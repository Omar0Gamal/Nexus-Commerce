package promotions

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"math/big"
	"strings"

	"backend-api/internal/db"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/redis/go-redis/v9"
)

// CartItem is used for tiered-discount computation.
type CartItem struct {
	ProductID uuid.UUID
	VariantID *uuid.UUID
	Quantity  int
	UnitPrice int64 // cents
}

// tier is a single step in the tiers JSONB array.
type tier struct {
	MinQty      int     `json:"min_qty"`
	DiscountPct float64 `json:"discount_pct"`
}

// ErrInsufficientPoints is returned when a customer has fewer points than requested.
var ErrInsufficientPoints = errors.New("insufficient loyalty points")

// Service handles flash sales, tiered discounts, loyalty, and referrals.
type Service struct {
	queries *db.Queries
	rdb     *redis.Client
}

func NewService(queries *db.Queries, rdb *redis.Client) *Service {
	return &Service{queries: queries, rdb: rdb}
}

// GetFlashSalePrice returns the active flash-sale price for a product/variant, or nil if none.
func (s *Service) GetFlashSalePrice(ctx context.Context, shopID, productID uuid.UUID, variantID *uuid.UUID) (*int32, error) {
	arg := db.GetFlashSalePriceParams{
		ShopID:    shopID,
		ProductID: productID,
	}
	if variantID != nil {
		arg.VariantID = pgtype.UUID{Bytes: *variantID, Valid: true}
	}
	price, err := s.queries.GetFlashSalePrice(ctx, arg)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &price, nil
}

