package shipping

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

//
//	/shipping/zones            GET, POST
//	/shipping/zones/:id        GET, PATCH, DELETE
//	/shipping/zones/:id/rates  GET
//	/shipping/rates            GET
//	/shipping/rates/:id        GET, PATCH, DELETE
//	/shipping/rates            POST  — created via zone's rates endpoint too
func (h *Handler) RegisterRoutes(rg *gin.RouterGroup, requireAuth gin.HandlerFunc, requireStaff gin.HandlerFunc) {
	g := rg.Group("/shipping", requireAuth, requireStaff)

	// Zones
	zones := g.Group("/zones")
	{
		zones.GET("", h.ListZones)
		zones.POST("", h.CreateZone)
		zones.GET("/:id", h.GetZone)
		zones.PATCH("/:id", h.UpdateZone)
		zones.DELETE("/:id", h.DeleteZone)
		zones.GET("/:id/rates", h.ListRatesByZone)
	}

	// Rates (flat access)
	rates := g.Group("/rates")
	{
		rates.GET("", h.ListRates)
		rates.POST("", h.CreateRate)
		rates.GET("/:id", h.GetRate)
		rates.PATCH("/:id", h.UpdateRate)
		rates.DELETE("/:id", h.DeleteRate)
	}
}


func (h *Handler) CreateZone(c *gin.Context) {
	shopID := c.GetString("shop_id")
	var req CreateZoneRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	zone, err := h.svc.CreateZone(c.Request.Context(), shopID, req)
	if err != nil {
		h.handleError(c, err)
		return
	}
	response.Created(c, zone)
}

func (h *Handler) GetZone(c *gin.Context) {
	shopID := c.GetString("shop_id")
	zone, err := h.svc.GetZone(c.Request.Context(), shopID, c.Param("id"))
	if err != nil {
		h.handleError(c, err)
		return
	}
	response.OK(c, zone)
}

func (h *Handler) ListZones(c *gin.Context) {
	shopID := c.GetString("shop_id")
	zones, err := h.svc.ListZones(c.Request.Context(), shopID)
	if err != nil {
		response.InternalError(c)
		return
	}
	response.OK(c, zones)
}

func (h *Handler) UpdateZone(c *gin.Context) {
	shopID := c.GetString("shop_id")
	var req UpdateZoneRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	zone, err := h.svc.UpdateZone(c.Request.Context(), shopID, c.Param("id"), req)
	if err != nil {
		h.handleError(c, err)
		return
	}
	response.OK(c, zone)
}

func (h *Handler) DeleteZone(c *gin.Context) {
	shopID := c.GetString("shop_id")
	if err := h.svc.DeleteZone(c.Request.Context(), shopID, c.Param("id")); err != nil {
		h.handleError(c, err)
		return
	}
	response.NoContent(c)
}


func (h *Handler) CreateRate(c *gin.Context) {
	shopID := c.GetString("shop_id")
	var req CreateRateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	rate, err := h.svc.CreateRate(c.Request.Context(), shopID, req)
	if err != nil {
		h.handleError(c, err)
		return
	}
	response.Created(c, rate)
}

func (h *Handler) GetRate(c *gin.Context) {
	shopID := c.GetString("shop_id")
	rate, err := h.svc.GetRate(c.Request.Context(), shopID, c.Param("id"))
	if err != nil {
		h.handleError(c, err)
		return
	}
	response.OK(c, rate)
}

func (h *Handler) ListRates(c *gin.Context) {
	shopID := c.GetString("shop_id")
	rates, err := h.svc.ListRates(c.Request.Context(), shopID)
	if err != nil {
		response.InternalError(c)
		return
	}
	response.OK(c, rates)
}

func (h *Handler) ListRatesByZone(c *gin.Context) {
	shopID := c.GetString("shop_id")
	rates, err := h.svc.ListRatesByZone(c.Request.Context(), shopID, c.Param("id"))
	if err != nil {
		h.handleError(c, err)
		return
	}
	response.OK(c, rates)
}

func (h *Handler) UpdateRate(c *gin.Context) {
	shopID := c.GetString("shop_id")
	var req UpdateRateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	rate, err := h.svc.UpdateRate(c.Request.Context(), shopID, c.Param("id"), req)
	if err != nil {
		h.handleError(c, err)
		return
	}
	response.OK(c, rate)
}

func (h *Handler) DeleteRate(c *gin.Context) {
	shopID := c.GetString("shop_id")
	if err := h.svc.DeleteRate(c.Request.Context(), shopID, c.Param("id")); err != nil {
		h.handleError(c, err)
		return
	}
	response.NoContent(c)
}


func (h *Handler) handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrNoRate):
		response.UnprocessableEntity(c, err.Error())
	default:
		response.HandleError(c, err)
	}
}
