package staff

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"backend-api/internal/db"
	"backend-api/internal/email"
	"backend-api/internal/shared/pgutil"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

var (
	ErrStaffNotFound         = errors.New("staff member not found")
	ErrRoleNotFound          = errors.New("role not found")
	ErrAlreadyStaff          = errors.New("user is already a staff member of this shop")
	ErrCannotChangeOwnerRole = errors.New("cannot change owner role")
	ErrCannotRemoveOwner     = errors.New("cannot remove shop owner")
	ErrInvalidUUID           = errors.New("invalid UUID")
	ErrStaffLimit            = errors.New("staff seat limit reached for your plan")
	ErrInvalidEmail          = errors.New("invalid email address")
)

type BillingChecker interface {
	CheckStaffLimit(ctx context.Context, shopID string) error
}

type InviteStaffRequest struct {
	Email    string `json:"email" binding:"required,email"`
	RoleID   string `json:"role_id" binding:"required"`
	FullName string `json:"full_name"`
	Phone    string `json:"phone"`
}

type UpdateStaffRoleRequest struct {
	RoleID string `json:"role_id" binding:"required"`
}

type StaffMemberResponse struct {
	ID          string `json:"id"`
	ShopID      string `json:"shop_id"`
	UserID      string `json:"user_id"`
	RoleID      string `json:"role_id"`
	RoleName    string `json:"role_name"`
	IsOwner     bool   `json:"is_owner"`
	JoinedAt    string `json:"joined_at"`
	Email       string `json:"email"`
	FullName    string `json:"full_name"`
	Phone       string `json:"phone"`
	Status      string `json:"status"`
	InviteToken string `json:"invite_token,omitempty"`
}

type Service struct {
	queries *db.Queries
	pool    *pgxpool.Pool
	rdb     *redis.Client
	billing BillingChecker
	mailer  *email.Mailer
	logger  *zap.Logger
}

func NewService(
	queries *db.Queries,
	pool *pgxpool.Pool,
	rdb *redis.Client,
	billing BillingChecker,
	mailer *email.Mailer,
	logger *zap.Logger,
) *Service {
	return &Service{
		queries: queries,
		pool:    pool,
		rdb:     rdb,
		billing: billing,
		mailer:  mailer,
		logger:  logger,
	}
}

func (s *Service) Queries() *db.Queries {
	return s.queries
}

func (s *Service) ListStaff(ctx context.Context, shopID string) ([]StaffMemberResponse, error) {
	shopUUID, err := uuid.Parse(shopID)
	if err != nil {
		return nil, ErrInvalidUUID
	}

	rows, err := s.queries.ListStaffMembersDetailed(ctx, pgutil.ToUUID(shopUUID))
	if err != nil {
		return nil, fmt.Errorf("list staff: %w", err)
	}

	staffMembers := make([]StaffMemberResponse, len(rows))
	for i, r := range rows {
		roleIDStr := ""
		if r.RoleID.Valid {
			roleIDStr = uuid.UUID(r.RoleID.Bytes).String()
		}
		statusStr := "active"
		if r.UserStatus.Valid {
			statusStr = string(r.UserStatus.UserStatus)
		}
		joinedAtStr := ""
		if r.JoinedAt.Valid {
			joinedAtStr = r.JoinedAt.Time.Format(time.RFC3339)
		}

		staffMembers[i] = StaffMemberResponse{
			ID:       r.ID.String(),
			ShopID:   shopUUID.String(),
			UserID:   uuid.UUID(r.UserID.Bytes).String(),
			RoleID:   roleIDStr,
			RoleName: r.RoleName,
			IsOwner:  r.IsOwner,
			JoinedAt: joinedAtStr,
			Email:    r.Email,
			FullName: r.FullName,
			Phone:    r.Phone,
			Status:   statusStr,
		}
	}

	return staffMembers, nil
}

