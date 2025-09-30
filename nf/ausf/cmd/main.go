package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/your-org/5g-network/nf/ausf/internal/client"
	"github.com/your-org/5g-network/nf/ausf/internal/config"
	"github.com/your-org/5g-network/nf/ausf/internal/server"
	"github.com/your-org/5g-network/nf/ausf/internal/service"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	Version   = "dev"
	BuildTime = "unknown"
)

func main() {
	// Parse command line flags
	configPath := flag.String("config", "nf/ausf/config/ausf.yaml", "path to configuration file")
	flag.Parse()

	// Create logger
	logger := createLogger("info")
	defer logger.Sync()

	logger.Info("Starting AUSF (Authentication Server Function)",
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
		zap.String("udm_url", cfg.UDM.URL),
		zap.String("nrf_url", cfg.NRF.URL),
	)

	// Create UDM client
	udmClient := client.NewUDMClient(cfg.UDM.URL, cfg.UDM.Timeout, logger)
	logger.Info("UDM client initialized")

	// Create authentication service
	authService := service.NewAuthenticationService(udmClient, logger)
	logger.Info("Authentication service initialized")

	// Start cleanup goroutine for expired contexts
	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()

		for range ticker.C {
			authService.CleanupExpiredContexts()
		}
	}()

	// Create HTTP server
	srv := server.NewServer(cfg, authService, logger)

	// Create context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Register with NRF if enabled
	if cfg.NRF.Enabled {
		nrfClient := client.NewNRFClient(cfg.NRF.URL, logger)

		profile := &client.NFProfile{
			NFInstanceID: cfg.NF.InstanceID,
			NFType:       "AUSF",
			NFStatus:     "REGISTERED",
			PLMNID: client.PLMNID{
				MCC: cfg.PLMN.MCC,
				MNC: cfg.PLMN.MNC,
			},
			IPv4Addresses: []string{fmt.Sprintf("%s:%d", cfg.SBI.BindAddress, cfg.SBI.Port)},
			Capacity:      100,
			Priority:      1,
			AUSFInfo: &client.AUSFInfo{
				GroupID: "ausf-group-1",
			},
		}

		if err := nrfClient.Register(ctx, profile); err != nil {
			logger.Error("Failed to register with NRF", zap.Error(err))
		} else {
			logger.Info("Registered with NRF")

			// Start heartbeat goroutine
			go func() {
				ticker := time.NewTicker(cfg.NRF.HeartbeatInterval)
				defer ticker.Stop()

				for {
					select {
					case <-ticker.C:
						if err := nrfClient.Heartbeat(ctx, cfg.NF.InstanceID); err != nil {
							logger.Error("Heartbeat failed", zap.Error(err))
						}
					case <-ctx.Done():
						return
					}
				}
			}()

			// Deregister on shutdown
			defer func() {
				deregCtx, deregCancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer deregCancel()

				if err := nrfClient.Deregister(deregCtx, cfg.NF.InstanceID); err != nil {
					logger.Error("Failed to deregister from NRF", zap.Error(err))
				} else {
					logger.Info("Deregistered from NRF")
				}
			}()
		}
	}

	// Start HTTP server in a goroutine
	serverErrors := make(chan error, 1)
	go func() {
		logger.Info("AUSF started successfully",
			zap.String("address", fmt.Sprintf("%s:%d", cfg.SBI.BindAddress, cfg.SBI.Port)),
			zap.String("scheme", cfg.SBI.Scheme),
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

		logger.Info("AUSF shutdown complete")
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
