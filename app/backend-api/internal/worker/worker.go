// Package worker provides a Redis-backed background job queue.
//
// Jobs are serialised to JSON and pushed onto a Redis list.  A pool of
// worker goroutines continuously BRPOP from that list and dispatch each job
// to its registered handler.
package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

const (
	// Email jobs
	JobSendOrderConfirmation = "email:order_confirmation"
	JobSendOrderStatusUpdate = "email:order_status_update"
	JobSendLowStockAlert     = "email:low_stock_alert"
	JobSendPasswordReset     = "email:password_reset"
	JobSendAccountSuspended  = "email:account_suspended"

	// Webhook jobs
	JobDispatchWebhook = "webhook:dispatch"

	// Invoice jobs
	JobGenerateInvoice = "invoice:generate"

	// Phase 13: engagement jobs
	JobSendBackInStock   = "email:back_in_stock"
	JobSendReviewRequest = "email:review_request"
)

// DefaultQueueKey is the Redis list key used as the job queue.
const DefaultQueueKey = "nexus:jobs"

const (
	defaultMaxRetries  = 3
	defaultRetryDelay  = 2 * time.Second
	defaultDeadLetterQ = "nexus:jobs:dead"
)

// Job represents a single unit of background work.
type Job struct {
	// Type identifies which handler processes this job.
	Type string `json:"type"`
	// Payload holds job-specific data (JSON-encoded).
	Payload json.RawMessage `json:"payload"`
	// RetryCount tracks how many times the job has been retried.
	RetryCount int `json:"retry_count,omitempty"`
}

// Queue is a Redis-backed FIFO queue.
type Queue struct {
	rdb      *redis.Client
	queueKey string
}

// Pass DefaultQueueKey to use the standard key.
func NewQueue(rdb *redis.Client, queueKey string) *Queue {
	if queueKey == "" {
		queueKey = DefaultQueueKey
	}
	return &Queue{rdb: rdb, queueKey: queueKey}
}

// Enqueue serialises job and pushes it to the back of the Redis list.
func (q *Queue) Enqueue(ctx context.Context, job Job) error {
	data, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("marshal job: %w", err)
	}
	return q.rdb.LPush(ctx, q.queueKey, data).Err()
}

// EnqueueTyped is a convenience helper that accepts any payload struct.
func (q *Queue) EnqueueTyped(ctx context.Context, jobType string, payload any) error {
	raw, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal payload: %w", err)
	}
	return q.Enqueue(ctx, Job{Type: jobType, Payload: raw})
}

// Dequeue blocks until a job is available or the context is cancelled.
// It returns (job, nil) on success, (Job{}, context.Canceled) on shutdown.
func (q *Queue) Dequeue(ctx context.Context) (Job, error) {
	result, err := q.rdb.BRPop(ctx, 5*time.Second, q.queueKey).Result()
	if err != nil {
		if err == redis.Nil {
			// Timeout with no message — caller should retry.
			return Job{}, nil
		}
		return Job{}, fmt.Errorf("brpop: %w", err)
	}
	// result[0] = key, result[1] = value
	var job Job
	if err := json.Unmarshal([]byte(result[1]), &job); err != nil {
		return Job{}, fmt.Errorf("unmarshal job: %w", err)
	}
	return job, nil
}

// HandlerFunc processes a single job.
type HandlerFunc func(ctx context.Context, payload json.RawMessage) error

// Worker runs a pool of goroutines that consume jobs from a Queue.
type Worker struct {
	queue       *Queue
	handlers    map[string]HandlerFunc
	logger      *zap.Logger
	concurrency int
	maxRetries  int
	retryDelay  time.Duration
	deadLetterQ string
}

// concurrency is the number of parallel goroutines consuming the queue.
func NewWorker(queue *Queue, concurrency int, logger *zap.Logger) *Worker {
	if concurrency < 1 {
		concurrency = 1
	}
	if logger == nil {
		logger = zap.NewNop()
	}
	return &Worker{
		queue:       queue,
		handlers:    make(map[string]HandlerFunc),
		logger:      logger,
		concurrency: concurrency,
		maxRetries:  defaultMaxRetries,
		retryDelay:  defaultRetryDelay,
		deadLetterQ: defaultDeadLetterQ,
	}
}

