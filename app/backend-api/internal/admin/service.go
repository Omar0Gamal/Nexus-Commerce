package admin

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/netip"
	"runtime"
	"time"

	"backend-api/internal/db"
	"backend-api/internal/email"
	"backend-api/internal/shared/apperr"
	"backend-api/internal/shared/pgutil"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

var (
	ErrShopNotFound  = errors.New("shop not found")
	ErrInvalidUUID   = apperr.ErrInvalidUUID
	ErrInvalidStatus = errors.New("invalid shop status")
)

// Service handles platform admin business logic for managing shops.
type Service struct {
	queries   *db.Queries
	mailer    *email.Mailer
	pool      *pgxpool.Pool
	rdb       *redis.Client
	startTime time.Time
	logger    *zap.Logger
}

func NewService(queries *db.Queries, mailer *email.Mailer, pool *pgxpool.Pool, rdb *redis.Client, logger *zap.Logger) *Service {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &Service{queries: queries, mailer: mailer, pool: pool, rdb: rdb, startTime: time.Now(), logger: logger}
}

// validStatuses are the allowed shop statuses.
var validStatuses = map[string]bool{
	"active":    true,
	"suspended": true,
	"paused":    true,
}

// ListShops returns a paginated list of all shops.
func (s *Service) ListShops(ctx context.Context, limit, offset int32, sortField, sortOrder string) ([]ShopResponse, int64, error) {
	shops, err := s.queries.AdminListShops(ctx, db.AdminListShopsParams{
		Sort:   sortField,
		Order:  sortOrder,
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, 0, fmt.Errorf("list shops: %w", err)
	}

	total, err := s.queries.CountShops(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("count shops: %w", err)
	}

	result := make([]ShopResponse, len(shops))
	for i, shop := range shops {
		resp := ShopResponse{
			ID:          shop.ID.String(),
			Name:        shop.Name,
			Subdomain:   shop.Subdomain,
			Status:      string(shop.Status.ShopStatus),
			Currency:    pgutil.TextToString(shop.Currency),
			Timezone:    pgutil.TextToString(shop.Timezone),
			OwnerUserID: shop.OwnerUserID.String(),
			PlanID:      shop.PlanID.String(),
			PlanName:    shop.PlanName,
		}

		if shop.CustomDomain.Valid {
			resp.CustomDomain = &shop.CustomDomain.String
		}
		if shop.CreatedAt.Valid {
			resp.CreatedAt = shop.CreatedAt.Time.Format("2006-01-02T15:04:05Z")
		}
		
		result[i] = resp
	}

	return result, total, nil
}

// GetShop returns details for a single shop.
func (s *Service) GetShop(ctx context.Context, shopID string) (*ShopResponse, error) {
	id, err := uuid.Parse(shopID)
	if err != nil {
		return nil, ErrInvalidUUID
	}

	shop, err := s.queries.GetShop(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrShopNotFound
		}
		return nil, fmt.Errorf("get shop: %w", err)
	}

	resp := shopToResponse(shop)
	return &resp, nil
}

// ActorInfo carries the platform admin identity for writing audit logs.
type ActorInfo struct {
	UserID    string
	Name      string
	IPAddress string
}