func (s *Service) InviteStaff(
	ctx context.Context,
	shopID string,
	actorUserID string,
	actorName string,
	req InviteStaffRequest,
) (*StaffMemberResponse, error) {
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	if req.Email == "" || !strings.Contains(req.Email, "@") {
		return nil, ErrInvalidEmail
	}

	shopUUID, err := uuid.Parse(shopID)
	if err != nil {
		return nil, ErrInvalidUUID
	}

	roleUUID, err := uuid.Parse(req.RoleID)
	if err != nil {
		return nil, ErrInvalidUUID
	}

	role, err := s.queries.GetShopRole(ctx, roleUUID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrRoleNotFound
		}
		return nil, fmt.Errorf("get role: %w", err)
	}
	if role.ShopID.Valid && role.ShopID.Bytes != shopUUID {
		return nil, ErrRoleNotFound
	}

	if s.billing != nil {
		if err := s.billing.CheckStaffLimit(ctx, shopID); err != nil {
			return nil, ErrStaffLimit
		}
	}

	var targetUser db.User
	var isNewUser bool
	existingUser, err := s.queries.GetUserByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			newUser, createErr := s.queries.CreateUser(ctx, db.CreateUserParams{
				Email:    req.Email,
				FullName: pgutil.ToText(strings.TrimSpace(req.FullName)),
				Phone:    pgutil.ToText(strings.TrimSpace(req.Phone)),
			})
			if createErr != nil {
				return nil, fmt.Errorf("create user: %w", createErr)
			}

			updatedUser, updateErr := s.queries.UpdateUser(ctx, db.UpdateUserParams{
				ID:              newUser.ID,
				FullName:        newUser.FullName,
				Phone:           newUser.Phone,
				IsEmailVerified: newUser.IsEmailVerified,
				Status:          db.NullUserStatus{UserStatus: db.UserStatusPendingInvite, Valid: true},
			})
			if updateErr == nil {
				targetUser = updatedUser
			} else {
				targetUser = newUser
			}
			isNewUser = true
		} else {
			return nil, fmt.Errorf("get user: %w", err)
		}
	} else {
		targetUser = existingUser
	}

	_, err = s.queries.GetShopStaffByUserAndShop(ctx, db.GetShopStaffByUserAndShopParams{
		ShopID: pgutil.ToUUID(shopUUID),
		UserID: pgutil.ToUUID(targetUser.ID),
	})
	if err == nil {
		return nil, ErrAlreadyStaff
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("check shop staff: %w", err)
	}

	inviteToken, err := generateInviteToken()
	if err != nil {
		return nil, fmt.Errorf("generate invite token: %w", err)
	}

	payloadBytes, _ := json.Marshal(map[string]any{
		"shop_id":   shopID,
		"role_id":   req.RoleID,
		"role_name": role.Name,
		"email":     req.Email,
	})

	var invitedByUUID pgtype.UUID
	if actorUUID, parseErr := uuid.Parse(actorUserID); parseErr == nil {
		invitedByUUID = pgutil.ToUUID(actorUUID)
	}

	_, err = s.queries.CreateInvitation(ctx, db.CreateInvitationParams{
		Email:           req.Email,
		Token:           inviteToken,
		Type:            db.InviteTypeShopStaff,
		Payload:         payloadBytes,
		InvitedByUserID: invitedByUUID,
		ExpiresAt:       pgtype.Timestamptz{Time: time.Now().Add(7 * 24 * time.Hour), Valid: true},
	})
	if err != nil {
		return nil, fmt.Errorf("create invitation: %w", err)
	}

	createdStaff, err := s.queries.CreateShopStaff(ctx, db.CreateShopStaffParams{
		ShopID:  pgutil.ToUUID(shopUUID),
		UserID:  pgutil.ToUUID(targetUser.ID),
		RoleID:  pgutil.ToUUID(roleUUID),
		IsOwner: pgtype.Bool{Bool: false, Valid: true},
	})
	if err != nil {
		return nil, fmt.Errorf("create shop staff: %w", err)
	}

	s.invalidateRBAC(ctx, shopID, createdStaff.ID.String())

	if s.mailer != nil {
		shopName := "Store"
		if shopObj, shopErr := s.queries.GetShop(ctx, shopUUID); shopErr == nil && shopObj.Name != "" {
			shopName = shopObj.Name
		}
		inviter := actorName
		if inviter == "" {
			inviter = "The Store Administrator"
		}
		inviteLink := fmt.Sprintf("/admin/invite?token=%s", inviteToken)
		go func() {
			_ = s.mailer.SendStaffInvitation(req.Email, email.StaffInviteData{
				InviterName: inviter,
				ShopName:    shopName,
				RoleName:    role.Name,
				InviteLink:  inviteLink,
			})
		}()
	}

	statusStr := "active"
	if isNewUser || (targetUser.Status.Valid && targetUser.Status.UserStatus == db.UserStatusPendingInvite) {
		statusStr = "pending_invite"
	}

	return &StaffMemberResponse{
		ID:          createdStaff.ID.String(),
		ShopID:      shopUUID.String(),
		UserID:      targetUser.ID.String(),
		RoleID:      roleUUID.String(),
		RoleName:    role.Name,
		IsOwner:     false,
		JoinedAt:    createdStaff.JoinedAt.Time.Format(time.RFC3339),
		Email:       targetUser.Email,
		FullName:    targetUser.FullName.String,
		Phone:       targetUser.Phone.String,
		Status:      statusStr,
		InviteToken: inviteToken,
	}, nil
}

