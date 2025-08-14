package ai

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"text/template"
	"time"

	"backend-api/internal/db"
	"backend-api/internal/shared/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

// coming-soon response used when a provider or engine is not yet configured.
var comingSoon = AIResponse{
	Status:  "coming_soon",
	Message: "This AI feature is not yet configured. Set the required API keys to activate it.",
}

// Handler exposes shop-facing AI endpoints.
type Handler struct {
	client  *AIClient
	prompts *PromptRegistry
	db      *db.Queries
}

func NewHandler(client *AIClient, prompts *PromptRegistry, queries *db.Queries) *Handler {
	return &Handler{client: client, prompts: prompts, db: queries}
}

// All routes require auth + staff. Feature gates are applied per-endpoint.
func (h *Handler) RegisterRoutes(
	g *gin.RouterGroup,
	requireAuth gin.HandlerFunc,
	requireStaff gin.HandlerFunc,
	requireFeature func(string) gin.HandlerFunc,
) {
	ai := g.Group("", requireAuth, requireStaff)

	ai.POST("/products/:id/description", requireFeature("ai_text"), h.GenerateDescription)
	ai.POST("/products/:id/seo-meta", requireFeature("ai_seo"), h.GenerateSEOMeta)
	ai.POST("/products/:id/translate", requireFeature("ai_translate"), h.TranslateProduct)
	ai.POST("/products/:id/titles", requireFeature("ai_text"), h.SuggestTitles)
	ai.POST("/content/quality", requireFeature("ai_text"), h.AnalyzeContentQuality)
	ai.POST("/email/copy", requireFeature("ai_email"), h.GenerateEmailCopy)
	ai.POST("/analytics/report", requireFeature("ai_reports"), h.GenerateWeeklyReport)
	ai.GET("/usage", h.GetUsageStats)
}

// -------------------------------------------------------------------------- //
//  Product endpoints                                                           //
// -------------------------------------------------------------------------- //

// GenerateDescription generates a product description using the registered LLM.
func (h *Handler) GenerateDescription(c *gin.Context) {
	shopIDStr := c.GetString("shop_id")
	productID := c.Param("id")

	var req GenerateDescriptionReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if req.MaxWords <= 0 {
		req.MaxWords = 200
	}
	if req.Language == "" {
		req.Language = "en"
	}

	product, shop, err := h.fetchProductAndShop(c.Request.Context(), shopIDStr, productID)
	if err != nil {
		response.NotFound(c, "product not found")
		return
	}

	tmpl := h.prompts.Get("product_description")
	if tmpl == nil {
		response.InternalError(c)
		return
	}

	userPrompt, err := renderTemplate(tmpl.UserPromptTmpl, map[string]any{
		"ShopName":            shop.Name,
		"Title":               product.Title,
		"Category":            "",
		"ExistingDescription": product.Description.String,
		"Tone":                req.Tone,
		"MaxWords":            req.MaxWords,
		"Language":            req.Language,
	})
	if err != nil {
		response.InternalError(c)
		return
	}

	genResp, err := h.client.GenerateText(c.Request.Context(), shopIDStr, "ai_text", GenerateRequest{
		SystemPrompt: tmpl.SystemPrompt,
		UserPrompt:   userPrompt,
		Temperature:  tmpl.Temperature,
		MaxTokens:    tmpl.MaxTokens,
	})
	if errors.Is(err, ErrNotConfigured) {
		c.JSON(http.StatusServiceUnavailable, comingSoon)
		return
	}
	if errors.Is(err, ErrQuotaExceeded) {
		c.JSON(http.StatusTooManyRequests, AIResponse{Status: "quota_exceeded", Message: "Monthly AI quota exhausted."})
		return
	}
	if err != nil {
		response.InternalError(c)
		return
	}

	response.OK(c, AIResponse{Status: "ok", Data: map[string]string{"description": genResp.Content}})
}

