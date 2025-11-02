package ports

import (
	"context"
	"time"
)

type OutboxEvent struct {
	ID             int64
	EventType      string
	EventData      []byte
	IdempotencyKey string
	CreatedAt      time.Time
	PublishedAt    *time.Time
	RetryCount     int
	Status         OutboxStatus
}

type OutboxStatus string

const (
	OutboxStatusPending   OutboxStatus = "pending"
	OutboxStatusPublished OutboxStatus = "published"
	OutboxStatusFailed    OutboxStatus = "failed"
	OutboxStatusDLQ       OutboxStatus = "dlq"
)

type OutboxRepository interface {
	SaveEvent(ctx context.Context, event *OutboxEvent) error
	GetPendingEvents(ctx context.Context, limit int) ([]OutboxEvent, error)
	MarkAsPublished(ctx context.Context, eventID int64) error
	MarkAsFailed(ctx context.Context, eventID int64, retryCount int) error
	MoveToDLQ(ctx context.Context, eventID int64, reason string) error
	CheckIdempotencyKey(ctx context.Context, idempotencyKey string) (bool, error)
	
}

