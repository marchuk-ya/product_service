package middleware

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"product_service/products/internal/usecase/ports"
)

type HealthChecker struct {
	db                   *sql.DB
	publisherHealthChecker ports.EventPublisherHealthChecker
	logger               ports.Logger
	timeout              time.Duration
}

func NewHealthChecker(db *sql.DB, publisherHealthChecker ports.EventPublisherHealthChecker, logger ports.Logger) *HealthChecker {
	return &HealthChecker{
		db:                   db,
		publisherHealthChecker: publisherHealthChecker,
		logger:               logger,
		timeout:              5 * time.Second,
	}
}

func (h *HealthChecker) HealthCheckHandler(c *gin.Context) {
	status := http.StatusOK
	checks := make(map[string]string)

	ctx, cancel := context.WithTimeout(c.Request.Context(), h.timeout)
	defer cancel()

	dbHealthy := h.checkDatabase(ctx)
	if dbHealthy {
		checks["database"] = "healthy"
	} else {
		status = http.StatusServiceUnavailable
		checks["database"] = "unhealthy"
	}

	publisherHealthy := h.checkEventPublisher(ctx)
	if publisherHealthy {
		checks["event_publisher"] = "healthy"
	} else {
		status = http.StatusServiceUnavailable
		checks["event_publisher"] = "unhealthy"
	}

	response := gin.H{
		"status": "ok",
		"checks": checks,
	}

	if status != http.StatusOK {
		response["status"] = "degraded"
	}

	c.JSON(status, response)
}

func (h *HealthChecker) checkDatabase(ctx context.Context) bool {
	if h.db == nil {
		h.logger.Warn("Database connection is nil")
		return false
	}

	done := make(chan bool, 1)
	go func() {
		err := h.db.PingContext(ctx)
		done <- err == nil
	}()

	select {
	case healthy := <-done:
		if !healthy {
			h.logger.Warn("Health check failed: database ping failed")
		}
		return healthy
	case <-ctx.Done():
		h.logger.Warn("Health check failed: database ping timeout",
			ports.NewField("error", ctx.Err()),
		)
		return false
	}
}

func (h *HealthChecker) checkEventPublisher(ctx context.Context) bool {
	if h.publisherHealthChecker == nil {
		h.logger.Warn("Event publisher health checker is nil")
		return false
	}

	done := make(chan bool, 1)
	go func() {
		healthy := h.publisherHealthChecker.IsHealthy(ctx)
		done <- healthy
	}()

	select {
	case healthy := <-done:
		if !healthy {
			h.logger.Warn("Health check failed: event publisher is unhealthy")
		}
		return healthy
	case <-ctx.Done():
		h.logger.Warn("Health check failed: event publisher check timeout",
			ports.NewField("error", ctx.Err()),
		)
		return false
	}
}

