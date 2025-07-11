package billing

import "testing"

func TestNewMeteringServiceLoggerInitialization(t *testing.T) {
	m := NewMeteringService(nil, nil)
	if m == nil {
		t.Fatal("expected metering service instance")
	}
	if m.log == nil {
		t.Fatal("expected default logger to be initialized")
	}

	m2 := NewMeteringService(&Service{}, nil)
	if m2.log == nil {
		t.Fatal("expected logger to be initialized with non-nil service")
	}
}
