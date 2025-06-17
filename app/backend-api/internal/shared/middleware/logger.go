package middleware

import (
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

var logger *zap.Logger

// SetLogger configures the package-level logger used by middleware.
// Call this once at startup, before registering any middleware.
func SetLogger(l *zap.Logger) {
	logger = l
}

// StructuredLogger returns a middleware that logs HTTP requests
func StructuredLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		statusCode := c.Writer.Status()
		method := c.Request.Method
		clientIP := c.ClientIP()

		logger.Info("HTTP Request",
			zap.String("method", method),
			zap.String("path", path),
			zap.String("query", query),
			zap.Int("status", statusCode),
			zap.Duration("latency", latency),
			zap.String("ip", clientIP),
			zap.String("user-agent", c.Request.UserAgent()),
			zap.String("request_id", c.GetHeader("X-Request-ID")),
		)
	}
}

// SmartCORS returns a CORS middleware configured for development and production.
// In production the backend sits behind the gateway, which handles CORS.
// In development we also allow direct access and subdomain origins.
func SmartCORS() gin.HandlerFunc {
	config := cors.Config{
		AllowOriginFunc: func(origin string) bool {
			// Only allow real localhost origins (any port).
			return strings.HasPrefix(origin, "http://localhost") ||
				strings.HasPrefix(origin, "https://localhost")
		},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Tenant-ID", "X-Shop-ID"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}

	return cors.New(config)
}
