package admin

type CreateShopRequest struct {
	Name      string `json:"name" binding:"required,min=1,max=100"`
	Subdomain string `json:"subdomain" binding:"required,min=3,max=63,alphanum"`
	Currency  string `json:"currency" binding:"required"`
	Timezone  string `json:"timezone" binding:"required"`
	Country   string `json:"country" binding:"required"`
	Industry  string `json:"industry"`
}

type ShopResponse struct {
	ID           string  `json:"id"`
	Name         string  `json:"name"`
	Subdomain    string  `json:"subdomain"`
	CustomDomain *string `json:"custom_domain,omitempty"`
	Status       string  `json:"status"`
	Currency     string  `json:"currency"`
	Timezone     string  `json:"timezone"`
	OwnerUserID  string  `json:"owner_user_id"`
	PlanID       string  `json:"plan_id"`
	PlanName     string  `json:"plan_name"`
	CreatedAt    string  `json:"created_at"`
}

type PlatformStatsResponse struct {
	TotalShops     int64 `json:"total_shops"`
	ActiveShops    int64 `json:"active_shops"`
	SuspendedShops int64 `json:"suspended_shops"`
}

type UpdateShopStatusRequest struct {
	Status string `json:"status" binding:"required"`
}
