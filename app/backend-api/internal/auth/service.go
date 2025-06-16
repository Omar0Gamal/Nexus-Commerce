package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"

	"backend-api/internal/db"
	"backend-api/internal/email"
	"backend-api/internal/shared/pgutil"
	"backend-api/internal/shared/token"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrEmailTaken         = errors.New("email already registered")
	ErrSubdomainTaken     = errors.New("subdomain already taken")
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrAccountInactive    = errors.New("account is not active")
	ErrNotStaff           = errors.New("user is not staff of this shop")
	ErrPlanNotFound       = errors.New("plan not found")
	ErrShopNotFound       = errors.New("shop not found")
	ErrUserNotFound       = errors.New("user not found")
	ErrAlreadyMember      = errors.New("user is already a member of this shop")
	ErrStaffLimit         = errors.New("staff seat limit reached for current plan")
)

// StaffLimitChecker is the subset of billing.Service needed by auth.
// Using an interface breaks the import cycle auth ↔ billing.
type StaffLimitChecker interface {
	CheckStaffLimit(ctx context.Context, shopID string) error
}

// Service handles authentication business logic.
type Service struct {
	store     Store
	pool      *pgxpool.Pool
	tokenSvc  *token.Service
	billing   StaffLimitChecker
	mailer    *email.Mailer
	rdb       *redis.Client
	logger    *zap.Logger
	loginFail sync.Map
}

func NewService(queries *db.Queries, pool *pgxpool.Pool, tokenSvc *token.Service, billing StaffLimitChecker, mailer *email.Mailer, rdb *redis.Client, logger *zap.Logger) *Service {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &Service{
		store:    NewDBStore(queries),
		pool:     pool,
		tokenSvc: tokenSvc,
		billing:  billing,
		mailer:   mailer,
		rdb:      rdb,
		logger:   logger,
	}
}

// Queries returns the underlying *db.Queries (e.g. for sharedaudit).
func (s *Service) Queries() *db.Queries {
	return s.Queries()
}

