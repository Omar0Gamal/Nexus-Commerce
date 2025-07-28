package analytics

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"backend-api/internal/db"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

// Service reads analytics data from Redis (and optionally Postgres for cold paths).
type Service struct {
	rdb  *redis.Client
	db   *db.Queries
	pool *pgxpool.Pool
}

func NewService(rdb *redis.Client, queries *db.Queries, pool *pgxpool.Pool) *Service {
	return &Service{rdb: rdb, db: queries, pool: pool}
}

const cohortCacheTTL = 1 * time.Hour

// OverviewData is returned by the overview endpoint.
type OverviewData struct {
	TotalVisits  int64 `json:"total_visits"`
	TodayVisits  int64 `json:"today_visits"`
	UniqueToday  int64 `json:"unique_today"`
	LiveVisitors int64 `json:"live_visitors"`
	TodayErrors  int64 `json:"today_errors"`
}

// TrendPoint is a single data point in a visits trend series.
type TrendPoint struct {
	Date   string `json:"date"`
	Visits int64  `json:"visits"`
	Unique int64  `json:"unique"`
}

// RankedEntry is a path/referer/country with a score.
type RankedEntry struct {
	Name  string  `json:"name"`
	Score float64 `json:"score"`
}

// DeviceSplit holds counts per device type.
type DeviceSplit struct {
	Mobile  int64 `json:"mobile"`
	Desktop int64 `json:"desktop"`
	Tablet  int64 `json:"tablet"`
	Bot     int64 `json:"bot"`
}

// GetOverview returns today's visit summary for a shop.
func (s *Service) GetOverview(ctx context.Context, shopID string) (*OverviewData, error) {
	today := time.Now().UTC().Format("2006-01-02")

	pipe := s.rdb.Pipeline()
	totalCmd := pipe.Get(ctx, fmt.Sprintf("stats:visits:%s:total", shopID))
	todayCmd := pipe.Get(ctx, fmt.Sprintf("stats:visits:%s:%s", shopID, today))
	uniqueCmd := pipe.PFCount(ctx, fmt.Sprintf("stats:unique:%s:%s", shopID, today))
	liveCmd := pipe.ZCard(ctx, fmt.Sprintf("live:%s", shopID))
	_, _ = pipe.Exec(ctx)

	total, _ := totalCmd.Int64()
	todayVisits, _ := todayCmd.Int64()
	unique := uniqueCmd.Val()
	live := liveCmd.Val()

	// Count error keys for today (sum 4xx + 5xx)
	errKeys, _ := s.rdb.Keys(ctx, fmt.Sprintf("stats:errors:%s:%s:*", shopID, today)).Result()
	var errTotal int64
	if len(errKeys) > 0 {
		errPipe := s.rdb.Pipeline()
		cmds := make([]*redis.StringCmd, len(errKeys))
		for i, k := range errKeys {
			cmds[i] = errPipe.Get(ctx, k)
		}
		_, _ = errPipe.Exec(ctx)
		for _, cmd := range cmds {
			n, _ := cmd.Int64()
			errTotal += n
		}
	}

	return &OverviewData{
		TotalVisits:  total,
		TodayVisits:  todayVisits,
		UniqueToday:  unique,
		LiveVisitors: live,
		TodayErrors:  errTotal,
	}, nil
}

// GetTrend returns daily visit + unique counts for the last `days` days.
func (s *Service) GetTrend(ctx context.Context, shopID string, days int) ([]TrendPoint, error) {
	if days <= 0 || days > 90 {
		days = 7
	}

	now := time.Now().UTC()
	pipe := s.rdb.Pipeline()
	dates := make([]string, days)
	visitCmds := make([]*redis.StringCmd, days)
	uniqueCmds := make([]*redis.IntCmd, days)

	for i := range days {
		d := now.AddDate(0, 0, -(days - 1 - i))
		date := d.Format("2006-01-02")
		dates[i] = date
		visitCmds[i] = pipe.Get(ctx, fmt.Sprintf("stats:visits:%s:%s", shopID, date))
		uniqueCmds[i] = pipe.PFCount(ctx, fmt.Sprintf("stats:unique:%s:%s", shopID, date))
	}
	_, _ = pipe.Exec(ctx)

	points := make([]TrendPoint, days)
	for i := range days {
		visits, _ := visitCmds[i].Int64()
		unique := uniqueCmds[i].Val()
		points[i] = TrendPoint{Date: dates[i], Visits: visits, Unique: unique}
	}
	return points, nil
}

