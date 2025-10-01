package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/your-org/5g-network/common/metrics"
	"github.com/your-org/5g-network/nf/smf/internal/client"
	"github.com/your-org/5g-network/nf/smf/internal/config"
	smfcontext "github.com/your-org/5g-network/nf/smf/internal/context"
	"github.com/your-org/5g-network/nf/smf/internal/n4"
	"github.com/your-org/5g-network/nf/smf/internal/server"
	"github.com/your-org/5g-network/nf/smf/internal/service"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func main() {
	// Parse command line flags
	configPath := flag.String("config", "nf/smf/config/smf.yaml", "Path to configuration file")
	flag.Parse()

	// Initialize logger
	logger := initLogger("info")
	defer func() {
		if err := logger.Sync(); err != nil {
			// Ignore sync errors on stdout/stderr
		}
	}()

	logger.Info("Starting SMF (Session Management Function)",
		zap.String("config", *configPath),
	)

	// Load configuration
	cfg, err := config.Load(*configPath)
	if err != nil {
		logger.Fatal("Failed to load configuration", zap.Error(err))
	}

	logger.Info("Configuration loaded",
		zap.String("smf_name", cfg.SMF.Name),
		zap.String("sbi_address", fmt.Sprintf("%s:%d", cfg.SBI.IPv4, cfg.SBI.Port)),
		zap.String("nrf_url", cfg.NRF.URL),
	)

	// Initialize metrics server
	metricsServer := metrics.NewMetricsServer(9095, logger)
	go func() {
		logger.Info("Starting metrics server on :9095")
		if err := metricsServer.Start(); err != nil {
			logger.Error("Metrics server error", zap.Error(err))
		}
	}()
	defer metricsServer.Stop()

	// Set service up
	metrics.SetServiceUp(true)
	defer metrics.SetServiceUp(false)

	// Initialize NRF client
	nrfClient := client.NewNRFClient(cfg, logger)

	// Register with NRF
	if err := nrfClient.Register(); err != nil {
		logger.Error("Failed to register with NRF (continuing anyway)", zap.Error(err))
	}

	// Start NRF heartbeat
	go startNRFHeartbeat(nrfClient, cfg.NRF.HeartbeatInterval, logger)

	// Initialize PFCP client for UPF communication
	pfcpClient := n4.NewPFCPClient(
		cfg.UPF.DefaultUPF.NodeID,
		cfg.UPF.DefaultUPF.N4Address,
		logger,
	)

	// Establish PFCP association with UPF
	if err := pfcpClient.AssociatePFCPSession(); err != nil {
		logger.Error("Failed to establish PFCP association with UPF (continuing anyway)", zap.Error(err))
	}

	// Initialize SMF context
	smfContext := smfcontext.NewSMFContext(
		cfg.UPF.DefaultUPF.NodeID,
		cfg.UPF.DefaultUPF.N4Address,
	)

	// Initialize session service
	sessionService, err := service.NewSessionService(cfg, smfContext, pfcpClient, logger)
	if err != nil {
		logger.Fatal("Failed to create session service", zap.Error(err))
	}

	// Initialize HTTP server
	smfServer := server.NewSMFServer(cfg, sessionService, logger)

	// Start HTTP server in goroutine
	serverErrors := make(chan error, 1)
	go func() {
		logger.Info("SMF HTTP server listening",
			zap.String("address", fmt.Sprintf("%s:%d", cfg.SBI.IPv4, cfg.SBI.Port)),
		)
		serverErrors <- smfServer.Start()
	}()

	logger.Info("SMF started successfully",
		zap.String("name", cfg.SMF.Name),
		zap.String("plmn", fmt.Sprintf("MCC=%s, MNC=%s", cfg.SMF.PLMN.MCC, cfg.SMF.PLMN.MNC)),
	)

	// Wait for interrupt signal or server error
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		logger.Fatal("Server error", zap.Error(err))
	case sig := <-shutdown:
		logger.Info("Shutdown signal received", zap.String("signal", sig.String()))

		// Deregister from NRF
		if err := nrfClient.Deregister(); err != nil {
			logger.Error("Failed to deregister from NRF", zap.Error(err))
		}

		// Graceful shutdown with timeout
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := smfServer.Stop(ctx); err != nil {
			logger.Error("Error during server shutdown", zap.Error(err))
		}

		logger.Info("SMF shutdown complete")
	}
}

// initLogger initializes the logger
func initLogger(level string) *zap.Logger {
	var zapLevel zapcore.Level
	if err := zapLevel.UnmarshalText([]byte(level)); err != nil {
		zapLevel = zapcore.InfoLevel
	}

	config := zap.Config{
		Level:            zap.NewAtomicLevelAt(zapLevel),
		Development:      false,
		Encoding:         "console",
		EncoderConfig:    zap.NewDevelopmentEncoderConfig(),
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
	}

	logger, err := config.Build()
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize logger: %v", err))
	}

	return logger
}

// startNRFHeartbeat starts periodic NRF heartbeat
func startNRFHeartbeat(nrfClient *client.NRFClient, interval time.Duration, logger *zap.Logger) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		if err := nrfClient.SendHeartbeat(); err != nil {
			logger.Error("NRF heartbeat failed", zap.Error(err))
		}
	}
}