// Register creates a new user + shop + owner role + staff entry in a single transaction.
func (s *Service) Register(ctx context.Context, req RegisterRequest) (*AuthResponse, error) {
	if req.Currency == "" {
		req.Currency = "EGP"
	}
	if req.Timezone == "" {
		req.Timezone = "Africa/Cairo"
	}
	if req.PlanName == "" {
		req.PlanName = "basic"
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	_, err = s.Queries().GetUserByEmail(ctx, req.Email)
	if err == nil {
		return nil, ErrEmailTaken
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("check email: %w", err)
	}

	_, err = s.Queries().GetShopBySubdomain(ctx, req.Subdomain)
	if err == nil {
		return nil, ErrSubdomainTaken
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("check subdomain: %w", err)
	}

	plan, err := s.Queries().GetPlanByName(ctx, req.PlanName)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrPlanNotFound
		}
		return nil, fmt.Errorf("get plan: %w", err)
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	qtx := s.Queries().WithTx(tx)

	// 1. Create user
	user, err := qtx.CreateUser(ctx, db.CreateUserParams{
		Email:        req.Email,
		PasswordHash: pgutil.ToText(string(hashedPassword)),
		FullName:     pgutil.ToText(req.FullName),
		Phone:        pgutil.ToText(req.Phone),
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, ErrEmailTaken
		}
		return nil, fmt.Errorf("create user: %w", err)
	}

	// 2. Create shop
	shop, err := qtx.CreateShop(ctx, db.CreateShopParams{
		PlanID:       plan.ID,
		OwnerUserID:  user.ID,
		Name:         req.ShopName,
		Subdomain:    req.Subdomain,
		CustomDomain: pgtype.Text{},
		Status:       db.NullShopStatus{ShopStatus: db.ShopStatusActive, Valid: true},
		Currency:     pgutil.ToText(req.Currency),
		Timezone:     pgutil.ToText(req.Timezone),
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			if strings.Contains(pgErr.ConstraintName, "subdomain") {
				return nil, ErrSubdomainTaken
			}
			return nil, ErrEmailTaken
		}
		return nil, fmt.Errorf("create shop: %w", err)
	}

	// 3. Create "owner" system role
	ownerRole, err := qtx.CreateShopRole(ctx, db.CreateShopRoleParams{
		ShopID:       pgutil.ToUUID(shop.ID),
		ParentRoleID: pgtype.UUID{},
		Name:         "owner",
		Permissions:  []byte(`["*"]`),
		IsSystemRole: pgtype.Bool{Bool: true, Valid: true},
	})
	if err != nil {
		return nil, fmt.Errorf("create owner role: %w", err)
	}

	// 3b. Create additional system roles (manager / staff / viewer).
	systemRoles := []struct {
		name  string
		perms string
	}{
		{"manager", `{"products.read":true,"products.write":true,"orders.read":true,"orders.write":true,"customers.read":true,"customers.write":true,"analytics.read":true,"settings.read":true,"settings.write":true,"marketing.read":true,"marketing.write":true,"support.read":true,"support.write":true}`},
		{"staff", `{"products.read":true,"products.write":true,"orders.read":true,"orders.write":true,"customers.read":true,"support.read":true,"support.write":true}`},
		{"viewer", `{"products.read":true,"orders.read":true,"customers.read":true,"analytics.read":true,"support.read":true}`},
	}
	for _, sr := range systemRoles {
		_, err = qtx.CreateShopRole(ctx, db.CreateShopRoleParams{
			ShopID:       pgutil.ToUUID(shop.ID),
			ParentRoleID: pgtype.UUID{},
			Name:         sr.name,
			Permissions:  []byte(sr.perms),
			IsSystemRole: pgtype.Bool{Bool: true, Valid: true},
		})
		if err != nil {
			return nil, fmt.Errorf("create %s role: %w", sr.name, err)
		}
	}

	// 4. Link user as owner staff
	_, err = qtx.CreateShopStaff(ctx, db.CreateShopStaffParams{
		ShopID:  pgutil.ToUUID(shop.ID),
		UserID:  pgutil.ToUUID(user.ID),
		RoleID:  pgutil.ToUUID(ownerRole.ID),
		IsOwner: pgtype.Bool{Bool: true, Valid: true},
	})
	if err != nil {
		return nil, fmt.Errorf("create shop staff: %w", err)
	}

	// 5. Seed default payment methods (disabled until the merchant configures them)
	// Paymob: enabled once merchant sets their sub_merchant_code via dashboard settings
	_, _ = qtx.CreateShopPaymentMethod(ctx, db.CreateShopPaymentMethodParams{
		ShopID:               pgutil.ToUUID(shop.ID),
		Provider:             db.PaymentProviderPaymob,
		IsEnabled:            pgtype.Bool{Bool: false, Valid: true},
		EncryptedCredentials: `{}`,
	})
	// Cash on Delivery: enabled by default
	_, _ = qtx.CreateShopPaymentMethod(ctx, db.CreateShopPaymentMethodParams{
		ShopID:               pgutil.ToUUID(shop.ID),
		Provider:             db.PaymentProviderCashOnDelivery,
		IsEnabled:            pgtype.Bool{Bool: true, Valid: true},
		EncryptedCredentials: `{}`,
	})

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	return s.generateAuthResponse(user, &shop, "owner", true)
}

