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
	"backend-api/internal/email"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/google/uuid"
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

type Worker struct {
	rdb    *redis.Client
	db     *db.Queries
	mailer *email.Mailer
	logger *zap.Logger

	cachedStreams      []string
	streamsLastFetched time.Time
}

func NewWorker(rdb *redis.Client, queries *db.Queries, mailer *email.Mailer, logger *zap.Logger) *Worker {
	return &Worker{rdb: rdb, db: queries, mailer: mailer, logger: logger}
}

func (w *Worker) Run(ctx context.Context) {
	go w.ensureConsumerGroups(ctx)
	go w.runDailyScheduler(ctx)
	go w.runAlertEngine(ctx)
	w.consume(ctx)
}

func (w *Worker) ensureConsumerGroups(ctx context.Context) {
	shops, err := w.db.ListShopsByStatus(ctx, db.NullShopStatus{ShopStatus: db.ShopStatusActive, Valid: true})
	if err != nil {
		w.logger.Warn("analytics worker: could not list shops for consumer group setup", zap.Error(err))
		return
	}
	for _, shop := range shops {
		stream := streamKeyPrefix + shop.ID.String()
		// MKSTREAM creates stream if absent; ignore BUSYGROUP error if already registered.
		w.rdb.XGroupCreateMkStream(ctx, stream, consumerGroup, "0")
	}
}

func (w *Worker) consume(ctx context.Context) {
	for {
		if ctx.Err() != nil {
			return
		}
		w.consumeOnce(ctx)
	}
}

func (w *Worker) consumeOnce(ctx context.Context) {
	var streams []string
	if time.Since(w.streamsLastFetched) < 10*time.Minute && len(w.cachedStreams) > 0 {
		streams = w.cachedStreams
	} else {
		var cursor uint64
		for {
			keys, nextCursor, err := w.rdb.Scan(ctx, cursor, streamKeyPrefix+"*", 100).Result()
			if err != nil {
				break
			}
			for _, k := range keys {
				if !strings.HasPrefix(k, dlqKeyPrefix) {
					streams = append(streams, k)
				}
			}
			cursor = nextCursor
			if cursor == 0 {
				break
			}
		}
		w.cachedStreams = streams
		w.streamsLastFetched = time.Now()
	}

	if len(streams) == 0 {
		select {
		case <-ctx.Done():
		case <-time.After(5 * time.Second):
		}
		return
	}

	streamArgs := make([]string, 0, len(streams)*2)
	streamArgs = append(streamArgs, streams...)
	for range streams {
		streamArgs = append(streamArgs, ">")
	}

	streamResults, err := w.rdb.XReadGroup(ctx, &redis.XReadGroupArgs{
		Group:    consumerGroup,
		Consumer: consumerName,
		Streams:  streamArgs,
		Count:    100,
		Block:    5 * time.Second,
	}).Result()
	if err != nil {
		if err == redis.Nil {
			return
		}
		// If consumer group missing on newly created stream, attempt initialization.
		go w.ensureConsumerGroups(context.Background())
		return
	}

	for _, xstream := range streamResults {
		shopID := strings.TrimPrefix(xstream.Stream, streamKeyPrefix)
		for _, msg := range xstream.Messages {
			if err := w.processMessage(ctx, shopID, xstream.Stream, msg); err != nil {
				w.logger.Error("analytics worker: failed to process message",
					zap.String("stream", xstream.Stream),
					zap.String("id", msg.ID),
					zap.Error(err))
				w.rdb.XAdd(ctx, &redis.XAddArgs{
					Stream: dlqKeyPrefix + shopID,
					MaxLen: 10_000,
					Approx: true,
					Values: msg.Values,
				})
			}
			// Prevent poison pill retry loop by acknowledging unprocessable messages after DLQ push.
			w.rdb.XAck(ctx, xstream.Stream, consumerGroup, msg.ID)
		}
	}
}

