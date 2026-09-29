package analytics

import (
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"time"

	"backend-api/internal/ai"
	"backend-api/internal/shared/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type Handler struct {
	svc      *Service
	aiClient *ai.AIClient
}

type TopProductItem struct {
	ID            string  `json:"id"`
	Title         string  `json:"title"`
	SKU           string  `json:"sku"`
	ImageURL      string  `json:"image_url"`
	UnitsSold     int64   `json:"units_sold"`
	Revenue       float64 `json:"revenue"`
	StockQuantity int32   `json:"stock_quantity"`
}

type LowStockItem struct {
	ID                string `json:"id"`
	Title             string `json:"title"`
	SKU               string `json:"sku"`
	ImageURL          string `json:"image_url"`
	StockQuantity     int32  `json:"stock_quantity"`
	LowStockThreshold int32  `json:"low_stock_threshold"`
}

type RecentOrderItem struct {
	ID                string `json:"id"`
	OrderNumber       int32  `json:"order_number"`
	CustomerName      string `json:"customer_name"`
	CustomerEmail     string `json:"customer_email"`
	TotalPrice        string `json:"total_price"`
	Status            string `json:"status"`
	PaymentStatus     string `json:"payment_status"`
	FulfillmentStatus string `json:"fulfillment_status"`
	CreatedAt         string `json:"created_at"`
}

type RevenueChartPoint struct {
	Date    string  `json:"date"`
	Revenue float64 `json:"revenue"`
	Orders  int64   `json:"orders"`
	Visits  int64   `json:"visits"`
}

type DashboardSummaryResponse struct {
	TotalRevenue         float64             `json:"total_revenue"`
	TodayRevenue         float64             `json:"today_revenue"`
	RevenueChange        float64             `json:"revenue_change"`
	RevenueIsPositive    bool                `json:"revenue_is_positive"`
	TotalOrders          int64               `json:"total_orders"`
	TodayOrders          int64               `json:"today_orders"`
	PendingOrders        int64               `json:"pending_orders"`
	OrdersChange         float64             `json:"orders_change"`
	OrdersIsPositive     bool                `json:"orders_is_positive"`
	ActiveCustomers      int64               `json:"active_customers"`
	CustomersChange      float64             `json:"customers_change"`
	CustomersIsPositive  bool                `json:"customers_is_positive"`
	ActiveVisitors       int64               `json:"active_visitors"`
	AverageOrderValue    float64             `json:"average_order_value"`
	ConversionRate       float64             `json:"conversion_rate"`
	ConversionChange     float64             `json:"conversion_change"`
	ConversionIsPositive bool                `json:"conversion_is_positive"`
	TopProducts          []TopProductItem    `json:"top_products"`
	LowStockProducts     []LowStockItem      `json:"low_stock_products"`
	RecentOrders         []RecentOrderItem   `json:"recent_orders"`
	RevenueChart         []RevenueChartPoint `json:"revenue_chart"`

	TotalRevenueAlt         float64             `json:"totalRevenue"`
	TotalOrdersAlt          int64               `json:"totalOrders"`
	ActiveCustomersAlt      int64               `json:"activeCustomers"`
	ActiveVisitorsAlt       int64               `json:"activeVisitors"`
	AverageOrderValueAlt    float64             `json:"averageOrderValue"`
	ConversionRateAlt       float64             `json:"conversionRate"`
	RevenueChangeAlt        float64             `json:"revenueChange"`
	RevenueIsPositiveAlt    bool                `json:"revenueIsPositive"`
	OrdersChangeAlt         float64             `json:"ordersChange"`
	OrdersIsPositiveAlt     bool                `json:"ordersIsPositive"`
	CustomersChangeAlt      float64             `json:"customersChange"`
	CustomersIsPositiveAlt  bool                `json:"customersIsPositive"`
	ConversionChangeAlt     float64             `json:"conversionChange"`
	ConversionIsPositiveAlt bool                `json:"conversionIsPositive"`
	TopProductsAlt          []TopProductItem    `json:"topProducts"`
	LowStockProductsAlt     []LowStockItem      `json:"lowStockProducts"`
	RecentOrdersAlt         []RecentOrderItem   `json:"recentOrders"`
	RevenueChartAlt         []RevenueChartPoint `json:"revenueChart"`
}

func NewHandler(svc *Service, aiClient *ai.AIClient) *Handler {
	return &Handler{svc: svc, aiClient: aiClient}
}

