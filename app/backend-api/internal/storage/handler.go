package storage

import (
	"fmt"
	"time"

	"backend-api/internal/shared/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Handler struct {
	client *Client
}

func NewHandler(client *Client) *Handler {
	return &Handler{client: client}
}

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup, requireAuth gin.HandlerFunc, requireStaff gin.HandlerFunc) {
	storageGroup := rg.Group("/storage", requireAuth, requireStaff)
	{
		storageGroup.POST("/presigned", h.GeneratePresignedURL)
	}
}

type PresignedURLRequest struct {
	FileName    string `json:"file_name" binding:"required"`
	ContentType string `json:"content_type" binding:"required"`
}

// GeneratePresignedURL generates a Cloudflare R2 pre-signed URL to allow direct browser uploads
// for product images or merchant logos, bypassing the Go backend.
func (h *Handler) GeneratePresignedURL(c *gin.Context) {
	shopID := c.GetString("shop_id")
	if shopID == "" {
		response.Unauthorized(c, "missing shop context")
		return
	}

	var req PresignedURLRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	timestamp := time.Now().Unix()
	uniqueID := uuid.New().String()
	key := fmt.Sprintf("shops/%s/uploads/%d_%s_%s", shopID, timestamp, uniqueID, req.FileName)

	uploadURL, publicURL, err := h.client.PresignUpload(c.Request.Context(), key, req.ContentType, 900)
	if err != nil {
		response.InternalError(c)
		return
	}

	response.OK(c, gin.H{
		"upload_url": uploadURL,
		"public_url": publicURL,
		"key":        key,
	})
}
