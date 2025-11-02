package repository

import (
	"context"
	"product_service/products/internal/domain"
	"product_service/products/internal/usecase/ports"
	"time"
)

type MetricsProductRepositoryDecorator struct {
	repo    ports.ProductRepository
	metrics ports.MetricsCollector
}

func NewMetricsProductRepositoryDecorator(repo ports.ProductRepository, metrics ports.MetricsCollector) ports.ProductRepository {
	if metrics == nil {
		return repo
	}
	return &MetricsProductRepositoryDecorator{
		repo:    repo,
		metrics: metrics,
	}
}

func (d *MetricsProductRepositoryDecorator) Create(ctx context.Context, product *domain.Product) error {
	start := time.Now()
	err := d.repo.Create(ctx, product)
	if d.metrics != nil {
		d.metrics.RecordDatabaseQueryDuration(time.Since(start))
	}
	return err
}

func (d *MetricsProductRepositoryDecorator) GetByID(ctx context.Context, id int) (*domain.Product, error) {
	start := time.Now()
	product, err := d.repo.GetByID(ctx, id)
	if d.metrics != nil {
		d.metrics.RecordDatabaseQueryDuration(time.Since(start))
	}
	return product, err
}

func (d *MetricsProductRepositoryDecorator) List(ctx context.Context, page, limit int) ([]domain.Product, int, error) {
	start := time.Now()
	products, total, err := d.repo.List(ctx, page, limit)
	if d.metrics != nil {
		d.metrics.RecordDatabaseQueryDuration(time.Since(start))
	}
	return products, total, err
}

func (d *MetricsProductRepositoryDecorator) Delete(ctx context.Context, id int) error {
	start := time.Now()
	err := d.repo.Delete(ctx, id)
	if d.metrics != nil {
		d.metrics.RecordDatabaseQueryDuration(time.Since(start))
	}
	return err
}

