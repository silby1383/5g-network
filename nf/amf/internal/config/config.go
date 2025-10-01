package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config represents the AMF configuration
type Config struct {
	NF             NFConfig             `yaml:"nf"`
	SBI            SBIConfig            `yaml:"sbi"`
	NRF            NRFConfig            `yaml:"nrf"`
	AUSF           AUSFConfig           `yaml:"ausf"`
	UDM            UDMConfig            `yaml:"udm"`
	PLMN           PLMNConfig           `yaml:"plmn"`
	AMF            AMFConfig            `yaml:"amf"`
	Security       SecurityConfig       `yaml:"security"`
	NetworkSlicing NetworkSlicingConfig `yaml:"network_slicing"`
	Timers         TimersConfig         `yaml:"timers"`
	Observability  ObservabilityConfig  `yaml:"observability"`
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

// AUSFConfig contains AUSF client configuration
type AUSFConfig struct {
	URL     string        `yaml:"url"`
	Timeout time.Duration `yaml:"timeout"`
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
	TAC string `yaml:"tac"` // Tracking Area Code
}

// AMFConfig contains AMF-specific configuration
type AMFConfig struct {
	RegionID        uint8    `yaml:"region_id"`
	SetID           uint16   `yaml:"set_id"`
	Pointer         uint8    `yaml:"pointer"`
	SupportedSNSSAI []SNSSAI `yaml:"supported_snssai"`
	SupportedDNN    []string `yaml:"supported_dnn"`
}

// SNSSAI represents Single Network Slice Selection Assistance Information
type SNSSAI struct {
	SST uint8  `yaml:"sst"` // Slice/Service Type
	SD  string `yaml:"sd"`  // Slice Differentiator
}

// SecurityConfig contains security configuration
type SecurityConfig struct {
	IntegrityOrder []string `yaml:"integrity_order"`
	CipheringOrder []string `yaml:"ciphering_order"`
}

// NetworkSlicingConfig contains network slicing configuration
type NetworkSlicingConfig struct {
	Enabled bool `yaml:"enabled"`
}

// TimersConfig contains NAS timer configuration (in seconds)
type TimersConfig struct {
	T3502 int `yaml:"t3502"` // Registration retry
	T3512 int `yaml:"t3512"` // Periodic registration
	T3522 int `yaml:"t3522"` // Deregistration
	T3550 int `yaml:"t3550"` // NAS message
	T3560 int `yaml:"t3560"` // Authentication
	T3570 int `yaml:"t3570"` // Identity request
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

	if c.AUSF.URL == "" {
		return fmt.Errorf("ausf.url is required")
	}

	if c.UDM.URL == "" {
		return fmt.Errorf("udm.url is required")
	}

	if c.PLMN.MCC == "" || c.PLMN.MNC == "" || c.PLMN.TAC == "" {
		return fmt.Errorf("plmn.mcc, plmn.mnc, and plmn.tac are required")
	}

	if len(c.AMF.SupportedSNSSAI) == 0 {
		return fmt.Errorf("at least one supported S-NSSAI must be configured")
	}

	if len(c.Security.IntegrityOrder) == 0 {
		return fmt.Errorf("at least one integrity algorithm must be configured")
	}

	if len(c.Security.CipheringOrder) == 0 {
		return fmt.Errorf("at least one ciphering algorithm must be configured")
	}

	return nil
}

// GetSBIURL returns the full SBI URL
func (c *Config) GetSBIURL() string {
	return fmt.Sprintf("%s://%s:%d", c.SBI.Scheme, c.SBI.BindAddress, c.SBI.Port)
}

// GetAMFID returns the AMF ID (Region + Set + Pointer)
func (c *Config) GetAMFID() string {
	// AMF ID format: <Region><Set><Pointer>
	// Region: 8 bits, Set: 10 bits, Pointer: 6 bits (24 bits total)
	return fmt.Sprintf("%02X%03X%02X", c.AMF.RegionID, c.AMF.SetID, c.AMF.Pointer)
}

// GetGUAMI returns the Globally Unique AMF Identifier
func (c *Config) GetGUAMI() string {
	// GUAMI format: <PLMN><AMF-ID>
	return fmt.Sprintf("%s%s-%s", c.PLMN.MCC, c.PLMN.MNC, c.GetAMFID())
}
