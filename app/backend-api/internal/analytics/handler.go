package analytics

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"backend-api/internal/ai"
	"backend-api/internal/shared/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

// Handler serves the analytics read API endpoints.
type Handler struct {
	svc      *Service
	aiClient *ai.AIClient
}

func NewHandler(svc *Service, aiClient *ai.AIClient) *Handler {
	return &Handler{svc: svc, aiClient: aiClient}
}

func (h *Handler) Overview(c *gin.Context) {
	shopID := c.GetString("shop_id")
	data, err := h.svc.GetOverview(c.Request.Context(), shopID)
	if err != nil {
		response.InternalError(c)
		return
	}
	response.OK(c, data)
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

// live visitor counts every 10 seconds until the client disconnects.
func (h *Handler) LiveStream(c *gin.Context) {
	shopID := c.GetString("shop_id")
	ctx := c.Request.Context()

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	// Push immediately on connect
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
		data map[string]int64
		err  error
	)
	switch hmType {
	case "move":
		data, err = h.svc.GetHeatmapMove(c.Request.Context(), shopID, urlPath, date)
	default:
		data, err = h.svc.GetHeatmapClick(c.Request.Context(), shopID, urlPath, date)
	}
	if err != nil {
		response.InternalError(c)
		return
	}
	response.OK(c, data)
}

func (h *Handler) ScrollDepth(c *gin.Context) {
	shopID := c.GetString("shop_id")
	urlPath := c.Query("path")
	if urlPath == "" {
		urlPath = "/"
	}
	dateFrom := parseDate(c.Query("date_from"), -7)
	dateTo := parseDate(c.Query("date_to"), 0)

	data, err := h.svc.GetScrollDepth(c.Request.Context(), shopID, urlPath, dateFrom, dateTo)
	if err != nil {
		response.InternalError(c)
		return
	}
	response.OK(c, data)
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
	data, err := h.svc.GetFunnel(c.Request.Context(), shopID, dateFrom, dateTo)
	if err != nil {
		response.InternalError(c)
		return
	}
	response.OK(c, data)
}

func (h *Handler) ProductPerformance(c *gin.Context) {
	shopID := c.GetString("shop_id")
	sortBy := c.DefaultQuery("sort", "revenue")
	limit := parseLimit(c, 20)
	dateFrom := parseDate(c.Query("date_from"), -30)
	dateTo := parseDate(c.Query("date_to"), 0)
	data, err := h.svc.GetProductPerformance(c.Request.Context(), shopID, sortBy, limit, dateFrom, dateTo)
	if err != nil {
		response.InternalError(c)
		return
	}
	response.OK(c, data)
}

func (h *Handler) DeadStock(c *gin.Context) {
	shopID := c.GetString("shop_id")
	data, err := h.svc.GetDeadStock(c.Request.Context(), shopID)
	if err != nil {
		response.InternalError(c)
		return
	}
	response.OK(c, data)
}

func (h *Handler) Cohorts(c *gin.Context) {
	shopID := c.GetString("shop_id")
	dateFrom := parseDate(c.Query("date_from"), -180)
	dateTo := parseDate(c.Query("date_to"), 0)
	data, err := h.svc.GetCohorts(c.Request.Context(), shopID, dateFrom, dateTo)
	if err != nil {
		response.InternalError(c)
		return
	}
	response.OK(c, data)
}

func (h *Handler) Retention(c *gin.Context) {
	shopID := c.GetString("shop_id")
	period := c.DefaultQuery("period", "monthly")
	data, err := h.svc.GetRetention(c.Request.Context(), shopID, period)
	if err != nil {
		response.InternalError(c)
		return
	}
	response.OK(c, data)
}

func (h *Handler) AtRiskCustomers(c *gin.Context) {
	shopID := c.GetString("shop_id")
	limit := parseLimit(c, 50)
	data, err := h.svc.GetAtRiskCustomers(c.Request.Context(), shopID, limit)
	if err != nil {
		response.InternalError(c)
		return
	}
	response.OK(c, data)
}

func (h *Handler) Revenue(c *gin.Context) {
	shopID := c.GetString("shop_id")
	groupBy := c.DefaultQuery("group_by", "date")
	dateFrom := parseDate(c.Query("date_from"), -30)
	dateTo := parseDate(c.Query("date_to"), 0)
	data, err := h.svc.GetRevenue(c.Request.Context(), shopID, groupBy, dateFrom, dateTo)
	if err != nil {
		response.InternalError(c)
		return
	}
	response.OK(c, data)
}

func (h *Handler) Forecast(c *gin.Context) {
	shopID := c.GetString("shop_id")
	metric := c.DefaultQuery("metric", "revenue")
	horizon, _ := strconv.Atoi(c.DefaultQuery("horizon", "30"))
	if horizon <= 0 || horizon > 90 {
		horizon = 30
	}
	data, err := h.svc.GetForecast(c.Request.Context(), shopID, metric, horizon)
	if err != nil {
		response.InternalError(c)
		return
	}
	response.OK(c, data)
}

func (h *Handler) InventoryForecast(c *gin.Context) {
	shopID := c.GetString("shop_id")
	days, _ := strconv.Atoi(c.DefaultQuery("days_threshold", "14"))
	data, err := h.svc.GetInventoryForecast(c.Request.Context(), shopID, days)
	if err != nil {
		response.InternalError(c)
		return
	}
	response.OK(c, data)
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
	data, err := h.svc.GetABTestResults(c.Request.Context(), testID)
	if err != nil {
		response.InternalError(c)
		return
	}
	response.OK(c, data)
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

func (h *Handler) SearchTerms(c *gin.Context) {
	shopID := c.GetString("shop_id")
	sortBy := c.DefaultQuery("sort", "count")
	limit := parseLimit(c, 20)
	dateFrom := parseDate(c.Query("date_from"), -30)
	dateTo := parseDate(c.Query("date_to"), 0)
	data, err := h.svc.GetSearchTerms(c.Request.Context(), shopID, sortBy, limit, dateFrom, dateTo)
	if err != nil {
		response.InternalError(c)
		return
	}
	response.OK(c, data)
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
	advanced.POST("/ab-tests", h.CreateABTest)
	advanced.GET("/ab-tests/:id", h.GetABTest)
	advanced.POST("/ab-tests/:id/conclude", h.ConcludeABTest)
}
