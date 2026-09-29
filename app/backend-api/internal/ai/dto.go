package ai

type GenerateDescriptionReq struct {
	Tone     string `json:"tone"`      // "professional" | "casual" | "luxury"
	Language string `json:"language"`  // "en" | "ar"
	MaxWords int    `json:"max_words"` // default 200
}

type GenerateSEOMetaReq struct {
	Language string `json:"language"` // "en" | "ar"
}

type TranslateReq struct {
	TargetLanguage string `json:"target_language" binding:"required"` // "ar" | "en" | "fr"
}

type SuggestTitlesReq struct {
	Count    int    `json:"count"`    // 3–10, default 5
	Language string `json:"language"` // "en" | "ar"
}

type ContentQualityReq struct {
	Text string `json:"text" binding:"required"` // the content to analyse
}

type GenerateEmailCopyReq struct {
	EventType string         `json:"event_type" binding:"required"` // "welcome" | "order_confirm" | "abandoned_cart"
	Context   map[string]any `json:"context"`
}

type WeeklyReportReq struct {
	FromDate string `json:"from_date" binding:"required"` // YYYY-MM-DD
	ToDate   string `json:"to_date"   binding:"required"` // YYYY-MM-DD
}

type AIResponse struct {
	Status  string `json:"status"` // "ok" | "coming_soon" | "quota_exceeded"
	Data    any    `json:"data,omitempty"`
	Message string `json:"message,omitempty"`
}
