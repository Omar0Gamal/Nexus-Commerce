package seo

import (
	"context"
	"time"

	"go.uber.org/zap"
)

const workerInterval = 24 * time.Hour

type Worker struct {
	svc    *Service
	logger *zap.Logger
}

func NewWorker(svc *Service, logger *zap.Logger) *Worker {
	return &Worker{svc: svc, logger: logger}
}

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