// GenerateSEOMeta generates SEO title and description for a product.
func (h *Handler) GenerateSEOMeta(c *gin.Context) {
	shopIDStr := c.GetString("shop_id")
	productID := c.Param("id")

	var req GenerateSEOMetaReq
	_ = c.ShouldBindJSON(&req)
	if req.Language == "" {
		req.Language = "en"
	}

	product, _, err := h.fetchProductAndShop(c.Request.Context(), shopIDStr, productID)
	if err != nil {
		response.NotFound(c, "product not found")
		return
	}

	tmpl := h.prompts.Get("seo_metadata")
	if tmpl == nil {
		response.InternalError(c)
		return
	}

	desc := product.Description.String
	if len(desc) > 300 {
		desc = desc[:300]
	}

	userPrompt, err := renderTemplate(tmpl.UserPromptTmpl, map[string]any{
		"Title":              product.Title,
		"Category":           "",
		"DescriptionExcerpt": desc,
		"Keywords":           "",
		"Language":           req.Language,
	})
	if err != nil {
		response.InternalError(c)
		return
	}

	genResp, err := h.client.GenerateText(c.Request.Context(), shopIDStr, "ai_seo", GenerateRequest{
		SystemPrompt: tmpl.SystemPrompt,
		UserPrompt:   userPrompt,
		Temperature:  tmpl.Temperature,
		MaxTokens:    tmpl.MaxTokens,
	})
	if errors.Is(err, ErrNotConfigured) {
		c.JSON(http.StatusServiceUnavailable, comingSoon)
		return
	}
	if errors.Is(err, ErrQuotaExceeded) {
		c.JSON(http.StatusTooManyRequests, AIResponse{Status: "quota_exceeded", Message: "Monthly AI quota exhausted."})
		return
	}
	if err != nil {
		response.InternalError(c)
		return
	}

	response.OK(c, AIResponse{Status: "ok", Data: map[string]string{"seo_content": genResp.Content}})
}

// TranslateProduct translates a product's title and description.
func (h *Handler) TranslateProduct(c *gin.Context) {
	shopIDStr := c.GetString("shop_id")
	productID := c.Param("id")

	var req TranslateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	product, _, err := h.fetchProductAndShop(c.Request.Context(), shopIDStr, productID)
	if err != nil {
		response.NotFound(c, "product not found")
		return
	}

	tmpl := h.prompts.Get("translation")
	if tmpl == nil {
		response.InternalError(c)
		return
	}

	userPrompt, err := renderTemplate(tmpl.UserPromptTmpl, map[string]any{
		"ProductTitle":   product.Title,
		"CurrentText":    product.Description.String,
		"TargetLanguage": req.TargetLanguage,
	})
	if err != nil {
		response.InternalError(c)
		return
	}

	genResp, err := h.client.GenerateText(c.Request.Context(), shopIDStr, "ai_translate", GenerateRequest{
		SystemPrompt: tmpl.SystemPrompt,
		UserPrompt:   userPrompt,
		Temperature:  tmpl.Temperature,
		MaxTokens:    tmpl.MaxTokens,
	})
	if errors.Is(err, ErrNotConfigured) {
		c.JSON(http.StatusServiceUnavailable, comingSoon)
		return
	}
	if errors.Is(err, ErrQuotaExceeded) {
		c.JSON(http.StatusTooManyRequests, AIResponse{Status: "quota_exceeded", Message: "Monthly AI quota exhausted."})
		return
	}
	if err != nil {
		response.InternalError(c)
		return
	}

	response.OK(c, AIResponse{Status: "ok", Data: map[string]string{"translation": genResp.Content}})
}

