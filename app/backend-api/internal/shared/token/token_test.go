package token_test

import (
	"strings"
	"testing"
	"time"

	"backend-api/internal/shared/token"
)

func TestGenerateAndValidateAccessToken(t *testing.T) {
	svc := token.NewService("test-secret-key")

	claims := token.Claims{
		UserID:    "user-123",
		Email:     "test@example.com",
		FullName:  "Test User",
		ActorType: token.ActorStaff,
		ShopID:    "shop-456",
		Role:      "admin",
	}

	tok, err := svc.GenerateAccessToken(claims)
	if err != nil {
		t.Fatalf("GenerateAccessToken: %v", err)
	}
	if tok == "" {
		t.Fatal("expected non-empty token")
	}

	parsed, err := svc.ValidateToken(tok)
	if err != nil {
		t.Fatalf("ValidateToken: %v", err)
	}
	if parsed.UserID != claims.UserID {
		t.Errorf("UserID: got %q, want %q", parsed.UserID, claims.UserID)
	}
	if parsed.Email != claims.Email {
		t.Errorf("Email: got %q, want %q", parsed.Email, claims.Email)
	}
	if parsed.ActorType != claims.ActorType {
		t.Errorf("ActorType: got %q, want %q", parsed.ActorType, claims.ActorType)
	}
	if parsed.ShopID != claims.ShopID {
		t.Errorf("ShopID: got %q, want %q", parsed.ShopID, claims.ShopID)
	}
}

func TestValidateToken_WrongSecret(t *testing.T) {
	svc1 := token.NewService("secret-a")
	svc2 := token.NewService("secret-b")

	tok, err := svc1.GenerateAccessToken(token.Claims{
		UserID:    "user-1",
		ActorType: token.ActorCustomer,
	})
	if err != nil {
		t.Fatalf("GenerateAccessToken: %v", err)
	}

	_, err = svc2.ValidateToken(tok)
	if err == nil {
		t.Fatal("expected error when validating with wrong secret")
	}
}

func TestValidateToken_Expired(t *testing.T) {
	svc := token.NewService("test-secret")

	tok, err := svc.GenerateTokenWithDuration(token.Claims{
		UserID:    "user-1",
		ActorType: token.ActorStaff,
	}, -1*time.Second)
	if err != nil {
		t.Fatalf("GenerateTokenWithDuration: %v", err)
	}

	_, err = svc.ValidateToken(tok)
	if err != token.ErrTokenExpired {
		t.Errorf("expected ErrTokenExpired, got %v", err)
	}
}

func TestValidateToken_Malformed(t *testing.T) {
	svc := token.NewService("test-secret")

	_, err := svc.ValidateToken("not.a.jwt")
	if err != token.ErrInvalidToken {
		t.Errorf("expected ErrInvalidToken, got %v", err)
	}
}

func TestGenerateRefreshToken(t *testing.T) {
	svc := token.NewService("test-secret")

	tok, err := svc.GenerateRefreshToken("user-1", "user@test.com", token.ActorCustomer)
	if err != nil {
		t.Fatalf("GenerateRefreshToken: %v", err)
	}

	claims, err := svc.ValidateToken(tok)
	if err != nil {
		t.Fatalf("ValidateToken: %v", err)
	}
	if claims.UserID != "user-1" {
		t.Errorf("UserID: got %q, want %q", claims.UserID, "user-1")
	}
	if claims.ActorType != token.ActorCustomer {
		t.Errorf("ActorType: got %q, want %q", claims.ActorType, token.ActorCustomer)
	}
}

func TestGeneratePasswordResetToken(t *testing.T) {
	svc := token.NewService("test-secret")

	tok, err := svc.GeneratePasswordResetToken("user-42")
	if err != nil {
		t.Fatalf("GeneratePasswordResetToken: %v", err)
	}
	if !strings.Contains(tok, ".") {
		t.Fatal("expected JWT format")
	}

	userID, err := svc.ValidatePasswordResetToken(tok)
	if err != nil {
		t.Fatalf("ValidatePasswordResetToken: %v", err)
	}
	if userID != "user-42" {
		t.Errorf("userID: got %q, want %q", userID, "user-42")
	}
}

func TestValidatePasswordResetToken_WrongSecret(t *testing.T) {
	svc1 := token.NewService("secret-a")
	svc2 := token.NewService("secret-b")

	tok, _ := svc1.GeneratePasswordResetToken("user-1")
	_, err := svc2.ValidatePasswordResetToken(tok)
	if err == nil {
		t.Fatal("expected error with wrong secret")
	}
}

func TestActorTypeConstants(t *testing.T) {
	if token.ActorStaff == "" || token.ActorCustomer == "" || token.ActorPlatformAdmin == "" {
		t.Fatal("actor type constants must be non-empty")
	}
}
