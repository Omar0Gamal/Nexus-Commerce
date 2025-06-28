package cart


// CartItem represents a single item in the cart.
type CartItem struct {
	ProductID   string  `json:"product_id"`
	VariantID   string  `json:"variant_id,omitempty"`
	Title       string  `json:"title"`
	Price       float64 `json:"price"`
	Quantity    int     `json:"quantity"`
	ImageURL    string  `json:"image_url,omitempty"`
	Slug        string  `json:"slug,omitempty"`
	VariantName string  `json:"variant_name,omitempty"`
}

// Cart is the full shopping cart.
type Cart struct {
	Items      []CartItem `json:"items"`
	TotalItems int        `json:"total_items"`
	TotalPrice float64    `json:"total_price"`
}

// AddItemRequest is the JSON body for adding an item.
type AddItemRequest struct {
	ProductID string `json:"product_id" binding:"required"`
	VariantID string `json:"variant_id,omitempty"`
	Quantity  int    `json:"quantity" binding:"required,min=1"`
}

// UpdateItemRequest is the JSON body for updating item quantity.
// Quantity=0 is valid and signals removal (handled in service layer).
type UpdateItemRequest struct {
	Quantity int `json:"quantity" binding:"min=0"`
}

// CheckoutRequest is the JSON body for converting the cart to an order.
type CheckoutRequest struct {
	Email           string          `json:"email,omitempty"` // used to send order confirmation; required for guests
	ShippingAddress ShippingAddress `json:"shipping_address" binding:"required"`
}

// ShippingAddress mirrors the order's shipping address.
type ShippingAddress struct {
	FirstName string `json:"first_name" binding:"required"`
	LastName  string `json:"last_name" binding:"required"`
	Address1  string `json:"address1" binding:"required"`
	Address2  string `json:"address2,omitempty"`
	City      string `json:"city" binding:"required"`
	State     string `json:"state,omitempty"`
	Country   string `json:"country" binding:"required"`
	ZipCode   string `json:"zip_code,omitempty"`
	Phone     string `json:"phone,omitempty"`
}
