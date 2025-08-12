package seo

import (
	"context"
	"time"

	"go.uber.org/zap"
)

const workerInterval = 24 * time.Hour

// Worker runs nightly SEO maintenance tasks.
type Worker struct {
	svc    *Service
	logger *zap.Logger
}

// NewWorker creates an SEO background worker.
func NewWorker(svc *Service, logger *zap.Logger) *Worker {
	return &Worker{svc: svc, logger: logger}
}

// Run starts the periodic worker. It fires once at startup and then every 24 h.
// Respects ctx for clean shutdown.
func (w *Worker) Run(ctx context.Context) {
	w.logger.Info("seo worker: starting")
	w.runOnce(ctx)

	ticker := time.NewTicker(workerInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			w.logger.Info("seo worker: stopping")
			return
		case <-ticker.C:
			w.runOnce(ctx)
		}
	}
}

func (w *Worker) runOnce(ctx context.Context) {
	w.logger.Info("seo worker: warming sitemaps and auditing shops")
	w.svc.WarmSitemapForAllShops(ctx)
}