// Login authenticates a user by email/password and returns JWT tokens.
// If 2FA is enabled on the account, it returns a TwoFAChallengeResponse instead.
func (s *Service) Login(ctx context.Context, req LoginRequest) (any, error) {
	// Brute-force protection
	bfKey := staffBruteForceKey(req.Email)
	if err := s.checkBruteForce(ctx, bfKey); err != nil {
		return nil, ErrRateLimited
	}

	user, err := s.Queries().GetUserByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			s.recordFailedLogin(ctx, bfKey)
			return nil, ErrInvalidCredentials
		}
		return nil, fmt.Errorf("get user: %w", err)
	}

	if !user.Status.Valid || user.Status.UserStatus != db.UserStatusActive {
		return nil, ErrAccountInactive
	}

	if !user.PasswordHash.Valid {
		s.recordFailedLogin(ctx, bfKey)
		return nil, ErrInvalidCredentials
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash.String), []byte(req.Password)); err != nil {
		s.recordFailedLogin(ctx, bfKey)
		return nil, ErrInvalidCredentials
	}

	// Successful auth — clear brute-force counter
	s.clearFailedLogin(ctx, bfKey)

	secrets, sErr := s.Queries().GetUserSecrets(ctx, user.ID)
	if sErr == nil && secrets.Is2faEnabled.Bool {
		challengeToken, tErr := s.tokenSvc.Generate2FAChallengeToken(
			user.ID.String(), user.Email, "staff", req.ShopID,
		)
		if tErr != nil {
			return nil, fmt.Errorf("generate 2fa challenge token: %w", tErr)
		}
		return &TwoFAChallengeResponse{Requires2FA: true, LoginToken: challengeToken}, nil
	}

	if req.ShopID != "" {
		shopID, err := uuid.Parse(req.ShopID)
		if err != nil {
			return nil, ErrShopNotFound
		}

		staff, roleName, err := s.getStaffContext(ctx, shopID, user.ID)
		if err != nil {
			return nil, err
		}

		shop, err := s.Queries().GetShop(ctx, shopID)
		if err != nil {
			return nil, ErrShopNotFound
		}

		return s.generateAuthResponse(user, &shop, roleName, staff.IsOwner.Bool)
	}

	// No shop context — return token without shop claim.
	// User can call /auth/select-shop later.
	return s.generateAuthResponse(user, nil, "", false)
}

// RefreshToken validates a refresh token and issues a new access + refresh pair.
func (s *Service) RefreshToken(ctx context.Context, req RefreshRequest) (*AuthResponse, error) {
	claims, err := s.tokenSvc.ValidateToken(req.RefreshToken)
	if err != nil {
		return nil, token.ErrInvalidToken
	}

	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		return nil, token.ErrInvalidToken
	}

	user, err := s.Queries().GetUser(ctx, userID)
	if err != nil {
		return nil, ErrUserNotFound
	}

	if !user.Status.Valid || user.Status.UserStatus != db.UserStatusActive {
		return nil, ErrAccountInactive
	}

	// Optional: switch/maintain shop context
	if req.ShopID != "" {
		shopID, err := uuid.Parse(req.ShopID)
		if err != nil {
			return nil, ErrShopNotFound
		}

		staff, roleName, err := s.getStaffContext(ctx, shopID, user.ID)
		if err != nil {
			return nil, err
		}

		shop, err := s.Queries().GetShop(ctx, shopID)
		if err != nil {
			return nil, ErrShopNotFound
		}

		return s.generateAuthResponse(user, &shop, roleName, staff.IsOwner.Bool)
	}

	return s.generateAuthResponse(user, nil, "", false)
}

// SelectShop generates new tokens scoped to a specific shop the user belongs to.
func (s *Service) SelectShop(ctx context.Context, userID string, req SelectShopRequest) (*AuthResponse, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, ErrUserNotFound
	}

	user, err := s.Queries().GetUser(ctx, uid)
	if err != nil {
		return nil, ErrUserNotFound
	}

	shopID, err := uuid.Parse(req.ShopID)
	if err != nil {
		return nil, ErrShopNotFound
	}

	staff, roleName, err := s.getStaffContext(ctx, shopID, uid)
	if err != nil {
		return nil, err
	}

	shop, err := s.Queries().GetShop(ctx, shopID)
	if err != nil {
		return nil, ErrShopNotFound
	}

	return s.generateAuthResponse(user, &shop, roleName, staff.IsOwner.Bool)
}