// SuggestTitles suggests alternative product titles.
func (h *Handler) SuggestTitles(c *gin.Context) {
	shopIDStr := c.GetString("shop_id")
	productID := c.Param("id")

	var req SuggestTitlesReq
	_ = c.ShouldBindJSON(&req)
	if req.Count <= 0 {
		req.Count = 5
	}
	if req.Language == "" {
		req.Language = "en"
	}

	product, _, err := h.fetchProductAndShop(c.Request.Context(), shopIDStr, productID)
	if err != nil {
		response.NotFound(c, "product not found")
		return
	}

	tmpl := h.prompts.Get("title_suggestions")
	if tmpl == nil {
		response.InternalError(c)
		return
	}

	userPrompt, err := renderTemplate(tmpl.UserPromptTmpl, map[string]any{
		"Title":       product.Title,
		"Description": product.Description.String,
		"Count":       req.Count,
		"Language":    req.Language,
	})
	if err != nil {
		response.InternalError(c)
		return
	}

	genResp, err := h.client.GenerateText(c.Request.Context(), shopIDStr, "ai_text", GenerateRequest{
		SystemPrompt: tmpl.SystemPrompt,
		UserPrompt:   userPrompt,
		Temperature:  tmpl.Temperature,
		MaxTokens:    tmpl.MaxTokens,
	})
	if errors.Is(err, ErrNotConfigured) {
		c.JSON(http.StatusServiceUnavailable, comingSoon)
		return
	}
	if errors.Is(err, ErrQuotaExceeded) {
		c.JSON(http.StatusTooManyRequests, AIResponse{Status: "quota_exceeded", Message: "Monthly AI quota exhausted."})
		return
	}
	if err != nil {
		response.InternalError(c)
		return
	}

	response.OK(c, AIResponse{Status: "ok", Data: map[string]string{"titles": genResp.Content}})
}

// -------------------------------------------------------------------------- //
//  Content / email / analytics endpoints                                       //
// -------------------------------------------------------------------------- //

// AnalyzeContentQuality analyzes text for quality and SEO friendliness.
func (h *Handler) AnalyzeContentQuality(c *gin.Context) {
	shopIDStr := c.GetString("shop_id")

	var req ContentQualityReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	tmpl := h.prompts.Get("content_quality")
	if tmpl == nil {
		response.InternalError(c)
		return
	}

	userPrompt, err := renderTemplate(tmpl.UserPromptTmpl, map[string]any{
		"Text": req.Text,
	})
	if err != nil {
		response.InternalError(c)
		return
	}

	genResp, err := h.client.GenerateText(c.Request.Context(), shopIDStr, "ai_text", GenerateRequest{
		SystemPrompt: tmpl.SystemPrompt,
		UserPrompt:   userPrompt,
		Temperature:  tmpl.Temperature,
		MaxTokens:    tmpl.MaxTokens,
	})
	if errors.Is(err, ErrNotConfigured) {
		c.JSON(http.StatusServiceUnavailable, comingSoon)
		return
	}
	if errors.Is(err, ErrQuotaExceeded) {
		c.JSON(http.StatusTooManyRequests, AIResponse{Status: "quota_exceeded", Message: "Monthly AI quota exhausted."})
		return
	}
	if err != nil {
		response.InternalError(c)
		return
	}

	response.OK(c, AIResponse{Status: "ok", Data: map[string]string{"analysis": genResp.Content}})
}

// GenerateEmailCopy generates marketing email copy.
func (h *Handler) GenerateEmailCopy(c *gin.Context) {
	shopIDStr := c.GetString("shop_id")

	var req GenerateEmailCopyReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	tmpl := h.prompts.Get("email_copy")
	if tmpl == nil {
		response.InternalError(c)
		return
	}

	userPrompt, err := renderTemplate(tmpl.UserPromptTmpl, map[string]any{
		"EventType": req.EventType,
		"Context":   fmt.Sprintf("%v", req.Context),
	})
	if err != nil {
		response.InternalError(c)
		return
	}

	genResp, err := h.client.GenerateText(c.Request.Context(), shopIDStr, "ai_email", GenerateRequest{
		SystemPrompt: tmpl.SystemPrompt,
		UserPrompt:   userPrompt,
		Temperature:  tmpl.Temperature,
		MaxTokens:    tmpl.MaxTokens,
	})
	if errors.Is(err, ErrNotConfigured) {
		c.JSON(http.StatusServiceUnavailable, comingSoon)
		return
	}
	if errors.Is(err, ErrQuotaExceeded) {
		c.JSON(http.StatusTooManyRequests, AIResponse{Status: "quota_exceeded", Message: "Monthly AI quota exhausted."})
		return
	}
	if err != nil {
		response.InternalError(c)
		return
	}

	response.OK(c, AIResponse{Status: "ok", Data: map[string]string{"email_copy": genResp.Content}})
}

