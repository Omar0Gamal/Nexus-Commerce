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
	RefreshTokenDuration       = 7 * 24 * time.Hour
	PasswordResetTokenDuration = 30 * time.Minute

	ActorStaff         = "staff"
	ActorCustomer      = "customer"
	ActorPlatformAdmin = "platform_admin"
)

type Claims struct {
	UserID    string `json:"user_id"`
	Email     string `json:"email"`
	FullName  string `json:"full_name,omitempty"`
	ActorType string `json:"actor_type"`
	ShopID    string `json:"shop_id,omitempty"`
	StaffID   string `json:"staff_id,omitempty"`
	Role      string `json:"role,omitempty"`
	IsOwner   bool   `json:"is_owner,omitempty"`
	jwt.RegisteredClaims
}

type PasswordResetClaims struct {
	UserID    string `json:"user_id"`
	TokenType string `json:"token_type"`
	jwt.RegisteredClaims
}

type Service struct {
	secret []byte
}

func NewService(secret string) *Service {
	return &Service{secret: []byte(secret)}
}

func (s *Service) GenerateAccessToken(claims Claims) (string, error) {
	return s.generateTokenWithDuration(claims, AccessTokenDuration)
}

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

const TwoFAChallengeDuration = 5 * time.Minute

type TwoFAChallengeClaims struct {
	UserID    string `json:"user_id"`
	Email     string `json:"email"`
	ActorType string `json:"actor_type"`
	ShopID    string `json:"shop_id,omitempty"`
	TokenType string `json:"token_type"`
	jwt.RegisteredClaims
}

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
