package coupons

import (
	"errors"

	"backend-api/internal/shared/response"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// All routes require authentication and staff-level access.
func (h *Handler) RegisterRoutes(rg *gin.RouterGroup, requireAuth gin.HandlerFunc, requireStaff gin.HandlerFunc) {
	group := rg.Group("/coupons", requireAuth, requireStaff)
	{
		group.GET("", h.ListCoupons)
		group.POST("", h.CreateCoupon)
		group.GET("/:id", h.GetCoupon)
		group.PATCH("/:id", h.UpdateCoupon)
		group.PATCH("/:id/toggle", h.ToggleCoupon)
		group.DELETE("/:id", h.DeleteCoupon)
	}
}

func (h *Handler) CreateCoupon(c *gin.Context) {
	shopID := c.GetString("shop_id")

	var req CreateCouponRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	coupon, err := h.svc.CreateCoupon(c.Request.Context(), shopID, req)
	if err != nil {
		h.handleError(c, err)
		return
	}
	response.Created(c, coupon)
}

func (h *Handler) GetCoupon(c *gin.Context) {
	shopID := c.GetString("shop_id")

	coupon, err := h.svc.GetCoupon(c.Request.Context(), shopID, c.Param("id"))
	if err != nil {
		h.handleError(c, err)
		return
	}
	response.OK(c, coupon)
}

func (h *Handler) ListCoupons(c *gin.Context) {
	shopID := c.GetString("shop_id")

	coupons, err := h.svc.ListCoupons(c.Request.Context(), shopID)
	if err != nil {
		response.InternalError(c)
		return
	}
	response.OK(c, coupons)
}

func (h *Handler) UpdateCoupon(c *gin.Context) {
	shopID := c.GetString("shop_id")

	var req UpdateCouponRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	coupon, err := h.svc.UpdateCoupon(c.Request.Context(), shopID, c.Param("id"), req)
	if err != nil {
		h.handleError(c, err)
		return
	}
	response.OK(c, coupon)
}

func (h *Handler) ToggleCoupon(c *gin.Context) {
	shopID := c.GetString("shop_id")

	coupon, err := h.svc.ToggleCoupon(c.Request.Context(), shopID, c.Param("id"))
	if err != nil {
		h.handleError(c, err)
		return
	}
	response.OK(c, coupon)
}

func (h *Handler) DeleteCoupon(c *gin.Context) {
	shopID := c.GetString("shop_id")

	if err := h.svc.DeleteCoupon(c.Request.Context(), shopID, c.Param("id")); err != nil {
		h.handleError(c, err)
		return
	}
	response.NoContent(c)
}

func (h *Handler) handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrInvalidType):
		response.BadRequest(c, err.Error())
	case errors.Is(err, ErrCodeTaken):
		response.Conflict(c, err.Error())
	case errors.Is(err, ErrExpired),
		errors.Is(err, ErrUsageLimitReached),
		errors.Is(err, ErrMinOrder),
		errors.Is(err, ErrInactive):
		response.UnprocessableEntity(c, err.Error())
	default:
		response.HandleError(c, err)
	}
}
