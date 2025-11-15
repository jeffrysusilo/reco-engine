package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/yourusername/reco-engine/internal/processor"
	"github.com/yourusername/reco-engine/internal/store"
	"github.com/yourusername/reco-engine/internal/util/config"
	"github.com/yourusername/reco-engine/internal/util/logger"
	"go.uber.org/zap"
)

func main() {
	// Load configuration
	cfg, err := config.Load("")
	if err != nil {
		panic(fmt.Sprintf("Failed to load config: %v", err))
	}

	// Initialize logger
	if err := logger.Init(cfg.Observability.Logging.Level, cfg.Observability.Logging.Format); err != nil {
		panic(fmt.Sprintf("Failed to initialize logger: %v", err))
	}
	defer logger.Sync()

	logger.Info("Starting Stream Processor Service")

	// Initialize Redis
	redisStore, err := store.NewRedisStore(cfg.Redis)
	if err != nil {
		logger.Fatal("Failed to connect to Redis", zap.Error(err))
	}
	defer redisStore.Close()

	// Initialize service
	svc := processor.NewService(cfg, redisStore)
	defer svc.Close()

	// Create context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start processing in goroutine
	errCh := make(chan error, 1)
	go func() {
		if err := svc.Start(ctx); err != nil {
			errCh <- err
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-quit:
		logger.Info("Received shutdown signal")
		cancel()
	case err := <-errCh:
		logger.Error("Processor error", zap.Error(err))
		cancel()
	}

	logger.Info("Processor exited")
}
