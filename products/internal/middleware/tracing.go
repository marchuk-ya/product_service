package middleware

import (
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
)

func TracingMiddleware() gin.HandlerFunc {
	return otelgin.Middleware("product_service")
}
