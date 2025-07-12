package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"network-scanner/internal/config"
	"network-scanner/internal/handler"
	"network-scanner/internal/scanner"
	"network-scanner/pkg/logger"
)

func main() {
	log := logger.New()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	portScanner := scanner.NewPortScanner(log, cfg.Scanner.Timeout, cfg.Scanner.MaxRetries, cfg.Scanner.RetryDelay)
	handlers := handler.NewHandler(log, portScanner)

	srv := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      handlers.InitRoutes(),
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	go func() {
		log.Infof("Starting server on port %s", cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Errorf("Server forced to shutdown: %v", err)
	}

	log.Info("Server exited properly")
}