// GetShopSettings returns the current shop's name, subdomain, currency, plan, SEO and notification prefs.
func (s *Service) GetShopSettings(ctx context.Context, shopID string) (*ShopSettingsResponse, error) {
	sid, err := uuid.Parse(shopID)
	if err != nil {
		return nil, ErrShopNotFound
	}
	shop, err := s.Queries().GetShop(ctx, sid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrShopNotFound
		}
		return nil, fmt.Errorf("get shop: %w", err)
	}
	resp := &ShopSettingsResponse{
		ID:        shop.ID.String(),
		Name:      shop.Name,
		Subdomain: shop.Subdomain,
		Currency:  pgutil.TextToString(shop.Currency),
	}
	if planName, err := s.Queries().GetShopPlanName(ctx, sid); err == nil {
		resp.PlanName = planName
	}
	resp.SeoTitle = shop.SeoTitle.String
	resp.SeoDescription = shop.SeoDescription.String
	resp.FaviconURL = shop.FaviconUrl.String
	resp.CustomDomain = shop.CustomDomain.String
	if prefsJSON, err := s.Queries().GetShopNotificationPrefs(ctx, sid); err == nil {
		var prefs map[string]bool
		if jsonErr := json.Unmarshal(prefsJSON, &prefs); jsonErr == nil {
			resp.NotificationPrefs = prefs
		}
	}
	return resp, nil
}

// UpdateShopSettings updates the name and/or currency of a shop (staff, tenant-scoped).
func (s *Service) UpdateShopSettings(ctx context.Context, shopID string, req UpdateShopSettingsRequest) (*ShopSettingsResponse, error) {
	sid, err := uuid.Parse(shopID)
	if err != nil {
		return nil, ErrShopNotFound
	}

	params := db.UpdateShopParams{ID: sid}
	if req.Name != "" {
		params.Name = pgtype.Text{String: req.Name, Valid: true}
	}
	if req.Currency != "" {
		params.Currency = pgtype.Text{String: req.Currency, Valid: true}
	}
	if req.CustomDomain != "" {
		// Normalise: strip protocol + trailing slash, lowercase
		cd := strings.ToLower(strings.TrimSpace(req.CustomDomain))
		cd = strings.TrimPrefix(cd, "https://")
		cd = strings.TrimPrefix(cd, "http://")
		cd = strings.TrimRight(cd, "/")
		params.CustomDomain = pgtype.Text{String: cd, Valid: true}
	}

	shop, err := s.Queries().UpdateShop(ctx, params)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrShopNotFound
		}
		return nil, fmt.Errorf("update shop: %w", err)
	}

	if req.SeoTitle != "" || req.SeoDescription != "" || req.FaviconURL != "" {
		_ = s.Queries().UpdateShopSEO(ctx, db.UpdateShopSEOParams{
			ID:             sid,
			SeoTitle:       pgtype.Text{String: req.SeoTitle, Valid: req.SeoTitle != ""},
			SeoDescription: pgtype.Text{String: req.SeoDescription, Valid: req.SeoDescription != ""},
			FaviconUrl:     pgtype.Text{String: req.FaviconURL, Valid: req.FaviconURL != ""},
		})
	}

	resp := &ShopSettingsResponse{
		ID:        shop.ID.String(),
		Name:      shop.Name,
		Subdomain: shop.Subdomain,
		Currency:  pgutil.TextToString(shop.Currency),
	}
	if planName, err := s.Queries().GetShopPlanName(ctx, sid); err == nil {
		resp.PlanName = planName
	}
	// Use request values directly — no redundant DB round-trip needed.
	resp.SeoTitle = req.SeoTitle
	resp.SeoDescription = req.SeoDescription
	resp.FaviconURL = req.FaviconURL
	resp.CustomDomain = shop.CustomDomain.String

	// Publish gateway cache invalidation so the new custom domain is
	// immediately resolvable (and the old one stops matching).
	if s.rdb != nil && req.CustomDomain != "" {
		if payload, marshalErr := json.Marshal(map[string]string{"target": "shop", "key": shopID}); marshalErr == nil {
			_ = s.rdb.Publish(context.Background(), "events:gateway:cache_purge", string(payload))
		}
	}

	if req.NotificationPrefs != nil {
		if prefsJSON, marshalErr := json.Marshal(req.NotificationPrefs); marshalErr == nil {
			_ = s.Queries().UpdateShopNotificationPrefs(ctx, db.UpdateShopNotificationPrefsParams{
				ID:                sid,
				NotificationPrefs: prefsJSON,
			})
		}
		resp.NotificationPrefs = req.NotificationPrefs
	} else {
		// Return current prefs even when not updating them.
		if prefsJSON, err := s.Queries().GetShopNotificationPrefs(ctx, sid); err == nil {
			var prefs map[string]bool
			if jsonErr := json.Unmarshal(prefsJSON, &prefs); jsonErr == nil {
				resp.NotificationPrefs = prefs
			}
		}
	}

	return resp, nil
}

