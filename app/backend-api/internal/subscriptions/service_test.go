package subscriptions

import (
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"
)

func TestMapErr(t *testing.T) {
	if got := mapErr(nil); got != nil {
		t.Fatalf("mapErr(nil) = %v, want nil", got)
	}

	if got := mapErr(pgx.ErrNoRows); !errors.Is(got, ErrNotFound) {
		t.Fatalf("mapErr(ErrNoRows) = %v, want ErrNotFound", got)
	}

	boom := errors.New("boom")
	if got := mapErr(boom); !errors.Is(got, boom) {
		t.Fatalf("mapErr(boom) = %v, want original error", got)
	}
}
