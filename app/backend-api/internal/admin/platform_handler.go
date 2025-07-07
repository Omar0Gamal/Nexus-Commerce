package admin

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"time"

	"backend-api/internal/db"
	"backend-api/internal/shared/response"
	"backend-api/internal/shared/token"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

// PlatformHandler adds the extended platform-admin API endpoints that require
// the token service (impersonation) and access to AI config.
type PlatformHandler struct {
	svc      *Service
	tokenSvc *token.Service
}

func NewPlatformHandler(svc *Service, tokenSvc *token.Service) *PlatformHandler {
	return &PlatformHandler{svc: svc, tokenSvc: tokenSvc}
}

// GET  /api/v1/platform-admin/shops          — paginated list with MRR, plan, last_active_at
// GET  /api/v1/platform-admin/revenue        — platform MRR / ARR / churn
// GET  /api/v1/platform-admin/shops/:id      — full shop detail + invoices
// PUT  /api/v1/platform-admin/shops/:id/plan — force plan change
// POST /api/v1/platform-admin/shops/:id/impersonate — issues 1h staff JWT
// PUT  /api/v1/platform-admin/ai-config      — rotate AI API keys
// GET  /api/v1/platform-admin/health         — aggregate health dashboard
// GET  /api/v1/platform-admin/billing        — MRR, plan distribution, shop counts
func (h *PlatformHandler) RegisterRoutes(rg *gin.RouterGroup, authMW, requireAdmin gin.HandlerFunc) {
	g := rg.Group("/platform-admin")
	g.Use(authMW, requireAdmin)
	{
		g.GET("/shops", h.ListShops)
		g.GET("/revenue", h.GetRevenue)
		g.GET("/shops/:id", h.GetShopDetail)
		g.PUT("/shops/:id/plan", h.ForcePlanChange)
		g.POST("/shops/:id/impersonate", h.ImpersonateShop)
		g.PUT("/ai-config", h.UpsertAIConfig)
		g.GET("/health", h.GetHealth)
		g.GET("/billing", h.GetBillingSummary)
	}
}

func (h *PlatformHandler) ListShops(c *gin.Context) {
	page := queryInt(c, "page", 1)
	perPage := queryInt(c, "limit", 20) // Use limit instead of per_page if frontend sends limit
	if perPage > 100 {
		perPage = 100
	}
	offset := int32((page - 1) * perPage)

	sortField := c.DefaultQuery("sort", "created_at")
	sortOrder := c.DefaultQuery("order", "desc")

	shops, total, err := h.svc.ListShops(c.Request.Context(), int32(perPage), offset, sortField, sortOrder)
	if err != nil {
		response.InternalError(c)
		return
	}

	totalPages := int(response.CalcTotalPages(total, int64(perPage)))
	response.OKWithMeta(c, shops, response.PaginationMeta{
		Page:       page,
		PerPage:    perPage,
		TotalItems: total,
		TotalPages: int64(totalPages),
	})
}

func (h *PlatformHandler) GetRevenue(c *gin.Context) {
	stats, err := h.getPlatformRevenue(c.Request.Context())
	if err != nil {
		response.InternalError(c)
		return
	}
	response.OK(c, stats)
}

type revenueStats struct {
	MRR         string `json:"mrr"`
	ARR         string `json:"arr"`
	ActiveShops int64  `json:"active_shops"`
	TotalShops  int64  `json:"total_shops"`
}

func (h *PlatformHandler) getPlatformRevenue(ctx context.Context) (*revenueStats, error) {
	total, err := h.svc.queries.CountShops(ctx)
	if err != nil {
		return nil, err
	}
	active, err := h.svc.queries.CountShopsByStatus(ctx, db.NullShopStatus{ShopStatus: db.ShopStatusActive, Valid: true})
	if err != nil {
		return nil, err
	}

	// MRR = sum of all active shops' plan monthly prices.
	shops, err := h.svc.queries.ListShopsByStatus(ctx, db.NullShopStatus{ShopStatus: db.ShopStatusActive, Valid: true})
	if err != nil {
		return nil, err
	}
	planCache := map[uuid.UUID]db.Plan{}
	var totalMRR float64
	for _, s := range shops {
		p, ok := planCache[s.PlanID]
		if !ok {
			p, err = h.svc.queries.GetPlan(ctx, s.PlanID)
			if err != nil {
				continue
			}
			planCache[s.PlanID] = p
		}
		totalMRR += parseMoney(numericToStr(p.MonthlyPrice))
	}

	mrrStr := fmt.Sprintf("%.2f", totalMRR)
	arrStr := fmt.Sprintf("%.2f", totalMRR*12)

	return &revenueStats{
		MRR:         mrrStr,
		ARR:         arrStr,
		ActiveShops: active,
		TotalShops:  total,
	}, nil
}

