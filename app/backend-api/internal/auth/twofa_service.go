package auth

import (
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"image/png"
	"io"
	"strings"
	"time"

	"backend-api/internal/db"
	"backend-api/internal/shared/token"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

var (
	Err2FANotEnabled  = errors.New("2FA is not enabled for this account")
	Err2FAAlreadyOn   = errors.New("2FA is already enabled")
	ErrInvalid2FACode = errors.New("invalid 2FA code")
	ErrInvalidBackup  = errors.New("invalid or already-used backup code")
	ErrRateLimited    = errors.New("too many failed login attempts")
)

const maxLoginAttempts = 10

type fallbackFailedLogin struct {
	Count     int
	ExpiresAt time.Time
}

func staffBruteForceKey(email string) string {
	return fmt.Sprintf("failed_login:staff:%s", strings.ToLower(email))
}

func customerBruteForceKey(shopID, email string) string {
	return fmt.Sprintf("failed_login:%s:%s", shopID, strings.ToLower(email))
}

// checkBruteForce returns ErrRateLimited if the key is at or above maxLoginAttempts.
func (s *Service) checkBruteForce(ctx context.Context, key string) error {
	if s.rdb == nil {
		return s.checkFallbackBruteForce(key)
	}

	val, err := s.rdb.Get(ctx, key).Int()
	if err == redis.Nil {
		s.clearFallbackFailedLogin(key)
		return nil
	}
	if err != nil {
		s.logger.Warn("redis unavailable during brute-force check; using local fallback",
			zap.String("key", key),
			zap.Error(err),
		)
		return s.checkFallbackBruteForce(key)
	}
	if val >= maxLoginAttempts {
		return ErrRateLimited
	}
	return nil
}

// recordFailedLogin increments the attempt counter with a progressive TTL.
func (s *Service) recordFailedLogin(ctx context.Context, key string) {
	if s.rdb == nil {
		s.recordFallbackFailedLogin(key)
		return
	}

	val, err := s.rdb.Incr(ctx, key).Result()
	if err != nil {
		s.logger.Warn("redis unavailable during failed login increment; using local fallback",
			zap.String("key", key),
			zap.Error(err),
		)
		s.recordFallbackFailedLogin(key)
		return
	}

	ttl := failedLoginTTL(int(val))
	if err := s.rdb.Expire(ctx, key, ttl).Err(); err != nil {
		s.logger.Warn("failed to set brute-force key TTL",
			zap.String("key", key),
			zap.Error(err),
		)
	}
}

// clearFailedLogin deletes the brute-force counter after a successful login.
func (s *Service) clearFailedLogin(ctx context.Context, key string) {
	if s.rdb != nil {
		if err := s.rdb.Del(ctx, key).Err(); err != nil {
			s.logger.Warn("failed to clear redis brute-force key",
				zap.String("key", key),
				zap.Error(err),
			)
		}
	}
	s.clearFallbackFailedLogin(key)
}

func failedLoginTTL(count int) time.Duration {
	switch {
	case count >= 5:
		return time.Hour
	case count >= 3:
		return 15 * time.Minute
	case count >= 2:
		return 5 * time.Minute
	default:
		return time.Minute
	}
}

func (s *Service) checkFallbackBruteForce(key string) error {
	v, ok := s.loginFail.Load(key)
	if !ok {
		return nil
	}
	entry, ok := v.(fallbackFailedLogin)
	if !ok {
		s.loginFail.Delete(key)
		return nil
	}
	if time.Now().After(entry.ExpiresAt) {
		s.loginFail.Delete(key)
		return nil
	}
	if entry.Count >= maxLoginAttempts {
		return ErrRateLimited
	}
	return nil
}

func (s *Service) recordFallbackFailedLogin(key string) {
	now := time.Now()
	count := 1
	if v, ok := s.loginFail.Load(key); ok {
		if entry, ok := v.(fallbackFailedLogin); ok && now.Before(entry.ExpiresAt) {
			count = entry.Count + 1
		}
	}

	s.loginFail.Store(key, fallbackFailedLogin{
		Count:     count,
		ExpiresAt: now.Add(failedLoginTTL(count)),
	})
}

func (s *Service) clearFallbackFailedLogin(key string) {
	s.loginFail.Delete(key)
}

// encryptTOTPSecret encrypts a TOTP secret with AES-256-GCM.
// If encKey is empty the secret is returned as-is (dev mode).
func encryptTOTPSecret(secret, encKey string) (string, error) {
	if encKey == "" {
		return secret, nil
	}
	key, err := hex.DecodeString(encKey)
	if err != nil || len(key) != 32 {
		return "", fmt.Errorf("TOTP_ENCRYPTION_KEY must be a 64-char hex string (32 bytes)")
	}
	block, err := aes.NewCipher(key)
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
	ciphertext := gcm.Seal(nonce, nonce, []byte(secret), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// decryptTOTPSecret reverses encryptTOTPSecret.
func decryptTOTPSecret(stored, encKey string) (string, error) {
	if encKey == "" {
		return stored, nil
	}
	key, err := hex.DecodeString(encKey)
	if err != nil || len(key) != 32 {
		return "", fmt.Errorf("TOTP_ENCRYPTION_KEY must be a 64-char hex string (32 bytes)")
	}
	data, err := base64.StdEncoding.DecodeString(stored)
	if err != nil {
		return "", fmt.Errorf("decode totp secret: %w", err)
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", errors.New("ciphertext too short")
	}
	plaintext, err := gcm.Open(nil, data[:nonceSize], data[nonceSize:], nil)
	if err != nil {
		return "", fmt.Errorf("decrypt totp secret: %w", err)
	}
	return string(plaintext), nil
}

// Setup2FA generates a new TOTP secret + backup codes, stores them (unactivated),
// and returns the provisioning URI + QR code + plaintext backup codes.
func (s *Service) Setup2FA(ctx context.Context, userID, encKey string) (*Setup2FAResponse, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, ErrUserNotFound
	}

	user, err := s.Queries().GetUser(ctx, uid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("get user: %w", err)
	}

	// Check 2FA not already active
	secrets, err := s.Queries().GetUserSecrets(ctx, uid)
	if err == nil && secrets.Is2faEnabled.Bool {
		return nil, Err2FAAlreadyOn
	}

	// Generate TOTP key
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "NexusCommerce",
		AccountName: user.Email,
		Algorithm:   otp.AlgorithmSHA1,
		Digits:      otp.DigitsSix,
	})
	if err != nil {
		return nil, fmt.Errorf("generate totp key: %w", err)
	}

	// Encrypt the secret before storing
	encSecret, err := encryptTOTPSecret(key.Secret(), encKey)
	if err != nil {
		return nil, fmt.Errorf("encrypt totp secret: %w", err)
	}

	// Generate 8 backup codes
	plainCodes := make([]string, 8)
	hashedCodes := make([]string, 8)
	for i := range plainCodes {
		buf := make([]byte, 5)
		if _, err := rand.Read(buf); err != nil {
			return nil, fmt.Errorf("generate backup code: %w", err)
		}
		code := strings.ToUpper(hex.EncodeToString(buf))
		plainCodes[i] = code[:5] + "-" + code[5:]
		h, err := bcrypt.GenerateFromPassword([]byte(code), 10)
		if err != nil {
			return nil, fmt.Errorf("hash backup code: %w", err)
		}
		hashedCodes[i] = string(h)
	}

	codesJSON, err := json.Marshal(hashedCodes)
	if err != nil {
		return nil, fmt.Errorf("marshal backup codes: %w", err)
	}

	// Upsert user_secrets — store secret but keep is_2fa_enabled = false until Verify2FA
	if err := s.Queries().Enable2FA(ctx, db.Enable2FAParams{
		UserID:      uid,
		TotpSecret:  pgtype.Text{String: encSecret, Valid: true},
		BackupCodes: codesJSON,
	}); err != nil {
		// If no row exists, create one first
		_, createErr := s.Queries().CreateUserSecrets(ctx, db.CreateUserSecretsParams{
			UserID:       uid,
			TotpSecret:   pgtype.Text{String: encSecret, Valid: true},
			BackupCodes:  codesJSON,
			Is2faEnabled: pgtype.Bool{Bool: false, Valid: true},
		})
		if createErr != nil {
			return nil, fmt.Errorf("store 2fa setup: %w", createErr)
		}
	}

	// Generate QR code PNG → base64
	img, err := key.Image(200, 200)
	if err != nil {
		return nil, fmt.Errorf("generate qr image: %w", err)
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, fmt.Errorf("encode qr png: %w", err)
	}
	qrB64 := "data:image/png;base64," + base64.StdEncoding.EncodeToString(buf.Bytes())

	return &Setup2FAResponse{
		ProvisioningURI: key.URL(),
		QRCodeBase64:    qrB64,
		BackupCodes:     plainCodes,
	}, nil
}

