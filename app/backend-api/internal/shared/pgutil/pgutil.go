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

func ToUUID(id uuid.UUID) pgtype.UUID {
	return pgtype.UUID{Bytes: id, Valid: true}
}

func ParseUUID(s string) (pgtype.UUID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return pgtype.UUID{}, err
	}
	return pgtype.UUID{Bytes: id, Valid: true}, nil
}

func ParseStdUUID(s string) (uuid.UUID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return uuid.UUID{}, apperr.ErrInvalidUUID
	}
	return id, nil
}

func UUIDToString(u pgtype.UUID) string {
	if !u.Valid {
		return ""
	}
	return uuid.UUID(u.Bytes).String()
}

func ToText(s string) pgtype.Text {
	if s == "" {
		return pgtype.Text{}
	}
	return pgtype.Text{String: s, Valid: true}
}

func TextToString(t pgtype.Text) string {
	if t.Valid {
		return t.String
	}
	return ""
}

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

func NumericToStringPtr(n pgtype.Numeric) *string {
	if !n.Valid {
		return nil
	}
	s := NumericToString(n)
	return &s
}

func NumericToFloat(n pgtype.Numeric) float64 {
	f, err := n.Float64Value()
	if err != nil || !f.Valid {
		return 0
	}
	return f.Float64
}

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

// FloatToNumericCents rounds via integer cent arithmetic to avoid IEEE-754 representation drift.
func FloatToNumericCents(f float64) pgtype.Numeric {
	cents := int64(math.Round(f * 100))
	return pgtype.Numeric{Int: big.NewInt(cents), Exp: -2, Valid: true}
}

func Int4ToInt32(i pgtype.Int4) int32 {
	if !i.Valid {
		return 0
	}
	return i.Int32
}