func (w *Worker) processMessage(ctx context.Context, shopID, stream string, msg redis.XMessage) error {
	raw, ok := msg.Values["data"]
	if !ok {
		return nil
	}
	eventDataStr, _ := raw.(string)
	if eventDataStr == "" {
		return nil
	}

	var evt sharedanalytics.Event
	if err := json.Unmarshal([]byte(eventDataStr), &evt); err != nil {
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
		pipe.Get(ctx, fmt.Sprintf("session_last_search:%s:%s", shopID, sid))
	}

	_, _ = pipe.Exec(ctx)

	if sid != "" && revenue > 0 {
		if q, err := w.rdb.Get(ctx, fmt.Sprintf("session_last_search:%s:%s", shopID, sid)).Result(); err == nil && q != "" {
			w.rdb.ZIncrBy(ctx, fmt.Sprintf("stats:search_revenue:%s:%s", shopID, date), float64(revenue), q)
		}
	}

	w.rdb.Del(ctx, fmt.Sprintf("forecast:revenue:%s", shopID))

	if custID != "" {
		w.writeCohort(ctx, shopID, custID, date)
	}
}

func (w *Worker) writeCohort(ctx context.Context, shopID, customerID, date string) {
	firstOrderKey := fmt.Sprintf("customer_first_order:%s:%s", shopID, customerID)
	firstMonth, err := w.rdb.Get(ctx, firstOrderKey).Result()
	if err != nil {
		yearMonth := date[:7]
		w.rdb.Set(ctx, firstOrderKey, yearMonth, 63072000*time.Second)
		w.rdb.SAdd(ctx, fmt.Sprintf("cohort:%s:%s", shopID, yearMonth), customerID)
		return
	}
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
	forecastJSON, _ := json.Marshal(forecast)
	w.rdb.Set(ctx, fmt.Sprintf("forecast:revenue:%s", shopID), string(forecastJSON), forecastCacheTTL)
}

func (w *Worker) forecastTraffic(ctx context.Context, shopID string) {
	series := w.readDailySeries(ctx, "stats:visits", shopID, 90)
	if countNonZero(series) < 14 {
		return
	}
	forecast := HoltWinters(series, 30, 0.3, 0.1)
	trafficJSON, _ := json.Marshal(forecast)
	w.rdb.Set(ctx, fmt.Sprintf("forecast:traffic:%s", shopID), string(trafficJSON), forecastCacheTTL)
}

