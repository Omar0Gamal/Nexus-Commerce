package webhooks

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"io"
)

func prepareKey(key []byte) []byte {
	k := make([]byte, 32)
	copy(k, key)
	return k
}

// Returns base64(nonce || ciphertext+tag). Empty key returns plaintext unchanged for dev mode.
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

// Returns raw value unchanged if key is empty or decryption fails (legacy plaintext support).
func decryptSecret(key []byte, value string) string {
	if len(key) == 0 {
		return value
	}

	decodedBytes, err := base64.StdEncoding.DecodeString(value)
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

	if len(decodedBytes) < gcm.NonceSize() {
		return value
	}

	nonce, ciphertext := decodedBytes[:gcm.NonceSize()], decodedBytes[gcm.NonceSize():]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return value
	}

	return string(plaintext)
}
