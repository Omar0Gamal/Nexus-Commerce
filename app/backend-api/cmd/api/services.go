package main

import (
	"backend-api/internal/addresses"
	"backend-api/internal/admin"
	"backend-api/internal/ai"
	"backend-api/internal/analytics"
	"backend-api/internal/auth"
	"backend-api/internal/billing"
	"backend-api/internal/cart"
	"backend-api/internal/catalog"
	"backend-api/internal/coupons"
	"backend-api/internal/currency"
	"backend-api/internal/customers"
	"backend-api/internal/inventory"
	"backend-api/internal/notifications"
	"backend-api/internal/orders"
	"backend-api/internal/paymob"
	"backend-api/internal/promotions"
	"backend-api/internal/returns"
	"backend-api/internal/reviews"
	"backend-api/internal/seo"
	"backend-api/internal/shared/token"
	"backend-api/internal/shipping"
	"backend-api/internal/socialproof"
	"backend-api/internal/storage"
	"backend-api/internal/subscriptions"
	"backend-api/internal/support"
	"backend-api/internal/webhooks"
	"backend-api/internal/wishlists"

	"go.uber.org/zap"
)

// services holds all initialised service and handler instances.
type services struct {
	// Auth
	Token     *token.Service
	Auth      *auth.Handler
	StaffAuth *auth.Service
	AdminSvc  *admin.Service

	// Handlers
	Billing       *billing.Handler
	Webhook       *webhooks.Handler
	Catalog       *catalog.Handler
	Orders        *orders.Handler
	Admin         *admin.Handler
	Platform      *admin.PlatformHandler
	Cart          *cart.Handler
	Customers     *customers.Handler
	Addresses     *addresses.Handler
	Coupons       *coupons.Handler
	Returns       *returns.Handler
	Shipping      *shipping.Handler
	Wishlists     *wishlists.Handler
	Reviews       *reviews.Handler
	SocialProof   *socialproof.Handler
	Notif         *notifications.Handler
	Currency      *currency.Handler
	Inventory     *inventory.Handler
	Promotions    *promotions.Handler
	Paymob        *paymob.Handler
	Analytics     *analytics.Handler
	Support       *support.Handler
	Subscriptions *subscriptions.Handler

	// AI (text generation + self-hosted engine)
	AI *ai.Handler

	// SEO
	SEO    *seo.Handler
	SEOSvc *seo.Service

	// Services exposed for workers / route helpers
	CurrencySvc  *currency.Service
	BillingSvc   *billing.Service
	MeteringSvc  *billing.MeteringService
	AnalyticsSvc *analytics.Service
}

