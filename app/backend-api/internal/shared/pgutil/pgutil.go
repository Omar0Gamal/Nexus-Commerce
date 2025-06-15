// Package pgutil provides conversion helpers between Go types and pgx/pgtype types.
package pgutil

import (
	"fmt"
	"math"
	"math/big"
	"strings"

	"backend-api/internal/shared/apperr"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

// ToUUID converts a google/uuid.UUID to pgtype.UUID.
func ToUUID(id uuid.UUID) pgtype.UUID {
	return pgtype.UUID{Bytes: id, Valid: true}
}

// ParseUUID parses a string UUID into pgtype.UUID.
func ParseUUID(s string) (pgtype.UUID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return pgtype.UUID{}, err
	}
	return pgtype.UUID{Bytes: id, Valid: true}, nil
}

// ParseStdUUID parses a string UUID into uuid.UUID, returning apperr.ErrInvalidUUID on failure.
func ParseStdUUID(s string) (uuid.UUID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return uuid.UUID{}, apperr.ErrInvalidUUID
	}
	return id, nil
}

// UUIDToString converts pgtype.UUID to its string representation, or "" if invalid.
func UUIDToString(u pgtype.UUID) string {
	if !u.Valid {
		return ""
	}
	return uuid.UUID(u.Bytes).String()
}

// ToText converts a Go string to pgtype.Text. An empty string produces an invalid (NULL) Text.
func ToText(s string) pgtype.Text {
	if s == "" {
		return pgtype.Text{}
	}
	return pgtype.Text{String: s, Valid: true}
}

// TextToString converts pgtype.Text to a Go string, returning "" if NULL.
func TextToString(t pgtype.Text) string {
	if t.Valid {
		return t.String
	}
	return ""
}

// NumericToString renders pgtype.Numeric as a two-decimal-place string (e.g. "99.99").
func NumericToString(n pgtype.Numeric) string {
	if !n.Valid || n.Int == nil {
		return "0.00"
	}
	f, _ := n.Float64Value()
	if f.Valid {
		return fmt.Sprintf("%.2f", f.Float64)
	}
	return "0.00"
}

// NumericToStringPtr returns nil if the numeric is invalid, otherwise a pointer
// to the two-decimal string representation.
func NumericToStringPtr(n pgtype.Numeric) *string {
	if !n.Valid {
		return nil
	}
	s := NumericToString(n)
	return &s
}

// NumericToFloat converts pgtype.Numeric to float64, returning 0 on error.
func NumericToFloat(n pgtype.Numeric) float64 {
	f, err := n.Float64Value()
	if err != nil || !f.Valid {
		return 0
	}
	return f.Float64
}

// Float64ToNumeric converts a float64 to pgtype.Numeric with two decimal places.
func Float64ToNumeric(f float64) pgtype.Numeric {
	if f == 0 {
		return pgtype.Numeric{Valid: false}
	}
	s := fmt.Sprintf("%.2f", f)
	parts := strings.SplitN(s, ".", 2)
	intStr := parts[0] + parts[1]
	n := new(big.Int)
	n.SetString(intStr, 10)
	return pgtype.Numeric{Int: n, Exp: -2, Valid: true, NaN: false}
}

// FloatToNumericCents converts a float64 price to pgtype.Numeric using integer
// cent arithmetic to avoid floating-point precision issues.
func FloatToNumericCents(f float64) pgtype.Numeric {
	cents := int64(math.Round(f * 100))
	return pgtype.Numeric{Int: big.NewInt(cents), Exp: -2, Valid: true}
}

// Int4ToInt32 converts pgtype.Int4 to int32, returning 0 if invalid.
func Int4ToInt32(i pgtype.Int4) int32 {
	if !i.Valid {
		return 0
	}
	return i.Int32
}
