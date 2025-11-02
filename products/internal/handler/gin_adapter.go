package handler

import (
	"time"
	
	"github.com/gin-gonic/gin"
	"product_service/products/internal/usecase"
	"product_service/products/internal/usecase/ports"
)

type GinProductHandler struct {
	httpHandler *HTTPProductHandler
}

func NewGinProductHandler(useCase usecase.ProductUseCase, logger ports.Logger, metrics ports.MetricsCollector, requestTimeout, readTimeout time.Duration) *GinProductHandler {
	return &GinProductHandler{
		httpHandler: NewHTTPProductHandler(useCase, logger, metrics, requestTimeout, readTimeout),
	}
}

func (h *GinProductHandler) CreateProduct(c *gin.Context) {
	h.httpHandler.CreateProduct(c.Writer, c.Request)
}

func (h *GinProductHandler) GetProducts(c *gin.Context) {
	h.httpHandler.GetProducts(c.Writer, c.Request)
}

func (h *GinProductHandler) DeleteProduct(c *gin.Context) {
	id := c.Param("id")
	h.httpHandler.DeleteProduct(id, c.Writer, c.Request)
}

