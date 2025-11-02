package bootstrap

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"product_service/products/internal/config"
	"product_service/products/internal/infrastructure/messaging"
	"product_service/products/internal/infrastructure/metrics"
	"product_service/products/internal/infrastructure/tracing"
	"product_service/products/internal/repository"
	"product_service/products/internal/usecase"
	"product_service/products/internal/usecase/ports"
)

type App struct {
	Config        *config.AppConfig
	Logger        *zap.Logger
	DB            *config.Dependencies
	Router        *gin.Engine
	HTTPServer    *http.Server
	OutboxWorker  *messaging.OutboxWorker
	Publisher     ports.EventPublisher
	ProductStm    *repository.PreparedStatements
	OutboxStm     *repository.PreparedStatements
	ProductUseCase usecase.ProductUseCase
	TracerProvider *tracing.TracerProvider
	RateLimiter   interface{ Stop() }
}

func Bootstrap() (*App, error) {
	appConfig, err := config.LoadAppConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	logger, err := config.NewLogger()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize logger: %w", err)
	}

	deps := config.NewDependencies(logger)
	defer func() {
		if err != nil {
			deps.Close()
		}
	}()

	if err := initDatabase(appConfig, deps, logger); err != nil {
		return nil, err
	}

	metricsCollector := metrics.NewPrometheusMetrics()

	productRepo, outboxRepo, productStm, outboxStm, err := initRepositories(deps.DB, metricsCollector, appConfig.Outbox.MaxBatchSize)
	if err != nil {
		return nil, err
	}

	uowFactory := initUnitOfWorkFactory(deps.DB, productStm, outboxStm, metricsCollector)

	publisher, outboxWorker, err := initMessaging(appConfig, logger, outboxRepo, metricsCollector)
	if err != nil {
		return nil, err
	}

	transactionalEventPublisher := messaging.NewTransactionalEventPublisher(publisher)

	handlerLogger := initLoggerAdapters(logger)

	retrier := initRetrier()

	domainService := initDomainService()

	appService := initApplicationService(
		uowFactory,
		transactionalEventPublisher,
		handlerLogger,
		retrier,
		metricsCollector,
	)

	productUseCase := initUseCase(
		productRepo,
		appService,
		domainService,
		handlerLogger,
	)

	productHandler := initHandlers(productUseCase, handlerLogger, metricsCollector, appConfig)

	healthChecker, rateLimiter := initMiddleware(deps.DB, publisher, handlerLogger, metricsCollector)

	tracerProvider := initTracing(appConfig, logger)

	router, httpServer := initRouter(
		productHandler,
		healthChecker,
		rateLimiter,
		metricsCollector,
		tracerProvider,
		handlerLogger,
		appConfig,
	)

	return &App{
		Config:        appConfig,
		Logger:        logger,
		DB:            deps,
		Router:        router,
		HTTPServer:    httpServer,
		OutboxWorker:  outboxWorker,
		Publisher:     publisher,
		ProductStm:    productStm,
		OutboxStm:     outboxStm,
		ProductUseCase: productUseCase,
		TracerProvider: tracerProvider,
		RateLimiter:   rateLimiter,
	}, nil
}

func (a *App) Start() error {
	a.Logger.Info("Starting Products service", zap.String("port", a.Config.Server.Port))
	if a.HTTPServer != nil {
		return a.HTTPServer.ListenAndServe()
	}
	return fmt.Errorf("HTTP server not initialized")
}

func (a *App) Shutdown(ctx context.Context) error {
	a.Logger.Info("Shutting down server...")

	if a.TracerProvider != nil {
		a.Logger.Info("Shutting down tracer provider...")
		if err := a.TracerProvider.Shutdown(ctx); err != nil {
			a.Logger.Error("Failed to shutdown tracer provider", zap.Error(err))
		}
	}

	if a.RateLimiter != nil {
		a.Logger.Info("Stopping rate limiter...")
		a.RateLimiter.Stop()
	}

	a.Logger.Info("Stopping outbox worker...")
	if a.OutboxWorker != nil {
		a.OutboxWorker.Stop()
	}

	if shutdownable, ok := a.ProductUseCase.(usecase.Shutdownable); ok {
		if err := shutdownable.Shutdown(ctx); err != nil {
			a.Logger.Error("Failed to shutdown use case gracefully", zap.Error(err))
		}
	}

	if a.Publisher != nil {
		if err := a.Publisher.Close(); err != nil {
			a.Logger.Error("Failed to close publisher", zap.Error(err))
		}
	}

	if a.ProductStm != nil {
		if err := a.ProductStm.Close(); err != nil {
			a.Logger.Error("Failed to close product statements", zap.Error(err))
		}
	}
	if a.OutboxStm != nil {
		if err := a.OutboxStm.Close(); err != nil {
			a.Logger.Error("Failed to close outbox statements", zap.Error(err))
		}
	}

	if err := a.HTTPServer.Shutdown(ctx); err != nil {
		a.Logger.Error("Server forced to shutdown", zap.Error(err))
		return err
	}

	if err := a.DB.Close(); err != nil {
		a.Logger.Error("Failed to close database", zap.Error(err))
		return err
	}

	a.Logger.Info("Server exited")
	return nil
}

