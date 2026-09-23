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

type Handler struct {
	client  *Client
	queries *db.Queries
	pool    *pgxpool.Pool
	encKey  []byte
	rdb     *redis.Client
	logger  *zap.Logger
}

func NewHandler(client *Client, queries *db.Queries, pool *pgxpool.Pool, encKey []byte, rdb *redis.Client, logger *zap.Logger) *Handler {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &Handler{client: client, queries: queries, pool: pool, encKey: encKey, rdb: rdb, logger: logger}
}

type paymobCredentials struct {
	SubMerchantCode string `json:"sub_merchant_code"`
}

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup, requireAuth, requireStaff gin.HandlerFunc) {
	payments := rg.Group("/payments")
	payments.POST("/initiate", h.Initiate)
	payments.GET("/paymob-settings", requireAuth, requireStaff, h.GetPaymobSettings)
	payments.PATCH("/paymob-settings", requireAuth, requireStaff, h.UpdatePaymobSettings)
}

func (h *Handler) RegisterPublicRoutes(rg *gin.RouterGroup) {
	payments := rg.Group("/payments")
	payments.POST("/webhook", h.Webhook)
}

type InitiateRequest struct {
	OrderID  string      `json:"order_id" binding:"required"`
	Billing  BillingData `json:"billing"`
	Currency string      `json:"currency"`
}

type InitiateResponse struct {
	PaymentURL string `json:"payment_url"`
	OrderID    string `json:"order_id"`
}

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

	billing := fillBillingDefaults(req.Billing)

	orderUUID, err := uuid.Parse(req.OrderID)
	if err != nil {
		response.BadRequest(c, "invalid order_id")
		return
	}
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

	amountCents := int64(0)
	if order.TotalPrice.Valid {
		rawTotalVal, valueErr := order.TotalPrice.Value()
		if valueErr != nil {
			response.BadRequest(c, "invalid order total")
			return
		}
		priceStr, ok := rawTotalVal.(string)
		if !ok || priceStr == "" {
			response.BadRequest(c, "invalid order total")
			return
		}
		dec, parseErr := decimal.NewFromString(priceStr)
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

type PaymobSettingsRequest struct {
	IsEnabled       bool   `json:"is_enabled"`
	SubMerchantCode string `json:"sub_merchant_code"`
}

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

type WebhookPayload struct {
	Obj  TransactionObj `json:"obj"`
	HMAC string         `json:"hmac"`
}

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

type PaymobOrder struct {
	ID              int64  `json:"id"`
	MerchantOrderID string `json:"merchant_order_id"`
}

type SourceData struct {
	Pan     string `json:"pan"`
	Type    string `json:"type"`
	SubType string `json:"sub_type"`
}

func (h *Handler) Webhook(c *gin.Context) {
	var payload WebhookPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	if !h.validateHMAC(payload) {
		response.Forbidden(c, "HMAC validation failed")
		return
	}

	if !payload.Obj.Success || payload.Obj.Pending {
		c.JSON(http.StatusOK, gin.H{"status": "ignored"})
		return
	}

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

	var shopUUID uuid.UUID
	err = h.pool.QueryRow(c.Request.Context(),
		`SELECT shop_id FROM orders WHERE id = $1`, orderUUID,
	).Scan(&shopUUID)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"status": "order not found"})
		return
	}

	_, err = h.pool.Exec(c.Request.Context(),
		`UPDATE orders SET status = 'paid' WHERE id = $1 AND shop_id = $2 AND status = 'pending'`,
		orderUUID,
		shopUUID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update order status"})
		return
	}

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
			eventBytes, err := json.Marshal(event)
			if err != nil {
				h.logger.Warn("failed to marshal payment analytics event", zap.Error(err))
				return
			}
			if err := h.rdb.XAdd(publishCtx, &redis.XAddArgs{
				Stream: fmt.Sprintf("events:analytics:%s", shopIDStr),
				MaxLen: 50_000,
				Approx: true,
				Values: map[string]any{"data": string(eventBytes)},
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

// Paymob concatenates specific fields in a fixed order and signs with HMAC-SHA512.
func (h *Handler) validateHMAC(payload WebhookPayload) bool {
	secret := h.client.HMACSecret()
	if secret == "" {
		return false
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
