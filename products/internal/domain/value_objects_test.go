package domain

import (
	"errors"
	"strings"
	"testing"
)

func TestNewProductName(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantErr     bool
		errType     error
		description string
	}{
		{
			name:        "valid product name",
			input:       "Test Product",
			wantErr:     false,
			description: "should create ProductName with valid input",
		},
		{
			name:        "valid product name with spaces",
			input:       "  Test Product  ",
			wantErr:     false,
			description: "should trim spaces and create ProductName",
		},
		{
			name:        "empty string",
			input:       "",
			wantErr:     true,
			errType:     ErrInvalidProductName,
			description: "should return error for empty string",
		},
		{
			name:        "only spaces",
			input:       "   ",
			wantErr:     true,
			errType:     ErrInvalidProductName,
			description: "should return error for only spaces after trim",
		},
		{
			name:        "name with 255 characters",
			input:       string(make([]byte, 255)),
			wantErr:     false,
			description: "should accept name with maximum length",
		},
		{
			name:        "name with 256 characters",
			input:       string(make([]byte, 256)),
			wantErr:     true,
			description: "should reject name exceeding maximum length",
		},
		{
			name:        "unicode characters",
			input:       "Продукт 产品 محصول",
			wantErr:     false,
			description: "should accept unicode characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewProductName(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Errorf("NewProductName() expected error, got nil")
					return
				}
				if tt.errType != nil && err != tt.errType {
					if !errors.Is(err, tt.errType) {
						t.Errorf("NewProductName() expected error type %v, got %v", tt.errType, err)
					}
				}
			} else {
				if err != nil {
					t.Errorf("NewProductName() unexpected error: %v", err)
					return
				}
				if got.Value() != strings.TrimSpace(tt.input) {
					t.Errorf("NewProductName() value = %v, want %v", got.Value(), strings.TrimSpace(tt.input))
				}
			}
		})
	}
}

func TestProductName_Value(t *testing.T) {
	name, err := NewProductName("Test Product")
	if err != nil {
		t.Fatalf("NewProductName() unexpected error: %v", err)
	}

	if name.Value() != "Test Product" {
		t.Errorf("ProductName.Value() = %v, want %v", name.Value(), "Test Product")
	}
}

func TestProductName_String(t *testing.T) {
	name, err := NewProductName("Test Product")
	if err != nil {
		t.Fatalf("NewProductName() unexpected error: %v", err)
	}

	if name.String() != "Test Product" {
		t.Errorf("ProductName.String() = %v, want %v", name.String(), "Test Product")
	}

	if name.String() != name.Value() {
		t.Errorf("ProductName.String() = %v, should equal Value() = %v", name.String(), name.Value())
	}
}

func TestNewPrice(t *testing.T) {
	tests := []struct {
		name        string
		input       float64
		wantErr     bool
		errType     error
		description string
	}{
		{
			name:        "valid positive price",
			input:       99.99,
			wantErr:     false,
			description: "should create Price with valid positive value",
		},
		{
			name:        "valid integer price",
			input:       100,
			wantErr:     false,
			description: "should create Price with integer value",
		},
		{
			name:        "very small positive price",
			input:       0.01,
			wantErr:     false,
			description: "should accept very small positive price",
		},
		{
			name:        "zero price",
			input:       0,
			wantErr:     true,
			errType:     ErrInvalidProductPrice,
			description: "should return error for zero price",
		},
		{
			name:        "negative price",
			input:       -10.50,
			wantErr:     true,
			errType:     ErrInvalidProductPrice,
			description: "should return error for negative price",
		},
		{
			name:        "maximum allowed price",
			input:       1e15,
			wantErr:     false,
			description: "should accept maximum allowed price",
		},
		{
			name:        "price exceeding maximum",
			input:       1e15 + 1,
			wantErr:     true,
			description: "should reject price exceeding maximum",
		},
		{
			name:        "very large price",
			input:       1e16,
			wantErr:     true,
			description: "should reject very large price",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewPrice(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Errorf("NewPrice() expected error, got nil")
					return
				}
				if tt.errType != nil {
					if !errors.Is(err, tt.errType) {
						t.Errorf("NewPrice() expected error type %v, got %v", tt.errType, err)
					}
				}
			} else {
				if err != nil {
					t.Errorf("NewPrice() unexpected error: %v", err)
					return
				}
				if got.Value() != tt.input {
					t.Errorf("NewPrice() value = %v, want %v", got.Value(), tt.input)
				}
			}
		})
	}
}

func TestPrice_Value(t *testing.T) {
	price, err := NewPrice(99.99)
	if err != nil {
		t.Fatalf("NewPrice() unexpected error: %v", err)
	}

	if price.Value() != 99.99 {
		t.Errorf("Price.Value() = %v, want %v", price.Value(), 99.99)
	}
}

func TestPrice_Float64(t *testing.T) {
	price, err := NewPrice(99.99)
	if err != nil {
		t.Fatalf("NewPrice() unexpected error: %v", err)
	}

	if price.Float64() != 99.99 {
		t.Errorf("Price.Float64() = %v, want %v", price.Float64(), 99.99)
	}

	if price.Float64() != price.Value() {
		t.Errorf("Price.Float64() = %v, should equal Value() = %v", price.Float64(), price.Value())
	}
}

func TestProductName_Immutable(t *testing.T) {
	name1, err := NewProductName("Product 1")
	if err != nil {
		t.Fatalf("NewProductName() unexpected error: %v", err)
	}

	name2, err := NewProductName("Product 2")
	if err != nil {
		t.Fatalf("NewProductName() unexpected error: %v", err)
	}

	if name1.Value() == name2.Value() {
		t.Error("ProductName should be immutable - different instances should have different values")
	}
}

func TestPrice_Immutable(t *testing.T) {
	price1, err := NewPrice(99.99)
	if err != nil {
		t.Fatalf("NewPrice() unexpected error: %v", err)
	}

	price2, err := NewPrice(199.99)
	if err != nil {
		t.Fatalf("NewPrice() unexpected error: %v", err)
	}

	if price1.Value() == price2.Value() {
		t.Error("Price should be immutable - different instances should have different values")
	}
}
