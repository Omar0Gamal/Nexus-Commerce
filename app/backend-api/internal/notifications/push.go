package notifications

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"backend-api/internal/db"

	webpush "github.com/SherClockHolmes/webpush-go"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// PushSender delivers VAPID Web Push notifications.
type PushSender struct {
	queries      *db.Queries
	vapidPublic  string
	vapidPrivate string
	vapidEmail   string
	logger       *zap.Logger
}

func NewPushSender(queries *db.Queries, public, private, email string, logger *zap.Logger) *PushSender {
	return &PushSender{
		queries:      queries,
		vapidPublic:  public,
		vapidPrivate: private,
		vapidEmail:   email,
		logger:       logger,
	}
}

// Send fetches all push subscriptions for the recipient and delivers the notification.
func (p *PushSender) Send(ctx context.Context, shopID uuid.UUID, userType string, userID uuid.UUID, title, body string) error {
	subs, err := p.queries.GetPushSubscriptions(ctx, db.GetPushSubscriptionsParams{
		ShopID:   shopID,
		UserType: userType,
		UserID:   userID,
	})
	if err != nil {
		return fmt.Errorf("push.GetPushSubscriptions: %w", err)
	}

	payload, _ := json.Marshal(map[string]string{"title": title, "body": body})

	for _, sub := range subs {
		if err := p.sendOne(ctx, shopID, sub, payload); err != nil {
			p.logger.Error("push send", zap.String("endpoint", sub.Endpoint), zap.Error(err))
		}
	}
	return nil
}

func (p *PushSender) sendOne(ctx context.Context, shopID uuid.UUID, sub db.PushSubscription, payload []byte) error {
	resp, err := webpush.SendNotification(payload, &webpush.Subscription{
		Endpoint: sub.Endpoint,
		Keys: webpush.Keys{
			Auth:   sub.AuthKey,
			P256dh: sub.P256dh,
		},
	}, &webpush.Options{
		Subscriber:      p.vapidEmail,
		VAPIDPublicKey:  p.vapidPublic,
		VAPIDPrivateKey: p.vapidPrivate,
		TTL:             86400,
	})
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// 410 Gone — the subscription is expired; remove it.
	if resp.StatusCode == http.StatusGone {
		if delErr := p.queries.DeletePushSubscription(ctx, db.DeletePushSubscriptionParams{
			Endpoint: sub.Endpoint,
			ShopID:   shopID,
		}); delErr != nil {
			p.logger.Error("push delete expired subscription", zap.Error(delErr))
		}
	}
	return nil
}