func (w *Worker) forecastInventory(ctx context.Context, shopID string) {
	shopUUID, err := uuid.Parse(shopID)
	if err != nil {
		return
	}

	products, err := w.db.ListProducts(ctx, db.ListProductsParams{
		ShopID: pgtype.UUID{Bytes: shopUUID, Valid: true},
		Limit:  1000,
	})
	if err != nil {
		w.logger.Error("forecastInventory: list products", zap.Error(err))
		return
	}

	now := time.Now().UTC()
	for _, p := range products {
		if !p.TrackInventory.Valid || !p.TrackInventory.Bool || p.StockQuantity < 0 {
			continue
		}

		pid := p.ID.String()
		pipe := w.rdb.Pipeline()
		cmds := make([]*redis.FloatCmd, 30)
		for i := 0; i < 30; i++ {
			d := now.AddDate(0, 0, -i).Format("2006-01-02")
			cmds[i] = pipe.ZScore(ctx, fmt.Sprintf("stats:product_sales:%s:%s", shopID, d), pid)
		}
		_, _ = pipe.Exec(ctx)

		var totalRev float64
		salesDays := 0
		for _, cmd := range cmds {
			if v, err := cmd.Result(); err == nil && v > 0 {
				totalRev += v
				salesDays++
			}
		}

		avgDailySales := 0.0
		var price float64
		if f, err := p.Price.Float64Value(); err == nil && f.Valid {
			price = f.Float64
		}

		if salesDays > 0 && price > 0 {
			avgDailySales = (totalRev / price) / 30.0
		}
		if avgDailySales < 0.1 {
			avgDailySales = 0.1
		}

		daysUntilDepletion := float64(p.StockQuantity) / avgDailySales

		entry := map[string]any{
			"product_id":           pid,
			"title":                p.Title,
			"stock_quantity":       p.StockQuantity,
			"avg_daily_sales":      avgDailySales,
			"days_until_depletion": daysUntilDepletion,
		}

		if entryJSON, err := json.Marshal(entry); err == nil {
			w.rdb.Set(ctx, fmt.Sprintf("forecast:inventory:%s:%s", shopID, pid), entryJSON, inventoryCacheTTL)
		}
	}
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

func (w *Worker) runDailyDigest(ctx context.Context) {
	shops, err := w.db.ListShopsByStatus(ctx, db.NullShopStatus{ShopStatus: db.ShopStatusActive, Valid: true})
	if err != nil {
		w.logger.Warn("runDailyDigest: could not list shops", zap.Error(err))
		return
	}
	
	yesterday := time.Now().UTC().AddDate(0, 0, -1).Format("2006-01-02")
	
	for _, shop := range shops {
		shopID := shop.ID.String()
		user, err := w.db.GetUser(ctx, shop.OwnerUserID)
		if err != nil {
			continue
		}
		
		visitsStr := w.rdb.Get(ctx, fmt.Sprintf("stats:visits:%s:%s", shopID, yesterday)).Val()
		revenueStr := w.rdb.Get(ctx, fmt.Sprintf("stats:revenue:%s:%s", shopID, yesterday)).Val()
		uniqueStr := w.rdb.Get(ctx, fmt.Sprintf("stats:unique_visitors:%s:%s", shopID, yesterday)).Val()
		
		visits, _ := strconv.ParseInt(visitsStr, 10, 64)
		unique, _ := strconv.ParseInt(uniqueStr, 10, 64)
		
		if revenueStr == "" {
			revenueStr = "0.00"
		}
		
		digestPayload := email.DigestData{
			ShopName:       shop.Name,
			Date:           yesterday,
			TodayVisits:    visits,
			UniqueVisitors: unique,
			Revenue:        revenueStr,
		}
		
		if err := w.mailer.SendDailyDigest(user.Email, digestPayload); err != nil {
			w.logger.Error("failed to send daily digest", zap.String("shop_id", shopID), zap.Error(err))
		}
	}
}

func (w *Worker) runABTestAutoConclusion(ctx context.Context) {
	tests, err := w.db.ListABTestsForAutoConclusion(ctx)
	if err != nil {
		w.logger.Error("ab test auto conclude: list", zap.Error(err))
		return
	}
	for _, t := range tests {
		testIDStr := t.ID.String()
		shopIDStr := t.ShopID.String()

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

func countNonZero(s []float64) int {
	n := 0
	for _, v := range s {
		if v != 0 {
			n++
		}
	}
	return n
}

func (w *Worker) RecordImpression(ctx context.Context, testID, variant string) {
	w.rdb.Incr(ctx, fmt.Sprintf("abtest:%s:variant_%s:impressions", testID, variant))
}

func (w *Worker) RecordConversion(ctx context.Context, testID, variant string) {
	w.rdb.Incr(ctx, fmt.Sprintf("abtest:%s:variant_%s:conversions", testID, variant))
}

func (w *Worker) runAlertEngine(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
		
		shops, err := w.db.ListShopsByStatus(ctx, db.NullShopStatus{ShopStatus: db.ShopStatusActive, Valid: true})
		if err != nil {
			continue
		}
		
		today := time.Now().UTC().Format("2006-01-02")
		yesterday := time.Now().UTC().AddDate(0, 0, -1).Format("2006-01-02")
		
		for _, shop := range shops {
			shopID := shop.ID.String()
			
			revTodayStr := w.rdb.Get(ctx, fmt.Sprintf("stats:revenue:%s:%s", shopID, today)).Val()
			revYestStr := w.rdb.Get(ctx, fmt.Sprintf("stats:revenue:%s:%s", shopID, yesterday)).Val()
			
			revToday, _ := strconv.ParseFloat(revTodayStr, 64)
			revYest, _ := strconv.ParseFloat(revYestStr, 64)
			
			if revYest > 100.0 && revToday < revYest*0.2 {
				w.logger.Warn("ALERT: Revenue dropped unexpectedly", zap.String("shop_id", shopID), zap.Float64("today", revToday), zap.Float64("yesterday", revYest))
			}
			
			errTodayStr := w.rdb.Get(ctx, fmt.Sprintf("stats:errors:%s:%s", shopID, today)).Val()
			errToday, _ := strconv.ParseInt(errTodayStr, 10, 64)
			if errToday > 50 {
				w.logger.Warn("ALERT: Server errors spiked", zap.String("shop_id", shopID), zap.Int64("errors", errToday))
			}
		}
	}
}
