package billing

import (
	"testing"
)

func TestFeatureEnabled_Bool(t *testing.T) {
	cases := []struct {
		name         string
		featureValue any
		want         bool
	}{
		{"bool true", true, true},
		{"bool false", false, false},
		{"string true", "true", true},
		{"string 1", "1", true},
		{"string yes", "yes", true},
		{"string false", "false", false},
		{"string empty", "", false},
		{"float64 nonzero", float64(1), true},
		{"float64 zero", float64(0), false},
		{"nil", nil, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			features := map[string]any{"k": tc.featureValue}
			got := featureEnabled(features, "k")
			if got != tc.want {
				t.Errorf("featureEnabled(%v) = %v, want %v", tc.featureValue, got, tc.want)
			}
		})
	}
}

func TestFeatureEnabled_MissingKey(t *testing.T) {
	features := map[string]any{"other": true}
	if featureEnabled(features, "missing") {
		t.Error("expected false for missing key")
	}
}

func TestFeatureEnabled_NilMap(t *testing.T) {
	if featureEnabled(nil, "any") {
		t.Error("expected false for nil map")
	}
}
