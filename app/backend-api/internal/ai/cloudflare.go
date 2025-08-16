package ai

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// CloudflareProvider implements LLMProvider using Cloudflare Workers AI.
// Requires the apiKey to be in the format "account_id:api_token".
type CloudflareProvider struct {
	inner *OpenAIProvider
}

func NewCloudflareProvider(apiKey, model string) *CloudflareProvider {
	if model == "" {
		model = "@cf/meta/llama-3-8b-instruct"
	}

	// Parse "account_id:api_token"
	parts := strings.SplitN(apiKey, ":", 2)
	accountID := ""
	token := apiKey
	if len(parts) == 2 {
		accountID = parts[0]
		token = parts[1]
	}

	baseURL := fmt.Sprintf("https://api.cloudflare.com/client/v4/accounts/%s/ai/v1/chat/completions", accountID)

	return &CloudflareProvider{
		inner: &OpenAIProvider{
			apiKey:  token,
			model:   model,
			baseURL: baseURL,
			http:    &http.Client{Timeout: 30 * time.Second},
		},
	}
}

func (p *CloudflareProvider) Name() string { return "cloudflare" }

func (p *CloudflareProvider) Generate(ctx context.Context, req GenerateRequest) (GenerateResponse, error) {
	if p.inner.apiKey == "" {
		return GenerateResponse{}, ErrNotConfigured
	}
	return p.inner.generate(ctx, req, "cloudflare")
}
