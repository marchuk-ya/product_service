package retry

import (
	"context"
	"fmt"
	"time"
)

type Config struct {
	MaxAttempts   int
	BaseBackoff   time.Duration
	MaxBackoff    time.Duration
	InitialDelay  time.Duration
}

func DefaultConfig() Config {
	return Config{
		MaxAttempts:  5,
		BaseBackoff:  1 * time.Second,
		MaxBackoff:   30 * time.Second,
		InitialDelay: 100 * time.Millisecond,
	}
}

func Do(ctx context.Context, cfg Config, fn func() error) error {
	var lastErr error
	
	for attempt := 0; attempt < cfg.MaxAttempts; attempt++ {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		
		err := fn()
		if err == nil {
			return nil
		}
		
		lastErr = err
		
		if attempt == cfg.MaxAttempts-1 {
			break
		}
		
		delay := calculateBackoff(cfg, attempt)
		
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
		}
	}
	
	return fmt.Errorf("failed after %d attempts: %w", cfg.MaxAttempts, lastErr)
}

func calculateBackoff(cfg Config, attempt int) time.Duration {
	if attempt == 0 {
		return cfg.InitialDelay
	}
	
	delay := cfg.BaseBackoff * time.Duration(1<<uint(attempt-1))
	
	if delay > cfg.MaxBackoff {
		delay = cfg.MaxBackoff
	}
	
	return delay
}

