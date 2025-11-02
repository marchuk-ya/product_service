package usecase

import (
	"context"
	"errors"
	"fmt"
	"product_service/products/internal/domain"
	domainServices "product_service/products/internal/domain/services"
	"product_service/products/mocks"
	"testing"

	"go.uber.org/mock/gomock"
)

func TestProductUseCase_CreateProduct_WithGeneratedMocks(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockProductRepository(ctrl)
	mockAppService := mocks.NewMockProductApplicationService(ctrl)
	mockLogger := mocks.NewMockLogger(ctrl)

	domainService := domainServices.NewProductDomainService(nil)

	useCase := NewProductUseCase(
		mockRepo,
		mockAppService,
		domainService,
		mockLogger,
	)

	ctx := context.Background()
	name := "Test Product"
	price := 99.99
	idempotencyKey := "test-key-123"

	mockAppService.EXPECT().
		CreateProductWithEvent(ctx, gomock.Any(), idempotencyKey).
		DoAndReturn(func(ctx context.Context, p *domain.Product, key string) error {
			p.ID = 1
			return nil
		})

	result, err := useCase.CreateProduct(ctx, name, price, idempotencyKey)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if result.ID == 0 {
		t.Error("Expected product to have an ID")
	}

	if result.Name.Value() != name {
		t.Errorf("Expected name %s, got %s", name, result.Name.Value())
	}

	if result.Price.Value() != price {
		t.Errorf("Expected price %f, got %f", price, result.Price.Value())
	}
}

func TestProductUseCase_CreateProduct_InvalidInput_WithGeneratedMocks(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockProductRepository(ctrl)
	mockAppService := mocks.NewMockProductApplicationService(ctrl)
	mockLogger := mocks.NewMockLogger(ctrl)
	domainService := domainServices.NewProductDomainService(nil)

	useCase := NewProductUseCase(
		mockRepo,
		mockAppService,
		domainService,
		mockLogger,
	)

	ctx := context.Background()

	mockLogger.EXPECT().Warn(gomock.Any(), gomock.Any()).AnyTimes()

	_, err := useCase.CreateProduct(ctx, "", 99.99, "")
	if err == nil {
		t.Error("Expected error for empty name")
	}

	_, err = useCase.CreateProduct(ctx, "Test", -10, "")
	if err == nil {
		t.Error("Expected error for negative price")
	}

	_, err = useCase.CreateProduct(ctx, "Test", 0, "")
	if err == nil {
		t.Error("Expected error for zero price")
	}
}

func TestProductUseCase_GetProducts_WithGeneratedMocks(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockProductRepository(ctrl)
	mockAppService := mocks.NewMockProductApplicationService(ctrl)
	mockLogger := mocks.NewMockLogger(ctrl)
	domainService := domainServices.NewProductDomainService(nil)

	useCase := NewProductUseCase(
		mockRepo,
		mockAppService,
		domainService,
		mockLogger,
	)

	ctx := context.Background()

	testProducts := make([]domain.Product, 15)
	for i := 0; i < 15; i++ {
		product, _ := domain.NewProduct(fmt.Sprintf("Product %d", i+1), float64((i+1)*10))
		product.ID = i + 1
		testProducts[i] = *product
	}

	mockLogger.EXPECT().Debug(gomock.Any(), gomock.Any()).AnyTimes()

	mockRepo.EXPECT().
		List(ctx, 1, 10).
		Return(testProducts[0:10], 15, nil)

	products, total, err := useCase.GetProducts(ctx, 1, 10)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if total != 15 {
		t.Errorf("Expected total 15, got %d", total)
	}

	if len(products) != 10 {
		t.Errorf("Expected 10 products, got %d", len(products))
	}

	mockRepo.EXPECT().
		List(ctx, 2, 10).
		Return(testProducts[10:15], 15, nil)

	products, total, err = useCase.GetProducts(ctx, 2, 10)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(products) != 5 {
		t.Errorf("Expected 5 products on page 2, got %d", len(products))
	}
}

func TestProductUseCase_DeleteProduct_WithGeneratedMocks(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockProductRepository(ctrl)
	mockAppService := mocks.NewMockProductApplicationService(ctrl)
	mockLogger := mocks.NewMockLogger(ctrl)
	domainService := domainServices.NewProductDomainService(nil)

	useCase := NewProductUseCase(
		mockRepo,
		mockAppService,
		domainService,
		mockLogger,
	)

	ctx := context.Background()
	productID := 1

	product, err := domain.NewProduct("Test Product", 99.99)
	if err != nil {
		t.Fatalf("Failed to create product: %v", err)
	}
	product.ID = productID

	mockRepo.EXPECT().
		GetByID(ctx, productID).
		Return(product, nil)

	product.RecordDeleteEvent()

	mockAppService.EXPECT().
		DeleteProductWithEvent(ctx, product, "").
		Return(nil)

	err = useCase.DeleteProduct(ctx, productID, "")
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
}

func TestProductUseCase_DeleteProduct_NotFound_WithGeneratedMocks(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockProductRepository(ctrl)
	mockAppService := mocks.NewMockProductApplicationService(ctrl)
	mockLogger := mocks.NewMockLogger(ctrl)
	domainService := domainServices.NewProductDomainService(nil)

	useCase := NewProductUseCase(
		mockRepo,
		mockAppService,
		domainService,
		mockLogger,
	)

	ctx := context.Background()

	mockLogger.EXPECT().Warn(gomock.Any(), gomock.Any()).AnyTimes()

	mockRepo.EXPECT().
		GetByID(ctx, 999).
		Return(nil, domain.ErrProductNotFound)

	err := useCase.DeleteProduct(ctx, 999, "")
	if !errors.Is(err, domain.ErrProductNotFound) {
		t.Errorf("Expected ErrProductNotFound, got: %v", err)
	}
}

