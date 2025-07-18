package paymob

import (
	"context"
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"backend-api/internal/db"
	sharedanalytics "backend-api/internal/shared/analytics"
	"backend-api/internal/shared/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

const maxPaymobAmountCents int64 = 100_000_000 // 1,000,000.00 EGP

// Handler handles Paymob payment initiation and webhook callbacks.
type Handler struct {
	client  *Client
	queries *db.Queries
	pool    *pgxpool.Pool
	encKey  []byte        // AES-256 key for encrypting shop credentials at rest
	rdb     *redis.Client // optional — for analytics event publishing
	logger  *zap.Logger
}

// encKey is the AES-256 encryption key for credentials at rest (pass nil/empty to
// skip encryption in development).
func NewHandler(client *Client, queries *db.Queries, pool *pgxpool.Pool, encKey []byte, rdb *redis.Client, logger *zap.Logger) *Handler {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &Handler{client: client, queries: queries, pool: pool, encKey: encKey, rdb: rdb, logger: logger}
}

// paymobCredentials is the JSON stored in shop_payment_methods.encrypted_credentials
// for the Paymob provider. sub_merchant_code is assigned by Paymob when a merchant
// is registered as a sub-merchant under the platform's master account.
type paymobCredentials struct {
	SubMerchantCode string `json:"sub_merchant_code"`
}

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup, requireAuth, requireStaff gin.HandlerFunc) {
	payments := rg.Group("/payments")
	payments.POST("/initiate", h.Initiate)
	payments.GET("/paymob-settings", requireAuth, requireStaff, h.GetPaymobSettings)
	payments.PATCH("/paymob-settings", requireAuth, requireStaff, h.UpdatePaymobSettings)
}

// tenant header — used for server-to-server callbacks from Paymob.
func (h *Handler) RegisterPublicRoutes(rg *gin.RouterGroup) {
	payments := rg.Group("/payments")
	payments.POST("/webhook", h.Webhook)
}

// InitiateRequest is the body sent by the frontend to start a Paymob payment.
// ShopID is NOT required in the body — it is read from the trusted X-Shop-ID
// context set by TenantMiddleware (injected by the gateway).
type InitiateRequest struct {
	OrderID  string      `json:"order_id" binding:"required"`
	Billing  BillingData `json:"billing"`
	Currency string      `json:"currency"` // defaults to "EGP"
}

// InitiateResponse contains the Paymob redirect URL.
type InitiateResponse struct {
	PaymentURL string `json:"payment_url"`
	OrderID    string `json:"order_id"`
}

