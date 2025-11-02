package logging

import (
	"product_service/products/internal/usecase/ports"
	"go.uber.org/zap"
)

var _ ports.Logger = (*zapLoggerAdapter)(nil)

type zapLoggerAdapter struct {
	logger *zap.Logger
}

func NewZapLoggerAdapter(logger *zap.Logger) ports.Logger {
	return &zapLoggerAdapter{
		logger: logger,
	}
}

func (l *zapLoggerAdapter) Info(msg string, fields ...ports.Field) {
	l.logger.Info(msg, l.convertFields(fields)...)
}

func (l *zapLoggerAdapter) Warn(msg string, fields ...ports.Field) {
	l.logger.Warn(msg, l.convertFields(fields)...)
}

func (l *zapLoggerAdapter) Error(msg string, fields ...ports.Field) {
	l.logger.Error(msg, l.convertFields(fields)...)
}

func (l *zapLoggerAdapter) Debug(msg string, fields ...ports.Field) {
	l.logger.Debug(msg, l.convertFields(fields)...)
}

func (l *zapLoggerAdapter) convertFields(fields []ports.Field) []zap.Field {
	zapFields := make([]zap.Field, len(fields))
	for i, f := range fields {
		zapFields[i] = zap.Any(f.Key, f.Value)
	}
	return zapFields
}

