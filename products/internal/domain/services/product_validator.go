package services

import (
	"errors"
	"product_service/products/internal/domain"
)

type ProductDomainService interface {
	ValidateProductForCreation(name string, price float64) error
	
	ValidateProductForUpdate(name string, price float64) error
	
	CanDeleteProduct(product *domain.Product) error
}

type ProductValidator interface {
	ValidateProductName(name string) error
}

type productDomainService struct {
	validator ProductValidator
}

func NewProductDomainService(validator ProductValidator) ProductDomainService {
	return &productDomainService{
		validator: validator,
	}
}

var _ ProductDomainService = (*productDomainService)(nil)

func (s *productDomainService) ValidateProductForCreation(name string, price float64) error {
	
	if s.validator != nil {
		if err := s.validator.ValidateProductName(name); err != nil {
			return domain.WrapDomainError(err, "VALIDATION_FAILED", "product name validation failed")
		}
	}
	
	return nil
}

func (s *productDomainService) ValidateProductForUpdate(name string, price float64) error {
	return s.ValidateProductForCreation(name, price)
}

func (s *productDomainService) CanDeleteProduct(product *domain.Product) error {
	if product == nil {
		return errors.New("product is nil")
	}
	
	
	return nil
}

