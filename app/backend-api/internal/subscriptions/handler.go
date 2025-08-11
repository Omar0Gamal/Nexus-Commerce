package subscriptions

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

func (h *Handler) RegisterStaffRoutes(rg *gin.RouterGroup, requireAuth, requireStaff gin.HandlerFunc) {
	s := rg.Group("/subscriptions", requireAuth, requireStaff)
	{
		s.POST("/plans", h.CreatePlan)
		s.GET("/plans", h.ListPlans)
		s.GET("/plans/:id", h.GetPlan)
		s.PATCH("/plans/:id", h.UpdatePlan)
		s.DELETE("/plans/:id", h.DeletePlan)
		s.GET("", h.ListShopSubscriptions)
		s.GET("/:id", h.GetSubscription)
	}
}

func (h *Handler) RegisterCustomerRoutes(rg *gin.RouterGroup, requireAuth, requireCustomer gin.HandlerFunc) {
	s := rg.Group("/my-subscriptions", requireAuth, requireCustomer)
	{
		s.GET("/plans", h.ListPlans)
		s.POST("", h.Subscribe)
		s.GET("", h.ListMySubscriptions)
		s.GET("/:id", h.GetMySubscription)
		s.POST("/:id/cancel", h.Cancel)
		s.POST("/:id/pause", h.Pause)
		s.POST("/:id/resume", h.Resume)
	}
}


func (h *Handler) CreatePlan(c *gin.Context) {
	shopID := c.GetString("shop_id")
	var req CreatePlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	plan, err := h.svc.CreatePlan(c.Request.Context(), shopID, req)
	if err != nil {
		h.handleError(c, err)
		return
	}
	response.Created(c, plan)
}

func (h *Handler) GetPlan(c *gin.Context) {
	shopID := c.GetString("shop_id")
	plan, err := h.svc.GetPlan(c.Request.Context(), shopID, c.Param("id"))
	if err != nil {
		h.handleError(c, err)
		return
	}
	response.OK(c, plan)
}

func (h *Handler) ListPlans(c *gin.Context) {
	shopID := c.GetString("shop_id")
	plans, err := h.svc.ListPlans(c.Request.Context(), shopID)
	if err != nil {
		h.handleError(c, err)
		return
	}
	response.OK(c, plans)
}

func (h *Handler) UpdatePlan(c *gin.Context) {
	shopID := c.GetString("shop_id")
	var req UpdatePlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	plan, err := h.svc.UpdatePlan(c.Request.Context(), shopID, c.Param("id"), req)
	if err != nil {
		h.handleError(c, err)
		return
	}
	response.OK(c, plan)
}

func (h *Handler) DeletePlan(c *gin.Context) {
	shopID := c.GetString("shop_id")
	if err := h.svc.DeletePlan(c.Request.Context(), shopID, c.Param("id")); err != nil {
		h.handleError(c, err)
		return
	}
	response.NoContent(c)
}

func (h *Handler) GetSubscription(c *gin.Context) {
	shopID := c.GetString("shop_id")
	sub, err := h.svc.GetSubscription(c.Request.Context(), shopID, c.Param("id"))
	if err != nil {
		h.handleError(c, err)
		return
	}
	response.OK(c, sub)
}

func (h *Handler) ListShopSubscriptions(c *gin.Context) {
	shopID := c.GetString("shop_id")
	var q ListSubscriptionsQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	subs, err := h.svc.ListShopSubscriptions(c.Request.Context(), shopID, q.Page, q.PerPage)
	if err != nil {
		h.handleError(c, err)
		return
	}
	response.OK(c, subs)
}


func (h *Handler) Subscribe(c *gin.Context) {
	shopID := c.GetString("shop_id")
	customerID := c.GetString("customer_id")
	var req SubscribeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	sub, err := h.svc.Subscribe(c.Request.Context(), shopID, customerID, req)
	if err != nil {
		h.handleError(c, err)
		return
	}
	response.Created(c, sub)
}

func (h *Handler) ListMySubscriptions(c *gin.Context) {
	shopID := c.GetString("shop_id")
	customerID := c.GetString("customer_id")
	subs, err := h.svc.ListCustomerSubscriptions(c.Request.Context(), shopID, customerID)
	if err != nil {
		h.handleError(c, err)
		return
	}
	response.OK(c, subs)
}

func (h *Handler) GetMySubscription(c *gin.Context) {
	shopID := c.GetString("shop_id")
	sub, err := h.svc.GetSubscription(c.Request.Context(), shopID, c.Param("id"))
	if err != nil {
		h.handleError(c, err)
		return
	}
	response.OK(c, sub)
}

func (h *Handler) Cancel(c *gin.Context) {
	shopID := c.GetString("shop_id")
	sub, err := h.svc.CancelSubscription(c.Request.Context(), shopID, c.Param("id"))
	if err != nil {
		h.handleError(c, err)
		return
	}
	response.OK(c, sub)
}

func (h *Handler) Pause(c *gin.Context) {
	shopID := c.GetString("shop_id")
	sub, err := h.svc.PauseSubscription(c.Request.Context(), shopID, c.Param("id"))
	if err != nil {
		h.handleError(c, err)
		return
	}
	response.OK(c, sub)
}

func (h *Handler) Resume(c *gin.Context) {
	shopID := c.GetString("shop_id")
	sub, err := h.svc.ResumeSubscription(c.Request.Context(), shopID, c.Param("id"))
	if err != nil {
		h.handleError(c, err)
		return
	}
	response.OK(c, sub)
}


func (h *Handler) handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrNotFound):
		response.NotFound(c, "Subscription not found")
	case errors.Is(err, ErrInvalidUUID):
		response.BadRequest(c, "Invalid ID format")
	case errors.Is(err, ErrConflict):
		response.Conflict(c, err.Error())
	default:
		response.InternalError(c)
	}
}
