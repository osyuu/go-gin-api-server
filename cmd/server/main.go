package main

import (
	"context"
	"fmt"
	"go-gin-api-server/config"
	"go-gin-api-server/internal/database"
	"go-gin-api-server/internal/server"
	"go-gin-api-server/pkg/logger"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"
)

func main() {
	cfg := config.LoadConfig()

	if err := logger.Init(cfg.Env); err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer func() {
		if err := logger.Log.Sync(); err != nil {
			log.Printf("Failed to sync logger: %v", err)
		}
	}()

	// Initialize database
	if err := database.InitDatabase(cfg.Database); err != nil {
		logger.Log.Fatal("Failed to initialize database", zap.Error(err))
	}
	defer func() {
		if err := database.CloseDatabase(); err != nil {
			logger.Log.Error("Failed to close database", zap.Error(err))
		}
	}()

	// Auto migrate database tables
	if err := database.AutoMigrate(); err != nil {
		logger.Log.Fatal("Failed to migrate database", zap.Error(err))
	}

	router := server.NewServer(cfg)

	srv := &http.Server{
		Addr:              fmt.Sprintf(":%s", cfg.Port),
		Handler:           router,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	go func() {
		logger.Log.Info("Server is running on",
			zap.String("port", srv.Addr),
			zap.String("env", cfg.Env),
		)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Log.Fatal("Server error", zap.Error(err))
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Log.Info("Shutting down server...")

	// wait for 5 seconds before shutting down
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// shutdown server
	if err := srv.Shutdown(ctx); err != nil {
		logger.Log.Fatal("Forced to shutdown", zap.Error(err))
	}

	logger.Log.Info("Server exited gracefully")
}
