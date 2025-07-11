package billing

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"backend-api/internal/db"
	"backend-api/internal/email"
	"backend-api/internal/paymob"
	"backend-api/internal/shared/pgutil"

	"github.com/jackc/pgx/v5/pgtype"
	"go.uber.org/zap"
)

// InvoiceWorker runs on the 1st of each month and creates subscription invoices
// for all active shops, then generates a Paymob payment link and emails it to
// the shop owner.
type InvoiceWorker struct {
	db     *db.Queries
	paymob *paymob.Client
	mailer *email.Mailer
	logger *zap.Logger
}

func NewInvoiceWorker(queries *db.Queries, paymobClient *paymob.Client, mailer *email.Mailer, logger *zap.Logger) *InvoiceWorker {
	return &InvoiceWorker{
		db:     queries,
		paymob: paymobClient,
		mailer: mailer,
		logger: logger,
	}
}

// Run blocks until ctx is cancelled, firing invoice generation on the 1st of
// each month at 00:10 UTC. It uses a ticker that checks every minute and fires
// only on the correct day/hour/minute to stay idempotent across restarts.
func (w *InvoiceWorker) Run(ctx context.Context) {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case t := <-ticker.C:
			utc := t.UTC()
			if utc.Day() == 1 && utc.Hour() == 0 && utc.Minute() == 10 {
				w.logger.Info("invoice worker: starting monthly run")
				w.generateMonthlyInvoices(ctx)
			}
		}
	}
}

func (w *InvoiceWorker) generateMonthlyInvoices(ctx context.Context) {
	shops, err := w.db.ListShopsByStatus(ctx, db.NullShopStatus{
		ShopStatus: db.ShopStatusActive,
		Valid:      true,
	})
	if err != nil {
		w.logger.Error("invoice worker: list active shops", zap.Error(err))
		return
	}

	for _, shop := range shops {
		if err := w.processShopInvoice(ctx, shop); err != nil {
			w.logger.Error("invoice worker: process shop",
				zap.String("shop_id", shop.ID.String()),
				zap.Error(err),
			)
		}
	}
}

func (w *InvoiceWorker) processShopInvoice(ctx context.Context, shop db.Shop) error {
	plan, err := w.db.GetPlan(ctx, shop.PlanID)
	if err != nil {
		return fmt.Errorf("get plan: %w", err)
	}

	// Compute amount: plan monthly_price is stored as a Postgres numeric string.
	amount := plan.MonthlyPrice

	// idempotency: skip if invoice already exists for this month.
	now := time.Now().UTC()
	periodStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	invoices, err := w.db.ListShopInvoices(ctx, db.ListShopInvoicesParams{
		ShopID: pgtype.UUID{Bytes: shop.ID, Valid: true},
		Limit:  1,
		Offset: 0,
	})
	if err == nil && len(invoices) > 0 {
		latest := invoices[0]
		if latest.CreatedAt.Valid && latest.CreatedAt.Time.After(periodStart) {
			// Invoice for this period already exists.
			return nil
		}
	}

	// Create invoice with "open" status first.
	invoice, err := w.db.CreateShopInvoice(ctx, db.CreateShopInvoiceParams{
		ShopID:           pgtype.UUID{Bytes: shop.ID, Valid: true},
		Amount:           amount,
		Status:           db.InvoiceStatusOpen,
		BillingReason:    pgtype.Text{String: fmt.Sprintf("subscription_%s_%d", now.Month().String(), now.Year()), Valid: true},
		HostedInvoiceUrl: pgtype.Text{},
	})
	if err != nil {
		return fmt.Errorf("create invoice: %w", err)
	}

	// Generate Paymob payment link if configured.
	paymentURL, err := w.generatePaymentLink(shop, plan, invoice.ID.String())
	if err != nil {
		w.logger.Warn("invoice worker: paymob payment link failed, invoice created without URL",
			zap.String("invoice_id", invoice.ID.String()),
			zap.Error(err),
		)
	} else if paymentURL != "" {
		// Update invoice with the payment URL.
		if _, err := w.db.UpdateShopInvoiceStatus(ctx, db.UpdateShopInvoiceStatusParams{
			ID:     invoice.ID,
			Status: db.InvoiceStatusOpen,
		}); err != nil {
			w.logger.Warn("invoice worker: update invoice url", zap.Error(err))
		}
	}

	// Send email to shop owner.
	user, err := w.db.GetUser(ctx, shop.OwnerUserID)
	if err != nil {
		return fmt.Errorf("get owner: %w", err)
	}

	amountStr := pgutil.NumericToString(amount)
	if w.mailer != nil && w.mailer.Enabled() {
		_ = w.mailer.SendInvoiceReady(user.Email, email.InvoiceData{
			ShopName:    shop.Name,
			PlanName:    plan.Name,
			Amount:      amountStr,
			Currency:    "EGP",
			PaymentURL:  paymentURL,
			InvoiceID:   invoice.ID.String(),
			PeriodStart: periodStart.Format("Jan 2, 2006"),
			PeriodEnd:   time.Date(now.Year(), now.Month()+1, 0, 0, 0, 0, 0, time.UTC).Format("Jan 2, 2006"),
		})
	}

	w.logger.Info("invoice worker: invoice created",
		zap.String("shop_id", shop.ID.String()),
		zap.String("invoice_id", invoice.ID.String()),
		zap.String("amount", amountStr),
	)
	return nil
}

func (w *InvoiceWorker) generatePaymentLink(shop db.Shop, plan db.Plan, merchantOrderID string) (string, error) {
	if w.paymob == nil {
		return "", nil
	}

	token, err := w.paymob.Authenticate()
	if err != nil {
		return "", fmt.Errorf("paymob auth: %w", err)
	}

	amountStr := pgutil.NumericToString(plan.MonthlyPrice)
	amountFloat, ok := new(big.Float).SetString(amountStr)
	if !ok {
		return "", fmt.Errorf("invalid plan amount: %s", amountStr)
	}
	amountF64, _ := amountFloat.Float64()
	amountCents := int64(amountF64 * 100)

	currency := "EGP"
	if shop.Currency.Valid && shop.Currency.String != "" {
		currency = shop.Currency.String
	}

	paymobOrderID, err := w.paymob.RegisterOrder(token, amountCents, currency, merchantOrderID)
	if err != nil {
		return "", fmt.Errorf("register order: %w", err)
	}

	paymentKey, err := w.paymob.GetPaymentKey(token, paymobOrderID, amountCents, currency,
		paymob.BillingData{FirstName: shop.Name, LastName: "Invoice"},
		"", 0,
	)
	if err != nil {
		return "", fmt.Errorf("get payment key: %w", err)
	}

	return w.paymob.IFrameURL(paymentKey), nil
}
