// Package email provides a simple SMTP mailer for transactional emails.
package email

import (
	"bytes"
	"fmt"
	"html/template"
	"net/smtp"
	"strings"

	"go.uber.org/zap"
)

// Mailer sends transactional emails via SMTP.
// When Host is empty the mailer is disabled (all sends are no-ops that log a warning).
type Mailer struct {
	host     string
	port     int
	username string
	password string
	from     string
	logger   *zap.Logger
}

// New creates a Mailer. If host is empty the mailer will be disabled.
func New(host string, port int, username, password, from string, logger *zap.Logger) *Mailer {
	return &Mailer{host: host, port: port, username: username, password: password, from: from, logger: logger}
}

// Enabled reports whether SMTP credentials are configured.
func (m *Mailer) Enabled() bool { return m.host != "" }


// OrderData holds all the information needed to render an order email.
type OrderData struct {
	OrderNumber  int32
	OrderID      string
	TotalPrice   string
	Status       string
	CustomerName string
	Items        []OrderItemData
}

// OrderItemData is a single line item in an order email.
type OrderItemData struct {
	Name     string
	Quantity int32
	Price    string
}

// PasswordResetData holds data for a password-reset email.
type PasswordResetData struct {
	FullName  string
	ResetLink string
	ExpiresIn string // e.g. "30 minutes"
}

// LowStockData holds data for a low-stock alert email.
type LowStockData struct {
	ShopName     string
	ProductTitle string
	ProductID    string
	SKU          string
	CurrentStock int32
	Threshold    int32
}

// AccountSuspendedData holds data for an account-suspended notification.
type AccountSuspendedData struct {
	ShopName     string
	Reason       string
	SupportEmail string
}


// SendOrderConfirmation sends an order confirmation email to the customer.
func (m *Mailer) SendOrderConfirmation(to string, data OrderData) error {
	if !m.Enabled() {
		m.logger.Info("SMTP not configured, skipping order confirmation", zap.String("to", to))
		return nil
	}
	subject := fmt.Sprintf("Order Confirmed — #%d", data.OrderNumber)
	return m.renderAndSend(to, subject, confirmationTmpl, data)
}

// SendOrderStatusUpdate sends an order status-change email to the customer.
func (m *Mailer) SendOrderStatusUpdate(to string, data OrderData) error {
	if !m.Enabled() {
		m.logger.Info("SMTP not configured, skipping order status update", zap.String("to", to))
		return nil
	}
	subject := fmt.Sprintf("Order #%d — Status Updated to %s", data.OrderNumber, capitalize(data.Status))
	return m.renderAndSend(to, subject, statusUpdateTmpl, data)
}

// SendPasswordReset sends a password-reset link to the user.
func (m *Mailer) SendPasswordReset(to string, data PasswordResetData) error {
	if !m.Enabled() {
		m.logger.Info("SMTP not configured, skipping password reset", zap.String("to", to))
		return nil
	}
	return m.renderAndSend(to, "Reset your password", passwordResetTmpl, data)
}

// SendLowStockAlert notifies the shop owner that a product is low in stock.
func (m *Mailer) SendLowStockAlert(to string, data LowStockData) error {
	if !m.Enabled() {
		m.logger.Info("SMTP not configured, skipping low-stock alert", zap.String("to", to))
		return nil
	}
	subject := fmt.Sprintf("Low stock alert: %s", data.ProductTitle)
	return m.renderAndSend(to, subject, lowStockTmpl, data)
}

// SendAccountSuspended notifies that a shop account has been suspended.
func (m *Mailer) SendAccountSuspended(to string, data AccountSuspendedData) error {
	if !m.Enabled() {
		m.logger.Info("SMTP not configured, skipping account suspension notice", zap.String("to", to))
		return nil
	}
	return m.renderAndSend(to, "Your account has been suspended", accountSuspendedTmpl, data)
}


