package orders

import (
	"context"
	"fmt"
	"backend-api/internal/paymob"
	"errors"

	"backend-api/internal/shared/response"

	"github.com/gin-gonic/gin"
)

type CartService interface {
	GetCartForCheckout(ctx context.Context, shopID, identity string) ([]OrderItemInput, error)
	ClearCart(ctx context.Context, shopID, identity string) error
}

type Handler struct {
	svc          *Service
	cartSvc      CartService
	paymobClient *paymob.Client
}

func NewHandler(svc *Service, cartSvc CartService, paymobClient *paymob.Client) *Handler {
	return &Handler{svc: svc, cartSvc: cartSvc, paymobClient: paymobClient}
}

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup, requireAuth, requireStaff gin.HandlerFunc) {
	ord := rg.Group("/orders")
	{
		ord.GET("", requireAuth, requireStaff, h.ListOrders)
		ord.POST("", requireAuth, requireStaff, h.CreateOrder)
		ord.GET("/:id", requireAuth, requireStaff, h.GetOrder)
		ord.DELETE("/:id", requireAuth, requireStaff, h.DeleteOrder)
		ord.PATCH("/:id/status", requireAuth, requireStaff, h.UpdateOrderStatus)
		ord.POST("/bulk-update-status", requireAuth, requireStaff, h.BulkUpdateOrderStatus)
	}

	rg.GET("/stats", requireAuth, requireStaff, h.GetStats)
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

type BulkUpdateOrderStatusRequest struct {
	OrderIDs []string `json:"order_ids" binding:"required,min=1"`
	Status   string   `json:"status" binding:"required"`
}

func (h *Handler) BulkUpdateOrderStatus(c *gin.Context) {
	shopID := c.GetString("shop_id")

	var req BulkUpdateOrderStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	updatedCount := 0
	for _, id := range req.OrderIDs {
		if _, err := h.svc.UpdateOrderStatus(c.Request.Context(), shopID, id, req.Status); err == nil {
			updatedCount++
		}
	}

	c.JSON(200, gin.H{"updated_count": updatedCount})
}

func (h *Handler) RegisterCustomerRoutes(rg *gin.RouterGroup, requireAuth, requireCustomer gin.HandlerFunc) {
	my := rg.Group("/orders/my")
	{
		my.GET("", requireAuth, requireCustomer, h.ListMyOrders)
		my.GET("/:id", requireAuth, requireCustomer, h.GetMyOrder)
	}
}

func (h *Handler) ListMyOrders(c *gin.Context) {
	shopID := c.GetString("shop_id")
	customerID := c.GetString("user_id")

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

// RegisterCheckoutRoutes registers the checkout intent route, allowing optional auth for guest checkout.
func (h *Handler) RegisterCheckoutRoutes(rg *gin.RouterGroup, optionalAuth gin.HandlerFunc) {
	rg.POST("/checkout/intent", optionalAuth, h.CheckoutIntent)
}

type CheckoutIntentRequest struct {
	CartToken       string          `json:"cart_token"`
	Email           string          `json:"email"`
	PaymentMethod   string          `json:"payment_method"`
	ShippingAddress ShippingAddress `json:"shipping_address" binding:"required"`
}

func (h *Handler) CheckoutIntent(c *gin.Context) {
	shopID := c.GetString("shop_id")
	customerID := c.GetString("user_id")
	
	var req CheckoutIntentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	identity := "guest:" + req.CartToken
	if customerID != "" {
		identity = "customer:" + customerID
	}

	items, err := h.cartSvc.GetCartForCheckout(c.Request.Context(), shopID, identity)
	if err != nil {
		h.handleError(c, err)
		return
	}

	req.ShippingAddress.Email = req.Email
	orderReq := CreateOrderRequest{
		CustomerID:      customerID,
		CustomerEmail:   req.Email,
		ShippingAddress: req.ShippingAddress,
		Items:           items,
	}
	
	order, err := h.svc.CreateOrder(c.Request.Context(), shopID, orderReq)
	if err != nil {
		h.handleError(c, err)
		return
	}

	_ = h.cartSvc.ClearCart(c.Request.Context(), shopID, identity)

	if req.PaymentMethod == "test" || req.PaymentMethod == "cod" {
		if req.PaymentMethod == "test" {
			_, _ = h.svc.UpdateOrderStatus(c.Request.Context(), shopID, order.ID, "paid")
		}
		response.OK(c, gin.H{
			"intent_id":   order.ID,
			"payment_url": fmt.Sprintf("/order-confirmation?order_id=%s&simulated=true", order.ID),
		})
		return
	}

	token, err := h.paymobClient.Authenticate()
	if err != nil {
		response.OK(c, gin.H{
			"intent_id":   order.ID,
			"payment_url": fmt.Sprintf("/order-confirmation?order_id=%s&simulated=true", order.ID),
		})
		return
	}

	var amountCents int64
	var parsedPrice float64
	fmt.Sscanf(order.TotalPrice, "%f", &parsedPrice)
	amountCents = int64(parsedPrice * 100)

	paymobOrderID, err := h.paymobClient.RegisterOrder(token, amountCents, "EGP", order.ID)
	if err != nil {
		response.OK(c, gin.H{
			"intent_id":   order.ID,
			"payment_url": fmt.Sprintf("/order-confirmation?order_id=%s&simulated=true", order.ID),
		})
		return
	}

	billingData := paymob.BillingData{
		FirstName: req.ShippingAddress.FirstName,
		LastName:  req.ShippingAddress.LastName,
		Email:     req.Email,
		Phone:     req.ShippingAddress.Phone,
		City:      req.ShippingAddress.City,
		Country:   req.ShippingAddress.Country,
		Street:    req.ShippingAddress.Address1,
		Building:  "NA",
		Floor:     "NA",
		Apartment: "NA",
		State:     req.ShippingAddress.State,
	}
	if req.Email == "" {
		billingData.Email = "no-reply@example.com"
	}
	if req.ShippingAddress.Phone == "" {
		billingData.Phone = "+201000000000"
	}
	if req.ShippingAddress.State == "" {
		billingData.State = "NA"
	}

	paymentKey, err := h.paymobClient.GetPaymentKey(token, paymobOrderID, amountCents, "EGP", billingData, "", 0)
	if err != nil {
		response.OK(c, gin.H{
			"intent_id":   order.ID,
			"payment_url": fmt.Sprintf("/order-confirmation?order_id=%s&simulated=true", order.ID),
		})
		return
	}

	paymentURL := h.paymobClient.IFrameURL(paymentKey)

	response.OK(c, gin.H{
		"intent_id":   order.ID,
		"payment_url": paymentURL,
	})
}

func (h *Handler) RegisterStorefrontRoutes(rg *gin.RouterGroup) {
	rg.GET("/storefront/orders/track", h.TrackOrder)
}

func (h *Handler) TrackOrder(c *gin.Context) {
	shopID := c.GetString("shop_id")
	if shopID == "" {
		response.BadRequest(c, "shop_id missing")
		return
	}

	email := c.Query("email")
	trackingNumberStr := c.Query("tracking_number")

	if email == "" || trackingNumberStr == "" {
		response.BadRequest(c, "email and tracking_number are required")
		return
	}

	var trackingNumber int32
	_, err := fmt.Sscanf(trackingNumberStr, "%d", &trackingNumber)
	if err != nil {
		response.BadRequest(c, "invalid tracking number format")
		return
	}

	order, err := h.svc.TrackOrder(c.Request.Context(), shopID, trackingNumber, email)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response.OK(c, order)
}
