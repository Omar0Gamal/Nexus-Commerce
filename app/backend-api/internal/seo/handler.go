package seo

import (
	"errors"
	"net/http"
	"strconv"

	"backend-api/internal/shared/apperr"
	"backend-api/internal/shared/pgutil"
	"backend-api/internal/shared/response"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// These must be registered before TenantMiddleware is applied.
func (h *Handler) RegisterPublicRoutes(r *gin.Engine) {
	// Tenant context is injected by the gateway via X-Shop-ID header and is
	// available when TenantMiddleware runs. These routes need it, so they are
	// mounted on a group that uses TenantMiddleware independently.
	pub := r.Group("/")
	pub.Use(shopContextFromHeader())
	pub.GET("/sitemap.xml", h.Sitemap)
	pub.GET("/robots.txt", h.RobotsTxt)
}

func (h *Handler) RegisterRoutes(api *gin.RouterGroup, requireAuth, requireStaff gin.HandlerFunc, requireFeature func(string) gin.HandlerFunc) {
	seo := api.Group("/seo", requireAuth, requireStaff)
	{
		// SEO audit — gated by seo_scoring feature flag
		seo.GET("/audit", requireFeature("seo_scoring"), h.GetAuditReport)
		seo.POST("/audit/run", requireFeature("seo_scoring"), h.RunAudit)

		// Autofill blank SEO fields from title/description
		seo.POST("/autofill", requireFeature("seo_advanced"), h.AutofillSEO)

		// URL redirects — gated by seo_redirects feature flag
		seo.GET("/redirects", requireFeature("seo_redirects"), h.ListRedirects)
		seo.POST("/redirects", requireFeature("seo_redirects"), h.CreateRedirect)
		seo.DELETE("/redirects/:id", requireFeature("seo_redirects"), h.DeleteRedirect)
	}
}


func (h *Handler) Sitemap(c *gin.Context) {
	shopID, err := pgutil.ParseStdUUID(c.GetString("shop_id"))
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}
	baseURL := c.GetString("shop_base_url")
	if baseURL == "" {
		baseURL = "https://" + c.Request.Host
	}

	xml, err := h.svc.GenerateSitemap(c.Request.Context(), shopID, baseURL)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}
	c.Data(http.StatusOK, "application/xml; charset=utf-8", xml)
}

func (h *Handler) RobotsTxt(c *gin.Context) {
	baseURL := c.GetString("shop_base_url")
	if baseURL == "" {
		baseURL = "https://" + c.Request.Host
	}
	text := h.svc.Robots(c.Request.Context(), baseURL)
	c.Data(http.StatusOK, "text/plain; charset=utf-8", []byte(text))
}


func (h *Handler) RunAudit(c *gin.Context) {
	shopID, err := pgutil.ParseStdUUID(c.GetString("shop_id"))
	if err != nil {
		response.BadRequest(c, "invalid shop context")
		return
	}
	count, err := h.svc.RunSEOAudit(c.Request.Context(), shopID)
	if err != nil {
		response.InternalError(c)
		return
	}
	response.OK(c, gin.H{"products_scored": count})
}

func (h *Handler) GetAuditReport(c *gin.Context) {
	shopID, err := pgutil.ParseStdUUID(c.GetString("shop_id"))
	if err != nil {
		response.BadRequest(c, "invalid shop context")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	rows, total, err := h.svc.GetAuditReport(c.Request.Context(), shopID, page, limit)
	if err != nil {
		response.InternalError(c)
		return
	}
	response.OK(c, gin.H{
		"data":  rows,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

func (h *Handler) AutofillSEO(c *gin.Context) {
	shopID, err := pgutil.ParseStdUUID(c.GetString("shop_id"))
	if err != nil {
		response.BadRequest(c, "invalid shop context")
		return
	}
	count, err := h.svc.AutofillSEO(c.Request.Context(), shopID)
	if err != nil {
		response.InternalError(c)
		return
	}
	response.OK(c, gin.H{"products_updated": count})
}


type createRedirectRequest struct {
	FromPath   string `json:"from_path" binding:"required"`
	ToPath     string `json:"to_path" binding:"required"`
	StatusCode int16  `json:"status_code"`
}

func (h *Handler) CreateRedirect(c *gin.Context) {
	shopID, err := pgutil.ParseStdUUID(c.GetString("shop_id"))
	if err != nil {
		response.BadRequest(c, "invalid shop context")
		return
	}

	var req createRedirectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	r, err := h.svc.CreateRedirect(c.Request.Context(), shopID, req.FromPath, req.ToPath, req.StatusCode)
	if err != nil {
		response.InternalError(c)
		return
	}
	response.Created(c, r)
}

func (h *Handler) ListRedirects(c *gin.Context) {
	shopID, err := pgutil.ParseStdUUID(c.GetString("shop_id"))
	if err != nil {
		response.BadRequest(c, "invalid shop context")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))

	rows, total, err := h.svc.ListRedirects(c.Request.Context(), shopID, page, limit)
	if err != nil {
		response.InternalError(c)
		return
	}
	response.OK(c, gin.H{
		"data":  rows,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

func (h *Handler) DeleteRedirect(c *gin.Context) {
	shopID, err := pgutil.ParseStdUUID(c.GetString("shop_id"))
	if err != nil {
		response.BadRequest(c, "invalid shop context")
		return
	}

	id, err := pgutil.ParseStdUUID(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid redirect id")
		return
	}

	if err := h.svc.DeleteRedirect(c.Request.Context(), shopID, id); err != nil {
		if errors.Is(err, apperr.ErrNotFound) {
			response.NotFound(c, "redirect not found")
			return
		}
		response.InternalError(c)
		return
	}
	response.NoContent(c)
}


// shopContextFromHeader is a lightweight middleware that extracts X-Shop-ID
// and X-Shop-Base-URL headers into the Gin context for the public sitemap/robots routes.
func shopContextFromHeader() gin.HandlerFunc {
	return func(c *gin.Context) {
		if shopID := c.GetHeader("X-Shop-ID"); shopID != "" {
			c.Set("shop_id", shopID)
		}
		if baseURL := c.GetHeader("X-Shop-Base-URL"); baseURL != "" {
			c.Set("shop_base_url", baseURL)
		}
		c.Next()
	}
}