// initServices constructs every service and handler, wiring all dependencies
// via constructors. No setter injection is used.
func initServices(d *infra) *services {
	tokenSvc := token.NewService(d.Cfg.JWTSecret)

	// Core services (no cross-service deps)
	billingSvc := billing.NewService(d.Queries)
	shippingSvc := shipping.NewService(d.Queries)
	couponsSvc := coupons.NewService(d.Queries, d.Redis)
	webhookSvc := webhooks.NewService(d.Queries, d.JobQueue, []byte(d.Cfg.PaymobCredentialsKey))
	addressesSvc := addresses.NewService(d.Queries, d.Pool)
	wishlistsSvc := wishlists.NewService(d.Queries)
	reviewsSvc := reviews.NewService(d.Queries)
	socialProofSvc := socialproof.NewService(d.Queries, d.Redis)
	notifSvc := notifications.NewService(d.Queries, d.Redis, d.Cfg.VAPIDPublicKey, d.Cfg.VAPIDPrivateKey, d.Cfg.VAPIDEmail, d.Logger)
	currencySvc := currency.NewService(d.Queries, d.Redis)
	inventorySvc := inventory.NewService(d.Queries, d.Mailer, d.Logger)
	promotionsSvc := promotions.NewService(d.Queries, d.Redis)
	returnsSvc := returns.NewService(d.Queries, d.PaymobClient)
	adminSvc := admin.NewService(d.Queries, d.Mailer, d.Pool, d.Redis, d.Logger)

	supportSvc := support.NewService(d.Queries, d.Logger)
	subscriptionsSvc := subscriptions.NewService(d.Queries)
	seoSvc := seo.NewService(d.Queries, d.Redis, d.Logger)

	// AI provider wiring — activates automatically when env vars are set

	aiClient := ai.NewAIClient(d.Queries, d.Logger)
	if d.Cfg.AIPrimaryAPIKey != "" {
		primary := buildAIProvider(d.Cfg.AIPrimaryProvider, d.Cfg.AIPrimaryAPIKey, d.Cfg.AIPrimaryModel)
		var fallback ai.LLMProvider
		if d.Cfg.AIFallbackAPIKey != "" {
			fallback = buildAIProvider(d.Cfg.AIFallbackProvider, d.Cfg.AIFallbackAPIKey, d.Cfg.AIFallbackModel)
		}
		aiClient.Configure(primary, fallback)
	}
	promptsDir := d.Cfg.AIPromptsDir
	if promptsDir == "" {
		promptsDir = "internal/ai/prompts"
	}
	promptRegistry, _ := ai.LoadPromptRegistry(promptsDir)
	aiHandler := ai.NewHandler(aiClient, promptRegistry, d.Queries)

	// Services with cross-service deps
	catalogSvc := catalog.NewService(d.Queries, d.Pool, billingSvc, d.Redis)
	ordersSvc := orders.NewService(d.Queries, d.Pool, orders.ServiceDeps{
		Coupons:  couponsSvc,
		Shipping: shippingSvc,
		Mailer:   d.Mailer,
		Logger:   d.Logger,
		Webhooks: webhookSvc,
		Jobs:     d.JobQueue,
		Redis:    d.Redis,
	})
	cartStore := cart.NewStore(d.Redis)
	cartSvc := cart.NewService(cartStore, d.Queries, d.Pool, d.Redis)
	customersSvc := customers.NewService(d.Queries, ordersSvc)

	// Auth services
	staffAuthSvc := auth.NewService(d.Queries, d.Pool, tokenSvc, billingSvc, d.Mailer, d.Redis, d.Logger)
	customerAuthSvc := auth.NewCustomerService(d.Queries, tokenSvc, d.Redis)
	adminAuthSvc := auth.NewAdminService(d.Queries, tokenSvc)

	// Analytics
	analyticsSvc := analytics.NewService(d.Redis, d.Queries, d.Pool)

	// Metering
	meteringSvc := billing.NewMeteringService(billingSvc, d.Redis)

	// Storage (Cloudflare R2)
	var storageClient *storage.Client
	if d.Cfg.R2AccountID != "" {
		var err error
		storageClient, err = storage.New(
			d.Cfg.R2AccountID,
			d.Cfg.R2AccessKey,
			d.Cfg.R2SecretKey,
			d.Cfg.R2Bucket,
			d.Cfg.R2PublicURL,
		)
		if err != nil {
			d.Logger.Error("failed to init R2 storage client", zap.Error(err))
		}
	}

	// ── Handlers ──
	return &services{
		Token:     tokenSvc,
		StaffAuth: staffAuthSvc,
		Auth:      auth.NewHandler(staffAuthSvc, customerAuthSvc, adminAuthSvc, d.Cfg.FrontendURL, d.Cfg.TOTPEncryptionKey),
		AdminSvc:  adminSvc,

		Billing:       billing.NewHandler(billingSvc),
		Webhook:       webhooks.NewHandler(webhookSvc),
		Catalog:       catalog.NewHandler(catalogSvc, storageClient, d.Redis),
		Orders:        orders.NewHandler(ordersSvc),
		Admin:         admin.NewHandler(adminSvc),
		Platform:      admin.NewPlatformHandler(adminSvc, tokenSvc),
		Cart:          cart.NewHandler(cartSvc, ordersSvc, d.Logger),
		Customers:     customers.NewHandler(customersSvc),
		Addresses:     addresses.NewHandler(addressesSvc),
		Coupons:       coupons.NewHandler(couponsSvc),
		Returns:       returns.NewHandler(returnsSvc),
		Shipping:      shipping.NewHandler(shippingSvc),
		Wishlists:     wishlists.NewHandler(wishlistsSvc),
		Reviews:       reviews.NewHandler(reviewsSvc),
		SocialProof:   socialproof.NewHandler(socialProofSvc),
		Notif:         notifications.NewHandler(notifSvc),
		Currency:      currency.NewHandler(currencySvc),
		Inventory:     inventory.NewHandler(inventorySvc),
		Promotions:    promotions.NewHandler(promotionsSvc),
		Paymob:        paymob.NewHandler(d.PaymobClient, d.Queries, d.Pool, []byte(d.Cfg.PaymobCredentialsKey), d.Redis, d.Logger),
		Analytics:     analytics.NewHandler(analyticsSvc, aiClient),
		Support:       support.NewHandler(supportSvc),
		Subscriptions: subscriptions.NewHandler(subscriptionsSvc),
		AI:            aiHandler,

		SEO:    seo.NewHandler(seoSvc),
		SEOSvc: seoSvc,

		CurrencySvc:  currencySvc,
		BillingSvc:   billingSvc,
		MeteringSvc:  meteringSvc,
		AnalyticsSvc: analyticsSvc,
	}
}

// buildAIProvider constructs the appropriate LLMProvider by name.
func buildAIProvider(provider, apiKey, model string) ai.LLMProvider {
	switch provider {
	case "gemini":
		return ai.NewGeminiProvider(apiKey, model)
	case "cloudflare":
		return ai.NewCloudflareProvider(apiKey, model)
	case "groq":
		return ai.NewGroqProvider(apiKey, model)
	case "deepseek":
		return ai.NewDeepSeekProvider(apiKey, model)
	default:
		return ai.NewOpenAIProvider(apiKey, model)
	}
}
