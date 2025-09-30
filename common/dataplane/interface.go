package dataplane

import (
	"context"
	"net"
	"time"
)

// DataPlane defines the interface for UPF data plane implementations.
// This allows swapping between simulated and eBPF/XDP implementations.
type DataPlane interface {
	// Initialize the data plane with configuration
	Initialize(ctx context.Context, config *Config) error

	// Install packet detection rules from PFCP session
	InstallPDR(ctx context.Context, sessionID uint64, pdr *PDR) error

	// Install forwarding action rules
	InstallFAR(ctx context.Context, sessionID uint64, far *FAR) error

	// Install QoS enforcement rules
	InstallQER(ctx context.Context, sessionID uint64, qer *QER) error

	// Install usage reporting rules
	InstallURR(ctx context.Context, sessionID uint64, urr *URR) error

	// Remove rules
	RemovePDR(ctx context.Context, sessionID uint64, pdrID uint16) error
	RemoveFAR(ctx context.Context, sessionID uint64, farID uint16) error
	RemoveQER(ctx context.Context, sessionID uint64, qerID uint16) error
	RemoveURR(ctx context.Context, sessionID uint64, urrID uint32) error

	// Remove entire session
	RemoveSession(ctx context.Context, sessionID uint64) error

	// Process a packet (for simulation)
	ProcessPacket(ctx context.Context, packet *Packet) error

	// Get statistics
	GetStats(ctx context.Context) (*Stats, error)

	// Shutdown
	Shutdown(ctx context.Context) error
}

// Config holds data plane configuration
type Config struct {
	// Interface names
	N3Interface string // Interface to RAN
	N6Interface string // Interface to Data Network
	N9Interface string // Interface to other UPFs

	// IP addresses
	N3Address net.IP
	N6Address net.IP
	N9Address net.IP

	// Data plane type
	Type string // "simulated", "ebpf", "xdp"

	// Performance settings
	Workers    int
	BufferSize int
	QueueDepth int
	MTU        int
}

// PDR (Packet Detection Rule) - 3GPP TS 29.244
type PDR struct {
	PDRID              uint16
	Precedence         uint32
	PDI                *PacketDetectionInfo
	OuterHeaderRemoval *OuterHeaderRemoval
	FARID              uint16
	URRID              []uint32
	QERID              []uint16
}

// PacketDetectionInfo defines how to match packets
type PacketDetectionInfo struct {
	SourceInterface    string // "ACCESS", "CORE", "CP-FUNCTION"
	NetworkInstance    string
	LocalFTEID         *FTEID
	UEIPAddress        *UEIPAddress
	SDFFilter          []string
	ApplicationID      string
	EthernetPDUSession *EthernetPDUSession
	QFI                uint8
}

// FAR (Forwarding Action Rule)
type FAR struct {
	FARID                uint16
	ApplyAction          uint8 // Bit flags: DROP=0x01, FORW=0x02, BUFF=0x04, NOCP=0x08, DUPL=0x10
	ForwardingParameters *ForwardingParameters
	BARID                uint16
}

// ForwardingParameters defines where to forward packets
type ForwardingParameters struct {
	DestinationInterface string // "ACCESS", "CORE", "CP-FUNCTION"
	NetworkInstance      string
	OuterHeaderCreation  *OuterHeaderCreation
	ForwardingPolicy     string
	HeaderEnrichment     *HeaderEnrichment
	TrafficEndpointID    uint8
}

// QER (QoS Enforcement Rule)
type QER struct {
	QERID              uint16
	QFI                uint8
	GateStatus         uint8 // OPEN=0, CLOSED=1
	MBR                *MBR  // Maximum Bit Rate
	GBR                *GBR  // Guaranteed Bit Rate
	PacketRate         *PacketRate
	DLFlowLevelMarking *DLFlowLevelMarking
	QoSFlowIdentifier  uint8
	ReflectiveQoS      bool
}

// URR (Usage Reporting Rule)
type URR struct {
	URRID             uint32
	MeasurementMethod uint8 // VOLUM=0x01, DURAT=0x02, EVENT=0x04
	ReportingTriggers uint32
	MeasurementPeriod time.Duration
	VolumeThreshold   *VolumeThreshold
	TimeThreshold     time.Duration
	QuotaHoldingTime  time.Duration
}

// FTEID (Fully Qualified TEID)
type FTEID struct {
	TEID     uint32
	IPv4     net.IP
	IPv6     net.IP
	ChooseID bool
}

// UEIPAddress represents UE IP address
type UEIPAddress struct {
	IPv4       net.IP
	IPv6       net.IP
	IPv6Prefix uint8
}

// OuterHeaderCreation defines GTP-U encapsulation
type OuterHeaderCreation struct {
	Description uint16 // GTP-U/UDP/IPv4, GTP-U/UDP/IPv6, etc.
	TEID        uint32
	IPv4        net.IP
	IPv6        net.IP
	PortNumber  uint16
}

// OuterHeaderRemoval defines GTP-U decapsulation
type OuterHeaderRemoval struct {
	Description uint8 // GTP-U/UDP/IPv4, GTP-U/UDP/IPv6
}

// MBR (Maximum Bit Rate)
type MBR struct {
	Uplink   uint64 // bps
	Downlink uint64 // bps
}

// GBR (Guaranteed Bit Rate)
type GBR struct {
	Uplink   uint64 // bps
	Downlink uint64 // bps
}

// PacketRate for rate limiting
type PacketRate struct {
	UplinkRate   uint16 // packets per time unit
	DownlinkRate uint16
	TimeUnit     uint8 // MINUTE=0, 6MINUTES=1, HOUR=2, DAY=3, WEEK=4
}

// VolumeThreshold for usage reporting
type VolumeThreshold struct {
	Total    uint64 // bytes
	Uplink   uint64
	Downlink uint64
}

// DLFlowLevelMarking for QoS marking
type DLFlowLevelMarking struct {
	ToSTrafficClass uint16
}

// HeaderEnrichment for header manipulation
type HeaderEnrichment struct {
	Name  string
	Value string
}

// EthernetPDUSession for Ethernet PDU sessions
type EthernetPDUSession struct {
	EtherType uint16
}

// Packet represents a network packet
type Packet struct {
	Data      []byte
	Timestamp time.Time
	SrcIP     net.IP
	DstIP     net.IP
	SrcPort   uint16
	DstPort   uint16
	Protocol  uint8
	Interface string // "N3", "N6", "N9"
	TEID      uint32 // For GTP-U packets
}

// Stats holds data plane statistics
type Stats struct {
	// Packet counters
	PacketsProcessed uint64
	PacketsDropped   uint64
	PacketsForwarded uint64
	PacketsBuffered  uint64

	// Byte counters
	BytesProcessed uint64
	BytesForwarded uint64

	// Session counters
	ActiveSessions uint32
	TotalSessions  uint64

	// GTP-U tunnel counters
	ActiveTunnels uint32

	// QoS stats
	QoSViolations uint64

	// Errors
	Errors map[string]uint64

	// Timestamp
	Timestamp time.Time
}

// Error types
const (
	ErrSessionNotFound = "session_not_found"
	ErrInvalidPDR      = "invalid_pdr"
	ErrInvalidFAR      = "invalid_far"
	ErrInvalidQER      = "invalid_qer"
	ErrInvalidPacket   = "invalid_packet"
	ErrQueueFull       = "queue_full"
)
