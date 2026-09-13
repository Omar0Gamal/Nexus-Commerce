package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// TenantMiddleware enforces X-Gateway-Secret to prevent header spoofing and requires X-Shop-ID.
func TenantMiddleware(gatewaySecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if gatewaySecret != "" {
			if secret := c.GetHeader("X-Gateway-Secret"); secret == "" || secret != gatewaySecret {
				c.JSON(http.StatusUnauthorized, gin.H{
					"error": "Unauthorized gateway",
				})
				c.Abort()
				return
			}
		}

		shopID := c.GetHeader("X-Shop-ID")
		if shopID == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Missing tenant context. X-Shop-ID header is required.",
			})
			c.Abort()
			return
		}

		c.Set("shop_id", shopID)
		c.Next()
	}
}

func OptionalTenantMiddleware(gatewaySecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		shopID := c.GetHeader("X-Shop-ID")
		if shopID != "" {
			if gatewaySecret != "" {
				if secret := c.GetHeader("X-Gateway-Secret"); secret == "" || secret != gatewaySecret {
					c.JSON(http.StatusUnauthorized, gin.H{
						"error": "Unauthorized gateway",
					})
					c.Abort()
					return
				}
			}
			c.Set("shop_id", shopID)
		}
		c.Next()
	}
}
