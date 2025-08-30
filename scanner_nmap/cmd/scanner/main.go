package scanner

import (
	"context"
	"os"
	"os/signal"
	"scanner_nmap/internal/config"
	"scanner_nmap/internal/usecases"
	"scanner_nmap/pkg/logger"
	"syscall"
)

func main() {
	log := logger.New()
	cfg := config.Load()

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	if err := usecases.Run(ctx, cfg, *log); err != nil {
		log.Errorf("Service failed: %v", err)
		os.Exit(1)
	}

	log.Info("Service stopped gracefully")
}
