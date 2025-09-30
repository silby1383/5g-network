package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config holds the NRF configuration
type Config struct {
	SBI           SBIConfig           `yaml:"sbi"`
	NF            NFConfig            `yaml:"nf"`
	Database      DatabaseConfig      `yaml:"database"`
	Observability ObservabilityConfig `yaml:"observability"`
}

// SBIConfig holds Service Based Interface configuration
type SBIConfig struct {
	Scheme      string    `yaml:"scheme"`       // http or https
	BindAddress string    `yaml:"bind_address"` // 0.0.0.0
	Port        int       `yaml:"port"`         // 8080
	TLS         TLSConfig `yaml:"tls"`
}

// TLSConfig holds TLS configuration
type TLSConfig struct {
	Enabled  bool   `yaml:"enabled"`
	CertFile string `yaml:"cert_file"`
	KeyFile  string `yaml:"key_file"`
}

// NFConfig holds NF-specific configuration
type NFConfig struct {
	Name        string `yaml:"name"`        // nrf-1
	InstanceID  string `yaml:"instance_id"` // UUID
	Description string `yaml:"description"`

	// NF Management configuration
	Heartbeat HeartbeatConfig `yaml:"heartbeat"`
}

// HeartbeatConfig holds heartbeat configuration
type HeartbeatConfig struct {
	Enabled  bool `yaml:"enabled"`
	Interval int  `yaml:"interval"` // seconds
	Timeout  int  `yaml:"timeout"`  // seconds
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Type      string `yaml:"type"`      // memory, redis, clickhouse
	URL       string `yaml:"url"`       // connection string
	Retention int    `yaml:"retention"` // days
}

// ObservabilityConfig holds observability configuration
type ObservabilityConfig struct {
	Metrics MetricsConfig `yaml:"metrics"`
	Tracing TracingConfig `yaml:"tracing"`
	Logging LoggingConfig `yaml:"logging"`
}

// MetricsConfig holds metrics configuration
type MetricsConfig struct {
	Enabled bool `yaml:"enabled"`
	Port    int  `yaml:"port"`
}

// TracingConfig holds tracing configuration
type TracingConfig struct {
	Enabled  bool   `yaml:"enabled"`
	Exporter string `yaml:"exporter"` // otlp, jaeger
	Endpoint string `yaml:"endpoint"`
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Level  string `yaml:"level"`  // debug, info, warn, error
	Format string `yaml:"format"` // json, text
}

// Load loads configuration from YAML file
func Load(path string) (*Config, error) {
	// Read file
	data, err := os.ReadFile(path)
	if err != nil {
		// Return default configuration if file doesn't exist
		if os.IsNotExist(err) {
			return DefaultConfig(), nil
		}
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse YAML
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &cfg, nil
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.SBI.Port <= 0 || c.SBI.Port > 65535 {
		return fmt.Errorf("invalid SBI port: %d", c.SBI.Port)
	}

	if c.SBI.Scheme != "http" && c.SBI.Scheme != "https" {
		return fmt.Errorf("invalid SBI scheme: %s (must be http or https)", c.SBI.Scheme)
	}

	if c.SBI.TLS.Enabled {
		if c.SBI.TLS.CertFile == "" || c.SBI.TLS.KeyFile == "" {
			return fmt.Errorf("TLS enabled but cert/key files not specified")
		}
	}

	if c.NF.InstanceID == "" {
		return fmt.Errorf("NF instance ID is required")
	}

	return nil
}

// DefaultConfig returns a default configuration
func DefaultConfig() *Config {
	return &Config{
		SBI: SBIConfig{
			Scheme:      "http",
			BindAddress: "0.0.0.0",
			Port:        8080,
			TLS: TLSConfig{
				Enabled: false,
			},
		},
		NF: NFConfig{
			Name:        "nrf-1",
			InstanceID:  "00000000-0000-0000-0000-000000000001",
			Description: "Network Repository Function",
			Heartbeat: HeartbeatConfig{
				Enabled:  true,
				Interval: 30,
				Timeout:  60,
			},
		},
		Database: DatabaseConfig{
			Type:      "memory",
			Retention: 7,
		},
		Observability: ObservabilityConfig{
			Metrics: MetricsConfig{
				Enabled: true,
				Port:    9090,
			},
			Tracing: TracingConfig{
				Enabled:  false,
				Exporter: "otlp",
			},
			Logging: LoggingConfig{
				Level:  "info",
				Format: "json",
			},
		},
	}
}
