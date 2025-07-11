package billing

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// RequireFeature returns a Gin middleware that gates the route on whether the
// authenticated shop's active plan declares the given feature flag as truthy
// in its features JSONB object.
//
// Usage:
//
//	customDomainGate := billingService.RequireFeature("custom_domain")
//	r.POST("/shops/:id/domain", requireAuth, requireStaff, customDomainGate, handler)
//
// A feature is considered "enabled" when features[key] is a bool true, the
// string "true", or any non-zero number.
func (s *Service) RequireFeature(featureKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		shopID := c.GetString("shop_id")
		if shopID == "" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "shop context missing"})
			return
		}

		features, err := s.GetPlanFeatures(c.Request.Context(), shopID)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "could not check plan features"})
			return
		}

		if !featureEnabled(features, featureKey) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error":   "feature_not_available",
				"feature": featureKey,
				"message": "This feature is not available on your current plan. Please upgrade.",
			})
			return
		}

		c.Next()
	}
}

// featureEnabled checks whether a key in the features map represents a truthy
// value. Supported value types: bool, string ("true"), float64 (non-zero).
func featureEnabled(features map[string]any, key string) bool {
	v, ok := features[key]
	if !ok {
		return false
	}
	switch val := v.(type) {
	case bool:
		return val
	case string:
		return val == "true" || val == "1" || val == "yes"
	case float64:
		return val != 0
	default:
		return false
	}
}
