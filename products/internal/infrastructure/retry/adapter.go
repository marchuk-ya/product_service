package retry

import (
	"context"
	"product_service/products/internal/usecase/ports"
	"time"
)

type RetryAdapter struct{}

func NewRetryAdapter() ports.Retrier {
	return &RetryAdapter{}
}

func (r *RetryAdapter) Do(ctx context.Context, cfg ports.RetryConfig, fn func() error) error {
	infraCfg := Config{
		MaxAttempts:  cfg.MaxAttempts,
		BaseBackoff:  cfg.BaseBackoff,
		MaxBackoff:   cfg.MaxBackoff,
		InitialDelay: cfg.InitialDelay,
	}

	return Do(ctx, infraCfg, fn)
}

func DefaultRetryConfig() ports.RetryConfig {
	return ports.RetryConfig{
		MaxAttempts:  3,
		BaseBackoff:  100 * time.Millisecond,
		MaxBackoff:   1 * time.Second,
		InitialDelay: 50 * time.Millisecond,
	}
}

