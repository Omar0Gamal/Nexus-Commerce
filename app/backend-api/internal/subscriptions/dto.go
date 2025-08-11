package subscriptions

import "time"


type CreatePlanRequest struct {
	Name         string  `json:"name" binding:"required"`
	Description  string  `json:"description"`
	ProductID    string  `json:"product_id"`
	Price        float64 `json:"price" binding:"required,gt=0"`
	BillingCycle string  `json:"billing_cycle" binding:"required,oneof=weekly monthly quarterly annual"`
	TrialDays    int32   `json:"trial_days"`
	IsActive     *bool   `json:"is_active"`
}

type UpdatePlanRequest struct {
	Name         string  `json:"name"`
	Description  string  `json:"description"`
	Price        float64 `json:"price"`
	BillingCycle string  `json:"billing_cycle" binding:"omitempty,oneof=weekly monthly quarterly annual"`
	TrialDays    int32   `json:"trial_days"`
	IsActive     *bool   `json:"is_active"`
}

type PlanResponse struct {
	ID           string  `json:"id"`
	ShopID       string  `json:"shop_id"`
	Name         string  `json:"name"`
	Description  string  `json:"description,omitempty"`
	ProductID    *string `json:"product_id,omitempty"`
	Price        string  `json:"price"`
	BillingCycle string  `json:"billing_cycle"`
	TrialDays    int32   `json:"trial_days"`
	IsActive     bool    `json:"is_active"`
	CreatedAt    string  `json:"created_at"`
}


type SubscribeRequest struct {
	PlanID string `json:"plan_id" binding:"required,uuid"`
}

type SubscriptionResponse struct {
	ID                 string  `json:"id"`
	ShopID             string  `json:"shop_id"`
	CustomerID         string  `json:"customer_id"`
	PlanID             string  `json:"plan_id"`
	PlanName           string  `json:"plan_name"`
	PlanPrice          string  `json:"plan_price"`
	BillingCycle       string  `json:"billing_cycle"`
	Status             string  `json:"status"`
	CurrentPeriodStart string  `json:"current_period_start"`
	CurrentPeriodEnd   string  `json:"current_period_end"`
	NextBillingAt      string  `json:"next_billing_at"`
	TrialEnd           *string `json:"trial_end,omitempty"`
	CancelledAt        *string `json:"cancelled_at,omitempty"`
	CreatedAt          string  `json:"created_at"`
}

type ListSubscriptionsQuery struct {
	Page    int `form:"page,default=1"`
	PerPage int `form:"per_page,default=20"`
}

// nextBillingDate calculates the next billing timestamp based on the billing cycle.
func nextBillingDate(from time.Time, cycle string) time.Time {
	switch cycle {
	case "weekly":
		return from.AddDate(0, 0, 7)
	case "quarterly":
		return from.AddDate(0, 3, 0)
	case "annual":
		return from.AddDate(1, 0, 0)
	default: // monthly
		return from.AddDate(0, 1, 0)
	}
}
