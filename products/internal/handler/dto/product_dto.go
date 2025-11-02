package dto

type CreateProductRequest struct {
	Name  string  `json:"name" binding:"required,min=1,max=255"`
	Price float64 `json:"price" binding:"required,gt=0"`
}

type ProductListResponse struct {
	Products []ProductResponse `json:"products"`
	Page     int               `json:"page"`
	Limit    int               `json:"limit"`
	Total    int               `json:"total"`
}

type ProductResponse struct {
	ID        int     `json:"id"`
	Name      string  `json:"name"`
	Price     float64 `json:"price"`
	CreatedAt string  `json:"created_at"`
}