// SetRetryPolicy configures retry behavior for failed jobs.
func (w *Worker) SetRetryPolicy(maxRetries int, retryDelay time.Duration) {
	if maxRetries < 0 {
		maxRetries = 0
	}
	if retryDelay <= 0 {
		retryDelay = defaultRetryDelay
	}
	w.maxRetries = maxRetries
	w.retryDelay = retryDelay
}

// SetDeadLetterQueue overrides the queue key for permanently failed jobs.
func (w *Worker) SetDeadLetterQueue(queueKey string) {
	if queueKey == "" {
		w.deadLetterQ = defaultDeadLetterQ
		return
	}
	w.deadLetterQ = queueKey
}

// Register associates a job type with a handler.
// Call before Start.
func (w *Worker) Register(jobType string, handler HandlerFunc) {
	w.handlers[jobType] = handler
}

// Start launches the worker goroutines and blocks until ctx is cancelled.
// Typically called in a separate goroutine: go worker.Start(ctx)
func (w *Worker) Start(ctx context.Context) {
	w.logger.Info("worker pool starting", zap.Int("concurrency", w.concurrency))

	sem := make(chan struct{}, w.concurrency)

	for {
		select {
		case <-ctx.Done():
			w.logger.Info("worker pool shutting down")
			return
		default:
		}

		sem <- struct{}{}
		go func() {
			defer func() { <-sem }()
			job, err := w.queue.Dequeue(ctx)
			if err != nil {
				if ctx.Err() == nil {
					w.logger.Error("dequeue error", zap.Error(err))
				}
				return
			}
			// Timeout returned empty job
			if job.Type == "" {
				return
			}
			w.process(ctx, job)
		}()
	}
}

func (w *Worker) process(ctx context.Context, job Job) {
	handler, ok := w.handlers[job.Type]
	if !ok {
		w.logger.Warn("no handler for job type", zap.String("type", job.Type))
		return
	}

	start := time.Now()
	if err := handler(ctx, job.Payload); err != nil {
		if retryErr := w.handleFailure(ctx, job, err); retryErr != nil {
			w.logger.Error("failed to schedule retry/dead-letter",
				zap.String("type", job.Type),
				zap.Int("retry_count", job.RetryCount),
				zap.Error(retryErr),
			)
		}
		w.logger.Error("job failed",
			zap.String("type", job.Type),
			zap.Int("retry_count", job.RetryCount),
			zap.Duration("elapsed", time.Since(start)),
			zap.Error(err),
		)
		return
	}

	w.logger.Info("job completed",
		zap.String("type", job.Type),
		zap.Int("retry_count", job.RetryCount),
		zap.Duration("elapsed", time.Since(start)),
	)
}

func (w *Worker) handleFailure(ctx context.Context, job Job, handlerErr error) error {
	if job.RetryCount < w.maxRetries {
		retryJob := job
		retryJob.RetryCount++

		// Apply a small increasing backoff before re-enqueueing.
		delay := time.Duration(retryJob.RetryCount) * w.retryDelay
		go func(j Job, d time.Duration) {
			timer := time.NewTimer(d)
			defer timer.Stop()
			select {
			case <-ctx.Done():
				return
			case <-timer.C:
			}
			if err := w.queue.Enqueue(ctx, j); err != nil {
				w.logger.Error("failed to enqueue retry job",
					zap.String("type", j.Type),
					zap.Int("retry_count", j.RetryCount),
					zap.Error(err),
				)
			}
		}(retryJob, delay)

		w.logger.Warn("job scheduled for retry",
			zap.String("type", retryJob.Type),
			zap.Int("retry_count", retryJob.RetryCount),
			zap.Duration("delay", delay),
		)
		return nil
	}

	deadPayload := map[string]any{
		"job":         job,
		"failed_at":   time.Now().UTC().Format(time.RFC3339),
		"error":       handlerErr.Error(),
		"max_retries": w.maxRetries,
	}
	data, err := json.Marshal(deadPayload)
	if err != nil {
		return fmt.Errorf("marshal dead-letter payload: %w", err)
	}
	if err := w.queue.rdb.LPush(ctx, w.deadLetterQ, data).Err(); err != nil {
		return fmt.Errorf("enqueue dead-letter: %w", err)
	}

	w.logger.Warn("job moved to dead-letter queue",
		zap.String("type", job.Type),
		zap.Int("retry_count", job.RetryCount),
		zap.String("dead_letter_queue", w.deadLetterQ),
	)
	return nil
}
