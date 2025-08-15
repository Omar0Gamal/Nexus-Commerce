package ai

import (
	"context"
	"net/http"
	"time"
)

// GroqProvider implements LLMProvider using the Groq API (OpenAI-compatible).
// Endpoint: https://api.groq.com/openai/v1/chat/completions
// Returns ErrNotConfigured when apiKey is empty.
type GroqProvider struct {
	inner *OpenAIProvider
}

func NewGroqProvider(apiKey, model string) *GroqProvider {
	if model == "" {
		model = "llama-3.3-70b-versatile"
	}
	return &GroqProvider{
		inner: &OpenAIProvider{
			apiKey:  apiKey,
			model:   model,
			baseURL: "https://api.groq.com/openai/v1/chat/completions",
			http:    &http.Client{Timeout: 30 * time.Second},
		},
	}
}

func (p *GroqProvider) Name() string { return "groq" }

func (p *GroqProvider) Generate(ctx context.Context, req GenerateRequest) (GenerateResponse, error) {
	if p.inner.apiKey == "" {
		return GenerateResponse{}, ErrNotConfigured
	}
	return p.inner.generate(ctx, req, "groq")
}
