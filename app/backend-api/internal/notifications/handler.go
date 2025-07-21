package notifications

import (
	"strconv"

	"backend-api/internal/shared/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup, requireAuth gin.HandlerFunc) {
	notif := rg.Group("/notifications")
	notif.Use(requireAuth)
	{
		notif.GET("", h.GetNotifications)
		notif.GET("/unread-count", h.GetUnreadCount)
		notif.PATCH("/:id/read", h.MarkRead)
		notif.PATCH("/read-all", h.MarkAllRead)
		notif.GET("/stream", h.SSEStream)
		notif.POST("/push/subscribe", h.RegisterPush)
		notif.DELETE("/push/subscribe", h.UnregisterPush)
	}
}

func (h *Handler) RegisterCustomerPrefsRoutes(rg *gin.RouterGroup, requireAuth gin.HandlerFunc) {
	prefs := rg.Group("/customer/notifications/prefs")
	prefs.Use(requireAuth)
	{
		prefs.GET("", h.GetNotifPrefs)
		prefs.PUT("", h.UpdateNotifPrefs)
	}
}

func (h *Handler) GetNotifications(c *gin.Context) {
	shopID, recipientType, recipientID, ok := extractActor(c)
	if !ok {
		response.Unauthorized(c, "unauthorized")
		return
	}

	limit := int32(20)
	if l := c.Query("limit"); l != "" {
		v, err := strconv.Atoi(l)
		if err == nil && v > 0 && v <= 100 {
			limit = int32(v)
		}
	}
	offset := int32(0)
	if o := c.Query("offset"); o != "" {
		v, err := strconv.Atoi(o)
		if err == nil && v >= 0 {
			offset = int32(v)
		}
	}

	items, err := h.svc.GetNotifications(c.Request.Context(), shopID, recipientType, recipientID, limit, offset)
	if err != nil {
		response.InternalError(c)
		return
	}
	response.OK(c, items)
}

func (h *Handler) GetUnreadCount(c *gin.Context) {
	shopID, recipientType, recipientID, ok := extractActor(c)
	if !ok {
		response.Unauthorized(c, "unauthorized")
		return
	}

	count, err := h.svc.GetUnreadCount(c.Request.Context(), shopID, recipientType, recipientID)
	if err != nil {
		response.InternalError(c)
		return
	}
	response.OK(c, gin.H{"count": count})
}

func (h *Handler) MarkRead(c *gin.Context) {
	shopID, _, _, ok := extractActor(c)
	if !ok {
		response.Unauthorized(c, "unauthorized")
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid notification id")
		return
	}

	if err := h.svc.MarkNotificationRead(c.Request.Context(), id, shopID); err != nil {
		response.InternalError(c)
		return
	}
	response.NoContent(c)
}

func (h *Handler) MarkAllRead(c *gin.Context) {
	shopID, recipientType, recipientID, ok := extractActor(c)
	if !ok {
		response.Unauthorized(c, "unauthorized")
		return
	}

	if err := h.svc.MarkAllNotificationsRead(c.Request.Context(), shopID, recipientType, recipientID); err != nil {
		response.InternalError(c)
		return
	}
	response.NoContent(c)
}

func (h *Handler) RegisterPush(c *gin.Context) {
	shopID, userType, userID, ok := extractActor(c)
	if !ok {
		response.Unauthorized(c, "unauthorized")
		return
	}

	var req struct {
		Endpoint  string `json:"endpoint"  binding:"required"`
		P256dh    string `json:"p256dh"    binding:"required"`
		AuthKey   string `json:"auth"      binding:"required"`
		UserAgent string `json:"userAgent"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := h.svc.UpsertPushSubscription(c.Request.Context(), shopID, userType, userID, req.Endpoint, req.P256dh, req.AuthKey, req.UserAgent); err != nil {
		response.InternalError(c)
		return
	}
	response.NoContent(c)
}

func (h *Handler) UnregisterPush(c *gin.Context) {
	shopID, _, _, ok := extractActor(c)
	if !ok {
		response.Unauthorized(c, "unauthorized")
		return
	}

	var req struct {
		Endpoint string `json:"endpoint" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := h.svc.DeletePushSubscription(c.Request.Context(), shopID, req.Endpoint); err != nil {
		response.InternalError(c)
		return
	}
	response.NoContent(c)
}

func (h *Handler) GetNotifPrefs(c *gin.Context) {
	_, _, recipientID, ok := extractActor(c)
	if !ok {
		response.Unauthorized(c, "unauthorized")
		return
	}

	prefs, err := h.svc.GetCustomerNotifPrefs(c.Request.Context(), recipientID)
	if err != nil {
		// Return defaults if not set yet.
		response.OK(c, gin.H{
			"customer_id":         recipientID,
			"email_order_updates": true,
			"email_marketing":     false,
			"push_order_updates":  true,
			"push_marketing":      false,
			"in_app_all":          true,
		})
		return
	}
	response.OK(c, prefs)
}

func (h *Handler) UpdateNotifPrefs(c *gin.Context) {
	_, _, recipientID, ok := extractActor(c)
	if !ok {
		response.Unauthorized(c, "unauthorized")
		return
	}

	var req struct {
		EmailOrderUpdates bool `json:"email_order_updates"`
		EmailMarketing    bool `json:"email_marketing"`
		PushOrderUpdates  bool `json:"push_order_updates"`
		PushMarketing     bool `json:"push_marketing"`
		InAppAll          bool `json:"in_app_all"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	prefs, err := h.svc.UpsertCustomerNotifPrefs(c.Request.Context(), recipientID,
		req.EmailOrderUpdates, req.EmailMarketing, req.PushOrderUpdates, req.PushMarketing, req.InAppAll)
	if err != nil {
		response.InternalError(c)
		return
	}
	response.OK(c, prefs)
}
