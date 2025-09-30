package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config represents the AUSF configuration
type Config struct {
	NF            NFConfig            `yaml:"nf"`
	SBI           SBIConfig           `yaml:"sbi"`
	NRF           NRFConfig           `yaml:"nrf"`
	UDM           UDMConfig           `yaml:"udm"`
	PLMN          PLMNConfig          `yaml:"plmn"`
	Auth          AuthConfig          `yaml:"auth"`
	Observability ObservabilityConfig `yaml:"observability"`
}

// NFConfig contains NF instance configuration
type NFConfig struct {
	Name        string `yaml:"name"`
	InstanceID  string `yaml:"instance_id"`
	Description string `yaml:"description"`
}

// SBIConfig contains Service-Based Interface configuration
type SBIConfig struct {
	Scheme      string    `yaml:"scheme"`
	BindAddress string    `yaml:"bind_address"`
	Port        int       `yaml:"port"`
	TLS         TLSConfig `yaml:"tls"`
}

// TLSConfig contains TLS configuration
type TLSConfig struct {
	Enabled  bool   `yaml:"enabled"`
	CertFile string `yaml:"cert_file"`
	KeyFile  string `yaml:"key_file"`
}

// NRFConfig contains NRF client configuration
type NRFConfig struct {
	URL               string        `yaml:"url"`
	Enabled           bool          `yaml:"enabled"`
	HeartbeatInterval time.Duration `yaml:"heartbeat_interval"`
}

// UDMConfig contains UDM client configuration
type UDMConfig struct {
	URL     string        `yaml:"url"`
	Timeout time.Duration `yaml:"timeout"`
}

// PLMNConfig contains PLMN configuration
type PLMNConfig struct {
	MCC string `yaml:"mcc"` // Mobile Country Code
	MNC string `yaml:"mnc"` // Mobile Network Code
}

// AuthConfig contains authentication configuration
type AuthConfig struct {
	Methods       []string `yaml:"methods"`        // Supported auth methods
	DefaultMethod string   `yaml:"default_method"` // Default method
}

// ObservabilityConfig contains observability settings
type ObservabilityConfig struct {
	Metrics MetricsConfig `yaml:"metrics"`
	Tracing TracingConfig `yaml:"tracing"`
	Logging LoggingConfig `yaml:"logging"`
}

// MetricsConfig contains metrics configuration
type MetricsConfig struct {
	Enabled bool `yaml:"enabled"`
	Port    int  `yaml:"port"`
}

// TracingConfig contains tracing configuration
type TracingConfig struct {
	Enabled  bool   `yaml:"enabled"`
	Exporter string `yaml:"exporter"`
	Endpoint string `yaml:"endpoint"`
}

// LoggingConfig contains logging configuration
type LoggingConfig struct {
	Level  string `yaml:"level"`
	Format string `yaml:"format"`
}

// Load loads configuration from a YAML file
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &config, nil
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.NF.Name == "" {
		return fmt.Errorf("nf.name is required")
	}

	if c.NF.InstanceID == "" {
		return fmt.Errorf("nf.instance_id is required")
	}

	if c.SBI.Port <= 0 || c.SBI.Port > 65535 {
		return fmt.Errorf("invalid sbi.port: %d", c.SBI.Port)
	}

	if c.NRF.Enabled && c.NRF.URL == "" {
		return fmt.Errorf("nrf.url is required when nrf.enabled is true")
	}

	if c.UDM.URL == "" {
		return fmt.Errorf("udm.url is required")
	}

	if c.PLMN.MCC == "" || c.PLMN.MNC == "" {
		return fmt.Errorf("plmn.mcc and plmn.mnc are required")
	}

	if len(c.Auth.Methods) == 0 {
		return fmt.Errorf("at least one authentication method must be configured")
	}

	return nil
}

// GetSBIURL returns the full SBI URL
func (c *Config) GetSBIURL() string {
	return fmt.Sprintf("%s://%s:%d", c.SBI.Scheme, c.SBI.BindAddress, c.SBI.Port)
}
