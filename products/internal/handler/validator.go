package handler

import (
	"fmt"
	"product_service/products/internal/handler/dto"
)

func ValidateCreateProductRequest(req dto.CreateProductRequest) error {
	if req.Name == "" {
		return fmt.Errorf("name is required")
	}
	if req.Price <= 0 {
		return fmt.Errorf("price must be greater than zero")
	}
	return nil
}

