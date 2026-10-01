package main

import (
	"net/http"
	"time"

	"backend-api/internal/config"
	"backend-api/internal/shared/middleware"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

func newRouter(cfg *config.Config, svc *services, pool *pgxpool.Pool, rdb *redis.Client) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.SmartCORS())
	r.Use(middleware.StructuredLogger())
	r.Use(middleware.PrometheusInstrumentation())

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	r.GET("/ready", func(c *gin.Context) {
		if err := pool.Ping(c.Request.Context()); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "not ready", "db": false})
			return
		}
		if err := rdb.Ping(c.Request.Context()).Err(); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "not ready", "redis": false})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ready"})
	})

	// Registered before TenantMiddleware because search engine crawlers hit root sitemap/robots without X-Shop-ID.
	svc.SEO.RegisterPublicRoutes(r)

	requireAuth := middleware.RequireAuth(svc.Token)
	requireStaff := middleware.RequireStaff()
	requireCustomer := middleware.RequireCustomer()
	requireAdmin := middleware.RequirePlatformAdmin()
	authRateLimiter := middleware.NewIPRateLimiter(rdb, 20, time.Minute)

	staffAPI := r.Group("/api/v1")
	svc.Auth.RegisterStaffRoutes(staffAPI, requireAuth, requireStaff, authRateLimiter)
	svc.Auth.RegisterAppRoutes(staffAPI, requireAuth)

	adminAPI := r.Group("/api/v1")
	svc.Auth.RegisterAdminRoutes(adminAPI, requireAuth, requireAdmin, authRateLimiter)
	svc.Admin.RegisterRoutes(adminAPI, requireAuth, requireAdmin)
	svc.Platform.RegisterRoutes(adminAPI, requireAuth, requireAdmin)

	customerAPI := r.Group("/api/v1")
	customerAPI.Use(middleware.TenantMiddleware(cfg.GatewaySecret))
	svc.Auth.RegisterCustomerRoutes(customerAPI, requireAuth, requireCustomer, authRateLimiter)
	svc.Orders.RegisterCustomerRoutes(customerAPI, requireAuth, requireCustomer)
	svc.Addresses.RegisterCustomerRoutes(customerAPI, requireAuth, requireCustomer)

	api := r.Group("/api/v1")
	api.Use(middleware.TenantMiddleware(cfg.GatewaySecret))

	svc.Catalog.RegisterRoutes(api, requireAuth, svc.Auth.RequirePermission("products.read"), svc.MeteringSvc.RequireQuota("products"))
	api.POST("/catalog/import/shopify", requireAuth, svc.Auth.RequirePermission("products.write"), svc.Catalog.ImportShopify)
	svc.Orders.RegisterRoutes(api, requireAuth, svc.Auth.RequirePermission("orders.read"))
	svc.Coupons.RegisterRoutes(api, requireAuth, svc.Auth.RequirePermission("marketing.read"))
	svc.Returns.RegisterRoutes(api, requireAuth, svc.Auth.RequirePermission("returns.read"))
	svc.Shipping.RegisterRoutes(api, requireAuth, svc.Auth.RequirePermission("settings.read"))
	svc.Billing.RegisterRoutes(api, requireAuth, svc.Auth.RequirePermission("billing.read"))
	svc.Customers.RegisterRoutes(api, requireAuth, svc.Auth.RequirePermission("customers.read"))
	svc.Customers.RegisterCustomerSelfServiceRoutes(api, requireAuth, requireCustomer)
	svc.Webhook.RegisterRoutes(api, requireAuth, svc.Auth.RequirePermission("settings.read"))
	svc.Auth.RegisterShopRoutes(api, requireAuth, requireStaff)

	// OptionalAuth binds customer identity when present while allowing guest carts via X-Cart-Token.
	optionalAuth := middleware.OptionalAuth(svc.Token)
	svc.Cart.RegisterRoutes(api, optionalAuth)
	svc.Orders.RegisterCheckoutRoutes(api, optionalAuth)
	svc.Orders.RegisterStorefrontRoutes(api)

	svc.Wishlists.RegisterRoutes(api, requireAuth, requireCustomer)
	svc.Reviews.RegisterCustomerRoutes(api, requireAuth, requireCustomer, authRateLimiter)
	svc.Reviews.RegisterStaffRoutes(api, requireAuth, svc.Auth.RequirePermission("products.read"))
	svc.SocialProof.RegisterRoutes(api)

	svc.Notif.RegisterRoutes(api, requireAuth)
	svc.Notif.RegisterCustomerPrefsRoutes(customerAPI, requireAuth)

	svc.Currency.RegisterPublicRoutes(api)
	svc.Currency.RegisterStaffRoutes(api, requireAuth)

	svc.Inventory.RegisterRoutes(api, requireAuth, svc.Auth.RequirePermission("inventory.read"))

	svc.Storage.RegisterRoutes(api, requireAuth, svc.Auth.RequirePermission("products.write"))
	svc.Promotions.RegisterStaffRoutes(api, requireAuth, svc.Auth.RequirePermission("marketing.read"))
	svc.Promotions.RegisterCustomerRoutes(customerAPI, requireAuth, requireCustomer)
	svc.Paymob.RegisterRoutes(api, requireAuth, svc.Auth.RequirePermission("settings.read"))

	svc.Staff.RegisterRoutes(api, requireAuth, requireStaff)
	svc.Auth.RegisterRoleRoutes(api, requireAuth, requireStaff)

	analyticsGroup := api.Group("/analytics")
	analyticsGroup.Use(requireAuth, svc.Auth.RequirePermission("analytics.read"), authRateLimiter, middleware.Timeout(30*time.Second))
	svc.Analytics.RegisterRoutes(analyticsGroup, svc.BillingSvc.RequireFeature)

	svc.Support.RegisterStaffRoutes(api, requireAuth, svc.Auth.RequirePermission("support.read"))
	svc.Support.RegisterCustomerRoutes(customerAPI, requireAuth, requireCustomer)
	svc.Support.RegisterTenantRoutes(api, requireAuth, svc.Auth.RequirePermission("support.read"))

	svc.Subscriptions.RegisterStaffRoutes(api, requireAuth, svc.Auth.RequirePermission("billing.read"))
	svc.Subscriptions.RegisterCustomerRoutes(customerAPI, requireAuth, requireCustomer)

	aiGroup := api.Group("/ai")
	aiGroup.Use(authRateLimiter, middleware.Timeout(60*time.Second))
	svc.AI.RegisterRoutes(aiGroup, requireAuth, svc.Auth.RequirePermission("products.write"), svc.BillingSvc.RequireFeature)

	svc.SEO.RegisterRoutes(api, requireAuth, svc.Auth.RequirePermission("seo.read"), svc.BillingSvc.RequireFeature)

	publicAPI := r.Group("/api/v1")
	svc.Paymob.RegisterPublicRoutes(publicAPI)
	svc.Billing.RegisterPublicRoutes(publicAPI)
	publicAPI.GET("/storefront/shop/:domain", svc.Auth.GetShopContextByDomain)
	publicAPI.POST("/analytics/ingest", svc.Analytics.Ingest)

	return r
}
