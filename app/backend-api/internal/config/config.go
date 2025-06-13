package config

import (
	"log"
	"os"

	"github.com/spf13/viper"
)

type Config struct {
	Port        string
	DatabaseURL string
	Environment string
	JWTSecret   string
	RedisAddr   string

	// Database connection pool tuning
	DBMaxConns          int32
	DBMinConns          int32
	DBMaxConnLifetime   int // seconds
	DBMaxConnIdleTime   int // seconds
	DBHealthCheckPeriod int // seconds

	// Paymob payment gateway
	PaymobAPIKey        string
	PaymobIntegrationID int
	PaymobIframeID      int
	PaymobHMACSecret    string
	// 32+ byte key used to encrypt shop payment-method credentials at rest.
	// Leave empty in development to skip encryption (plaintext stored).
	PaymobCredentialsKey string

	// Transactional email (SMTP)
	SMTPHost     string
	SMTPPort     int
	SMTPUsername string
	SMTPPassword string
	SMTPFrom     string

	// Frontend URL used for building email links (e.g. password reset)
	FrontendURL string

	// Upload directory for product images / assets
	UploadDir string

	// Background worker tuning
	WorkerConcurrency int
	MetricsPort       string

	// VAPID keys for Web Push notifications
	VAPIDPublicKey  string
	VAPIDPrivateKey string
	VAPIDEmail      string // e.g. "mailto:admin@example.com"

	// 2FA: 32-byte hex key for AES-256-GCM encryption of TOTP secrets at rest.
	// If empty in development, secrets are stored as plaintext.
	TOTPEncryptionKey string

	// Cloudflare R2 object storage (for product images).
	// If R2AccountID is empty, the catalog handler falls back to local disk uploads.
	R2AccountID string
	R2AccessKey string
	R2SecretKey string
	R2Bucket    string
	R2PublicURL string // e.g. "https://pub-xxx.r2.dev"

	// AI Engine (self-hosted GPU VPS over WireGuard).
	// Empty = disabled; all engine calls return ErrNotConfigured.
	AIEngineBaseURL string
	AIEngineAPIKey  string

	// LLM text generation provider (Gemini / OpenAI / Groq / DeepSeek).
	// Empty API key = disabled; all text-gen calls return ErrNotConfigured.
	AIPrimaryProvider  string // "gemini" | "openai" | "groq" | "deepseek"
	AIPrimaryAPIKey    string
	AIPrimaryModel     string // e.g. "gemini-2.0-flash-exp"
	AIFallbackProvider string
	AIFallbackAPIKey   string
	AIFallbackModel    string
	AIPromptsDir       string // default: "internal/ai/prompts"
}

