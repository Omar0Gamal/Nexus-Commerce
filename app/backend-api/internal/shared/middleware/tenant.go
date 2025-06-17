package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// TenantMiddleware extracts the shop identity from the X-Shop-ID header
// injected by the Gateway after tenant resolution. This is the trusted source.
// Direct API calls without a valid X-Shop-ID are rejected with 401.
func TenantMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		shopID := c.GetHeader("X-Shop-ID")
		if shopID == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Missing tenant context. X-Shop-ID header is required.",
			})
			c.Abort()
			return
		}

		// Store shop_id in context — all modules read from here
		c.Set("shop_id", shopID)
		c.Next()
	}
}

// OptionalTenantMiddleware extracts X-Shop-ID but doesn't require it.
// Useful for public/shared endpoints that behave differently per tenant.
func OptionalTenantMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		shopID := c.GetHeader("X-Shop-ID")
		if shopID != "" {
			c.Set("shop_id", shopID)
		}
		c.Next()
	}
}
