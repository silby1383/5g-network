package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/your-org/5g-network/common/metrics"
	"github.com/your-org/5g-network/nf/upf/internal/client"
	"github.com/your-org/5g-network/nf/upf/internal/config"
	upfcontext "github.com/your-org/5g-network/nf/upf/internal/context"
	"github.com/your-org/5g-network/nf/upf/internal/gtpu"
	"github.com/your-org/5g-network/nf/upf/internal/pfcp"
	"github.com/your-org/5g-network/nf/upf/internal/server"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	Version   = "dev"
	BuildTime = "unknown"
)

func main() {
	var configPath string
	flag.StringVar(&configPath, "config", "nf/upf/config/upf.yaml", "Path to configuration file")
	flag.Parse()

	// Initialize logger
	logger := initLogger()
	defer logger.Sync()

	logger.Info("Starting UPF (User Plane Function)",
		zap.String("version", Version),
		zap.String("build_time", BuildTime))

	// Load configuration
	cfg, err := config.Load(configPath)
	if err != nil {
		logger.Fatal("Failed to load configuration", zap.Error(err))
	}

	logger.Info("Configuration loaded",
		zap.String("pfcp_bind", cfg.GetPFCPAddress()),
		zap.String("n3_bind", cfg.GetN3Address()),
		zap.String("node_id", cfg.PFCP.NodeID))

	// Create UPF context
	upfCtx := upfcontext.NewUPFContext()
	logger.Info("UPF context initialized")

	// Create PFCP server (N4)
	pfcpServer := pfcp.NewPFCPServer(cfg, upfCtx, logger)
	logger.Info("PFCP server initialized")

	// Create GTP-U handler (N3/N6)
	gtpuHandler := gtpu.NewGTPUHandler(cfg, upfCtx, logger)
	logger.Info("GTP-U handler initialized")

	// Create admin/monitoring HTTP server
	httpServer := server.NewServer(cfg, upfCtx, gtpuHandler, logger)
	logger.Info("HTTP admin server initialized")

	// Create context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize metrics server (UPF uses port 9098, admin server uses 9096)
	metricsServer := metrics.NewMetricsServer(9098, logger)
	go func() {
		logger.Info("Starting metrics server on :9098")
		if err := metricsServer.Start(); err != nil {
			logger.Error("Metrics server error", zap.Error(err))
		}
	}()
	defer metricsServer.Stop()

	// Set service up
	metrics.SetServiceUp(true)
	defer metrics.SetServiceUp(false)

	// Register with NRF if enabled
	if cfg.NRF.Enabled {
		nrfClient := client.NewNRFClient(cfg.NRF.URL, logger)

		profile := &client.NFProfile{
			NFInstanceID: cfg.NF.InstanceID,
			NFType:       "UPF",
			NFStatus:     "REGISTERED",
			PLMNID: client.PLMNID{
				MCC: cfg.PLMN.MCC,
				MNC: cfg.PLMN.MNC,
			},
			IPv4Addresses: []string{fmt.Sprintf("%s:%d", cfg.PFCP.BindAddress, cfg.PFCP.Port)},
			Capacity:      100,
			Priority:      1,
			UPFInfo: &client.UPFInfo{
				SNSSAIUPFInfoList: []client.SNSSAIUPFInfo{
					{
						SNSSAI: client.SNSSAI{SST: 1},
						DNNUPFInfoList: []client.DNNInfo{
							{DNN: "internet"},
						},
					},
				},
				InterfaceUPFInfo: []client.InterfaceInfo{
					{
						InterfaceType: "N3",
						IPv4Addresses: []string{cfg.N3.LocalAddress},
					},
					{
						InterfaceType: "N6",
						IPv4Addresses: []string{cfg.N6.Gateway},
					},
				},
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
				if err := nrfClient.Deregister(context.Background(), cfg.NF.InstanceID); err != nil {
					logger.Error("Failed to deregister from NRF", zap.Error(err))
				}
			}()
		}
	}

	// Start PFCP server
	pfcpErrChan := make(chan error, 1)
	go func() {
		if err := pfcpServer.Start(ctx); err != nil {
			pfcpErrChan <- fmt.Errorf("PFCP server error: %w", err)
		}
	}()

	// Start GTP-U handler
	gtpuErrChan := make(chan error, 1)
	go func() {
		if err := gtpuHandler.Start(ctx); err != nil {
			gtpuErrChan <- fmt.Errorf("GTP-U handler error: %w", err)
		}
	}()

	// Start HTTP admin server
	httpErrChan := make(chan error, 1)
	go func() {
		if err := httpServer.Start(); err != nil && err != http.ErrServerClosed {
			httpErrChan <- fmt.Errorf("HTTP server error: %w", err)
		}
	}()

	logger.Info("UPF started successfully",
		zap.String("pfcp_address", cfg.GetPFCPAddress()),
		zap.String("n3_address", cfg.GetN3Address()),
		zap.String("admin_port", ":9096"))

	// Wait for shutdown signal or error
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	select {
	case sig := <-sigChan:
		logger.Info("Received shutdown signal", zap.String("signal", sig.String()))
	case err := <-pfcpErrChan:
		logger.Fatal("PFCP server failed", zap.Error(err))
	case err := <-gtpuErrChan:
		logger.Fatal("GTP-U handler failed", zap.Error(err))
	case err := <-httpErrChan:
		logger.Fatal("HTTP server failed", zap.Error(err))
	}

	// Graceful shutdown
	logger.Info("Shutting down UPF...")
	cancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := httpServer.Stop(shutdownCtx); err != nil {
		logger.Error("Error stopping HTTP server", zap.Error(err))
	}

	logger.Info("UPF shutdown complete")
}

func initLogger() *zap.Logger {
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "ts",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	config := zap.Config{
		Level:            zap.NewAtomicLevelAt(zap.InfoLevel),
		Development:      false,
		Encoding:         "json",
		EncoderConfig:    encoderConfig,
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
	}

	logger, _ := config.Build()
	return logger
}
