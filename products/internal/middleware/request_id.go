package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"

	"github.com/gin-gonic/gin"
)

const requestIDKey = "request_id"

func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = generateRequestID()
		}

		c.Set(requestIDKey, requestID)

		c.Header("X-Request-ID", requestID)

		c.Request = c.Request.WithContext(c.Request.Context())

		c.Next()
	}
}

func GetRequestID(c *gin.Context) string {
	if id, exists := c.Get(requestIDKey); exists {
		if str, ok := id.(string); ok {
			return str
		}
	}
	return ""
}

func GetRequestIDFromContext(ctx interface{}) string {
	if c, ok := ctx.(*gin.Context); ok {
		return GetRequestID(c)
	}
	return ""
}

func generateRequestID() string {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return hex.EncodeToString([]byte(http.TimeFormat))
	}
	return hex.EncodeToString(bytes)
}