// UpdateShopStatus changes the status of a shop (active, suspended, paused)
// and records an audit log entry for the acting platform admin.
func (s *Service) UpdateShopStatus(ctx context.Context, shopID, status string, actor ActorInfo) error {
	if !validStatuses[status] {
		return ErrInvalidStatus
	}

	id, err := uuid.Parse(shopID)
	if err != nil {
		return ErrInvalidUUID
	}

	// Fetch current state so we can record the before/after in the audit log.
	shop, err := s.queries.GetShop(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrShopNotFound
		}
		return fmt.Errorf("get shop: %w", err)
	}

	err = s.queries.UpdateShopStatus(ctx, db.UpdateShopStatusParams{
		ID: id,
		Status: db.NullShopStatus{
			ShopStatus: db.ShopStatus(status),
			Valid:      true,
		},
	})
	if err != nil {
		return fmt.Errorf("update shop status: %w", err)
	}

	// Write audit log — fire-and-forget (failure must not roll back the update).
	s.writeAuditLog(ctx, shopID, id, actor, "update_status", "shop", map[string]any{
		"old_status": string(shop.Status.ShopStatus),
		"new_status": status,
	})

	// Send suspension email to shop owner when the shop is suspended.
	if status == "suspended" && s.mailer != nil {
		go func() {
			bgCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			owner, err := s.queries.GetUser(bgCtx, shop.OwnerUserID)
			if err != nil {
				s.logger.Error("failed to get shop owner for suspension email",
					zap.String("shop_id", shopID),
					zap.Error(err),
				)
				return
			}
			data := email.AccountSuspendedData{
				ShopName:     shop.Name,
				SupportEmail: "support@nexuscommerce.io",
			}
			if err := s.mailer.SendAccountSuspended(owner.Email, data); err != nil {
				s.logger.Error("failed to send account suspended email",
					zap.String("shop_id", shopID),
					zap.Error(err),
				)
			}
		}()
	}

	return nil
}

// writeAuditLog inserts a single audit log row. Errors are silently discarded
// so an audit write failure never affects the primary operation.
func (s *Service) writeAuditLog(
	ctx context.Context,
	shopIDStr string,
	resourceID uuid.UUID,
	actor ActorInfo,
	action, resourceType string,
	changes map[string]any,
) {
	changesJSON, err := json.Marshal(changes)
	if err != nil {
		changesJSON = []byte("{}")
	}

	var actorPgUUID pgtype.UUID
	if uid, err := uuid.Parse(actor.UserID); err == nil {
		actorPgUUID = pgtype.UUID{Bytes: uid, Valid: true}
	}

	var shopPgUUID pgtype.UUID
	if sid, err := uuid.Parse(shopIDStr); err == nil {
		shopPgUUID = pgtype.UUID{Bytes: sid, Valid: true}
	}

	var ipPtr *netip.Addr
	if actor.IPAddress != "" {
		if addr, err := netip.ParseAddr(actor.IPAddress); err == nil {
			ipPtr = &addr
		}
	}

	if err := s.queries.CreateAuditLog(ctx, db.CreateAuditLogParams{
		ShopID:       shopPgUUID,
		ActorUserID:  actorPgUUID,
		ActorName:    pgtype.Text{String: actor.Name, Valid: actor.Name != ""},
		IpAddress:    ipPtr,
		Action:       action,
		ResourceType: resourceType,
		ResourceID:   pgtype.UUID{Bytes: resourceID, Valid: true},
		Changes:      changesJSON,
	}); err != nil {
		s.logger.Error("failed to write admin audit log",
			zap.String("shop_id", shopIDStr),
			zap.String("action", action),
			zap.String("resource_type", resourceType),
			zap.Error(err),
		)
	}
}

// GetStats returns aggregate platform statistics.
func (s *Service) GetStats(ctx context.Context) (*PlatformStatsResponse, error) {
	total, err := s.queries.CountShops(ctx)
	if err != nil {
		return nil, fmt.Errorf("count shops: %w", err)
	}

	active, err := s.queries.CountShopsByStatus(ctx, db.NullShopStatus{
		ShopStatus: db.ShopStatusActive,
		Valid:      true,
	})
	if err != nil {
		return nil, fmt.Errorf("count active shops: %w", err)
	}

	suspended, err := s.queries.CountShopsByStatus(ctx, db.NullShopStatus{
		ShopStatus: db.ShopStatusSuspended,
		Valid:      true,
	})
	if err != nil {
		return nil, fmt.Errorf("count suspended shops: %w", err)
	}

	return &PlatformStatsResponse{
		TotalShops:     total,
		ActiveShops:    active,
		SuspendedShops: suspended,
	}, nil
}

