package webhooks

import (
	"errors"

	"backend-api/internal/shared/response"

	"github.com/gin-gonic/gin"
)

// Handler exposes webhook management endpoints.
type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup, authMW, requireStaff gin.HandlerFunc) {
	wh := rg.Group("/webhooks")
	wh.Use(authMW, requireStaff)
	{
		wh.GET("", h.List)
		wh.POST("", h.Create)
		wh.GET("/:id", h.Get)
		wh.PATCH("/:id", h.Update)
		wh.POST("/:id/toggle", h.Toggle)
		wh.DELETE("/:id", h.Delete)
	}
}


func (h *Handler) List(c *gin.Context) {
	shopID := c.GetString("shop_id")
	if shopID == "" {
		response.Unauthorized(c, "missing shop context")
		return
	}

	items, err := h.svc.ListWebhooks(c.Request.Context(), shopID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response.OK(c, items)
}

func (h *Handler) Create(c *gin.Context) {
	shopID := c.GetString("shop_id")
	if shopID == "" {
		response.Unauthorized(c, "missing shop context")
		return
	}

	var req CreateWebhookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	wh, err := h.svc.CreateWebhook(c.Request.Context(), shopID, req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response.Created(c, wh)
}

func (h *Handler) Get(c *gin.Context) {
	shopID := c.GetString("shop_id")
	if shopID == "" {
		response.Unauthorized(c, "missing shop context")
		return
	}

	wh, err := h.svc.GetWebhook(c.Request.Context(), shopID, c.Param("id"))
	if err != nil {
		h.handleError(c, err)
		return
	}

	response.OK(c, wh)
}

func (h *Handler) Update(c *gin.Context) {
	shopID := c.GetString("shop_id")
	if shopID == "" {
		response.Unauthorized(c, "missing shop context")
		return
	}

	var req UpdateWebhookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	wh, err := h.svc.UpdateWebhook(c.Request.Context(), shopID, c.Param("id"), req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response.OK(c, wh)
}

func (h *Handler) Toggle(c *gin.Context) {
	shopID := c.GetString("shop_id")
	if shopID == "" {
		response.Unauthorized(c, "missing shop context")
		return
	}

	wh, err := h.svc.ToggleWebhook(c.Request.Context(), shopID, c.Param("id"))
	if err != nil {
		h.handleError(c, err)
		return
	}

	response.OK(c, wh)
}

func (h *Handler) Delete(c *gin.Context) {
	shopID := c.GetString("shop_id")
	if shopID == "" {
		response.Unauthorized(c, "missing shop context")
		return
	}

	if err := h.svc.DeleteWebhook(c.Request.Context(), shopID, c.Param("id")); err != nil {
		h.handleError(c, err)
		return
	}

	response.NoContent(c)
}


func (h *Handler) handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrBadTopic):
		response.BadRequest(c, "invalid topic; allowed: order.created, order.status_changed, product.low_stock")
	default:
		response.HandleError(c, err)
	}
}
