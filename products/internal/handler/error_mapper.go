package handler

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"product_service/products/internal/domain"
	"product_service/products/internal/usecase"
	"product_service/products/internal/usecase/ports"
)

type ErrorMapper struct {
	logger ports.Logger
}

func NewErrorMapper(logger ports.Logger) *ErrorMapper {
	return &ErrorMapper{
		logger: logger,
	}
}

type HTTPErrorResponse struct {
	Error     string            `json:"error"`
	Code      string            `json:"code,omitempty"`
	Details   map[string]string `json:"details,omitempty"`
	RequestID string            `json:"request_id,omitempty"`
}

func (m *ErrorMapper) MapToHTTPError(w http.ResponseWriter, err error, ctx context.Context) {
	if err == nil {
		return
	}

	requestID := ""
	if reqID, ok := ctx.Value("request_id").(string); ok {
		requestID = reqID
	}

	if ctx.Err() == context.DeadlineExceeded {
		m.logger.Error("Request timeout",
			ports.NewField("error", err),
			ports.NewField("request_id", requestID),
		)
		m.writeErrorResponse(w, http.StatusRequestTimeout, "Request timeout", "TIMEOUT", nil, requestID)
		return
	}

	if ctx.Err() == context.Canceled {
		m.logger.Warn("Request canceled",
			ports.NewField("error", err),
			ports.NewField("request_id", requestID),
		)
		m.writeErrorResponse(w, http.StatusRequestTimeout, "Request canceled", "CANCELED", nil, requestID)
		return
	}

	if errors.Is(err, domain.ErrProductNotFound) {
		m.logger.Warn("Product not found",
			ports.NewField("error", err),
			ports.NewField("request_id", requestID),
		)
		m.writeErrorResponse(w, http.StatusNotFound, "Product not found", "PRODUCT_NOT_FOUND", nil, requestID)
		return
	}

	if errors.Is(err, domain.ErrInvalidProductName) {
		m.logger.Warn("Invalid product name",
			ports.NewField("error", err),
			ports.NewField("request_id", requestID),
		)
		m.writeErrorResponse(w, http.StatusBadRequest, "Invalid product name", "INVALID_PRODUCT_NAME", nil, requestID)
		return
	}

	if errors.Is(err, domain.ErrInvalidProductPrice) {
		m.logger.Warn("Invalid product price",
			ports.NewField("error", err),
			ports.NewField("request_id", requestID),
		)
		m.writeErrorResponse(w, http.StatusBadRequest, "Invalid product price", "INVALID_PRODUCT_PRICE", nil, requestID)
		return
	}

	if usecase.IsProductNotFound(err) {
		m.logger.Warn("Product not found",
			ports.NewField("error", err),
			ports.NewField("request_id", requestID),
		)
		m.writeErrorResponse(w, http.StatusNotFound, "Product not found", "PRODUCT_NOT_FOUND", nil, requestID)
		return
	}

	if usecase.IsInvalidProductName(err) {
		m.logger.Warn("Invalid product name",
			ports.NewField("error", err),
			ports.NewField("request_id", requestID),
		)
		message := err.Error()
		m.writeErrorResponse(w, http.StatusBadRequest, message, "INVALID_PRODUCT_NAME", nil, requestID)
		return
	}

	if usecase.IsInvalidProductPrice(err) {
		m.logger.Warn("Invalid product price",
			ports.NewField("error", err),
			ports.NewField("request_id", requestID),
		)
		message := err.Error()
		m.writeErrorResponse(w, http.StatusBadRequest, message, "INVALID_PRODUCT_PRICE", nil, requestID)
		return
	}

	if errors.Is(err, sql.ErrNoRows) {
		m.logger.Warn("Resource not found (database)",
			ports.NewField("error", err),
			ports.NewField("request_id", requestID),
		)
		m.writeErrorResponse(w, http.StatusNotFound, "Resource not found", "NOT_FOUND", nil, requestID)
		return
	}

	if errors.Is(err, sql.ErrConnDone) || errors.Is(err, sql.ErrTxDone) {
		m.logger.Error("Database connection error",
			ports.NewField("error", err),
			ports.NewField("request_id", requestID),
		)
		m.writeErrorResponse(w, http.StatusInternalServerError, "Database connection error", "DATABASE_CONNECTION_ERROR", nil, requestID)
		return
	}

	errStr := err.Error()
	if containsAny(errStr, []string{"database", "sql", "connection", "transaction", "postgres", "pgx"}) {
		m.logger.Error("Database-related error",
			ports.NewField("error", err),
			ports.NewField("request_id", requestID),
		)
		m.writeErrorResponse(w, http.StatusInternalServerError, "Database error occurred", "DATABASE_ERROR", nil, requestID)
		return
	}

	m.logger.Error("Internal server error",
		ports.NewField("error", err),
		ports.NewField("error_type", "UNKNOWN"),
		ports.NewField("request_id", requestID),
	)
	m.writeErrorResponse(w, http.StatusInternalServerError, "Internal server error", "INTERNAL_ERROR", nil, requestID)
}

func (m *ErrorMapper) writeErrorResponse(
	w http.ResponseWriter,
	status int,
	message string,
	code string,
	details map[string]string,
	requestID string,
) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	response := HTTPErrorResponse{
		Error: message,
		Code:  code,
		Details: details,
	}

	if requestID != "" {
		response.RequestID = requestID
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		m.logger.Error("Failed to encode error response",
			ports.NewField("error", err),
		)
		http.Error(w, message, status)
	}
}

func (m *ErrorMapper) writeError(w http.ResponseWriter, status int, message string) {
	m.writeErrorResponse(w, status, message, "", nil, "")
}

func containsAny(s string, substrings []string) bool {
	for _, substr := range substrings {
		if len(s) >= len(substr) {
			for i := 0; i <= len(s)-len(substr); i++ {
				if s[i:i+len(substr)] == substr {
					return true
				}
			}
		}
	}
	return false
}

