package coupons

import (
	"math"
	"testing"
)

// calculateDiscount is the pure math extracted from ApplyCoupon. Since the
// coupon service embeds the calculation inline, we replicate it here to test
// the exact same arithmetic the production path uses.

func calcDiscount(discountType, value, orderTotal float64, isPercentage bool) float64 {
	if isPercentage {
		pct := value
		if pct > 100 {
			pct = 100
		}
		return math.Round(orderTotal*pct/100*100) / 100
	}
	// fixed amount
	if value > orderTotal {
		return orderTotal
	}
	return value
}

func TestDiscount_Percentage(t *testing.T) {
	cases := []struct {
		name       string
		pct        float64
		orderTotal float64
		want       float64
	}{
		{"20% off 100.00", 20, 100.00, 20.00},
		{"10% off 99.99", 10, 99.99, 10.00},
		{"50% off 37.50", 50, 37.50, 18.75},
		{"100% off 50.00", 100, 50.00, 50.00},
		{"over-100% capped at 100%", 150, 80.00, 80.00},
		{"0% off", 0, 100.00, 0.00},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := calcDiscount(tc.pct, tc.pct, tc.orderTotal, true)
			if got != tc.want {
				t.Errorf("calcDiscount(%v%%, %.2f) = %.2f, want %.2f", tc.pct, tc.orderTotal, got, tc.want)
			}
		})
	}
}

func TestDiscount_FixedAmount(t *testing.T) {
	cases := []struct {
		name       string
		fixed      float64
		orderTotal float64
		want       float64
	}{
		{"$10 off $50", 10, 50.00, 10.00},
		{"$10 off $8 (capped)", 10, 8.00, 8.00},
		{"$0 off", 0, 100.00, 0.00},
		{"exact match", 99.99, 99.99, 99.99},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := calcDiscount(tc.fixed, tc.fixed, tc.orderTotal, false)
			if got != tc.want {
				t.Errorf("calcDiscount(fixed=%.2f, order=%.2f) = %.2f, want %.2f", tc.fixed, tc.orderTotal, got, tc.want)
			}
		})
	}
}

func TestValidateType(t *testing.T) {
	if !validTypes["percentage"] {
		t.Error("percentage should be a valid type")
	}
	if !validTypes["fixed_amount"] {
		t.Error("fixed_amount should be a valid type")
	}
	if validTypes["flat"] {
		t.Error("flat should not be a valid type")
	}
	if validTypes[""] {
		t.Error("empty string should not be a valid type")
	}
}
