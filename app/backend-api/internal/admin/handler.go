package admin

import (
	"errors"
	"strconv"

	"backend-api/internal/shared/response"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// These routes require RequireAuth + RequirePlatformAdmin middleware.
func (h *Handler) RegisterRoutes(rg *gin.RouterGroup, authMW, requireAdmin gin.HandlerFunc) {
	admin := rg.Group("/admin")
	admin.Use(authMW, requireAdmin)
	{
		admin.GET("/stats", h.GetStats)
		admin.GET("/shops", h.ListShops)
		admin.GET("/shops/:id", h.GetShop)
		admin.PATCH("/shops/:id/status", h.UpdateShopStatus)
	}
}


func (h *Handler) GetStats(c *gin.Context) {
	stats, err := h.svc.GetStats(c.Request.Context())
	if err != nil {
		response.InternalError(c)
		return
	}
	response.OK(c, stats)
}

func (h *Handler) ListShops(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))

	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}

	offset := int32((page - 1) * perPage)
	limit := int32(perPage)

	shops, total, err := h.svc.ListShops(c.Request.Context(), limit, offset)
	if err != nil {
		response.InternalError(c)
		return
	}

	totalPages := int(response.CalcTotalPages(total, int64(perPage)))

	response.OKWithMeta(c, shops, gin.H{
		"page":        page,
		"per_page":    perPage,
		"total_items": total,
		"total_pages": totalPages,
	})
}

func (h *Handler) GetShop(c *gin.Context) {
	shop, err := h.svc.GetShop(c.Request.Context(), c.Param("id"))
	if err != nil {
		h.handleError(c, err)
		return
	}
	response.OK(c, shop)
}

func (h *Handler) UpdateShopStatus(c *gin.Context) {
	var req UpdateShopStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	actor := ActorInfo{
		UserID:    c.GetString("user_id"),
		Name:      c.GetString("full_name"),
		IPAddress: c.ClientIP(),
	}

	err := h.svc.UpdateShopStatus(c.Request.Context(), c.Param("id"), req.Status, actor)
	if err != nil {
		h.handleError(c, err)
		return
	}

	// Return updated shop
	shop, err := h.svc.GetShop(c.Request.Context(), c.Param("id"))
	if err != nil {
		response.InternalError(c)
		return
	}

	response.OK(c, shop)
}


func (h *Handler) handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrShopNotFound):
		response.NotFound(c, "Shop not found")
	case errors.Is(err, ErrInvalidUUID):
		response.BadRequest(c, "Invalid shop ID")
	case errors.Is(err, ErrInvalidStatus):
		response.BadRequest(c, "Invalid status. Must be: active, suspended, or paused")
	default:
		response.InternalError(c)
	}
}
