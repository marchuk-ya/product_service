package events

import (
	"encoding/json"
	"product_service/products/internal/domain"
	"time"
)

type InfrastructureEvent struct {
	Type      string    `json:"type"`
	ProductID int       `json:"product_id"`
	Timestamp time.Time `json:"timestamp"`
}

func ToInfrastructureEvent(event domain.DomainEvent) InfrastructureEvent {
	switch e := event.(type) {
	case domain.ProductCreatedEvent:
		return InfrastructureEvent{
			Type:      e.EventType(),
			ProductID: e.ProductID,
			Timestamp: e.OccurredAt(),
		}
	case domain.ProductDeletedEvent:
		return InfrastructureEvent{
			Type:      e.EventType(),
			ProductID: e.ProductID,
			Timestamp: e.OccurredAt(),
		}
	default:
		return InfrastructureEvent{
			Type:      event.EventType(),
			Timestamp: event.OccurredAt(),
		}
	}
}

func (e InfrastructureEvent) ToJSON() ([]byte, error) {
	return json.Marshal(e)
}

const (
	EventTypeProductCreated = "PRODUCT_CREATED"
	EventTypeProductDeleted = "PRODUCT_DELETED"
)

