package main

import (
	"arp_scanner/internal/config"
	"arp_scanner/internal/service"
	"arp_scanner/pkg/logger"
	"context"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	log := logger.New()
	cfg := config.Load()

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	if err := service.Run(ctx, cfg, *log); err != nil {
		log.Errorf("Service failed: %v", err)
		os.Exit(1)
	}

	log.Info("Service stopped gracefully")
}
