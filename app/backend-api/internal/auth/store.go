package auth

import (
	"context"

	"backend-api/internal/db"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type Store interface {
	Queries() *db.Queries
	// Users
	GetUser(ctx context.Context, id uuid.UUID) (db.User, error)
	GetUserByEmail(ctx context.Context, email string) (db.User, error)
	CreateUser(ctx context.Context, p db.CreateUserParams) (db.User, error)
	UpdateUser(ctx context.Context, p db.UpdateUserParams) (db.User, error)
	UpdateUserPassword(ctx context.Context, p db.UpdateUserPasswordParams) error

	// Secrets (2FA)
	GetUserSecrets(ctx context.Context, userID uuid.UUID) (db.UserSecret, error)
	CreateUserSecrets(ctx context.Context, p db.CreateUserSecretsParams) (db.UserSecret, error)
	UpdateUserSecrets(ctx context.Context, p db.UpdateUserSecretsParams) (db.UserSecret, error)
	Enable2FA(ctx context.Context, p db.Enable2FAParams) error
	Disable2FA(ctx context.Context, userID uuid.UUID) error

	// Shops
	GetShop(ctx context.Context, id uuid.UUID) (db.Shop, error)
	GetShopBySubdomain(ctx context.Context, subdomain string) (db.Shop, error)
	CreateShop(ctx context.Context, p db.CreateShopParams) (db.Shop, error)

	// Staff
	GetShopStaff(ctx context.Context, id uuid.UUID) (db.ShopStaff, error)
	GetShopStaffByUserAndShop(ctx context.Context, p db.GetShopStaffByUserAndShopParams) (db.ShopStaff, error)
	ListStaffWithUsers(ctx context.Context, shopID uuid.UUID) ([]db.ListStaffWithUsersRow, error)
	CreateShopStaff(ctx context.Context, p db.CreateShopStaffParams) (db.ShopStaff, error)
	DeleteShopStaffProtected(ctx context.Context, p db.DeleteShopStaffProtectedParams) error

	// Roles
	GetShopRole(ctx context.Context, id uuid.UUID) (db.ShopRole, error)
	ListShopRoles(ctx context.Context, shopID uuid.UUID) ([]db.ShopRole, error)
	CreateShopRole(ctx context.Context, p db.CreateShopRoleParams) (db.ShopRole, error)
	UpdateShopRole(ctx context.Context, p db.UpdateShopRoleParams) (db.ShopRole, error)
	DeleteShopRole(ctx context.Context, id uuid.UUID) error
	CountStaffByRole(ctx context.Context, roleID uuid.UUID) (int64, error)
	UpdateShopStaffRole(ctx context.Context, p db.UpdateShopStaffRoleParams) (db.ShopStaff, error)

	// Platform admins
	GetPlatformAdmin(ctx context.Context, id uuid.UUID) (db.PlatformAdmin, error)
	GetPlatformAdminByEmail(ctx context.Context, email string) (db.PlatformAdmin, error)

	// Customers
	GetCustomer(ctx context.Context, p db.GetCustomerParams) (db.Customer, error)
	GetCustomerByEmail(ctx context.Context, p db.GetCustomerByEmailParams) (db.Customer, error)
	CreateCustomer(ctx context.Context, p db.CreateCustomerParams) (db.Customer, error)
	UpdateCustomer(ctx context.Context, p db.UpdateCustomerParams) (db.Customer, error)
	UpdateCustomerPassword(ctx context.Context, p db.UpdateCustomerPasswordParams) error

	// Transactions
	WithTx(tx pgx.Tx) Store
}

// DBStore wraps *db.Queries and implements Store.
// Add methods here as needed; the compiler enforces completeness.
type DBStore struct{ q *db.Queries }

func NewDBStore(q *db.Queries) Store { return &DBStore{q: q} }

func (s *DBStore) Queries() *db.Queries { return s.q }

func (s *DBStore) WithTx(tx pgx.Tx) Store { return &DBStore{q: s.q.WithTx(tx)} }

// Users

func (s *DBStore) GetUser(ctx context.Context, id uuid.UUID) (db.User, error) {
	return s.q.GetUser(ctx, id)
}

func (s *DBStore) GetUserByEmail(ctx context.Context, email string) (db.User, error) {
	return s.q.GetUserByEmail(ctx, email)
}

func (s *DBStore) CreateUser(ctx context.Context, p db.CreateUserParams) (db.User, error) {
	return s.q.CreateUser(ctx, p)
}

func (s *DBStore) UpdateUser(ctx context.Context, p db.UpdateUserParams) (db.User, error) {
	return s.q.UpdateUser(ctx, p)
}

func (s *DBStore) UpdateUserPassword(ctx context.Context, p db.UpdateUserPasswordParams) error {
	return s.q.UpdateUserPassword(ctx, p)
}

// Secrets (2FA)

func (s *DBStore) GetUserSecrets(ctx context.Context, userID uuid.UUID) (db.UserSecret, error) {
	return s.q.GetUserSecrets(ctx, userID)
}

func (s *DBStore) CreateUserSecrets(ctx context.Context, p db.CreateUserSecretsParams) (db.UserSecret, error) {
	return s.q.CreateUserSecrets(ctx, p)
}

func (s *DBStore) UpdateUserSecrets(ctx context.Context, p db.UpdateUserSecretsParams) (db.UserSecret, error) {
	return s.q.UpdateUserSecrets(ctx, p)
}

func (s *DBStore) Enable2FA(ctx context.Context, p db.Enable2FAParams) error {
	return s.q.Enable2FA(ctx, p)
}

func (s *DBStore) Disable2FA(ctx context.Context, userID uuid.UUID) error {
	return s.q.Disable2FA(ctx, userID)
}

// Shops

func (s *DBStore) GetShop(ctx context.Context, id uuid.UUID) (db.Shop, error) {
	return s.q.GetShop(ctx, id)
}

func (s *DBStore) GetShopBySubdomain(ctx context.Context, subdomain string) (db.Shop, error) {
	return s.q.GetShopBySubdomain(ctx, subdomain)
}

func (s *DBStore) CreateShop(ctx context.Context, p db.CreateShopParams) (db.Shop, error) {
	return s.q.CreateShop(ctx, p)
}

// Staff

func (s *DBStore) GetShopStaff(ctx context.Context, id uuid.UUID) (db.ShopStaff, error) {
	return s.q.GetShopStaff(ctx, id)
}

func (s *DBStore) GetShopStaffByUserAndShop(ctx context.Context, p db.GetShopStaffByUserAndShopParams) (db.ShopStaff, error) {
	return s.q.GetShopStaffByUserAndShop(ctx, p)
}

func (s *DBStore) ListStaffWithUsers(ctx context.Context, shopID uuid.UUID) ([]db.ListStaffWithUsersRow, error) {
	return s.q.ListStaffWithUsers(ctx, pgtype.UUID{Bytes: shopID, Valid: true})
}

func (s *DBStore) CreateShopStaff(ctx context.Context, p db.CreateShopStaffParams) (db.ShopStaff, error) {
	return s.q.CreateShopStaff(ctx, p)
}

func (s *DBStore) DeleteShopStaffProtected(ctx context.Context, p db.DeleteShopStaffProtectedParams) error {
	return s.q.DeleteShopStaffProtected(ctx, p)
}

// Roles

func (s *DBStore) GetShopRole(ctx context.Context, id uuid.UUID) (db.ShopRole, error) {
	return s.q.GetShopRole(ctx, id)
}

func (s *DBStore) ListShopRoles(ctx context.Context, shopID uuid.UUID) ([]db.ShopRole, error) {
	return s.q.ListShopRoles(ctx, pgtype.UUID{Bytes: shopID, Valid: true})
}

func (s *DBStore) CreateShopRole(ctx context.Context, p db.CreateShopRoleParams) (db.ShopRole, error) {
	return s.q.CreateShopRole(ctx, p)
}

func (s *DBStore) UpdateShopRole(ctx context.Context, p db.UpdateShopRoleParams) (db.ShopRole, error) {
	return s.q.UpdateShopRole(ctx, p)
}

func (s *DBStore) DeleteShopRole(ctx context.Context, id uuid.UUID) error {
	return s.q.DeleteShopRole(ctx, id)
}

func (s *DBStore) CountStaffByRole(ctx context.Context, roleID uuid.UUID) (int64, error) {
	return s.q.CountStaffByRole(ctx, pgtype.UUID{Bytes: roleID, Valid: true})
}

func (s *DBStore) UpdateShopStaffRole(ctx context.Context, p db.UpdateShopStaffRoleParams) (db.ShopStaff, error) {
	return s.q.UpdateShopStaffRole(ctx, p)
}

// Platform admins

func (s *DBStore) GetPlatformAdmin(ctx context.Context, id uuid.UUID) (db.PlatformAdmin, error) {
	return s.q.GetPlatformAdmin(ctx, id)
}

func (s *DBStore) GetPlatformAdminByEmail(ctx context.Context, email string) (db.PlatformAdmin, error) {
	return s.q.GetPlatformAdminByEmail(ctx, email)
}

// Customers

func (s *DBStore) GetCustomer(ctx context.Context, p db.GetCustomerParams) (db.Customer, error) {
	return s.q.GetCustomer(ctx, p)
}

func (s *DBStore) GetCustomerByEmail(ctx context.Context, p db.GetCustomerByEmailParams) (db.Customer, error) {
	return s.q.GetCustomerByEmail(ctx, p)
}

func (s *DBStore) CreateCustomer(ctx context.Context, p db.CreateCustomerParams) (db.Customer, error) {
	return s.q.CreateCustomer(ctx, p)
}

func (s *DBStore) UpdateCustomer(ctx context.Context, p db.UpdateCustomerParams) (db.Customer, error) {
	return s.q.UpdateCustomer(ctx, p)
}

func (s *DBStore) UpdateCustomerPassword(ctx context.Context, p db.UpdateCustomerPasswordParams) error {
	return s.q.UpdateCustomerPassword(ctx, p)
}
