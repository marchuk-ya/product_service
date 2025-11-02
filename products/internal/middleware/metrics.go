package middleware

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"product_service/products/internal/usecase/ports"
)

func MetricsMiddleware(metricsCollector ports.MetricsCollector) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		duration := time.Since(start)

		method := c.Request.Method
		endpoint := c.FullPath()
		if endpoint == "" {
			endpoint = c.Request.URL.Path
		}
		status := strconv.Itoa(c.Writer.Status())

		metricsCollector.RecordRequestDuration(method, endpoint, status, duration)
		metricsCollector.IncrementRequestCount(method, endpoint, status)
	}
}

func LoggingMiddleware(logger ports.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		duration := time.Since(start)
		requestID := GetRequestID(c)

		logger.Info("HTTP request",
			ports.NewField("request_id", requestID),
			ports.NewField("method", c.Request.Method),
			ports.NewField("path", c.Request.URL.Path),
			ports.NewField("status", c.Writer.Status()),
			ports.NewField("duration", duration),
			ports.NewField("ip", c.ClientIP()),
		)
	}
}

