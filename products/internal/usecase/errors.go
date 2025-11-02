package usecase

import (
	"errors"
	"fmt"
)

var (
	ErrInvalidInput     = errors.New("invalid input")
	ErrNotFound         = errors.New("resource not found")
	ErrAlreadyExists    = errors.New("resource already exists")
	ErrOperationFailed  = errors.New("operation failed")
)

const (
	ErrCodeValidation      = "VALIDATION_ERROR"
	ErrCodeNotFound        = "NOT_FOUND"
	ErrCodeAlreadyExists   = "ALREADY_EXISTS"
	ErrCodeOperationFailed = "OPERATION_FAILED"
	ErrCodeDatabaseError   = "DATABASE_ERROR"
	ErrCodeTimeout         = "TIMEOUT"
)

type UseCaseError struct {
	Err     error
	Message string
	Code    string
	Context map[string]interface{}
}

func (e *UseCaseError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	if e.Err != nil {
		return e.Err.Error()
	}
	return "unknown error"
}

func (e *UseCaseError) Unwrap() error {
	return e.Err
}

func (e *UseCaseError) WithContext(key string, value interface{}) *UseCaseError {
	if e.Context == nil {
		e.Context = make(map[string]interface{})
	}
	e.Context[key] = value
	return e
}

func NewUseCaseError(err error, message string, code string) *UseCaseError {
	return &UseCaseError{
		Err:     err,
		Message: message,
		Code:    code,
		Context: make(map[string]interface{}),
	}
}

func WrapError(err error, message string) error {
	if err == nil {
		return nil
	}
	if useCaseErr, ok := err.(*UseCaseError); ok {
		return &UseCaseError{
			Err:     useCaseErr.Err,
			Message: message,
			Code:    useCaseErr.Code,
			Context: useCaseErr.Context,
		}
	}
	return fmt.Errorf("%s: %w", message, err)
}

func WrapErrorWithCode(err error, message string, code string) *UseCaseError {
	return &UseCaseError{
		Err:     err,
		Message: message,
		Code:    code,
		Context: make(map[string]interface{}),
	}
}

