package paymob

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

// If key is empty, plaintext is returned unchanged to support dev environments.
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

// If key is empty or decryption fails, the raw value is returned to preserve
// backwards compatibility with legacy plaintext rows during key rotation.
func decryptCredentials(key []byte, cipherText string) (string, error) {
	if len(key) == 0 {
		return cipherText, nil
	}

	rawBytes, err := base64.StdEncoding.DecodeString(cipherText)
	if err != nil {
		return cipherText, nil
	}

	block, err := aes.NewCipher(prepareKey(key))
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	if len(rawBytes) < gcm.NonceSize() {
		return cipherText, nil
	}

	nonce, ciphertext := rawBytes[:gcm.NonceSize()], rawBytes[gcm.NonceSize():]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return cipherText, nil
	}

	return string(plaintext), nil
}
