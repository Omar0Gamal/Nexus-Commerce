// Package worker provides a Redis-backed background job queue.
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
	JobSendOrderConfirmation = "email:order_confirmation"
	JobSendOrderStatusUpdate = "email:order_status_update"
	JobSendLowStockAlert     = "email:low_stock_alert"
	JobSendPasswordReset     = "email:password_reset"
	JobSendAccountSuspended  = "email:account_suspended"

	JobDispatchWebhook = "webhook:dispatch"

	JobGenerateInvoice = "invoice:generate"

	JobSendBackInStock   = "email:back_in_stock"
	JobSendReviewRequest = "email:review_request"
)

const DefaultQueueKey = "nexus:jobs"

const (
	defaultMaxRetries  = 3
	defaultRetryDelay  = 2 * time.Second
	defaultDeadLetterQ = "nexus:jobs:dead"
)

type Job struct {
	Type       string          `json:"type"`
	Payload    json.RawMessage `json:"payload"`
	RetryCount int             `json:"retry_count,omitempty"`
}

type Queue struct {
	rdb      *redis.Client
	queueKey string
}

func NewQueue(rdb *redis.Client, queueKey string) *Queue {
	if queueKey == "" {
		queueKey = DefaultQueueKey
	}
	return &Queue{rdb: rdb, queueKey: queueKey}
}

func (q *Queue) Enqueue(ctx context.Context, job Job) error {
	jobBytes, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("marshal job: %w", err)
	}
	return q.rdb.LPush(ctx, q.queueKey, jobBytes).Err()
}

func (q *Queue) EnqueueTyped(ctx context.Context, jobType string, payload any) error {
	raw, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal payload: %w", err)
	}
	return q.Enqueue(ctx, Job{Type: jobType, Payload: raw})
}

func (q *Queue) Dequeue(ctx context.Context) (Job, error) {
	popResult, err := q.rdb.BRPop(ctx, 5*time.Second, q.queueKey).Result()
	if err != nil {
		if err == redis.Nil {
			return Job{}, nil
		}
		return Job{}, fmt.Errorf("brpop: %w", err)
	}
	var job Job
	if err := json.Unmarshal([]byte(popResult[1]), &job); err != nil {
		return Job{}, fmt.Errorf("unmarshal job: %w", err)
	}
	return job, nil
}

type HandlerFunc func(ctx context.Context, payload json.RawMessage) error

type Worker struct {
	queue       *Queue
	handlers    map[string]HandlerFunc
	logger      *zap.Logger
	concurrency int
	maxRetries  int
	retryDelay  time.Duration
	deadLetterQ string
}

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

func (w *Worker) SetDeadLetterQueue(queueKey string) {
	if queueKey == "" {
		w.deadLetterQ = defaultDeadLetterQ
		return
	}
	w.deadLetterQ = queueKey
}

func (w *Worker) Register(jobType string, handler HandlerFunc) {
	w.handlers[jobType] = handler
}

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

	defer func() {
		if r := recover(); r != nil {
			err := fmt.Errorf("panic in job handler: %v", r)
			w.logger.Error("job panic recovered",
				zap.String("type", job.Type),
				zap.Int("retry_count", job.RetryCount),
				zap.Any("panic", r),
			)
			if retryErr := w.handleFailure(ctx, job, err); retryErr != nil {
				w.logger.Error("failed to schedule retry/dead-letter after panic",
					zap.String("type", job.Type),
					zap.Error(retryErr),
				)
			}
		}
	}()

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
	deadBytes, err := json.Marshal(deadPayload)
	if err != nil {
		return fmt.Errorf("marshal dead-letter payload: %w", err)
	}
	if err := w.queue.rdb.LPush(ctx, w.deadLetterQ, deadBytes).Err(); err != nil {
		return fmt.Errorf("enqueue dead-letter: %w", err)
	}

	w.logger.Warn("job moved to dead-letter queue",
		zap.String("type", job.Type),
		zap.Int("retry_count", job.RetryCount),
		zap.String("dead_letter_queue", w.deadLetterQ),
	)
	return nil
}
