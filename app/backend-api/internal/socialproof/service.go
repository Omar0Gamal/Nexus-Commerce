package socialproof

import (
	"context"
	"fmt"
	"time"

	"backend-api/internal/db"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/redis/go-redis/v9"
)

// Service handles social proof event storage and retrieval.
type Service struct {
	q   *db.Queries
	rdb *redis.Client
}

func NewService(q *db.Queries, rdb *redis.Client) *Service {
	return &Service{q: q, rdb: rdb}
}

// RecordSale inserts a social proof event when an order is completed.
func (s *Service) RecordSale(ctx context.Context, shopID, productID, city, country string) error {
	shopUUID, err := uuid.Parse(shopID)
	if err != nil {
		return fmt.Errorf("invalid shop_id: %w", err)
	}
	productUUID, err := uuid.Parse(productID)
	if err != nil {
		return fmt.Errorf("invalid product_id: %w", err)
	}

	return s.q.InsertSocialProofEvent(ctx, db.InsertSocialProofEventParams{
		ShopID:    shopUUID,
		ProductID: productUUID,
		City:      pgtype.Text{String: city, Valid: city != ""},
		Country:   pgtype.Text{String: country, Valid: country != ""},
	})
}

// GetRecent returns the latest N social proof events for a shop.
func (s *Service) GetRecent(ctx context.Context, shopID string, limit int32) ([]RecentEvent, error) {
	shopUUID, err := uuid.Parse(shopID)
	if err != nil {
		return nil, fmt.Errorf("invalid shop_id: %w", err)
	}

	if limit <= 0 || limit > 20 {
		limit = 5
	}

	rows, err := s.q.GetRecentSocialProofEvents(ctx, db.GetRecentSocialProofEventsParams{
		ShopID: shopUUID,
		Limit:  limit,
	})
	if err != nil {
		return nil, fmt.Errorf("get recent events: %w", err)
	}

	events := make([]RecentEvent, 0, len(rows))
	for _, r := range rows {
		e := RecentEvent{
			ProductID:    r.ProductID.String(),
			ProductTitle: r.ProductTitle,
			ProductSlug:  r.ProductSlug,
			CreatedAt:    r.CreatedAt.Time,
		}
		if r.City.Valid {
			e.City = r.City.String
		}
		if r.Country.Valid {
			e.Country = r.Country.String
		}
		events = append(events, e)
	}
	return events, nil
}

// Purge removes social proof events older than 7 days.
func (s *Service) Purge(ctx context.Context, shopID string) error {
	shopUUID, err := uuid.Parse(shopID)
	if err != nil {
		return err
	}
	return s.q.PurgeStaleSocialProofEvents(ctx, shopUUID)
}

// TrackShare increments the share counter for a product+channel in Redis.
// Key format: stats:shares:{shopID}:{YYYY-MM-DD}:{channel}  (sorted set, member = productID)
func (s *Service) TrackShare(ctx context.Context, shopID, productID, channel string) error {
	if s.rdb == nil {
		return nil
	}
	date := time.Now().UTC().Format("2006-01-02")
	key := fmt.Sprintf("stats:shares:%s:%s:%s", shopID, date, channel)
	return s.rdb.ZIncrBy(ctx, key, 1, productID).Err()
}

// ShareStats holds aggregated share counts.
type ShareStats struct {
	Channels    map[string]int64    `json:"channels"`
	TopProducts []ProductShareCount `json:"top_products"`
}

// ProductShareCount is a product with its total share count.
type ProductShareCount struct {
	ProductID string `json:"product_id"`
	Count     int64  `json:"count"`
}

// GetShareStats aggregates share counts from Redis for a date range.
func (s *Service) GetShareStats(ctx context.Context, shopID, startDate, endDate string) (*ShareStats, error) {
	if s.rdb == nil {
		return &ShareStats{Channels: map[string]int64{}, TopProducts: []ProductShareCount{}}, nil
	}

	channels := []string{"facebook", "twitter", "whatsapp", "copy"}
	stats := &ShareStats{
		Channels:    make(map[string]int64),
		TopProducts: []ProductShareCount{},
	}

	productTotals := make(map[string]int64)

	// Parse date range
	start, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		start = time.Now().UTC().AddDate(0, 0, -30)
	}
	end, err := time.Parse("2006-01-02", endDate)
	if err != nil {
		end = time.Now().UTC()
	}

	for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
		dateStr := d.Format("2006-01-02")
		for _, ch := range channels {
			key := fmt.Sprintf("stats:shares:%s:%s:%s", shopID, dateStr, ch)
			members, err := s.rdb.ZRangeWithScores(ctx, key, 0, -1).Result()
			if err != nil {
				continue
			}
			for _, m := range members {
				stats.Channels[ch] += int64(m.Score)
				productTotals[m.Member.(string)] += int64(m.Score)
			}
		}
	}

	// Build top products (top 10)
	type kv struct {
		k string
		v int64
	}
	kvs := make([]kv, 0, len(productTotals))
	for k, v := range productTotals {
		kvs = append(kvs, kv{k, v})
	}
	// Simple sort by count descending
	for i := 0; i < len(kvs); i++ {
		for j := i + 1; j < len(kvs); j++ {
			if kvs[j].v > kvs[i].v {
				kvs[i], kvs[j] = kvs[j], kvs[i]
			}
		}
	}
	limit := 10
	if len(kvs) < limit {
		limit = len(kvs)
	}
	for _, item := range kvs[:limit] {
		stats.TopProducts = append(stats.TopProducts, ProductShareCount{
			ProductID: item.k,
			Count:     item.v,
		})
	}

	return stats, nil
}