func (h *Handler) Overview(c *gin.Context) {
	shopID := c.GetString("shop_id")
	overviewData, err := h.svc.GetOverview(c.Request.Context(), shopID)
	if err != nil {
		response.InternalError(c)
		return
	}
	response.OK(c, overviewData)
}

func (h *Handler) Trend(c *gin.Context) {
	shopID := c.GetString("shop_id")
	days := 7
	if p := c.Query("period"); p == "30d" {
		days = 30
	}
	points, err := h.svc.GetTrend(c.Request.Context(), shopID, days)
	if err != nil {
		response.InternalError(c)
		return
	}
	response.OK(c, points)
}

func (h *Handler) Live(c *gin.Context) {
	shopID := c.GetString("shop_id")
	count, err := h.svc.GetLive(c.Request.Context(), shopID)
	if err != nil {
		response.InternalError(c)
		return
	}
	response.OK(c, gin.H{"live_visitors": count})
}

func (h *Handler) DashboardSummary(c *gin.Context) {
	shopID := c.GetString("shop_id")
	
	resp, err := h.svc.GetDashboardSummary(c.Request.Context(), shopID)
	if err != nil {
		response.InternalError(c)
		return
	}

	response.OK(c, resp)
}

func (h *Handler) Ingest(c *gin.Context) {
	var ev struct {
		Event     string         `json:"event"`
		ShopID    string         `json:"shop_id"`
		SessionID string         `json:"session_id"`
		Payload   map[string]any `json:"payload"`
	}
	if err := c.ShouldBindJSON(&ev); err != nil {
		response.BadRequest(c, "invalid payload")
		return
	}

	if ev.Event == "" || ev.ShopID == "" {
		response.BadRequest(c, "event and shop_id are required")
		return
	}

	h.svc.RecordLiveVisitor(c.Request.Context(), ev.ShopID, ev.SessionID)

	h.svc.PublishEvent(c.Request.Context(), ev.Event, ev.ShopID, ev.SessionID, ev.Payload)

	response.OK(c, gin.H{"status": "ingested"})
}

func (h *Handler) Pages(c *gin.Context) {
	shopID := c.GetString("shop_id")
	limit := parseLimit(c, 20)
	entries, err := h.svc.GetTopPages(c.Request.Context(), shopID, limit)
	if err != nil {
		response.InternalError(c)
		return
	}
	response.OK(c, entries)
}

func (h *Handler) Referrers(c *gin.Context) {
	shopID := c.GetString("shop_id")
	limit := parseLimit(c, 20)
	entries, err := h.svc.GetTopReferrers(c.Request.Context(), shopID, limit)
	if err != nil {
		response.InternalError(c)
		return
	}
	response.OK(c, entries)
}

func (h *Handler) Geo(c *gin.Context) {
	shopID := c.GetString("shop_id")
	limit := parseLimit(c, 20)
	entries, err := h.svc.GetGeo(c.Request.Context(), shopID, limit)
	if err != nil {
		response.InternalError(c)
		return
	}
	response.OK(c, entries)
}

func (h *Handler) Devices(c *gin.Context) {
	shopID := c.GetString("shop_id")
	split, err := h.svc.GetDevices(c.Request.Context(), shopID)
	if err != nil {
		response.InternalError(c)
		return
	}
	response.OK(c, split)
}

func (h *Handler) Sparkline(c *gin.Context) {
	shopID := c.GetString("shop_id")
	points, err := h.svc.GetSparkline(c.Request.Context(), shopID)
	if err != nil {
		response.InternalError(c)
		return
	}
	response.OK(c, points)
}

func (h *Handler) LiveStream(c *gin.Context) {
	shopID := c.GetString("shop_id")
	ctx := c.Request.Context()

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	if count, err := h.svc.GetLive(ctx, shopID); err == nil {
		fmt.Fprintf(c.Writer, "data: {\"live_visitors\":%d}\n\n", count)
		c.Writer.Flush()
	}

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			count, err := h.svc.GetLive(ctx, shopID)
			if err != nil {
				return
			}
			fmt.Fprintf(c.Writer, "data: {\"live_visitors\":%d}\n\n", count)
			c.Writer.Flush()
		}
	}
}


