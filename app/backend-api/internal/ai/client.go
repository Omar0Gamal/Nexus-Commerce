package ai

import (
	"context"
	"time"

	"backend-api/internal/db"

	"go.uber.org/zap"
)

type AIClient struct {
	primary  LLMProvider
	fallback LLMProvider
	db       *db.Queries
	logger   *zap.Logger
}

func NewAIClient(queries *db.Queries, logger *zap.Logger) *AIClient {
	return &AIClient{db: queries, logger: logger}
}

func (c *AIClient) Configure(primary, fallback LLMProvider) {
	c.primary = primary
	c.fallback = fallback
}

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
