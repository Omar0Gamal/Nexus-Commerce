package notifications

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"backend-api/internal/db"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// Event is the canonical structure for fanning out a notification.
type Event struct {
	ShopID        uuid.UUID
	RecipientType string // "staff" | "customer"
	RecipientID   uuid.UUID
	EventType     string // "order_placed", "order_shipped", "review_pending", etc.
	Channel       string // "in_app" | "email" | "push" — may be empty to fan-out all
	Title         string
	Body          string
	ActionURL     string
	Metadata      map[string]any
}

// Service handles creating and delivering notifications.
type Service struct {
	queries *db.Queries
	rdb     *redis.Client
	push    *PushSender
	logger  *zap.Logger
}

// vapidPublic/vapidPrivate/vapidEmail are optional — omit to disable Web Push.
func NewService(queries *db.Queries, rdb *redis.Client, vapidPublic, vapidPrivate, vapidEmail string, logger *zap.Logger) *Service {
	var pusher *PushSender
	if vapidPublic != "" && vapidPrivate != "" {
		pusher = NewPushSender(queries, vapidPublic, vapidPrivate, vapidEmail, logger)
	}
	return &Service{queries: queries, rdb: rdb, push: pusher, logger: logger}
}

// Notify fans out a notification to in-app (always), and optionally push.
// Email fan-out is handled separately by the worker system.
func (s *Service) Notify(ctx context.Context, e Event) error {
	metadata, err := json.Marshal(e.Metadata)
	if err != nil {
		metadata = []byte("{}")
	}

	actionURL := pgtype.Text{}
	if e.ActionURL != "" {
		actionURL = pgtype.Text{String: e.ActionURL, Valid: true}
	}

	channel := e.Channel
	if channel == "" {
		channel = "in_app"
	}

	n, err := s.queries.CreateNotification(ctx, db.CreateNotificationParams{
		ShopID:        e.ShopID,
		RecipientType: e.RecipientType,
		RecipientID:   e.RecipientID,
		Channel:       channel,
		EventType:     e.EventType,
		Title:         e.Title,
		Body:          e.Body,
		ActionUrl:     actionURL,
		Metadata:      metadata,
	})
	if err != nil {
		return fmt.Errorf("notifications.Notify: create record: %w", err)
	}

	// Publish to Redis for SSE subscribers.
	s.publishSSE(ctx, e.ShopID, e.RecipientType, e.RecipientID, n)

	// Fan out Web Push if available.
	if s.push != nil {
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			if pushErr := s.push.Send(ctx, e.ShopID, e.RecipientType, e.RecipientID, e.Title, e.Body); pushErr != nil {
				s.logger.Error("push send failed", zap.Error(pushErr))
			}
		}()
	}

	return nil
}

// sseChannel returns the Redis Pub/Sub channel name for a recipient.
func sseChannel(shopID uuid.UUID, recipientType string, recipientID uuid.UUID) string {
	return fmt.Sprintf("notify:%s:%s:%s", shopID, recipientType, recipientID)
}

// publishSSE publishes to Redis so any open SSE stream picks it up.
func (s *Service) publishSSE(ctx context.Context, shopID uuid.UUID, recipientType string, recipientID uuid.UUID, n db.Notification) {
	if s.rdb == nil {
		return
	}
	payload, err := json.Marshal(n)
	if err != nil {
		return
	}
	ch := sseChannel(shopID, recipientType, recipientID)
	if err := s.rdb.Publish(ctx, ch, payload).Err(); err != nil {
		s.logger.Error("Redis publish failed",
			zap.String("channel", ch),
			zap.Error(err))
	}
}


// GetNotifications retrieves paginated notifications for a recipient.
func (s *Service) GetNotifications(ctx context.Context, shopID uuid.UUID, recipientType string, recipientID uuid.UUID, limit, offset int32) ([]db.Notification, error) {
	items, err := s.queries.GetNotifications(ctx, db.GetNotificationsParams{
		ShopID:        shopID,
		RecipientType: recipientType,
		RecipientID:   recipientID,
		Limit:         limit,
		Offset:        offset,
	})
	if err != nil {
		return nil, err
	}
	if items == nil {
		return []db.Notification{}, nil
	}
	return items, nil
}

// GetUnreadCount returns the number of unread notifications for a recipient.
func (s *Service) GetUnreadCount(ctx context.Context, shopID uuid.UUID, recipientType string, recipientID uuid.UUID) (int64, error) {
	return s.queries.GetUnreadCount(ctx, db.GetUnreadCountParams{
		ShopID:        shopID,
		RecipientType: recipientType,
		RecipientID:   recipientID,
	})
}

// MarkNotificationRead marks a single notification as read.
func (s *Service) MarkNotificationRead(ctx context.Context, id, shopID uuid.UUID) error {
	return s.queries.MarkNotificationRead(ctx, db.MarkNotificationReadParams{
		ID:     id,
		ShopID: shopID,
	})
}

// MarkAllNotificationsRead marks all notifications as read for a recipient.
func (s *Service) MarkAllNotificationsRead(ctx context.Context, shopID uuid.UUID, recipientType string, recipientID uuid.UUID) error {
	return s.queries.MarkAllNotificationsRead(ctx, db.MarkAllNotificationsReadParams{
		ShopID:        shopID,
		RecipientType: recipientType,
		RecipientID:   recipientID,
	})
}

// UpsertPushSubscription registers or updates a Web Push subscription.
func (s *Service) UpsertPushSubscription(ctx context.Context, shopID uuid.UUID, userType string, userID uuid.UUID, endpoint, p256dh, authKey, userAgent string) error {
	ua := pgtype.Text{}
	if userAgent != "" {
		ua = pgtype.Text{String: userAgent, Valid: true}
	}
	return s.queries.UpsertPushSubscription(ctx, db.UpsertPushSubscriptionParams{
		ShopID:    shopID,
		UserType:  userType,
		UserID:    userID,
		Endpoint:  endpoint,
		P256dh:    p256dh,
		AuthKey:   authKey,
		UserAgent: ua,
	})
}

// DeletePushSubscription removes a Web Push subscription by endpoint.
func (s *Service) DeletePushSubscription(ctx context.Context, shopID uuid.UUID, endpoint string) error {
	return s.queries.DeletePushSubscription(ctx, db.DeletePushSubscriptionParams{
		Endpoint: endpoint,
		ShopID:   shopID,
	})
}

// GetCustomerNotifPrefs returns a customer's notification preferences.
func (s *Service) GetCustomerNotifPrefs(ctx context.Context, customerID uuid.UUID) (db.CustomerNotificationPref, error) {
	return s.queries.GetCustomerNotifPrefs(ctx, customerID)
}

// UpsertCustomerNotifPrefs saves a customer's notification preferences.
func (s *Service) UpsertCustomerNotifPrefs(ctx context.Context, customerID uuid.UUID, emailOrderUpdates, emailMarketing, pushOrderUpdates, pushMarketing, inAppAll bool) (db.CustomerNotificationPref, error) {
	return s.queries.UpsertCustomerNotifPrefs(ctx, db.UpsertCustomerNotifPrefsParams{
		CustomerID:        customerID,
		EmailOrderUpdates: emailOrderUpdates,
		EmailMarketing:    emailMarketing,
		PushOrderUpdates:  pushOrderUpdates,
		PushMarketing:     pushMarketing,
		InAppAll:          inAppAll,
	})
}
