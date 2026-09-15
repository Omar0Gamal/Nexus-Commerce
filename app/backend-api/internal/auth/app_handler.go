package auth

import (
	"backend-api/internal/shared/response"

	"github.com/gin-gonic/gin"
)

type ReferenceDataResponse struct {
	Countries  []string `json:"countries"`
	Currencies []string `json:"currencies"`
	Timezones  []string `json:"timezones"`
	Industries []string `json:"industries"`
}

type BootstrapResponse struct {
	User                UserResponse         `json:"user"`
	ActiveShop          *ShopSettingsResponse `json:"active_shop,omitempty"`
	BillingStatus       string               `json:"billing_status"`
	UnreadNotifications int                  `json:"unread_notifications"`
}

func (h *Handler) RegisterAppRoutes(rg *gin.RouterGroup, authMW gin.HandlerFunc) {
	rg.GET("/bootstrap", authMW, h.Bootstrap)
	rg.GET("/reference-data", h.ReferenceData)
}

func (h *Handler) Bootstrap(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		response.Unauthorized(c, "not authenticated")
		return
	}

	resp := BootstrapResponse{
		User: UserResponse{
			ID:       userID,
			Email:    c.GetString("email"),
			FullName: c.GetString("full_name"),
		},
		BillingStatus:       "active",
		UnreadNotifications: 0,
	}

	response.OK(c, resp)
}

func (h *Handler) ReferenceData(c *gin.Context) {
	response.OK(c, ReferenceDataResponse{
		Countries:  []string{"EG", "US", "UK", "AE", "SA"},
		Currencies: []string{"EGP", "USD", "GBP", "EUR", "SAR", "AED"},
		Timezones:  []string{"Africa/Cairo", "America/New_York", "Europe/London", "Asia/Dubai", "Asia/Riyadh"},
		Industries: []string{"Retail", "Fashion", "Electronics", "Food", "Services", "Other"},
	})
}