func (s *Service) UpdateStaffRole(ctx context.Context, shopID, staffID, roleID string) (*StaffMemberResponse, error) {
	shopUUID, err := uuid.Parse(shopID)
	if err != nil {
		return nil, ErrInvalidUUID
	}
	staffUUID, err := uuid.Parse(staffID)
	if err != nil {
		return nil, ErrInvalidUUID
	}
	roleUUID, err := uuid.Parse(roleID)
	if err != nil {
		return nil, ErrInvalidUUID
	}

	staffMember, err := s.queries.GetStaffMemberDetailed(ctx, db.GetStaffMemberDetailedParams{
		ID:     staffUUID,
		ShopID: pgutil.ToUUID(shopUUID),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrStaffNotFound
		}
		return nil, fmt.Errorf("get staff member: %w", err)
	}

	if staffMember.IsOwner {
		return nil, ErrCannotChangeOwnerRole
	}

	role, err := s.queries.GetShopRole(ctx, roleUUID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrRoleNotFound
		}
		return nil, fmt.Errorf("get role: %w", err)
	}
	if role.ShopID.Valid && role.ShopID.Bytes != shopUUID {
		return nil, ErrRoleNotFound
	}

	_, err = s.queries.UpdateShopStaffRoleForShop(ctx, db.UpdateShopStaffRoleForShopParams{
		ID:     staffUUID,
		ShopID: pgutil.ToUUID(shopUUID),
		RoleID: pgutil.ToUUID(roleUUID),
	})
	if err != nil {
		return nil, fmt.Errorf("update staff role: %w", err)
	}

	s.invalidateRBAC(ctx, shopID, staffID)

	statusStr := "active"
	if staffMember.UserStatus.Valid {
		statusStr = string(staffMember.UserStatus.UserStatus)
	}

	return &StaffMemberResponse{
		ID:       staffUUID.String(),
		ShopID:   shopUUID.String(),
		UserID:   uuid.UUID(staffMember.UserID.Bytes).String(),
		RoleID:   roleUUID.String(),
		RoleName: role.Name,
		IsOwner:  staffMember.IsOwner,
		JoinedAt: staffMember.JoinedAt.Time.Format(time.RFC3339),
		Email:    staffMember.Email,
		FullName: staffMember.FullName,
		Phone:    staffMember.Phone,
		Status:   statusStr,
	}, nil
}

func (s *Service) RemoveStaff(ctx context.Context, shopID, staffID string) error {
	shopUUID, err := uuid.Parse(shopID)
	if err != nil {
		return ErrInvalidUUID
	}
	staffUUID, err := uuid.Parse(staffID)
	if err != nil {
		return ErrInvalidUUID
	}

	staffMember, err := s.queries.GetShopStaff(ctx, staffUUID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrStaffNotFound
		}
		return fmt.Errorf("get staff: %w", err)
	}

	if staffMember.ShopID.Bytes != shopUUID {
		return ErrStaffNotFound
	}

	if staffMember.IsOwner.Bool {
		return ErrCannotRemoveOwner
	}

	err = s.queries.DeleteShopStaffProtected(ctx, db.DeleteShopStaffProtectedParams{
		ID:     staffUUID,
		ShopID: pgutil.ToUUID(shopUUID),
	})
	if err != nil {
		return fmt.Errorf("delete staff: %w", err)
	}

	s.invalidateRBAC(ctx, shopID, staffID)
	return nil
}

func (s *Service) invalidateRBAC(ctx context.Context, shopID, staffID string) {
	if s.rdb != nil {
		cacheKey := fmt.Sprintf("rbac:%s:%s", shopID, staffID)
		_ = s.rdb.Del(ctx, cacheKey).Err()
	}
}

func generateInviteToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
