package domain

import "time"

type ProductEvent struct {
	Type      string    `json:"type"`
	ProductID int       `json:"product_id"`
	Timestamp time.Time `json:"timestamp"`
}

const (
	EventTypeProductCreated = "PRODUCT_CREATED"
	EventTypeProductDeleted = "PRODUCT_DELETED"
)

