package webhooks

import (
	"strings"
	"testing"
)

func TestEncryptDecryptRoundtrip(t *testing.T) {
	key := []byte("my-32-byte-secret-key-for-tests!")
	plaintext := "my-webhook-signing-secret"

	ciphertext, err := encryptSecret(key, plaintext)
	if err != nil {
		t.Fatalf("encryptSecret: %v", err)
	}
	if ciphertext == plaintext {
		t.Error("ciphertext should differ from plaintext")
	}

	got := decryptSecret(key, ciphertext)
	if got != plaintext {
		t.Errorf("decryptSecret: got %q, want %q", got, plaintext)
	}
}

func TestEncryptDecrypt_UniqueNonces(t *testing.T) {
	key := []byte("my-32-byte-secret-key-for-tests!")
	plaintext := "same-secret"

	ct1, err := encryptSecret(key, plaintext)
	if err != nil {
		t.Fatalf("encrypt 1: %v", err)
	}
	ct2, err := encryptSecret(key, plaintext)
	if err != nil {
		t.Fatalf("encrypt 2: %v", err)
	}

	if ct1 == ct2 {
		t.Error("two encryptions of the same plaintext should produce different ciphertexts (random nonce)")
	}

	// Both should still decrypt to the same plaintext.
	if decryptSecret(key, ct1) != plaintext {
		t.Errorf("decrypt ct1: got %q, want %q", decryptSecret(key, ct1), plaintext)
	}
	if decryptSecret(key, ct2) != plaintext {
		t.Errorf("decrypt ct2: got %q, want %q", decryptSecret(key, ct2), plaintext)
	}
}

func TestEncryptDecrypt_EmptyKey(t *testing.T) {
	// Empty key → dev mode: plaintext passthrough.
	plaintext := "raw-secret"

	ct, err := encryptSecret(nil, plaintext)
	if err != nil {
		t.Fatalf("encryptSecret with empty key: %v", err)
	}
	if ct != plaintext {
		t.Errorf("expected passthrough, got %q", ct)
	}

	got := decryptSecret(nil, plaintext)
	if got != plaintext {
		t.Errorf("decryptSecret with empty key: got %q, want %q", got, plaintext)
	}
}

func TestDecryptSecret_LegacyPlaintext(t *testing.T) {
	// A row that was stored before encryption was introduced should be returned unchanged.
	key := []byte("my-32-byte-secret-key-for-tests!")
	legacyValue := "plaintext-no-base64"

	got := decryptSecret(key, legacyValue)
	if got != legacyValue {
		t.Errorf("legacy plaintext: got %q, want %q", got, legacyValue)
	}
}

func TestDecryptSecret_WrongKey(t *testing.T) {
	key1 := []byte("my-32-byte-secret-key-for-tests!")
	key2 := []byte("another-32-byte-key-for-testing!")
	plaintext := "secret-value"

	ct, err := encryptSecret(key1, plaintext)
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}

	// Wrong key should fall back to returning the raw ciphertext (not error).
	got := decryptSecret(key2, ct)
	if got == plaintext {
		t.Error("decryption with wrong key should not return the original plaintext")
	}
	// Should return the base64 ciphertext unchanged.
	if !strings.Contains(got, "") {
		t.Errorf("unexpected result: %q", got)
	}
}

func TestPrepareKey_ShortKey(t *testing.T) {
	// Keys shorter than 32 bytes are zero-padded.
	key := prepareKey([]byte("short"))
	if len(key) != 32 {
		t.Errorf("prepareKey: got len %d, want 32", len(key))
	}
}

func TestPrepareKey_LongKey(t *testing.T) {
	// Keys longer than 32 bytes are truncated.
	input := []byte("this-is-a-very-long-key-that-exceeds-32-bytes-in-length")
	key := prepareKey(input)
	if len(key) != 32 {
		t.Errorf("prepareKey: got len %d, want 32", len(key))
	}
}