// Initiate runs the 3-step Paymob flow and returns the iFrame redirect URL.
func (h *Handler) Initiate(c *gin.Context) {
	var req InitiateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	currency := req.Currency
	if currency == "" {
		currency = "EGP"
	}

	// Hydrate billing defaults to avoid Paymob validation errors
	billing := fillBillingDefaults(req.Billing)

	// Look up the order to get its total price
	orderUUID, err := uuid.Parse(req.OrderID)
	if err != nil {
		response.BadRequest(c, "invalid order_id")
		return
	}
	// Read shop_id from the trusted tenant context (set by TenantMiddleware from gateway).
	shopIDStr := c.GetString("shop_id")
	shopUUID, err := uuid.Parse(shopIDStr)
	if err != nil {
		response.BadRequest(c, "missing or invalid shop context")
		return
	}

	order, err := h.queries.GetOrder(c.Request.Context(), db.GetOrderParams{
		ID:     orderUUID,
		ShopID: shopUUID,
	})
	if err != nil {
		response.NotFound(c, "order not found")
		return
	}

	// Convert total_price to cents with exact decimal arithmetic.
	amountCents := int64(0)
	if order.TotalPrice.Valid {
		v, valueErr := order.TotalPrice.Value()
		if valueErr != nil {
			response.BadRequest(c, "invalid order total")
			return
		}
		s, ok := v.(string)
		if !ok || s == "" {
			response.BadRequest(c, "invalid order total")
			return
		}
		dec, parseErr := decimal.NewFromString(s)
		if parseErr != nil {
			response.BadRequest(c, "invalid order total")
			return
		}
		amountDec := dec.Mul(decimal.NewFromInt(100)).Round(0)
		if !amountDec.IsInteger() {
			response.BadRequest(c, "invalid order total")
			return
		}
		amountCents = amountDec.IntPart()
	}
	if amountCents <= 0 {
		response.BadRequest(c, "order total must be greater than zero")
		return
	}
	if amountCents > maxPaymobAmountCents {
		response.BadRequest(c, "order total exceeds maximum allowed value")
		return
	}

	token, err := h.client.Authenticate()
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": fmt.Sprintf("Paymob auth failed: %v", err)})
		return
	}

	paymobOrderID, err := h.client.RegisterOrder(token, amountCents, currency, req.OrderID)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": fmt.Sprintf("Paymob order registration failed: %v", err)})
		return
	}

	// If the shop has a Paymob payment method with a sub_merchant_code configured,
	// compute the split: merchant gets (total - platform_fee), platform keeps the fee.
	var subMerchantCode string
	var subMerchantAmountCents int64

	pm, pmErr := h.queries.GetShopPaymentMethodByProvider(c.Request.Context(), db.GetShopPaymentMethodByProviderParams{
		ShopID:   pgtype.UUID{Bytes: shopUUID, Valid: true},
		Provider: db.PaymentProviderPaymob,
	})
	if pmErr == nil && pm.IsEnabled.Bool {
		rawCreds, _ := decryptCredentials(h.encKey, pm.EncryptedCredentials)
		var creds paymobCredentials
		if jsonErr := json.Unmarshal([]byte(rawCreds), &creds); jsonErr != nil {
			h.logger.Warn("failed to parse paymob sub-merchant credentials",
				zap.String("shop_id", shopUUID.String()),
				zap.Error(jsonErr),
			)
		} else if creds.SubMerchantCode != "" {
			subMerchantCode = creds.SubMerchantCode

			// Get plan transaction fee percent for this shop
			var feePercent float64
			if feeErr := h.pool.QueryRow(c.Request.Context(),
				`SELECT p.transaction_fee_percent FROM shops s JOIN plans p ON p.id = s.plan_id WHERE s.id = $1`,
				shopUUID,
			).Scan(&feePercent); feeErr == nil {
				platformFeeCents := decimal.NewFromInt(amountCents).Mul(decimal.NewFromFloat(feePercent)).Div(decimal.NewFromInt(100)).Round(0).IntPart()
				subMerchantAmountCents = amountCents - platformFeeCents
			}
		}
	}

	paymentKey, err := h.client.GetPaymentKey(token, paymobOrderID, amountCents, currency, billing, subMerchantCode, subMerchantAmountCents)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": fmt.Sprintf("Paymob payment key failed: %v", err)})
		return
	}

	response.OK(c, InitiateResponse{
		PaymentURL: h.client.IFrameURL(paymentKey),
		OrderID:    req.OrderID,
	})
}

// PaymobSettingsResponse is returned by GET /api/v1/payments/paymob-settings.
type PaymobSettingsResponse struct {
	IsEnabled       bool   `json:"is_enabled"`
	SubMerchantCode string `json:"sub_merchant_code"`
}

// PaymobSettingsRequest is the body for PATCH /api/v1/payments/paymob-settings.
type PaymobSettingsRequest struct {
	IsEnabled       bool   `json:"is_enabled"`
	SubMerchantCode string `json:"sub_merchant_code"`
}

// GetPaymobSettings returns the shop's Paymob sub-merchant settings.
func (h *Handler) GetPaymobSettings(c *gin.Context) {
	shopIDStr := c.GetString("shop_id")
	shopUUID, err := uuid.Parse(shopIDStr)
	if err != nil {
		response.BadRequest(c, "missing or invalid shop context")
		return
	}

	pm, err := h.queries.GetShopPaymentMethodByProvider(c.Request.Context(), db.GetShopPaymentMethodByProviderParams{
		ShopID:   pgtype.UUID{Bytes: shopUUID, Valid: true},
		Provider: db.PaymentProviderPaymob,
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			// No row yet — return defaults
			response.OK(c, PaymobSettingsResponse{IsEnabled: false, SubMerchantCode: ""})
			return
		}
		response.InternalError(c)
		return
	}

	rawCreds, _ := decryptCredentials(h.encKey, pm.EncryptedCredentials)
	var creds paymobCredentials
	if err := json.Unmarshal([]byte(rawCreds), &creds); err != nil {
		h.logger.Warn("failed to parse paymob settings credentials",
			zap.String("shop_id", shopUUID.String()),
			zap.Error(err),
		)
	}

	response.OK(c, PaymobSettingsResponse{
		IsEnabled:       pm.IsEnabled.Bool,
		SubMerchantCode: creds.SubMerchantCode,
	})
}

