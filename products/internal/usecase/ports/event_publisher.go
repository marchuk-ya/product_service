package ports

import "context"

type EventPublisher interface {
	PublishProductCreated(ctx context.Context, productID int) error
	PublishProductDeleted(ctx context.Context, productID int) error
	Close() error
}

type EventPublisherHealthChecker interface {
	IsHealthy(ctx context.Context) bool
}