func (h *Handler) Heatmap(c *gin.Context) {
	shopID := c.GetString("shop_id")
	urlPath := c.Query("path")
	if urlPath == "" {
		urlPath = "/"
	}
	hmType := c.Query("type")
	date := c.Query("date")
	if date == "" {
		date = time.Now().UTC().Format("2006-01-02")
	}

	var (
		heatmapPoints map[string]int64
		err           error
	)
	switch hmType {
	case "move":
		heatmapPoints, err = h.svc.GetHeatmapMove(c.Request.Context(), shopID, urlPath, date)
	default:
		heatmapPoints, err = h.svc.GetHeatmapClick(c.Request.Context(), shopID, urlPath, date)
	}
	if err != nil {
		response.InternalError(c)
		return
	}
	response.OK(c, heatmapPoints)
}

func (h *Handler) ScrollDepth(c *gin.Context) {
	shopID := c.GetString("shop_id")
	urlPath := c.Query("path")
	if urlPath == "" {
		urlPath = "/"
	}
	dateFrom := parseDate(c.Query("date_from"), -7)
	dateTo := parseDate(c.Query("date_to"), 0)

	depthData, err := h.svc.GetScrollDepth(c.Request.Context(), shopID, urlPath, dateFrom, dateTo)
	if err != nil {
		response.InternalError(c)
		return
	}
	response.OK(c, depthData)
}

func parseDate(s string, offsetDays int) time.Time {
	if t, err := time.Parse("2006-01-02", s); err == nil {
		return t.UTC()
	}
	return time.Now().UTC().AddDate(0, 0, offsetDays).Truncate(24 * time.Hour)
}

func parseLimit(c *gin.Context, def int) int {
	if s := c.Query("limit"); s != "" {
		if n, err := strconv.Atoi(s); err == nil && n > 0 && n <= 100 {
			return n
		}
	}
	return def
}


func (h *Handler) FunnelInsights(c *gin.Context) {
	shopID := c.GetString("shop_id")
	shopUUID, err := uuid.Parse(shopID)
	if err != nil {
		response.BadRequest(c, "invalid shop id")
		return
	}
	
	dateFrom := parseDate(c.Query("date_from"), -7)
	dateTo := parseDate(c.Query("date_to"), 0)

	start := pgtype.Timestamptz{Time: dateFrom, Valid: true}
	end := pgtype.Timestamptz{Time: dateTo, Valid: true}

	insights, err := h.svc.GenerateFunnelInsights(c.Request.Context(), shopUUID, start, end, h.aiClient)
	if err != nil {
		response.InternalError(c)
		return
	}
	response.OK(c, insights)
}

func (h *Handler) Funnel(c *gin.Context) {
	shopID := c.GetString("shop_id")
	dateFrom := parseDate(c.Query("date_from"), -7)
	dateTo := parseDate(c.Query("date_to"), 0)
	funnelData, err := h.svc.GetFunnel(c.Request.Context(), shopID, dateFrom, dateTo)
	if err != nil {
		response.InternalError(c)
		return
	}
	response.OK(c, funnelData)
}

func (h *Handler) ProductPerformance(c *gin.Context) {
	shopID := c.GetString("shop_id")
	sortBy := c.DefaultQuery("sort", "revenue")
	limit := parseLimit(c, 20)
	dateFrom := parseDate(c.Query("date_from"), -30)
	dateTo := parseDate(c.Query("date_to"), 0)
	performanceData, err := h.svc.GetProductPerformance(c.Request.Context(), shopID, sortBy, limit, dateFrom, dateTo)
	if err != nil {
		response.InternalError(c)
		return
	}
	response.OK(c, performanceData)
}

func (h *Handler) DeadStock(c *gin.Context) {
	shopID := c.GetString("shop_id")
	deadStockData, err := h.svc.GetDeadStock(c.Request.Context(), shopID)
	if err != nil {
		response.InternalError(c)
		return
	}
	response.OK(c, deadStockData)
}

func (h *Handler) Cohorts(c *gin.Context) {
	shopID := c.GetString("shop_id")
	dateFrom := parseDate(c.Query("date_from"), -180)
	dateTo := parseDate(c.Query("date_to"), 0)
	cohortData, err := h.svc.GetCohorts(c.Request.Context(), shopID, dateFrom, dateTo)
	if err != nil {
		response.InternalError(c)
		return
	}
	response.OK(c, cohortData)
}

func (h *Handler) Retention(c *gin.Context) {
	shopID := c.GetString("shop_id")
	period := c.DefaultQuery("period", "monthly")
	retentionData, err := h.svc.GetRetention(c.Request.Context(), shopID, period)
	if err != nil {
		response.InternalError(c)
		return
	}
	response.OK(c, retentionData)
}