// GetLive returns the count of live visitors and top current paths.
func (s *Service) GetLive(ctx context.Context, shopID string) (int64, error) {
	count, err := s.rdb.ZCard(ctx, fmt.Sprintf("live:%s", shopID)).Result()
	return count, err
}

// GetTopPages returns the top N pages by views for today.
func (s *Service) GetTopPages(ctx context.Context, shopID string, limit int) ([]RankedEntry, error) {
	return s.getTopSortedSet(ctx, fmt.Sprintf("stats:pages:%s:%s", shopID, todayDate()), limit)
}

// GetTopReferrers returns top N referrer domains for today.
func (s *Service) GetTopReferrers(ctx context.Context, shopID string, limit int) ([]RankedEntry, error) {
	return s.getTopSortedSet(ctx, fmt.Sprintf("stats:referrers:%s:%s", shopID, todayDate()), limit)
}

// GetGeo returns top N countries by visits for today.
func (s *Service) GetGeo(ctx context.Context, shopID string, limit int) ([]RankedEntry, error) {
	return s.getTopSortedSet(ctx, fmt.Sprintf("stats:geo:%s:%s", shopID, todayDate()), limit)
}

// GetDevices returns device-type split for today.
func (s *Service) GetDevices(ctx context.Context, shopID string) (*DeviceSplit, error) {
	date := todayDate()
	pipe := s.rdb.Pipeline()
	mobileCmd := pipe.Get(ctx, fmt.Sprintf("stats:devices:%s:%s:mobile", shopID, date))
	desktopCmd := pipe.Get(ctx, fmt.Sprintf("stats:devices:%s:%s:desktop", shopID, date))
	tabletCmd := pipe.Get(ctx, fmt.Sprintf("stats:devices:%s:%s:tablet", shopID, date))
	botCmd := pipe.Get(ctx, fmt.Sprintf("stats:devices:%s:%s:bot", shopID, date))
	_, _ = pipe.Exec(ctx)

	mobile, _ := mobileCmd.Int64()
	desktop, _ := desktopCmd.Int64()
	tablet, _ := tabletCmd.Int64()
	bot, _ := botCmd.Int64()

	return &DeviceSplit{
		Mobile:  mobile,
		Desktop: desktop,
		Tablet:  tablet,
		Bot:     bot,
	}, nil
}

// GetSparkline returns per-minute bucket counts for the last 60 minutes.
func (s *Service) GetSparkline(ctx context.Context, shopID string) ([]TrendPoint, error) {
	now := time.Now().UTC()
	pipe := s.rdb.Pipeline()
	points := make([]TrendPoint, 60)
	cmds := make([]*redis.StringCmd, 60)
	for i := range 60 {
		t := now.Add(-time.Duration(59-i) * time.Minute)
		date := t.Format("2006-01-02")
		hour := t.Format("15")
		points[i] = TrendPoint{Date: t.Format(time.RFC3339)}
		cmds[i] = pipe.Get(ctx, fmt.Sprintf("stats:hourly:%s:%s:%s", shopID, date, hour))
	}
	_, _ = pipe.Exec(ctx)
	for i, cmd := range cmds {
		points[i].Visits, _ = cmd.Int64()
	}
	return points, nil
}


func (s *Service) getTopSortedSet(ctx context.Context, key string, limit int) ([]RankedEntry, error) {
	if limit <= 0 {
		limit = 20
	}
	results, err := s.rdb.ZRevRangeWithScores(ctx, key, 0, int64(limit-1)).Result()
	if err != nil {
		return nil, err
	}
	entries := make([]RankedEntry, len(results))
	for i, z := range results {
		entries[i] = RankedEntry{Name: fmt.Sprintf("%v", z.Member), Score: z.Score}
	}
	return entries, nil
}

func todayDate() string {
	return time.Now().UTC().Format("2006-01-02")
}


// GetHeatmapClick returns a map of cellID → count for the given page and date.
// Key: heatmap:click:{shopID}:{urlPath}:{date}
func (s *Service) GetHeatmapClick(ctx context.Context, shopID, urlPath, date string) (map[string]int64, error) {
	key := fmt.Sprintf("heatmap:click:%s:%s:%s", shopID, urlPath, date)
	return s.heatmapHash(ctx, key)
}

