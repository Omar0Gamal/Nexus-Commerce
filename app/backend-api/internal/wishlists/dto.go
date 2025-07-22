package wishlists

import "time"

// WishlistItem is the response DTO for a single wishlist entry.
type WishlistItem struct {
	ID           string    `json:"id"`
	ProductID    string    `json:"product_id"`
	VariantID    *string   `json:"variant_id,omitempty"`
	ProductTitle string    `json:"product_title"`
	ProductSlug  string    `json:"product_slug"`
	ProductPrice float64   `json:"product_price"`
	ProductStock int32     `json:"product_stock"`
	CreatedAt    time.Time `json:"created_at"`
}

// AddRequest is the body for POST /wishlist.
type AddRequest struct {
	ProductID string  `json:"product_id" binding:"required"`
	VariantID *string `json:"variant_id,omitempty"`
}

// MergeRequest is the body for POST /wishlist/merge (guest → authenticated).
type MergeRequest struct {
	ProductIDs []string `json:"product_ids" binding:"required"`
}
