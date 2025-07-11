package billing

// ChangePlanRequest is the body for PATCH /billing/plan.
type ChangePlanRequest struct {
	PlanID string `json:"plan_id" binding:"required"`
}
