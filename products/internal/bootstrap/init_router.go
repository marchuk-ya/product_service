package bootstrap

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"product_service/products/internal/config"
	"product_service/products/internal/handler"
	"product_service/products/internal/infrastructure/tracing"
	"product_service/products/internal/middleware"
	"product_service/products/internal/usecase/ports"
)

func initRouter(
	productHandler *handler.GinProductHandler,
	healthChecker *middleware.HealthChecker,
	rateLimiter *middleware.RateLimiter,
	metricsCollector ports.MetricsCollector,
	tracerProvider *tracing.TracerProvider,
	handlerLogger ports.Logger,
	appConfig *config.AppConfig,
) (*gin.Engine, *http.Server) {
	router := gin.New()

	if tracerProvider != nil {
		router.Use(middleware.TracingMiddleware())
	}

	router.Use(middleware.RequestIDMiddleware())
	router.Use(middleware.LoggingMiddleware(handlerLogger))
	router.Use(middleware.RecoveryMiddleware(handlerLogger))
	router.Use(middleware.MetricsMiddleware(metricsCollector))
	router.Use(rateLimiter.RateLimitMiddleware())

	router.GET("/health", healthChecker.HealthCheckHandler)

	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	v1 := router.Group("/api/v1")
	{
		v1.POST("/products", productHandler.CreateProduct)
		v1.GET("/products", productHandler.GetProducts)
		v1.DELETE("/products/:id", productHandler.DeleteProduct)
	}

	httpServer := &http.Server{
		Addr:    ":" + appConfig.Server.Port,
		Handler: router,
	}

	return router, httpServer
}

