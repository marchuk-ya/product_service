package messaging

import (
	"context"
	"time"
)

type contextKey string

const (
	TimestampKey contextKey = "timestamp"
)

func WithTimestamp(ctx context.Context, timestamp time.Time) context.Context {
	return context.WithValue(ctx, TimestampKey, timestamp)
}

func GetTimestamp(ctx context.Context) (time.Time, bool) {
	ts, ok := ctx.Value(TimestampKey).(time.Time)
	return ts, ok
}

