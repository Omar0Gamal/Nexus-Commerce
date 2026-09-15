package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"backend-api/internal/db"
	"backend-api/internal/shared/pgutil"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

var (
	ErrRoleNotFound  = errors.New("role not found")
	ErrRoleInUse     = errors.New("role has assigned staff members and cannot be deleted")
	ErrSystemRole    = errors.New("system roles cannot be modified or deleted")
	ErrStaffNotFound = errors.New("staff member not found")
	ErrInvalidUUID   = errors.New("invalid uuid")
)


type RoleResponse struct {
	ID           string          `json:"id"`
	Name         string          `json:"name"`
	Permissions  json.RawMessage `json:"permissions"`
	IsSystemRole bool            `json:"is_system_role"`
	ParentRoleID *string         `json:"parent_role_id,omitempty"`
}

type CreateRoleRequest struct {
	Name        string          `json:"name"        binding:"required"`
	Permissions json.RawMessage `json:"permissions" binding:"required"`
}

type UpdateRoleRequest struct {
	Name        *string         `json:"name"`
	Permissions json.RawMessage `json:"permissions"`
}

type AssignRoleRequest struct {
	RoleID string `json:"role_id" binding:"required"`
}

func (s *Service) ListRoles(ctx context.Context, shopID string) ([]RoleResponse, error) {
	id, err := uuid.Parse(shopID)
	if err != nil {
		return nil, ErrInvalidUUID
	}
	roles, err := s.Queries().ListShopRoles(ctx, pgutil.ToUUID(id))
	if err != nil {
		return nil, fmt.Errorf("list roles: %w", err)
	}
	roleResponses := make([]RoleResponse, len(roles))
	for i, r := range roles {
		roleResponses[i] = roleToResponse(r)
	}
	return roleResponses, nil
}

func (s *Service) CreateRole(ctx context.Context, shopID string, req CreateRoleRequest) (*RoleResponse, error) {
	id, err := uuid.Parse(shopID)
	if err != nil {
		return nil, ErrInvalidUUID
	}
	if !json.Valid(req.Permissions) {
		return nil, fmt.Errorf("permissions must be valid JSON")
	}
	role, err := s.Queries().CreateShopRole(ctx, db.CreateShopRoleParams{
		ShopID:       pgutil.ToUUID(id),
		ParentRoleID: pgtype.UUID{},
		Name:         req.Name,
		Permissions:  []byte(req.Permissions),
		IsSystemRole: pgtype.Bool{Bool: false, Valid: true},
	})
	if err != nil {
		return nil, fmt.Errorf("create role: %w", err)
	}
	resp := roleToResponse(role)
	return &resp, nil
}

func (s *Service) UpdateRole(ctx context.Context, shopID, roleID string, req UpdateRoleRequest) (*RoleResponse, error) {
	rid, err := uuid.Parse(roleID)
	if err != nil {
		return nil, ErrInvalidUUID
	}

	existing, err := s.Queries().GetShopRole(ctx, rid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrRoleNotFound
		}
		return nil, fmt.Errorf("get role: %w", err)
	}
	if existing.IsSystemRole.Bool {
		return nil, ErrSystemRole
	}

	var nameArg pgtype.Text
	if req.Name != nil {
		nameArg = pgtype.Text{String: *req.Name, Valid: true}
	}
	var permsArg []byte
	if req.Permissions != nil && json.Valid(req.Permissions) {
		permsArg = []byte(req.Permissions)
	}

	updated, err := s.Queries().UpdateShopRole(ctx, db.UpdateShopRoleParams{
		ID:           rid,
		Name:         nameArg,
		Permissions:  permsArg,
		ParentRoleID: pgtype.UUID{},
	})
	if err != nil {
		return nil, fmt.Errorf("update role: %w", err)
	}

	// Invalidate RBAC cache for shop roles to maintain cross-session consistency.
	s.InvalidateRBACCacheForShop(ctx, shopID)

	resp := roleToResponse(updated)
	return &resp, nil
}

func (s *Service) DeleteRole(ctx context.Context, shopID, roleID string) error {
	rid, err := uuid.Parse(roleID)
	if err != nil {
		return ErrInvalidUUID
	}

	existing, err := s.Queries().GetShopRole(ctx, rid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrRoleNotFound
		}
		return fmt.Errorf("get role: %w", err)
	}
	if existing.IsSystemRole.Bool {
		return ErrSystemRole
	}

	count, err := s.Queries().CountStaffByRole(ctx, pgutil.ToUUID(rid))
	if err != nil {
		return fmt.Errorf("check role members: %w", err)
	}
	if count > 0 {
		return ErrRoleInUse
	}

	if err := s.Queries().DeleteShopRole(ctx, rid); err != nil {
		return fmt.Errorf("delete role: %w", err)
	}
	return nil
}

func (s *Service) AssignStaffRole(ctx context.Context, shopID, staffID, roleID string) error {
	sid, err := uuid.Parse(staffID)
	if err != nil {
		return ErrInvalidUUID
	}
	rid, err := uuid.Parse(roleID)
	if err != nil {
		return ErrInvalidUUID
	}

	role, err := s.Queries().GetShopRole(ctx, rid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrRoleNotFound
		}
		return fmt.Errorf("get role: %w", err)
	}
	shopUUID, _ := uuid.Parse(shopID)
	if role.ShopID.Bytes != shopUUID {
		return ErrRoleNotFound
	}

	_, err = s.Queries().UpdateShopStaffRole(ctx, db.UpdateShopStaffRoleParams{
		ID:     sid,
		RoleID: pgutil.ToUUID(rid),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrStaffNotFound
		}
		return fmt.Errorf("assign role: %w", err)
	}

	s.InvalidateRBACCache(ctx, shopID, staffID)
	return nil
}


func roleToResponse(r db.ShopRole) RoleResponse {
	resp := RoleResponse{
		ID:           r.ID.String(),
		Name:         r.Name,
		Permissions:  json.RawMessage(r.Permissions),
		IsSystemRole: r.IsSystemRole.Bool,
	}
	if r.ParentRoleID.Valid {
		s := uuid.UUID(r.ParentRoleID.Bytes).String()
		resp.ParentRoleID = &s
	}
	return resp
}
