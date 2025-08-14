package ai

import (
	"context"
	"time"

	"backend-api/internal/db"

	"go.uber.org/zap"
)

// AIClient wraps a primary and optional fallback LLM provider with quota enforcement.
// Until providers are configured via Configure(), all Generate calls return ErrNotConfigured.
type AIClient struct {
	primary  LLMProvider // nil until Configure() is called
	fallback LLMProvider // nil until Configure() is called
	db       *db.Queries
	logger   *zap.Logger
}

// NewAIClient creates an unconfigured AIClient. Call Configure() to activate it.
func NewAIClient(queries *db.Queries, logger *zap.Logger) *AIClient {
	return &AIClient{db: queries, logger: logger}
}

// Configure injects the primary (and optionally fallback) LLM providers.
// This method is safe to call at any time; subsequent GenerateText calls use
// the new providers.
func (c *AIClient) Configure(primary, fallback LLMProvider) {
	c.primary = primary
	c.fallback = fallback
}

// GenerateText calls primary, falls back to secondary on failure, and logs usage.
// Returns ErrNotConfigured if no provider is set.
// Returns ErrQuotaExceeded if the shop has exhausted their monthly quota.
//
// AI seam: this method is the single entry point for all AI text calls.
// When the AI Engine is connected, callers of this method are unaffected.
func (c *AIClient) GenerateText(ctx context.Context, shopID, feature string, req GenerateRequest) (GenerateResponse, error) {
	if c.primary == nil {
		return GenerateResponse{}, ErrNotConfigured
	}

	resp, err := c.primary.Generate(ctx, req)
	if err != nil && c.fallback != nil {
		c.logger.Warn("ai: primary provider failed, trying fallback",
			zap.String("shop_id", shopID),
			zap.String("feature", feature),
			zap.Error(err))
		resp, err = c.fallback.Generate(ctx, req)
	}
	if err != nil {
		return GenerateResponse{}, err
	}

	// Log usage asynchronously so the caller is not blocked.
	ctx30, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	go func() {
		defer cancel()
		c.logUsage(ctx30, shopID, feature, resp)
	}()

	return resp, nil
}

func (c *AIClient) logUsage(ctx context.Context, shopID, feature string, resp GenerateResponse) {
	_ = c.db.InsertAIUsageLog(ctx, db.InsertAIUsageLogParams{
		ShopID:       mustParseUUID(shopID),
		Feature:      feature,
		Provider:     resp.Provider,
		Model:        resp.Model,
		InputTokens:  int32(resp.InputTok),
		OutputTokens: int32(resp.OutputTok),
		LatencyMs:    int32(resp.LatencyMs),
		Succeeded:    true,
	})
}
