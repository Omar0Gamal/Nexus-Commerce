package customers

import "backend-api/internal/orders"

// CustomerSummary is a lightweight customer row used in list views.
type CustomerSummary struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	FirstName string `json:"first_name,omitempty"`
	LastName  string `json:"last_name,omitempty"`
	Phone     string `json:"phone,omitempty"`
	CreatedAt string `json:"created_at"`
}

// CustomerDetail is the full customer profile including order history.
type CustomerDetail struct {
	CustomerSummary
	TotalOrders int                    `json:"total_orders"`
	TotalSpent  string                 `json:"total_spent"`
	Orders      []orders.OrderResponse `json:"orders"`
}

// ListCustomersParams holds query parameters for listing customers.
type ListCustomersParams struct {
	Search  string `form:"search"`
	Page    int    `form:"page,default=1"`
	PerPage int    `form:"per_page,default=20"`
}
