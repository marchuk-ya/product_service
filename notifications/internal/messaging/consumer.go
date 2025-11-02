package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"product_service/notifications/internal/domain"

	"github.com/streadway/amqp"
	"go.uber.org/zap"
)

type Consumer interface {
	Start(ctx context.Context) error
	Stop() error
}

type rabbitMQConsumer struct {
	conn      *amqp.Connection
	channel   *amqp.Channel
	exchange  string
	queueName string
	logger    *zap.Logger
	done      chan bool
}

func NewRabbitMQConsumer(connStr, exchange string, logger *zap.Logger) (Consumer, error) {
	conn, err := amqp.Dial(connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

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
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to declare exchange: %w", err)
	}

	queue, err := ch.QueueDeclare(
		"",
		false,
		false,
		true,
		false,
		nil,
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to declare queue: %w", err)
	}

	err = ch.QueueBind(
		queue.Name,
		"",
		exchange,
		false,
		nil,
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to bind queue: %w", err)
	}

	return &rabbitMQConsumer{
		conn:      conn,
		channel:   ch,
		exchange:  exchange,
		queueName: queue.Name,
		logger:    logger,
		done:      make(chan bool),
	}, nil
}

func (c *rabbitMQConsumer) Start(ctx context.Context) error {
	msgs, err := c.channel.Consume(
		c.queueName,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to register consumer: %w", err)
	}

	c.logger.Info("Started consuming messages",
		zap.String("exchange", c.exchange),
		zap.String("queue", c.queueName))

	go func() {
		for {
			select {
			case <-ctx.Done():
				c.logger.Info("Context cancelled, stopping consumer")
				return
			case <-c.done:
				c.logger.Info("Consumer stopped")
				return
			case msg, ok := <-msgs:
				if !ok {
					c.logger.Info("Message channel closed")
					return
				}
				c.handleMessage(msg)
			}
		}
	}()

	return nil
}

func (c *rabbitMQConsumer) handleMessage(msg amqp.Delivery) {
	var event domain.ProductEvent
	if err := json.Unmarshal(msg.Body, &event); err != nil {
		c.logger.Error("Failed to unmarshal message",
			zap.Error(err),
			zap.String("body", string(msg.Body)))
		msg.Nack(false, false)
		return
	}

	if event.Type != domain.EventTypeProductCreated && event.Type != domain.EventTypeProductDeleted {
		c.logger.Warn("Unknown event type",
			zap.String("type", event.Type),
			zap.String("body", string(msg.Body)))
		msg.Ack(false)
		return
	}

	c.logger.Info("Received product event",
		zap.String("type", event.Type),
		zap.Int("product_id", event.ProductID),
		zap.Time("timestamp", event.Timestamp),
		zap.String("raw_json", string(msg.Body)))

	if err := msg.Ack(false); err != nil {
		c.logger.Error("Failed to acknowledge message", zap.Error(err))
	} else {
		c.logger.Debug("Message acknowledged successfully",
			zap.String("type", event.Type),
			zap.Int("product_id", event.ProductID))
	}
}

func (c *rabbitMQConsumer) Stop() error {
	close(c.done)
	if c.channel != nil {
		c.channel.Close()
	}
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

