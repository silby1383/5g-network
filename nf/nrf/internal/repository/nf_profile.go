package repository

import (
	"time"
)

// NFType represents the type of Network Function
type NFType string

const (
	NFTypeAMF   NFType = "AMF"
	NFTypeSMF   NFType = "SMF"
	NFTypeUPF   NFType = "UPF"
	NFTypeAUSF  NFType = "AUSF"
	NFTypeUDM   NFType = "UDM"
	NFTypeUDR   NFType = "UDR"
	NFTypePCF   NFType = "PCF"
	NFTypeNRF   NFType = "NRF"
	NFTypeNSSF  NFType = "NSSF"
	NFTypeNEF   NFType = "NEF"
	NFTypeNWDAF NFType = "NWDAF"
)

// NFStatus represents the status of a Network Function
type NFStatus string

const (
	NFStatusRegistered     NFStatus = "REGISTERED"
	NFStatusSuspended      NFStatus = "SUSPENDED"
	NFStatusUndiscoverable NFStatus = "UNDISCOVERABLE"
)

// NFProfile represents a Network Function profile (TS 29.510)
type NFProfile struct {
	// Basic Information
	NFInstanceID   string   `json:"nfInstanceId"`
	NFType         NFType   `json:"nfType"`
	NFStatus       NFStatus `json:"nfStatus"`
	HeartBeatTimer int      `json:"heartBeatTimer,omitempty"` // in seconds

	// Network Information
	PLMNID        *PLMNID  `json:"plmnId,omitempty"`
	SNSSAIs       []SNSSAI `json:"sNssais,omitempty"`
	FQDN          string   `json:"fqdn,omitempty"`
	IPv4Addresses []string `json:"ipv4Addresses,omitempty"`
	IPv6Addresses []string `json:"ipv6Addresses,omitempty"`

	// Capacity
	Capacity int `json:"capacity,omitempty"` // 0-65535
	Load     int `json:"load,omitempty"`     // 0-100
	Priority int `json:"priority,omitempty"` // 0-65535

	// Service Information
	NFServices []NFService `json:"nfServices,omitempty"`

	// Location
	Locality string `json:"locality,omitempty"`

	// NF-specific profiles
	AMFInfo *AMFInfo `json:"amfInfo,omitempty"`
	SMFInfo *SMFInfo `json:"smfInfo,omitempty"`
	UPFInfo *UPFInfo `json:"upfInfo,omitempty"`

	// Metadata
	RecoveryTime  time.Time `json:"recoveryTime,omitempty"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
	LastHeartbeat time.Time `json:"lastHeartbeat"`
}

// PLMNID represents a Public Land Mobile Network ID
type PLMNID struct {
	MCC string `json:"mcc"` // Mobile Country Code (3 digits)
	MNC string `json:"mnc"` // Mobile Network Code (2 or 3 digits)
}

// SNSSAI represents Single Network Slice Selection Assistance Information
type SNSSAI struct {
	SST int    `json:"sst"`          // Slice/Service Type (0-255)
	SD  string `json:"sd,omitempty"` // Slice Differentiator (6 hex digits)
}

// NFService represents a service provided by an NF
type NFService struct {
	ServiceInstanceID string             `json:"serviceInstanceId"`
	ServiceName       string             `json:"serviceName"`
	Versions          []NFServiceVersion `json:"versions"`
	Scheme            string             `json:"scheme"` // http, https
	FQDN              string             `json:"fqdn,omitempty"`
	IPv4Addresses     []string           `json:"ipv4EndPoints,omitempty"`
	IPv6Addresses     []string           `json:"ipv6EndPoints,omitempty"`
	Port              int                `json:"port,omitempty"`
	Priority          int                `json:"priority,omitempty"`
	Capacity          int                `json:"capacity,omitempty"`
	Load              int                `json:"load,omitempty"`
	APIPrefix         string             `json:"apiPrefix,omitempty"`
	SupportedFeatures string             `json:"supportedFeatures,omitempty"`
}

// NFServiceVersion represents a service version
type NFServiceVersion struct {
	APIVersionInURI string `json:"apiVersionInUri"` // e.g., "v1"
	APIFullVersion  string `json:"apiFullVersion"`  // e.g., "1.0.0"
}

// AMFInfo contains AMF-specific information
type AMFInfo struct {
	AMFSetID        string           `json:"amfSetId"`
	AMFRegionID     string           `json:"amfRegionId"`
	GUAMIList       []GUAMI          `json:"guamiList,omitempty"`
	TaiList         []TAI            `json:"taiList,omitempty"`
	N2InterfaceInfo *N2InterfaceInfo `json:"n2InterfaceAmfInfo,omitempty"`
}

// GUAMI represents Globally Unique AMF Identifier
type GUAMI struct {
	PLMNID      PLMNID `json:"plmnId"`
	AMFRegionID string `json:"amfRegionId"`
	AMFSetID    string `json:"amfSetId"`
	AMFPointer  string `json:"amfPointer"`
}

// TAI represents Tracking Area Identity
type TAI struct {
	PLMNID PLMNID `json:"plmnId"`
	TAC    string `json:"tac"` // Tracking Area Code
}

// N2InterfaceInfo represents N2 interface information
type N2InterfaceInfo struct {
	IPv4Addresses []string `json:"ipv4EndpointAddresses,omitempty"`
	IPv6Addresses []string `json:"ipv6EndpointAddresses,omitempty"`
	AMFName       string   `json:"amfName,omitempty"`
}

// SMFInfo contains SMF-specific information
type SMFInfo struct {
	SNSSAIs     []SNSSAI     `json:"sNssaiSmfInfoList,omitempty"`
	TaiList     []TAI        `json:"taiList,omitempty"`
	PGWFQDNList []string     `json:"pgwFqdn,omitempty"`
	AccessType  []string     `json:"accessType,omitempty"`
	SMFInfoList []SNSSAIInfo `json:"sNssaiUpfInfoList,omitempty"`
}

// SNSSAIInfo represents S-NSSAI specific information
type SNSSAIInfo struct {
	SNSSAI  SNSSAI   `json:"sNssai"`
	DNNList []string `json:"dnnList,omitempty"`
}

// UPFInfo contains UPF-specific information
type UPFInfo struct {
	SNSSAIs              []SNSSAI           `json:"sNssaiUpfInfoList,omitempty"`
	SMFServingArea       []string           `json:"smfServingArea,omitempty"`
	InterfaceUpfInfoList []InterfaceUpfInfo `json:"interfaceUpfInfoList,omitempty"`
}

// InterfaceUpfInfo represents UPF interface information
type InterfaceUpfInfo struct {
	InterfaceType   string   `json:"interfaceType"` // N3, N6, N9
	IPv4Addresses   []string `json:"ipv4EndpointAddresses,omitempty"`
	IPv6Addresses   []string `json:"ipv6EndpointAddresses,omitempty"`
	NetworkInstance string   `json:"networkInstance,omitempty"`
}

// IsValid validates the NF profile
func (p *NFProfile) IsValid() bool {
	if p.NFInstanceID == "" {
		return false
	}
	if p.NFType == "" {
		return false
	}
	if p.NFStatus == "" {
		return false
	}
	return true
}

// UpdateHeartbeat updates the last heartbeat time
func (p *NFProfile) UpdateHeartbeat() {
	p.LastHeartbeat = time.Now()
	p.UpdatedAt = time.Now()
}

// IsExpired checks if the NF has expired based on heartbeat timer
func (p *NFProfile) IsExpired() bool {
	if p.HeartBeatTimer == 0 {
		return false // No heartbeat requirement
	}

	timeout := time.Duration(p.HeartBeatTimer) * time.Second
	expiryTime := p.LastHeartbeat.Add(timeout)
	return time.Now().After(expiryTime)
}
