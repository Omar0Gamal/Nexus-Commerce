package auth

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestFailedLoginTTL(t *testing.T) {
	tests := []struct {
		count int
		want  time.Duration
	}{
		{1, time.Minute},
		{2, 5 * time.Minute},
		{3, 15 * time.Minute},
		{5, time.Hour},
		{9, time.Hour},
	}

	for _, tt := range tests {
		if got := failedLoginTTL(tt.count); got != tt.want {
			t.Fatalf("failedLoginTTL(%d) = %v, want %v", tt.count, got, tt.want)
		}
	}
}

func TestBruteForceFallbackRateLimitAndClear(t *testing.T) {
	svc := &Service{}
	ctx := context.Background()
	key := "failed_login:test@example.com"

	for i := 0; i < maxLoginAttempts; i++ {
		svc.recordFailedLogin(ctx, key)
	}

	if err := svc.checkBruteForce(ctx, key); !errors.Is(err, ErrRateLimited) {
		t.Fatalf("expected ErrRateLimited, got %v", err)
	}

	svc.clearFailedLogin(ctx, key)
	if err := svc.checkBruteForce(ctx, key); err != nil {
		t.Fatalf("expected no error after clear, got %v", err)
	}
}

func TestBruteForceFallbackExpiry(t *testing.T) {
	svc := &Service{}
	ctx := context.Background()
	key := "failed_login:expired@example.com"

	svc.loginFail.Store(key, fallbackFailedLogin{
		Count:     maxLoginAttempts,
		ExpiresAt: time.Now().Add(-1 * time.Minute),
	})

	if err := svc.checkBruteForce(ctx, key); err != nil {
		t.Fatalf("expected expired fallback entry to be ignored, got %v", err)
	}

	if _, ok := svc.loginFail.Load(key); ok {
		t.Fatal("expected expired fallback entry to be removed")
	}
}
