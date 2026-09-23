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

func (c *Client) Authenticate() (string, error) {
	body, _ := json.Marshal(authRequest{APIKey: c.apiKey})
	resp, err := c.post(baseURL+"/auth/tokens", body, "")
	if err != nil {
		return "", fmt.Errorf("paymob auth: %w", err)
	}

	var authResp authResponse
	if err := json.Unmarshal(resp, &authResp); err != nil {
		return "", fmt.Errorf("paymob auth decode: %w", err)
	}
	if authResp.Token == "" {
		return "", fmt.Errorf("paymob auth: empty token returned")
	}
	return authResp.Token, nil
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

	var orderResp registerOrderResponse
	if err := json.Unmarshal(resp, &orderResp); err != nil {
		return 0, fmt.Errorf("paymob register order decode: %w", err)
	}
	return orderResp.ID, nil
}


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

	var keyResp paymentKeyResponse
	if err := json.Unmarshal(resp, &keyResp); err != nil {
		return "", fmt.Errorf("paymob payment key decode: %w", err)
	}
	if keyResp.Token == "" {
		return "", fmt.Errorf("paymob payment key: empty token")
	}
	return keyResp.Token, nil
}

func (c *Client) IFrameURL(paymentKey string) string {
	return fmt.Sprintf("https://accept.paymob.com/api/acceptance/iframes/%d?payment_token=%s", c.iframeID, paymentKey)
}

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

	var refundResp refundResponse
	if err := json.Unmarshal(resp, &refundResp); err != nil {
		return 0, fmt.Errorf("paymob refund decode: %w", err)
	}
	if !refundResp.Success {
		return 0, fmt.Errorf("paymob refund failed: %s", refundResp.Message)
	}
	return refundResp.ID, nil
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

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("paymob API error %d: %s", resp.StatusCode, string(respBytes))
	}
	return respBytes, nil
}
