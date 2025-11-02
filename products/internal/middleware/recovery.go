package middleware

import (
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
	"product_service/products/internal/usecase/ports"
)

func RecoveryMiddleware(logger ports.Logger) gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		logger.Error("Panic recovered",
			ports.NewField("error", recovered),
			ports.NewField("path", c.Request.URL.Path),
			ports.NewField("method", c.Request.Method),
			ports.NewField("stack", string(debug.Stack())),
		)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Internal server error",
		})
	})
}