// ListStaff returns all staff members for a shop.
func (s *Service) ListStaff(ctx context.Context, shopID string) ([]StaffMemberResponse, error) {
	sid, err := uuid.Parse(shopID)
	if err != nil {
		return nil, ErrShopNotFound
	}
	members, err := s.Queries().ListStaffWithUsers(ctx, pgutil.ToUUID(sid))
	if err != nil {
		return nil, fmt.Errorf("list staff: %w", err)
	}
	result := make([]StaffMemberResponse, 0, len(members))
	for _, m := range members {
		result = append(result, StaffMemberResponse{
			ID:       m.ID.String(),
			UserID:   uuid.UUID(m.UserID.Bytes).String(),
			Email:    m.Email,
			FullName: pgutil.TextToString(m.FullName),
			Phone:    pgutil.TextToString(m.Phone),
			IsOwner:  m.IsOwner,
			JoinedAt: m.JoinedAt.Time.Format("2006-01-02T15:04:05Z07:00"),
		})
	}
	return result, nil
}

// AddStaff adds an existing user as a staff member of the shop by email.
func (s *Service) AddStaff(ctx context.Context, shopID string, req AddStaffRequest) (*StaffMemberResponse, error) {
	// Enforce plan staff seat limit if billing service is available.
	if s.billing != nil {
		if err := s.billing.CheckStaffLimit(ctx, shopID); err != nil {
			return nil, ErrStaffLimit
		}
	}

	sid, err := uuid.Parse(shopID)
	if err != nil {
		return nil, ErrShopNotFound
	}
	user, err := s.Queries().GetUserByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("get user: %w", err)
	}
	roleID, err := s.Queries().GetOrCreateStaffRole(ctx, pgutil.ToUUID(sid))
	if err != nil {
		return nil, fmt.Errorf("get staff role: %w", err)
	}
	staff, err := s.Queries().CreateShopStaff(ctx, db.CreateShopStaffParams{
		ShopID:  pgutil.ToUUID(sid),
		UserID:  pgutil.ToUUID(user.ID),
		RoleID:  pgutil.ToUUID(roleID),
		IsOwner: pgtype.Bool{Bool: false, Valid: true},
	})
	if err != nil {
		var pgErr interface{ SQLState() string }
		if errors.As(err, &pgErr) && pgErr.SQLState() == "23505" {
			return nil, ErrAlreadyMember
		}
		return nil, fmt.Errorf("add staff: %w", err)
	}
	resp := &StaffMemberResponse{
		ID:       staff.ID.String(),
		UserID:   user.ID.String(),
		Email:    user.Email,
		FullName: pgutil.TextToString(user.FullName),
		Phone:    pgutil.TextToString(user.Phone),
		IsOwner:  false,
		JoinedAt: staff.JoinedAt.Time.Format("2006-01-02T15:04:05Z07:00"),
	}
	return resp, nil
}

// RemoveStaff removes a staff member from a shop (owners cannot be removed).
func (s *Service) RemoveStaff(ctx context.Context, shopID, staffID string) error {
	sid, err := uuid.Parse(shopID)
	if err != nil {
		return ErrShopNotFound
	}
	smid, err := uuid.Parse(staffID)
	if err != nil {
		return fmt.Errorf("invalid staff id")
	}
	return s.Queries().DeleteShopStaffProtected(ctx, db.DeleteShopStaffProtectedParams{
		ID:     smid,
		ShopID: pgutil.ToUUID(sid),
	})
}

