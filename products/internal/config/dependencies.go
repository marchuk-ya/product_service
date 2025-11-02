package config

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"
	"product_service/products/internal/infrastructure/retry"
)

type Dependencies struct {
	DB     *sql.DB
	Logger *zap.Logger
}

func NewDependencies(logger *zap.Logger) *Dependencies {
	return &Dependencies{
		Logger: logger,
	}
}

func (d *Dependencies) InitDatabase(ctx context.Context, connStr string, dbConfig DatabaseConfig) error {
	retryCfg := retry.DefaultConfig()
	retryCfg.MaxAttempts = 5
	retryCfg.BaseBackoff = 2 * time.Second
	retryCfg.MaxBackoff = 10 * time.Second
	retryCfg.InitialDelay = 1 * time.Second

	var db *sql.DB
	var err error

	err = retry.Do(ctx, retryCfg, func() error {
		db, err = sql.Open("pgx", connStr)
		if err != nil {
			return fmt.Errorf("failed to open database: %w", err)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to open database after retries: %w", err)
	}

	err = retry.Do(ctx, retryCfg, func() error {
		if err := db.Ping(); err != nil {
			return fmt.Errorf("failed to ping database: %w", err)
		}
		return nil
	})
	if err != nil {
		db.Close()
		return fmt.Errorf("failed to ping database after retries: %w", err)
	}

	db.SetMaxOpenConns(dbConfig.MaxOpenConns)
	db.SetMaxIdleConns(dbConfig.MaxIdleConns)
	db.SetConnMaxLifetime(dbConfig.ConnMaxLifetime)
	db.SetConnMaxIdleTime(dbConfig.ConnMaxIdleTime)

	d.DB = db
	d.Logger.Info("Database connection established",
		zap.Int("max_open_conns", dbConfig.MaxOpenConns),
		zap.Int("max_idle_conns", dbConfig.MaxIdleConns),
		zap.Duration("conn_max_lifetime", dbConfig.ConnMaxLifetime),
		zap.Duration("conn_max_idle_time", dbConfig.ConnMaxIdleTime),
	)
	return nil
}

func (d *Dependencies) Close() error {
	if d.DB != nil {
		if err := d.DB.Close(); err != nil {
			return fmt.Errorf("failed to close database: %w", err)
		}
	}
	return nil
}

func NewLogger() (*zap.Logger, error) {
	logger, err := zap.NewProduction()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize logger: %w", err)
	}
	return logger, nil
}