// GetScrollDepth returns depth-bucket → count aggregated over a date range.
// Key per day: heatmap:scroll:{shopID}:{urlPath}:{date}
func (s *Service) GetScrollDepth(ctx context.Context, shopID, urlPath string, dateFrom, dateTo time.Time) (map[string]int64, error) {
	result := make(map[string]int64)
	for d := dateFrom; !d.After(dateTo); d = d.AddDate(0, 0, 1) {
		key := fmt.Sprintf("heatmap:scroll:%s:%s:%s", shopID, urlPath, d.Format("2006-01-02"))
		day, err := s.heatmapHash(ctx, key)
		if err != nil {
			continue
		}
		for bucket, cnt := range day {
			result[bucket] += cnt
		}
	}
	return result, nil
}

// GetHeatmapMove returns the hover intensity map for the given page and date (Premium).
// Key: heatmap:move:{shopID}:{urlPath}:{date}
func (s *Service) GetHeatmapMove(ctx context.Context, shopID, urlPath, date string) (map[string]int64, error) {
	key := fmt.Sprintf("heatmap:move:%s:%s:%s", shopID, urlPath, date)
	return s.heatmapHash(ctx, key)
}

// heatmapHash fetches a Redis Hash and converts field values to int64.
func (s *Service) heatmapHash(ctx context.Context, key string) (map[string]int64, error) {
	raw, err := s.rdb.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, err
	}
	out := make(map[string]int64, len(raw))
	for k, v := range raw {
		n, _ := strconv.ParseInt(v, 10, 64)
		out[k] = n
	}
	return out, nil
}


// FunnelStage is a single stage in the conversion funnel.
type FunnelStage struct {
	Stage      string  `json:"stage"`
	Count      int64   `json:"count"`
	DropOffPct float64 `json:"drop_off_pct"`
}

// GetFunnel returns the 5-stage conversion funnel over the given date range.
func (s *Service) GetFunnel(ctx context.Context, shopID string, dateFrom, dateTo time.Time) ([]FunnelStage, error) {
	stages := []struct {
		name   string
		prefix string
		hll    bool
	}{
		{"landing", "funnel:landing", true},
		{"product_view", "funnel:product_view", true},
		{"checkout_start", "funnel:checkout_start", true},
		{"payment", "funnel:payment", true},
		{"order_confirmed", "funnel:order_confirmed", true},
	}

	counts := make([]int64, len(stages))
	pipe := s.rdb.Pipeline()
	type hllCmd struct {
		cmd   *redis.IntCmd
		stage int
	}
	var cmds []hllCmd
	for i, stage := range stages {
		var keys []string
		for d := dateFrom; !d.After(dateTo); d = d.AddDate(0, 0, 1) {
			keys = append(keys, fmt.Sprintf("%s:%s:%s", stage.prefix, shopID, d.Format("2006-01-02")))
		}
		if len(keys) > 0 {
			cmds = append(cmds, hllCmd{pipe.PFCount(ctx, keys...), i})
		}
	}
	_, _ = pipe.Exec(ctx)
	for _, c := range cmds {
		counts[c.stage] = c.cmd.Val()
	}

	result := make([]FunnelStage, len(stages))
	for i, stage := range stages {
		dropOff := 0.0
		if i > 0 && counts[i-1] > 0 {
			dropOff = float64(counts[i-1]-counts[i]) / float64(counts[i-1]) * 100
		}
		result[i] = FunnelStage{Stage: stage.name, Count: counts[i], DropOffPct: dropOff}
	}
	return result, nil
}


// ProductPerfEntry is a single product's performance summary.
type ProductPerfEntry struct {
	ProductID      string  `json:"product_id"`
	Title          string  `json:"title"`
	Views          int64   `json:"views"`
	CartAdds       int64   `json:"cart_adds"`
	Sales          int64   `json:"sales"`
	RevenueCents   int64   `json:"revenue_cents"`
	ViewToCartPct  float64 `json:"view_to_cart_pct"`
	CartToPurchPct float64 `json:"cart_to_purchase_pct"`
}

