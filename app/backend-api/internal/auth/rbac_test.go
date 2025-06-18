package auth

import (
	"encoding/json"
	"testing"
)

// hasPermission and isTruthy are pure functions — no DB/Redis needed.

func TestHasPermission_ArrayWildcard(t *testing.T) {
	perms, _ := json.Marshal([]string{"*"})
	if !hasPermission(perms, "products.delete") {
		t.Error("wildcard array should grant any permission")
	}
}

func TestHasPermission_ArrayExact(t *testing.T) {
	perms, _ := json.Marshal([]string{"products.read", "orders.write"})
	if !hasPermission(perms, "products.read") {
		t.Error("exact match should be granted")
	}
	if hasPermission(perms, "products.delete") {
		t.Error("unlisted permission should be denied")
	}
}

func TestHasPermission_ObjectWildcard(t *testing.T) {
	perms, _ := json.Marshal(map[string]any{"*": true})
	if !hasPermission(perms, "anything") {
		t.Error("object wildcard should grant any permission")
	}
}

func TestHasPermission_ObjectExact(t *testing.T) {
	perms, _ := json.Marshal(map[string]any{
		"products.read": true,
		"orders.write":  true,
	})
	if !hasPermission(perms, "products.read") {
		t.Error("exact key true should be granted")
	}
	if hasPermission(perms, "products.delete") {
		t.Error("absent key should be denied")
	}
}

func TestHasPermission_ObjectFalsyValues(t *testing.T) {
	perms, _ := json.Marshal(map[string]any{
		"products.read": false,
		"orders.write":  "false",
	})
	if hasPermission(perms, "products.read") {
		t.Error("false bool should be denied")
	}
	if hasPermission(perms, "orders.write") {
		t.Error(`"false" string should be denied`)
	}
}

func TestHasPermission_InvalidJSON(t *testing.T) {
	if hasPermission([]byte("not-json"), "perm") {
		t.Error("invalid JSON should deny")
	}
}

func TestHasPermission_EmptyArray(t *testing.T) {
	perms, _ := json.Marshal([]string{})
	if hasPermission(perms, "anything") {
		t.Error("empty array should deny all")
	}
}

func TestIsTruthy(t *testing.T) {
	cases := []struct {
		val  any
		want bool
	}{
		{true, true},
		{false, false},
		{"true", true},
		{"1", true},
		{"false", false},
		{"", false},
		{float64(1), true},
		{float64(0), false},
		{nil, false},
		{42, false}, // unknown type
	}
	for _, tc := range cases {
		got := isTruthy(tc.val)
		if got != tc.want {
			t.Errorf("isTruthy(%v) = %v, want %v", tc.val, got, tc.want)
		}
	}
}
