package pgutil_test

import (
	"testing"

	"backend-api/internal/shared/apperr"
	"backend-api/internal/shared/pgutil"

	"github.com/google/uuid"
)

func TestParseStdUUID_Valid(t *testing.T) {
	id := uuid.New()
	got, err := pgutil.ParseStdUUID(id.String())
	if err != nil {
		t.Fatalf("ParseStdUUID: unexpected error: %v", err)
	}
	if got != id {
		t.Errorf("ParseStdUUID: got %v, want %v", got, id)
	}
}

func TestParseStdUUID_Invalid(t *testing.T) {
	_, err := pgutil.ParseStdUUID("not-a-uuid")
	if err != apperr.ErrInvalidUUID {
		t.Errorf("ParseStdUUID: got %v, want %v", err, apperr.ErrInvalidUUID)
	}
}

func TestParseStdUUID_Empty(t *testing.T) {
	_, err := pgutil.ParseStdUUID("")
	if err != apperr.ErrInvalidUUID {
		t.Errorf("ParseStdUUID: got %v, want %v", err, apperr.ErrInvalidUUID)
	}
}

func TestParseUUID_Valid(t *testing.T) {
	id := uuid.New()
	pgUUID, err := pgutil.ParseUUID(id.String())
	if err != nil {
		t.Fatalf("ParseUUID: unexpected error: %v", err)
	}
	if !pgUUID.Valid {
		t.Error("ParseUUID: expected Valid=true")
	}
	if uuid.UUID(pgUUID.Bytes) != id {
		t.Errorf("ParseUUID: bytes mismatch: got %v, want %v", uuid.UUID(pgUUID.Bytes), id)
	}
}

func TestParseUUID_Invalid(t *testing.T) {
	_, err := pgutil.ParseUUID("not-a-uuid")
	if err == nil {
		t.Error("ParseUUID: expected error for invalid input")
	}
}

func TestToUUID(t *testing.T) {
	id := uuid.New()
	pgUUID := pgutil.ToUUID(id)
	if !pgUUID.Valid {
		t.Error("ToUUID: expected Valid=true")
	}
	if uuid.UUID(pgUUID.Bytes) != id {
		t.Errorf("ToUUID: bytes mismatch: got %v, want %v", uuid.UUID(pgUUID.Bytes), id)
	}
}

func TestUUIDToString_Valid(t *testing.T) {
	id := uuid.New()
	pgUUID := pgutil.ToUUID(id)
	got := pgutil.UUIDToString(pgUUID)
	if got != id.String() {
		t.Errorf("UUIDToString: got %q, want %q", got, id.String())
	}
}

func TestUUIDToString_Invalid(t *testing.T) {
	pgUUID := pgutil.ToUUID(uuid.UUID{})
	pgUUID.Valid = false
	got := pgutil.UUIDToString(pgUUID)
	if got != "" {
		t.Errorf("UUIDToString for invalid UUID: got %q, want empty string", got)
	}
}
