package usecase

import (
	"errors"
	"product_service/products/internal/domain"
)

type DomainError struct {
	Err error
	Message string
}

func (e *DomainError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return e.Err.Error()
}

func (e *DomainError) Unwrap() error {
	return e.Err
}

func IsDomainError(err error) bool {
	var domainErr *DomainError
	return errors.As(err, &domainErr)
}

func WrapDomainError(domainErr error, message string) error {
	return &DomainError{
		Err:     domainErr,
		Message: message,
	}
}

func IsProductNotFound(err error) bool {
	return errors.Is(err, domain.ErrProductNotFound)
}

func IsInvalidProductName(err error) bool {
	return errors.Is(err, domain.ErrInvalidProductName)
}

func IsInvalidProductPrice(err error) bool {
	return errors.Is(err, domain.ErrInvalidProductPrice)
}

