package auth

import (
	"context"
	"errors"
	"fmt"

	"backend-api/internal/db"
	"backend-api/internal/shared/pgutil"
	"backend-api/internal/shared/token"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrAdminNotFound = errors.New("admin not found")
)

// AdminService handles platform admin authentication.
type AdminService struct {
	queries  *db.Queries
	tokenSvc *token.Service
}

func NewAdminService(queries *db.Queries, tokenSvc *token.Service) *AdminService {
	return &AdminService{
		queries:  queries,
		tokenSvc: tokenSvc,
	}
}


// Login authenticates a platform admin by email/password.
func (s *AdminService) Login(ctx context.Context, req AdminLoginRequest) (*AuthResponse, error) {
	admin, err := s.queries.GetPlatformAdminByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrInvalidCredentials
		}
		return nil, fmt.Errorf("get admin: %w", err)
	}

	// Verify password (platform_admins.password_hash is NOT NULL)
	if err := bcrypt.CompareHashAndPassword([]byte(admin.PasswordHash), []byte(req.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	return s.generateAdminAuthResponse(admin)
}


// RefreshToken validates an admin's refresh token and issues new tokens.
func (s *AdminService) RefreshToken(ctx context.Context, refreshToken string) (*AuthResponse, error) {
	claims, err := s.tokenSvc.ValidateToken(refreshToken)
	if err != nil {
		return nil, token.ErrInvalidToken
	}

	if claims.ActorType != token.ActorPlatformAdmin {
		return nil, token.ErrInvalidToken
	}

	adminID, err := uuid.Parse(claims.UserID)
	if err != nil {
		return nil, token.ErrInvalidToken
	}

	admin, err := s.queries.GetPlatformAdmin(ctx, adminID)
	if err != nil {
		return nil, ErrAdminNotFound
	}

	return s.generateAdminAuthResponse(admin)
}


// Me returns the current platform admin profile.
func (s *AdminService) Me(ctx context.Context, adminID string) (*AdminMeResponse, error) {
	aid, err := uuid.Parse(adminID)
	if err != nil {
		return nil, ErrAdminNotFound
	}

	admin, err := s.queries.GetPlatformAdmin(ctx, aid)
	if err != nil {
		return nil, ErrAdminNotFound
	}

	return &AdminMeResponse{
		Admin: toAdminResponse(admin),
	}, nil
}


func (s *AdminService) generateAdminAuthResponse(admin db.PlatformAdmin) (*AuthResponse, error) {
	claims := token.Claims{
		UserID:    admin.ID.String(),
		Email:     admin.Email,
		FullName:  pgutil.TextToString(admin.FullName),
		ActorType: token.ActorPlatformAdmin,
	}

	accessToken, err := s.tokenSvc.GenerateAccessToken(claims)
	if err != nil {
		return nil, fmt.Errorf("generate access token: %w", err)
	}

	refreshToken, err := s.tokenSvc.GenerateRefreshToken(admin.ID.String(), admin.Email, token.ActorPlatformAdmin)
	if err != nil {
		return nil, fmt.Errorf("generate refresh token: %w", err)
	}

	return &AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int(token.AccessTokenDuration.Seconds()),
		ActorType:    token.ActorPlatformAdmin,
		User: UserResponse{
			ID:       admin.ID.String(),
			Email:    admin.Email,
			FullName: pgutil.TextToString(admin.FullName),
		},
	}, nil
}

func toAdminResponse(a db.PlatformAdmin) AdminResponse {
	return AdminResponse{
		ID:       a.ID.String(),
		Email:    a.Email,
		FullName: pgutil.TextToString(a.FullName),
	}
}
