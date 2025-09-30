package config

import (
	"fmt"
	"os"
	"time"

	"github.com/your-org/5g-network/nf/udr/internal/clickhouse"
	"gopkg.in/yaml.v3"
)

// Config holds the UDR configuration
type Config struct {
	NF            NFConfig            `yaml:"nf"`
	SBI           SBIConfig           `yaml:"sbi"`
	ClickHouse    clickhouse.Config   `yaml:"clickhouse"`
	NRF           NRFConfig           `yaml:"nrf"`
	Observability ObservabilityConfig `yaml:"observability"`
}

// NFConfig holds NF-specific configuration
type NFConfig struct {
	Name        string `yaml:"name"`
	InstanceID  string `yaml:"instance_id"`
	Description string `yaml:"description"`
}

// SBIConfig holds Service Based Interface configuration
type SBIConfig struct {
	Scheme      string    `yaml:"scheme"`
	BindAddress string    `yaml:"bind_address"`
	Port        int       `yaml:"port"`
	TLS         TLSConfig `yaml:"tls"`
}

// TLSConfig holds TLS configuration
type TLSConfig struct {
	Enabled  bool   `yaml:"enabled"`
	CertFile string `yaml:"cert_file"`
	KeyFile  string `yaml:"key_file"`
}

// NRFConfig holds NRF client configuration
type NRFConfig struct {
	URL     string `yaml:"url"`
	Enabled bool   `yaml:"enabled"`
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
	Exporter string `yaml:"exporter"`
	Endpoint string `yaml:"endpoint"`
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Level  string `yaml:"level"`
	Format string `yaml:"format"`
}

// Load loads configuration from YAML file
func Load(path string) (*Config, error) {
	// Read file
	data, err := os.ReadFile(path)
	if err != nil {
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
		return fmt.Errorf("invalid SBI scheme: %s", c.SBI.Scheme)
	}

	if c.NF.InstanceID == "" {
		return fmt.Errorf("NF instance ID is required")
	}

	if len(c.ClickHouse.Addresses) == 0 {
		return fmt.Errorf("ClickHouse addresses are required")
	}

	return nil
}

// DefaultConfig returns a default configuration
func DefaultConfig() *Config {
	return &Config{
		NF: NFConfig{
			Name:        "udr-1",
			InstanceID:  "00000000-0000-0000-0000-000000000002",
			Description: "Unified Data Repository",
		},
		SBI: SBIConfig{
			Scheme:      "http",
			BindAddress: "0.0.0.0",
			Port:        8081,
			TLS: TLSConfig{
				Enabled: false,
			},
		},
		ClickHouse: clickhouse.Config{
			Addresses:    []string{"localhost:9000"},
			Database:     "udr",
			Username:     "default",
			Password:     "",
			MaxOpenConns: 10,
			MaxIdleConns: 5,
			Timeout:      10 * time.Second,
		},
		NRF: NRFConfig{
			URL:     "http://localhost:8080",
			Enabled: true,
		},
		Observability: ObservabilityConfig{
			Metrics: MetricsConfig{
				Enabled: true,
				Port:    9091,
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
