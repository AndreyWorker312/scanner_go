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
)

func main() {
	// Инициализация логгера
	log := logger.New()

	// Загрузка конфигурации
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Инициализация репозитория
	repo, err := repository.NewPostgresRepository(cfg.DB)
	if err != nil {
		log.Fatalf("Failed to initialize repository: %v", err)
	}
	defer repo.Close()

	// Инициализация сканера
	portScanner := scanner.NewPortScanner(log, cfg.Scanner.Timeout)

	// Инициализация обработчиков
	handlers := handler.NewHandler(log, repo, portScanner)

	// Создание HTTP сервера
	srv := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      handlers.InitRoutes(),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	// Запуск сервера в горутине
	go func() {
		log.Infof("Starting server on port %s", cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Ожидание сигналов для graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down server...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Errorf("Server forced to shutdown: %v", err)
	}

	log.Info("Server exited properly")
}
