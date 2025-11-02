.PHONY: generate-mocks test test-cover lint staticcheck clean

generate-mocks:
	@echo "Generating mocks..."
	@mkdir -p mocks
	mockgen -source=internal/usecase/ports/product_repository.go -destination=mocks/mock_product_repository.go -package=mocks
	mockgen -source=internal/usecase/ports/outbox_repository.go -destination=mocks/mock_outbox_repository.go -package=mocks
	mockgen -source=internal/usecase/ports/unit_of_work.go -destination=mocks/mock_unit_of_work.go -package=mocks
	mockgen -source=internal/usecase/ports/domain_event_publisher.go -destination=mocks/mock_domain_event_publisher.go -package=mocks
	mockgen -source=internal/usecase/ports/event_publisher.go -destination=mocks/mock_event_publisher.go -package=mocks
	mockgen -source=internal/usecase/ports/logger.go -destination=mocks/mock_logger.go -package=mocks
	mockgen -source=internal/usecase/ports/metrics_collector.go -destination=mocks/mock_metrics_collector.go -package=mocks
	mockgen -source=internal/usecase/ports/application_service.go -destination=mocks/mock_application_service.go -package=mocks
	mockgen -source=internal/usecase/ports/retry.go -destination=mocks/mock_retry.go -package=mocks
	@echo "Mocks generated successfully!"

test:
	go test ./...

install-mockgen:
	go install go.uber.org/mock/mockgen@latest

