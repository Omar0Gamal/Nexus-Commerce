package notifications

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Holds the HTTP connection open, subscribes to Redis Pub/Sub for the
// authenticated user, and writes SSE data frames on every new notification.
// A ping comment is written every 30 s to keep the connection alive.
func (h *Handler) SSEStream(c *gin.Context) {
	shopID, recipientType, recipientID, ok := extractActor(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	channel := sseChannel(shopID, recipientType, recipientID)
	pubsub := h.svc.rdb.Subscribe(c.Request.Context(), channel)
	defer pubsub.Close()

	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("X-Accel-Buffering", "no") // nginx: disable buffering
	c.Writer.WriteHeader(http.StatusOK)
	c.Writer.Flush()

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	msgCh := pubsub.Channel()
	for {
		select {
		case <-c.Request.Context().Done():
			return
		case msg, ok := <-msgCh:
			if !ok {
				return
			}
			fmt.Fprintf(c.Writer, "data: %s\n\n", msg.Payload)
			c.Writer.Flush()
		case <-ticker.C:
			fmt.Fprintf(c.Writer, ": ping\n\n")
			c.Writer.Flush()
		}
	}
}

// extractActor reads shop_id, actor_type, and user_id from gin context.
func extractActor(c *gin.Context) (shopID uuid.UUID, recipientType string, recipientID uuid.UUID, ok bool) {
	shopIDStr := c.GetString("shop_id")
	actorType := c.GetString("actor_type")
	userIDStr := c.GetString("user_id")

	if shopIDStr == "" || actorType == "" || userIDStr == "" {
		return
	}

	var err error
	shopID, err = uuid.Parse(shopIDStr)
	if err != nil {
		return
	}
	recipientID, err = uuid.Parse(userIDStr)
	if err != nil {
		return
	}

	recipientType = actorType
	ok = true
	return
}
