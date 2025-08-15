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

// GeminiProvider implements LLMProvider using the Google Generative Language API.
// Model: gemini-2.0-flash-exp (or any model name passed at construction).
// Returns ErrNotConfigured when apiKey is empty.
type GeminiProvider struct {
	apiKey string
	model  string
	http   *http.Client
}

// in the unconfigured state (all calls return ErrNotConfigured).
func NewGeminiProvider(apiKey, model string) *GeminiProvider {
	if model == "" {
		model = "gemini-2.0-flash-exp"
	}
	return &GeminiProvider{
		apiKey: apiKey,
		model:  model,
		http:   &http.Client{Timeout: 30 * time.Second},
	}
}

func (p *GeminiProvider) Name() string { return "gemini" }

func (p *GeminiProvider) Generate(ctx context.Context, req GenerateRequest) (GenerateResponse, error) {
	if p.apiKey == "" {
		return GenerateResponse{}, ErrNotConfigured
	}

	combined := req.SystemPrompt + "\n\n" + req.UserPrompt

	body := map[string]any{
		"contents": []map[string]any{
			{"parts": []map[string]any{{"text": combined}}},
		},
		"generationConfig": map[string]any{
			"temperature":     req.Temperature,
			"maxOutputTokens": req.MaxTokens,
		},
	}

	start := time.Now()
	raw, err := p.post(ctx, body)
	if err != nil {
		return GenerateResponse{}, err
	}
	latency := time.Since(start).Milliseconds()

	var resp geminiResponse
	if err := json.Unmarshal(raw, &resp); err != nil {
		return GenerateResponse{}, fmt.Errorf("gemini: decode response: %w", err)
	}
	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return GenerateResponse{}, fmt.Errorf("gemini: empty response")
	}

	return GenerateResponse{
		Content:   resp.Candidates[0].Content.Parts[0].Text,
		Model:     p.model,
		Provider:  "gemini",
		InputTok:  resp.UsageMetadata.PromptTokenCount,
		OutputTok: resp.UsageMetadata.CandidatesTokenCount,
		LatencyMs: latency,
	}, nil
}

func (p *GeminiProvider) post(ctx context.Context, body any) ([]byte, error) {
	url := fmt.Sprintf(
		"https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s",
		p.model, p.apiKey,
	)
	payload, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("gemini: http error: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("gemini: read body: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("gemini: status %d: %s", resp.StatusCode, string(data))
	}
	return data, nil
}

// geminiResponse is the minimal shape of a Gemini API response.
type geminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
	UsageMetadata struct {
		PromptTokenCount     int `json:"promptTokenCount"`
		CandidatesTokenCount int `json:"candidatesTokenCount"`
	} `json:"usageMetadata"`
}
