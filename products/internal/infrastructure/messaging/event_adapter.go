package messaging

import (
	"encoding/json"
	"fmt"
	"product_service/products/internal/domain"
	"product_service/products/internal/infrastructure/events"
	"product_service/products/internal/usecase/ports"
	"time"
)

type EventAdapter interface {
	AdaptEvent(outboxEvent ports.OutboxEvent) (eventType string, productID int, timestamp time.Time, err error)
}

type domainEventAdapter struct{}

func NewDomainEventAdapter() EventAdapter {
	return &domainEventAdapter{}
}

func (a *domainEventAdapter) AdaptEvent(outboxEvent ports.OutboxEvent) (string, int, time.Time, error) {
	var eventData struct {
		Type      string    `json:"type"`
		ProductID int       `json:"product_id"`
		Timestamp time.Time `json:"timestamp"`
	}

	if err := json.Unmarshal(outboxEvent.EventData, &eventData); err != nil {
		return "", 0, time.Time{}, fmt.Errorf("failed to unmarshal domain event: %w", err)
	}

	if eventData.Type == "" {
		return "", 0, time.Time{}, fmt.Errorf("missing event type in domain event")
	}

	createdEventType := domain.ProductCreatedEvent{}.EventType()
	deletedEventType := domain.ProductDeletedEvent{}.EventType()
	if eventData.Type == createdEventType || eventData.Type == deletedEventType {
		if eventData.ProductID == 0 {
			return "", 0, time.Time{}, fmt.Errorf("missing product_id in product event")
		}
	}

	timestamp := eventData.Timestamp
	if timestamp.IsZero() {
		timestamp = time.Now()
	}

	return eventData.Type, eventData.ProductID, timestamp, nil
}

type infrastructureEventAdapter struct{}

func NewInfrastructureEventAdapter() EventAdapter {
	return &infrastructureEventAdapter{}
}

func (a *infrastructureEventAdapter) AdaptEvent(outboxEvent ports.OutboxEvent) (string, int, time.Time, error) {
	var infraEvent events.InfrastructureEvent
	if err := json.Unmarshal(outboxEvent.EventData, &infraEvent); err != nil {
		return "", 0, time.Time{}, fmt.Errorf("failed to unmarshal infrastructure event: %w", err)
	}

	timestamp := infraEvent.Timestamp
	if timestamp.IsZero() {
		timestamp = time.Now()
	}

	return infraEvent.Type, infraEvent.ProductID, timestamp, nil
}

type smartEventAdapter struct {
	domainAdapter        EventAdapter
	infrastructureAdapter EventAdapter
}

func NewSmartEventAdapter() EventAdapter {
	return &smartEventAdapter{
		domainAdapter:        NewDomainEventAdapter(),
		infrastructureAdapter: NewInfrastructureEventAdapter(),
	}
}

func (a *smartEventAdapter) AdaptEvent(outboxEvent ports.OutboxEvent) (string, int, time.Time, error) {
	var infraEvent events.InfrastructureEvent
	if err := json.Unmarshal(outboxEvent.EventData, &infraEvent); err == nil && infraEvent.Type != "" {
		timestamp := infraEvent.Timestamp
		if timestamp.IsZero() {
			timestamp = time.Now()
		}
		return infraEvent.Type, infraEvent.ProductID, timestamp, nil
	}

	return a.domainAdapter.AdaptEvent(outboxEvent)
}