func (h *PlatformHandler) GetShopDetail(c *gin.Context) {
	shop, err := h.svc.GetShop(c.Request.Context(), c.Param("id"))
	if err != nil {
		h.handleError(c, err)
		return
	}
	response.OK(c, shop)
}

func (h *PlatformHandler) ForcePlanChange(c *gin.Context) {
	shopIDStr := c.Param("id")
	shopID, err := uuid.Parse(shopIDStr)
	if err != nil {
		response.BadRequest(c, "invalid shop id")
		return
	}

	var req struct {
		PlanID string `json:"plan_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	planID, err := uuid.Parse(req.PlanID)
	if err != nil {
		response.BadRequest(c, "invalid plan id")
		return
	}

	// Verify plan exists.
	if _, err := h.svc.queries.GetPlan(c.Request.Context(), planID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			response.NotFound(c, "plan not found")
		} else {
			response.InternalError(c)
		}
		return
	}

	// Use UpdateShop preserving existing fields.
	shopRow, err := h.svc.queries.GetShop(c.Request.Context(), shopID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			response.NotFound(c, "shop not found")
		} else {
			response.InternalError(c)
		}
		return
	}
	if _, err := h.svc.queries.UpdateShop(c.Request.Context(), db.UpdateShopParams{
		ID:               shopID,
		Name:             pgtype.Text{String: shopRow.Name, Valid: true},
		Subdomain:        pgtype.Text{String: shopRow.Subdomain, Valid: true},
		CustomDomain:     shopRow.CustomDomain,
		Status:           shopRow.Status,
		Currency:         shopRow.Currency,
		Timezone:         shopRow.Timezone,
		PlanID:           pgtype.UUID{Bytes: planID, Valid: true},
		CurrentPeriodEnd: shopRow.CurrentPeriodEnd,
		IsOverdue:        shopRow.IsOverdue,
	}); err != nil {
		response.InternalError(c)
		return
	}

	actor := ActorInfo{
		UserID:    c.GetString("user_id"),
		Name:      c.GetString("full_name"),
		IPAddress: c.ClientIP(),
	}
	h.svc.writeAuditLog(c.Request.Context(), shopIDStr, shopID, actor,
		"force_plan_change", "shop", map[string]any{"new_plan_id": req.PlanID})

	response.OK(c, gin.H{"status": "ok"})
}

// Issues a 1-hour staff JWT scoped to the target shop's owner account.
func (h *PlatformHandler) ImpersonateShop(c *gin.Context) {
	shopIDStr := c.Param("id")
	shopID, err := uuid.Parse(shopIDStr)
	if err != nil {
		response.BadRequest(c, "invalid shop id")
		return
	}

	shopPgUUID := pgtype.UUID{Bytes: shopID, Valid: true}
	ownerStaff, err := h.svc.queries.GetShopOwner(c.Request.Context(), shopPgUUID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			response.NotFound(c, "shop owner not found")
		} else {
			response.InternalError(c)
		}
		return
	}

	ownerUser, err := h.svc.queries.GetUser(c.Request.Context(), ownerStaff.UserID.Bytes)
	if err != nil {
		response.InternalError(c)
		return
	}

	claims := token.Claims{
		UserID:    ownerUser.ID.String(),
		Email:     ownerUser.Email,
		FullName:  ownerUser.FullName.String,
		ActorType: token.ActorStaff,
		ShopID:    shopIDStr,
		StaffID:   ownerStaff.ID.String(),
		IsOwner:   true,
	}

	impersonationToken, err := h.tokenSvc.GenerateTokenWithDuration(claims, time.Hour)
	if err != nil {
		response.InternalError(c)
		return
	}

	// Audit log.
	adminID := c.GetString("user_id")
	actor := ActorInfo{
		UserID:    adminID,
		Name:      c.GetString("full_name"),
		IPAddress: c.ClientIP(),
	}
	h.svc.writeAuditLog(c.Request.Context(), shopIDStr, shopID, actor,
		"impersonate", "shop", map[string]any{"impersonated_user": ownerUser.Email})

	response.OK(c, gin.H{
		"access_token": impersonationToken,
		"expires_in":   3600,
		"shop_id":      shopIDStr,
		"note":         "impersonation token — expires in 1 hour",
	})
}

func (h *PlatformHandler) UpsertAIConfig(c *gin.Context) {
	var req struct {
		PrimaryProvider     string `json:"primary_provider"`
		PrimaryAPIKey       string `json:"primary_api_key"`
		PrimaryModel        string `json:"primary_model"`
		FallbackProvider    string `json:"fallback_provider"`
		FallbackAPIKey      string `json:"fallback_api_key"`
		FallbackModel       string `json:"fallback_model"`
		EngineBaseURL       string `json:"engine_base_url"`
		EngineAPIKey        string `json:"engine_api_key"`
		MaxTokensPerRequest int32  `json:"max_tokens_per_request"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if req.PrimaryProvider == "" {
		response.BadRequest(c, "primary_provider is required")
		return
	}
	if req.MaxTokensPerRequest <= 0 {
		req.MaxTokensPerRequest = 2048
	}

	cfg, err := h.svc.queries.UpsertAIConfig(c.Request.Context(), db.UpsertAIConfigParams{
		PrimaryProvider:         req.PrimaryProvider,
		PrimaryApiKeyEncrypted:  pgtype.Text{String: req.PrimaryAPIKey, Valid: req.PrimaryAPIKey != ""},
		PrimaryModel:            req.PrimaryModel,
		FallbackProvider:        pgtype.Text{String: req.FallbackProvider, Valid: req.FallbackProvider != ""},
		FallbackApiKeyEncrypted: pgtype.Text{String: req.FallbackAPIKey, Valid: req.FallbackAPIKey != ""},
		FallbackModel:           pgtype.Text{String: req.FallbackModel, Valid: req.FallbackModel != ""},
		EngineBaseUrl:           pgtype.Text{String: req.EngineBaseURL, Valid: req.EngineBaseURL != ""},
		EngineApiKeyEncrypted:   pgtype.Text{String: req.EngineAPIKey, Valid: req.EngineAPIKey != ""},
		MaxTokensPerRequest:     req.MaxTokensPerRequest,
	})
	if err != nil {
		response.InternalError(c)
		return
	}

	response.OK(c, gin.H{
		"primary_provider": cfg.PrimaryProvider,
		"primary_model":    cfg.PrimaryModel,
		"updated_at":       cfg.UpdatedAt.Time.Format(time.RFC3339),
	})
}

func (h *PlatformHandler) GetHealth(c *gin.Context) {
	health, err := h.svc.GetHealth(c.Request.Context())
	if err != nil {
		response.InternalError(c)
		return
	}
	response.OK(c, health)
}

func (h *PlatformHandler) GetBillingSummary(c *gin.Context) {
	summary, err := h.svc.GetBillingSummary(c.Request.Context())
	if err != nil {
		response.InternalError(c)
		return
	}
	response.OK(c, summary)
}

func (h *PlatformHandler) handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrShopNotFound):
		response.NotFound(c, "shop not found")
	case errors.Is(err, ErrInvalidUUID):
		response.BadRequest(c, "invalid shop id")
	default:
		response.InternalError(c)
	}
}

func queryInt(c *gin.Context, key string, defaultVal int) int {
	v := c.DefaultQuery(key, "")
	if v == "" {
		return defaultVal
	}
	n := 0
	if _, err := fmt.Sscanf(v, "%d", &n); err == nil && n > 0 {
		return n
	}
	return defaultVal
}

func parseMoney(s string) float64 {
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return f
}

// numericToStr converts a pgtype.Numeric to a decimal string.
func numericToStr(n pgtype.Numeric) string {
	if !n.Valid || n.Int == nil {
		return "0"
	}
	val := new(big.Float).SetInt(n.Int)
	if n.Exp > 0 {
		for i := int32(0); i < n.Exp; i++ {
			val.Mul(val, big.NewFloat(10))
		}
	} else if n.Exp < 0 {
		for i := int32(0); i > n.Exp; i-- {
			val.Quo(val, big.NewFloat(10))
		}
	}
	return val.Text('f', 2)
}