// UpdatePaymobSettings upserts the shop's Paymob sub-merchant settings.
func (h *Handler) UpdatePaymobSettings(c *gin.Context) {
	shopIDStr := c.GetString("shop_id")
	shopUUID, err := uuid.Parse(shopIDStr)
	if err != nil {
		response.BadRequest(c, "missing or invalid shop context")
		return
	}

	var req PaymobSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	creds := paymobCredentials{SubMerchantCode: req.SubMerchantCode}
	credsJSON, _ := json.Marshal(creds)

	storedCreds, err := encryptCredentials(h.encKey, string(credsJSON))
	if err != nil {
		response.InternalError(c)
		return
	}

	_, err = h.pool.Exec(c.Request.Context(),
		`INSERT INTO shop_payment_methods (shop_id, provider, is_enabled, encrypted_credentials)
		 VALUES ($1, 'paymob', $2, $3)
		 ON CONFLICT (shop_id, provider) DO UPDATE SET
			 is_enabled = EXCLUDED.is_enabled,
			 encrypted_credentials = EXCLUDED.encrypted_credentials`,
		shopUUID,
		req.IsEnabled,
		storedCreds,
	)
	if err != nil {
		response.InternalError(c)
		return
	}

	response.OK(c, PaymobSettingsResponse{
		IsEnabled:       req.IsEnabled,
		SubMerchantCode: req.SubMerchantCode,
	})
}

// WebhookPayload is the root object Paymob sends to the callback URL.
type WebhookPayload struct {
	Obj  TransactionObj `json:"obj"`
	HMAC string         `json:"hmac"`
}

// TransactionObj holds transaction details from the Paymob callback.
type TransactionObj struct {
	ID                   int64       `json:"id"`
	Pending              bool        `json:"pending"`
	AmountCents          int64       `json:"amount_cents"`
	Success              bool        `json:"success"`
	IsAuth               bool        `json:"is_auth"`
	IsCapture            bool        `json:"is_capture"`
	IsStandalonePayment  bool        `json:"is_standalone_payment"`
	IsVoided             bool        `json:"is_voided"`
	IsRefunded           bool        `json:"is_refunded"`
	Is3DSecure           bool        `json:"is_3d_secure"`
	IntegrationID        int64       `json:"integration_id"`
	HasParentTransaction bool        `json:"has_parent_transaction"`
	Order                PaymobOrder `json:"order"`
	CreatedAt            string      `json:"created_at"`
	Currency             string      `json:"currency"`
	ErrorOccured         bool        `json:"error_occured"`
	Owner                int64       `json:"owner"`
	SourceData           SourceData  `json:"source_data"`
}

// PaymobOrder is the embedded order info in the transaction callback.
type PaymobOrder struct {
	ID              int64  `json:"id"`
	MerchantOrderID string `json:"merchant_order_id"`
}

// SourceData contains payment instrument details.
type SourceData struct {
	Pan     string `json:"pan"`
	Type    string `json:"type"`
	SubType string `json:"sub_type"`
}

