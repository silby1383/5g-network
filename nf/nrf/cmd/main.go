package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/your-org/5g-network/nf/nrf/internal/config"
	"github.com/your-org/5g-network/nf/nrf/internal/server"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	Version   = "dev"
	BuildTime = "unknown"
)

func main() {
	// Parse command-line flags
	configPath := flag.String("config", "config/nrf.yaml", "Path to configuration file")
	logLevel := flag.String("log-level", "info", "Log level (debug, info, warn, error)")
	flag.Parse()

	// Initialize logger
	logger := initLogger(*logLevel)
	defer logger.Sync()

	logger.Info("Starting NRF (Network Repository Function)",
		zap.String("version", Version),
		zap.String("build_time", BuildTime),
	)

	// Load configuration
	cfg, err := config.Load(*configPath)
	if err != nil {
		logger.Fatal("Failed to load configuration", zap.Error(err))
	}

	logger.Info("Configuration loaded",
		zap.String("sbi_bind", cfg.SBI.BindAddress),
		zap.Int("sbi_port", cfg.SBI.Port),
	)

	// Create context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create and start NRF server
	nrfServer, err := server.NewNRFServer(cfg, logger)
	if err != nil {
		logger.Fatal("Failed to create NRF server", zap.Error(err))
	}

	// Start server in goroutine
	errChan := make(chan error, 1)
	go func() {
		if err := nrfServer.Start(ctx); err != nil {
			errChan <- err
		}
	}()

	logger.Info("NRF started successfully",
		zap.String("address", fmt.Sprintf("%s:%d", cfg.SBI.BindAddress, cfg.SBI.Port)),
		zap.String("scheme", cfg.SBI.Scheme),
	)

	// Wait for shutdown signal or error
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	select {
	case err := <-errChan:
		logger.Error("Server error", zap.Error(err))
	case sig := <-sigChan:
		logger.Info("Received shutdown signal", zap.String("signal", sig.String()))
	}

	// Graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	logger.Info("Shutting down NRF...")
	if err := nrfServer.Stop(shutdownCtx); err != nil {
		logger.Error("Error during shutdown", zap.Error(err))
	}

	logger.Info("NRF stopped")
}

// initLogger initializes the logger
func initLogger(level string) *zap.Logger {
	// Parse log level
	var zapLevel zapcore.Level
	if err := zapLevel.UnmarshalText([]byte(level)); err != nil {
		zapLevel = zapcore.InfoLevel
	}

	// Create logger config
	config := zap.Config{
		Level:            zap.NewAtomicLevelAt(zapLevel),
		Development:      false,
		Encoding:         "json",
		EncoderConfig:    zap.NewProductionEncoderConfig(),
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
	}

	// Build logger
	logger, err := config.Build(
		zap.AddCaller(),
		zap.AddStacktrace(zapcore.ErrorLevel),
	)
	if err != nil {
		panic(fmt.Sprintf("failed to initialize logger: %v", err))
	}

	return logger
}
