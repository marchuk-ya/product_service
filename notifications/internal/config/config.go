package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

type Config struct {
	RabbitMQURL string
	Exchange    string
	Port        string
	Logger      *zap.Logger
}

func LoadConfig() (*Config, error) {
	_ = godotenv.Load()

	logger, err := zap.NewProduction()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize logger: %w", err)
	}

	rmqHost := getEnv("RABBITMQ_HOST", "localhost")
	rmqPort := getEnv("RABBITMQ_PORT", "5672")
	rmqUser := getEnv("RABBITMQ_USER", "guest")
	rmqPassword := getEnv("RABBITMQ_PASSWORD", "guest")

	rabbitMQURL := fmt.Sprintf("amqp://%s:%s@%s:%s", rmqUser, rmqPassword, rmqHost, rmqPort)

	exchange := getEnv("RABBITMQ_EXCHANGE", "products_events")
	port := getEnv("NOTIFICATIONS_SERVICE_PORT", "8081")

	return &Config{
		RabbitMQURL: rabbitMQURL,
		Exchange:    exchange,
		Port:        port,
		Logger:      logger,
	}, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

