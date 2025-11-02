package ports

import (
	"context"
	"product_service/products/internal/domain"
)

type ProductRepository interface {
	Create(ctx context.Context, product *domain.Product) error
	GetByID(ctx context.Context, id int) (*domain.Product, error)
	List(ctx context.Context, page, limit int) ([]domain.Product, int, error)
	Delete(ctx context.Context, id int) error
}

