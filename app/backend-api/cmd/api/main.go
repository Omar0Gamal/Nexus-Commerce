package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"backend-api/internal/config"
	"backend-api/internal/db"
	"backend-api/internal/email"
	"backend-api/internal/paymob"
	"backend-api/internal/shared/middleware"
	"backend-api/internal/worker"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type infra struct {
	Cfg          *config.Config
	Queries      *db.Queries
	Pool         *pgxpool.Pool
	Redis        *redis.Client
	Mailer       *email.Mailer
	PaymobClient *paymob.Client
	JobQueue     *worker.Queue
	Logger       *zap.Logger
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "fatal: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	cfg := config.Load()

	logger, err := zap.NewProduction()
	if err != nil {
		return fmt.Errorf("create logger: %w", err)
	}
	defer logger.Sync() //nolint:errcheck

	// Share the single logger with middleware.
	middleware.SetLogger(logger)

	// ── Infrastructure ──
	deps, err := initInfra(cfg, logger)
	if err != nil {
		return fmt.Errorf("init infrastructure: %w", err)
	}
	defer deps.Close()

	// ── Services ──
	svc := initServices(deps)

	// ── Background workers ──
	workerCtx, cancelWorker := context.WithCancel(context.Background())
	defer cancelWorker()
	startWorkers(workerCtx, deps, svc)

	// ── HTTP server with graceful shutdown ──
	router := newRouter(svc, deps.Pool, deps.Redis)

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           router,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	// Listen for OS signals to trigger graceful shutdown.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	errCh := make(chan error, 1)
	go func() {
		logger.Info("backend API starting", zap.String("port", cfg.Port))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	// Block until signal or server error.
	select {
	case <-ctx.Done():
		logger.Info("shutdown signal received, draining connections…")
	case err := <-errCh:
		return fmt.Errorf("server error: %w", err)
	}

	cancelWorker() // stop background workers

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("server forced to shutdown", zap.Error(err))
	}
	logger.Info("server stopped gracefully")
	return nil
}

// openPool creates a pgxpool with tuning from config.
func openPool(cfg *config.Config) (*pgxpool.Pool, error) {
	ctx := context.Background()

	pgxCfg, err := pgxpool.ParseConfig(cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("parse pool config: %w", err)
	}
	pgxCfg.MaxConns = cfg.DBMaxConns
	pgxCfg.MinConns = cfg.DBMinConns
	pgxCfg.MaxConnLifetime = time.Duration(cfg.DBMaxConnLifetime) * time.Second
	pgxCfg.MaxConnIdleTime = time.Duration(cfg.DBMaxConnIdleTime) * time.Second
	pgxCfg.HealthCheckPeriod = time.Duration(cfg.DBHealthCheckPeriod) * time.Second

	pool, err := pgxpool.NewWithConfig(ctx, pgxCfg)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("ping database: %w", err)
	}
	return pool, nil
}

func initInfra(cfg *config.Config, logger *zap.Logger) (*infra, error) {
	pool, err := openPool(cfg)
	if err != nil {
		return nil, fmt.Errorf("database pool: %w", err)
	}

	rdb := redis.NewClient(&redis.Options{Addr: cfg.RedisAddr})
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		logger.Warn("Redis unavailable (cart will not work)", zap.Error(err))
	}

	deps := &infra{
		Cfg:     cfg,
		Queries: db.New(pool),
		Pool:    pool,
		Redis:   rdb,
		Mailer: email.New(
			cfg.SMTPHost, cfg.SMTPPort,
			cfg.SMTPUsername, cfg.SMTPPassword, cfg.SMTPFrom,
			logger,
		),
		PaymobClient: paymob.NewClient(
			cfg.PaymobAPIKey, cfg.PaymobIntegrationID,
			cfg.PaymobIframeID, cfg.PaymobHMACSecret,
		),
		JobQueue: worker.NewQueue(rdb, worker.DefaultQueueKey),
		Logger:   logger,
	}
	return deps, nil
}

func (i *infra) Close() {
	if i.Pool != nil {
		i.Pool.Close()
	}
	if i.Redis != nil {
		i.Redis.Close()
	}
}
