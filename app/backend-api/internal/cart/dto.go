package cart


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

type Cart struct {
	Items      []CartItem `json:"items"`
	TotalItems int        `json:"total_items"`
	TotalPrice float64    `json:"total_price"`
}

type AddItemRequest struct {
	ProductID string `json:"product_id" binding:"required"`
	VariantID string `json:"variant_id,omitempty"`
	Quantity  int    `json:"quantity" binding:"required,min=1"`
}

// Quantity=0 is valid and signals removal.
type UpdateItemRequest struct {
	Quantity int `json:"quantity" binding:"min=0"`
}

type CheckoutRequest struct {
	Email           string          `json:"email,omitempty"`
	ShippingAddress ShippingAddress `json:"shipping_address" binding:"required"`
}

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
