package messaging

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"product_service/products/internal/domain"
	"product_service/products/internal/usecase/ports"
)

var _ ports.DomainEventPublisher = (*TransactionalEventPublisher)(nil)

type TransactionalEventPublisher struct {
	publisher ports.EventPublisher
}

func NewTransactionalEventPublisher(
	publisher ports.EventPublisher,
) *TransactionalEventPublisher {
	return &TransactionalEventPublisher{
		publisher: publisher,
	}
}

func (p *TransactionalEventPublisher) PublishDomainEvent(
	ctx context.Context,
	event domain.DomainEvent,
	outboxRepo ports.OutboxRepository,
) error {
	eventDataJSON, err := event.MarshalJSON()
	if err != nil {
		return fmt.Errorf("failed to marshal event data: %w", err)
	}

	idempotencyKey, err := generateIdempotencyKey()
	if err != nil {
		return fmt.Errorf("failed to generate idempotency key: %w", err)
	}

	outboxEvent := &ports.OutboxEvent{
		EventType:      event.EventType(),
		EventData:      eventDataJSON,
		IdempotencyKey: idempotencyKey,
		Status:         ports.OutboxStatusPending,
	}

	return outboxRepo.SaveEvent(ctx, outboxEvent)
}

func (p *TransactionalEventPublisher) PublishDomainEventWithIdempotencyKey(
	ctx context.Context,
	event domain.DomainEvent,
	idempotencyKey string,
	outboxRepo ports.OutboxRepository,
) error {
	eventDataJSON, err := event.MarshalJSON()
	if err != nil {
		return fmt.Errorf("failed to marshal event data: %w", err)
	}

	if idempotencyKey == "" {
		idempotencyKey, err = generateIdempotencyKey()
		if err != nil {
			return fmt.Errorf("failed to generate idempotency key: %w", err)
		}
	}

	outboxEvent := &ports.OutboxEvent{
		EventType:      event.EventType(),
		EventData:      eventDataJSON,
		IdempotencyKey: idempotencyKey,
		Status:         ports.OutboxStatusPending,
	}

	err = outboxRepo.SaveEvent(ctx, outboxEvent)
	if err != nil {
		return err
	}

	if outboxEvent.ID == 0 && idempotencyKey != "" {
		return nil
	}

	return nil
}

func generateIdempotencyKey() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}
	return hex.EncodeToString(bytes), nil
}

