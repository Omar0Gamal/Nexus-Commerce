// Package token provides JWT access and refresh token generation and validation.
package token

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrInvalidToken = errors.New("invalid or expired token")
	ErrTokenExpired = errors.New("token has expired")
)

const (
	AccessTokenDuration        = 15 * time.Minute
	RefreshTokenDuration       = 7 * 24 * time.Hour // 7 days
	PasswordResetTokenDuration = 30 * time.Minute

	// Actor types for JWT claims
	ActorStaff         = "staff"
	ActorCustomer      = "customer"
	ActorPlatformAdmin = "platform_admin"
)

// Claims holds the JWT payload for both access and refresh tokens.
type Claims struct {
	UserID    string `json:"user_id"`
	Email     string `json:"email"`
	FullName  string `json:"full_name,omitempty"`
	ActorType string `json:"actor_type"` // "staff", "customer", "platform_admin"
	// ShopID + Role are populated when the user is acting within a shop context.
	ShopID  string `json:"shop_id,omitempty"`
	StaffID string `json:"staff_id,omitempty"`
	Role    string `json:"role,omitempty"`
	IsOwner bool   `json:"is_owner,omitempty"`
	jwt.RegisteredClaims
}

// PasswordResetClaims holds the minimal JWT payload for password-reset tokens.
type PasswordResetClaims struct {
	UserID    string `json:"user_id"`
	TokenType string `json:"token_type"` // always "password_reset"
	jwt.RegisteredClaims
}

// Service handles JWT operations.
type Service struct {
	secret []byte
}

func NewService(secret string) *Service {
	return &Service{secret: []byte(secret)}
}

// GenerateAccessToken creates a short-lived JWT for API authentication.
func (s *Service) GenerateAccessToken(claims Claims) (string, error) {
	return s.generateTokenWithDuration(claims, AccessTokenDuration)
}

// GenerateTokenWithDuration creates a JWT with a custom expiry duration.
// Used for impersonation tokens issued by platform admins.
func (s *Service) GenerateTokenWithDuration(claims Claims, duration time.Duration) (string, error) {
	return s.generateTokenWithDuration(claims, duration)
}

func (s *Service) generateTokenWithDuration(claims Claims, duration time.Duration) (string, error) {
	now := time.Now()
	claims.RegisteredClaims = jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(now.Add(duration)),
		IssuedAt:  jwt.NewNumericDate(now),
		Issuer:    "nexus-commerce",
		Subject:   claims.UserID,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.secret)
}

// GenerateRefreshToken creates a long-lived JWT for token rotation.
func (s *Service) GenerateRefreshToken(userID, email, actorType string) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID:    userID,
		Email:     email,
		ActorType: actorType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(RefreshTokenDuration)),
			IssuedAt:  jwt.NewNumericDate(now),
			Issuer:    "nexus-commerce",
			Subject:   userID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.secret)
}

// ValidateToken parses and validates a JWT string, returning its claims.
func (s *Service) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return s.secret, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

// GeneratePasswordResetToken creates a short-lived JWT for password reset.
func (s *Service) GeneratePasswordResetToken(userID string) (string, error) {
	now := time.Now()
	claims := PasswordResetClaims{
		UserID:    userID,
		TokenType: "password_reset",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(PasswordResetTokenDuration)),
			IssuedAt:  jwt.NewNumericDate(now),
			Issuer:    "nexus-commerce",
			Subject:   userID,
		},
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return t.SignedString(s.secret)
}

// ValidatePasswordResetToken parses and validates a password-reset JWT.
func (s *Service) ValidatePasswordResetToken(tokenString string) (userID string, err error) {
	t, err := jwt.ParseWithClaims(tokenString, &PasswordResetClaims{}, func(tok *jwt.Token) (interface{}, error) {
		if _, ok := tok.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return s.secret, nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return "", ErrTokenExpired
		}
		return "", ErrInvalidToken
	}
	claims, ok := t.Claims.(*PasswordResetClaims)
	if !ok || !t.Valid || claims.TokenType != "password_reset" {
		return "", ErrInvalidToken
	}
	return claims.UserID, nil
}

// TwoFAChallengeDuration is the lifetime of the short-lived 2FA challenge token.
const TwoFAChallengeDuration = 5 * time.Minute

// TwoFAChallengeClaims is the JWT payload for a 2FA challenge token.
type TwoFAChallengeClaims struct {
	UserID    string `json:"user_id"`
	Email     string `json:"email"`
	ActorType string `json:"actor_type"` // "staff"
	ShopID    string `json:"shop_id,omitempty"`
	TokenType string `json:"token_type"` // always "2fa_challenge"
	jwt.RegisteredClaims
}

// Generate2FAChallengeToken creates a short-lived JWT that can be exchanged for
// full auth tokens once the caller proves their TOTP code.
func (s *Service) Generate2FAChallengeToken(userID, email, actorType, shopID string) (string, error) {
	now := time.Now()
	claims := TwoFAChallengeClaims{
		UserID:    userID,
		Email:     email,
		ActorType: actorType,
		ShopID:    shopID,
		TokenType: "2fa_challenge",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(TwoFAChallengeDuration)),
			IssuedAt:  jwt.NewNumericDate(now),
			Issuer:    "nexus-commerce",
			Subject:   userID,
		},
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return t.SignedString(s.secret)
}

// Validate2FAChallengeToken verifies a 2FA challenge JWT.
func (s *Service) Validate2FAChallengeToken(tokenString string) (*TwoFAChallengeClaims, error) {
	t, err := jwt.ParseWithClaims(tokenString, &TwoFAChallengeClaims{}, func(tok *jwt.Token) (interface{}, error) {
		if _, ok := tok.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return s.secret, nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		return nil, ErrInvalidToken
	}
	claims, ok := t.Claims.(*TwoFAChallengeClaims)
	if !ok || !t.Valid || claims.TokenType != "2fa_challenge" {
		return nil, ErrInvalidToken
	}
	return claims, nil
}
