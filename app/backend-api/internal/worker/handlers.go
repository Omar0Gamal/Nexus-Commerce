package worker

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"backend-api/internal/email"
	"backend-api/internal/storage"
	"backend-api/internal/db"
	"github.com/jung-kurt/gofpdf"
)


type OrderConfirmationPayload struct {
	To   string          `json:"to"`
	Data email.OrderData `json:"data"`
}

type OrderStatusUpdatePayload struct {
	To   string          `json:"to"`
	Data email.OrderData `json:"data"`
}

type LowStockAlertPayload struct {
	To   string             `json:"to"`
	Data email.LowStockData `json:"data"`
}

type PasswordResetPayload struct {
	To   string                  `json:"to"`
	Data email.PasswordResetData `json:"data"`
}

type AccountSuspendedPayload struct {
	To   string                     `json:"to"`
	Data email.AccountSuspendedData `json:"data"`
}

type WebhookDispatchPayload struct {
	TargetURL string         `json:"target_url"`
	Topic     string         `json:"topic"`
	SecretKey string         `json:"secret_key"`
	Body      map[string]any `json:"body"`
}

type InvoiceGeneratePayload struct {
	ShopID  string `json:"shop_id"`
	OrderID string `json:"order_id"`
}

type BackInStockPayload struct {
	To           string `json:"to"`
	CustomerName string `json:"customer_name"`
	ProductTitle string `json:"product_title"`
	ProductURL   string `json:"product_url"`
}

type ReviewRequestItem struct {
	ProductTitle string `json:"product_title"`
	ReviewURL    string `json:"review_url"`
}

type ReviewRequestPayload struct {
	To           string              `json:"to"`
	CustomerName string              `json:"customer_name"`
	Items        []ReviewRequestItem `json:"items"`
}

func EmailHandlers(mailer *email.Mailer) map[string]HandlerFunc {
	return map[string]HandlerFunc{
		JobSendWinback: func(ctx context.Context, raw json.RawMessage) error {
			var p WinbackPayload
			if err := json.Unmarshal(raw, &p); err != nil {
				return fmt.Errorf("unmarshal: %w", err)
			}
			return mailer.SendWinbackEmail(p.To, email.WinbackData{CouponCode: p.CouponCode})
		},
		JobSendOrderConfirmation: func(ctx context.Context, raw json.RawMessage) error {
			var p OrderConfirmationPayload
			if err := json.Unmarshal(raw, &p); err != nil {
				return fmt.Errorf("unmarshal: %w", err)
			}
			return mailer.SendOrderConfirmation(p.To, p.Data)
		},
		JobSendOrderStatusUpdate: func(ctx context.Context, raw json.RawMessage) error {
			var p OrderStatusUpdatePayload
			if err := json.Unmarshal(raw, &p); err != nil {
				return fmt.Errorf("unmarshal: %w", err)
			}
			return mailer.SendOrderStatusUpdate(p.To, p.Data)
		},
		JobSendLowStockAlert: func(ctx context.Context, raw json.RawMessage) error {
			var p LowStockAlertPayload
			if err := json.Unmarshal(raw, &p); err != nil {
				return fmt.Errorf("unmarshal: %w", err)
			}
			return mailer.SendLowStockAlert(p.To, p.Data)
		},
		JobSendPasswordReset: func(ctx context.Context, raw json.RawMessage) error {
			var p PasswordResetPayload
			if err := json.Unmarshal(raw, &p); err != nil {
				return fmt.Errorf("unmarshal: %w", err)
			}
			return mailer.SendPasswordReset(p.To, p.Data)
		},
		JobSendAccountSuspended: func(ctx context.Context, raw json.RawMessage) error {
			var p AccountSuspendedPayload
			if err := json.Unmarshal(raw, &p); err != nil {
				return fmt.Errorf("unmarshal: %w", err)
			}
			return mailer.SendAccountSuspended(p.To, p.Data)
		},
		JobSendBackInStock: func(ctx context.Context, raw json.RawMessage) error {
			var p BackInStockPayload
			if err := json.Unmarshal(raw, &p); err != nil {
				return fmt.Errorf("unmarshal: %w", err)
			}
			return mailer.SendBackInStock(p.To, email.BackInStockData{
				CustomerName: p.CustomerName,
				ProductTitle: p.ProductTitle,
				ProductURL:   p.ProductURL,
			})
		},
		JobSendReviewRequest: func(ctx context.Context, raw json.RawMessage) error {
			var p ReviewRequestPayload
			if err := json.Unmarshal(raw, &p); err != nil {
				return fmt.Errorf("unmarshal: %w", err)
			}
			items := make([]email.ReviewRequestItem, 0, len(p.Items))
			for _, it := range p.Items {
				items = append(items, email.ReviewRequestItem{
					ProductTitle: it.ProductTitle,
					ReviewURL:    it.ReviewURL,
				})
			}
			return mailer.SendReviewRequest(p.To, email.ReviewRequestData{
				CustomerName: p.CustomerName,
				Items:        items,
			})
		},
	}
}

func WebhookHandler() HandlerFunc {
	httpClient := &http.Client{Timeout: 10 * time.Second}

	return func(ctx context.Context, raw json.RawMessage) error {
		var p WebhookDispatchPayload
		if err := json.Unmarshal(raw, &p); err != nil {
			return fmt.Errorf("unmarshal: %w", err)
		}

		body, err := json.Marshal(p.Body)
		if err != nil {
			return fmt.Errorf("marshal body: %w", err)
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.TargetURL, bytes.NewReader(body))
		if err != nil {
			return fmt.Errorf("build request: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Webhook-Topic", p.Topic)

		if p.SecretKey != "" {
			mac := hmac.New(sha256.New, []byte(p.SecretKey))
			mac.Write(body)
			req.Header.Set("X-Webhook-Signature", "sha256="+hex.EncodeToString(mac.Sum(nil)))
		}

		resp, err := httpClient.Do(req)
		if err != nil {
			return fmt.Errorf("deliver webhook: %w", err)
		}
		resp.Body.Close()

		if resp.StatusCode >= 500 {
			return fmt.Errorf("webhook endpoint returned %d", resp.StatusCode)
		}
		return nil
	}
}

func InvoiceHandler(storageClient *storage.Client, q *db.Queries) HandlerFunc {
	return func(ctx context.Context, raw json.RawMessage) error {
		var p InvoiceGeneratePayload
		if err := json.Unmarshal(raw, &p); err != nil {
			return fmt.Errorf("unmarshal: %w", err)
		}
		
		pdf := gofpdf.New("P", "mm", "A4", "")
		pdf.AddPage()
		pdf.SetFont("Arial", "B", 16)
		pdf.Cell(40, 10, fmt.Sprintf("Invoice for Order %s", p.OrderID))
		pdf.Ln(10)
		pdf.SetFont("Arial", "", 12)
		pdf.Cell(40, 10, fmt.Sprintf("Shop ID: %s", p.ShopID))
		
		var buf bytes.Buffer
		if err := pdf.Output(&buf); err != nil {
			return fmt.Errorf("generate pdf: %w", err)
		}
		
		if storageClient != nil {
			key := fmt.Sprintf("invoices/%s/%s.pdf", p.ShopID, p.OrderID)
			_, err := storageClient.Upload(ctx, key, "application/pdf", &buf)
			if err != nil {
				return fmt.Errorf("upload pdf: %w", err)
			}
		}
		return nil
	}
}



const JobSendWinback = "email:winback"

type WinbackPayload struct {
	To         string `json:"to"`
	CouponCode string `json:"coupon_code"`
}
