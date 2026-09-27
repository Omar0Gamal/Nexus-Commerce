package notifications

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

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
	c.Writer.Header().Set("X-Accel-Buffering", "no") // Disable reverse-proxy response buffering for SSE.
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
