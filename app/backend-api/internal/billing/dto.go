package billing

type ChangePlanRequest struct {
	PlanID string `json:"plan_id" binding:"required"`
}
