package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/your-org/5g-network/nf/amf/internal/client"
	"github.com/your-org/5g-network/nf/amf/internal/config"
	amfcontext "github.com/your-org/5g-network/nf/amf/internal/context"
	"github.com/your-org/5g-network/nf/amf/internal/server"
	"github.com/your-org/5g-network/nf/amf/internal/service"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	Version   = "dev"
	BuildTime = "unknown"
)

func main() {
	// Parse command line flags
	configPath := flag.String("config", "nf/amf/config/amf.yaml", "path to configuration file")
	flag.Parse()

	// Create logger
	logger := createLogger("info")
	defer logger.Sync()

	logger.Info("Starting AMF (Access and Mobility Management Function)",
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
		zap.String("ausf_url", cfg.AUSF.URL),
		zap.String("udm_url", cfg.UDM.URL),
		zap.String("guami", cfg.GetGUAMI()),
	)

	// Create AUSF client
	ausfClient := client.NewAUSFClient(cfg.AUSF.URL, cfg.AUSF.Timeout, logger)
	logger.Info("AUSF client initialized")

	// Create UE context manager
	contextManager := amfcontext.NewUEContextManager()
	logger.Info("UE context manager initialized")

	// Create registration service
	registrationService := service.NewRegistrationService(cfg, ausfClient, contextManager, logger)
	logger.Info("Registration service initialized")

	// Create HTTP server
	srv := server.NewServer(cfg, registrationService, contextManager, logger)

	// Register with NRF if enabled (simplified - no actual NRF registration for now)
	if cfg.NRF.Enabled {
		logger.Info("NRF registration configured",
			zap.String("nrf_url", cfg.NRF.URL),
		)
		// TODO: Implement NRF registration
	}

	// Start HTTP server in a goroutine
	serverErrors := make(chan error, 1)
	go func() {
		logger.Info("AMF started successfully",
			zap.String("address", fmt.Sprintf("%s:%d", cfg.SBI.BindAddress, cfg.SBI.Port)),
			zap.String("scheme", cfg.SBI.Scheme),
			zap.String("guami", cfg.GetGUAMI()),
		)
		serverErrors <- srv.Start()
	}()

	// Wait for interrupt signal or server error
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		logger.Fatal("Server error", zap.Error(err))
	case sig := <-shutdown:
		logger.Info("Shutdown signal received", zap.String("signal", sig.String()))

		// Create shutdown context with timeout
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer shutdownCancel()

		// Gracefully shutdown the server
		if err := srv.Stop(shutdownCtx); err != nil {
			logger.Error("Failed to gracefully shutdown server", zap.Error(err))
		}

		logger.Info("AMF shutdown complete")
	}
}

// createLogger creates a structured logger
func createLogger(level string) *zap.Logger {
	var zapLevel zapcore.Level
	switch level {
	case "debug":
		zapLevel = zapcore.DebugLevel
	case "info":
		zapLevel = zapcore.InfoLevel
	case "warn":
		zapLevel = zapcore.WarnLevel
	case "error":
		zapLevel = zapcore.ErrorLevel
	default:
		zapLevel = zapcore.InfoLevel
	}

	config := zap.NewProductionConfig()
	config.Level = zap.NewAtomicLevelAt(zapLevel)
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	logger, err := config.Build()
	if err != nil {
		panic(fmt.Sprintf("failed to create logger: %v", err))
	}

	return logger
}