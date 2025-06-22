// Package metrics defines all Prometheus metrics for the backend API.
package metrics

import (
	"context"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

var (
	// RequestsTotal counts HTTP requests served by the backend.
	RequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "backend_requests_total",
		Help: "Total number of HTTP requests processed by the backend API.",
	}, []string{"method", "status_class"})

	// RequestDuration tracks request latency.
	RequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "backend_request_duration_seconds",
		Help:    "HTTP request latency in seconds.",
		Buckets: prometheus.DefBuckets,
	}, []string{"method", "status_class"})

	// DBQueryDuration tracks database query latency by query name.
	DBQueryDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "backend_db_query_duration_seconds",
		Help:    "Database query execution time in seconds.",
		Buckets: []float64{.005, .01, .025, .05, .1, .25, .5, 1},
	}, []string{"query_name"})

	// PgxpoolIdleConns tracks idle DB connections.
	PgxpoolIdleConns = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "backend_pgxpool_idle_conns",
		Help: "Number of idle connections in the pgxpool.",
	})

	// PgxpoolAcquiredConns tracks acquired DB connections.
	PgxpoolAcquiredConns = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "backend_pgxpool_acquired_conns",
		Help: "Number of acquired connections in the pgxpool.",
	})

	// AnalyticsWorkerEvents counts events processed by the analytics worker.
	AnalyticsWorkerEvents = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "analytics_worker_events_processed_total",
		Help: "Total analytics events processed by the worker.",
	}, []string{"event_type"})

	// AnalyticsWorkerDLQDepth tracks the dead-letter queue depth.
	AnalyticsWorkerDLQDepth = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "analytics_worker_dlq_depth",
		Help: "Number of events currently in the analytics dead-letter queue.",
	})
)

// StatusClass converts an HTTP status code to a class string like "2xx", "5xx".
func StatusClass(code int) string {
	switch {
	case code < 200:
		return "1xx"
	case code < 300:
		return "2xx"
	case code < 400:
		return "3xx"
	case code < 500:
		return "4xx"
	default:
		return "5xx"
	}
}

// StartMetricsServer launches a Prometheus metrics HTTP server on the given
// address (e.g. ":9090") in a background goroutine. It stops when ctx is done.
func StartMetricsServer(ctx context.Context, addr string, logger *zap.Logger) {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	srv := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 3 * time.Second,
	}

	go func() {
		logger.Info("metrics server listening", zap.String("addr", addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("metrics server error", zap.Error(err))
		}
	}()

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil && err != http.ErrServerClosed {
			logger.Warn("metrics server shutdown error", zap.Error(err))
		}
	}()
}

// CollectPoolStats periodically updates pgxpool gauge metrics.
func CollectPoolStats(ctx context.Context, pool *pgxpool.Pool, interval time.Duration, logger *zap.Logger) {
	if logger == nil {
		logger = zap.NewNop()
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			stat := pool.Stat()
			PgxpoolIdleConns.Set(float64(stat.IdleConns()))
			acquired := stat.AcquiredConns()
			total := stat.TotalConns()

			PgxpoolAcquiredConns.Set(float64(acquired))
			if total > 0 {
				utilization := float64(acquired) / float64(total)
				if utilization >= 0.90 {
					logger.Warn("database pool utilization is high",
						zap.Int32("acquired", acquired),
						zap.Int32("total", total),
						zap.Float64("utilization", utilization),
					)
				}
			}
		}
	}
}
