package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"

	"product_service/notifications/internal/config"
	"product_service/notifications/internal/messaging"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		panic(fmt.Sprintf("Failed to load config: %v", err))
	}

	consumer, err := messaging.NewRabbitMQConsumer(cfg.RabbitMQURL, cfg.Exchange, cfg.Logger)
	if err != nil {
		cfg.Logger.Fatal("Failed to initialize RabbitMQ consumer", zap.Error(err))
	}
	defer consumer.Stop()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := consumer.Start(ctx); err != nil {
		cfg.Logger.Fatal("Failed to start consumer", zap.Error(err))
	}

	router := gin.New()
	router.Use(gin.Logger())

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}

	go func() {
		cfg.Logger.Info("Starting Notifications service", zap.String("port", cfg.Port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			cfg.Logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	cfg.Logger.Info("Shutting down server...")

	cancel()

	time.Sleep(2 * time.Second)

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		cfg.Logger.Error("Server forced to shutdown", zap.Error(err))
	}

	cfg.Logger.Info("Server exited")
}

