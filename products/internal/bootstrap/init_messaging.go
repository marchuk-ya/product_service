package bootstrap

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	"product_service/products/internal/config"
	"product_service/products/internal/infrastructure/messaging"
	"product_service/products/internal/usecase/ports"
)

func initMessaging(
	appConfig *config.AppConfig,
	logger *zap.Logger,
	outboxRepo ports.OutboxRepository,
	metrics ports.MetricsCollector,
) (ports.EventPublisher, *messaging.OutboxWorker, error) {
	rmqCtx, rmqCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer rmqCancel()

	publisher, err := messaging.NewRabbitMQPublisher(
		rmqCtx,
		appConfig.RabbitMQ.URL(),
		appConfig.RabbitMQ.Exchange,
		logger,
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to initialize RabbitMQ publisher: %w", err)
	}

	workerCtx := context.Background()
	outboxWorker := messaging.NewOutboxWorker(
		outboxRepo,
		publisher,
		logger,
		appConfig.Outbox.Interval,
		appConfig.Outbox.BatchSize,
		appConfig.Outbox.MaxRetries,
		appConfig.Outbox.BaseBackoff,
		appConfig.Outbox.MaxBackoff,
		metrics,
		appConfig.Outbox.Concurrency,
	)
	outboxWorker.Start(workerCtx)

	return publisher, outboxWorker, nil
}

