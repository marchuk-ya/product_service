package repository

import (
	"encoding/json"
	"time"
)

type OutboxEvent struct {
	ID             int64
	EventType      string
	EventData      json.RawMessage
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

