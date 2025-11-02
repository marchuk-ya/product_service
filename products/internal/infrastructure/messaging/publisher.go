package messaging

import (
	"context"
	"fmt"
	"product_service/products/internal/infrastructure/events"
	"product_service/products/internal/infrastructure/retry"
	"product_service/products/internal/usecase/ports"
	"time"

	"github.com/streadway/amqp"
	"go.uber.org/zap"
)

var _ ports.EventPublisher = (*rabbitMQPublisher)(nil)

var _ ports.EventPublisherHealthChecker = (*rabbitMQPublisher)(nil)

type rabbitMQPublisher struct {
	conn     *amqp.Connection
	channel  *amqp.Channel
	exchange string
	logger   *zap.Logger
}

func (p *rabbitMQPublisher) IsHealthy(ctx context.Context) bool {
	if p.conn == nil || p.channel == nil {
		return false
	}
	
	if p.conn.IsClosed() {
		return false
	}
	
	_, err := p.channel.QueueDeclare(
		"",
		false,
		true,
		false,
		false,
		nil,
	)
	
	return err == nil
}

func NewRabbitMQPublisher(ctx context.Context, connStr, exchange string, logger *zap.Logger) (ports.EventPublisher, error) {
	retryCfg := retry.DefaultConfig()
	retryCfg.MaxAttempts = 5
	retryCfg.BaseBackoff = 2 * time.Second
	retryCfg.MaxBackoff = 10 * time.Second
	retryCfg.InitialDelay = 1 * time.Second

	var conn *amqp.Connection
	var ch *amqp.Channel
	var err error

	err = retry.Do(ctx, retryCfg, func() error {
		conn, err = amqp.Dial(connStr)
		if err != nil {
			return fmt.Errorf("failed to dial RabbitMQ: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ after retries: %w", err)
	}

	err = retry.Do(ctx, retryCfg, func() error {
		ch, err = conn.Channel()
		if err != nil {
			if conn != nil {
				conn.Close()
			}
			return fmt.Errorf("failed to create channel: %w", err)
		}
		return nil
	})
	if err != nil {
		if conn != nil {
			conn.Close()
		}
		return nil, fmt.Errorf("failed to create channel after retries: %w", err)
	}

	err = retry.Do(ctx, retryCfg, func() error {
		err = ch.ExchangeDeclare(
			exchange,
			"fanout",
			true,
			false,
			false,
			false,
			nil,
		)
		if err != nil {
			return fmt.Errorf("failed to declare exchange: %w", err)
		}
		return nil
	})
	if err != nil {
		if ch != nil {
			ch.Close()
		}
		if conn != nil {
			conn.Close()
		}
		return nil, fmt.Errorf("failed to declare exchange after retries: %w", err)
	}

	return &rabbitMQPublisher{
		conn:     conn,
		channel:  ch,
		exchange: exchange,
		logger:   logger,
	}, nil
}

func (p *rabbitMQPublisher) PublishProductCreated(ctx context.Context, productID int) error {
	timestamp, ok := GetTimestamp(ctx)
	if !ok {
		timestamp = time.Now()
	}
	
	event := events.InfrastructureEvent{
		Type:      events.EventTypeProductCreated,
		ProductID: productID,
		Timestamp: timestamp,
	}
	return p.publishInfrastructureEvent(ctx, event)
}

func (p *rabbitMQPublisher) PublishProductDeleted(ctx context.Context, productID int) error {
	timestamp, ok := GetTimestamp(ctx)
	if !ok {
		timestamp = time.Now()
	}
	
	event := events.InfrastructureEvent{
		Type:      events.EventTypeProductDeleted,
		ProductID: productID,
		Timestamp: timestamp,
	}
	return p.publishInfrastructureEvent(ctx, event)
}

func (p *rabbitMQPublisher) publishInfrastructureEvent(ctx context.Context, event events.InfrastructureEvent) error {
	body, err := event.ToJSON()
	if err != nil {
		return err
	}

	err = p.channel.Publish(
		p.exchange,
		"",
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
			Timestamp:   event.Timestamp,
		},
	)
	if err != nil {
		p.logger.Error("Failed to publish event", zap.Error(err), zap.String("type", event.Type))
		return err
	}

	p.logger.Info("Event published successfully", 
		zap.String("type", event.Type), 
		zap.Int("product_id", event.ProductID))
	return nil
}

func (p *rabbitMQPublisher) Close() error {
	if p.channel != nil {
		p.channel.Close()
	}
	if p.conn != nil {
		return p.conn.Close()
	}
	return nil
}

