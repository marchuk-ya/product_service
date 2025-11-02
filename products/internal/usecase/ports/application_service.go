package ports

import (
	"context"
	"product_service/products/internal/domain"
)

type ProductApplicationService interface {
	CreateProductWithEvent(ctx context.Context, product *domain.Product, idempotencyKey string) error
	
	DeleteProductWithEvent(ctx context.Context, product *domain.Product, idempotencyKey string) error
}

type UoWFactory interface {
	CreateUnitOfWork() UnitOfWork
}

type BatchOutboxRepository interface {
	OutboxRepository
	
	SaveEventsBatch(ctx context.Context, events []*OutboxEvent) error
}

