package promotions

import (
	"errors"
	"strconv"

	"backend-api/internal/shared/ginutil"
	"backend-api/internal/shared/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) RegisterStaffRoutes(rg *gin.RouterGroup, requireAuth, requireStaff gin.HandlerFunc) {
	fs := rg.Group("/flash-sales")
	fs.Use(requireAuth, requireStaff)
	{
		fs.GET("", h.ListFlashSales)
		fs.POST("", h.CreateFlashSale)
		fs.POST("/:id/items", h.AddFlashSaleItem)
		fs.DELETE("/:id", h.DeleteFlashSale)
	}

	td := rg.Group("/tiered-discounts")
	td.Use(requireAuth, requireStaff)
	{
		td.GET("", h.ListTieredDiscounts)
		td.POST("", h.CreateTieredDiscount)
	}

	lp := rg.Group("/loyalty/program")
	lp.Use(requireAuth, requireStaff)
	{
		lp.GET("", h.GetLoyaltyProgram)
		lp.PUT("", h.UpdateLoyaltyProgram)
	}

	// Staff referral config + stats
	refStaff := rg.Group("/referral")
	refStaff.Use(requireAuth, requireStaff)
	{
		refStaff.GET("/config", h.GetReferralConfig)
		refStaff.PATCH("/config", h.UpdateReferralConfig)
		refStaff.GET("/stats", h.GetReferralStats)
	}

	// Staff loyalty tier management + leaderboard + manual adjust
	tierStaff := rg.Group("/loyalty")
	tierStaff.Use(requireAuth, requireStaff)
	{
		tierStaff.GET("/tiers", h.ListLoyaltyTiers)
		tierStaff.POST("/tiers", h.CreateLoyaltyTier)
		tierStaff.PATCH("/tiers/:id", h.UpdateLoyaltyTier)
		tierStaff.DELETE("/tiers/:id", h.DeleteLoyaltyTier)
		tierStaff.POST("/adjust", h.ManualAdjustPoints)
		tierStaff.GET("/leaderboard", h.GetLeaderboard)
	}
}

func (h *Handler) RegisterCustomerRoutes(rg *gin.RouterGroup, requireAuth, requireCustomer gin.HandlerFunc) {
	loy := rg.Group("/loyalty")
	loy.Use(requireAuth, requireCustomer)
	{
		loy.GET("/balance", h.GetMyBalance)
		loy.POST("/redeem", h.RedeemPoints)
		loy.GET("/transactions", h.GetTransactions)
		loy.GET("/my-tier", h.GetMyTierStatus)
		loy.GET("/public-tiers", h.GetPublicTiers)
	}

	ref := rg.Group("/referral")
	ref.Use(requireAuth, requireCustomer)
	{
		ref.GET("/code", h.GetMyReferralCode)
		ref.GET("/my-referrals", h.GetMyReferrals)
	}
}


func (h *Handler) ListFlashSales(c *gin.Context) {
	sid, ok := ginutil.ShopUUID(c)
	if !ok {
		return
	}
	sales, err := h.svc.ListFlashSales(c.Request.Context(), sid)
	if err != nil {
		response.InternalError(c)
		return
	}
	response.OK(c, sales)
}

func (h *Handler) CreateFlashSale(c *gin.Context) {
	sid, ok := ginutil.ShopUUID(c)
	if !ok {
		return
	}
	var req CreateFlashSaleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	sale, err := h.svc.CreateFlashSale(c.Request.Context(), sid, req)
	if err != nil {
		response.InternalError(c)
		return
	}
	response.Created(c, sale)
}

func (h *Handler) AddFlashSaleItem(c *gin.Context) {
	sid, ok := ginutil.ShopUUID(c)
	if !ok {
		return
	}
	fsID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid flash sale id")
		return
	}
	var req AddFlashSaleItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	productID, err := uuid.Parse(req.ProductID)
	if err != nil {
		response.BadRequest(c, "invalid product_id")
		return
	}
	var variantUUID *uuid.UUID
	if req.VariantID != nil {
		vid, err := uuid.Parse(*req.VariantID)
		if err != nil {
			response.BadRequest(c, "invalid variant_id")
			return
		}
		variantUUID = &vid
	}
	item, err := h.svc.AddFlashSaleItem(c.Request.Context(), sid, fsID, productID, variantUUID, req.SalePriceCents)
	if err != nil {
		response.InternalError(c)
		return
	}
	response.Created(c, item)
}

func (h *Handler) DeleteFlashSale(c *gin.Context) {
	sid, ok := ginutil.ShopUUID(c)
	if !ok {
		return
	}
	fsID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}
	if err := h.svc.DeleteFlashSale(c.Request.Context(), sid, fsID); err != nil {
		response.InternalError(c)
		return
	}
	response.OK(c, gin.H{"deleted": true})
}


