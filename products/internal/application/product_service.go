package application

import (
	"context"
	"errors"
	"fmt"
	"product_service/products/internal/domain"
	"product_service/products/internal/usecase/ports"
	"time"
)

var _ ports.ProductApplicationService = (*ProductService)(nil)

var _ ports.UoWFactory = (*ProductService)(nil)

type ProductService struct {
	uowFactory     ports.UoWFactory
	eventPublisher ports.DomainEventPublisher
	logger         ports.Logger
	retrier        ports.Retrier
	retryConfig    ports.RetryConfig
	metrics        ports.MetricsCollector
}

func NewProductService(
	uowFactory ports.UoWFactory,
	eventPublisher ports.DomainEventPublisher,
	logger ports.Logger,
	retrier ports.Retrier,
	metrics ports.MetricsCollector,
) *ProductService {
	retryCfg := ports.RetryConfig{
		MaxAttempts:  3,
		BaseBackoff:  100 * time.Millisecond,
		MaxBackoff:   1 * time.Second,
		InitialDelay: 50 * time.Millisecond,
	}

	return &ProductService{
		uowFactory:     uowFactory,
		eventPublisher: eventPublisher,
		logger:         logger,
		retrier:        retrier,
		retryConfig:    retryCfg,
		metrics:        metrics,
	}
}

func (s *ProductService) CreateUnitOfWork() ports.UnitOfWork {
	return s.uowFactory.CreateUnitOfWork()
}

type TransactionResult struct {
	Success bool
	Error   error
}

func (s *ProductService) ExecuteInTransaction(ctx context.Context, fn func(uow ports.UnitOfWork) error) error {
	if s.retrier == nil {
		uow := s.uowFactory.CreateUnitOfWork()
		if err := uow.Begin(ctx); err != nil {
			return fmt.Errorf("failed to start transaction: %w", err)
		}
		
		originalErr := fn(uow)
		
		defer func() {
			if originalErr != nil {
				if rollbackErr := uow.Rollback(); rollbackErr != nil {
					s.logger.Error("Failed to rollback transaction",
						ports.NewField("error", rollbackErr),
						ports.NewField("original_error", originalErr),
					)
				}
			} else {
				if commitErr := uow.Commit(); commitErr != nil {
					s.logger.Error("Failed to commit transaction",
						ports.NewField("error", commitErr),
					)
					if rollbackErr := uow.Rollback(); rollbackErr != nil {
						s.logger.Error("Failed to rollback after failed commit",
							ports.NewField("error", rollbackErr),
							ports.NewField("commit_error", commitErr),
						)
					}
				}
			}
		}()
		
		return originalErr
	}

	var lastErr error
	var retryAttempts int
	err := s.retrier.Do(ctx, s.retryConfig, func() error {
		retryAttempts++
		if retryAttempts > 1 && s.metrics != nil {
			s.metrics.IncrementTransactionRetry()
		}
		
		uow := s.uowFactory.CreateUnitOfWork()

		if err := uow.Begin(ctx); err != nil {
			s.logger.Error("Failed to start transaction",
				ports.NewField("error", err),
				ports.NewField("retry_attempt", retryAttempts),
			)
			return fmt.Errorf("failed to start transaction: %w", err)
		}

		var originalErr error
		defer func() {
			if originalErr != nil {
				if rollbackErr := uow.Rollback(); rollbackErr != nil {
					s.logger.Error("Failed to rollback transaction",
						ports.NewField("error", rollbackErr),
						ports.NewField("original_error", originalErr),
					)
				}
			} else {
				if commitErr := uow.Commit(); commitErr != nil {
					s.logger.Error("Failed to commit transaction",
						ports.NewField("error", commitErr),
					)
					if rollbackErr := uow.Rollback(); rollbackErr != nil {
						s.logger.Error("Failed to rollback after failed commit",
							ports.NewField("error", rollbackErr),
							ports.NewField("commit_error", commitErr),
						)
					}
					originalErr = fmt.Errorf("failed to commit transaction: %w", commitErr)
				}
			}
		}()

		originalErr = fn(uow)
		lastErr = originalErr
		return originalErr
	})

	if err != nil {
		if s.metrics != nil {
			s.metrics.IncrementTransactionRetryFailed()
		}
		return NewRetryExhaustedError(s.retryConfig.MaxAttempts, lastErr)
	}

	if retryAttempts > 1 && s.metrics != nil {
		s.metrics.IncrementTransactionRetrySuccess()
	}

	return nil
}

