package returns

// CreateReturnRequest holds the fields required to open a return request.
type CreateReturnRequest struct {
	OrderID    string `json:"order_id" binding:"required"`
	CustomerID string `json:"customer_id,omitempty"`
	Reason     string `json:"reason" binding:"required,min=5"`
}

// UpdateReturnRequest holds the fields for updating a return request status.
type UpdateReturnRequest struct {
	Status string  `json:"status" binding:"required"`
	Notes  *string `json:"notes,omitempty"`
}

// ProcessRefundRequest holds the fields for processing a refund.
type ProcessRefundRequest struct {
	// AmountCents is the refund amount in the smallest currency unit (e.g. piasters).
	// If zero, the full order total is used by the Paymob flow.
	AmountCents  int64    `json:"amount_cents,omitempty"`
	RefundAmount *float64 `json:"refund_amount,omitempty"` // in currency units (e.g. EGP)
}

// ReturnResponse is the JSON-serialisable representation of a return request.
type ReturnResponse struct {
	ID             string  `json:"id"`
	ShopID         string  `json:"shop_id"`
	OrderID        string  `json:"order_id"`
	CustomerID     *string `json:"customer_id,omitempty"`
	Reason         string  `json:"reason"`
	Status         string  `json:"status"`
	PaymobRefundID *int64  `json:"paymob_refund_id,omitempty"`
	RefundAmount   *string `json:"refund_amount,omitempty"`
	Notes          *string `json:"notes,omitempty"`
	CreatedAt      string  `json:"created_at"`
	UpdatedAt      string  `json:"updated_at"`
}