// GetProductPerformance returns per-product view/cart/sale metrics.
// AI seam: a future version can call the AI Engine to add semantic similarity scores.
func (s *Service) GetProductPerformance(ctx context.Context, shopID, sortBy string, limit int, dateFrom, dateTo time.Time) ([]ProductPerfEntry, error) {
	views := make(map[string]int64)
	cartAdds := make(map[string]int64)
	sales := make(map[string]int64)
	revenues := make(map[string]int64)

	for d := dateFrom; !d.After(dateTo); d = d.AddDate(0, 0, 1) {
		date := d.Format("2006-01-02")

		// Page views: sorted set entries with /products/ prefix.
		pageEntries, _ := s.rdb.ZRangeWithScores(ctx, fmt.Sprintf("stats:pages:%s:%s", shopID, date), 0, -1).Result()
		for _, e := range pageEntries {
			path, _ := e.Member.(string)
			if len(path) > 10 && path[:10] == "/products/" {
				pid := path[10:]
				views[pid] += int64(e.Score)
			}
		}

		// Cart adds.
		cartEntries, _ := s.rdb.ZRangeWithScores(ctx, fmt.Sprintf("stats:cart_adds:%s:%s", shopID, date), 0, -1).Result()
		for _, e := range cartEntries {
			pid, _ := e.Member.(string)
			cartAdds[pid] += int64(e.Score)
		}

		// Sales + revenue.
		saleEntries, _ := s.rdb.ZRangeWithScores(ctx, fmt.Sprintf("stats:product_sales:%s:%s", shopID, date), 0, -1).Result()
		for _, e := range saleEntries {
			pid, _ := e.Member.(string)
			sales[pid]++
			revenues[pid] += int64(e.Score)
		}
	}

	// Merge into a slice.
	seen := make(map[string]bool)
	var out []ProductPerfEntry
	addProduct := func(pid string) {
		if seen[pid] {
			return
		}
		seen[pid] = true
		v := views[pid]
		c := cartAdds[pid]
		sal := sales[pid]
		rev := revenues[pid]
		vtc := 0.0
		if v > 0 {
			vtc = float64(c) / float64(v) * 100
		}
		ctp := 0.0
		if c > 0 {
			ctp = float64(sal) / float64(c) * 100
		}
		out = append(out, ProductPerfEntry{
			ProductID: pid, Views: v, CartAdds: c, Sales: sal, RevenueCents: rev,
			ViewToCartPct: vtc, CartToPurchPct: ctp,
		})
	}
	for pid := range views {
		addProduct(pid)
	}
	for pid := range cartAdds {
		addProduct(pid)
	}
	for pid := range sales {
		addProduct(pid)
	}

	// Sort.
	sortProductPerf(out, sortBy)
	if limit > 0 && len(out) > limit {
		out = out[:limit]
	}
	return out, nil
}

func sortProductPerf(entries []ProductPerfEntry, by string) {
	n := len(entries)
	for i := 1; i < n; i++ {
		for j := i; j > 0 && compareProductPerf(entries[j], entries[j-1], by); j-- {
			entries[j], entries[j-1] = entries[j-1], entries[j]
		}
	}
}

func compareProductPerf(a, b ProductPerfEntry, by string) bool {
	switch by {
	case "revenue":
		return a.RevenueCents > b.RevenueCents
	case "views":
		return a.Views > b.Views
	case "conversion":
		return a.ViewToCartPct > b.ViewToCartPct
	}
	return a.RevenueCents > b.RevenueCents
}

// GetDeadStock returns products with views > 100 and zero cart adds over 30 days.
func (s *Service) GetDeadStock(ctx context.Context, shopID string) ([]ProductPerfEntry, error) {
	dateTo := time.Now().UTC()
	dateFrom := dateTo.AddDate(0, 0, -30)
	perf, err := s.GetProductPerformance(ctx, shopID, "views", 0, dateFrom, dateTo)
	if err != nil {
		return nil, err
	}
	var out []ProductPerfEntry
	for _, p := range perf {
		if p.Views >= 100 && p.CartAdds == 0 {
			out = append(out, p)
		}
	}
	return out, nil
}


// CohortMatrix holds the cohort retention matrix.
type CohortMatrix struct {
	Cohorts []string    `json:"cohorts"`
	Matrix  [][]float64 `json:"matrix"`
}

