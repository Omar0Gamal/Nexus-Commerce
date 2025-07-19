package webhooks

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"backend-api/internal/db"
	"backend-api/internal/shared/apperr"
	"backend-api/internal/worker"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)


const (
	TopicOrderCreated       = "order.created"
	TopicOrderStatusChanged = "order.status_changed"
	TopicProductLowStock    = "product.low_stock"
)

var validTopics = map[string]bool{
	TopicOrderCreated:       true,
	TopicOrderStatusChanged: true,
	TopicProductLowStock:    true,
}


var (
	ErrNotFound    = apperr.ErrNotFound
	ErrInvalidUUID = apperr.ErrInvalidUUID
	ErrBadTopic    = errors.New("invalid or unsupported webhook topic")
)


// WebhookResponse is the API representation of a webhook subscription.
type WebhookResponse struct {
	ID        string `json:"id"`
	Topic     string `json:"topic"`
	TargetURL string `json:"target_url"`
	IsActive  bool   `json:"is_active"`
	// SecretKey is intentionally omitted from responses for security.
}

// CreateWebhookRequest is the payload for registering a new webhook.
type CreateWebhookRequest struct {
	Topic     string `json:"topic"      binding:"required"`
	TargetURL string `json:"target_url" binding:"required,url"`
	SecretKey string `json:"secret_key"` // optional; HMAC-SHA256 signing key
	IsActive  *bool  `json:"is_active"`  // defaults to true
}

// UpdateWebhookRequest supports partial updates.
type UpdateWebhookRequest struct {
	Topic     *string `json:"topic"`
	TargetURL *string `json:"target_url"`
	SecretKey *string `json:"secret_key"`
	IsActive  *bool   `json:"is_active"`
}

// Event is the body sent in a webhook POST call.
type Event struct {
	Topic     string         `json:"topic"`
	ShopID    string         `json:"shop_id"`
	Timestamp string         `json:"timestamp"`
	Data      map[string]any `json:"data"`
}


// Service manages webhook subscriptions and dispatching.
type Service struct {
	q          *db.Queries
	httpClient *http.Client
	jobs       *worker.Queue
	encKey     []byte
}

