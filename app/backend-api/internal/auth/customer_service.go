package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"backend-api/internal/db"
	sharedanalytics "backend-api/internal/shared/analytics"
	"backend-api/internal/shared/pgutil"
	"backend-api/internal/shared/token"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrCustomerEmailTaken = errors.New("customer email already registered for this shop")
	ErrCustomerNotFound   = errors.New("customer not found")
)

func incrementCustomerBF(rdb *redis.Client, ctx context.Context, key string) {
	attempts, _ := rdb.Incr(ctx, key).Result()
	var ttl time.Duration
	switch {
	case attempts >= 5:
		ttl = time.Hour
	case attempts >= 3:
		ttl = 15 * time.Minute
	case attempts >= 2:
		ttl = 5 * time.Minute
	default:
		ttl = time.Minute
	}
	rdb.Expire(ctx, key, ttl)
}

type CustomerService struct {
	queries  *db.Queries
	tokenSvc *token.Service
	rdb      *redis.Client
}

func NewCustomerService(queries *db.Queries, tokenSvc *token.Service, rdb *redis.Client) *CustomerService {
	return &CustomerService{
		queries:  queries,
		tokenSvc: tokenSvc,
		rdb:      rdb,
	}
}

func (s *CustomerService) Register(ctx context.Context, shopID string, req CustomerRegisterRequest) (*AuthResponse, error) {
	sid, err := uuid.Parse(shopID)
	if err != nil {
		return nil, ErrShopNotFound
	}
	pgShopID := pgtype.UUID{Bytes: sid, Valid: true}

	_, err = s.queries.GetCustomerByEmail(ctx, db.GetCustomerByEmailParams{
		Email:  req.Email,
		ShopID: pgShopID,
	})
	if err == nil {
		return nil, ErrCustomerEmailTaken
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("check customer email: %w", err)
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	customer, err := s.queries.CreateCustomer(ctx, db.CreateCustomerParams{
		ShopID:       pgShopID,
		Email:        req.Email,
		PasswordHash: pgtype.Text{String: string(hashedPassword), Valid: true},
		FirstName:    pgutil.ToText(req.FirstName),
		LastName:     pgutil.ToText(req.LastName),
		Phone:        pgutil.ToText(req.Phone),
	})
	if err != nil {
		return nil, fmt.Errorf("create customer: %w", err)
	}

	if s.rdb != nil {
		custID := customer.ID.String()
		sharedanalytics.Publish(ctx, s.rdb, shopID, sharedanalytics.Event{
			Event:  "customer_signup",
			ShopID: shopID,
			Payload: map[string]any{
				"customer_id": custID,
				"source":      "storefront",
			},
		})
	}

	return s.generateCustomerAuthResponse(customer, shopID)
}

func (s *CustomerService) Login(ctx context.Context, shopID string, req CustomerLoginRequest) (*AuthResponse, error) {
	sid, err := uuid.Parse(shopID)
	if err != nil {
		return nil, ErrShopNotFound
	}
	pgShopID := pgtype.UUID{Bytes: sid, Valid: true}

	bfKey := customerBruteForceKey(shopID, req.Email)
	if cnt, _ := s.rdb.Get(ctx, bfKey).Int(); cnt >= maxLoginAttempts {
		return nil, ErrRateLimited
	}

	customer, err := s.queries.GetCustomerByEmail(ctx, db.GetCustomerByEmailParams{
		Email:  req.Email,
		ShopID: pgShopID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			incrementCustomerBF(s.rdb, ctx, bfKey)
			return nil, ErrInvalidCredentials
		}
		return nil, fmt.Errorf("get customer: %w", err)
	}

	if !customer.PasswordHash.Valid || customer.PasswordHash.String == "" {
		incrementCustomerBF(s.rdb, ctx, bfKey)
		return nil, ErrInvalidCredentials
	}
	if err := bcrypt.CompareHashAndPassword([]byte(customer.PasswordHash.String), []byte(req.Password)); err != nil {
		incrementCustomerBF(s.rdb, ctx, bfKey)
		return nil, ErrInvalidCredentials
	}

	s.rdb.Del(ctx, bfKey)
	
	if s.rdb != nil {
		sharedanalytics.Publish(ctx, s.rdb, shopID, sharedanalytics.Event{
			Event:  "customer_login",
			ShopID: shopID,
			Payload: map[string]any{
				"customer_id": customer.ID.String(),
			},
		})
	}

	return s.generateCustomerAuthResponse(customer, shopID)
}

func (s *CustomerService) RefreshToken(ctx context.Context, shopID, refreshToken string) (*AuthResponse, error) {
	claims, err := s.tokenSvc.ValidateToken(refreshToken)
	if err != nil {
		return nil, token.ErrInvalidToken
	}

	if claims.ActorType != token.ActorCustomer {
		return nil, token.ErrInvalidToken
	}

	customerID, err := uuid.Parse(claims.UserID)
	if err != nil {
		return nil, token.ErrInvalidToken
	}

	sid, err := uuid.Parse(shopID)
	if err != nil {
		return nil, ErrShopNotFound
	}

	customer, err := s.queries.GetCustomer(ctx, db.GetCustomerParams{
		ID:     customerID,
		ShopID: pgtype.UUID{Bytes: sid, Valid: true},
	})
	if err != nil {
		return nil, ErrCustomerNotFound
	}

	return s.generateCustomerAuthResponse(customer, shopID)
}

func (s *CustomerService) Me(ctx context.Context, customerID, shopID string) (*CustomerMeResponse, error) {
	cid, err := uuid.Parse(customerID)
	if err != nil {
		return nil, ErrCustomerNotFound
	}
	sid, err := uuid.Parse(shopID)
	if err != nil {
		return nil, ErrShopNotFound
	}

	customer, err := s.queries.GetCustomer(ctx, db.GetCustomerParams{
		ID:     cid,
		ShopID: pgtype.UUID{Bytes: sid, Valid: true},
	})
	if err != nil {
		return nil, ErrCustomerNotFound
	}

	return &CustomerMeResponse{
		Customer: toCustomerResponse(customer),
	}, nil
}

func (s *CustomerService) UpdateMe(ctx context.Context, customerID, shopID string, req CustomerUpdateMeRequest) (*CustomerMeResponse, error) {
	cid, err := uuid.Parse(customerID)
	if err != nil {
		return nil, ErrCustomerNotFound
	}
	sid, err := uuid.Parse(shopID)
	if err != nil {
		return nil, ErrShopNotFound
	}

	customer, err := s.queries.UpdateCustomer(ctx, db.UpdateCustomerParams{
		FirstName: pgutil.ToText(req.FirstName),
		LastName:  pgutil.ToText(req.LastName),
		Phone:     pgutil.ToText(req.Phone),
		ID:        cid,
		ShopID:    pgtype.UUID{Bytes: sid, Valid: true},
	})
	if err != nil {
		return nil, fmt.Errorf("update customer: %w", err)
	}

	return &CustomerMeResponse{
		Customer: toCustomerResponse(customer),
	}, nil
}


func (s *CustomerService) ChangePassword(ctx context.Context, customerID, shopID string, req CustomerChangePasswordRequest) error {
	cid, err := uuid.Parse(customerID)
	if err != nil {
		return ErrCustomerNotFound
	}
	sid, err := uuid.Parse(shopID)
	if err != nil {
		return ErrShopNotFound
	}

	customer, err := s.queries.GetCustomer(ctx, db.GetCustomerParams{
		ID:     cid,
		ShopID: pgtype.UUID{Bytes: sid, Valid: true},
	})
	if err != nil {
		return ErrCustomerNotFound
	}

	if !customer.PasswordHash.Valid || customer.PasswordHash.String == "" {
		return ErrInvalidCredentials
	}
	if err := bcrypt.CompareHashAndPassword([]byte(customer.PasswordHash.String), []byte(req.CurrentPassword)); err != nil {
		return ErrInvalidCredentials
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	return s.queries.UpdateCustomerPassword(ctx, db.UpdateCustomerPasswordParams{
		ID:           cid,
		ShopID:       pgtype.UUID{Bytes: sid, Valid: true},
		PasswordHash: pgtype.Text{String: string(hashed), Valid: true},
	})
}


func (s *CustomerService) generateCustomerAuthResponse(customer db.Customer, shopID string) (*AuthResponse, error) {
	claims := token.Claims{
		UserID:    customer.ID.String(),
		Email:     customer.Email,
		FullName:  pgutil.TextToString(customer.FirstName),
		ActorType: token.ActorCustomer,
		ShopID:    shopID,
	}

	accessToken, err := s.tokenSvc.GenerateAccessToken(claims)
	if err != nil {
		return nil, fmt.Errorf("generate access token: %w", err)
	}

	refreshToken, err := s.tokenSvc.GenerateRefreshToken(customer.ID.String(), customer.Email, token.ActorCustomer)
	if err != nil {
		return nil, fmt.Errorf("generate refresh token: %w", err)
	}

	custResp := toCustomerResponse(customer)
	return &AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int(token.AccessTokenDuration.Seconds()),
		ActorType:    token.ActorCustomer,
		Customer:     &custResp,
	}, nil
}

func toCustomerResponse(c db.Customer) CustomerResponse {
	resp := CustomerResponse{
		ID:        c.ID.String(),
		Email:     c.Email,
		FirstName: pgutil.TextToString(c.FirstName),
		LastName:  pgutil.TextToString(c.LastName),
		Phone:     pgutil.TextToString(c.Phone),
	}
	if c.CreatedAt.Valid {
		resp.CreatedAt = c.CreatedAt.Time.Format("2006-01-02T15:04:05Z")
	}
	return resp
}
