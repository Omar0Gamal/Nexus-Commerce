package currency

import (
	"backend-api/internal/db"
	"backend-api/internal/shared/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) RegisterPublicRoutes(rg *gin.RouterGroup) {
	rg.GET("/currencies", h.GetActiveCurrencies)
}

func (h *Handler) RegisterStaffRoutes(rg *gin.RouterGroup, requireAuth gin.HandlerFunc) {
	locale := rg.Group("/shop/locale")
	locale.Use(requireAuth)
	{
		locale.GET("", h.GetShopLocale)
		locale.PUT("", h.UpdateShopLocale)
	}
}

func (h *Handler) GetActiveCurrencies(c *gin.Context) {
	currencies, err := h.svc.queries.GetActiveCurrencies(c.Request.Context())
	if err != nil {
		response.InternalError(c)
		return
	}
	if currencies == nil {
		currencies = []db.Currency{}
	}
	response.OK(c, currencies)
}

func (h *Handler) GetShopLocale(c *gin.Context) {
	shopIDStr := c.GetString("shop_id")
	shopID, err := uuid.Parse(shopIDStr)
	if err != nil {
		response.BadRequest(c, "invalid shop_id")
		return
	}
	settings, err := h.svc.GetShopSettings(c.Request.Context(), shopID)
	if err != nil {
		response.InternalError(c)
		return
	}
	response.OK(c, settings)
}

func (h *Handler) UpdateShopLocale(c *gin.Context) {
	shopIDStr := c.GetString("shop_id")
	shopID, err := uuid.Parse(shopIDStr)
	if err != nil {
		response.BadRequest(c, "invalid shop_id")
		return
	}

	var req struct {
		BaseCurrency      string   `json:"base_currency"      binding:"required"`
		DisplayCurrencies []string `json:"display_currencies" binding:"required"`
		DefaultLocale     string   `json:"default_locale"     binding:"required"`
		Timezone          string   `json:"timezone"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	result, err := h.svc.UpdateShopLocale(c.Request.Context(), shopID, req.BaseCurrency, req.DisplayCurrencies, req.DefaultLocale, req.Timezone)
	if err != nil {
		response.InternalError(c)
		return
	}
	response.OK(c, result)
}
