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
	"network-scanner/internal/repository"
	"network-scanner/internal/scanner"
	"network-scanner/pkg/logger"

	httpSwagger "github.com/swaggo/http-swagger" // SWAGGER
	_ "network-scanner/docs"                     // SWAGGER
)

// @title Network Scanner API
// @version 1.0
// @description API для сканирования портов и просмотра истории запросов
// @host localhost:8080
// @BasePath /api/v1
func main() {
	log := logger.New()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	repo, err := repository.NewPostgresRepository(cfg.DB)
	if err != nil {
		log.Fatalf("Failed to initialize repository: %v", err)
	}
	defer repo.Close()

	portScanner := scanner.NewPortScanner(log, cfg.Scanner.Timeout, cfg.Scanner.MaxRetries, cfg.Scanner.RetryDelay)

	handlers := handler.NewHandler(log, repo, portScanner)

	r := handlers.InitRoutes()
	r.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler) // SWAGGER

	srv := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
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
