package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/your-org/5g-network/common/metrics"
	"github.com/your-org/5g-network/nf/udr/internal/clickhouse"
	"github.com/your-org/5g-network/nf/udr/internal/client"
	"github.com/your-org/5g-network/nf/udr/internal/config"
	"github.com/your-org/5g-network/nf/udr/internal/repository"
	"github.com/your-org/5g-network/nf/udr/internal/server"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	Version   = "dev"
	BuildTime = "unknown"
)

func main() {
	// Parse command-line flags
	configPath := flag.String("config", "config/udr.yaml", "Path to configuration file")
	logLevel := flag.String("log-level", "info", "Log level (debug, info, warn, error)")
	initSchema := flag.Bool("init-schema", false, "Initialize ClickHouse schema")
	flag.Parse()

	// Initialize logger
	logger := initLogger(*logLevel)
	defer logger.Sync()

	logger.Info("Starting UDR (Unified Data Repository)",
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
		zap.Strings("clickhouse_addresses", cfg.ClickHouse.Addresses),
	)

	// Create ClickHouse client
	chClient, err := clickhouse.NewClient(&cfg.ClickHouse, logger)
	if err != nil {
		logger.Fatal("Failed to create ClickHouse client", zap.Error(err))
	}
	defer chClient.Close()

	logger.Info("Connected to ClickHouse successfully")

	// Initialize schema if requested
	if *initSchema {
		logger.Info("Initializing ClickHouse schema...")
		if err := initializeSchema(chClient, logger); err != nil {
			logger.Fatal("Failed to initialize schema", zap.Error(err))
		}
		logger.Info("Schema initialized successfully")
		return
	}

	// Create repository
	repo := repository.NewClickHouseRepository(chClient, logger)

	// Create context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize metrics server
	metricsServer := metrics.NewMetricsServer(9091, logger)
	go func() {
		logger.Info("Starting metrics server on :9091")
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
			NFType:       "UDR",
			NFStatus:     "REGISTERED",
			PLMNID: client.PLMNID{
				MCC: cfg.PLMN.MCC,
				MNC: cfg.PLMN.MNC,
			},
			IPv4Addresses: []string{fmt.Sprintf("%s:%d", cfg.SBI.BindAddress, cfg.SBI.Port)},
			Capacity:      100,
			Priority:      1,
			UDRInfo: &client.UDRInfo{
				GroupID: "udr-group-1",
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

	// Create and start UDR server
	udrServer, err := server.NewUDRServer(cfg, repo, logger)
	if err != nil {
		logger.Fatal("Failed to create UDR server", zap.Error(err))
	}

	// Start server in goroutine
	errChan := make(chan error, 1)
	go func() {
		if err := udrServer.Start(ctx); err != nil {
			errChan <- err
		}
	}()

	logger.Info("UDR started successfully",
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

	logger.Info("Shutting down UDR...")
	if err := udrServer.Stop(shutdownCtx); err != nil {
		logger.Error("Error during shutdown", zap.Error(err))
	}

	logger.Info("UDR stopped")
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

// initializeSchema initializes the ClickHouse database schema
func initializeSchema(client *clickhouse.Client, logger *zap.Logger) error {
	// Read schema SQL file
	schemaSQL, err := os.ReadFile("nf/udr/internal/clickhouse/schema.sql")
	if err != nil {
		// Try alternative path
		schemaSQL, err = os.ReadFile("internal/clickhouse/schema.sql")
		if err != nil {
			return fmt.Errorf("failed to read schema file: %w", err)
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Split schema into individual statements
	// ClickHouse Go client doesn't support multi-statement execution
	statements := splitSQLStatements(string(schemaSQL))

	logger.Info("Executing schema statements", zap.Int("total_statements", len(statements)))

	for i, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" || strings.HasPrefix(stmt, "--") {
			continue
		}

		logger.Debug("Executing statement", zap.Int("statement_num", i+1))
		if err := client.Exec(ctx, stmt); err != nil {
			logger.Error("Failed to execute statement",
				zap.Int("statement_num", i+1),
				zap.String("statement", stmt[:min(100, len(stmt))]),
				zap.Error(err),
			)
			return fmt.Errorf("failed to execute statement %d: %w", i+1, err)
		}
	}

	return nil
}

// splitSQLStatements splits SQL script into individual statements
func splitSQLStatements(sql string) []string {
	var statements []string
	var current strings.Builder
	inStatement := false

	lines := strings.Split(sql, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Skip empty lines and standalone comments
		if trimmed == "" {
			if inStatement {
				current.WriteString("\n")
			}
			continue
		}

		if strings.HasPrefix(trimmed, "--") && !inStatement {
			continue
		}

		// Check if this line starts a statement
		if !inStatement {
			if strings.HasPrefix(strings.ToUpper(trimmed), "CREATE") ||
				strings.HasPrefix(strings.ToUpper(trimmed), "INSERT") ||
				strings.HasPrefix(strings.ToUpper(trimmed), "ALTER") ||
				strings.HasPrefix(strings.ToUpper(trimmed), "DROP") {
				inStatement = true
			}
		}

		if inStatement {
			current.WriteString(line)
			current.WriteString("\n")

			// Check if statement ends with semicolon
			if strings.HasSuffix(trimmed, ";") {
				statements = append(statements, current.String())
				current.Reset()
				inStatement = false
			}
		}
	}

	// Add any remaining statement
	if current.Len() > 0 {
		statements = append(statements, current.String())
	}

	return statements
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
