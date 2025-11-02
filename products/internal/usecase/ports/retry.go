package ports

import (
	"context"
	"time"
)

type RetryConfig struct {
	MaxAttempts  int
	BaseBackoff  time.Duration
	MaxBackoff   time.Duration
	InitialDelay time.Duration
}

type Retrier interface {
	Do(ctx context.Context, cfg RetryConfig, fn func() error) error
}

