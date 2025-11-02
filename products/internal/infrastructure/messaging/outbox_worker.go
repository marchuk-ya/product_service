package messaging

import (
	"context"
	"fmt"
	"product_service/products/internal/infrastructure/events"
	"product_service/products/internal/usecase/ports"
	"sync"
	"time"

	"go.uber.org/zap"
)

const (
	shutdownGracePeriod = 5 * time.Second
)

type OutboxWorker struct {
	outboxRepo    ports.OutboxRepository
	publisher      ports.EventPublisher
	eventAdapter   EventAdapter
	logger        *zap.Logger
	metrics       ports.MetricsCollector
	interval      time.Duration
	batchSize     int
	maxRetries    int
	baseBackoff   time.Duration
	maxBackoff    time.Duration
	concurrency   int
	stopChan      chan struct{}
	doneChan      chan struct{}
}

func NewOutboxWorker(
	outboxRepo ports.OutboxRepository,
	publisher ports.EventPublisher,
	logger *zap.Logger,
	interval time.Duration,
	batchSize int,
	maxRetries int,
	baseBackoff time.Duration,
	maxBackoff time.Duration,
	metrics ports.MetricsCollector,
	concurrency int,
) *OutboxWorker {
	if concurrency <= 0 {
		concurrency = 3
	}
	return &OutboxWorker{
		outboxRepo:   outboxRepo,
		publisher:    publisher,
		eventAdapter: NewSmartEventAdapter(),
		logger:       logger,
		metrics:      metrics,
		interval:     interval,
		batchSize:    batchSize,
		maxRetries:   maxRetries,
		baseBackoff:  baseBackoff,
		maxBackoff:   maxBackoff,
		concurrency:  concurrency,
		stopChan:     make(chan struct{}),
		doneChan:     make(chan struct{}),
	}
}

func (w *OutboxWorker) Start(ctx context.Context) {
	go w.run(ctx)
}

func (w *OutboxWorker) Stop() {
	close(w.stopChan)
	<-w.doneChan
}

func (w *OutboxWorker) run(ctx context.Context) {
	defer close(w.doneChan)

	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			w.logger.Info("Outbox worker stopped: context cancelled")
			return
		case <-w.stopChan:
			w.logger.Info("Outbox worker stopped: stop signal received")
			return
		case <-ticker.C:
			w.processPendingEvents(ctx)
		}
	}
}

func (w *OutboxWorker) processPendingEvents(ctx context.Context) {
	events, err := w.outboxRepo.GetPendingEvents(ctx, w.batchSize)
	if err != nil {
		w.logger.Error("Failed to get pending events", zap.Error(err))
		return
	}

	if len(events) == 0 {
		return
	}

	w.logger.Info("Processing pending events", zap.Int("count", len(events)), zap.Int("concurrency", w.concurrency))

	sem := make(chan struct{}, w.concurrency)
	var wg sync.WaitGroup

	for _, event := range events {
		select {
		case <-ctx.Done():
			w.logger.Warn("Context cancelled, stopping event processing",
				zap.Int64("event_id", event.ID),
				zap.String("event_type", event.EventType),
			)
			return
		default:
		}
		
		wg.Add(1)
		go func(e ports.OutboxEvent) {
			defer wg.Done()
			
			select {
			case <-ctx.Done():
				w.logger.Warn("Context cancelled, skipping event",
					zap.Int64("event_id", e.ID),
					zap.String("event_type", e.EventType),
				)
				return
			case sem <- struct{}{}:
				defer func() { <-sem }()
			}
			
			w.processEvent(ctx, e)
		}(event)
	}

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()
	
	select {
	case <-ctx.Done():
		w.logger.Warn("Context cancelled while waiting for events to complete")
		select {
		case <-done:
		case <-time.After(shutdownGracePeriod):
			w.logger.Warn("Timeout waiting for events to complete after context cancellation",
				zap.Duration("grace_period", shutdownGracePeriod),
			)
		}
	case <-done:
	}
}

func (w *OutboxWorker) processEvent(ctx context.Context, event ports.OutboxEvent) {
	if ctx.Err() != nil {
		w.logger.Warn("Context cancelled before processing event",
			zap.Int64("event_id", event.ID),
			zap.String("event_type", event.EventType),
			zap.Error(ctx.Err()),
		)
		return
	}
	
	if err := w.publishEventWithRetry(ctx, event); err != nil {
		if ctx.Err() != nil {
			w.logger.Warn("Context cancelled during event publication",
				zap.Int64("event_id", event.ID),
				zap.String("event_type", event.EventType),
				zap.Error(ctx.Err()),
			)
			return
		}
		w.handlePublishFailure(ctx, event, err)
		return
	}

	if ctx.Err() != nil {
		w.logger.Warn("Context cancelled after event publication",
			zap.Int64("event_id", event.ID),
			zap.String("event_type", event.EventType),
			zap.Error(ctx.Err()),
		)
		return
	}

	if err := w.markEventAsPublished(ctx, event); err != nil {
		w.logger.Error("Failed to mark event as published",
			zap.Int64("event_id", event.ID),
			zap.Error(err),
		)
		return
	}

	w.logger.Info("Event published successfully",
		zap.Int64("event_id", event.ID),
		zap.String("event_type", event.EventType),
	)
}

