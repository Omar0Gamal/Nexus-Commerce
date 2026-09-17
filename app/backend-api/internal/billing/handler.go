package billing

import (
	"errors"
	"fmt"
	"strconv"

	"backend-api/internal/paymob"
	sharedaudit "backend-api/internal/shared/audit"
	"backend-api/internal/shared/response"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	svc          *Service
	paymobClient *paymob.Client
}

func NewHandler(svc *Service, paymobClient *paymob.Client) *Handler {
	return &Handler{svc: svc, paymobClient: paymobClient}
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
		billing.POST("/portal", h.GeneratePortalLink)
	}
}

func (h *Handler) ListPlans(c *gin.Context) {
	plans, err := h.svc.ListPlans(c.Request.Context())
	if err != nil {
		response.InternalError(c)
		return
	}
	response.OK(c, plans)
}

func (h *Handler) GetPlan(c *gin.Context) {
	plan, err := h.svc.GetPlan(c.Request.Context(), c.Param("id"))
	if err != nil {
		h.handleError(c, err)
		return
	}
	response.OK(c, plan)
}

func (h *Handler) GetBillingStatus(c *gin.Context) {
	shopID := c.GetString("shop_id")

	status, err := h.svc.GetShopBillingStatus(c.Request.Context(), shopID)
	if err != nil {
		h.handleError(c, err)
		return
	}
	response.OK(c, status)
}

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

func (h *Handler) GeneratePortalLink(c *gin.Context) {
	shopID := c.GetString("shop_id")

	status, err := h.svc.GetShopBillingStatus(c.Request.Context(), shopID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	token, err := h.paymobClient.Authenticate()
	if err != nil {
		h.handleError(c, err)
		return
	}

	var monthlyPriceFloat float64
	fmt.Sscanf(status.Plan.MonthlyPrice, "%f", &monthlyPriceFloat)
	amountCents := int64(monthlyPriceFloat * 100)

	paymobOrderID, err := h.paymobClient.RegisterOrder(token, amountCents, "EGP", "")
	if err != nil {
		h.handleError(c, err)
		return
	}

	billingData := paymob.BillingData{
		FirstName: "Admin",
		LastName:  "User",
		Email:     "admin@nexuscommerce.io",
		Phone:     "+201000000000",
		City:      "Cairo",
		Country:   "EGY",
		Street:    "NA",
		Building:  "NA",
		Floor:     "NA",
		Apartment: "NA",
		State:     "NA",
	}

	paymentKey, err := h.paymobClient.GetPaymentKey(token, paymobOrderID, amountCents, "EGP", billingData, "", 0)
	if err != nil {
		h.handleError(c, err)
		return
	}

	url := h.paymobClient.IFrameURL(paymentKey)

	response.OK(c, gin.H{
		"url": url,
	})
}