func (h *Handler) ListTieredDiscounts(c *gin.Context) {
	sid, ok := ginutil.ShopUUID(c)
	if !ok {
		return
	}
	tds, err := h.svc.ListTieredDiscounts(c.Request.Context(), sid)
	if err != nil {
		response.InternalError(c)
		return
	}
	response.OK(c, tds)
}

func (h *Handler) CreateTieredDiscount(c *gin.Context) {
	sid, ok := ginutil.ShopUUID(c)
	if !ok {
		return
	}
	var req CreateTieredDiscountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	// Validate optional target_id format before delegating to service.
	if req.TargetID != nil {
		if _, err := uuid.Parse(*req.TargetID); err != nil {
			response.BadRequest(c, "invalid target_id")
			return
		}
	}
	td, err := h.svc.CreateTieredDiscount(c.Request.Context(), sid, req)
	if err != nil {
		response.InternalError(c)
		return
	}
	response.Created(c, td)
}


func (h *Handler) GetLoyaltyProgram(c *gin.Context) {
	sid, ok := ginutil.ShopUUID(c)
	if !ok {
		return
	}
	prog, err := h.svc.GetLoyaltyProgram(c.Request.Context(), sid)
	if errors.Is(err, pgx.ErrNoRows) {
		response.NotFound(c, "loyalty program not configured")
		return
	}
	if err != nil {
		response.InternalError(c)
		return
	}
	response.OK(c, prog)
}

func (h *Handler) UpdateLoyaltyProgram(c *gin.Context) {
	sid, ok := ginutil.ShopUUID(c)
	if !ok {
		return
	}
	var req UpdateLoyaltyProgramRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	prog, err := h.svc.UpsertLoyaltyProgram(c.Request.Context(), sid, req)
	if err != nil {
		response.InternalError(c)
		return
	}
	response.OK(c, prog)
}


func (h *Handler) GetMyBalance(c *gin.Context) {
	sid, ok := ginutil.ShopUUID(c)
	if !ok {
		return
	}
	cid, ok := ginutil.CustomerUUID(c)
	if !ok {
		return
	}
	balance, err := h.svc.GetLoyaltyBalance(c.Request.Context(), sid, cid)
	if err != nil {
		response.InternalError(c)
		return
	}
	response.OK(c, gin.H{"points_balance": balance})
}

func (h *Handler) RedeemPoints(c *gin.Context) {
	sid, ok := ginutil.ShopUUID(c)
	if !ok {
		return
	}
	cid, ok := ginutil.CustomerUUID(c)
	if !ok {
		return
	}
	var req RedeemPointsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	discountCents, err := h.svc.RedeemPoints(c.Request.Context(), sid, cid, req.Points, req.OrderAmountCents)
	if errors.Is(err, ErrInsufficientPoints) {
		response.BadRequest(c, "insufficient loyalty points")
		return
	}
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, gin.H{"discount_cents": discountCents})
}

func (h *Handler) GetTransactions(c *gin.Context) {
	sid, ok := ginutil.ShopUUID(c)
	if !ok {
		return
	}
	cid, ok := ginutil.CustomerUUID(c)
	if !ok {
		return
	}
	limit := int32(50)
	offset := int32(0)
	if l := c.Query("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 {
			limit = int32(v)
		}
	}
	if o := c.Query("offset"); o != "" {
		if v, err := strconv.Atoi(o); err == nil && v >= 0 {
			offset = int32(v)
		}
	}
	txs, err := h.svc.GetLoyaltyTransactions(c.Request.Context(), sid, cid, limit, offset)
	if err != nil {
		response.InternalError(c)
		return
	}
	response.OK(c, txs)
}


func (h *Handler) GetMyReferralCode(c *gin.Context) {
	sid, ok := ginutil.ShopUUID(c)
	if !ok {
		return
	}
	cid, ok := ginutil.CustomerUUID(c)
	if !ok {
		return
	}
	code, err := h.svc.GetOrCreateReferralCode(c.Request.Context(), sid, cid)
	if err != nil {
		response.InternalError(c)
		return
	}
	response.OK(c, gin.H{"code": code})
}

func (h *Handler) GetMyReferrals(c *gin.Context) {
	sid, ok := ginutil.ShopUUID(c)
	if !ok {
		return
	}
	cid, ok := ginutil.CustomerUUID(c)
	if !ok {
		return
	}
	rows, err := h.svc.GetMyReferrals(c.Request.Context(), sid, cid)
	if err != nil {
		response.InternalError(c)
		return
	}
	response.OK(c, rows)
}

