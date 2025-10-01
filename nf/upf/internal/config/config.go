package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config holds the UPF configuration
type Config struct {
	NF            NFConfig            `yaml:"nf"`
	PFCP          PFCPConfig          `yaml:"pfcp"`
	N3            N3Config            `yaml:"n3"`
	N6            N6Config            `yaml:"n6"`
	N9            N9Config            `yaml:"n9"`
	PLMN          PLMNConfig          `yaml:"plmn"`
	DNN           []DNNConfig         `yaml:"dnn"`
	QoS           QoSConfig           `yaml:"qos"`
	Forwarding    ForwardingConfig    `yaml:"forwarding"`
	NRF           NRFConfig           `yaml:"nrf"`
	Observability ObservabilityConfig `yaml:"observability"`
}

// NFConfig holds NF-specific configuration
type NFConfig struct {
	Name        string `yaml:"name"`
	InstanceID  string `yaml:"instance_id"`
	Description string `yaml:"description"`
}

// PFCPConfig holds PFCP (N4) interface configuration
type PFCPConfig struct {
	BindAddress string `yaml:"bind_address"`
	Port        int    `yaml:"port"`
	NodeID      string `yaml:"node_id"`
}

// N3Config holds N3 interface configuration (gNB-UPF)
type N3Config struct {
	BindAddress  string `yaml:"bind_address"`
	Port         int    `yaml:"port"`
	LocalAddress string `yaml:"local_address"`
}

// N6Config holds N6 interface configuration (Data Network)
type N6Config struct {
	InterfaceName string `yaml:"interface_name"`
	Subnet        string `yaml:"subnet"`
	Gateway       string `yaml:"gateway"`
	DNSPrimary    string `yaml:"dns_primary"`
	DNSSecondary  string `yaml:"dns_secondary"`
}

// N9Config holds N9 interface configuration (UPF-UPF)
type N9Config struct {
	Enabled     bool   `yaml:"enabled"`
	BindAddress string `yaml:"bind_address"`
	Port        int    `yaml:"port"`
}

// PLMNConfig holds PLMN configuration
type PLMNConfig struct {
	MCC string `yaml:"mcc"`
	MNC string `yaml:"mnc"`
}

// DNNConfig holds Data Network Name configuration
type DNNConfig struct {
	Name    string `yaml:"name"`
	CIDR    string `yaml:"cidr"`
	Gateway string `yaml:"gateway"`
}

// QoSConfig holds QoS configuration
type QoSConfig struct {
	MaxUplinkBitrate   uint64 `yaml:"max_uplink_bitrate"`
	MaxDownlinkBitrate uint64 `yaml:"max_downlink_bitrate"`
	DefaultQFI         uint8  `yaml:"default_qfi"`
}

// ForwardingConfig holds forwarding configuration
type ForwardingConfig struct {
	MaxSessions        int           `yaml:"max_sessions"`
	SessionIdleTimeout time.Duration `yaml:"session_idle_timeout"`
	BufferSize         int           `yaml:"buffer_size"`
}

// NRFConfig holds NRF client configuration
type NRFConfig struct {
	URL               string        `yaml:"url"`
	Enabled           bool          `yaml:"enabled"`
	HeartbeatInterval time.Duration `yaml:"heartbeat_interval"`
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

// Load reads the configuration from a file
func Load(configPath string) (*Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Set defaults
	if config.PFCP.Port == 0 {
		config.PFCP.Port = 8805
	}
	if config.N3.Port == 0 {
		config.N3.Port = 2152
	}
	if config.N9.Port == 0 {
		config.N9.Port = 2153
	}
	if config.Forwarding.BufferSize == 0 {
		config.Forwarding.BufferSize = 65535
	}

	return &config, nil
}

// GetPFCPAddress returns the PFCP bind address
func (c *Config) GetPFCPAddress() string {
	return fmt.Sprintf("%s:%d", c.PFCP.BindAddress, c.PFCP.Port)
}

// GetN3Address returns the N3 bind address
func (c *Config) GetN3Address() string {
	return fmt.Sprintf("%s:%d", c.N3.BindAddress, c.N3.Port)
}

// GetN9Address returns the N9 bind address
func (c *Config) GetN9Address() string {
	return fmt.Sprintf("%s:%d", c.N9.BindAddress, c.N9.Port)
}