func Load() *Config {
	viper.SetConfigName("config")
	viper.SetConfigType("env")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./app/backend-api")
	viper.AutomaticEnv()

	// Set defaults
	viper.SetDefault("PORT", "8080")
	viper.SetDefault("ENVIRONMENT", "development")
	viper.SetDefault("JWT_SECRET", "dev-secret-change-in-production")
	viper.SetDefault("REDIS_ADDR", "redis:6379")
	viper.SetDefault("SMTP_PORT", 587)
	viper.SetDefault("DB_MAX_CONNS", 100)
	viper.SetDefault("DB_MIN_CONNS", 10)
	viper.SetDefault("DB_MAX_CONN_LIFETIME", 3600)
	viper.SetDefault("DB_MAX_CONN_IDLE_TIME", 1800)
	viper.SetDefault("DB_HEALTH_CHECK_PERIOD", 60)
	viper.SetDefault("UPLOAD_DIR", "/uploads")
	viper.SetDefault("FRONTEND_URL", "https://shop.nexuscommerce.io")
	viper.SetDefault("WORKER_CONCURRENCY", 4)
	viper.SetDefault("METRICS_PORT", "9091")

	if err := viper.ReadInConfig(); err != nil {
		log.Printf("No config file found, using environment variables and defaults: %v", err)
	}

	// Try to get DATABASE_URL from environment if not in config
	dbURL := viper.GetString("DATABASE_URL")
	if dbURL == "" {
		dbURL = os.Getenv("DATABASE_URL")
	}

	cfg := &Config{
		Port:        viper.GetString("PORT"),
		DatabaseURL: dbURL,
		Environment: viper.GetString("ENVIRONMENT"),
		JWTSecret:   viper.GetString("JWT_SECRET"),
		RedisAddr:   viper.GetString("REDIS_ADDR"),

		DBMaxConns:          int32(viper.GetInt("DB_MAX_CONNS")),
		DBMinConns:          int32(viper.GetInt("DB_MIN_CONNS")),
		DBMaxConnLifetime:   viper.GetInt("DB_MAX_CONN_LIFETIME"),
		DBMaxConnIdleTime:   viper.GetInt("DB_MAX_CONN_IDLE_TIME"),
		DBHealthCheckPeriod: viper.GetInt("DB_HEALTH_CHECK_PERIOD"),

		PaymobAPIKey:         viper.GetString("PAYMOB_API_KEY"),
		PaymobIntegrationID:  viper.GetInt("PAYMOB_INTEGRATION_ID"),
		PaymobIframeID:       viper.GetInt("PAYMOB_IFRAME_ID"),
		PaymobHMACSecret:     viper.GetString("PAYMOB_HMAC_SECRET"),
		PaymobCredentialsKey: viper.GetString("PAYMOB_CREDENTIALS_KEY"),

		SMTPHost:     viper.GetString("SMTP_HOST"),
		SMTPPort:     viper.GetInt("SMTP_PORT"),
		SMTPUsername: viper.GetString("SMTP_USERNAME"),
		SMTPPassword: viper.GetString("SMTP_PASSWORD"),
		SMTPFrom:     viper.GetString("SMTP_FROM"),

		FrontendURL: viper.GetString("FRONTEND_URL"),
		UploadDir:   viper.GetString("UPLOAD_DIR"),

		WorkerConcurrency: viper.GetInt("WORKER_CONCURRENCY"),
		MetricsPort:       viper.GetString("METRICS_PORT"),

		VAPIDPublicKey:  viper.GetString("VAPID_PUBLIC_KEY"),
		VAPIDPrivateKey: viper.GetString("VAPID_PRIVATE_KEY"),
		VAPIDEmail:      viper.GetString("VAPID_EMAIL"),

		TOTPEncryptionKey: viper.GetString("TOTP_ENCRYPTION_KEY"),

		R2AccountID: viper.GetString("R2_ACCOUNT_ID"),
		R2AccessKey: viper.GetString("R2_ACCESS_KEY"),
		R2SecretKey: viper.GetString("R2_SECRET_KEY"),
		R2Bucket:    viper.GetString("R2_BUCKET"),
		R2PublicURL: viper.GetString("R2_PUBLIC_URL"),

		AIEngineBaseURL: viper.GetString("AI_ENGINE_BASE_URL"),
		AIEngineAPIKey:  viper.GetString("AI_ENGINE_API_KEY"),

		AIPrimaryProvider:  viper.GetString("AI_PRIMARY_PROVIDER"),
		AIPrimaryAPIKey:    viper.GetString("AI_PRIMARY_API_KEY"),
		AIPrimaryModel:     viper.GetString("AI_PRIMARY_MODEL"),
		AIFallbackProvider: viper.GetString("AI_FALLBACK_PROVIDER"),
		AIFallbackAPIKey:   viper.GetString("AI_FALLBACK_API_KEY"),
		AIFallbackModel:    viper.GetString("AI_FALLBACK_MODEL"),
		AIPromptsDir:       viper.GetString("AI_PROMPTS_DIR"),
	}

	// Guard against running with the default insecure JWT secret in production.
	if cfg.Environment == "production" && cfg.JWTSecret == "dev-secret-change-in-production" {
		log.Fatal("JWT_SECRET must be explicitly set in production — refusing to start with the default insecure value")
	}

	// Guard against unencrypted payment credentials in production.
	if cfg.Environment == "production" && cfg.PaymobCredentialsKey == "" {
		log.Fatal("PAYMOB_CREDENTIALS_KEY must be set in production — shop payment credentials would be stored in plaintext")
	}

	// Guard against missing Paymob HMAC secret in production (webhook verification would be skipped).
	if cfg.Environment == "production" && cfg.PaymobHMACSecret == "" {
		log.Fatal("PAYMOB_HMAC_SECRET must be set in production — Paymob webhook signatures cannot be verified")
	}

	return cfg
}
