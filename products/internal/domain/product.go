package domain

import (
	"errors"
	"time"
)

type Product struct {
	ID        int
	Name      ProductName
	Price     Price
	CreatedAt time.Time

	domainEvents []DomainEvent
}

var (
	ErrInvalidProductName  = errors.New("product name cannot be empty")
	ErrInvalidProductPrice = errors.New("product price must be greater than zero")
)

func (p *Product) Validate() error {
	if p.Name.Value() == "" {
		return ErrInvalidProductName
	}
	if p.Price.Value() <= 0 {
		return ErrInvalidProductPrice
	}
	return nil
}

func NewProduct(name string, price float64) (*Product, error) {
	productName, err := NewProductName(name)
	if err != nil {
		return nil, err
	}

	productPrice, err := NewPrice(price)
	if err != nil {
		return nil, err
	}

	product := &Product{
		Name:         productName,
		Price:        productPrice,
		CreatedAt:    time.Now(),
		domainEvents: make([]DomainEvent, 0),
	}

	return product, nil
}

func (p *Product) RecordCreatedEvent() {
	event := NewProductCreatedEvent(p.ID, p)
	p.recordDomainEvent(event)
}

func (p *Product) recordDomainEvent(event DomainEvent) {
	if p.domainEvents == nil {
		p.domainEvents = make([]DomainEvent, 0)
	}
	p.domainEvents = append(p.domainEvents, event)
}

func (p *Product) DomainEvents() []DomainEvent {
	if len(p.domainEvents) == 0 {
		return []DomainEvent{}
	}
	return append([]DomainEvent{}, p.domainEvents...)
}

func (p *Product) ClearDomainEvents() {
	p.domainEvents = nil
}

func (p *Product) RecordDeleteEvent() {
	event := NewProductDeletedEvent(p.ID)
	p.recordDomainEvent(event)
}
