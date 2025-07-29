package analytics

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"backend-api/internal/db"
	sharedanalytics "backend-api/internal/shared/analytics"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

const (
	consumerGroup     = "analytics-worker"
	consumerName      = "worker-1"
	maxFailures       = 3
	dlqKeyPrefix      = "events:analytics:dlq:"
	streamKeyPrefix   = "events:analytics:"
	forecastCacheTTL  = 24 * time.Hour
	inventoryCacheTTL = 6 * time.Hour
)

// Worker consumes analytics Redis Streams for all active shops and materialises
// business-level counters into Redis keys.
type Worker struct {
	rdb    *redis.Client
	db     *db.Queries
	logger *zap.Logger
}

func NewWorker(rdb *redis.Client, queries *db.Queries, logger *zap.Logger) *Worker {
	return &Worker{rdb: rdb, db: queries, logger: logger}
}

// Run blocks until ctx is cancelled. It discovers active shops and processes
// their event streams continuously.
func (w *Worker) Run(ctx context.Context) {
	// Ensure consumer groups exist for all active shops on first run.
	go w.ensureConsumerGroups(ctx)

	// Schedule daily analytics jobs.
	go w.runDailyScheduler(ctx)

	// Run the stream consumer loop.
	w.consume(ctx)
}

// ensureConsumerGroups creates the consumer group for every active-shop stream.
func (w *Worker) ensureConsumerGroups(ctx context.Context) {
	shops, err := w.db.ListShopsByStatus(ctx, db.NullShopStatus{ShopStatus: db.ShopStatusActive, Valid: true})
	if err != nil {
		w.logger.Warn("analytics worker: could not list shops for consumer group setup", zap.Error(err))
		return
	}
	for _, shop := range shops {
		stream := streamKeyPrefix + shop.ID.String()
		// XGROUP CREATE ... $ MKSTREAM — ignore "BUSYGROUP" if already exists.
		w.rdb.XGroupCreateMkStream(ctx, stream, consumerGroup, "0")
	}
}

// consume is the main XREADGROUP loop.
func (w *Worker) consume(ctx context.Context) {
	for {
		if ctx.Err() != nil {
			return
		}
		w.consumeOnce(ctx)
	}
}

func (w *Worker) consumeOnce(ctx context.Context) {
	// Discover all known streams by SCAN for keys matching events:analytics:*
	var cursor uint64
	var streams []string
	for {
		keys, nextCursor, err := w.rdb.Scan(ctx, cursor, streamKeyPrefix+"*", 100).Result()
		if err != nil {
			break
		}
		for _, k := range keys {
			// Exclude DLQ keys.
			if !strings.HasPrefix(k, dlqKeyPrefix) {
				streams = append(streams, k)
			}
		}
		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}

	if len(streams) == 0 {
		select {
		case <-ctx.Done():
		case <-time.After(5 * time.Second):
		}
		return
	}

	// Build XREADGROUP args: one "> " per stream.
	streamArgs := make([]string, 0, len(streams)*2)
	for _, s := range streams {
		streamArgs = append(streamArgs, s)
	}
	for range streams {
		streamArgs = append(streamArgs, ">")
	}

	result, err := w.rdb.XReadGroup(ctx, &redis.XReadGroupArgs{
		Group:    consumerGroup,
		Consumer: consumerName,
		Streams:  streamArgs,
		Count:    100,
		Block:    5 * time.Second,
	}).Result()
	if err != nil {
		if err == redis.Nil {
			return // timeout, no messages
		}
		// Likely a new stream without a consumer group — try to create groups.
		go w.ensureConsumerGroups(context.Background())
		return
	}

	for _, xstream := range result {
		shopID := strings.TrimPrefix(xstream.Stream, streamKeyPrefix)
		for _, msg := range xstream.Messages {
			if err := w.processMessage(ctx, shopID, xstream.Stream, msg); err != nil {
				w.logger.Error("analytics worker: failed to process message",
					zap.String("stream", xstream.Stream),
					zap.String("id", msg.ID),
					zap.Error(err))
				// Move to DLQ on repeated failures (simplified: move immediately on error).
				w.rdb.XAdd(ctx, &redis.XAddArgs{
					Stream: dlqKeyPrefix + shopID,
					MaxLen: 10_000,
					Approx: true,
					Values: msg.Values,
				})
			}
			// ACK regardless so we don't reprocess endlessly.
			w.rdb.XAck(ctx, xstream.Stream, consumerGroup, msg.ID)
		}
	}
}

