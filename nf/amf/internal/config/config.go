package config

// Config holds the AMF configuration
type Config struct {
	SBI            SBIConfig
	NGAP           NGAPConfig
	GUAMI          GUAMI
	Observability  ObservabilityConfig
}

// SBIConfig holds Service Based Interface configuration
type SBIConfig struct {
	BindAddress string
}

// NGAPConfig holds NGAP interface configuration
type NGAPConfig struct {
	BindAddress string
}

// GUAMI holds Globally Unique AMF Identifier
type GUAMI struct {
	PLMNIdentity PLMNIdentity
	AMFRegionID  string
	AMFSetID     string
}

// PLMNIdentity holds PLMN (Public Land Mobile Network) identity
type PLMNIdentity struct {
	MCC string // Mobile Country Code
	MNC string // Mobile Network Code
}

// ObservabilityConfig holds observability configuration
type ObservabilityConfig struct {
	EBPF EBPFConfig
}

// EBPFConfig holds eBPF tracing configuration
type EBPFConfig struct {
	Enabled bool
}

// Load loads configuration from file
func Load(path string) (*Config, error) {
	// TODO: Implement configuration loading from YAML
	return &Config{
		SBI: SBIConfig{
			BindAddress: "0.0.0.0:8080",
		},
		NGAP: NGAPConfig{
			BindAddress: "0.0.0.0:38412",
		},
		GUAMI: GUAMI{
			PLMNIdentity: PLMNIdentity{
				MCC: "001",
				MNC: "01",
			},
			AMFRegionID: "01",
			AMFSetID:    "001",
		},
		Observability: ObservabilityConfig{
			EBPF: EBPFConfig{
				Enabled: false,
			},
		},
	}, nil
}