// Verify2FA validates a TOTP code and permanently enables 2FA on the account.
func (s *Service) Verify2FA(ctx context.Context, userID, code, encKey string) error {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return ErrUserNotFound
	}

	secrets, err := s.Queries().GetUserSecrets(ctx, uid)
	if err != nil {
		return fmt.Errorf("get user secrets: %w", err)
	}
	if !secrets.TotpSecret.Valid || secrets.TotpSecret.String == "" {
		return Err2FANotEnabled
	}

	rawSecret, err := decryptTOTPSecret(secrets.TotpSecret.String, encKey)
	if err != nil {
		return err
	}

	if !totp.Validate(code, rawSecret) {
		return ErrInvalid2FACode
	}

	// Mark 2FA as fully enabled
	if _, err := s.Queries().UpdateUserSecrets(ctx, db.UpdateUserSecretsParams{
		UserID:       uid,
		Is2faEnabled: pgtype.Bool{Bool: true, Valid: true},
	}); err != nil {
		return fmt.Errorf("enable 2fa: %w", err)
	}
	return nil
}

// Challenge2FA validates a TOTP code (or backup code) against the 2FA challenge token
// and returns full auth tokens on success.
func (s *Service) Challenge2FA(ctx context.Context, req Challenge2FARequest, encKey string) (*AuthResponse, error) {
	claims, err := s.tokenSvc.Validate2FAChallengeToken(req.LoginToken)
	if err != nil {
		return nil, token.ErrInvalidToken
	}

	uid, err := uuid.Parse(claims.UserID)
	if err != nil {
		return nil, token.ErrInvalidToken
	}

	secrets, err := s.Queries().GetUserSecrets(ctx, uid)
	if err != nil || !secrets.Is2faEnabled.Bool {
		return nil, Err2FANotEnabled
	}

	rawSecret, err := decryptTOTPSecret(secrets.TotpSecret.String, encKey)
	if err != nil {
		return nil, err
	}

	// Try TOTP first, then backup code
	codeValid := totp.Validate(req.Code, rawSecret)
	if !codeValid {
		if err := s.validateAndConsumeBackupCode(ctx, uid, req.Code, secrets.BackupCodes); err != nil {
			return nil, ErrInvalid2FACode
		}
	}

	// Re-fetch user and rebuild full auth response
	user, err := s.Queries().GetUser(ctx, uid)
	if err != nil {
		return nil, ErrUserNotFound
	}

	if claims.ShopID != "" {
		shopID, err := uuid.Parse(claims.ShopID)
		if err == nil {
			shop, sErr := s.Queries().GetShop(ctx, shopID)
			if sErr == nil {
				staff, roleName, rErr := s.getStaffContext(ctx, shopID, uid)
				if rErr == nil {
					return s.generateAuthResponse(user, &shop, roleName, staff.IsOwner.Bool)
				}
			}
		}
	}
	return s.generateAuthResponse(user, nil, "", false)
}