func (w *Worker) processMessage(ctx context.Context, shopID, stream string, msg redis.XMessage) error {
	raw, ok := msg.Values["data"]
	if !ok {
		return nil
	}
	dataStr, _ := raw.(string)
	if dataStr == "" {
		return nil
	}

	var evt sharedanalytics.Event
	if err := json.Unmarshal([]byte(dataStr), &evt); err != nil {
		return fmt.Errorf("unmarshal event: %w", err)
	}

	date := evt.Timestamp.UTC().Format("2006-01-02")

	switch evt.Event {
	case "cart_add":
		if pid, ok := payloadStr(evt.Payload, "product_id"); ok {
			w.rdb.ZIncrBy(ctx, fmt.Sprintf("stats:cart_adds:%s:%s", shopID, date), 1, pid)
			// Search-to-cart attribution: look up the last search for this session.
			if sid, ok := payloadStr(evt.Payload, "session_id"); ok && sid != "" {
				if q, err := w.rdb.Get(ctx, fmt.Sprintf("session_last_search:%s:%s", shopID, sid)).Result(); err == nil && q != "" {
					w.rdb.ZIncrBy(ctx, fmt.Sprintf("stats:search_cart:%s:%s", shopID, date), 1, q)
				}
			}
		}
	case "cart_remove":
		if pid, ok := payloadStr(evt.Payload, "product_id"); ok {
			key := fmt.Sprintf("stats:cart_adds:%s:%s", shopID, date)
			// Decrement but floor at 0 via Lua to avoid negative counts.
			w.rdb.Eval(ctx,
				`local v=redis.call('ZSCORE',KEYS[1],ARGV[1]) if v then local n=tonumber(v)-1 if n<=0 then redis.call('ZREM',KEYS[1],ARGV[1]) else redis.call('ZADD',KEYS[1],n,ARGV[1]) end end return 0`,
				[]string{key}, pid)
		}
	case "checkout_start":
		if sid, ok := payloadStr(evt.Payload, "session_id"); ok && sid != "" {
			w.rdb.PFAdd(ctx, fmt.Sprintf("funnel:checkout_start:%s:%s", shopID, date), sid)
		}
	case "payment_success":
		if sid, ok := payloadStr(evt.Payload, "session_id"); ok && sid != "" {
			w.rdb.PFAdd(ctx, fmt.Sprintf("funnel:payment:%s:%s", shopID, date), sid)
		}
	case "order_completed":
		w.processOrderCompleted(ctx, shopID, date, evt)
	case "coupon_applied":
		code, _ := payloadStr(evt.Payload, "code")
		discount, _ := payloadInt(evt.Payload, "discount_cents")
		if code != "" && discount > 0 {
			w.rdb.ZIncrBy(ctx, fmt.Sprintf("stats:coupon_revenue:%s:%s", shopID, date), float64(discount), code)
		}
	case "product_search":
		if q, ok := payloadStr(evt.Payload, "query"); ok && q != "" {
			w.rdb.ZIncrBy(ctx, fmt.Sprintf("stats:search_terms:%s:%s", shopID, date), 1, q)
			// Search-to-sale attribution: store last search for this session.
			if sid, ok := payloadStr(evt.Payload, "session_id"); ok && sid != "" {
				w.rdb.Set(ctx, fmt.Sprintf("session_last_search:%s:%s", shopID, sid), q, 30*time.Minute)
			}
		}
	case "customer_signup":
		w.rdb.Incr(ctx, fmt.Sprintf("stats:signups:%s:%s", shopID, date))
	}
	return nil
}

