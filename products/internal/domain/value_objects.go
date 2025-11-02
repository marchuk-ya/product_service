package domain

import (
	"errors"
	"strings"
)

type ProductName struct {
	value string
}

func NewProductName(name string) (ProductName, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return ProductName{}, ErrInvalidProductName
	}
	if len(name) > 255 {
		return ProductName{}, errors.New("product name cannot exceed 255 characters")
	}
	return ProductName{value: name}, nil
}

func (n ProductName) Value() string {
	return n.value
}

func (n ProductName) String() string {
	return n.value
}

type Price struct {
	value float64
}

func NewPrice(price float64) (Price, error) {
	if price <= 0 {
		return Price{}, ErrInvalidProductPrice
	}
	if price > 1e15 {
		return Price{}, errors.New("product price cannot exceed 1e15")
	}
	return Price{value: price}, nil
}

func (p Price) Value() float64 {
	return p.value
}

func (p Price) Float64() float64 {
	return p.value
}

