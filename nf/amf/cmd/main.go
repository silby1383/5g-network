package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/your-org/5g-network/nf/amf/internal/config"
	"github.com/your-org/5g-network/nf/amf/internal/server"
	// "github.com/your-org/5g-network/observability/ebpf" // TODO: Fix eBPF loader
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	Version   = "dev"
	BuildTime = "unknown"
)

func main() {
	// Parse command-line flags
	configPath := flag.String("config", "config/amf.yaml", "Path to configuration file")
	logLevel := flag.String("log-level", "info", "Log level (debug, info, warn, error)")
	enableEBPF := flag.Bool("ebpf", true, "Enable eBPF tracing")
	flag.Parse()

	// Initialize logger
	logger := initLogger(*logLevel)
	defer logger.Sync()

	logger.Info("Starting AMF",
		zap.String("version", Version),
		zap.String("build_time", BuildTime),
	)

	// Load configuration
	cfg, err := config.Load(*configPath)
	if err != nil {
		logger.Fatal("Failed to load configuration", zap.Error(err))
	}

	// Create context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// TODO: Initialize eBPF tracer when loader is fixed
	_ = enableEBPF // Suppress unused variable warning
	if cfg.Observability.EBPF.Enabled {
		logger.Info("eBPF tracing requested but not yet implemented")
	}

	// Create and start AMF server
	amfServer, err := server.NewAMFServer(cfg, logger)
	if err != nil {
		logger.Fatal("Failed to create AMF server", zap.Error(err))
	}

	// Start server
	if err := amfServer.Start(ctx); err != nil {
		logger.Fatal("Failed to start AMF server", zap.Error(err))
	}

	logger.Info("AMF started successfully",
		zap.String("sbi_address", cfg.SBI.BindAddress),
		zap.String("ngap_address", cfg.NGAP.BindAddress),
		zap.String("guami", fmt.Sprintf("%s-%s-%s-%s",
			cfg.GUAMI.PLMNIdentity.MCC,
			cfg.GUAMI.PLMNIdentity.MNC,
			cfg.GUAMI.AMFRegionID,
			cfg.GUAMI.AMFSetID,
		)),
	)

	// Wait for shutdown signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	sig := <-sigChan
	logger.Info("Received shutdown signal", zap.String("signal", sig.String()))

	// Graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	logger.Info("Shutting down AMF...")
	if err := amfServer.Stop(shutdownCtx); err != nil {
		logger.Error("Error during shutdown", zap.Error(err))
	}

	logger.Info("AMF stopped")
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
