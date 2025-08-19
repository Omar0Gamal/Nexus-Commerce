package main

import (
	"context"
	"time"

	"backend-api/internal/analytics"
	"backend-api/internal/billing"
	"backend-api/internal/cart"
	"backend-api/internal/currency"
	"backend-api/internal/metrics"
	"backend-api/internal/seo"
	"backend-api/internal/worker"
)

// startWorkers launches all background goroutines. They respect ctx for
// cancellation and will exit when the context is cancelled during shutdown.
func startWorkers(ctx context.Context, d *infra, svc *services) {
	// Job queue processor
	jw := worker.NewWorker(d.JobQueue, d.Cfg.WorkerConcurrency, d.Logger)
	for jobType, h := range worker.EmailHandlers(d.Mailer) {
		jw.Register(jobType, h)
	}
	jw.Register(worker.JobDispatchWebhook, worker.WebhookHandler())
	jw.Register(worker.JobGenerateInvoice, worker.InvoiceHandler())
	go jw.Start(ctx)

	// Domain workers
	go analytics.NewWorker(d.Redis, d.Queries, d.Logger).Run(ctx)
	go worker.NewSupportAutomationWorker(d.Queries, d.Logger).Run(ctx)
	go billing.NewInvoiceWorker(d.Queries, d.PaymobClient, d.Mailer, d.Logger).Run(ctx)
	go cart.NewAbandonedCartWorker(d.Queries, d.Redis, d.Mailer, d.Logger, d.Cfg.FrontendURL).Run(ctx)
	go currency.NewRateWorker(svc.CurrencySvc, d.Logger).Run(ctx)
	go worker.NewSubscriptionRenewalWorker(d.Queries, d.Logger).Run(ctx)
	go worker.NewLoyaltyExpiryWorker(d.Queries, d.Logger).Run(ctx)
	go seo.NewWorker(svc.SEOSvc, d.Logger).Run(ctx)

	// Metrics
	go metrics.StartMetricsServer(ctx, ":"+d.Cfg.MetricsPort, d.Logger)
	go metrics.CollectPoolStats(ctx, d.Pool, 15*time.Second, d.Logger)
}
