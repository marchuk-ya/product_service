package dto

import (
	"product_service/products/internal/domain"
	"time"
)

func ToProductResponse(p *domain.Product) ProductResponse {
	return ProductResponse{
		ID:        p.ID,
		Name:      p.Name.Value(),
		Price:     p.Price.Value(),
		CreatedAt: p.CreatedAt.Format(time.RFC3339),
	}
}

func ToProductResponseList(products []domain.Product) []ProductResponse {
	responses := make([]ProductResponse, len(products))
	for i, p := range products {
		responses[i] = ToProductResponse(&p)
	}
	return responses
}