// GetStaffProfile returns the profile of the current (authenticated) staff member.
func (s *Service) GetStaffProfile(ctx context.Context, userID string) (*UserResponse, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, ErrUserNotFound
	}
	user, err := s.Queries().GetUser(ctx, uid)
	if err != nil {
		return nil, ErrUserNotFound
	}
	resp := toUserResponse(user)
	return &resp, nil
}

// UpdateStaffProfile updates full_name and/or phone for the authenticated staff member.
func (s *Service) UpdateStaffProfile(ctx context.Context, userID string, req StaffUpdateProfileRequest) (*UserResponse, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, ErrUserNotFound
	}
	var fnPg, phPg pgtype.Text
	if req.FullName != "" {
		fnPg = pgtype.Text{String: req.FullName, Valid: true}
	}
	if req.Phone != "" {
		phPg = pgtype.Text{String: req.Phone, Valid: true}
	}
	user, err := s.Queries().UpdateUser(ctx, db.UpdateUserParams{
		ID:       uid,
		FullName: fnPg,
		Phone:    phPg,
	})
	if err != nil {
		return nil, fmt.Errorf("update profile: %w", err)
	}
	resp := toUserResponse(user)
	return &resp, nil
}

// StaffChangePassword verifies the current password and sets a new one.
func (s *Service) StaffChangePassword(ctx context.Context, userID string, req StaffChangePasswordRequest) error {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return ErrUserNotFound
	}
	user, err := s.Queries().GetUser(ctx, uid)
	if err != nil {
		return ErrUserNotFound
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash.String), []byte(req.CurrentPassword)); err != nil {
		return ErrInvalidCredentials
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}
	return s.Queries().UpdateUserPassword(ctx, db.UpdateUserPasswordParams{
		ID:           uid,
		PasswordHash: pgtype.Text{String: string(hash), Valid: true},
	})
}

// mapAuthErr translates pgx duplicate-key errors to ErrEmailTaken.
func mapAuthErr(err error) error {
	if err == nil {
		return nil
	}
	var pgErr interface{ SQLState() string }
	if errors.As(err, &pgErr) && pgErr.SQLState() == "23505" {
		return ErrEmailTaken
	}
	return err
}

// Me returns the current user info and all shops they belong to.
func (s *Service) Me(ctx context.Context, userID string) (*MeResponse, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, ErrUserNotFound
	}

	user, err := s.Queries().GetUser(ctx, uid)
	if err != nil {
		return nil, ErrUserNotFound
	}

	// Single query — join shop_staff + shops + shop_roles; no N+1.
	rows, err := s.Queries().ListShopsForUser(ctx, pgutil.ToUUID(uid))
	if err != nil {
		return nil, fmt.Errorf("list shops: %w", err)
	}

	shops := make([]ShopBrief, 0, len(rows))
	for _, row := range rows {
		roleName := "staff"
		if row.RoleName.Valid {
			roleName = row.RoleName.String
		}
		shops = append(shops, ShopBrief{
			ID:        row.ShopID.String(),
			Name:      row.ShopName,
			Subdomain: row.Subdomain,
			Role:      roleName,
			IsOwner:   row.IsOwner,
		})
	}

	return &MeResponse{
		User:  toUserResponse(user),
		Shops: shops,
	}, nil
}

// getStaffContext retrieves a user's staff entry + role name for a given shop.
func (s *Service) getStaffContext(ctx context.Context, shopID, userID uuid.UUID) (db.ShopStaff, string, error) {
	staff, err := s.Queries().GetShopStaffByUserAndShop(ctx, db.GetShopStaffByUserAndShopParams{
		ShopID: pgutil.ToUUID(shopID),
		UserID: pgutil.ToUUID(userID),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.ShopStaff{}, "", ErrNotStaff
		}
		return db.ShopStaff{}, "", fmt.Errorf("get staff: %w", err)
	}

	roleName := "staff"
	if staff.RoleID.Valid {
		roleID, err := uuid.FromBytes(staff.RoleID.Bytes[:])
		if err == nil {
			role, err := s.Queries().GetShopRole(ctx, roleID)
			if err == nil {
				roleName = role.Name
			}
		}
	}

	return staff, roleName, nil
}

