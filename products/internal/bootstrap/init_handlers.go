package bootstrap

import (
	"product_service/products/internal/config"
	"product_service/products/internal/handler"
	"product_service/products/internal/infrastructure/logging"
	"product_service/products/internal/usecase"
	"product_service/products/internal/usecase/ports"

	"go.uber.org/zap"
)

func initHandlers(
	productUseCase usecase.ProductUseCase,
	handlerLogger ports.Logger,
	metricsCollector ports.MetricsCollector,
	appConfig *config.AppConfig,
) *handler.GinProductHandler {
	return handler.NewGinProductHandler(
		productUseCase,
		handlerLogger,
		metricsCollector,
		appConfig.Server.RequestTimeout,
		appConfig.Server.ReadTimeout,
	)
}

func initLoggerAdapters(logger *zap.Logger) ports.Logger {
	return logging.NewZapLoggerAdapter(logger)
}