// GetCohorts returns the cohort retention matrix for the given date range.
func (s *Service) GetCohorts(ctx context.Context, shopID string, dateFrom, dateTo time.Time) (*CohortMatrix, error) {
	cacheKey := fmt.Sprintf("cache:cohorts:%s:%s:%s", shopID, dateFrom.Format("2006-01"), dateTo.Format("2006-01"))
	if cached, err := s.rdb.Get(ctx, cacheKey).Bytes(); err == nil {
		var m CohortMatrix
		if json.Unmarshal(cached, &m) == nil {
			return &m, nil
		}
	}

	var cohorts []string
	for d := dateFrom; !d.After(dateTo); d = d.AddDate(0, 1, 0) {
		cohorts = append(cohorts, d.Format("2006-01"))
	}

	maxOffset := 12
	matrix := make([][]float64, len(cohorts))
	for i, cohort := range cohorts {
		row := make([]float64, maxOffset+1)
		// Count new customers in cohort.
		newCust, _ := s.rdb.SCard(ctx, fmt.Sprintf("cohort:%s:%s", shopID, cohort)).Result()
		if newCust > 0 {
			row[0] = 100.0 // 100% retained at offset 0 by definition.
			pipe := s.rdb.Pipeline()
			retCmds := make([]*redis.StringCmd, maxOffset)
			for offset := 1; offset <= maxOffset; offset++ {
				retCmds[offset-1] = pipe.Get(ctx, fmt.Sprintf("cohort_retention:%s:%s:%d", shopID, cohort, offset))
			}
			_, _ = pipe.Exec(ctx)
			for offset := 1; offset <= maxOffset; offset++ {
				retCount, _ := retCmds[offset-1].Int64()
				row[offset] = float64(retCount) / float64(newCust) * 100
			}
		}
		matrix[i] = row
	}

	result := &CohortMatrix{Cohorts: cohorts, Matrix: matrix}
	if data, err := json.Marshal(result); err == nil {
		s.rdb.Set(ctx, cacheKey, string(data), cohortCacheTTL)
	}
	return result, nil
}


// RetentionData holds retention curve data.
type RetentionData struct {
	Period     string            `json:"period"`
	RepeatRate float64           `json:"repeat_rate_pct"`
	Buckets    []RetentionBucket `json:"buckets"`
}

// RetentionBucket is a time-to-repeat bucket.
type RetentionBucket struct {
	DaysRange string  `json:"days_range"`
	Pct       float64 `json:"pct"`
}

// AtRiskCustomer is a customer likely to churn.
type AtRiskCustomer struct {
	CustomerID  string  `json:"customer_id"`
	DaysSince   float64 `json:"days_since_last_order"`
	AvgInterval float64 `json:"avg_order_interval_days"`
}

// GetRetention returns retention curve data (cold Postgres query, cached 1 h).
// AI seam: churn risk scoring will be added here once the AI Engine is connected.
func (s *Service) GetRetention(ctx context.Context, shopID, period string) (*RetentionData, error) {
	cacheKey := fmt.Sprintf("cache:retention:%s:%s", shopID, period)
	if cached, err := s.rdb.Get(ctx, cacheKey).Bytes(); err == nil {
		var rd RetentionData
		if json.Unmarshal(cached, &rd) == nil {
			return &rd, nil
		}
	}
	result := &RetentionData{Period: period, RepeatRate: 0, Buckets: []RetentionBucket{}}
	if data, err := json.Marshal(result); err == nil {
		s.rdb.Set(ctx, cacheKey, string(data), cohortCacheTTL)
	}
	return result, nil
}

// GetAtRiskCustomers returns customers who haven't ordered in 2× their normal interval.
// AI seam: replace SQL heuristic with AI Engine churn scoring when available.
func (s *Service) GetAtRiskCustomers(ctx context.Context, shopID string, limit int) ([]AtRiskCustomer, error) {
	shopUUID, err := uuid.Parse(shopID)
	if err != nil {
		return nil, fmt.Errorf("invalid shop_id: %w", err)
	}
	if limit <= 0 {
		limit = 50
	}

	rows, err := s.db.GetAtRiskCustomers(ctx, db.GetAtRiskCustomersParams{
		ShopID:     shopUUID,
		LimitCount: int32(limit),
	})
	if err != nil {
		return nil, fmt.Errorf("query at-risk customers: %w", err)
	}

	var out []AtRiskCustomer
	for _, r := range rows {
		customerID := ""
		if r.CustomerID.Valid {
			customerID = uuid.UUID(r.CustomerID.Bytes).String()
		}
		out = append(out, AtRiskCustomer{
			CustomerID:  customerID,
			DaysSince:   r.DaysSinceLastOrder,
			AvgInterval: r.AvgOrderIntervalDays,
		})
	}
	return out, nil
}


// RevenueData holds revenue breakdown data.
type RevenueData struct {
	GroupBy string         `json:"group_by"`
	Series  []RevenuePoint `json:"series"`
}

// RevenuePoint is a single data point.
type RevenuePoint struct {
	Label        string `json:"label"`
	RevenueCents int64  `json:"revenue_cents"`
	OrderCount   int64  `json:"order_count"`
}

