package main

import (
	"net/http"
	"time"

	"backend-api/internal/shared/middleware"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

func newRouter(svc *services, pool *pgxpool.Pool, rdb *redis.Client) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.SmartCORS())
	r.Use(middleware.StructuredLogger())
	r.Use(middleware.PrometheusInstrumentation())

	// Health check (no auth required)
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Readiness check — verifies DB and Redis are reachable.
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

	// SEO public routes (sitemap.xml, robots.txt) — registered before tenant middleware
	svc.SEO.RegisterPublicRoutes(r)

	// ── Shared middleware instances ──
	requireAuth := middleware.RequireAuth(svc.Token)
	requireStaff := middleware.RequireStaff()
	requireCustomer := middleware.RequireCustomer()
	requireAdmin := middleware.RequirePlatformAdmin()
	authRL := middleware.NewIPRateLimiter(rdb, 20, time.Minute)

	// ── Staff auth routes (no tenant context needed) ──
	staffAPI := r.Group("/api/v1")
	svc.Auth.RegisterStaffRoutes(staffAPI, requireAuth, requireStaff, authRL)

	// ── Admin auth routes (no tenant context needed) ──
	adminAPI := r.Group("/api/v1")
	svc.Auth.RegisterAdminRoutes(adminAPI, requireAuth, requireAdmin, authRL)

	// ── Admin management routes (platform admin only, no tenant context) ──
	svc.Admin.RegisterRoutes(adminAPI, requireAuth, requireAdmin)
	svc.Platform.RegisterRoutes(adminAPI, requireAuth, requireAdmin)

	// ── Customer auth routes (tenant-scoped via X-Shop-ID from gateway) ──
	customerAPI := r.Group("/api/v1")
	customerAPI.Use(middleware.TenantMiddleware())
	svc.Auth.RegisterCustomerRoutes(customerAPI, requireAuth, requireCustomer, authRL)
	svc.Orders.RegisterCustomerRoutes(customerAPI, requireAuth, requireCustomer)
	svc.Addresses.RegisterCustomerRoutes(customerAPI, requireAuth, requireCustomer)

	// ── Tenant-scoped API routes ──
	api := r.Group("/api/v1")
	api.Use(middleware.TenantMiddleware())

	svc.Catalog.RegisterRoutes(api, requireAuth, requireStaff, svc.MeteringSvc.RequireQuota("products"))
	svc.Orders.RegisterRoutes(api, requireAuth, requireStaff)
	svc.Coupons.RegisterRoutes(api, requireAuth, requireStaff)
	svc.Returns.RegisterRoutes(api, requireAuth, requireStaff)
	svc.Shipping.RegisterRoutes(api, requireAuth, requireStaff)
	svc.Billing.RegisterRoutes(api, requireAuth, requireStaff)
	svc.Customers.RegisterRoutes(api, requireAuth, requireStaff)
	svc.Customers.RegisterCustomerSelfServiceRoutes(api, requireAuth, requireCustomer)
	svc.Webhook.RegisterRoutes(api, requireAuth, requireStaff)
	svc.Auth.RegisterShopRoutes(api, requireAuth, requireStaff)

	// Cart routes use OptionalAuth so authenticated customers get a persistent cart.
	// Unauthenticated requests fall back to the X-Cart-Token guest flow.
	optionalAuth := middleware.OptionalAuth(svc.Token)
	svc.Cart.RegisterRoutes(api, optionalAuth)

	svc.Wishlists.RegisterRoutes(api, requireAuth, requireCustomer)
	svc.Reviews.RegisterCustomerRoutes(api, requireAuth, requireCustomer, authRL)
	svc.Reviews.RegisterStaffRoutes(api, requireAuth, requireStaff)
	svc.SocialProof.RegisterRoutes(api)

	svc.Notif.RegisterRoutes(api, requireAuth)
	svc.Notif.RegisterCustomerPrefsRoutes(customerAPI, requireAuth)

	svc.Currency.RegisterPublicRoutes(api)
	svc.Currency.RegisterStaffRoutes(api, requireAuth)

	svc.Inventory.RegisterRoutes(api, requireAuth, requireStaff)

	svc.Promotions.RegisterStaffRoutes(api, requireAuth, requireStaff)
	svc.Promotions.RegisterCustomerRoutes(customerAPI, requireAuth, requireCustomer)
	svc.Paymob.RegisterRoutes(api, requireAuth, requireStaff)

	// Role & RBAC management (staff-only, tenant-scoped)
	svc.Auth.RegisterRoleRoutes(api, requireAuth, requireStaff)

	// Analytics routes are gated per feature flag inside RegisterRoutes.
	analyticsGroup := api.Group("/analytics")
	svc.Analytics.RegisterRoutes(analyticsGroup, svc.BillingSvc.RequireFeature)

	// Support module (tickets + KB articles)
	svc.Support.RegisterStaffRoutes(api, requireAuth, requireStaff)
	svc.Support.RegisterCustomerRoutes(customerAPI, requireAuth, requireCustomer)

	// Subscriptions module
	svc.Subscriptions.RegisterStaffRoutes(api, requireAuth, requireStaff)
	svc.Subscriptions.RegisterCustomerRoutes(customerAPI, requireAuth, requireCustomer)

	// AI endpoints (text generation + self-hosted engine features)
	aiGroup := api.Group("/ai")
	svc.AI.RegisterRoutes(aiGroup, requireAuth, requireStaff, svc.BillingSvc.RequireFeature)

	// SEO endpoints
	svc.SEO.RegisterRoutes(api, requireAuth, requireStaff, svc.BillingSvc.RequireFeature)

	// ── Public routes (no tenant context required) ──
	publicAPI := r.Group("/api/v1")
	svc.Paymob.RegisterPublicRoutes(publicAPI)
	svc.Billing.RegisterPublicRoutes(publicAPI)

	return r
}