func (w *OutboxWorker) handlePublishFailure(ctx context.Context, event ports.OutboxEvent, publishErr error) {
	w.logger.Error("Failed to publish event after retries",
		zap.Int64("event_id", event.ID),
		zap.String("event_type", event.EventType),
		zap.Int("retry_count", event.RetryCount),
		zap.Error(publishErr),
	)
	
	retryCount := event.RetryCount + 1
	
	if w.metrics != nil {
		w.metrics.RecordOutboxRetryAttempt(event.EventType, retryCount)
	}
	
	if retryCount > w.maxRetries {
		w.logger.Warn("Event exceeded max retries, moving to DLQ",
			zap.Int64("event_id", event.ID),
			zap.String("event_type", event.EventType),
			zap.Int("retry_count", retryCount),
		)
		
		reason := fmt.Sprintf("Failed after %d retry attempts: %v", retryCount, publishErr)
		if dlqErr := w.outboxRepo.MoveToDLQ(ctx, event.ID, reason); dlqErr != nil {
			w.logger.Error("Failed to move event to DLQ",
				zap.Int64("event_id", event.ID),
				zap.Error(dlqErr),
			)
			w.markEventAsFailed(ctx, event.ID, retryCount)
		}
		
		if w.metrics != nil {
			w.metrics.RecordOutboxEventProcessed(event.EventType, "dlq")
		}
	} else {
		w.markEventAsFailed(ctx, event.ID, retryCount)
		
		if w.metrics != nil {
			w.metrics.RecordOutboxEventProcessed(event.EventType, "failed")
		}
	}
}

func (w *OutboxWorker) markEventAsFailed(ctx context.Context, eventID int64, retryCount int) {
	if err := w.outboxRepo.MarkAsFailed(ctx, eventID, retryCount); err != nil {
		w.logger.Error("Failed to mark event as failed",
			zap.Int64("event_id", eventID),
			zap.Error(err),
		)
	}
}

func (w *OutboxWorker) markEventAsPublished(ctx context.Context, event ports.OutboxEvent) error {
	if err := w.outboxRepo.MarkAsPublished(ctx, event.ID); err != nil {
		return err
	}
	
	if w.metrics != nil {
		w.metrics.RecordOutboxEventProcessed(event.EventType, "published")
	}
	
	return nil
}

func (w *OutboxWorker) publishEventWithRetry(ctx context.Context, event ports.OutboxEvent) error {
	var lastErr error
	
	for attempt := 0; attempt <= w.maxRetries; attempt++ {
		if attempt > 0 {
			backoff := w.baseBackoff * time.Duration(1<<uint(attempt-1))
			if backoff > w.maxBackoff {
				backoff = w.maxBackoff
			}
			
			w.logger.Info("Retrying event publication",
				zap.Int64("event_id", event.ID),
				zap.String("event_type", event.EventType),
				zap.Int("attempt", attempt),
				zap.Duration("backoff", backoff),
			)
			
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(backoff):
			}
		}
		
		if err := w.publishEvent(ctx, event); err == nil {
			return nil
		} else {
			lastErr = err
			if attempt < w.maxRetries {
				w.logger.Warn("Event publication failed, will retry",
					zap.Int64("event_id", event.ID),
					zap.String("event_type", event.EventType),
					zap.Int("attempt", attempt+1),
					zap.Error(err),
				)
			}
		}
	}
	
	return fmt.Errorf("failed after %d attempts: %w", w.maxRetries+1, lastErr)
}

func (w *OutboxWorker) publishEvent(ctx context.Context, event ports.OutboxEvent) error {
	eventType, productID, timestamp, err := w.eventAdapter.AdaptEvent(event)
	if err != nil {
		return fmt.Errorf("failed to adapt event: %w", err)
	}

	if !timestamp.IsZero() {
		ctx = WithTimestamp(ctx, timestamp)
	}

	switch eventType {
	case events.EventTypeProductCreated:
		return w.publisher.PublishProductCreated(ctx, productID)
	case events.EventTypeProductDeleted:
		return w.publisher.PublishProductDeleted(ctx, productID)
	default:
		return fmt.Errorf("unknown event type: %s", eventType)
	}
}

