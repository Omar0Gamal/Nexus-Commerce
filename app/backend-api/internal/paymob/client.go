// Package paymob wraps the Paymob Accept payment gateway API.
// Flow: Auth → Register Order → Payment Key → Redirect to iFrame
package paymob

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const baseURL = "https://accept.paymob.com/api"

// Client is a thin HTTP wrapper around the Paymob Accept API.
type Client struct {
	apiKey        string
	integrationID int
	iframeID      int
	hmacSecret    string
	httpClient    *http.Client
}

func NewClient(apiKey string, integrationID, iframeID int, hmacSecret string) *Client {
	return &Client{
		apiKey:        apiKey,
		integrationID: integrationID,
		iframeID:      iframeID,
		hmacSecret:    hmacSecret,
		httpClient:    &http.Client{Timeout: 30 * time.Second},
	}
}


type authRequest struct {
	APIKey string `json:"api_key"`
}

type authResponse struct {
	Token string `json:"token"`
}

// Authenticate obtains a short-lived API token from Paymob.
func (c *Client) Authenticate() (string, error) {
	body, _ := json.Marshal(authRequest{APIKey: c.apiKey})
	resp, err := c.post(baseURL+"/auth/tokens", body, "")
	if err != nil {
		return "", fmt.Errorf("paymob auth: %w", err)
	}

	var result authResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return "", fmt.Errorf("paymob auth decode: %w", err)
	}
	if result.Token == "" {
		return "", fmt.Errorf("paymob auth: empty token returned")
	}
	return result.Token, nil
}


type registerOrderRequest struct {
	AuthToken       string        `json:"auth_token"`
	DeliveryNeeded  bool          `json:"delivery_needed"`
	AmountCents     int64         `json:"amount_cents"`
	Currency        string        `json:"currency"`
	Items           []interface{} `json:"items"`
	MerchantOrderID string        `json:"merchant_order_id,omitempty"`
}

type registerOrderResponse struct {
	ID int64 `json:"id"`
}

// RegisterOrder creates a Paymob order record and returns the Paymob order ID.
func (c *Client) RegisterOrder(token string, amountCents int64, currency, merchantOrderID string) (int64, error) {
	body, _ := json.Marshal(registerOrderRequest{
		AuthToken:       token,
		DeliveryNeeded:  false,
		AmountCents:     amountCents,
		Currency:        currency,
		Items:           []interface{}{},
		MerchantOrderID: merchantOrderID,
	})
	resp, err := c.post(baseURL+"/ecommerce/orders", body, "")
	if err != nil {
		return 0, fmt.Errorf("paymob register order: %w", err)
	}

	var result registerOrderResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return 0, fmt.Errorf("paymob register order decode: %w", err)
	}
	return result.ID, nil
}


// BillingData holds customer billing information required by Paymob.
type BillingData struct {
	FirstName  string `json:"first_name"`
	LastName   string `json:"last_name"`
	Email      string `json:"email"`
	Phone      string `json:"phone"`
	City       string `json:"city"`
	Country    string `json:"country"`
	Street     string `json:"street"`
	Building   string `json:"building"`
	Floor      string `json:"floor"`
	Apartment  string `json:"apartment"`
	PostalCode string `json:"postal_code"`
	State      string `json:"state"`
}

// SubMerchant defines a split-payment destination for Paymob's marketplace flow.
// The platform receives the difference between the order total and sub-merchant amount.
type SubMerchant struct {
	AmountCents     int64  `json:"amount_cents"`
	Description     string `json:"description"`
	SubMerchantCode string `json:"sub_merchant_code"`
}

type paymentKeyRequest struct {
	AuthToken     string        `json:"auth_token"`
	AmountCents   int64         `json:"amount_cents"`
	Expiration    int           `json:"expiration"`
	OrderID       int64         `json:"order_id"`
	BillingData   BillingData   `json:"billing_data"`
	Currency      string        `json:"currency"`
	IntegrationID int           `json:"integration_id"`
	SubMerchants  []SubMerchant `json:"sub_merchants,omitempty"`
}

type paymentKeyResponse struct {
	Token string `json:"token"`
}

// GetPaymentKey retrieves a Paymob payment key (used to load the iFrame).
// If subMerchantCode is non-empty, a sub_merchants split entry is added so Paymob
// automatically routes subMerchantAmountCents to the merchant; the remainder stays
// with the platform as the transaction fee.
func (c *Client) GetPaymentKey(token string, paymobOrderID, amountCents int64, currency string, billing BillingData, subMerchantCode string, subMerchantAmountCents int64) (string, error) {
	req := paymentKeyRequest{
		AuthToken:     token,
		AmountCents:   amountCents,
		Expiration:    3600,
		OrderID:       paymobOrderID,
		BillingData:   billing,
		Currency:      currency,
		IntegrationID: c.integrationID,
	}
	if subMerchantCode != "" && subMerchantAmountCents > 0 {
		req.SubMerchants = []SubMerchant{
			{
				AmountCents:     subMerchantAmountCents,
				Description:     "Merchant payout",
				SubMerchantCode: subMerchantCode,
			},
		}
	}
	body, _ := json.Marshal(req)
	resp, err := c.post(baseURL+"/acceptance/payment_keys", body, "")
	if err != nil {
		return "", fmt.Errorf("paymob payment key: %w", err)
	}

	var result paymentKeyResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return "", fmt.Errorf("paymob payment key decode: %w", err)
	}
	if result.Token == "" {
		return "", fmt.Errorf("paymob payment key: empty token")
	}
	return result.Token, nil
}

// IFrameURL returns the redirect URL for the Paymob-hosted payment iFrame.
func (c *Client) IFrameURL(paymentKey string) string {
	return fmt.Sprintf("https://accept.paymob.com/api/acceptance/iframes/%d?payment_token=%s", c.iframeID, paymentKey)
}

// HMACSecret returns the HMAC secret for webhook validation.
func (c *Client) HMACSecret() string {
	return c.hmacSecret
}


type refundRequest struct {
	AuthToken     string `json:"auth_token"`
	TransactionID string `json:"transaction_id"`
	AmountCents   int64  `json:"amount_cents"`
}

type refundResponse struct {
	ID      int64  `json:"id"`
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

// Refund issues a full or partial refund for a transaction through Paymob.
// transactionID is the Paymob transaction ID (stored in the payments table).
// amountCents is the refund amount in the smallest currency unit.
// Returns the Paymob refund transaction ID.
func (c *Client) Refund(transactionID string, amountCents int64) (int64, error) {
	token, err := c.Authenticate()
	if err != nil {
		return 0, fmt.Errorf("paymob refund auth: %w", err)
	}

	body, _ := json.Marshal(refundRequest{
		AuthToken:     token,
		TransactionID: transactionID,
		AmountCents:   amountCents,
	})

	resp, err := c.post(baseURL+"/acceptance/void_refund/refund", body, "")
	if err != nil {
		return 0, fmt.Errorf("paymob refund: %w", err)
	}

	var result refundResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return 0, fmt.Errorf("paymob refund decode: %w", err)
	}
	if !result.Success {
		return 0, fmt.Errorf("paymob refund failed: %s", result.Message)
	}
	return result.ID, nil
}


func (c *Client) post(url string, body []byte, _ string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("paymob API error %d: %s", resp.StatusCode, string(data))
	}
	return data, nil
}