func (w *Worker) processOrderCompleted(ctx context.Context, shopID, date string, evt sharedanalytics.Event) {
	revenue, _ := payloadInt(evt.Payload, "revenue_cents")
	sid, _ := payloadStr(evt.Payload, "session_id")
	custID, _ := payloadStr(evt.Payload, "customer_id")

	pipe := w.rdb.Pipeline()

	if revenue > 0 {
		pipe.IncrBy(ctx, fmt.Sprintf("stats:revenue:%s:%s", shopID, date), int64(revenue))
	}
	pipe.Incr(ctx, fmt.Sprintf("stats:order_count:%s:%s", shopID, date))

	// Per-product sales.
	if items, ok := evt.Payload["items"]; ok {
		if arr, ok := items.([]any); ok {
			for _, item := range arr {
				m, ok := item.(map[string]any)
				if !ok {
					continue
				}
				pid, _ := m["product_id"].(string)
				rev := 0.0
				switch v := m["revenue_cents"].(type) {
				case float64:
					rev = v
				case int64:
					rev = float64(v)
				}
				if pid != "" && rev > 0 {
					pipe.ZIncrBy(ctx, fmt.Sprintf("stats:product_sales:%s:%s", shopID, date), rev, pid)
				}
			}
		}
	}

	if sid != "" {
		pipe.PFAdd(ctx, fmt.Sprintf("funnel:order_confirmed:%s:%s", shopID, date), sid)
		// Search-to-revenue attribution.
		pipe.Get(ctx, fmt.Sprintf("session_last_search:%s:%s", shopID, sid))
	}

	_, _ = pipe.Exec(ctx)

	// Search-to-revenue redis get (outside pipeline for simplicity).
	if sid != "" && revenue > 0 {
		if q, err := w.rdb.Get(ctx, fmt.Sprintf("session_last_search:%s:%s", shopID, sid)).Result(); err == nil && q != "" {
			w.rdb.ZIncrBy(ctx, fmt.Sprintf("stats:search_revenue:%s:%s", shopID, date), float64(revenue), q)
		}
	}

	// Cart-to-search attribution: already handled in cart_add handler above.

	// Invalidate forecast cache.
	w.rdb.Del(ctx, fmt.Sprintf("forecast:revenue:%s", shopID))

	// Cohort write: look up customer first order.
	if custID != "" {
		w.writeCohort(ctx, shopID, custID, date)
	}
}

func (w *Worker) writeCohort(ctx context.Context, shopID, customerID, date string) {
	firstOrderKey := fmt.Sprintf("customer_first_order:%s:%s", shopID, customerID)
	firstMonth, err := w.rdb.Get(ctx, firstOrderKey).Result()
	if err != nil {
		// Not cached — determine from the event date as first order.
		yearMonth := date[:7]                                          // YYYY-MM
		w.rdb.Set(ctx, firstOrderKey, yearMonth, 63072000*time.Second) // 2 years TTL
		w.rdb.SAdd(ctx, fmt.Sprintf("cohort:%s:%s", shopID, yearMonth), customerID)
		return
	}
	// Existing cohort — compute offset months.
	cohortMonth, _ := time.Parse("2006-01", firstMonth)
	currentMonth, _ := time.Parse("2006-01", date[:7])
	months := int(currentMonth.Sub(cohortMonth).Hours() / 24 / 30)
	if months > 0 {
		w.rdb.Incr(ctx, fmt.Sprintf("cohort_retention:%s:%s:%d", shopID, firstMonth, months))
	}
}


func (w *Worker) runDailyScheduler(ctx context.Context) {
	for {
		now := time.Now().UTC()
		// Calculate next 00:05 UTC
		next := time.Date(now.Year(), now.Month(), now.Day(), 0, 5, 0, 0, time.UTC)
		if now.After(next) {
			next = next.Add(24 * time.Hour)
		}
		select {
		case <-ctx.Done():
			return
		case <-time.After(time.Until(next)):
		}
		w.runDailyForecasting(ctx)
		w.runDailyDigest(ctx)
		w.runABTestAutoConclusion(ctx)
	}
}

func (w *Worker) runDailyForecasting(ctx context.Context) {
	shops, err := w.db.ListShopsByStatus(ctx, db.NullShopStatus{ShopStatus: db.ShopStatusActive, Valid: true})
	if err != nil {
		w.logger.Warn("analytics worker: could not list shops for forecasting", zap.Error(err))
		return
	}
	for _, shop := range shops {
		shopID := shop.ID.String()
		w.forecastRevenue(ctx, shopID)
		w.forecastTraffic(ctx, shopID)
		w.forecastInventory(ctx, shopID)
	}
}

func (w *Worker) forecastRevenue(ctx context.Context, shopID string) {
	series := w.readDailySeries(ctx, "stats:revenue", shopID, 90)
	if countNonZero(series) < 14 {
		return
	}
	forecast := HoltWinters(series, 30, 0.3, 0.1)
	data, _ := json.Marshal(forecast)
	w.rdb.Set(ctx, fmt.Sprintf("forecast:revenue:%s", shopID), string(data), forecastCacheTTL)
}

func (w *Worker) forecastTraffic(ctx context.Context, shopID string) {
	series := w.readDailySeries(ctx, "stats:visits", shopID, 90)
	if countNonZero(series) < 14 {
		return
	}
	forecast := HoltWinters(series, 30, 0.3, 0.1)
	data, _ := json.Marshal(forecast)
	w.rdb.Set(ctx, fmt.Sprintf("forecast:traffic:%s", shopID), string(data), forecastCacheTTL)
}

