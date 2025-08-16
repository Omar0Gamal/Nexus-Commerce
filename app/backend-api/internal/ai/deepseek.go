package ai

import (
	"context"
	"net/http"
	"time"
)

// DeepSeekProvider implements LLMProvider using the DeepSeek Chat API (OpenAI-compatible).
// Endpoint: https://api.deepseek.com/chat/completions
// Returns ErrNotConfigured when apiKey is empty.
type DeepSeekProvider struct {
	inner *OpenAIProvider
}

func NewDeepSeekProvider(apiKey, model string) *DeepSeekProvider {
	if model == "" {
		model = "deepseek-chat"
	}
	return &DeepSeekProvider{
		inner: &OpenAIProvider{
			apiKey:  apiKey,
			model:   model,
			baseURL: "https://api.deepseek.com/chat/completions",
			http:    &http.Client{Timeout: 30 * time.Second},
		},
	}
}

func (p *DeepSeekProvider) Name() string { return "deepseek" }

func (p *DeepSeekProvider) Generate(ctx context.Context, req GenerateRequest) (GenerateResponse, error) {
	if p.inner.apiKey == "" {
		return GenerateResponse{}, ErrNotConfigured
	}
	return p.inner.generate(ctx, req, "deepseek")
}
