package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"test_tcp/internal/config"
	"test_tcp/internal/service"
	"test_tcp/pkg/logger"
)

func main() {
	log := logger.New()
	cfg := config.Load()

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	if err := service.Run(ctx, cfg, log); err != nil {
		log.Errorf("Service failed: %v", err)
		os.Exit(1)
	}
	log.Info("Service stopped gracefully")
}
