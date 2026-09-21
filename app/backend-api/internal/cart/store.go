package cart

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	guestCartTTL    = 7 * 24 * time.Hour  // 7 days for anonymous carts
	customerCartTTL = 30 * 24 * time.Hour // 30 days for logged-in customers
)

type Store struct {
	rdb *redis.Client
}

func NewStore(rdb *redis.Client) *Store {
	return &Store{rdb: rdb}
}

// Key format:
// Customer: cart:{shopID}:customer:{customerID}
// Guest:    cart:{shopID}:guest:{cartToken}
func cartKey(shopID, identity string) string {
	return fmt.Sprintf("cart:%s:%s", shopID, identity)
}

func (s *Store) Get(ctx context.Context, shopID, identity string) ([]CartItem, error) {
	key := cartKey(shopID, identity)
	rawJSON, err := s.rdb.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return []CartItem{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("redis get cart: %w", err)
	}

	var items []CartItem
	if err := json.Unmarshal(rawJSON, &items); err != nil {
		return nil, fmt.Errorf("unmarshal cart: %w", err)
	}
	return items, nil
}

func (s *Store) Save(ctx context.Context, shopID, identity string, items []CartItem, isCustomer bool) error {
	key := cartKey(shopID, identity)
	rawJSON, err := json.Marshal(items)
	if err != nil {
		return fmt.Errorf("marshal cart: %w", err)
	}

	ttl := guestCartTTL
	if isCustomer {
		ttl = customerCartTTL
	}

	return s.rdb.Set(ctx, key, rawJSON, ttl).Err()
}

func (s *Store) Delete(ctx context.Context, shopID, identity string) error {
	return s.rdb.Del(ctx, cartKey(shopID, identity)).Err()
}