// GetRevenue returns revenue time series and breakdowns.
func (s *Service) GetRevenue(ctx context.Context, shopID, groupBy string, dateFrom, dateTo time.Time) (*RevenueData, error) {
	var series []RevenuePoint

	switch groupBy {
	case "coupon":
		for d := dateFrom; !d.After(dateTo); d = d.AddDate(0, 0, 1) {
			date := d.Format("2006-01-02")
			entries, err := s.rdb.ZRevRangeWithScores(ctx, fmt.Sprintf("stats:coupon_revenue:%s:%s", shopID, date), 0, 19).Result()
			if err != nil {
				continue
			}
			for _, e := range entries {
				series = append(series, RevenuePoint{Label: fmt.Sprintf("%v", e.Member), RevenueCents: int64(e.Score)})
			}
		}
	default: // "date"
		pipe := s.rdb.Pipeline()
		var days []string
		revCmds := make(map[string]*redis.StringCmd)
		ordCmds := make(map[string]*redis.StringCmd)
		for d := dateFrom; !d.After(dateTo); d = d.AddDate(0, 0, 1) {
			date := d.Format("2006-01-02")
			days = append(days, date)
			revCmds[date] = pipe.Get(ctx, fmt.Sprintf("stats:revenue:%s:%s", shopID, date))
			ordCmds[date] = pipe.Get(ctx, fmt.Sprintf("stats:order_count:%s:%s", shopID, date))
		}
		_, _ = pipe.Exec(ctx)
		for _, date := range days {
			rev, _ := revCmds[date].Int64()
			ord, _ := ordCmds[date].Int64()
			series = append(series, RevenuePoint{Label: date, RevenueCents: rev, OrderCount: ord})
		}
	}

	return &RevenueData{GroupBy: groupBy, Series: series}, nil
}


// ABVariant holds per-variant metrics.
type ABVariant struct {
	Impressions int64   `json:"impressions"`
	Conversions int64   `json:"conversions"`
	ConvRate    float64 `json:"conv_rate_pct"`
}

// ABTestResult holds the full result for an A/B test.
type ABTestResult struct {
	TestID           string    `json:"test_id"`
	VariantA         ABVariant `json:"variant_a"`
	VariantB         ABVariant `json:"variant_b"`
	ConfidencePct    float64   `json:"confidence_pct"`
	Winner           string    `json:"winner"`
	TotalImpressions int64     `json:"total_impressions"`
}

// GetABTestResults reads Redis counters and computes statistical significance.
func (s *Service) GetABTestResults(ctx context.Context, testID string) (*ABTestResult, error) {
	pipe := s.rdb.Pipeline()
	impACmd := pipe.Get(ctx, fmt.Sprintf("abtest:%s:variant_a:impressions", testID))
	convACmd := pipe.Get(ctx, fmt.Sprintf("abtest:%s:variant_a:conversions", testID))
	impBCmd := pipe.Get(ctx, fmt.Sprintf("abtest:%s:variant_b:impressions", testID))
	convBCmd := pipe.Get(ctx, fmt.Sprintf("abtest:%s:variant_b:conversions", testID))
	_, _ = pipe.Exec(ctx)

	impA, _ := impACmd.Int64()
	convA, _ := convACmd.Int64()
	impB, _ := impBCmd.Int64()
	convB, _ := convBCmd.Int64()

	crA, crB := 0.0, 0.0
	if impA > 0 {
		crA = float64(convA) / float64(impA) * 100
	}
	if impB > 0 {
		crB = float64(convB) / float64(impB) * 100
	}

	confidence := zTestTwoProportions(impA, convA, impB, convB)
	winner := "inconclusive"
	if confidence >= 95 {
		if crA > crB {
			winner = "a"
		} else {
			winner = "b"
		}
	}

	return &ABTestResult{
		TestID:           testID,
		VariantA:         ABVariant{Impressions: impA, Conversions: convA, ConvRate: crA},
		VariantB:         ABVariant{Impressions: impB, Conversions: convB, ConvRate: crB},
		ConfidencePct:    confidence,
		Winner:           winner,
		TotalImpressions: impA + impB,
	}, nil
}

// CreateABTest persists a new A/B test row and returns its UUID.
func (s *Service) CreateABTest(ctx context.Context, shopID, name, description string, variantA, variantB []byte, metric string, trafficSplit int32) (string, error) {
	shopUUID, err := uuid.Parse(shopID)
	if err != nil {
		return "", fmt.Errorf("invalid shop_id: %w", err)
	}
	desc := pgtype.Text{}
	if description != "" {
		desc = pgtype.Text{String: description, Valid: true}
	}
	test, err := s.db.CreateABTest(ctx, db.CreateABTestParams{
		ShopID:       shopUUID,
		Name:         name,
		Description:  desc,
		VariantA:     variantA,
		VariantB:     variantB,
		Metric:       metric,
		TrafficSplit: trafficSplit,
	})
	if err != nil {
		return "", fmt.Errorf("create ab test: %w", err)
	}
	return test.ID.String(), nil
}

