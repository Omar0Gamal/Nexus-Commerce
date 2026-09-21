package returns

type CreateReturnRequest struct {
	OrderID    string `json:"order_id" binding:"required"`
	CustomerID string `json:"customer_id,omitempty"`
	Reason     string `json:"reason" binding:"required,min=5"`
}

type UpdateReturnRequest struct {
	Status string  `json:"status" binding:"required"`
	Notes  *string `json:"notes,omitempty"`
}

type ProcessRefundRequest struct {
	AmountCents  int64    `json:"amount_cents,omitempty"`
	RefundAmount *float64 `json:"refund_amount,omitempty"`
}

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