// ApplyTieredDiscount computes total volume discount across cart items using order-level tiered discount.
func (s *Service) ApplyTieredDiscount(ctx context.Context, shopID uuid.UUID, items []CartItem) (int64, error) {
	discount, err := s.queries.GetActiveTieredDiscount(ctx, db.GetActiveTieredDiscountParams{
		ShopID:    shopID,
		AppliesTo: "order",
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}

	var tiers []tier
	if err := json.Unmarshal(discount.Tiers, &tiers); err != nil {
		return 0, fmt.Errorf("parse tiers: %w", err)
	}

	totalQty := 0
	totalCents := int64(0)
	for _, item := range items {
		totalQty += item.Quantity
		totalCents += int64(item.Quantity) * item.UnitPrice
	}

	pct := 0.0
	for _, t := range tiers {
		if totalQty >= t.MinQty && t.DiscountPct > pct {
			pct = t.DiscountPct
		}
	}
	return int64(math.Round(float64(totalCents) * pct / 100)), nil
}

// GetLoyaltyBalance returns the current points balance for a customer.
func (s *Service) GetLoyaltyBalance(ctx context.Context, shopID, customerID uuid.UUID) (int32, error) {
	acc, err := s.queries.GetLoyaltyAccount(ctx, db.GetLoyaltyAccountParams{
		ShopID:     shopID,
		CustomerID: customerID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	return acc.PointsBalance, nil
}

// EarnPoints awards points for a completed order (1 point per EGP by default).
func (s *Service) EarnPoints(ctx context.Context, shopID, customerID, orderID uuid.UUID, orderAmountCents int64) error {
	prog, err := s.queries.GetLoyaltyProgram(ctx, shopID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil // no program configured
	}
	if err != nil {
		return err
	}
	if !prog.IsActive {
		return nil
	}

	ppEgp, _ := prog.PointsPerEgp.Float64Value()
	ptsPerEGP := ppEgp.Float64
	if ptsPerEGP <= 0 {
		ptsPerEGP = 1.0
	}
	egpSpent := float64(orderAmountCents) / 100.0
	pts := int32(math.Round(egpSpent * ptsPerEGP))
	if pts <= 0 {
		return nil
	}

	if _, err := s.queries.UpsertLoyaltyAccount(ctx, db.UpsertLoyaltyAccountParams{
		ShopID:        shopID,
		CustomerID:    customerID,
		PointsBalance: pts,
	}); err != nil {
		return err
	}
	_, err = s.queries.RecordLoyaltyTransaction(ctx, db.RecordLoyaltyTransactionParams{
		ShopID:      shopID,
		CustomerID:  customerID,
		PointsDelta: pts,
		Reason:      "purchase",
		ReferenceID: pgtype.UUID{Bytes: orderID, Valid: true},
	})
	return err
}

// RedeemPoints validates and deducts loyalty points, returning the discount in cents.
func (s *Service) RedeemPoints(ctx context.Context, shopID, customerID uuid.UUID, points int32, orderAmountCents int64) (int64, error) {
	prog, err := s.queries.GetLoyaltyProgram(ctx, shopID)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, errors.New("loyalty program not configured")
	}
	if err != nil {
		return 0, err
	}
	if !prog.IsActive {
		return 0, errors.New("loyalty program inactive")
	}
	if points < int32(prog.MinRedeemPoints) {
		return 0, fmt.Errorf("minimum redemption is %d points", prog.MinRedeemPoints)
	}

	acc, err := s.queries.GetLoyaltyAccount(ctx, db.GetLoyaltyAccountParams{
		ShopID:     shopID,
		CustomerID: customerID,
	})
	if errors.Is(err, pgx.ErrNoRows) || acc.PointsBalance < points {
		return 0, ErrInsufficientPoints
	}
	if err != nil {
		return 0, err
	}

	eppVal, _ := prog.EgpPerPoint.Float64Value()
	egpRedeemed := float64(points) * eppVal.Float64
	discountCents := int64(math.Round(egpRedeemed * 100))

	// Cap at max_redeem_pct of order total.
	maxDiscount := int64(math.Round(float64(orderAmountCents) * float64(prog.MaxRedeemPct) / 100))
	if discountCents > maxDiscount {
		discountCents = maxDiscount
		egpRedeemed = float64(discountCents) / 100.0
		points = int32(math.Round(egpRedeemed / eppVal.Float64))
	}

	// Deduct.
	if _, err := s.queries.UpsertLoyaltyAccount(ctx, db.UpsertLoyaltyAccountParams{
		ShopID:        shopID,
		CustomerID:    customerID,
		PointsBalance: -points,
	}); err != nil {
		return 0, err
	}
	if _, err := s.queries.RecordLoyaltyTransaction(ctx, db.RecordLoyaltyTransactionParams{
		ShopID:      shopID,
		CustomerID:  customerID,
		PointsDelta: -points,
		Reason:      "redeem",
	}); err != nil {
		return 0, err
	}
	return discountCents, nil
}

// GetOrCreateReferralCode returns the customer's referral code, creating one if needed.
func (s *Service) GetOrCreateReferralCode(ctx context.Context, shopID, customerID uuid.UUID) (string, error) {
	rc, err := s.queries.GetReferralCode(ctx, db.GetReferralCodeParams{
		ShopID:     shopID,
		CustomerID: customerID,
	})
	if err == nil {
		return rc.Code, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return "", err
	}
	code, err := generateCode(8)
	if err != nil {
		return "", err
	}
	created, err := s.queries.CreateReferralCode(ctx, db.CreateReferralCodeParams{
		ShopID:     shopID,
		CustomerID: customerID,
		Code:       code,
	})
	if err != nil {
		return "", err
	}
	return created.Code, nil
}

// ProcessReferralConversion records a successful referral and rewards both parties.
func (s *Service) ProcessReferralConversion(ctx context.Context, shopID uuid.UUID, referralCode string, newCustomerID, orderID uuid.UUID) error {
	rc, err := s.queries.GetReferralCodeByCode(ctx, db.GetReferralCodeByCodeParams{
		ShopID: shopID,
		Code:   strings.ToUpper(referralCode),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil // no-op: unknown code
	}
	if err != nil {
		return err
	}
	if rc.CustomerID == newCustomerID {
		return nil // can't self-refer
	}

	// Award 100 pts to referrer by default.
	const referrerPts = 100
	if _, err := s.queries.RecordReferralConversion(ctx, db.RecordReferralConversionParams{
		ShopID:            shopID,
		ReferrerID:        rc.CustomerID,
		ReferredID:        newCustomerID,
		OrderID:           pgtype.UUID{Bytes: orderID, Valid: true},
		ReferrerRewardPts: referrerPts,
	}); err != nil {
		return err
	}
	if _, err := s.queries.UpsertLoyaltyAccount(ctx, db.UpsertLoyaltyAccountParams{
		ShopID:        shopID,
		CustomerID:    rc.CustomerID,
		PointsBalance: referrerPts,
	}); err != nil {
		return err
	}
	if _, err := s.queries.RecordLoyaltyTransaction(ctx, db.RecordLoyaltyTransactionParams{
		ShopID:      shopID,
		CustomerID:  rc.CustomerID,
		PointsDelta: referrerPts,
		Reason:      "referral",
		ReferenceID: pgtype.UUID{Bytes: newCustomerID, Valid: true},
	}); err != nil {
		return err
	}
	_ = s.queries.IncrementReferralCodeUses(ctx, rc.ID)
	return nil
}

// generateCode produces a random alphanumeric code of the given length.
func generateCode(n int) (string, error) {
	const chars = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	b := make([]byte, n)
	for i := range b {
		idx, err := rand.Int(rand.Reader, big.NewInt(int64(len(chars))))
		if err != nil {
			return "", err
		}
		b[i] = chars[idx.Int64()]
	}
	return string(b), nil
}


// ListFlashSales returns all active flash sales for a shop.
func (s *Service) ListFlashSales(ctx context.Context, shopID uuid.UUID) ([]db.FlashSale, error) {
	rows, err := s.queries.GetActiveFlashSales(ctx, shopID)
	if err != nil {
		return nil, err
	}
	if rows == nil {
		return []db.FlashSale{}, nil
	}
	return rows, nil
}

// CreateFlashSale creates a new flash sale.
func (s *Service) CreateFlashSale(ctx context.Context, shopID uuid.UUID, req CreateFlashSaleRequest) (db.FlashSale, error) {
	discVal := pgtype.Numeric{}
	_ = discVal.Scan(req.DiscountValue)
	var maxUses pgtype.Int4
	if req.MaxUses != nil {
		maxUses = pgtype.Int4{Int32: *req.MaxUses, Valid: true}
	}
	return s.queries.CreateFlashSale(ctx, db.CreateFlashSaleParams{
		ShopID:        shopID,
		Name:          req.Name,
		DiscountType:  req.DiscountType,
		DiscountValue: discVal,
		StartsAt:      pgtype.Timestamptz{Time: req.StartsAt, Valid: true},
		EndsAt:        pgtype.Timestamptz{Time: req.EndsAt, Valid: true},
		MaxUses:       maxUses,
	})
}

// AddFlashSaleItem adds a product to a flash sale.
func (s *Service) AddFlashSaleItem(ctx context.Context, shopID, flashSaleID, productID uuid.UUID, variantID *uuid.UUID, salePriceCents int32) (db.FlashSaleItem, error) {
	vid := pgtype.UUID{}
	if variantID != nil {
		vid = pgtype.UUID{Bytes: *variantID, Valid: true}
	}
	return s.queries.AddFlashSaleItem(ctx, db.AddFlashSaleItemParams{
		FlashSaleID:    flashSaleID,
		ShopID:         shopID,
		ProductID:      productID,
		VariantID:      vid,
		SalePriceCents: salePriceCents,
	})
}

// DeleteFlashSale removes a flash sale.
func (s *Service) DeleteFlashSale(ctx context.Context, shopID, id uuid.UUID) error {
	return s.queries.DeleteFlashSale(ctx, db.DeleteFlashSaleParams{
		ID:     id,
		ShopID: shopID,
	})
}

// ListTieredDiscounts returns all tiered discounts for a shop.
func (s *Service) ListTieredDiscounts(ctx context.Context, shopID uuid.UUID) ([]db.TieredDiscount, error) {
	rows, err := s.queries.GetTieredDiscounts(ctx, shopID)
	if err != nil {
		return nil, err
	}
	if rows == nil {
		return []db.TieredDiscount{}, nil
	}
	return rows, nil
}

// CreateTieredDiscount creates a new tiered discount.
func (s *Service) CreateTieredDiscount(ctx context.Context, shopID uuid.UUID, req CreateTieredDiscountRequest) (db.TieredDiscount, error) {
	targetID := pgtype.UUID{}
	if req.TargetID != nil {
		tid, err := uuid.Parse(*req.TargetID)
		if err != nil {
			return db.TieredDiscount{}, fmt.Errorf("invalid target_id: %w", err)
		}
		targetID = pgtype.UUID{Bytes: tid, Valid: true}
	}
	return s.queries.CreateTieredDiscount(ctx, db.CreateTieredDiscountParams{
		ShopID:    shopID,
		Name:      req.Name,
		AppliesTo: req.AppliesTo,
		TargetID:  targetID,
		Tiers:     []byte(req.Tiers),
	})
}

// GetLoyaltyProgram returns the loyalty program for a shop.
func (s *Service) GetLoyaltyProgram(ctx context.Context, shopID uuid.UUID) (db.LoyaltyProgram, error) {
	return s.queries.GetLoyaltyProgram(ctx, shopID)
}

// UpsertLoyaltyProgram creates or updates the loyalty program.
func (s *Service) UpsertLoyaltyProgram(ctx context.Context, shopID uuid.UUID, req UpdateLoyaltyProgramRequest) (db.LoyaltyProgram, error) {
	ppeVal := pgtype.Numeric{}
	_ = ppeVal.Scan(req.PointsPerEgp)
	eppVal := pgtype.Numeric{}
	_ = eppVal.Scan(req.EgpPerPoint)
	expiryMonths := pgtype.Int2{}
	if req.ExpiryMonths != nil {
		expiryMonths = pgtype.Int2{Int16: *req.ExpiryMonths, Valid: true}
	}
	return s.queries.UpsertLoyaltyProgram(ctx, db.UpsertLoyaltyProgramParams{
		ShopID:          shopID,
		PointsPerEgp:    ppeVal,
		EgpPerPoint:     eppVal,
		MinRedeemPoints: req.MinRedeemPoints,
		MaxRedeemPct:    req.MaxRedeemPct,
		ExpiryMonths:    expiryMonths,
		IsActive:        req.IsActive,
	})
}

// GetLoyaltyTransactions returns paginated loyalty transactions for a customer.
func (s *Service) GetLoyaltyTransactions(ctx context.Context, shopID, customerID uuid.UUID, limit, offset int32) ([]db.LoyaltyTransaction, error) {
	rows, err := s.queries.GetLoyaltyTransactions(ctx, db.GetLoyaltyTransactionsParams{
		ShopID:     shopID,
		CustomerID: customerID,
		Limit:      limit,
		Offset:     offset,
	})
	if err != nil {
		return nil, err
	}
	if rows == nil {
		return []db.LoyaltyTransaction{}, nil
	}
	return rows, nil
}


// GetReferralConfig returns the referral configuration for a shop.
func (s *Service) GetReferralConfig(ctx context.Context, shopID uuid.UUID) (db.ReferralConfig, error) {
	return s.queries.GetReferralConfig(ctx, shopID)
}

// UpsertReferralConfig creates or updates the referral configuration.
func (s *Service) UpsertReferralConfig(ctx context.Context, shopID uuid.UUID, req UpsertReferralConfigRequest) (db.ReferralConfig, error) {
	minOrder := pgtype.Numeric{}
	if req.MinOrderAmount != nil {
		_ = minOrder.Scan(*req.MinOrderAmount)
	}
	maxReferrals := pgtype.Int4{}
	if req.MaxReferralsPerCustomer != nil {
		maxReferrals = pgtype.Int4{Int32: int32(*req.MaxReferralsPerCustomer), Valid: true}
	}
	cookieWindow := int32(30)
	if req.CookieWindowDays != nil {
		cookieWindow = int32(*req.CookieWindowDays)
	}
	referrerRewardValue := pgtype.Numeric{}
	_ = referrerRewardValue.Scan(req.ReferrerRewardValue)
	refereeRewardValue := pgtype.Numeric{}
	_ = refereeRewardValue.Scan(req.RefereeRewardValue)

	return s.queries.UpsertReferralConfig(ctx, db.UpsertReferralConfigParams{
		ShopID:                  shopID,
		IsEnabled:               req.IsEnabled,
		ReferrerRewardType:      req.ReferrerRewardType,
		ReferrerRewardValue:     referrerRewardValue,
		RefereeRewardType:       req.RefereeRewardType,
		RefereeRewardValue:      refereeRewardValue,
		MinOrderAmount:          minOrder,
		MaxReferralsPerCustomer: maxReferrals,
		CookieWindowDays:        cookieWindow,
	})
}

// GetMyReferrals returns a list of conversions initiated by this customer.
func (s *Service) GetMyReferrals(ctx context.Context, shopID, customerID uuid.UUID) ([]db.GetMyReferralsRow, error) {
	rows, err := s.queries.GetMyReferrals(ctx, db.GetMyReferralsParams{
		ShopID:     shopID,
		CustomerID: customerID,
	})
	if err != nil {
		return nil, err
	}
	if rows == nil {
		return []db.GetMyReferralsRow{}, nil
	}
	return rows, nil
}

// GetReferralStats returns aggregate referral statistics for a shop.
func (s *Service) GetReferralStats(ctx context.Context, shopID uuid.UUID) (db.GetReferralStatsRow, error) {
	return s.queries.GetReferralStats(ctx, shopID)
}


// ListLoyaltyTiers returns all loyalty tiers for a shop ordered by position.
func (s *Service) ListLoyaltyTiers(ctx context.Context, shopID uuid.UUID) ([]db.LoyaltyTier, error) {
	rows, err := s.queries.ListLoyaltyTiers(ctx, shopID)
	if err != nil {
		return nil, err
	}
	if rows == nil {
		return []db.LoyaltyTier{}, nil
	}
	return rows, nil
}

// CreateLoyaltyTier creates a new loyalty tier for a shop.
func (s *Service) CreateLoyaltyTier(ctx context.Context, shopID uuid.UUID, req CreateLoyaltyTierRequest) (db.LoyaltyTier, error) {
	benefits := []byte(`{}`)
	if req.Benefits != nil {
		benefits = []byte(req.Benefits)
	}
	badgeColor := req.BadgeColor
	if badgeColor == "" {
		badgeColor = "#808080"
	}
	return s.queries.CreateLoyaltyTier(ctx, db.CreateLoyaltyTierParams{
		ShopID:            shopID,
		Name:              req.Name,
		MinLifetimePoints: req.MinLifetimePoints,
		Benefits:          benefits,
		BadgeLabel:        pgtype.Text{String: req.BadgeLabel, Valid: req.BadgeLabel != ""},
		BadgeColor:        pgtype.Text{String: badgeColor, Valid: true},
		Position:          req.Position,
	})
}

// UpdateLoyaltyTier updates an existing loyalty tier.
func (s *Service) UpdateLoyaltyTier(ctx context.Context, shopID, tierID uuid.UUID, req UpdateLoyaltyTierRequest) (db.LoyaltyTier, error) {
	var nameArg pgtype.Text
	if req.Name != nil {
		nameArg = pgtype.Text{String: *req.Name, Valid: true}
	}
	var minPtsArg pgtype.Int8
	if req.MinLifetimePoints != nil {
		minPtsArg = pgtype.Int8{Int64: *req.MinLifetimePoints, Valid: true}
	}
	var badgeLabelArg pgtype.Text
	if req.BadgeLabel != nil {
		badgeLabelArg = pgtype.Text{String: *req.BadgeLabel, Valid: true}
	}
	var badgeColorArg pgtype.Text
	if req.BadgeColor != nil {
		badgeColorArg = pgtype.Text{String: *req.BadgeColor, Valid: true}
	}
	var posArg pgtype.Int4
	if req.Position != nil {
		posArg = pgtype.Int4{Int32: *req.Position, Valid: true}
	}
	var benefitsArg []byte
	if req.Benefits != nil {
		benefitsArg = []byte(req.Benefits)
	}
	return s.queries.UpdateLoyaltyTier(ctx, db.UpdateLoyaltyTierParams{
		ID:                tierID,
		ShopID:            shopID,
		Name:              nameArg,
		MinLifetimePoints: minPtsArg,
		BadgeLabel:        badgeLabelArg,
		BadgeColor:        badgeColorArg,
		Position:          posArg,
		Benefits:          benefitsArg,
	})
}

// DeleteLoyaltyTier deletes a loyalty tier.
func (s *Service) DeleteLoyaltyTier(ctx context.Context, shopID, tierID uuid.UUID) error {
	return s.queries.DeleteLoyaltyTier(ctx, db.DeleteLoyaltyTierParams{
		ID:     tierID,
		ShopID: shopID,
	})
}

// GetCustomerTierStatus returns the customer's current tier, next tier, and progress.
func (s *Service) GetCustomerTierStatus(ctx context.Context, shopID, customerID uuid.UUID) (CustomerTierResponse, error) {
	resp := CustomerTierResponse{}

	account, err := s.queries.GetLoyaltyAccount(ctx, db.GetLoyaltyAccountParams{
		ShopID:     shopID,
		CustomerID: customerID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return resp, nil
		}
		return resp, err
	}
	resp.LifetimePoints = account.LifetimePoints

	current, err := s.queries.GetCustomerCurrentTier(ctx, db.GetCustomerCurrentTierParams{
		ShopID:     shopID,
		CustomerID: customerID,
	})
	if err == nil {
		s := tierToSummary(current)
		resp.CurrentTier = &s
	}

	next, err := s.queries.GetNextTier(ctx, db.GetNextTierParams{
		ShopID:     shopID,
		CustomerID: customerID,
	})
	if err == nil {
		n := tierToSummary(next)
		resp.NextTier = &n
		resp.PointsToNextTier = next.MinLifetimePoints - int64(account.LifetimePoints)
		if resp.CurrentTier != nil && next.MinLifetimePoints > resp.CurrentTier.MinLifetimePoints {
			span := next.MinLifetimePoints - resp.CurrentTier.MinLifetimePoints
			earned := int64(account.LifetimePoints) - resp.CurrentTier.MinLifetimePoints
			if span > 0 {
				resp.ProgressPercent = int(earned * 100 / span)
			}
		}
	}

	return resp, nil
}

// ManualAdjustPoints adds or subtracts points from a customer's balance (staff only).
func (s *Service) ManualAdjustPoints(ctx context.Context, shopID uuid.UUID, req ManualAdjustPointsRequest) error {
	customerID, err := uuid.Parse(req.CustomerID)
	if err != nil {
		return fmt.Errorf("invalid customer_id")
	}
	if _, err := s.queries.UpsertLoyaltyAccount(ctx, db.UpsertLoyaltyAccountParams{
		ShopID:        shopID,
		CustomerID:    customerID,
		PointsBalance: req.Delta,
	}); err != nil {
		return err
	}
	_, err = s.queries.RecordLoyaltyTransaction(ctx, db.RecordLoyaltyTransactionParams{
		ShopID:      shopID,
		CustomerID:  customerID,
		PointsDelta: req.Delta,
		Reason:      req.Reason,
	})
	return err
}

// GetLoyaltyLeaderboard returns the top customers by available points.
func (s *Service) GetLoyaltyLeaderboard(ctx context.Context, shopID uuid.UUID, limit int32) ([]db.GetLoyaltyLeaderboardRow, error) {
	rows, err := s.queries.GetLoyaltyLeaderboard(ctx, db.GetLoyaltyLeaderboardParams{
		ShopID: shopID,
		Limit:  limit,
	})
	if err != nil {
		return nil, err
	}
	if rows == nil {
		return []db.GetLoyaltyLeaderboardRow{}, nil
	}
	return rows, nil
}

func tierToSummary(t db.LoyaltyTier) TierSummary {
	return TierSummary{
		ID:                t.ID.String(),
		Name:              t.Name,
		MinLifetimePoints: t.MinLifetimePoints,
		BadgeLabel:        t.BadgeLabel.String,
		BadgeColor:        t.BadgeColor.String,
		Benefits:          t.Benefits,
	}
}
