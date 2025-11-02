package handler

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"product_service/products/internal/handler/dto"
	"product_service/products/internal/usecase"
	"product_service/products/internal/usecase/ports"
	"strconv"
	"time"
)

const (
	maxRequestBodySize = 1 << 20
)

type HTTPProductHandler struct {
	useCase        usecase.ProductUseCase
	logger         ports.Logger
	metrics        ports.MetricsCollector
	errorMapper    *ErrorMapper
	requestTimeout time.Duration
	readTimeout    time.Duration
}

func NewHTTPProductHandler(useCase usecase.ProductUseCase, logger ports.Logger, metrics ports.MetricsCollector, requestTimeout, readTimeout time.Duration) *HTTPProductHandler {
	return &HTTPProductHandler{
		useCase:        useCase,
		logger:         logger,
		metrics:        metrics,
		errorMapper:    NewErrorMapper(logger),
		requestTimeout: requestTimeout,
		readTimeout:    readTimeout,
	}
}

func (h *HTTPProductHandler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	limitedBody := io.LimitReader(r.Body, maxRequestBodySize)
	defer r.Body.Close()

	var req dto.CreateProductRequest
	if err := json.NewDecoder(limitedBody).Decode(&req); err != nil {
		h.logger.Warn("Invalid request body",
			ports.NewField("error", err),
		)
		h.writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := ValidateCreateProductRequest(req); err != nil {
		h.logger.Warn("Invalid product data",
			ports.NewField("error", err),
		)
		h.writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	idempotencyKey := r.Header.Get("Idempotency-Key")

	ctx, cancel := context.WithTimeout(r.Context(), h.requestTimeout)
	defer cancel()

	product, err := h.useCase.CreateProduct(ctx, req.Name, req.Price, idempotencyKey)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			h.handleContextError(ctx, w, "create_product", err)
			return
		}
		h.errorMapper.MapToHTTPError(w, err, ctx)
		return
	}

	response := dto.ToProductResponse(product)

	if h.metrics != nil {
		h.metrics.IncrementProductsCreated()
	}

	h.logger.Info("Product created",
		ports.NewField("id", product.ID),
		ports.NewField("name", product.Name.Value()),
	)
	h.writeJSON(w, http.StatusCreated, response)
}

func (h *HTTPProductHandler) GetProducts(w http.ResponseWriter, r *http.Request) {
	page, limit := ParsePaginationParams(r)

	ctx, cancel := context.WithTimeout(r.Context(), h.readTimeout)
	defer cancel()

	products, total, err := h.useCase.GetProducts(ctx, page, limit)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			h.handleContextError(ctx, w, "get_products", err)
			return
		}
		h.errorMapper.MapToHTTPError(w, err, ctx)
		return
	}

	response := dto.ProductListResponse{
		Products: dto.ToProductResponseList(products),
		Page:     page,
		Limit:    limit,
		Total:    total,
	}

	h.writeJSON(w, http.StatusOK, response)
}

func (h *HTTPProductHandler) DeleteProduct(idStr string, w http.ResponseWriter, r *http.Request) {
	if idStr == "" {
		h.writeError(w, http.StatusBadRequest, "Product ID is required")
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.logger.Warn("Invalid product ID",
			ports.NewField("id", idStr),
			ports.NewField("error", err),
		)
		h.writeError(w, http.StatusBadRequest, "Invalid product ID")
		return
	}

	idempotencyKey := r.Header.Get("Idempotency-Key")

	ctx, cancel := context.WithTimeout(r.Context(), h.requestTimeout)
	defer cancel()

	err = h.useCase.DeleteProduct(ctx, id, idempotencyKey)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			h.handleContextError(ctx, w, "delete_product", err, ports.NewField("product_id", id))
			return
		}
		h.errorMapper.MapToHTTPError(w, err, ctx)
		return
	}

	if h.metrics != nil {
		h.metrics.IncrementProductsDeleted()
	}

	h.logger.Info("Product deleted",
		ports.NewField("id", id),
	)
	h.writeJSON(w, http.StatusOK, map[string]string{"message": "Product deleted successfully"})
}


func (h *HTTPProductHandler) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.Error("Failed to encode JSON response",
			ports.NewField("error", err),
		)
	}
}

func (h *HTTPProductHandler) handleContextError(ctx context.Context, w http.ResponseWriter, operation string, err error, additionalFields ...ports.Field) {
	fields := []ports.Field{
		ports.NewField("operation", operation),
		ports.NewField("error", err),
	}
	fields = append(fields, additionalFields...)
	
	h.logger.Warn("Request timeout or cancelled", fields...)
	h.writeError(w, http.StatusRequestTimeout, "Request timeout")
}

func (h *HTTPProductHandler) writeError(w http.ResponseWriter, status int, message string) {
	h.writeJSON(w, status, map[string]string{"error": message})
}

