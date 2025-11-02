package ports

import (
	"context"
	"product_service/products/internal/domain"
)

type DomainEventPublisher interface {
	PublishDomainEvent(ctx context.Context, event domain.DomainEvent, outboxRepo OutboxRepository) error

	PublishDomainEventWithIdempotencyKey(
		ctx context.Context,
		event domain.DomainEvent,
		idempotencyKey string,
		outboxRepo OutboxRepository,
	) error
}

