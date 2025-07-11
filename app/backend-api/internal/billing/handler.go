package billing

import (
	"errors"
	"strconv"

	sharedaudit "backend-api/internal/shared/audit"
	"backend-api/internal/shared/response"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) RegisterPublicRoutes(rg *gin.RouterGroup) {
	rg.GET("/plans", h.ListPlans)
	rg.GET("/plans/:id", h.GetPlan)
}

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup, requireAuth gin.HandlerFunc, requireStaff gin.HandlerFunc) {
	billing := rg.Group("/billing", requireAuth, requireStaff)
	{
		billing.GET("/status", h.GetBillingStatus)
		billing.PATCH("/plan", h.ChangePlan)
		billing.GET("/invoices", h.ListInvoices)
		billing.GET("/onboarding", h.GetOnboardingChecklist)
	}
}

// ListPlans returns all subscription plans (public).
func (h *Handler) ListPlans(c *gin.Context) {
	plans, err := h.svc.ListPlans(c.Request.Context())
	if err != nil {
		response.InternalError(c)
		return
	}
	response.OK(c, plans)
}

// GetPlan returns a single plan by ID (public).
func (h *Handler) GetPlan(c *gin.Context) {
	plan, err := h.svc.GetPlan(c.Request.Context(), c.Param("id"))
	if err != nil {
		h.handleError(c, err)
		return
	}
	response.OK(c, plan)
}

// GetBillingStatus returns current plan + billing state for the authenticated shop.
func (h *Handler) GetBillingStatus(c *gin.Context) {
	shopID := c.GetString("shop_id")

	status, err := h.svc.GetShopBillingStatus(c.Request.Context(), shopID)
	if err != nil {
		h.handleError(c, err)
		return
	}
	response.OK(c, status)
}

// ChangePlan updates the shop's subscription plan.
func (h *Handler) ChangePlan(c *gin.Context) {
	shopID := c.GetString("shop_id")

	var req ChangePlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	status, err := h.svc.ChangePlan(c.Request.Context(), shopID, req.PlanID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	sharedaudit.Write(h.svc.Q(), sharedaudit.Entry{
		ShopID:       shopID,
		ActorUserID:  c.GetString("user_id"),
		ActorName:    c.GetString("user_name"),
		Action:       "billing.plan_change",
		ResourceType: "plan",
		ResourceID:   req.PlanID,
		Changes:      map[string]any{"plan_id": req.PlanID},
	})

	response.OK(c, status)
}

// ListInvoices returns the invoice history for the authenticated shop.
func (h *Handler) ListInvoices(c *gin.Context) {
	shopID := c.GetString("shop_id")

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	invoices, err := h.svc.ListInvoices(c.Request.Context(), shopID, page, pageSize)
	if err != nil {
		h.handleError(c, err)
		return
	}
	response.OK(c, invoices)
}

// GetOnboardingChecklist returns the shop's onboarding completion steps.
func (h *Handler) GetOnboardingChecklist(c *gin.Context) {
	shopID := c.GetString("shop_id")

	checklist, err := h.svc.GetOnboardingChecklist(c.Request.Context(), shopID)
	if err != nil {
		h.handleError(c, err)
		return
	}
	response.OK(c, checklist)
}

func (h *Handler) handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrPlanNotFound):
		response.NotFound(c, "plan not found")
	case errors.Is(err, ErrShopNotFound):
		response.NotFound(c, "shop not found")
	case errors.Is(err, ErrInvalidPlan):
		response.BadRequest(c, "invalid plan ID")
	default:
		response.InternalError(c)
	}
}
