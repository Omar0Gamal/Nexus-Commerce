package paymob

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"io"
)

// prepareKey pads or truncates key to exactly 32 bytes for AES-256.
func prepareKey(key []byte) []byte {
	k := make([]byte, 32)
	copy(k, key)
	return k
}

// encryptCredentials encrypts plaintext with AES-256-GCM.
//
// Returned value is base64(nonce || ciphertext+tag).
// If key is empty the plaintext is returned unchanged — this keeps the dev
// environment working without a key configured.
func encryptCredentials(key []byte, plaintext string) (string, error) {
	if len(key) == 0 {
		return plaintext, nil
	}

	block, err := aes.NewCipher(prepareKey(key))
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	sealed := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(sealed), nil
}

// decryptCredentials decrypts a base64-encoded AES-256-GCM ciphertext produced
// by encryptCredentials.
//
// If key is empty, or if decoding/decryption fails (e.g. legacy plaintext rows),
// the raw value is returned unchanged so existing rows continue to work during
// a key-rotation migration.
func decryptCredentials(key []byte, value string) (string, error) {
	if len(key) == 0 {
		return value, nil
	}

	data, err := base64.StdEncoding.DecodeString(value)
	if err != nil {
		// Not base64 — treat as pre-encryption legacy row.
		return value, nil
	}

	block, err := aes.NewCipher(prepareKey(key))
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	if len(data) < gcm.NonceSize() {
		// Too short to be AES-GCM output — legacy plaintext.
		return value, nil
	}

	nonce, ciphertext := data[:gcm.NonceSize()], data[gcm.NonceSize():]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		// Wrong key or corrupted — fall back to raw value rather than hard-failing.
		return value, nil
	}

	return string(plaintext), nil
}
