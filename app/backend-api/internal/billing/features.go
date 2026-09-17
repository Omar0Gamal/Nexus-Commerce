package billing

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

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

func featureEnabled(features map[string]any, key string) bool {
	rawValue, ok := features[key]
	if !ok {
		return false
	}
	switch featureValue := rawValue.(type) {
	case bool:
		return featureValue
	case string:
		return featureValue == "true" || featureValue == "1" || featureValue == "yes"
	case float64:
		return featureValue != 0
	default:
		return false
	}
}
