package analytics

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type Event struct {
	Event     string         `json:"event"`
	ShopID    string         `json:"shop_id"`
	SessionID string         `json:"session_id,omitempty"`
	Timestamp time.Time      `json:"timestamp"`
	Payload   map[string]any `json:"payload"`
}

// Publish writes to a capped Redis stream and discards errors so analytics outages never block callers.
func Publish(ctx context.Context, rdb *redis.Client, shopID string, e Event) {
	e.Timestamp = time.Now().UTC()
	payloadBytes, err := json.Marshal(e)
	if err != nil {
		return
	}
	rdb.XAdd(ctx, &redis.XAddArgs{
		Stream: fmt.Sprintf("events:analytics:%s", shopID),
		MaxLen: 50_000,
		Approx: true,
		Values: map[string]any{"data": string(payloadBytes)},
	})
}
