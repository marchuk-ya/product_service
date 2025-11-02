package application

import (
	"errors"
	"fmt"
)

type TransactionError struct {
	Operation string
	Err error
}

func (e *TransactionError) Error() string {
	if e.Operation != "" {
		return fmt.Sprintf("transaction %s failed: %v", e.Operation, e.Err)
	}
	return fmt.Sprintf("transaction failed: %v", e.Err)
}

func (e *TransactionError) Unwrap() error {
	return e.Err
}

func NewTransactionError(operation string, err error) *TransactionError {
	return &TransactionError{
		Operation: operation,
		Err:       err,
	}
}

type EventPublishError struct {
	ProductID int
	EventType string
	Err error
}

func (e *EventPublishError) Error() string {
	if e.EventType != "" {
		return fmt.Sprintf("failed to publish %s event for product %d: %v", e.EventType, e.ProductID, e.Err)
	}
	return fmt.Sprintf("failed to publish event for product %d: %v", e.ProductID, e.Err)
}

func (e *EventPublishError) Unwrap() error {
	return e.Err
}

func NewEventPublishError(productID int, eventType string, err error) *EventPublishError {
	return &EventPublishError{
		ProductID: productID,
		EventType: eventType,
		Err:       err,
	}
}

type RetryExhaustedError struct {
	MaxAttempts int
	LastErr error
}

func (e *RetryExhaustedError) Error() string {
	return fmt.Sprintf("retry exhausted after %d attempts: %v", e.MaxAttempts, e.LastErr)
}

func (e *RetryExhaustedError) Unwrap() error {
	return e.LastErr
}

func NewRetryExhaustedError(maxAttempts int, lastErr error) *RetryExhaustedError {
	return &RetryExhaustedError{
		MaxAttempts: maxAttempts,
		LastErr:     lastErr,
	}
}

func IsRetryExhaustedError(err error) bool {
	var retryErr *RetryExhaustedError
	return errors.As(err, &retryErr)
}