// generateAuthResponse builds the token pair + response DTOs.
func (s *Service) generateAuthResponse(user db.User, shop *db.Shop, roleName string, isOwner bool) (*AuthResponse, error) {
	claims := token.Claims{
		UserID:    user.ID.String(),
		Email:     user.Email,
		FullName:  pgutil.TextToString(user.FullName),
		ActorType: token.ActorStaff,
	}

	if shop != nil {
		claims.ShopID = shop.ID.String()
		claims.Role = roleName
		claims.IsOwner = isOwner
	}

	accessToken, err := s.tokenSvc.GenerateAccessToken(claims)
	if err != nil {
		return nil, fmt.Errorf("generate access token: %w", err)
	}

	refreshToken, err := s.tokenSvc.GenerateRefreshToken(user.ID.String(), user.Email, token.ActorStaff)
	if err != nil {
		return nil, fmt.Errorf("generate refresh token: %w", err)
	}

	resp := &AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int(token.AccessTokenDuration.Seconds()),
		ActorType:    token.ActorStaff,
		User:         toUserResponse(user),
	}

	if shop != nil {
		resp.Shop = &ShopBrief{
			ID:        shop.ID.String(),
			Name:      shop.Name,
			Subdomain: shop.Subdomain,
			Role:      roleName,
			IsOwner:   isOwner,
		}
	}

	return resp, nil
}

func toUserResponse(u db.User) UserResponse {
	return UserResponse{
		ID:              u.ID.String(),
		Email:           u.Email,
		FullName:        pgutil.TextToString(u.FullName),
		Phone:           pgutil.TextToString(u.Phone),
		IsEmailVerified: u.IsEmailVerified.Bool,
	}
}

// ErrInvalidResetToken is returned when a password-reset token is invalid or expired.
var ErrInvalidResetToken = errors.New("invalid or expired password reset token")

// ForgotPasswordRequest holds the email for a password-reset request.
type ForgotPasswordRequest struct {
	Email       string `json:"email" binding:"required,email"`
	FrontendURL string `json:"-"` // injected by handler, not from client
}

// ForgotPassword generates a signed reset token and sends the email.
// It always returns nil even if the email is not found (to prevent user enumeration).
func (s *Service) ForgotPassword(ctx context.Context, req ForgotPasswordRequest) error {
	user, err := s.Queries().GetUserByEmail(ctx, req.Email)
	if err != nil {
		// User not found — silent success to prevent enumeration.
		return nil
	}

	resetToken, err := s.tokenSvc.GeneratePasswordResetToken(user.ID.String())
	if err != nil {
		return fmt.Errorf("generate reset token: %w", err)
	}

	if s.mailer != nil {
		frontendURL := req.FrontendURL
		if frontendURL == "" {
			frontendURL = "https://app.nexuscommerce.io"
		}
		resetLink := fmt.Sprintf("%s/reset-password?token=%s", frontendURL, resetToken)
		fullName := pgutil.TextToString(user.FullName)
		if fullName == "" {
			fullName = user.Email
		}
		data := email.PasswordResetData{
			FullName:  fullName,
			ResetLink: resetLink,
			ExpiresIn: "30 minutes",
		}
		go func() {
			if err := s.mailer.SendPasswordReset(user.Email, data); err != nil {
				s.logger.Error("failed to send password reset email",
					zap.String("email", user.Email),
					zap.Error(err),
				)
			}
		}()
	}

	return nil
}

// ResetPasswordRequest holds the token and new password.
type ResetPasswordRequest struct {
	Token       string `json:"token" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8"`
}

// ResetPassword validates the reset token and updates the user's password.
func (s *Service) ResetPassword(ctx context.Context, req ResetPasswordRequest) error {
	userID, err := s.tokenSvc.ValidatePasswordResetToken(req.Token)
	if err != nil {
		return ErrInvalidResetToken
	}

	uid, err := uuid.Parse(userID)
	if err != nil {
		return ErrInvalidResetToken
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	return s.Queries().UpdateUserPassword(ctx, db.UpdateUserPasswordParams{
		ID:           uid,
		PasswordHash: pgtype.Text{String: string(hash), Valid: true},
	})
}
