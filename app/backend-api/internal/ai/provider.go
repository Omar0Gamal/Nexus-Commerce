package ai

import (
	"context"
	"errors"
)

// LLMProvider is the interface satisfied by all LLM provider implementations.
// Currently no implementation is wired — the AIClient returns ErrNotConfigured
// until ai_config is populated with a valid key.
type LLMProvider interface {
	Generate(ctx context.Context, req GenerateRequest) (GenerateResponse, error)
	Name() string
}

// GenerateRequest holds the parameters for an LLM generation call.
type GenerateRequest struct {
	SystemPrompt string
	UserPrompt   string
	OutputSchema any
	Temperature  float64
	MaxTokens    int
}

// GenerateResponse holds the result of an LLM generation call.
type GenerateResponse struct {
	Content   string
	Model     string
	Provider  string
	InputTok  int
	OutputTok int
	LatencyMs int64
}

// ErrNotConfigured is returned when no LLM provider has been configured.
var ErrNotConfigured = errors.New("AI provider not configured")

// ErrQuotaExceeded is returned when the shop has exhausted their monthly AI quota.
var ErrQuotaExceeded = errors.New("AI quota exceeded for this billing period")
