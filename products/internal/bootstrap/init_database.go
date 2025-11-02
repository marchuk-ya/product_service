package bootstrap

import (
	"context"
	"fmt"
	"os"
	"time"

	"go.uber.org/zap"

	"product_service/products/internal/config"
	"product_service/products/internal/database"
)

func initDatabase(appConfig *config.AppConfig, deps *config.Dependencies, logger *zap.Logger) error {
	dbCtx, dbCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer dbCancel()

	if err := deps.InitDatabase(dbCtx, appConfig.Database.ConnectionString(), appConfig.Database); err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}

	migrationsPath := "./migrations"
	if _, err := os.Stat(migrationsPath); os.IsNotExist(err) {
		migrationsPath = "/root/migrations"
	}
	if err := database.RunMigrations(deps.DB, migrationsPath, logger); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}

