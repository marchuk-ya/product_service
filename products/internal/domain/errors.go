package domain

import (
	"errors"
	"fmt"
)

var (
	ErrProductNotFound = errors.New("product not found")
	ErrInvalidInput    = errors.New("invalid input")
)

type DomainError struct {
	Code string
	Message string
	Err error
	Context map[string]interface{}
}

func (e *DomainError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	if e.Err != nil {
		return e.Err.Error()
	}
	return fmt.Sprintf("domain error: %s", e.Code)
}

func (e *DomainError) Unwrap() error {
	return e.Err
}

func (e *DomainError) WithContext(key string, value interface{}) *DomainError {
	if e.Context == nil {
		e.Context = make(map[string]interface{})
	}
	e.Context[key] = value
	return e
}

func NewDomainError(code, message string) *DomainError {
	return &DomainError{
		Code:    code,
		Message: message,
		Context: make(map[string]interface{}),
	}
}

func WrapDomainError(err error, code, message string) *DomainError {
	return &DomainError{
		Code:    code,
		Message: message,
		Err:     err,
		Context: make(map[string]interface{}),
	}
}

