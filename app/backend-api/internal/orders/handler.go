package orders

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

// GET routes are accessible to authenticated customers (for order history);
// write routes require authenticated staff.
//
//	/api/v1/orders              GET, POST
//	/api/v1/orders/:id          GET, DELETE
//	/api/v1/orders/:id/status   PATCH
//	/api/v1/stats               GET  ← dashboard aggregate stats
//	/api/v1/activity            GET  ← recent activity feed
func (h *Handler) RegisterRoutes(rg *gin.RouterGroup, requireAuth, requireStaff gin.HandlerFunc) {
	ord := rg.Group("/orders")
	{
		ord.GET("", requireAuth, requireStaff, h.ListOrders)
		ord.POST("", requireAuth, requireStaff, h.CreateOrder)
		ord.GET("/:id", requireAuth, requireStaff, h.GetOrder)
		ord.DELETE("/:id", requireAuth, requireStaff, h.DeleteOrder)
		ord.PATCH("/:id/status", requireAuth, requireStaff, h.UpdateOrderStatus)
	}

	// Dashboard stats
	rg.GET("/stats", requireAuth, requireStaff, h.GetStats)
	// Recent activity feed
	rg.GET("/activity", requireAuth, requireStaff, h.GetActivity)
}


func (h *Handler) CreateOrder(c *gin.Context) {
	shopID := c.GetString("shop_id")

	var req CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	order, err := h.svc.CreateOrder(c.Request.Context(), shopID, req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response.Created(c, order)
}

func (h *Handler) GetOrder(c *gin.Context) {
	shopID := c.GetString("shop_id")
	orderID := c.Param("id")

	order, err := h.svc.GetOrder(c.Request.Context(), shopID, orderID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response.OK(c, order)
}

func (h *Handler) ListOrders(c *gin.Context) {
	shopID := c.GetString("shop_id")

	var params ListOrdersParams
	if err := c.ShouldBindQuery(&params); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	orders, total, err := h.svc.ListOrders(c.Request.Context(), shopID, params)
	if err != nil {
		h.handleError(c, err)
		return
	}

	if params.PerPage < 1 {
		params.PerPage = 20
	}
	totalPages := total / int64(params.PerPage)
	if total%int64(params.PerPage) != 0 {
		totalPages++
	}

	response.OKWithMeta(c, orders, response.PaginationMeta{
		Page:       params.Page,
		PerPage:    params.PerPage,
		TotalItems: total,
		TotalPages: totalPages,
	})
}

func (h *Handler) UpdateOrderStatus(c *gin.Context) {
	shopID := c.GetString("shop_id")
	orderID := c.Param("id")

	var req UpdateOrderStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	order, err := h.svc.UpdateOrderStatus(c.Request.Context(), shopID, orderID, req.Status)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response.OK(c, order)
}

func (h *Handler) DeleteOrder(c *gin.Context) {
	shopID := c.GetString("shop_id")
	orderID := c.Param("id")

	if err := h.svc.DeleteOrder(c.Request.Context(), shopID, orderID); err != nil {
		h.handleError(c, err)
		return
	}

	response.NoContent(c)
}

//
//	/api/v1/orders/my          GET — customer's own orders
//	/api/v1/orders/my/:id      GET — single order (ownership verified)
func (h *Handler) RegisterCustomerRoutes(rg *gin.RouterGroup, requireAuth, requireCustomer gin.HandlerFunc) {
	my := rg.Group("/orders/my")
	{
		my.GET("", requireAuth, requireCustomer, h.ListMyOrders)
		my.GET("/:id", requireAuth, requireCustomer, h.GetMyOrder)
	}
}

func (h *Handler) ListMyOrders(c *gin.Context) {
	shopID := c.GetString("shop_id")
	customerID := c.GetString("user_id") // set by RequireAuth from JWT UserID claim

	orders, err := h.svc.ListMyOrders(c.Request.Context(), shopID, customerID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response.OK(c, orders)
}

func (h *Handler) GetMyOrder(c *gin.Context) {
	shopID := c.GetString("shop_id")
	customerID := c.GetString("user_id")
	orderID := c.Param("id")

	order, err := h.svc.GetMyOrder(c.Request.Context(), shopID, customerID, orderID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response.OK(c, order)
}

func (h *Handler) GetStats(c *gin.Context) {
	shopID := c.GetString("shop_id")
	stats, err := h.svc.GetDashboardStats(c.Request.Context(), shopID)
	if err != nil {
		response.InternalError(c)
		return
	}
	response.OK(c, stats)
}

func (h *Handler) GetActivity(c *gin.Context) {
	shopID := c.GetString("shop_id")
	items, err := h.svc.ListRecentActivity(c.Request.Context(), shopID, 10)
	if err != nil {
		response.InternalError(c)
		return
	}
	response.OK(c, items)
}

func (h *Handler) handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrNotFound):
		response.NotFound(c, "Order not found")
	case errors.Is(err, ErrInvalidUUID):
		response.BadRequest(c, "Invalid ID format")
	case errors.Is(err, ErrBadStatus):
		response.BadRequest(c, "Invalid order status. Must be one of: pending, paid, processing, shipped, completed, cancelled, refunded")
	case errors.Is(err, ErrInsufficientStock):
		response.UnprocessableEntity(c, err.Error())
	default:
		response.InternalError(c)
	}
}