func NewService(q *db.Queries, jobs *worker.Queue, encKey []byte) *Service {
	return &Service{
		q:      q,
		jobs:   jobs,
		encKey: encKey,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}


// CreateWebhook registers a new webhook subscription for a shop.
func (s *Service) CreateWebhook(ctx context.Context, shopID string, req CreateWebhookRequest) (*WebhookResponse, error) {
	if !validTopics[req.Topic] {
		return nil, ErrBadTopic
	}

	shopUUID, err := uuid.Parse(shopID)
	if err != nil {
		return nil, ErrInvalidUUID
	}

	active := true
	if req.IsActive != nil {
		active = *req.IsActive
	}

	encryptedSecret, err := encryptSecret(s.encKey, req.SecretKey)
	if err != nil {
		return nil, fmt.Errorf("encrypt webhook secret: %w", err)
	}

	wh, err := s.q.CreateWebhook(ctx, db.CreateWebhookParams{
		ShopID:    pgtype.UUID{Bytes: shopUUID, Valid: true},
		Topic:     req.Topic,
		TargetUrl: req.TargetURL,
		SecretKey: encryptedSecret,
		IsActive:  pgtype.Bool{Bool: active, Valid: true},
	})
	if err != nil {
		return nil, fmt.Errorf("create webhook: %w", err)
	}

	return toResponse(wh), nil
}

// GetWebhook retrieves a single webhook by ID, scoped to the shop.
func (s *Service) GetWebhook(ctx context.Context, shopID, webhookID string) (*WebhookResponse, error) {
	shopUUID, err := uuid.Parse(shopID)
	if err != nil {
		return nil, ErrInvalidUUID
	}
	whUUID, err := uuid.Parse(webhookID)
	if err != nil {
		return nil, ErrInvalidUUID
	}

	wh, err := s.q.GetWebhook(ctx, db.GetWebhookParams{
		ID:     whUUID,
		ShopID: pgtype.UUID{Bytes: shopUUID, Valid: true},
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get webhook: %w", err)
	}

	return toResponse(wh), nil
}

// ListWebhooks returns all webhooks for a shop.
func (s *Service) ListWebhooks(ctx context.Context, shopID string) ([]WebhookResponse, error) {
	shopUUID, err := uuid.Parse(shopID)
	if err != nil {
		return nil, ErrInvalidUUID
	}

	rows, err := s.q.ListWebhooks(ctx, pgtype.UUID{Bytes: shopUUID, Valid: true})
	if err != nil {
		return nil, fmt.Errorf("list webhooks: %w", err)
	}

	result := make([]WebhookResponse, len(rows))
	for i, wh := range rows {
		result[i] = *toResponse(wh)
	}
	return result, nil
}

// UpdateWebhook applies partial updates to a webhook.
func (s *Service) UpdateWebhook(ctx context.Context, shopID, webhookID string, req UpdateWebhookRequest) (*WebhookResponse, error) {
	shopUUID, err := uuid.Parse(shopID)
	if err != nil {
		return nil, ErrInvalidUUID
	}
	whUUID, err := uuid.Parse(webhookID)
	if err != nil {
		return nil, ErrInvalidUUID
	}

	if req.Topic != nil && !validTopics[*req.Topic] {
		return nil, ErrBadTopic
	}

	params := db.UpdateWebhookParams{
		ID:     whUUID,
		ShopID: pgtype.UUID{Bytes: shopUUID, Valid: true},
	}
	if req.Topic != nil {
		params.Topic = pgtype.Text{String: *req.Topic, Valid: true}
	}
	if req.TargetURL != nil {
		params.TargetUrl = pgtype.Text{String: *req.TargetURL, Valid: true}
	}
	if req.SecretKey != nil {
		enc, err := encryptSecret(s.encKey, *req.SecretKey)
		if err != nil {
			return nil, fmt.Errorf("encrypt webhook secret: %w", err)
		}
		params.SecretKey = pgtype.Text{String: enc, Valid: true}
	}
	if req.IsActive != nil {
		params.IsActive = pgtype.Bool{Bool: *req.IsActive, Valid: true}
	}

	wh, err := s.q.UpdateWebhook(ctx, params)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("update webhook: %w", err)
	}

	return toResponse(wh), nil
}

// ToggleWebhook flips the is_active flag on a webhook.
func (s *Service) ToggleWebhook(ctx context.Context, shopID, webhookID string) (*WebhookResponse, error) {
	shopUUID, err := uuid.Parse(shopID)
	if err != nil {
		return nil, ErrInvalidUUID
	}
	whUUID, err := uuid.Parse(webhookID)
	if err != nil {
		return nil, ErrInvalidUUID
	}

	wh, err := s.q.ToggleWebhook(ctx, db.ToggleWebhookParams{
		ID:     whUUID,
		ShopID: pgtype.UUID{Bytes: shopUUID, Valid: true},
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("toggle webhook: %w", err)
	}

	return toResponse(wh), nil
}

// DeleteWebhook removes a webhook subscription.
func (s *Service) DeleteWebhook(ctx context.Context, shopID, webhookID string) error {
	shopUUID, err := uuid.Parse(shopID)
	if err != nil {
		return ErrInvalidUUID
	}
	whUUID, err := uuid.Parse(webhookID)
	if err != nil {
		return ErrInvalidUUID
	}

	return s.q.DeleteWebhook(ctx, db.DeleteWebhookParams{
		ID:     whUUID,
		ShopID: pgtype.UUID{Bytes: shopUUID, Valid: true},
	})
}


// Dispatch sends an event to all active webhook subscriptions for the given
// shop and topic. Each delivery is attempted in a separate goroutine so the
// caller is never blocked. Failures are silently swallowed (best-effort).
func (s *Service) Dispatch(shopID, topic string, data map[string]any) {
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		shopUUID, err := uuid.Parse(shopID)
		if err != nil {
			return
		}

		subscribers, err := s.q.ListWebhooksByTopic(ctx, db.ListWebhooksByTopicParams{
			ShopID: pgtype.UUID{Bytes: shopUUID, Valid: true},
			Topic:  topic,
		})
		if err != nil {
			return
		}

		for _, wh := range subscribers {
			wh := wh // capture
			plainSecret := decryptSecret(s.encKey, wh.SecretKey)
			if s.jobs != nil {
				_ = s.jobs.EnqueueTyped(ctx, worker.JobDispatchWebhook, worker.WebhookDispatchPayload{
					TargetURL: wh.TargetUrl,
					Topic:     topic,
					SecretKey: plainSecret,
					Body: map[string]any{
						"topic":     topic,
						"shop_id":   shopID,
						"timestamp": time.Now().UTC().Format(time.RFC3339),
						"data":      data,
					},
				})
			}
		}
	}()
}


func toResponse(wh db.Webhook) *WebhookResponse {
	return &WebhookResponse{
		ID:        wh.ID.String(),
		Topic:     wh.Topic,
		TargetURL: wh.TargetUrl,
		IsActive:  wh.IsActive.Bool,
	}
}
