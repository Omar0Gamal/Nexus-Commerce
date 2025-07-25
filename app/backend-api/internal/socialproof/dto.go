package socialproof

import "time"

// RecentEvent is the public-facing social proof event.
type RecentEvent struct {
	ProductID    string    `json:"product_id"`
	ProductTitle string    `json:"product_title"`
	ProductSlug  string    `json:"product_slug"`
	City         string    `json:"city,omitempty"`
	Country      string    `json:"country,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}
