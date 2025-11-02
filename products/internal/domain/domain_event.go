package domain

import (
	"encoding/json"
	"time"
)

type DomainEvent interface {
	EventType() string
	OccurredAt() time.Time
	MarshalJSON() ([]byte, error)
}

type ProductCreatedEvent struct {
	ProductID int
	Product   *Product
	Timestamp time.Time
}

func NewProductCreatedEvent(productID int, product *Product) DomainEvent {
	return ProductCreatedEvent{
		ProductID: productID,
		Product:   product,
		Timestamp: time.Now(),
	}
}

func (e ProductCreatedEvent) EventType() string {
	return "PRODUCT_CREATED"
}

func (e ProductCreatedEvent) OccurredAt() time.Time {
	return e.Timestamp
}

func (e ProductCreatedEvent) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"type":       e.EventType(),
		"product_id": e.ProductID,
		"timestamp":  e.Timestamp,
	})
}

type ProductDeletedEvent struct {
	ProductID int
	Timestamp time.Time
}

func NewProductDeletedEvent(productID int) DomainEvent {
	return ProductDeletedEvent{
		ProductID: productID,
		Timestamp: time.Now(),
	}
}

func (e ProductDeletedEvent) EventType() string {
	return "PRODUCT_DELETED"
}

func (e ProductDeletedEvent) OccurredAt() time.Time {
	return e.Timestamp
}

func (e ProductDeletedEvent) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"type":       e.EventType(),
		"product_id": e.ProductID,
		"timestamp":  e.Timestamp,
	})
}

