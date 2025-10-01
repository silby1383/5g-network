package config

import (
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config represents the SMF configuration
type Config struct {
	SBI           SBIConfig           `yaml:"sbi"`
	NRF           NRFConfig           `yaml:"nrf"`
	UDM           UDMConfig           `yaml:"udm"`
	PCF           PCFConfig           `yaml:"pcf"`
	SMF           SMFConfig           `yaml:"smf"`
	UPF           UPFConfig           `yaml:"upf"`
	Observability ObservabilityConfig `yaml:"observability"`
}

// SBIConfig represents Service Based Interface configuration
type SBIConfig struct {
	Scheme string    `yaml:"scheme"`
	IPv4   string    `yaml:"ipv4"`
	Port   int       `yaml:"port"`
	TLS    TLSConfig `yaml:"tls"`
}

// TLSConfig represents TLS configuration
type TLSConfig struct {
	Enabled bool   `yaml:"enabled"`
	Cert    string `yaml:"cert"`
	Key     string `yaml:"key"`
}

// NRFConfig represents NRF client configuration
type NRFConfig struct {
	URL               string        `yaml:"url"`
	HeartbeatInterval time.Duration `yaml:"heartbeat_interval"`
}

// UDMConfig represents UDM client configuration
type UDMConfig struct {
	URL string `yaml:"url"`
}

// PCFConfig represents PCF client configuration
type PCFConfig struct {
	URL string `yaml:"url"`
}

// SMFConfig represents SMF-specific configuration
type SMFConfig struct {
	Name     string `yaml:"name"`
	SetID    string `yaml:"set_id"`
	RegionID string `yaml:"region_id"`
	PLMN     PLMN   `yaml:"plmn"`

	SupportedSNSSAI []SNSSAI `yaml:"supported_snssai"`
	SupportedDNN    []DNN    `yaml:"supported_dnn"`

	UESubnet           UESubnet `yaml:"ue_subnet"`
	DefaultSessionAMBR AMBR     `yaml:"default_session_ambr"`
}

// PLMN represents Public Land Mobile Network
type PLMN struct {
	MCC string `yaml:"mcc"`
	MNC string `yaml:"mnc"`
}

// SNSSAI represents Single Network Slice Selection Assistance Information
type SNSSAI struct {
	SST int    `yaml:"sst"`
	SD  string `yaml:"sd"`
}

// DNN represents Data Network Name
type DNN struct {
	DNN string    `yaml:"dnn"`
	DNS DNSConfig `yaml:"dns"`
}

// DNSConfig represents DNS configuration
type DNSConfig struct {
	IPv4 string `yaml:"ipv4"`
	IPv6 string `yaml:"ipv6"`
}

// UESubnet represents UE IP address pool
type UESubnet struct {
	IPv4 string `yaml:"ipv4"`
	IPv6 string `yaml:"ipv6"`
}

// AMBR represents Aggregate Maximum Bit Rate
type AMBR struct {
	Uplink   string `yaml:"uplink"`
	Downlink string `yaml:"downlink"`
}

// UPFConfig represents UPF configuration
type UPFConfig struct {
	DefaultUPF DefaultUPF `yaml:"default_upf"`
}

// DefaultUPF represents static UPF configuration
type DefaultUPF struct {
	NodeID    string `yaml:"node_id"`
	N4Address string `yaml:"n4_address"`
}

// ObservabilityConfig represents observability configuration
type ObservabilityConfig struct {
	LogLevel string     `yaml:"log_level"`
	OTEL     OTELConfig `yaml:"otel"`
	EBPF     EBPFConfig `yaml:"ebpf"`
}

// OTELConfig represents OpenTelemetry configuration
type OTELConfig struct {
	Enabled     bool   `yaml:"enabled"`
	Endpoint    string `yaml:"endpoint"`
	ServiceName string `yaml:"service_name"`
}

// EBPFConfig represents eBPF configuration
type EBPFConfig struct {
	Enabled bool `yaml:"enabled"`
}

// Load loads configuration from a YAML file
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// GetSBIURI returns the full SBI URI
func (c *Config) GetSBIURI() string {
	return c.SBI.Scheme + "://" + c.SBI.IPv4 + ":" + string(rune(c.SBI.Port))
}