func (h *Handler) GetReferralConfig(c *gin.Context) {
	sid, ok := ginutil.ShopUUID(c)
	if !ok {
		return
	}
	cfg, err := h.svc.GetReferralConfig(c.Request.Context(), sid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			response.OK(c, gin.H{"message": "no referral config set"})
			return
		}
		response.InternalError(c)
		return
	}
	response.OK(c, cfg)
}

func (h *Handler) UpdateReferralConfig(c *gin.Context) {
	sid, ok := ginutil.ShopUUID(c)
	if !ok {
		return
	}
	var req UpsertReferralConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	cfg, err := h.svc.UpsertReferralConfig(c.Request.Context(), sid, req)
	if err != nil {
		response.InternalError(c)
		return
	}
	response.OK(c, cfg)
}

func (h *Handler) GetReferralStats(c *gin.Context) {
	sid, ok := ginutil.ShopUUID(c)
	if !ok {
		return
	}
	stats, err := h.svc.GetReferralStats(c.Request.Context(), sid)
	if err != nil {
		response.InternalError(c)
		return
	}
	response.OK(c, stats)
}


func (h *Handler) ListLoyaltyTiers(c *gin.Context) {
	sid, ok := ginutil.ShopUUID(c)
	if !ok {
		return
	}
	tiers, err := h.svc.ListLoyaltyTiers(c.Request.Context(), sid)
	if err != nil {
		response.InternalError(c)
		return
	}
	response.OK(c, tiers)
}

func (h *Handler) GetPublicTiers(c *gin.Context) {
	sid, ok := ginutil.ShopUUID(c)
	if !ok {
		return
	}
	tiers, err := h.svc.ListLoyaltyTiers(c.Request.Context(), sid)
	if err != nil {
		response.InternalError(c)
		return
	}
	response.OK(c, tiers)
}

func (h *Handler) CreateLoyaltyTier(c *gin.Context) {
	sid, ok := ginutil.ShopUUID(c)
	if !ok {
		return
	}
	var req CreateLoyaltyTierRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	tier, err := h.svc.CreateLoyaltyTier(c.Request.Context(), sid, req)
	if err != nil {
		response.InternalError(c)
		return
	}
	response.Created(c, tier)
}

func (h *Handler) UpdateLoyaltyTier(c *gin.Context) {
	sid, ok := ginutil.ShopUUID(c)
	if !ok {
		return
	}
	tierID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid tier id")
		return
	}
	var req UpdateLoyaltyTierRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	tier, err := h.svc.UpdateLoyaltyTier(c.Request.Context(), sid, tierID, req)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			response.NotFound(c, "tier not found")
			return
		}
		response.InternalError(c)
		return
	}
	response.OK(c, tier)
}

func (h *Handler) DeleteLoyaltyTier(c *gin.Context) {
	sid, ok := ginutil.ShopUUID(c)
	if !ok {
		return
	}
	tierID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid tier id")
		return
	}
	if err := h.svc.DeleteLoyaltyTier(c.Request.Context(), sid, tierID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			response.NotFound(c, "tier not found")
			return
		}
		response.InternalError(c)
		return
	}
	response.NoContent(c)
}

func (h *Handler) GetMyTierStatus(c *gin.Context) {
	sid, ok := ginutil.ShopUUID(c)
	if !ok {
		return
	}
	cid, ok := ginutil.CustomerUUID(c)
	if !ok {
		return
	}
	status, err := h.svc.GetCustomerTierStatus(c.Request.Context(), sid, cid)
	if err != nil {
		response.InternalError(c)
		return
	}
	response.OK(c, status)
}

func (h *Handler) ManualAdjustPoints(c *gin.Context) {
	sid, ok := ginutil.ShopUUID(c)
	if !ok {
		return
	}
	var req ManualAdjustPointsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if err := h.svc.ManualAdjustPoints(c.Request.Context(), sid, req); err != nil {
		response.InternalError(c)
		return
	}
	response.NoContent(c)
}

func (h *Handler) GetLeaderboard(c *gin.Context) {
	sid, ok := ginutil.ShopUUID(c)
	if !ok {
		return
	}
	limitStr := c.DefaultQuery("limit", "20")
	var limit int32 = 20
	if n, err := strconv.Atoi(limitStr); err == nil && n > 0 && n <= 100 {
		limit = int32(n)
	}
	rows, err := h.svc.GetLoyaltyLeaderboard(c.Request.Context(), sid, limit)
	if err != nil {
		response.InternalError(c)
		return
	}
	response.OK(c, rows)
}
