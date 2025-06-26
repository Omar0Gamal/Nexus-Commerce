package orders

import "encoding/json"


// ShippingAddress is the JSONB structure for the order shipping address.
type ShippingAddress struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Address1  string `json:"address1"`
	Address2  string `json:"address2,omitempty"`
	City      string `json:"city"`
	State     string `json:"state,omitempty"`
	Country   string `json:"country"`
	ZipCode   string `json:"zip_code,omitempty"`
	Phone     string `json:"phone,omitempty"`
}

// OrderItemInput is a single line item for creating an order.
type OrderItemInput struct {
	ProductID string  `json:"product_id" binding:"required"`
	VariantID string  `json:"variant_id,omitempty"`
	Name      string  `json:"name" binding:"required"`
	Price     float64 `json:"price" binding:"required,gt=0"`
	Quantity  int32   `json:"quantity" binding:"required,gt=0"`
}

type CreateOrderRequest struct {
	CustomerID      string           `json:"customer_id,omitempty"`
	CustomerEmail   string           `json:"customer_email,omitempty"` // used to send confirmation email
	CouponCode      string           `json:"coupon_code,omitempty"`
	WeightKg        float64          `json:"weight_kg,omitempty"` // total order weight for weight-based shipping
	ShippingAddress ShippingAddress  `json:"shipping_address" binding:"required"`
	Items           []OrderItemInput `json:"items" binding:"required,min=1"`
}

type UpdateOrderStatusRequest struct {
	Status string `json:"status" binding:"required"`
}


// OrderResponse is the JSON response for a single order.
type OrderResponse struct {
	ID              string              `json:"id"`
	ShopID          string              `json:"shop_id"`
	CustomerID      *string             `json:"customer_id,omitempty"`
	OrderNumber     int32               `json:"order_number"`
	TotalPrice      string              `json:"total_price"`
	DiscountAmount  string              `json:"discount_amount"`
	ShippingFee     string              `json:"shipping_fee"`
	CouponID        *string             `json:"coupon_id,omitempty"`
	Status          string              `json:"status"`
	ShippingAddress json.RawMessage     `json:"shipping_address"`
	CreatedAt       string              `json:"created_at"`
	Items           []OrderItemResponse `json:"items,omitempty"`
}

// OrderItemResponse is the JSON response for a single order item.
type OrderItemResponse struct {
	ID        string  `json:"id"`
	OrderID   string  `json:"order_id"`
	ProductID *string `json:"product_id,omitempty"`
	VariantID *string `json:"variant_id,omitempty"`
	Name      string  `json:"name"`
	Price     string  `json:"price"`
	Quantity  int32   `json:"quantity"`
}

// ListOrdersParams holds pagination and filter parameters.
type ListOrdersParams struct {
	Page       int    `form:"page,default=1"`
	PerPage    int    `form:"per_page,default=20"`
	Status     string `form:"status"`
	CustomerID string `form:"customer_id"`
}

// DashboardStats is the response for GET /api/v1/stats.
type DashboardStats struct {
	TotalOrders    int64  `json:"total_orders"`
	TotalRevenue   string `json:"total_revenue"`
	PendingOrders  int64  `json:"pending_orders"`
	TotalCustomers int64  `json:"total_customers"`
}

// ActivityItem is a single recent activity event from the audit log.
type ActivityItem struct {
	ID           string `json:"id"`
	Action       string `json:"action"`
	ResourceType string `json:"resource_type"`
	ActorName    string `json:"actor_name"`
	Description  string `json:"description"`
	CreatedAt    string `json:"created_at"`
}