// GenerateWeeklyReport generates a natural-language analytics report.
func (h *Handler) GenerateWeeklyReport(c *gin.Context) {
	shopIDStr := c.GetString("shop_id")

	var req WeeklyReportReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	shopUUID, err := uuid.Parse(shopIDStr)
	if err != nil {
		response.BadRequest(c, "invalid shop id")
		return
	}
	shop, err := h.db.GetShop(c.Request.Context(), shopUUID)
	if err != nil {
		response.InternalError(c)
		return
	}

	tmpl := h.prompts.Get("weekly_report")
	if tmpl == nil {
		response.InternalError(c)
		return
	}

	userPrompt, err := renderTemplate(tmpl.UserPromptTmpl, map[string]any{
		"ShopName":    shop.Name,
		"DateRange":   req.FromDate + " to " + req.ToDate,
		"RevenueData": "not available",
		"TopProducts": "not available",
		"KeyMetrics":  "not available",
	})
	if err != nil {
		response.InternalError(c)
		return
	}

	genResp, err := h.client.GenerateText(c.Request.Context(), shopIDStr, "ai_reports", GenerateRequest{
		SystemPrompt: tmpl.SystemPrompt,
		UserPrompt:   userPrompt,
		Temperature:  tmpl.Temperature,
		MaxTokens:    tmpl.MaxTokens,
	})
	if errors.Is(err, ErrNotConfigured) {
		c.JSON(http.StatusServiceUnavailable, comingSoon)
		return
	}
	if errors.Is(err, ErrQuotaExceeded) {
		c.JSON(http.StatusTooManyRequests, AIResponse{Status: "quota_exceeded", Message: "Monthly AI quota exhausted."})
		return
	}
	if err != nil {
		response.InternalError(c)
		return
	}

	response.OK(c, AIResponse{Status: "ok", Data: map[string]string{"report": genResp.Content}})
}

// -------------------------------------------------------------------------- //
//  Engine / usage endpoints                                                    //
// -------------------------------------------------------------------------- //



// GetUsageStats returns current-month AI usage for the shop.
func (h *Handler) GetUsageStats(c *gin.Context) {
	shopIDStr := c.GetString("shop_id")
	shopUUID := mustParseUUID(shopIDStr)

	now := time.Now().UTC()
	start := pgtype.Timestamptz{
		Time:  time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC),
		Valid: true,
	}

	features := []string{"ai_text", "ai_seo", "ai_translate", "ai_email", "ai_reports"}
	usage := make(map[string]int64, len(features))
	for _, f := range features {
		count, _ := h.db.CountAIUsage(c.Request.Context(), db.CountAIUsageParams{
			ShopID:      shopUUID,
			Feature:     f,
			PeriodStart: start,
		})
		usage[f] = count
	}

	response.OK(c, usage)
}

// -------------------------------------------------------------------------- //
//  Internal helpers                                                            //
// -------------------------------------------------------------------------- //

// fetchProductAndShop fetches both the product and its owning shop in two DB calls.
func (h *Handler) fetchProductAndShop(ctx context.Context, shopIDStr, productIDStr string) (db.Product, db.Shop, error) {
	shopUUID, err := uuid.Parse(shopIDStr)
	if err != nil {
		return db.Product{}, db.Shop{}, err
	}
	productUUID, err := uuid.Parse(productIDStr)
	if err != nil {
		return db.Product{}, db.Shop{}, err
	}

	product, err := h.db.GetProduct(ctx, db.GetProductParams{
		ID:     productUUID,
		ShopID: pgtype.UUID{Bytes: shopUUID, Valid: true},
	})
	if err != nil {
		return db.Product{}, db.Shop{}, err
	}

	shop, err := h.db.GetShop(ctx, shopUUID)
	if err != nil {
		return db.Product{}, db.Shop{}, err
	}
	return product, shop, nil
}

// renderTemplate executes a Go text/template string with the given data.
func renderTemplate(tmplStr string, data map[string]any) (string, error) {
	t, err := template.New("prompt").Parse(strings.ReplaceAll(tmplStr, "{{.", "{{."))
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}