// Disable2FA validates the current TOTP code and then clears 2FA from the account.
func (s *Service) Disable2FA(ctx context.Context, userID, code, encKey string) error {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return ErrUserNotFound
	}

	secrets, err := s.Queries().GetUserSecrets(ctx, uid)
	if err != nil || !secrets.Is2faEnabled.Bool {
		return Err2FANotEnabled
	}

	rawSecret, err := decryptTOTPSecret(secrets.TotpSecret.String, encKey)
	if err != nil {
		return err
	}

	if !totp.Validate(code, rawSecret) {
		return ErrInvalid2FACode
	}

	return s.Queries().Disable2FA(ctx, uid)
}

// RecoverAccount consumes a backup code and returns full auth tokens.
func (s *Service) RecoverAccount(ctx context.Context, req RecoverAccountRequest, encKey string) (*AuthResponse, error) {
	claims, err := s.tokenSvc.Validate2FAChallengeToken(req.LoginToken)
	if err != nil {
		return nil, token.ErrInvalidToken
	}

	uid, err := uuid.Parse(claims.UserID)
	if err != nil {
		return nil, token.ErrInvalidToken
	}

	secrets, err := s.Queries().GetUserSecrets(ctx, uid)
	if err != nil || !secrets.Is2faEnabled.Bool {
		return nil, Err2FANotEnabled
	}

	if err := s.validateAndConsumeBackupCode(ctx, uid, req.BackupCode, secrets.BackupCodes); err != nil {
		return nil, ErrInvalidBackup
	}

	user, err := s.Queries().GetUser(ctx, uid)
	if err != nil {
		return nil, ErrUserNotFound
	}

	if claims.ShopID != "" {
		shopID, err := uuid.Parse(claims.ShopID)
		if err == nil {
			shop, sErr := s.Queries().GetShop(ctx, shopID)
			if sErr == nil {
				staff, roleName, rErr := s.getStaffContext(ctx, shopID, uid)
				if rErr == nil {
					return s.generateAuthResponse(user, &shop, roleName, staff.IsOwner.Bool)
				}
			}
		}
	}
	return s.generateAuthResponse(user, nil, "", false)
}

// validateAndConsumeBackupCode checks if the provided plain code matches any hashed
// backup code, removes it from the list, and persists the updated list.
func (s *Service) validateAndConsumeBackupCode(ctx context.Context, userID uuid.UUID, code string, codesJSON []byte) error {
	// Normalise input (strip dashes, uppercase)
	normalised := strings.ToUpper(strings.ReplaceAll(code, "-", ""))

	var hashes []string
	if err := json.Unmarshal(codesJSON, &hashes); err != nil {
		return ErrInvalidBackup
	}

	matchIdx := -1
	for i, h := range hashes {
		plain := strings.ToUpper(strings.ReplaceAll(normalised, "-", ""))
		if bcrypt.CompareHashAndPassword([]byte(h), []byte(plain)) == nil {
			matchIdx = i
			break
		}
	}
	if matchIdx == -1 {
		return ErrInvalidBackup
	}

	// Remove matched code
	hashes = append(hashes[:matchIdx], hashes[matchIdx+1:]...)
	updated, err := json.Marshal(hashes)
	if err != nil {
		return fmt.Errorf("marshal backup codes: %w", err)
	}

	_, err = s.Queries().UpdateUserSecrets(ctx, db.UpdateUserSecretsParams{
		UserID:      userID,
		BackupCodes: updated,
	})
	return err
}
