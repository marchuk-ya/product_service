package usecase

import (
	"context"
	"errors"
	"fmt"
	"product_service/products/internal/domain"
	"product_service/products/internal/domain/services"
	"product_service/products/internal/usecase/ports"
)

type ProductUseCase interface {
	CreateProduct(ctx context.Context, name string, price float64, idempotencyKey string) (*domain.Product, error)
	GetProducts(ctx context.Context, page, limit int) ([]domain.Product, int, error)
	DeleteProduct(ctx context.Context, id int, idempotencyKey string) error
}

type Shutdownable interface {
	Shutdown(ctx context.Context) error
}

type productUseCase struct {
	repo          ports.ProductRepository
	appService    ports.ProductApplicationService
	domainService services.ProductDomainService
	logger        ports.Logger
}

var _ Shutdownable = (*productUseCase)(nil)

func NewProductUseCase(
	repo ports.ProductRepository,
	appService ports.ProductApplicationService,
	domainService services.ProductDomainService,
	logger ports.Logger,
) ProductUseCase {
	return &productUseCase{
		repo:          repo,
		appService:    appService,
		domainService: domainService,
		logger:        logger,
	}
}

func (uc *productUseCase) CreateProduct(ctx context.Context, name string, price float64, idempotencyKey string) (*domain.Product, error) {
	if err := uc.domainService.ValidateProductForCreation(name, price); err != nil {
		uc.logger.Warn("Product validation failed",
			ports.NewField("error", err),
			ports.NewField("name", name),
			ports.NewField("price", price),
		)
		return nil, fmt.Errorf("product validation failed: %w", err)
	}

	product, err := domain.NewProduct(name, price)
	if err != nil {
		uc.logger.Warn("Failed to create product domain entity",
			ports.NewField("error", err),
			ports.NewField("name", name),
			ports.NewField("price", price),
		)
		return nil, fmt.Errorf("failed to create product: %w", err)
	}

	if err := uc.appService.CreateProductWithEvent(ctx, product, idempotencyKey); err != nil {
		return nil, fmt.Errorf("failed to create product: %w", err)
	}

	return product, nil
}

func (uc *productUseCase) GetProducts(ctx context.Context, page, limit int) ([]domain.Product, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	products, total, err := uc.repo.List(ctx, page, limit)
	if err != nil {
		uc.logger.Error("Failed to list products from repository",
			ports.NewField("error", err),
			ports.NewField("page", page),
			ports.NewField("limit", limit),
		)
		return nil, 0, fmt.Errorf("failed to list products from repository: %w", err)
	}

	uc.logger.Debug("Products listed successfully",
		ports.NewField("count", len(products)),
		ports.NewField("total", total),
		ports.NewField("page", page),
	)

	return products, total, nil
}

func (uc *productUseCase) DeleteProduct(ctx context.Context, id int, idempotencyKey string) error {
	if id <= 0 {
		uc.logger.Warn("Invalid product ID for deletion",
			ports.NewField("product_id", id),
		)
		return fmt.Errorf("invalid product id: %w", domain.ErrInvalidInput)
	}

	product, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, domain.ErrProductNotFound) {
			uc.logger.Warn("Product not found for deletion",
				ports.NewField("product_id", id),
			)
			return fmt.Errorf("product not found: %w", domain.ErrProductNotFound)
		}
		uc.logger.Error("Failed to get product for deletion",
			ports.NewField("error", err),
			ports.NewField("product_id", id),
		)
		return fmt.Errorf("failed to get product: %w", err)
	}

	if err := uc.domainService.CanDeleteProduct(product); err != nil {
		uc.logger.Warn("Product deletion validation failed",
			ports.NewField("error", err),
			ports.NewField("product_id", id),
		)
		return fmt.Errorf("cannot delete product: %w", err)
	}

	product.RecordDeleteEvent()

	if err := uc.appService.DeleteProductWithEvent(ctx, product, idempotencyKey); err != nil {
		uc.logger.Error("Failed to delete product",
			ports.NewField("error", err),
			ports.NewField("product_id", id),
		)
		return fmt.Errorf("failed to delete product: %w", err)
	}

	return nil
}

func (uc *productUseCase) Shutdown(ctx context.Context) error {
	return nil
}

