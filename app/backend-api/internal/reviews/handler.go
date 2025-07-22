package reviews

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

func (h *Handler) handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrAlreadyReviewed):
		response.Conflict(c, "you have already reviewed this product")
	default:
		response.HandleError(c, err)
	}
}

func (h *Handler) RegisterCustomerRoutes(rg *gin.RouterGroup, requireAuth, requireCustomer gin.HandlerFunc, helpfulRL ...gin.HandlerFunc) {
	rg.GET("/products/:id/reviews", h.ListReviews)
	rg.GET("/products/:id/reviews/summary", h.GetSummary)
	helpfulChain := append(helpfulRL, h.MarkHelpful)
	rg.POST("/products/:id/reviews/:reviewId/helpful", helpfulChain...)
	rg.POST("/products/:id/reviews", requireAuth, requireCustomer, h.SubmitReview)
}

func (h *Handler) RegisterStaffRoutes(rg *gin.RouterGroup, requireAuth, requireStaff gin.HandlerFunc) {
	grp := rg.Group("/catalog/reviews", requireAuth, requireStaff)
	{
		grp.GET("", h.ListPending)
		grp.PATCH("/:id", h.Moderate)
	}
}


func (h *Handler) SubmitReview(c *gin.Context) {
	shopID := c.GetString("shop_id")
	customerID := c.GetString("user_id")
	productID := c.Param("id")

	var req SubmitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	req.CustomerID = customerID
	req.ProductID = productID

	// moderationEnabled could come from shop settings; default to true (safe)
	review, err := h.svc.Submit(c.Request.Context(), shopID, req, true)
	if err != nil {
		h.handleError(c, err)
		return
	}
	response.Created(c, review)
}

func (h *Handler) ListReviews(c *gin.Context) {
	shopID := c.GetString("shop_id")
	productID := c.Param("id")

	limit := int32(20)
	offset := int32(0)
	if v := c.Query("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 100 {
			limit = int32(n)
		}
	}
	if v := c.Query("page"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 1 {
			offset = int32(n-1) * limit
		}
	}

	items, err := h.svc.List(c.Request.Context(), shopID, productID, limit, offset)
	if err != nil {
		response.InternalError(c)
		return
	}
	response.OK(c, items)
}

func (h *Handler) GetSummary(c *gin.Context) {
	shopID := c.GetString("shop_id")
	productID := c.Param("id")

	summary, err := h.svc.Summary(c.Request.Context(), shopID, productID)
	if err != nil {
		response.InternalError(c)
		return
	}
	response.OK(c, summary)
}

func (h *Handler) MarkHelpful(c *gin.Context) {
	shopID := c.GetString("shop_id")
	reviewID := c.Param("reviewId")

	if err := h.svc.MarkHelpful(c.Request.Context(), shopID, reviewID); err != nil {
		h.handleError(c, err)
		return
	}
	response.NoContent(c)
}

func (h *Handler) ListPending(c *gin.Context) {
	shopID := c.GetString("shop_id")

	limit := int32(20)
	offset := int32(0)
	if v := c.Query("page"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 1 {
			offset = int32(n-1) * limit
		}
	}

	items, err := h.svc.ListPending(c.Request.Context(), shopID, limit, offset)
	if err != nil {
		response.InternalError(c)
		return
	}
	response.OK(c, items)
}

func (h *Handler) Moderate(c *gin.Context) {
	shopID := c.GetString("shop_id")
	reviewID := c.Param("id")

	var req ModerateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	updated, err := h.svc.Moderate(c.Request.Context(), shopID, reviewID, req.Status)
	if err != nil {
		h.handleError(c, err)
		return
	}
	response.OK(c, updated)
}