func (m *Mailer) renderAndSend(to, subject, tmplStr string, data interface{}) error {
	body, err := renderTemplate(tmplStr, data)
	if err != nil {
		return fmt.Errorf("render email template: %w", err)
	}

	msg := buildMessage(m.from, to, subject, body)
	auth := smtp.PlainAuth("", m.username, m.password, m.host)
	addr := fmt.Sprintf("%s:%d", m.host, m.port)

	if err := smtp.SendMail(addr, auth, m.from, []string{to}, msg); err != nil {
		return fmt.Errorf("smtp send: %w", err)
	}
	m.logger.Info("email sent", zap.String("subject", subject), zap.String("to", to))
	return nil
}

func buildMessage(from, to, subject, htmlBody string) []byte {
	var b strings.Builder
	b.WriteString("MIME-Version: 1.0\r\n")
	b.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
	fmt.Fprintf(&b, "From: %s\r\n", from)
	fmt.Fprintf(&b, "To: %s\r\n", to)
	fmt.Fprintf(&b, "Subject: %s\r\n", subject)
	b.WriteString("\r\n")
	b.WriteString(htmlBody)
	return []byte(b.String())
}

func renderTemplate(tmplStr string, data interface{}) (string, error) {
	funcMap := template.FuncMap{
		"upper": strings.ToUpper,
	}
	t, err := template.New("email").Funcs(funcMap).Parse(tmplStr)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func capitalize(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}


const baseStyle = `
  body { margin:0; padding:24px; background:#f5f5f5; font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',sans-serif; }
  .card { max-width:560px; margin:0 auto; background:#fff; border-radius:10px; padding:36px; box-shadow:0 1px 3px rgba(0,0,0,.08); }
  h1 { font-size:22px; color:#111; margin:0 0 8px; }
  p { font-size:15px; color:#555; margin:0 0 24px; }
  .divider { border:none; border-top:1px solid #eee; margin:24px 0; }
  .label { font-size:11px; text-transform:uppercase; letter-spacing:.6px; color:#999; margin:0 0 4px; }
  .value { font-size:18px; font-weight:600; color:#111; margin:0 0 24px; }
  .item { display:flex; justify-content:space-between; padding:10px 0; border-bottom:1px solid #f0f0f0; font-size:14px; color:#333; }
  .total-row { display:flex; justify-content:space-between; padding:16px 0 0; font-size:15px; font-weight:700; color:#111; }
  .badge { display:inline-block; padding:4px 12px; border-radius:20px; font-size:13px; font-weight:600; background:#f0fdf4; color:#166534; }
  .footer { margin-top:32px; font-size:12px; color:#aaa; text-align:center; }
`

const confirmationTmpl = `<!DOCTYPE html>
<html lang="en">
<head><meta charset="UTF-8"><meta name="viewport" content="width=device-width,initial-scale=1">
<style>` + baseStyle + `</style></head>
<body>
  <div class="card">
    <h1>🎉 Order Confirmed!</h1>
    <p>Hi {{.CustomerName}}, thank you for your purchase. Your order has been placed successfully.</p>
    <hr class="divider">
    <p class="label">Order number</p>
    <p class="value">#{{.OrderNumber}}</p>

    {{range .Items}}
    <div class="item">
      <span>{{.Name}} &times; {{.Quantity}}</span>
      <span>{{.Price}} EGP</span>
    </div>
    {{end}}
    <div class="total-row">
      <span>Total</span>
      <span>{{.TotalPrice}} EGP</span>
    </div>

    <hr class="divider">
    <p>We'll send you another email when your order ships. If you have any questions, reply to this email.</p>
    <div class="footer">Nexus Commerce &mdash; Powered by your store</div>
  </div>
</body></html>`

const statusUpdateTmpl = `<!DOCTYPE html>
<html lang="en">
<head><meta charset="UTF-8"><meta name="viewport" content="width=device-width,initial-scale=1">
<style>` + baseStyle + `</style></head>
<body>
  <div class="card">
    <h1>Order Update</h1>
    <p>Hi {{.CustomerName}}, your order status has been updated.</p>
    <hr class="divider">
    <p class="label">Order number</p>
    <p class="value">#{{.OrderNumber}}</p>
    <p class="label">New status</p>
    <p><span class="badge">{{.Status}}</span></p>
    <hr class="divider">
    <p>Questions? Reply to this email and we'll be happy to help.</p>
    <div class="footer">Nexus Commerce &mdash; Powered by your store</div>
  </div>
</body></html>`

const passwordResetTmpl = `<!DOCTYPE html>
<html lang="en">
<head><meta charset="UTF-8"><meta name="viewport" content="width=device-width,initial-scale=1">
<style>` + baseStyle + `
  .btn { display:inline-block; padding:12px 28px; background:#111; color:#fff; border-radius:6px; text-decoration:none; font-size:15px; font-weight:600; }
</style></head>
<body>
  <div class="card">
    <h1>Reset your password</h1>
    <p>Hi {{.FullName}}, we received a request to reset the password for your account.</p>
    <p>Click the button below to set a new password. This link expires in <strong>{{.ExpiresIn}}</strong>.</p>
    <p><a class="btn" href="{{.ResetLink}}">Reset Password</a></p>
    <hr class="divider">
    <p style="font-size:13px;color:#999;">If you didn't request a password reset, you can safely ignore this email.</p>
    <div class="footer">Nexus Commerce &mdash; Powered by your store</div>
  </div>
</body></html>`

const lowStockTmpl = `<!DOCTYPE html>
<html lang="en">
<head><meta charset="UTF-8"><meta name="viewport" content="width=device-width,initial-scale=1">
<style>` + baseStyle + `</style></head>
<body>
  <div class="card">
    <h1>⚠️ Low Stock Alert</h1>
    <p>A product in your store <strong>{{.ShopName}}</strong> is running low on stock.</p>
    <hr class="divider">
    <p class="label">Product</p>
    <p class="value">{{.ProductTitle}}</p>
    {{if .SKU}}<p class="label">SKU</p><p style="font-size:14px;color:#555;margin:0 0 16px;">{{.SKU}}</p>{{end}}
    <p class="label">Current stock</p>
    <p class="value" style="color:#dc2626;">{{.CurrentStock}} units</p>
    <p class="label">Low stock threshold</p>
    <p style="font-size:14px;color:#555;margin:0 0 24px;">{{.Threshold}} units</p>
    <hr class="divider">
    <p>Please restock this product to avoid stockouts.</p>
    <div class="footer">Nexus Commerce &mdash; Powered by your store</div>
  </div>
</body></html>`

const accountSuspendedTmpl = `<!DOCTYPE html>
<html lang="en">
<head><meta charset="UTF-8"><meta name="viewport" content="width=device-width,initial-scale=1">
<style>` + baseStyle + `</style></head>
<body>
  <div class="card">
    <h1>Account Suspended</h1>
    <p>Your store <strong>{{.ShopName}}</strong> has been suspended.</p>
    {{if .Reason}}
    <hr class="divider">
    <p class="label">Reason</p>
    <p style="font-size:15px;color:#555;margin:0 0 24px;">{{.Reason}}</p>
    {{end}}
    <hr class="divider">
    <p>If you believe this is a mistake or would like to appeal, please contact us at <a href="mailto:{{.SupportEmail}}">{{.SupportEmail}}</a>.</p>
    <div class="footer">Nexus Commerce &mdash; Powered by your store</div>
  </div>
</body></html>`


// DigestData holds data for a daily analytics digest email.
type DigestData struct {
	ShopName        string
	Date            string
	TodayVisits     int64
	UniqueVisitors  int64
	Revenue         string
	TopPages        []DigestPageEntry
	TopReferrer     string
	IsAnomaly       bool
	AnomalyDeltaPct float64
	Week7AvgVisits  int64
}

// DigestPageEntry is a single page in the digest.
type DigestPageEntry struct {
	Path  string
	Views int64
}

// WeeklyReportData holds data for a weekly trend report email.
type WeeklyReportData struct {
	ShopName         string
	PeriodStart      string
	PeriodEnd        string
	TotalVisits      int64
	UniqueVisitors   int64
	TotalRevenue     string
	WoWVisitChange   float64
	WoWRevenueChange float64
	TopProducts      []ProductReportEntry
	WeeklyTrend      []TrendEntry
}

// ProductReportEntry is a product in the weekly report.
type ProductReportEntry struct {
	Title   string
	Views   int64
	Revenue string
}

// TrendEntry is a single day in the weekly trend.
type TrendEntry struct {
	Date   string
	Visits int64
}

// DepletionAlertData holds data for a stock depletion alert email.
type DepletionAlertData struct {
	ShopName           string
	ProductTitle       string
	ProductID          string
	StockQuantity      int32
	AvgDailySales      float64
	DaysUntilDepletion float64
	EditURL            string
}

// ABTestConcludedData holds data for an A/B test conclusion email.
type ABTestConcludedData struct {
	ShopName      string
	TestName      string
	Winner        string
	ConfidencePct float64
	VariantAConv  float64
	VariantBConv  float64
	TestDays      int
}


// SendDailyDigest sends the daily analytics digest email.
func (m *Mailer) SendDailyDigest(to string, data DigestData) error {
	if !m.Enabled() {
		m.logger.Info("SMTP not configured, skipping daily digest", zap.String("to", to))
		return nil
	}
	subject := fmt.Sprintf("Your daily analytics report — %s", data.Date)
	return m.renderAndSend(to, subject, dailyDigestTmpl, data)
}

// SendWeeklyReport sends the weekly trend report email.
func (m *Mailer) SendWeeklyReport(to string, data WeeklyReportData) error {
	if !m.Enabled() {
		m.logger.Info("SMTP not configured, skipping weekly report", zap.String("to", to))
		return nil
	}
	subject := fmt.Sprintf("Weekly report: %s – %s", data.PeriodStart, data.PeriodEnd)
	return m.renderAndSend(to, subject, weeklyReportTmpl, data)
}

// SendDepletionAlert sends a stock depletion forecast alert.
func (m *Mailer) SendDepletionAlert(to string, data DepletionAlertData) error {
	if !m.Enabled() {
		m.logger.Info("SMTP not configured, skipping depletion alert", zap.String("to", to))
		return nil
	}
	subject := fmt.Sprintf("⚠️ Stock depletion warning: %s", data.ProductTitle)
	return m.renderAndSend(to, subject, depletionAlertTmpl, data)
}

// SendABTestConcluded sends an A/B test conclusion notification.
func (m *Mailer) SendABTestConcluded(to string, data ABTestConcludedData) error {
	if !m.Enabled() {
		m.logger.Info("SMTP not configured, skipping A/B test conclusion", zap.String("to", to))
		return nil
	}
	subject := fmt.Sprintf("A/B test concluded: %s", data.TestName)
	return m.renderAndSend(to, subject, abTestConcludedTmpl, data)
}

// InvoiceData holds the data for an invoice email.
type InvoiceData struct {
	ShopName    string
	PlanName    string
	Amount      string
	Currency    string
	PaymentURL  string
	InvoiceID   string
	PeriodStart string
	PeriodEnd   string
}

// SendInvoiceReady sends a subscription invoice email with a payment link.
func (m *Mailer) SendInvoiceReady(to string, data InvoiceData) error {
	if !m.Enabled() {
		m.logger.Info("SMTP not configured, skipping invoice email", zap.String("to", to))
		return nil
	}
	subject := fmt.Sprintf("Your %s subscription invoice for %s – %s", data.PlanName, data.PeriodStart, data.PeriodEnd)
	return m.renderAndSend(to, subject, invoiceTmpl, data)
}

const invoiceTmpl = `<!DOCTYPE html>
<html lang="en">
<head><meta charset="UTF-8"><meta name="viewport" content="width=device-width,initial-scale=1">
<style>` + baseStyle + `
  .amount { font-size:36px; font-weight:700; color:#111; margin:8px 0 24px; }
  .btn { display:inline-block; background:#2563eb; color:#fff; text-decoration:none; padding:14px 28px; border-radius:8px; font-weight:600; font-size:16px; }
</style></head>
<body>
  <div class="card">
    <h1>💳 Subscription Invoice</h1>
    <p>Hello, here is your invoice for <strong>{{.ShopName}}</strong>.</p>
    <hr class="divider">
    <p class="label">Plan</p>
    <p class="value">{{.PlanName}}</p>
    <p class="label">Billing period</p>
    <p class="value">{{.PeriodStart}} – {{.PeriodEnd}}</p>
    <p class="label">Amount due</p>
    <p class="amount">{{.Amount}} {{.Currency}}</p>
    {{if .PaymentURL}}
    <a href="{{.PaymentURL}}" class="btn">Pay Now →</a>
    {{else}}
    <p style="color:#6b7280;font-size:14px;">No payment link available. Please contact support.</p>
    {{end}}
    <div class="footer">Invoice #{{.InvoiceID}} &mdash; Nexus Commerce</div>
  </div>
</body></html>`


const dailyDigestTmpl = `<!DOCTYPE html>
<html lang="en">
<head><meta charset="UTF-8"><meta name="viewport" content="width=device-width,initial-scale=1">
<style>` + baseStyle + `
  .stat-row { display:flex; gap:24px; margin-bottom:24px; }
  .stat { flex:1; background:#f9fafb; border-radius:8px; padding:16px; }
  .stat .num { font-size:28px; font-weight:700; color:#111; margin:4px 0; }
  .anomaly { background:#fef2f2; border-left:4px solid #ef4444; padding:12px 16px; border-radius:0 8px 8px 0; margin-bottom:24px; }
</style></head>
<body>
  <div class="card">
    <h1>📊 Daily Analytics — {{.Date}}</h1>
    <p>Here's your daily performance summary for <strong>{{.ShopName}}</strong>.</p>
    {{if .IsAnomaly}}
    <div class="anomaly">
      <strong>⚡ Traffic anomaly detected:</strong> Today's visits are {{printf "%.0f" .AnomalyDeltaPct}}% {{if gt .AnomalyDeltaPct 0.0}}above{{else}}below{{end}} your 7-day average ({{.Week7AvgVisits}} visits/day).
    </div>
    {{end}}
    <div class="stat-row">
      <div class="stat">
        <p class="label">Visits today</p>
        <p class="num">{{.TodayVisits}}</p>
      </div>
      <div class="stat">
        <p class="label">Unique visitors</p>
        <p class="num">{{.UniqueVisitors}}</p>
      </div>
      <div class="stat">
        <p class="label">Revenue</p>
        <p class="num">{{.Revenue}}</p>
      </div>
    </div>
    {{if .TopPages}}
    <hr class="divider">
    <p class="label">Top pages</p>
    {{range .TopPages}}
    <div class="item"><span>{{.Path}}</span><span>{{.Views}} views</span></div>
    {{end}}
    {{end}}
    {{if .TopReferrer}}
    <hr class="divider">
    <p class="label">Top referrer</p>
    <p style="font-size:14px;color:#555;margin:0 0 16px;">{{.TopReferrer}}</p>
    {{end}}
    <div class="footer">Nexus Commerce &mdash; Powered by your store</div>
  </div>
</body></html>`

const weeklyReportTmpl = `<!DOCTYPE html>
<html lang="en">
<head><meta charset="UTF-8"><meta name="viewport" content="width=device-width,initial-scale=1">
<style>` + baseStyle + `
  .stat-row { display:flex; gap:16px; margin-bottom:24px; flex-wrap:wrap; }
  .stat { flex:1; min-width:120px; background:#f9fafb; border-radius:8px; padding:16px; }
  .stat .num { font-size:22px; font-weight:700; color:#111; margin:4px 0; }
  .change { font-size:12px; font-weight:600; }
  .up { color:#16a34a; } .down { color:#dc2626; }
</style></head>
<body>
  <div class="card">
    <h1>📈 Weekly Report</h1>
    <p><strong>{{.ShopName}}</strong> — {{.PeriodStart}} to {{.PeriodEnd}}</p>
    <div class="stat-row">
      <div class="stat">
        <p class="label">Total visits</p>
        <p class="num">{{.TotalVisits}}</p>
        <span class="change {{if gt .WoWVisitChange 0.0}}up{{else}}down{{end}}">{{printf "%.1f" .WoWVisitChange}}% WoW</span>
      </div>
      <div class="stat">
        <p class="label">Unique visitors</p>
        <p class="num">{{.UniqueVisitors}}</p>
      </div>
      <div class="stat">
        <p class="label">Revenue</p>
        <p class="num">{{.TotalRevenue}}</p>
        <span class="change {{if gt .WoWRevenueChange 0.0}}up{{else}}down{{end}}">{{printf "%.1f" .WoWRevenueChange}}% WoW</span>
      </div>
    </div>
    {{if .TopProducts}}
    <hr class="divider">
    <p class="label">Top 5 products by views</p>
    {{range .TopProducts}}
    <div class="item"><span>{{.Title}}</span><span>{{.Views}} views · {{.Revenue}}</span></div>
    {{end}}
    {{end}}
    <div class="footer">Nexus Commerce &mdash; Powered by your store</div>
  </div>
</body></html>`

const depletionAlertTmpl = `<!DOCTYPE html>
<html lang="en">
<head><meta charset="UTF-8"><meta name="viewport" content="width=device-width,initial-scale=1">
<style>` + baseStyle + `</style></head>
<body>
  <div class="card">
    <h1>⚠️ Stock Depletion Warning</h1>
    <p>A product in <strong>{{.ShopName}}</strong> is projected to run out of stock soon.</p>
    <hr class="divider">
    <p class="label">Product</p>
    <p class="value">{{.ProductTitle}}</p>
    <p class="label">Current stock</p>
    <p class="value" style="color:#dc2626;">{{.StockQuantity}} units</p>
    <p class="label">Avg daily sales</p>
    <p style="font-size:14px;color:#555;margin:0 0 16px;">{{printf "%.1f" .AvgDailySales}} units/day</p>
    <p class="label">Projected depletion</p>
    <p class="value" style="color:#dc2626;">~{{printf "%.0f" .DaysUntilDepletion}} days</p>
    {{if .EditURL}}
    <hr class="divider">
    <p><a href="{{.EditURL}}" style="color:#2563eb;">Update stock levels →</a></p>
    {{end}}
    <div class="footer">Nexus Commerce &mdash; Powered by your store</div>
  </div>
</body></html>`

const abTestConcludedTmpl = `<!DOCTYPE html>
<html lang="en">
<head><meta charset="UTF-8"><meta name="viewport" content="width=device-width,initial-scale=1">
<style>` + baseStyle + `
  .winner { background:#f0fdf4; border-left:4px solid #16a34a; padding:12px 16px; border-radius:0 8px 8px 0; margin-bottom:24px; }
</style></head>
<body>
  <div class="card">
    <h1>🧪 A/B Test Concluded</h1>
    <p>Your A/B test <strong>{{.TestName}}</strong> in <strong>{{.ShopName}}</strong> has been automatically concluded after {{.TestDays}} days.</p>
    <div class="winner">
      <strong>Winner: Variant {{.Winner | upper}}</strong> — {{printf "%.1f" .ConfidencePct}}% statistical confidence
    </div>
    <hr class="divider">
    <p class="label">Variant A conversion rate</p>
    <p style="font-size:16px;color:#555;margin:0 0 16px;">{{printf "%.2f" .VariantAConv}}%</p>
    <p class="label">Variant B conversion rate</p>
    <p style="font-size:16px;color:#555;margin:0 0 16px;">{{printf "%.2f" .VariantBConv}}%</p>
    <div class="footer">Nexus Commerce &mdash; Powered by your store</div>
  </div>
</body></html>`


// BackInStockData holds data for a back-in-stock notification email.
type BackInStockData struct {
	CustomerName string
	ProductTitle string
	ProductURL   string
}

// ReviewRequestItem is a single product in a review-request email.
type ReviewRequestItem struct {
	ProductTitle string
	ReviewURL    string
}

// ReviewRequestData holds data for a post-purchase review request email.
type ReviewRequestData struct {
	CustomerName string
	Items        []ReviewRequestItem
}


// SendBackInStock notifies a customer that a wishlisted product is back in stock.
func (m *Mailer) SendBackInStock(to string, data BackInStockData) error {
	if !m.Enabled() {
		m.logger.Info("SMTP not configured, skipping back-in-stock notice", zap.String("to", to))
		return nil
	}
	subject := fmt.Sprintf("Back in stock: %s", data.ProductTitle)
	return m.renderAndSend(to, subject, backInStockTmpl, data)
}

// SendReviewRequest sends a post-purchase review request to a customer.
func (m *Mailer) SendReviewRequest(to string, data ReviewRequestData) error {
	if !m.Enabled() {
		m.logger.Info("SMTP not configured, skipping review request", zap.String("to", to))
		return nil
	}
	return m.renderAndSend(to, "How was your order? Leave a review", reviewRequestTmpl, data)
}

const backInStockTmpl = `<!DOCTYPE html>
<html lang="en">
<head><meta charset="UTF-8"><meta name="viewport" content="width=device-width,initial-scale=1">
<style>` + baseStyle + `
  .btn { display:inline-block; padding:12px 28px; background:#111; color:#fff; border-radius:6px; text-decoration:none; font-size:15px; font-weight:600; }
</style></head>
<body>
  <div class="card">
    <h1>🎉 Back in Stock!</h1>
    <p>Hi {{.CustomerName}}, great news — an item from your wishlist is back in stock.</p>
    <hr class="divider">
    <p class="label">Product</p>
    <p class="value">{{.ProductTitle}}</p>
    <p>Grab it before it sells out again!</p>
    <a href="{{.ProductURL}}" class="btn">Shop Now</a>
    <div class="footer">Nexus Commerce &mdash; Powered by your store</div>
  </div>
</body></html>`

const reviewRequestTmpl = `<!DOCTYPE html>
<html lang="en">
<head><meta charset="UTF-8"><meta name="viewport" content="width=device-width,initial-scale=1">
<style>` + baseStyle + `
  .btn { display:inline-block; padding:10px 22px; background:#111; color:#fff; border-radius:6px; text-decoration:none; font-size:14px; font-weight:600; }
  .product-row { padding:14px 0; border-bottom:1px solid #f0f0f0; display:flex; justify-content:space-between; align-items:center; }
</style></head>
<body>
  <div class="card">
    <h1>&#11088; How was your order?</h1>
    <p>Hi {{.CustomerName}}, we hope you're enjoying your purchase! We'd love to hear what you think.</p>
    <hr class="divider">
    {{range .Items}}
    <div class="product-row">
      <span style="font-size:15px;color:#333;">{{.ProductTitle}}</span>
      <a href="{{.ReviewURL}}" class="btn">Leave a Review</a>
    </div>
    {{end}}
    <hr class="divider">
    <p style="font-size:13px;color:#999;">Your feedback helps other shoppers and improves our products.</p>
    <div class="footer">Nexus Commerce &mdash; Powered by your store</div>
  </div>
</body></html>`


// AbandonedCartItem is a single cart item in an abandoned-cart recovery email.
type AbandonedCartItem struct {
	Title    string
	Quantity int
	Price    string
	ImageURL string
}

// AbandonedCartData holds data for an abandoned-cart recovery email.
type AbandonedCartData struct {
	CustomerName string
	Items        []AbandonedCartItem
	RecoverURL   string
	Touch        int // 1, 2, or 3 — controls urgency messaging
}


// SendAbandonedCart sends an abandoned-cart recovery email (touch 1, 2, or 3).
func (m *Mailer) SendAbandonedCart(to string, data AbandonedCartData) error {
	if !m.Enabled() {
		m.logger.Info("SMTP not configured, skipping abandoned cart email", zap.String("to", to))
		return nil
	}
	subjects := map[int]string{
		1: "You left something behind…",
		2: "Still thinking it over?",
		3: "Last chance — your cart is expiring soon!",
	}
	subject := subjects[data.Touch]
	if subject == "" {
		subject = "Your cart is waiting for you"
	}
	return m.renderAndSend(to, subject, abandonedCartTmpl, data)
}

const abandonedCartTmpl = `<!DOCTYPE html>
<html lang="en">
<head><meta charset="UTF-8"><meta name="viewport" content="width=device-width,initial-scale=1">
<style>` + baseStyle + `
  .btn { display:inline-block; padding:14px 32px; background:#111; color:#fff; border-radius:6px; text-decoration:none; font-size:15px; font-weight:600; text-align:center; }
  .item-row { display:flex; gap:12px; padding:12px 0; border-bottom:1px solid #f0f0f0; align-items:center; }
  .item-img { width:56px; height:56px; object-fit:cover; border-radius:6px; background:#f3f4f6; }
  .item-info { flex:1; font-size:14px; color:#333; }
  .item-price { font-size:14px; font-weight:600; color:#111; }
</style></head>
<body>
  <div class="card">
    {{if eq .Touch 1}}
    <h1>🛒 You left something behind</h1>
    <p>Hi {{if .CustomerName}}{{.CustomerName}}{{else}}there{{end}}, looks like you left some great items in your cart!</p>
    {{else if eq .Touch 2}}
    <h1>🤔 Still thinking it over?</h1>
    <p>Hi {{if .CustomerName}}{{.CustomerName}}{{else}}there{{end}}, your cart is still waiting for you. These items are popular — don't miss out!</p>
    {{else}}
    <h1>⏰ Last chance!</h1>
    <p>Hi {{if .CustomerName}}{{.CustomerName}}{{else}}there{{end}}, your cart is about to expire. Complete your purchase now before these items sell out.</p>
    {{end}}
    <hr class="divider">
    {{range .Items}}
    <div class="item-row">
      {{if .ImageURL}}<img class="item-img" src="{{.ImageURL}}" alt="">{{end}}
      <div class="item-info">
        <div>{{.Title}}</div>
        <div style="color:#6b7280;font-size:12px;margin-top:2px;">Qty: {{.Quantity}}</div>
      </div>
      <div class="item-price">{{.Price}} EGP</div>
    </div>
    {{end}}
    <hr class="divider">
    {{if .RecoverURL}}
    <p style="text-align:center;"><a href="{{.RecoverURL}}" class="btn">Complete My Order →</a></p>
    {{end}}
    <p style="font-size:13px;color:#999;margin-top:24px;">If you have any questions, just reply to this email — we're always happy to help.</p>
    <div class="footer">Nexus Commerce &mdash; Powered by your store</div>
  </div>
</body></html>`