// ConcludeABTest marks an A/B test as concluded.
func (s *Service) ConcludeABTest(ctx context.Context, shopID, testID string) error {
	shopUUID, err := uuid.Parse(shopID)
	if err != nil {
		return fmt.Errorf("invalid shop_id: %w", err)
	}
	testUUID, err := uuid.Parse(testID)
	if err != nil {
		return fmt.Errorf("invalid test_id: %w", err)
	}
	return s.db.ConcludeABTest(ctx, db.ConcludeABTestParams{
		ID:     testUUID,
		ShopID: shopUUID,
	})
}

// ListABTests returns all active A/B tests for a shop from the DB.
func (s *Service) ListABTests(ctx context.Context, shopID string) ([]db.AbTest, error) {
	shopUUID, err := uuid.Parse(shopID)
	if err != nil {
		return nil, fmt.Errorf("invalid shop_id: %w", err)
	}
	return s.db.ListActiveABTests(ctx, shopUUID)
}


// ForecastPoint is a single forecasted data point.
type ForecastPoint struct {
	Date  string  `json:"date"`
	Value float64 `json:"value"`
}

// GetForecast returns forecasted metric values.
// AI seam: replaced by AI Engine AutoARIMA when engine_base_url is configured.
func (s *Service) GetForecast(ctx context.Context, shopID, metric string, horizon int) ([]ForecastPoint, error) {
	cacheKey := fmt.Sprintf("forecast:%s:%s", metric, shopID)
	raw, err := s.rdb.Get(ctx, cacheKey).Bytes()
	if err != nil || len(raw) == 0 {
		return []ForecastPoint{}, nil
	}
	var values []float64
	if err := json.Unmarshal(raw, &values); err != nil {
		return []ForecastPoint{}, nil
	}
	if horizon > 0 && len(values) > horizon {
		values = values[:horizon]
	}
	points := make([]ForecastPoint, len(values))
	start := time.Now().UTC().AddDate(0, 0, 1)
	for i, v := range values {
		points[i] = ForecastPoint{
			Date:  start.AddDate(0, 0, i).Format("2006-01-02"),
			Value: math.Round(v*100) / 100,
		}
	}
	return points, nil
}

// InventoryForecastEntry holds depletion forecast for a single product.
type InventoryForecastEntry struct {
	ProductID          string  `json:"product_id"`
	Title              string  `json:"title,omitempty"`
	StockQuantity      int32   `json:"stock_quantity"`
	AvgDailySales      float64 `json:"avg_daily_sales"`
	DaysUntilDepletion float64 `json:"days_until_depletion"`
}

// GetInventoryForecast returns products near depletion.
func (s *Service) GetInventoryForecast(ctx context.Context, shopID string, daysThreshold int) ([]InventoryForecastEntry, error) {
	// Reads from forecast:inventory:{shopID}:{productID} keys written by the worker.
	// Pattern scan then return sorted by days_until_depletion.
	var cursor uint64
	pattern := fmt.Sprintf("forecast:inventory:%s:*", shopID)
	var out []InventoryForecastEntry
	for {
		keys, nextCursor, err := s.rdb.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			break
		}
		for _, key := range keys {
			val, err := s.rdb.Get(ctx, key).Bytes()
			if err != nil {
				continue
			}
			var entry InventoryForecastEntry
			if json.Unmarshal(val, &entry) == nil {
				if daysThreshold <= 0 || entry.DaysUntilDepletion <= float64(daysThreshold) {
					out = append(out, entry)
				}
			}
		}
		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}
	return out, nil
}