func shopToResponse(shop db.Shop) ShopResponse {
	resp := ShopResponse{
		ID:          shop.ID.String(),
		Name:        shop.Name,
		Subdomain:   shop.Subdomain,
		Status:      string(shop.Status.ShopStatus),
		Currency:    pgutil.TextToString(shop.Currency),
		Timezone:    pgutil.TextToString(shop.Timezone),
		OwnerUserID: shop.OwnerUserID.String(),
		PlanID:      shop.PlanID.String(),
	}

	if shop.CustomDomain.Valid {
		resp.CustomDomain = &shop.CustomDomain.String
	}

	if shop.CreatedAt.Valid {
		resp.CreatedAt = shop.CreatedAt.Time.Format("2006-01-02T15:04:05Z")
	}

	return resp
}

// DBHealthInfo is returned by GetHealth.
type DBHealthInfo struct {
	TotalConns int32  `json:"total_conns"`
	IdleConns  int32  `json:"idle_conns"`
	MaxConns   int32  `json:"max_conns"`
	Status     string `json:"status"`
}

// RedisHealthInfo is returned by GetHealth.
type RedisHealthInfo struct {
	Status string `json:"status"`
}

// HealthResponse is returned by GetHealth.
type HealthResponse struct {
	Database      DBHealthInfo    `json:"database"`
	Redis         RedisHealthInfo `json:"redis"`
	Goroutines    int             `json:"goroutines"`
	UptimeSeconds int64           `json:"uptime_seconds"`
}

// GetHealth returns aggregate service health info.
func (s *Service) GetHealth(ctx context.Context) (*HealthResponse, error) {
	dbInfo := DBHealthInfo{Status: "ok"}
	if s.pool != nil {
		stat := s.pool.Stat()
		dbInfo.TotalConns = stat.TotalConns()
		dbInfo.IdleConns = stat.IdleConns()
		dbInfo.MaxConns = stat.MaxConns()
	}

	redisInfo := RedisHealthInfo{Status: "ok"}
	if s.rdb != nil {
		if err := s.rdb.Ping(ctx).Err(); err != nil {
			redisInfo.Status = "degraded"
		}
	}

	return &HealthResponse{
		Database:      dbInfo,
		Redis:         redisInfo,
		Goroutines:    runtime.NumGoroutine(),
		UptimeSeconds: int64(time.Since(s.startTime).Seconds()),
	}, nil
}

// PlanDistribution is a single entry in BillingSummaryResponse.
type PlanDistribution struct {
	PlanName  string `json:"plan_name"`
	ShopCount int64  `json:"shop_count"`
}

// BillingSummaryResponse is returned by GetBillingSummary.
type BillingSummaryResponse struct {
	ActiveShops      int64              `json:"active_shops"`
	SuspendedShops   int64              `json:"suspended_shops"`
	NewLast30Days    int64              `json:"new_last_30_days"`
	MRR              interface{}        `json:"mrr"`
	ARR              interface{}        `json:"arr,omitempty"`
	PlanDistribution []PlanDistribution `json:"plan_distribution"`
}

// GetBillingSummary aggregates platform billing statistics.
func (s *Service) GetBillingSummary(ctx context.Context) (*BillingSummaryResponse, error) {
	row, err := s.queries.GetBillingSummary(ctx)
	if err != nil {
		return nil, fmt.Errorf("get billing summary: %w", err)
	}

	dist, err := s.queries.GetPlanDistribution(ctx)
	if err != nil {
		return nil, fmt.Errorf("get plan distribution: %w", err)
	}

	plans := make([]PlanDistribution, 0, len(dist))
	for _, d := range dist {
		plans = append(plans, PlanDistribution{PlanName: d.PlanName, ShopCount: d.ShopCount})
	}

	return &BillingSummaryResponse{
		ActiveShops:      row.ActiveShops,
		SuspendedShops:   row.SuspendedShops,
		NewLast30Days:    row.NewLast30d,
		MRR:              row.Mrr,
		PlanDistribution: plans,
	}, nil
}
