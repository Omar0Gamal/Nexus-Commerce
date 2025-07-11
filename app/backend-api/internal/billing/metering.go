package billing

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"backend-api/internal/db"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

const quotaCacheTTL = 60 * time.Second

// MeteringService wraps billing.Service with Redis for quota enforcement.
type MeteringService struct {
	svc *Service
	rdb *redis.Client
	log *zap.Logger
}

func NewMeteringService(svc *Service, rdb *redis.Client) *MeteringService {
	return &MeteringService{svc: svc, rdb: rdb, log: zap.NewNop()}
}

// RequireQuota returns a Gin middleware that enforces resource quota limits.
//
// resource is one of: "products", "staff".
// It reads the current count from Postgres (cached 60 s in Redis) and compares
// it against the plan's limit field. -1 or nil means unlimited.
func (m *MeteringService) RequireQuota(resource string) gin.HandlerFunc {
	return func(c *gin.Context) {
		shopID := c.GetString("shop_id")
		if shopID == "" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "shop context missing"})
			return
		}

		exceeded, err := m.isQuotaExceeded(c.Request.Context(), shopID, resource)
		if err != nil {
			// Non-fatal: let request through on quota check failure.
			c.Next()
			return
		}
		if exceeded {
			c.AbortWithStatusJSON(http.StatusUnprocessableEntity, gin.H{
				"error":    "quota_exceeded",
				"resource": resource,
				"message":  fmt.Sprintf("You have reached the %s limit for your current plan. Please upgrade.", resource),
			})
			return
		}

		c.Next()
	}
}

func (m *MeteringService) isQuotaExceeded(ctx context.Context, shopID, resource string) (bool, error) {
	shopUUID, err := uuid.Parse(shopID)
	if err != nil {
		return false, nil
	}

	limit, err := m.getPlanLimit(ctx, shopUUID, resource)
	if err != nil {
		return false, err
	}
	if limit < 0 {
		// -1 means unlimited.
		return false, nil
	}

	count, err := m.getCurrentCount(ctx, shopUUID, resource)
	if err != nil {
		return false, err
	}

	return count >= limit, nil
}

// getPlanLimit returns the plan limit for the resource, or -1 for unlimited.
func (m *MeteringService) getPlanLimit(ctx context.Context, shopID uuid.UUID, resource string) (int64, error) {
	cacheKey := fmt.Sprintf("quota_limit:%s:%s", shopID, resource)

	// Try Redis cache.
	if m.rdb != nil {
		if cached, err := m.rdb.Get(ctx, cacheKey).Result(); err == nil {
			if v, err := strconv.ParseInt(cached, 10, 64); err == nil {
				return v, nil
			}
		}
	}

	shop, err := m.svc.q.GetShop(ctx, shopID)
	if err != nil {
		return -1, fmt.Errorf("get shop: %w", err)
	}
	plan, err := m.svc.q.GetPlan(ctx, shop.PlanID)
	if err != nil {
		return -1, fmt.Errorf("get plan: %w", err)
	}

	var limit int64
	switch resource {
	case "products":
		if !plan.MaxProducts.Valid {
			limit = -1
		} else {
			limit = int64(plan.MaxProducts.Int32)
		}
	case "staff":
		if !plan.MaxStaffAccounts.Valid {
			limit = -1
		} else {
			limit = int64(plan.MaxStaffAccounts.Int32)
		}
	default:
		// Check features JSONB for custom quota keys (e.g. "max_api_keys").
		limit = -1
		if len(plan.Features) > 0 {
			var features map[string]any
			if err := json.Unmarshal(plan.Features, &features); err == nil {
				key := "max_" + resource
				if v, ok := features[key]; ok {
					switch n := v.(type) {
					case float64:
						limit = int64(n)
					case string:
						if parsed, err := strconv.ParseInt(n, 10, 64); err == nil {
							limit = parsed
						}
					}
				}
			}
		}
	}

	// Cache the limit.
	if m.rdb != nil {
		if err := m.rdb.Set(ctx, cacheKey, strconv.FormatInt(limit, 10), quotaCacheTTL).Err(); err != nil {
			m.log.Warn("failed to cache plan limit",
				zap.String("cache_key", cacheKey),
				zap.Error(err),
			)
		}
	}
	return limit, nil
}

// getCurrentCount returns the current resource count for the shop.
func (m *MeteringService) getCurrentCount(ctx context.Context, shopID uuid.UUID, resource string) (int64, error) {
	cacheKey := fmt.Sprintf("quota_count:%s:%s", shopID, resource)

	// Try Redis cache.
	if m.rdb != nil {
		if cached, err := m.rdb.Get(ctx, cacheKey).Result(); err == nil {
			if v, err := strconv.ParseInt(cached, 10, 64); err == nil {
				return v, nil
			}
		}
	}

	shopPgUUID := pgtype.UUID{Bytes: shopID, Valid: true}
	var count int64
	var err error

	switch resource {
	case "products":
		count, err = m.svc.q.CountProducts(ctx, db.CountProductsParams{
			ShopID: shopPgUUID,
		})
	case "staff":
		count, err = m.svc.q.CountShopStaff(ctx, shopPgUUID)
	default:
		return 0, nil
	}
	if err != nil {
		return 0, err
	}

	// Cache the count.
	if m.rdb != nil {
		if err := m.rdb.Set(ctx, cacheKey, strconv.FormatInt(count, 10), quotaCacheTTL).Err(); err != nil {
			m.log.Warn("failed to cache quota count",
				zap.String("cache_key", cacheKey),
				zap.Error(err),
			)
		}
	}
	return count, nil
}
