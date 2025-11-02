package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"

	"product_service/products/internal/bootstrap"
)

func main() {
	app, err := bootstrap.Bootstrap()
	if err != nil {
		panic(fmt.Sprintf("Failed to bootstrap application: %v", err))
	}
	defer app.Logger.Sync()

	go func() {
		if err := app.Start(); err != nil && err != http.ErrServerClosed {
			app.Logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	shutdownCtx, cancel := context.WithTimeout(context.Background(), app.Config.Server.ShutdownTimeout)
	defer cancel()


	if err := app.Shutdown(shutdownCtx); err != nil {
		app.Logger.Error("Failed to shutdown gracefully", zap.Error(err))
	}
}
