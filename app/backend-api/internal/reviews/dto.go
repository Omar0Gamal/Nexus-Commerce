package reviews

import "time"

type SubmitRequest struct {
	CustomerID string  `json:"-"`
	ProductID  string  `json:"-"`
	OrderID    *string `json:"order_id,omitempty"`
	Rating     int16   `json:"rating" binding:"required,min=1,max=5"`
	Title      *string `json:"title,omitempty"`
	Body       *string `json:"body,omitempty"`
}

type ReviewDTO struct {
	ID                 string    `json:"id"`
	ProductID          string    `json:"product_id"`
	CustomerID         string    `json:"customer_id"`
	CustomerName       string    `json:"customer_name,omitempty"`
	Rating             int16     `json:"rating"`
	Title              *string   `json:"title,omitempty"`
	Body               *string   `json:"body,omitempty"`
	Status             string    `json:"status"`
	IsVerifiedPurchase bool      `json:"is_verified_purchase"`
	HelpfulCount       int32     `json:"helpful_count"`
	CreatedAt          time.Time `json:"created_at"`
}

type PendingReviewDTO struct {
	ID           string    `json:"id"`
	ProductID    string    `json:"product_id"`
	ProductTitle string    `json:"product_title"`
	CustomerID   string    `json:"customer_id"`
	CustomerName string    `json:"customer_name,omitempty"`
	Rating       int16     `json:"rating"`
	Title        *string   `json:"title,omitempty"`
	Body         *string   `json:"body,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}

type ReviewSummary struct {
	TotalReviews int         `json:"total_reviews"`
	AvgRating    float64     `json:"avg_rating"`
	Distribution map[int]int `json:"distribution"`
}

type ModerateRequest struct {
	Status string `json:"status" binding:"required,oneof=approved rejected"`
}
