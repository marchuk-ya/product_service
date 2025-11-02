package bootstrap

import (
	"go.uber.org/zap"

	"product_service/products/internal/config"
	"product_service/products/internal/infrastructure/tracing"
)

func initTracing(appConfig *config.AppConfig, logger *zap.Logger) *tracing.TracerProvider {
	if !appConfig.Tracing.Enabled {
		return nil
	}

	tp, err := tracing.NewTracerProvider(appConfig.Tracing.ServiceName, appConfig.Tracing.OTLPEndpoint)
	if err != nil {
		logger.Warn("Failed to initialize tracing", zap.Error(err))
		return nil
	}

	logger.Info("Tracing initialized",
		zap.String("service", appConfig.Tracing.ServiceName),
		zap.String("endpoint", appConfig.Tracing.OTLPEndpoint),
	)
	return tp
}