func (h *Handler) AtRiskCustomers(c *gin.Context) {
	shopID := c.GetString("shop_id")
	limit := parseLimit(c, 50)
	riskData, err := h.svc.GetAtRiskCustomers(c.Request.Context(), shopID, limit)
	if err != nil {
		response.InternalError(c)
		return
	}
	response.OK(c, riskData)
}

func (h *Handler) Revenue(c *gin.Context) {
	shopID := c.GetString("shop_id")
	groupBy := c.DefaultQuery("group_by", "date")
	dateFrom := parseDate(c.Query("date_from"), -30)
	dateTo := parseDate(c.Query("date_to"), 0)
	revenueData, err := h.svc.GetRevenue(c.Request.Context(), shopID, groupBy, dateFrom, dateTo)
	if err != nil {
		response.InternalError(c)
		return
	}
	response.OK(c, revenueData)
}

func (h *Handler) Forecast(c *gin.Context) {
	shopID := c.GetString("shop_id")
	metric := c.DefaultQuery("metric", "revenue")
	horizon, _ := strconv.Atoi(c.DefaultQuery("horizon", "30"))
	if horizon <= 0 || horizon > 90 {
		horizon = 30
	}
	forecastData, err := h.svc.GetForecast(c.Request.Context(), shopID, metric, horizon)
	if err != nil {
		response.InternalError(c)
		return
	}
	response.OK(c, forecastData)
}

func (h *Handler) InventoryForecast(c *gin.Context) {
	shopID := c.GetString("shop_id")
	days, _ := strconv.Atoi(c.DefaultQuery("days_threshold", "14"))
	inventoryData, err := h.svc.GetInventoryForecast(c.Request.Context(), shopID, days)
	if err != nil {
		response.InternalError(c)
		return
	}
	response.OK(c, inventoryData)
}

