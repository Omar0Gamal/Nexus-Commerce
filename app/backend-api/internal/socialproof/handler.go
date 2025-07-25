package socialproof

import (
	"fmt"
	"net/url"
	"strconv"

	"backend-api/internal/shared/response"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.GET("/social-proof/recent", h.GetRecent)
	rg.POST("/social-proof/share/:product_id", h.Share)
}

func (h *Handler) GetRecent(c *gin.Context) {
	shopID := c.GetString("shop_id")

	limit := int32(5)
	if v := c.Query("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 20 {
			limit = int32(n)
		}
	}

	events, err := h.svc.GetRecent(c.Request.Context(), shopID, limit)
	if err != nil {
		response.InternalError(c)
		return
	}
	response.OK(c, events)
}

// Returns share URLs for the chosen channel and increments the share counter.
func (h *Handler) Share(c *gin.Context) {
	shopID := c.GetString("shop_id")
	productID := c.Param("product_id")
	channel := c.DefaultQuery("channel", "copy")

	// Track asynchronously (fire-and-forget)
	go func() {
		_ = h.svc.TrackShare(c.Request.Context(), shopID, productID, channel)
	}()

	// Build the product page URL from the Referer or a sensible default
	referer := c.GetHeader("Referer")
	productURL := referer
	if productURL == "" {
		productURL = fmt.Sprintf("/products/%s", productID)
	}

	encoded := url.QueryEscape(productURL)
	shareURLs := map[string]string{
		"url":      productURL,
		"facebook": fmt.Sprintf("https://www.facebook.com/sharer/sharer.php?u=%s", encoded),
		"twitter":  fmt.Sprintf("https://twitter.com/intent/tweet?url=%s", encoded),
		"whatsapp": fmt.Sprintf("https://wa.me/?text=%s", encoded),
		"copy":     productURL,
	}

	platform, ok := shareURLs[channel]
	if !ok {
		platform = productURL
	}

	response.OK(c, gin.H{
		"channel":   channel,
		"share_url": platform,
		"urls":      shareURLs,
	})
}
