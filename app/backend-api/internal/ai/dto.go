package ai

// GenerateDescriptionReq is the request body for the product description endpoint.
type GenerateDescriptionReq struct {
	Tone     string `json:"tone"`      // "professional" | "casual" | "luxury"
	Language string `json:"language"`  // "en" | "ar"
	MaxWords int    `json:"max_words"` // default 200
}

// GenerateSEOMetaReq is the request body for the SEO metadata endpoint.
type GenerateSEOMetaReq struct {
	Language string `json:"language"` // "en" | "ar"
}

// TranslateReq is the request body for the product translation endpoint.
type TranslateReq struct {
	TargetLanguage string `json:"target_language" binding:"required"` // "ar" | "en" | "fr"
}

// SuggestTitlesReq is the request body for the title suggestions endpoint.
type SuggestTitlesReq struct {
	Count    int    `json:"count"`    // 3–10, default 5
	Language string `json:"language"` // "en" | "ar"
}

// ContentQualityReq is the request body for the content quality analysis endpoint.
type ContentQualityReq struct {
	Text string `json:"text" binding:"required"` // the content to analyse
}

// GenerateEmailCopyReq is the request body for the email copy endpoint.
type GenerateEmailCopyReq struct {
	EventType string         `json:"event_type" binding:"required"` // "welcome" | "order_confirm" | "abandoned_cart"
	Context   map[string]any `json:"context"`
}

// WeeklyReportReq is the request body for the weekly analytics report endpoint.
type WeeklyReportReq struct {
	FromDate string `json:"from_date" binding:"required"` // YYYY-MM-DD
	ToDate   string `json:"to_date"   binding:"required"` // YYYY-MM-DD
}

// AIResponse is the standard envelope returned by all AI endpoints.
type AIResponse struct {
	Status  string `json:"status"` // "ok" | "coming_soon" | "quota_exceeded"
	Data    any    `json:"data,omitempty"`
	Message string `json:"message,omitempty"`
}