func (h *Handler) CreateABTest(c *gin.Context) {
	shopID := c.GetString("shop_id")
	var req struct {
		Name         string         `json:"name"`
		Description  string         `json:"description"`
		VariantA     map[string]any `json:"variant_a"`
		VariantB     map[string]any `json:"variant_b"`
		Metric       string         `json:"metric"`
		TrafficSplit int            `json:"traffic_split"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	varA, _ := json.Marshal(req.VariantA)
	varB, _ := json.Marshal(req.VariantB)
	testID, err := h.svc.CreateABTest(c.Request.Context(), shopID, req.Name, req.Description, varA, varB, req.Metric, int32(req.TrafficSplit))
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	c.JSON(201, gin.H{"test_id": testID})
}

func (h *Handler) GetABTest(c *gin.Context) {
	testID := c.Param("id")
	if testID == "" {
		c.JSON(400, gin.H{"error": "missing test id"})
		return
	}
	testResults, err := h.svc.GetABTestResults(c.Request.Context(), testID)
	if err != nil {
		response.InternalError(c)
		return
	}
	response.OK(c, testResults)
}

func (h *Handler) ConcludeABTest(c *gin.Context) {
	shopID := c.GetString("shop_id")
	testID := c.Param("id")
	if testID == "" {
		c.JSON(400, gin.H{"error": "missing test id"})
		return
	}
	if err := h.svc.ConcludeABTest(c.Request.Context(), shopID, testID); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"status": "concluded"})
}

func (h *Handler) ListABTests(c *gin.Context) {
	shopID := c.GetString("shop_id")
	tests, err := h.svc.ListABTests(c.Request.Context(), shopID)
	if err != nil {
		response.InternalError(c)
		return
	}
	type ABTestItem struct {
		ID               string  `json:"id"`
		Name             string  `json:"name"`
		Description      string  `json:"description"`
		Metric           string  `json:"metric"`
		TrafficSplit     int32   `json:"traffic_split"`
		Status           string  `json:"status"`
		CreatedAt        string  `json:"created_at"`
		Participants     int64   `json:"participants"`
		ConversionUplift float64 `json:"conversion_uplift"`
		Winner           string  `json:"winner,omitempty"`
	}

	abTestList := make([]ABTestItem, 0, len(tests))
	for _, t := range tests {
		desc := ""
		if t.Description.Valid {
			desc = t.Description.String
		}
		var createdAt string
		if t.CreatedAt.Valid {
			createdAt = t.CreatedAt.Time.Format(time.RFC3339)
		}

		item := ABTestItem{
			ID:           t.ID.String(),
			Name:         t.Name,
			Description:  desc,
			Metric:       t.Metric,
			TrafficSplit: t.TrafficSplit,
			Status:       t.Status,
			CreatedAt:    createdAt,
		}

		if testResults, err := h.svc.GetABTestResults(c.Request.Context(), t.ID.String()); err == nil && testResults != nil {
			item.Participants = testResults.TotalImpressions
			if testResults.VariantA.ConvRate > 0 || testResults.VariantB.ConvRate > 0 {
				item.ConversionUplift = math.Round((testResults.VariantB.ConvRate-testResults.VariantA.ConvRate)*10) / 10
			}
			item.Winner = testResults.Winner
		}
		abTestList = append(abTestList, item)
	}

	response.OK(c, abTestList)
}

func (h *Handler) SearchTerms(c *gin.Context) {
	shopID := c.GetString("shop_id")
	sortBy := c.DefaultQuery("sort", "count")
	limit := parseLimit(c, 20)
	dateFrom := parseDate(c.Query("date_from"), -30)
	dateTo := parseDate(c.Query("date_to"), 0)
	searchData, err := h.svc.GetSearchTerms(c.Request.Context(), shopID, sortBy, limit, dateFrom, dateTo)
	if err != nil {
		response.InternalError(c)
		return
	}
	response.OK(c, searchData)
}

func (h *Handler) SocialStats(c *gin.Context) {
	shopID := c.GetString("shop_id")
	start := c.DefaultQuery("start", time.Now().AddDate(0, 0, -30).Format("2006-01-02"))
	end := c.DefaultQuery("end", time.Now().Format("2006-01-02"))

	stats, err := h.svc.GetSocialStats(c.Request.Context(), shopID, start, end)
	if err != nil {
		response.InternalError(c)
		return
	}
	response.OK(c, stats)
}

// Each sub-group is gated by a feature flag via requireFeature.
func (h *Handler) RegisterRoutes(group *gin.RouterGroup, requireFeature func(string) gin.HandlerFunc) {
	core := group.Group("", requireFeature("core_analytics"))
	core.GET("/overview", h.Overview)
	core.GET("/trend", h.Trend)
	core.GET("/live", h.Live)
	core.GET("/dashboard-summary", h.DashboardSummary)
	core.GET("/export", h.Export)

	standard := group.Group("", requireFeature("standard_analytics"))
	standard.GET("/pages", h.Pages)
	standard.GET("/referrers", h.Referrers)
	standard.GET("/devices", h.Devices)
	standard.GET("/geo", h.Geo)
	standard.GET("/funnel", h.Funnel)
	standard.GET("/products/search-terms", h.SearchTerms)
	standard.GET("/social", h.SocialStats)

	realtime := group.Group("", requireFeature("realtime_analytics"))
	realtime.GET("/live/stream", h.LiveStream)
	realtime.GET("/sparkline", h.Sparkline)

	advanced := group.Group("", requireFeature("advanced_analytics"))
	advanced.GET("/heatmap", h.Heatmap)
	advanced.GET("/funnel/insights", h.FunnelInsights)
	advanced.GET("/heatmap/scroll", h.ScrollDepth)
	advanced.GET("/products/performance", h.ProductPerformance)
	advanced.GET("/products/dead-stock", h.DeadStock)
	advanced.GET("/cohorts", h.Cohorts)
	advanced.GET("/retention", h.Retention)
	advanced.GET("/customers/at-risk", h.AtRiskCustomers)
	advanced.GET("/revenue", h.Revenue)
	advanced.GET("/forecast", h.Forecast)
	advanced.GET("/inventory/forecast", h.InventoryForecast)
	advanced.GET("/ab-tests", h.ListABTests)
	advanced.POST("/ab-tests", h.CreateABTest)
	advanced.GET("/ab-tests/:id", h.GetABTest)
	advanced.POST("/ab-tests/:id/conclude", h.ConcludeABTest)
}

func (h *Handler) Export(c *gin.Context) {
	shopID := c.GetString("shop_id")
	csvBytes, err := h.svc.ExportOrdersCSV(c.Request.Context(), shopID)
	if err != nil {
		response.InternalError(c)
		return
	}
	c.Header("Content-Disposition", "attachment; filename=export.csv")
	c.Data(200, "text/csv", csvBytes)
}
