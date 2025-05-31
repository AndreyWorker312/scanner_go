package main

import (
	"context"
	"fmt"
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

	"github.com/golang-migrate/migrate/v4"                     // Для миграций
	_ "github.com/golang-migrate/migrate/v4/database/postgres" // Драйвер для PostgreSQL
	_ "github.com/golang-migrate/migrate/v4/source/file"       // Источник файловых миграций
	"github.com/rs/cors"                                       // Для CORS
	httpSwagger "github.com/swaggo/http-swagger"               // SWAGGER
	_ "network-scanner/docs"                                   // SWAGGER
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

	// Применяем миграции перед созданием репозитория
	if err := applyMigrations(cfg.DB); err != nil && err != migrate.ErrNoChange {
		log.Fatalf("Failed to apply migrations: %v", err)
	} else if err == nil {
		log.Info("Migrations applied successfully")
	} else {
		log.Info("Database is up to date")
	}

	repo, err := repository.NewPostgresRepository(cfg.DB)
	if err != nil {
		log.Fatalf("Failed to initialize repository: %v", err)
	}
	defer repo.Close()

	portScanner := scanner.NewPortScanner(log, cfg.Scanner.Timeout, cfg.Scanner.MaxRetries, cfg.Scanner.RetryDelay)

	handlers := handler.NewHandler(log, repo, portScanner)

	r := handlers.InitRoutes()

	// Настройка CORS middleware
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"}, // Разрешить все origins (для разработки)
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE"},
		AllowedHeaders:   []string{"Content-Type"},
		AllowCredentials: true,
	})

	// Обертываем роутер в CORS handler
	handler := c.Handler(r)

	// Добавляем Swagger после CORS
	r.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)
	handler = c.Handler(r)

	srv := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      handler, // Используем handler с CORS
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

func applyMigrations(cfg config.DBConfig) error {
	d, err := migrate.New(
		"file://migrations",
		fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
			cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Name))
	if err != nil {
		return err
	}
	defer d.Close()
	return d.Up()
}
