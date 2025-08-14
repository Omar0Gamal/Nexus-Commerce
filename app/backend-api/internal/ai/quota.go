package ai

import (
	"fmt"
	"net/http"
	"time"

	"backend-api/internal/db"
	"backend-api/internal/shared/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

// RequireAIFeature returns a Gin middleware that checks both the plan feature flag
// AND the monthly quota before allowing an AI endpoint through.
// key should be "ai_text_generation" or "ai_image_analysis".
func RequireAIFeature(key string, queries *db.Queries) gin.HandlerFunc {
	return func(c *gin.Context) {
		shopID := c.GetString("shop_id")
		if shopID == "" {
			response.BadRequest(c, "missing shop context")
			c.Abort()
			return
		}

		// Fetch shop via GetShop.
		shopUUID, err := uuid.Parse(shopID)
		if err != nil {
			response.BadRequest(c, "invalid shop id")
			c.Abort()
			return
		}
		shop, err := queries.GetShop(c.Request.Context(), shopUUID)
		if err != nil {
			response.InternalError(c)
			c.Abort()
			return
		}

		// Read feature flag from plan features JSONB.
		plan, err := queries.GetPlan(c.Request.Context(), shop.PlanID)
		if err != nil {
			response.InternalError(c)
			c.Abort()
			return
		}

		var features map[string]any
		if err := unmarshalJSONB(plan.Features, &features); err != nil || features[key] != true {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "feature_not_available",
				"message": fmt.Sprintf("The '%s' feature is not available on your current plan.", key),
			})
			c.Abort()
			return
		}

		// Check monthly quota.
		quotaKey := "ai_monthly_text_quota"
		if key == "ai_image_analysis" {
			quotaKey = "ai_monthly_image_quota"
		}
		quota, ok := features[quotaKey].(float64)
		if !ok || quota <= 0 {
			c.Next()
			return
		}

		now := time.Now().UTC()
		periodStart := pgtype.Timestamptz{
			Time:  time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC),
			Valid: true,
		}
		used, err := queries.CountAIUsage(c.Request.Context(), db.CountAIUsageParams{
			ShopID:      mustParseUUID(shopID),
			Feature:     key,
			PeriodStart: periodStart,
		})
		if err != nil {
			response.InternalError(c)
			c.Abort()
			return
		}

		if float64(used) >= quota {
			resetsAt := time.Date(now.Year(), now.Month()+1, 1, 0, 0, 0, 0, time.UTC)
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":     "ai_quota_exceeded",
				"resets_at": resetsAt.Format(time.RFC3339),
				"quota":     int(quota),
				"used":      used,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