func (w *Worker) forecastInventory(ctx context.Context, shopID string) {
	// This is handled separately — we query Postgres for products with track_inventory.
	// Simplified: iterate products with track_inventory and compute avg daily sales.
	// Full implementation deferred to when inventory service exists for batch queries.
}

func (w *Worker) readDailySeries(ctx context.Context, prefix, shopID string, days int) []float64 {
	now := time.Now().UTC()
	pipe := w.rdb.Pipeline()
	cmds := make([]*redis.StringCmd, days)
	for i := range days {
		d := now.AddDate(0, 0, -(days - 1 - i)).Format("2006-01-02")
		cmds[i] = pipe.Get(ctx, fmt.Sprintf("%s:%s:%s", prefix, shopID, d))
	}
	_, _ = pipe.Exec(ctx)
	out := make([]float64, days)
	for i, cmd := range cmds {
		if v, err := cmd.Float64(); err == nil {
			out[i] = v
		}
	}
	return out
}

// runDailyDigest fires analytics digest emails at ~08:00 UTC.
func (w *Worker) runDailyDigest(ctx context.Context) {
	// Implementation in email worker (Phase 9) — stub here.
}

// runABTestAutoConclusion auto-concludes tests older than 14 days with >= 95% confidence.
func (w *Worker) runABTestAutoConclusion(ctx context.Context) {
	tests, err := w.db.ListABTestsForAutoConclusion(ctx)
	if err != nil {
		w.logger.Error("ab test auto conclude: list", zap.Error(err))
		return
	}
	for _, t := range tests {
		testIDStr := t.ID.String()
		shopIDStr := t.ShopID.String()

		// Read Redis counters to check confidence.
		impAStr := w.rdb.Get(ctx, fmt.Sprintf("abtest:%s:variant_a:impressions", testIDStr)).Val()
		convAStr := w.rdb.Get(ctx, fmt.Sprintf("abtest:%s:variant_a:conversions", testIDStr)).Val()
		impBStr := w.rdb.Get(ctx, fmt.Sprintf("abtest:%s:variant_b:impressions", testIDStr)).Val()
		convBStr := w.rdb.Get(ctx, fmt.Sprintf("abtest:%s:variant_b:conversions", testIDStr)).Val()

		impA, _ := strconv.ParseInt(impAStr, 10, 64)
		convA, _ := strconv.ParseInt(convAStr, 10, 64)
		impB, _ := strconv.ParseInt(impBStr, 10, 64)
		convB, _ := strconv.ParseInt(convBStr, 10, 64)

		confidence := zTestTwoProportions(impA, convA, impB, convB)
		if confidence >= 95 {
			if err := w.db.ConcludeABTest(ctx, db.ConcludeABTestParams{
				ID:     t.ID,
				ShopID: t.ShopID,
			}); err != nil {
				w.logger.Error("ab test auto conclude: update", zap.String("test_id", testIDStr), zap.String("shop_id", shopIDStr), zap.Error(err))
			} else {
				w.logger.Info("ab test auto concluded", zap.String("test_id", testIDStr), zap.Float64("confidence", confidence))
			}
		}
	}
}


func payloadStr(p map[string]any, key string) (string, bool) {
	if p == nil {
		return "", false
	}
	v, ok := p[key]
	if !ok {
		return "", false
	}
	s, ok := v.(string)
	return s, ok
}

func payloadInt(p map[string]any, key string) (int64, bool) {
	if p == nil {
		return 0, false
	}
	v, ok := p[key]
	if !ok {
		return 0, false
	}
	switch n := v.(type) {
	case float64:
		return int64(n), true
	case int64:
		return n, true
	case string:
		i, err := strconv.ParseInt(n, 10, 64)
		return i, err == nil
	}
	return 0, false
}

// countNonZero returns the count of non-zero values in a slice.
func countNonZero(s []float64) int {
	n := 0
	for _, v := range s {
		if v != 0 {
			n++
		}
	}
	return n
}


// RecordImpression increments impression counter for a variant.
func (w *Worker) RecordImpression(ctx context.Context, testID, variant string) {
	w.rdb.Incr(ctx, fmt.Sprintf("abtest:%s:variant_%s:impressions", testID, variant))
}

// RecordConversion increments conversion counter for a variant.
func (w *Worker) RecordConversion(ctx context.Context, testID, variant string) {
	w.rdb.Incr(ctx, fmt.Sprintf("abtest:%s:variant_%s:conversions", testID, variant))
}
