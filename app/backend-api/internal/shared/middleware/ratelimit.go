package middleware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// NewIPRateLimiter returns a Gin middleware that enforces a fixed-window rate
// limit per client IP address using Redis as the counter store.
//
//   - limit: maximum requests allowed within the window
//   - window: duration of the sliding window (e.g. time.Minute)
//
// On Redis failure the middleware fails open (allows the request) so a Redis
// outage never blocks legitimate users.
func NewIPRateLimiter(rdb *redis.Client, limit int, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		windowSec := int64(window.Seconds())
		// Key bucketed to the current window slot.
		slot := time.Now().Unix() / windowSec
		key := fmt.Sprintf("rl:auth:%s:%d", ip, slot)

		count, err := rdb.Incr(c.Request.Context(), key).Result()
		if err != nil {
			// Fail open: Redis unavailable — let the request through.
			c.Next()
			return
		}

		// Set expiry only on the first increment in this window.
		if count == 1 {
			rdb.Expire(c.Request.Context(), key, window+5*time.Second)
		}

		remaining := limit - int(count)
		if remaining < 0 {
			remaining = 0
		}

		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", limit))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
		c.Header("X-RateLimit-Window", window.String())

		if int(count) > limit {
			c.Header("Retry-After", fmt.Sprintf("%d", windowSec))
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "Too many requests. Please slow down and try again later.",
			})
			return
		}

		c.Next()
	}
}
