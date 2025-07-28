// Package analytics provides a fire-and-forget event publisher that sends
// domain events to a Redis Pub/Sub channel for downstream consumption.
package analytics

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// Event is a domain event emitted by backend services.
type Event struct {
	Event     string         `json:"event"`
	ShopID    string         `json:"shop_id"`
	SessionID string         `json:"session_id,omitempty"`
	Timestamp time.Time      `json:"timestamp"`
	Payload   map[string]any `json:"payload"`
}

// Publish appends an analytics event to the shop's Redis Stream.
// The stream is capped at 50 000 entries (MAXLEN ~ 50000, approximate trim).
// It runs fire-and-forget: errors are silently discarded so that a Redis outage
// never blocks the hot path.
func Publish(ctx context.Context, rdb *redis.Client, shopID string, e Event) {
	e.Timestamp = time.Now().UTC()
	data, err := json.Marshal(e)
	if err != nil {
		return
	}
	// Best-effort — ignore stream errors
	rdb.XAdd(ctx, &redis.XAddArgs{
		Stream: fmt.Sprintf("events:analytics:%s", shopID),
		MaxLen: 50_000,
		Approx: true,
		Values: map[string]any{"data": string(data)},
	})
}