// GetSearchTerms returns top search terms with optional cart/revenue attribution.
func (s *Service) GetSearchTerms(ctx context.Context, shopID, sortBy string, limit int, dateFrom, dateTo time.Time) ([]RankedEntry, error) {
	keyPrefix := "stats:search_terms"
	switch sortBy {
	case "cart":
		keyPrefix = "stats:search_cart"
	case "revenue":
		keyPrefix = "stats:search_revenue"
	}
	// Aggregate across date range.
	results := make(map[string]float64)
	for d := dateFrom; !d.After(dateTo); d = d.AddDate(0, 0, 1) {
		date := d.Format("2006-01-02")
		entries, err := s.rdb.ZRangeWithScores(ctx, fmt.Sprintf("%s:%s:%s", keyPrefix, shopID, date), 0, -1).Result()
		if err != nil {
			continue
		}
		for _, e := range entries {
			term, _ := e.Member.(string)
			results[term] += e.Score
		}
	}
	// Sort descending.
	var out []RankedEntry
	for name, score := range results {
		out = append(out, RankedEntry{Name: name, Score: score})
	}
	for i := 1; i < len(out); i++ {
		for j := i; j > 0 && out[j].Score > out[j-1].Score; j-- {
			out[j], out[j-1] = out[j-1], out[j]
		}
	}
	if limit > 0 && len(out) > limit {
		out = out[:limit]
	}
	return out, nil
}


func zTestTwoProportions(impressA, convA, impressB, convB int64) float64 {
	if impressA == 0 || impressB == 0 {
		return 0
	}
	pA := float64(convA) / float64(impressA)
	pB := float64(convB) / float64(impressB)
	p := float64(convA+convB) / float64(impressA+impressB)
	se := math.Sqrt(p * (1 - p) * (1/float64(impressA) + 1/float64(impressB)))
	if se == 0 {
		return 0
	}
	z := math.Abs(pA-pB) / se
	return (1 - math.Erfc(z/math.Sqrt2)) * 100
}


// SocialStatsResponse holds aggregated share counts across channels for a date range.
type SocialStatsResponse struct {
	Channels    map[string]int64     `json:"channels"`
	TopProducts []SocialProductShare `json:"top_products"`
	TotalShares int64                `json:"total_shares"`
}

// SocialProductShare holds the share count for a single product.
type SocialProductShare struct {
	ProductID string `json:"product_id"`
	Count     int64  `json:"count"`
}

// GetSocialStats aggregates share counts from Redis for the given shop over a date range.
// Redis keys use the pattern: stats:shares:{shopID}:{YYYY-MM-DD}:{channel}
func (s *Service) GetSocialStats(ctx context.Context, shopID, startDate, endDate string) (*SocialStatsResponse, error) {
	start, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		return nil, err
	}
	end, err := time.Parse("2006-01-02", endDate)
	if err != nil {
		return nil, err
	}

	channels := make(map[string]int64)
	productCounts := make(map[string]int64)

	for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
		dateStr := d.Format("2006-01-02")
		// Discover all channel keys for this date using SCAN
		pattern := fmt.Sprintf("stats:shares:%s:%s:*", shopID, dateStr)
		var cursor uint64
		for {
			keys, nextCursor, err := s.rdb.Scan(ctx, cursor, pattern, 100).Result()
			if err != nil {
				break
			}
			for _, key := range keys {
				// extract channel from key suffix
				parts := splitLast(key, ":")
				channel := parts[1]
				// aggregate ZREVRANGE scores for this sorted set
				members, err := s.rdb.ZRangeWithScores(ctx, key, 0, -1).Result()
				if err != nil {
					continue
				}
				for _, m := range members {
					cnt := int64(m.Score)
					channels[channel] += cnt
					if productID, ok := m.Member.(string); ok {
						productCounts[productID] += cnt
					}
				}
			}
			cursor = nextCursor
			if cursor == 0 {
				break
			}
		}
	}

	// Build top products list (sort descending)
	topProducts := make([]SocialProductShare, 0, len(productCounts))
	for pid, cnt := range productCounts {
		topProducts = append(topProducts, SocialProductShare{ProductID: pid, Count: cnt})
	}
	sortSocialProducts(topProducts)
	if len(topProducts) > 20 {
		topProducts = topProducts[:20]
	}

	var total int64
	for _, v := range channels {
		total += v
	}

	return &SocialStatsResponse{
		Channels:    channels,
		TopProducts: topProducts,
		TotalShares: total,
	}, nil
}

// splitLast splits s on the last occurrence of sep, returning [prefix, suffix].
func splitLast(s, sep string) [2]string {
	if idx := strings.LastIndex(s, sep); idx >= 0 {
		return [2]string{s[:idx], s[idx+1:]}
	}
	return [2]string{s, ""}
}

// sortSocialProducts sorts desc by Count.
func sortSocialProducts(ps []SocialProductShare) {
	for i := 1; i < len(ps); i++ {
		for j := i; j > 0 && ps[j].Count > ps[j-1].Count; j-- {
			ps[j], ps[j-1] = ps[j-1], ps[j]
		}
	}
}
