package bootstrap

import (
	"product_service/products/internal/application"
	"product_service/products/internal/domain/services"
	"product_service/products/internal/usecase"
	"product_service/products/internal/usecase/ports"
)

func initDomainService() services.ProductDomainService {
	return services.NewProductDomainService(nil)
}

func initApplicationService(
	uowFactory ports.UoWFactory,
	eventPublisher ports.DomainEventPublisher,
	logger ports.Logger,
	retrier ports.Retrier,
	metrics ports.MetricsCollector,
) *application.ProductService {
	return application.NewProductService(
		uowFactory,
		eventPublisher,
		logger,
		retrier,
		metrics,
	)
}

func initUseCase(
	productRepo ports.ProductRepository,
	appService ports.ProductApplicationService,
	domainService services.ProductDomainService,
	logger ports.Logger,
) usecase.ProductUseCase {
	return usecase.NewProductUseCase(
		productRepo,
		appService,
		domainService,
		logger,
	)
}

