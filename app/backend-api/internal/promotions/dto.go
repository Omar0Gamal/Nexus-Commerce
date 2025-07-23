package promotions

import (
	"encoding/json"
	"time"
)

// CreateFlashSaleRequest holds the fields required to create a flash sale.
type CreateFlashSaleRequest struct {
	Name          string    `json:"name"           binding:"required"`
	DiscountType  string    `json:"discount_type"  binding:"required"`
	DiscountValue float64   `json:"discount_value" binding:"required"`
	StartsAt      time.Time `json:"starts_at"      binding:"required"`
	EndsAt        time.Time `json:"ends_at"        binding:"required"`
	MaxUses       *int32    `json:"max_uses"`
}

// AddFlashSaleItemRequest holds the fields required to add an item to a flash sale.
type AddFlashSaleItemRequest struct {
	ProductID      string  `json:"product_id"       binding:"required"`
	VariantID      *string `json:"variant_id"`
	SalePriceCents int32   `json:"sale_price_cents" binding:"required"`
}

// CreateTieredDiscountRequest holds the fields required to create a tiered discount.
type CreateTieredDiscountRequest struct {
	Name      string          `json:"name"       binding:"required"`
	AppliesTo string          `json:"applies_to" binding:"required"`
	TargetID  *string         `json:"target_id"`
	Tiers     json.RawMessage `json:"tiers"      binding:"required"`
}

// UpdateLoyaltyProgramRequest holds the fields for upserting a loyalty program.
type UpdateLoyaltyProgramRequest struct {
	PointsPerEgp    float64 `json:"points_per_egp"`
	EgpPerPoint     float64 `json:"egp_per_point"`
	MinRedeemPoints int32   `json:"min_redeem_points"`
	MaxRedeemPct    int16   `json:"max_redeem_pct"`
	ExpiryMonths    *int16  `json:"expiry_months"`
	IsActive        bool    `json:"is_active"`
}

// RedeemPointsRequest holds the fields for redeeming loyalty points.
type RedeemPointsRequest struct {
	Points           int32 `json:"points"             binding:"required"`
	OrderAmountCents int64 `json:"order_amount_cents" binding:"required"`
}

// UpsertReferralConfigRequest holds the fields for creating/updating referral config.
type UpsertReferralConfigRequest struct {
	IsEnabled               bool    `json:"is_enabled"`
	ReferrerRewardType      string  `json:"referrer_reward_type"  binding:"required"`
	ReferrerRewardValue     float64 `json:"referrer_reward_value" binding:"required"`
	RefereeRewardType       string  `json:"referee_reward_type"   binding:"required"`
	RefereeRewardValue      float64 `json:"referee_reward_value"  binding:"required"`
	MinOrderAmount          *string `json:"min_order_amount"`
	MaxReferralsPerCustomer *int    `json:"max_referrals_per_customer"`
	CookieWindowDays        *int    `json:"cookie_window_days"`
}

// CreateLoyaltyTierRequest holds the fields for creating a loyalty tier.
type CreateLoyaltyTierRequest struct {
	Name              string          `json:"name"                binding:"required"`
	MinLifetimePoints int64           `json:"min_lifetime_points" binding:"required"`
	BadgeLabel        string          `json:"badge_label"`
	BadgeColor        string          `json:"badge_color"`
	Benefits          json.RawMessage `json:"benefits"`
	Position          int32           `json:"position"`
}

// UpdateLoyaltyTierRequest holds the fields for updating a loyalty tier.
type UpdateLoyaltyTierRequest struct {
	Name              *string         `json:"name"`
	MinLifetimePoints *int64          `json:"min_lifetime_points"`
	BadgeLabel        *string         `json:"badge_label"`
	BadgeColor        *string         `json:"badge_color"`
	Benefits          json.RawMessage `json:"benefits"`
	Position          *int32          `json:"position"`
}

// ManualAdjustPointsRequest holds the fields for manually adjusting a customer's points.
type ManualAdjustPointsRequest struct {
	CustomerID string `json:"customer_id" binding:"required"`
	Delta      int32  `json:"delta"       binding:"required"`
	Reason     string `json:"reason"      binding:"required"`
}

// CustomerTierResponse is the response for GET /loyalty/my-tier.
type CustomerTierResponse struct {
	CurrentTier      *TierSummary `json:"current_tier"`
	NextTier         *TierSummary `json:"next_tier"`
	LifetimePoints   int32        `json:"lifetime_points"`
	PointsToNextTier int64        `json:"points_to_next_tier"`
	ProgressPercent  int          `json:"progress_percent"`
}

// TierSummary is a minimal tier representation.
type TierSummary struct {
	ID                string          `json:"id"`
	Name              string          `json:"name"`
	MinLifetimePoints int64           `json:"min_lifetime_points"`
	BadgeLabel        string          `json:"badge_label"`
	BadgeColor        string          `json:"badge_color"`
	Benefits          json.RawMessage `json:"benefits"`
}
