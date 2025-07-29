package analytics

import (
	"context"
	"encoding/json"
	"fmt"

	"backend-api/internal/ai"
	"backend-api/internal/db"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type FunnelInsightResult struct {
	Group             string `json:"group"`
	Count             int64  `json:"count"`
	LikelyCause       string `json:"likely_cause"`
	RecommendedAction string `json:"recommended_action"`
}

// GenerateFunnelInsights aggregates drop-off data using SQL and passes the summary to the AI to generate actionable marketing insights.
func (s *Service) GenerateFunnelInsights(ctx context.Context, shopID uuid.UUID, start, end pgtype.Timestamptz, aiClient *ai.AIClient) ([]FunnelInsightResult, error) {
	// 1. Math: Group drop-offs via SQL
	dropoffs, err := s.db.GetDynamicDropoffs(ctx, db.GetDynamicDropoffsParams{
		ShopID:         shopID,
		CreatedAtStart: start,
		CreatedAtEnd:   end,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get dropoffs: %w", err)
	}

	if len(dropoffs) == 0 {
		return []FunnelInsightResult{}, nil
	}

	// 2. Format the grouped data for the LLM
	dataBytes, _ := json.Marshal(dropoffs)

	systemPrompt := `You are an expert e-commerce data analyst and marketing consultant.
I will provide you with raw funnel drop-off groups (number of abandoned sessions at specific stages).
For each group, identify the likely root cause of the drop-off and suggest ONE highly actionable, automated marketing strategy or platform change to recover them or prevent future drops.
You MUST output your response as raw JSON matching this schema:
[
  {
    "group": "abandoned_after_shipping (Zone A)",
    "count": 150,
    "likely_cause": "Shipping shock or unsupported zone",
    "recommended_action": "Trigger automated email with a 10% flat shipping discount valid for 24 hours."
  }
]
Do not use markdown formatting like ` + "`" + `` + "`" + `json. Just return the raw JSON array.`

	userPrompt := fmt.Sprintf("Here is the cart abandonment data for the requested period:\n%s", string(dataBytes))

	// 3. AI: Generate Insights
	genResp, err := aiClient.GenerateText(ctx, shopID.String(), "ai_reports", ai.GenerateRequest{
		SystemPrompt: systemPrompt,
		UserPrompt:   userPrompt,
		Temperature:  0.7,
		MaxTokens:    1500,
	})
	if err != nil {
		return nil, fmt.Errorf("ai generation failed: %w", err)
	}

	// 4. Parse the LLM response
	var insights []FunnelInsightResult
	content := genResp.Content
	
	// Quick sanitize in case the LLM wrapped it in markdown
	if len(content) > 7 && content[:7] == "```json" {
		content = content[7:]
		if len(content) > 3 && content[len(content)-3:] == "```" {
			content = content[:len(content)-3]
		}
	}

	if err := json.Unmarshal([]byte(content), &insights); err != nil {
		return nil, fmt.Errorf("failed to parse AI response: %w. Raw: %s", err, genResp.Content)
	}

	return insights, nil
}
