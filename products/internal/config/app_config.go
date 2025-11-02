package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type AppConfig struct {
	Database    DatabaseConfig
	RabbitMQ    RabbitMQConfig
	Server      ServerConfig
	Tracing     TracingConfig
	Outbox      OutboxConfig
}

type DatabaseConfig struct {
	Host               string
	Port               string
	User               string
	Password           string
	Name               string
	MaxOpenConns       int
	MaxIdleConns       int
	ConnMaxLifetime    time.Duration
	ConnMaxIdleTime    time.Duration
}

func (d DatabaseConfig) ConnectionString() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		d.Host, d.Port, d.User, d.Password, d.Name)
}

type RabbitMQConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Exchange string
}

func (r RabbitMQConfig) URL() string {
	return fmt.Sprintf("amqp://%s:%s@%s:%s", r.User, r.Password, r.Host, r.Port)
}

type ServerConfig struct {
	Port                    string
	RequestTimeout         time.Duration
	ReadTimeout            time.Duration
	ShutdownTimeout        time.Duration
}

type TracingConfig struct {
	Enabled       bool
	OTLPEndpoint  string
	ServiceName   string
}

type OutboxConfig struct {
	MaxBatchSize  int
	Interval      time.Duration
	BatchSize     int
	MaxRetries    int
	BaseBackoff   time.Duration
	MaxBackoff    time.Duration
	Concurrency   int
}

func LoadAppConfig() (*AppConfig, error) {
	_ = godotenv.Load()

	return &AppConfig{
		Database: DatabaseConfig{
			Host:            getEnv("POSTGRES_HOST", "localhost"),
			Port:            getEnv("POSTGRES_PORT", "5432"),
			User:            getEnv("POSTGRES_USER", "postgres"),
			Password:        getEnv("POSTGRES_PASSWORD", "root"),
			Name:            getEnv("POSTGRES_DB", "legaltech_test"),
			MaxOpenConns:    getEnvAsInt("POSTGRES_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    getEnvAsInt("POSTGRES_MAX_IDLE_CONNS", 5),
			ConnMaxLifetime: getEnvAsDuration("POSTGRES_CONN_MAX_LIFETIME", 5*time.Minute),
			ConnMaxIdleTime: getEnvAsDuration("POSTGRES_CONN_MAX_IDLE_TIME", 5*time.Minute),
		},
		RabbitMQ: RabbitMQConfig{
			Host:     getEnv("RABBITMQ_HOST", "localhost"),
			Port:     getEnv("RABBITMQ_PORT", "5672"),
			User:     getEnv("RABBITMQ_USER", "guest"),
			Password: getEnv("RABBITMQ_PASSWORD", "guest"),
			Exchange: getEnv("RABBITMQ_EXCHANGE", "products_events"),
		},
		Server: ServerConfig{
			Port:            getEnv("PRODUCTS_SERVICE_PORT", "8080"),
			RequestTimeout:  getEnvAsDuration("REQUEST_TIMEOUT", 10*time.Second),
			ReadTimeout:     getEnvAsDuration("READ_TIMEOUT", 5*time.Second),
			ShutdownTimeout: getEnvAsDuration("SHUTDOWN_TIMEOUT", 15*time.Second),
		},
		Tracing: TracingConfig{
			Enabled:      getEnvAsBool("TRACING_ENABLED", false),
			OTLPEndpoint: getEnv("OTLP_ENDPOINT", "localhost:4318"),
			ServiceName:  getEnv("SERVICE_NAME", "product_service"),
		},
		Outbox: OutboxConfig{
			MaxBatchSize: getEnvAsInt("OUTBOX_MAX_BATCH_SIZE", 100),
			Interval:     getEnvAsDuration("OUTBOX_INTERVAL", 5*time.Second),
			BatchSize:    getEnvAsInt("OUTBOX_BATCH_SIZE", 50),
			MaxRetries:   getEnvAsInt("OUTBOX_MAX_RETRIES", 3),
			BaseBackoff:  getEnvAsDuration("OUTBOX_BASE_BACKOFF", 1*time.Second),
			MaxBackoff:   getEnvAsDuration("OUTBOX_MAX_BACKOFF", 30*time.Second),
			Concurrency:  getEnvAsInt("OUTBOX_CONCURRENCY", 3),
		},
	}, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

