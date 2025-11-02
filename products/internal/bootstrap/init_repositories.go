package bootstrap

import (
	"context"
	"database/sql"
	"fmt"

	"product_service/products/internal/repository"
	"product_service/products/internal/usecase/ports"
)

func initRepositories(db *sql.DB, metrics ports.MetricsCollector, maxBatchSize int) (ports.ProductRepository, ports.OutboxRepository, *repository.PreparedStatements, *repository.PreparedStatements, error) {
	productStm, err := repository.PrepareProductStatements(context.Background(), db)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("failed to prepare product statements: %w", err)
	}

	outboxStm, err := repository.PrepareOutboxStatements(context.Background(), db)
	if err != nil {
		productStm.Close()
		return nil, nil, nil, nil, fmt.Errorf("failed to prepare outbox statements: %w", err)
	}

	baseProductRepo := repository.NewPostgresProductRepository(db, productStm)
	
	productRepo := repository.NewMetricsProductRepositoryDecorator(baseProductRepo, metrics)
	
	outboxRepo := repository.NewPostgresOutboxRepositoryWithConfig(db, outboxStm, maxBatchSize)

	return productRepo, outboxRepo, productStm, outboxStm, nil
}

func initUnitOfWorkFactory(
	db *sql.DB,
	productStm *repository.PreparedStatements,
	outboxStm *repository.PreparedStatements,
	metrics ports.MetricsCollector,
) ports.UoWFactory {
	return repository.NewUoWFactory(db, productStm, outboxStm, metrics)
}

