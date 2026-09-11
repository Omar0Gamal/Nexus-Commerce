package db

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

// GetOrCreateStaffRole returns the "staff" system role for the given shop,
// creating it if it does not yet exist. This conditional upsert logic cannot be
// expressed as a single sqlc query, so it lives here.
func (q *Queries) GetOrCreateStaffRole(ctx context.Context, shopID pgtype.UUID) (uuid.UUID, error) {
	const getSQL = `SELECT id FROM shop_roles WHERE shop_id=$1 AND name='staff' LIMIT 1`
	var roleID uuid.UUID
	err := q.db.QueryRow(ctx, getSQL, shopID).Scan(&roleID)
	if err == nil {
		return roleID, nil
	}

	role, createErr := q.CreateShopRole(ctx, CreateShopRoleParams{
		ShopID:       shopID,
		ParentRoleID: pgtype.UUID{},
		Name:         "staff",
		Permissions:  []byte(`["read","write"]`),
		IsSystemRole: pgtype.Bool{Bool: true, Valid: true},
	})
	if createErr != nil {
		return uuid.UUID{}, createErr
	}
	return role.ID, nil
}