func (s *ProductService) CreateProductWithEvent(
	ctx context.Context,
	product *domain.Product,
	idempotencyKey string,
) error {
	return s.ExecuteInTransaction(ctx, func(uow ports.UnitOfWork) error {
		if err := uow.ProductRepository().Create(ctx, product); err != nil {
			s.logger.Error("Failed to save product to repository",
				ports.NewField("error", err),
				ports.NewField("product_id", product.ID),
			)
			return NewTransactionError("create product", err)
		}

		product.RecordCreatedEvent()

		events := product.DomainEvents()
		if len(events) == 0 {
			return fmt.Errorf("no domain events found in product")
		}

		outboxRepo := uow.OutboxRepository()
		if err := s.publishDomainEventsBatch(ctx, events, idempotencyKey, outboxRepo); err != nil {
			s.logger.Error("Failed to save events to outbox",
				ports.NewField("error", err),
				ports.NewField("product_id", product.ID),
			)
			if len(events) > 0 {
				return NewEventPublishError(product.ID, events[0].EventType(), err)
			}
			return NewTransactionError("publish events", err)
		}

		product.ClearDomainEvents()

		s.logger.Info("Product created successfully",
			ports.NewField("product_id", product.ID),
			ports.NewField("product_name", product.Name.Value()),
		)

		return nil
	})
}

func (s *ProductService) DeleteProductWithEvent(
	ctx context.Context,
	product *domain.Product,
	idempotencyKey string,
) error {
	return s.ExecuteInTransaction(ctx, func(uow ports.UnitOfWork) error {
		if err := uow.ProductRepository().Delete(ctx, product.ID); err != nil {
			if errors.Is(err, domain.ErrProductNotFound) {
				return fmt.Errorf("product not found: %w", domain.ErrProductNotFound)
			}
			return NewTransactionError("delete product", err)
		}

		events := product.DomainEvents()
		if len(events) == 0 {
			return fmt.Errorf("no domain events found in product - event should be recorded in Use Case layer")
		}

		outboxRepo := uow.OutboxRepository()
		if err := s.publishDomainEventsBatch(ctx, events, idempotencyKey, outboxRepo); err != nil {
			s.logger.Error("Failed to save events to outbox",
				ports.NewField("error", err),
				ports.NewField("product_id", product.ID),
			)
			if len(events) > 0 {
				return NewEventPublishError(product.ID, events[0].EventType(), err)
			}
			return NewTransactionError("publish events", err)
		}

		product.ClearDomainEvents()

		s.logger.Info("Product deleted successfully",
			ports.NewField("product_id", product.ID),
		)

		return nil
	})
}

func (s *ProductService) publishDomainEventsBatch(
	ctx context.Context,
	events []domain.DomainEvent,
	idempotencyKey string,
	outboxRepo ports.OutboxRepository,
) error {
	if batchRepo, ok := outboxRepo.(ports.BatchOutboxRepository); ok {
		outboxEvents := make([]*ports.OutboxEvent, 0, len(events))
		for _, event := range events {
			eventDataJSON, err := event.MarshalJSON()
			if err != nil {
				return fmt.Errorf("failed to marshal event data: %w", err)
			}

			outboxEvents = append(outboxEvents, &ports.OutboxEvent{
				EventType:      event.EventType(),
				EventData:      eventDataJSON,
				IdempotencyKey: idempotencyKey,
				Status:         ports.OutboxStatusPending,
			})
		}

		if s.metrics != nil {
			s.metrics.RecordBatchSize("save_events", len(outboxEvents))
		}

		return batchRepo.SaveEventsBatch(ctx, outboxEvents)
	}

		for _, event := range events {
			if err := s.eventPublisher.PublishDomainEventWithIdempotencyKey(ctx, event, idempotencyKey, outboxRepo); err != nil {
				return NewEventPublishError(0, event.EventType(), err)
			}
		}

	return nil
}