// Webhook processes Paymob transaction callbacks, validates HMAC, and updates order status.
func (h *Handler) Webhook(c *gin.Context) {
	var payload WebhookPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	// Validate HMAC
	if !h.validateHMAC(payload) {
		response.Forbidden(c, "HMAC validation failed")
		return
	}

	// Only act on successful, non-pending transactions
	if !payload.Obj.Success || payload.Obj.Pending {
		c.JSON(http.StatusOK, gin.H{"status": "ignored"})
		return
	}

	// Parse the merchant_order_id back to our internal order UUID
	merchantOrderID := payload.Obj.Order.MerchantOrderID
	if merchantOrderID == "" {
		c.JSON(http.StatusOK, gin.H{"status": "no merchant order id"})
		return
	}

	orderUUID, err := uuid.Parse(merchantOrderID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid merchant_order_id"})
		return
	}

	// Resolve shop_id from the order so the update is tenant-scoped.
	var shopUUID uuid.UUID
	err = h.pool.QueryRow(c.Request.Context(),
		`SELECT shop_id FROM orders WHERE id = $1`, orderUUID,
	).Scan(&shopUUID)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"status": "order not found"})
		return
	}

	// Update order status to "paid", scoped to the resolved shop.
	_, err = h.pool.Exec(c.Request.Context(),
		`UPDATE orders SET status = 'paid' WHERE id = $1 AND shop_id = $2 AND status = 'pending'`,
		orderUUID,
		shopUUID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update order status"})
		return
	}

	// Publish analytics payment_success event (fire-and-forget).
	if h.rdb != nil {
		shopIDStr := shopUUID.String()
		go func() {
			publishCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			event := sharedanalytics.Event{
				Event:  "payment_success",
				ShopID: shopIDStr,
				Payload: map[string]any{
					"order_id":     orderUUID.String(),
					"amount_cents": int64(payload.Obj.AmountCents),
				},
			}
			data, err := json.Marshal(event)
			if err != nil {
				h.logger.Warn("failed to marshal payment analytics event", zap.Error(err))
				return
			}
			if err := h.rdb.XAdd(publishCtx, &redis.XAddArgs{
				Stream: fmt.Sprintf("events:analytics:%s", shopIDStr),
				MaxLen: 50_000,
				Approx: true,
				Values: map[string]any{"data": string(data)},
			}).Err(); err != nil {
				h.logger.Warn("failed to publish payment analytics event",
					zap.String("shop_id", shopIDStr),
					zap.String("order_id", orderUUID.String()),
					zap.Error(err),
				)
			}
		}()
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// validateHMAC verifies the Paymob transaction callback HMAC-SHA512 signature.
// Paymob concatenates specific fields in a fixed order and signs with HMAC-SHA512.
func (h *Handler) validateHMAC(payload WebhookPayload) bool {
	secret := h.client.HMACSecret()
	if secret == "" {
		return false // no secret configured → reject all unauthenticated callbacks
	}

	t := payload.Obj
	concat := fmt.Sprintf("%d%s%s%s%s%d%d%s%s%s%s%s%s%d%d%s%s%s%s%s",
		t.AmountCents,
		t.CreatedAt,
		t.Currency,
		boolStr(t.ErrorOccured),
		boolStr(t.HasParentTransaction),
		t.ID,
		t.IntegrationID,
		boolStr(t.Is3DSecure),
		boolStr(t.IsAuth),
		boolStr(t.IsCapture),
		boolStr(t.IsRefunded),
		boolStr(t.IsStandalonePayment),
		boolStr(t.IsVoided),
		t.Order.ID,
		t.Owner,
		boolStr(t.Pending),
		t.SourceData.Pan,
		t.SourceData.SubType,
		t.SourceData.Type,
		boolStr(t.Success),
	)

	mac := hmac.New(sha512.New, []byte(secret))
	mac.Write([]byte(concat))
	expected := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(payload.HMAC), []byte(expected))
}

func boolStr(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

// fillBillingDefaults replaces empty billing fields with Paymob-required placeholder values.
func fillBillingDefaults(b BillingData) BillingData {
	if b.FirstName == "" {
		b.FirstName = "NA"
	}
	if b.LastName == "" {
		b.LastName = "NA"
	}
	if b.Email == "" {
		b.Email = "na@na.com"
	}
	if b.Phone == "" {
		b.Phone = "+201000000000"
	}
	if b.City == "" {
		b.City = "NA"
	}
	if b.Country == "" {
		b.Country = "EG"
	}
	if b.Street == "" {
		b.Street = "NA"
	}
	if b.Building == "" {
		b.Building = "NA"
	}
	if b.Floor == "" {
		b.Floor = "NA"
	}
	if b.Apartment == "" {
		b.Apartment = "NA"
	}
	if b.PostalCode == "" {
		b.PostalCode = "NA"
	}
	if b.State == "" {
		b.State = "NA"
	}
	return b
}
