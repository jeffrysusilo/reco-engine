package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/yourusername/reco-engine/internal/api"
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

	logger.Info("Starting Recommendation API Service")

	// Initialize Redis
	redisStore, err := store.NewRedisStore(cfg.Redis)
	if err != nil {
		logger.Fatal("Failed to connect to Redis", zap.Error(err))
	}
	defer redisStore.Close()

	// Initialize PostgreSQL
	pgStore, err := store.NewPostgresStore(cfg.Postgres)
	if err != nil {
		logger.Fatal("Failed to connect to PostgreSQL", zap.Error(err))
	}
	defer pgStore.Close()

	// Initialize service
	svc := api.NewService(cfg, redisStore, pgStore)

	// Initialize handler
	handler := api.NewHandler(svc)

	// Setup Gin router
	if cfg.Observability.Logging.Format == "json" {
		gin.SetMode(gin.ReleaseMode)
	}
	router := gin.Default()

	// CORS middleware
	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// Routes
	router.GET("/health", handler.HandleHealth)
	router.GET("/recommendations", handler.HandleGetRecommendations)
	router.GET("/popular", handler.HandleGetPopular)

	// Metrics endpoint
	if cfg.Observability.Metrics.Enabled {
		router.GET("/metrics", gin.WrapH(promhttp.Handler()))
	}

	// HTTP server
	addr := fmt.Sprintf("%s:%d", cfg.Server.API.Host, cfg.Server.API.Port)
	srv := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	// Start server in goroutine
	go func() {
		logger.Info("HTTP server listening", zap.String("addr", addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("Server exited")
}
