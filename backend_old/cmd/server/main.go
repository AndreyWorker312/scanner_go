package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"backend/internal/config"
	"backend/internal/handler"
	"backend/pkg/logger"
	"backend/pkg/queue"
)

func main() {
	log := logger.New()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	rabbitMQ, err := queue.NewRabbitMQ(queue.RabbitMQConfig{
		URL:          cfg.RabbitMQ.URL,
		ScannerQueue: cfg.RabbitMQ.ScannerQueue,
	})
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer rabbitMQ.Close()

	handlers := handler.NewHandler(log, rabbitMQ)

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
