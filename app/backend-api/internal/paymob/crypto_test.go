package paymob

import (
	"testing"
)

func TestEncryptDecryptRoundtrip(t *testing.T) {
	key := []byte("my-32-byte-secret-key-for-tests!")
	plaintext := `{"sub_merchant_code":"MER-001"}`

	ct, err := encryptCredentials(key, plaintext)
	if err != nil {
		t.Fatalf("encryptCredentials: %v", err)
	}
	if ct == plaintext {
		t.Error("ciphertext should differ from plaintext")
	}

	got, err := decryptCredentials(key, ct)
	if err != nil {
		t.Fatalf("decryptCredentials: %v", err)
	}
	if got != plaintext {
		t.Errorf("roundtrip: got %q, want %q", got, plaintext)
	}
}

func TestEncryptCredentials_EmptyKey(t *testing.T) {
	// Empty key → dev passthrough.
	plaintext := "raw-credentials"
	ct, err := encryptCredentials(nil, plaintext)
	if err != nil {
		t.Fatalf("encryptCredentials with empty key: %v", err)
	}
	if ct != plaintext {
		t.Errorf("expected passthrough, got %q", ct)
	}
}

func TestDecryptCredentials_EmptyKey(t *testing.T) {
	value := "raw-credentials"
	got, err := decryptCredentials(nil, value)
	if err != nil {
		t.Fatalf("decryptCredentials with empty key: %v", err)
	}
	if got != value {
		t.Errorf("expected passthrough, got %q", got)
	}
}

func TestDecryptCredentials_LegacyPlaintext(t *testing.T) {
	// A value that was stored before encryption was introduced:
	// It's not valid base64, so it should be returned unchanged.
	key := []byte("my-32-byte-secret-key-for-tests!")
	legacy := "not-base64-encoded-value"

	got, err := decryptCredentials(key, legacy)
	if err != nil {
		t.Fatalf("decryptCredentials legacy: %v", err)
	}
	if got != legacy {
		t.Errorf("legacy fallback: got %q, want %q", got, legacy)
	}
}

func TestEncryptCredentials_UniqueNonces(t *testing.T) {
	key := []byte("my-32-byte-secret-key-for-tests!")
	plaintext := "same-value"

	ct1, _ := encryptCredentials(key, plaintext)
	ct2, _ := encryptCredentials(key, plaintext)

	if ct1 == ct2 {
		t.Error("each encryption should produce a unique ciphertext (randomised nonce)")
	}
}

func TestPrepareKey_Short(t *testing.T) {
	k := prepareKey([]byte("short"))
	if len(k) != 32 {
		t.Errorf("prepareKey short: len=%d, want 32", len(k))
	}
}

func TestPrepareKey_Long(t *testing.T) {
	input := []byte("this-key-is-much-longer-than-thirty-two-bytes-total")
	k := prepareKey(input)
	if len(k) != 32 {
		t.Errorf("prepareKey long: len=%d, want 32", len(k))
	}
}
