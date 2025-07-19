package webhooks

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

// encryptSecret encrypts a webhook secret key using AES-256-GCM.
//
// Returns base64(nonce || ciphertext+tag).
// If key is empty the plaintext is returned unchanged (dev mode).
func encryptSecret(key []byte, plaintext string) (string, error) {
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

// decryptSecret decrypts a base64-encoded AES-256-GCM ciphertext produced by
// encryptSecret. If key is empty, or decryption fails (legacy plaintext rows),
// the raw value is returned unchanged.
func decryptSecret(key []byte, value string) string {
	if len(key) == 0 {
		return value
	}

	data, err := base64.StdEncoding.DecodeString(value)
	if err != nil {
		return value
	}

	block, err := aes.NewCipher(prepareKey(key))
	if err != nil {
		return value
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return value
	}

	if len(data) < gcm.NonceSize() {
		return value
	}

	nonce, ciphertext := data[:gcm.NonceSize()], data[gcm.NonceSize():]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return value
	}

	return string(plaintext)
}
