package bootstrap

import (
	"database/sql"
	"time"

	"product_service/products/internal/middleware"
	"product_service/products/internal/usecase/ports"
)

func initMiddleware(
	db *sql.DB,
	publisher ports.EventPublisher,
	handlerLogger ports.Logger,
	_ ports.MetricsCollector,
) (*middleware.HealthChecker, *middleware.RateLimiter) {
	var publisherHealthChecker ports.EventPublisherHealthChecker
	if healthChecker, ok := publisher.(ports.EventPublisherHealthChecker); ok {
		publisherHealthChecker = healthChecker
	}
	healthChecker := middleware.NewHealthChecker(db, publisherHealthChecker, handlerLogger)

	rateLimiter := middleware.NewRateLimiter(100, 1*time.Minute, handlerLogger)

	return healthChecker, rateLimiter
}

