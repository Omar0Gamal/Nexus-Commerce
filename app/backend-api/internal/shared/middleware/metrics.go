package middleware

import (
	"time"

	"backend-api/internal/metrics"

	"github.com/gin-gonic/gin"
)

// PrometheusInstrumentation returns a Gin middleware that records request
// metrics using the backend metrics package.
func PrometheusInstrumentation() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		duration := time.Since(start)
		statusClass := metrics.StatusClass(c.Writer.Status())
		metrics.RequestsTotal.WithLabelValues(c.Request.Method, statusClass).Inc()
		metrics.RequestDuration.WithLabelValues(c.Request.Method, statusClass).Observe(duration.Seconds())
	}
}
