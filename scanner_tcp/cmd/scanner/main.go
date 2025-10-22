package main

import (
	"context"
	"os"
	"os/signal"
	"scanner_tcp/internal/config"
	"scanner_tcp/internal/service"
	"scanner_tcp/pkg/logger"
	"syscall"
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

