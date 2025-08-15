package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// OpenAIProvider implements LLMProvider using the OpenAI Chat Completions API.
// Also usable as a base for any OpenAI-compatible endpoint (see GroqProvider,
// DeepSeekProvider which embed this type with a different base URL).
// Returns ErrNotConfigured when apiKey is empty.
type OpenAIProvider struct {
	apiKey  string
	model   string
	baseURL string
	http    *http.Client
}

// NewOpenAIProvider creates an OpenAIProvider targeting api.openai.com.
func NewOpenAIProvider(apiKey, model string) *OpenAIProvider {
	if model == "" {
		model = "gpt-4o-mini"
	}
	return &OpenAIProvider{
		apiKey:  apiKey,
		model:   model,
		baseURL: "https://api.openai.com/v1/chat/completions",
		http:    &http.Client{Timeout: 30 * time.Second},
	}
}

func (p *OpenAIProvider) Name() string { return "openai" }

func (p *OpenAIProvider) Generate(ctx context.Context, req GenerateRequest) (GenerateResponse, error) {
	if p.apiKey == "" {
		return GenerateResponse{}, ErrNotConfigured
	}
	return p.generate(ctx, req, "openai")
}

func (p *OpenAIProvider) generate(ctx context.Context, req GenerateRequest, providerName string) (GenerateResponse, error) {
	body := map[string]any{
		"model": p.model,
		"messages": []map[string]any{
			{"role": "system", "content": req.SystemPrompt},
			{"role": "user", "content": req.UserPrompt},
		},
		"max_tokens":  req.MaxTokens,
		"temperature": req.Temperature,
	}

	start := time.Now()
	raw, err := p.post(ctx, body)
	if err != nil {
		return GenerateResponse{}, err
	}
	latency := time.Since(start).Milliseconds()

	var resp openaiResponse
	if err := json.Unmarshal(raw, &resp); err != nil {
		return GenerateResponse{}, fmt.Errorf("%s: decode response: %w", providerName, err)
	}
	if len(resp.Choices) == 0 {
		return GenerateResponse{}, fmt.Errorf("%s: empty choices in response", providerName)
	}

	return GenerateResponse{
		Content:   resp.Choices[0].Message.Content,
		Model:     resp.Model,
		Provider:  providerName,
		InputTok:  resp.Usage.PromptTokens,
		OutputTok: resp.Usage.CompletionTokens,
		LatencyMs: latency,
	}, nil
}

func (p *OpenAIProvider) post(ctx context.Context, body any) ([]byte, error) {
	payload, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.baseURL, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.apiKey)

	resp, err := p.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http error: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status %d: %s", resp.StatusCode, string(data))
	}
	return data, nil
}

// openaiResponse is the minimal shape of an OpenAI-compatible chat completion response.
type openaiResponse struct {
	Model   string `json:"model"`
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
	} `json:"usage"`
}
