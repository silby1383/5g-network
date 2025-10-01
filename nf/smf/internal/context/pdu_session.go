package context

import (
	"sync"
	"time"
)

// PDUSessionType represents the type of PDU session
type PDUSessionType string

const (
	PDUSessionTypeIPv4     PDUSessionType = "IPV4"
	PDUSessionTypeIPv6     PDUSessionType = "IPV6"
	PDUSessionTypeIPv4v6   PDUSessionType = "IPV4V6"
	PDUSessionTypeEthernet PDUSessionType = "ETHERNET"
)

// PDUSessionState represents the state of a PDU session
type PDUSessionState string

const (
	PDUSessionStateInactive      PDUSessionState = "INACTIVE"
	PDUSessionStateActivePending PDUSessionState = "ACTIVE_PENDING"
	PDUSessionStateActive        PDUSessionState = "ACTIVE"
	PDUSessionStateModifying     PDUSessionState = "MODIFYING"
	PDUSessionStateReleasing     PDUSessionState = "RELEASING"
	PDUSessionStateReleased      PDUSessionState = "RELEASED"
)

// SSCMode represents Session and Service Continuity mode
type SSCMode int

const (
	SSCMode1 SSCMode = 1 // Connection stability
	SSCMode2 SSCMode = 2 // Connection relocation
	SSCMode3 SSCMode = 3 // Connection termination and re-establishment
)

// SNSSAI represents Single Network Slice Selection Assistance Information
type SNSSAI struct {
	SST int    `json:"sst"` // Slice/Service Type
	SD  string `json:"sd"`  // Slice Differentiator
}

// QoSFlowIdentifier represents QoS Flow ID (5QI)
type QoSFlowIdentifier uint8

// QoSFlow represents a QoS flow within a PDU session
type QoSFlow struct {
	QFI       QoSFlowIdentifier `json:"qfi"`           // QoS Flow Identifier (1-63)
	FiveQI    uint8             `json:"fiveQI"`        // 5G QoS Identifier
	Priority  uint8             `json:"priority"`      // Allocation and Retention Priority
	GBR       *BitRate          `json:"gbr,omitempty"` // Guaranteed Bit Rate (for GBR flows)
	MBR       *BitRate          `json:"mbr,omitempty"` // Maximum Bit Rate (for GBR flows)
	CreatedAt time.Time         `json:"createdAt"`
}

// BitRate represents uplink and downlink bit rates
type BitRate struct {
	Uplink   uint64 `json:"uplink"`   // bps
	Downlink uint64 `json:"downlink"` // bps
}

// PDUSession represents a PDU session
type PDUSession struct {
	mu sync.RWMutex

	// Session Identification
	SUPI         string `json:"supi"`
	PDUSessionID uint8  `json:"pduSessionId"` // 1-15
	DNN          string `json:"dnn"`
	SNSSAI       SNSSAI `json:"snssai"`

	// Session Characteristics
	PDUSessionType PDUSessionType  `json:"pduSessionType"`
	SSCMode        SSCMode         `json:"sscMode"`
	State          PDUSessionState `json:"state"`

	// UE IP Address
	UEIPv4Address string `json:"ueIpv4Address,omitempty"`
	UEIPv6Prefix  string `json:"ueIpv6Prefix,omitempty"`

	// Session AMBR
	SessionAMBR BitRate `json:"sessionAmbr"`

	// QoS Flows
	QoSFlows map[QoSFlowIdentifier]*QoSFlow `json:"qosFlows"`

	// UPF Information
	UPFNodeID       string `json:"upfNodeId"`
	UPFN4Address    string `json:"upfN4Address"`
	UPFTEIDUplink   uint32 `json:"upfTeidUplink"`   // F-TEID for uplink
	UPFTEIDDownlink uint32 `json:"upfTeidDownlink"` // F-TEID for downlink

	// gNB Information (via AMF)
	GNBTEIDUplink uint32 `json:"gnbTeidUplink"`
	GNBN3Address  string `json:"gnbN3Address"`

	// Timestamps
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// NewPDUSession creates a new PDU session
func NewPDUSession(supi string, pduSessionID uint8, dnn string, snssai SNSSAI) *PDUSession {
	now := time.Now()
	return &PDUSession{
		SUPI:           supi,
		PDUSessionID:   pduSessionID,
		DNN:            dnn,
		SNSSAI:         snssai,
		PDUSessionType: PDUSessionTypeIPv4,
		SSCMode:        SSCMode1,
		State:          PDUSessionStateInactive,
		QoSFlows:       make(map[QoSFlowIdentifier]*QoSFlow),
		CreatedAt:      now,
		UpdatedAt:      now,
	}
}

// UpdateState updates the session state
func (s *PDUSession) UpdateState(state PDUSessionState) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.State = state
	s.UpdatedAt = time.Now()
}

// GetState returns the current session state
func (s *PDUSession) GetState() PDUSessionState {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.State
}

// AddQoSFlow adds a QoS flow to the session
func (s *PDUSession) AddQoSFlow(flow *QoSFlow) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.QoSFlows[flow.QFI] = flow
	s.UpdatedAt = time.Now()
}

// RemoveQoSFlow removes a QoS flow from the session
func (s *PDUSession) RemoveQoSFlow(qfi QoSFlowIdentifier) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.QoSFlows, qfi)
	s.UpdatedAt = time.Now()
}

// SetUEIPAddress sets the UE IP address
func (s *PDUSession) SetUEIPAddress(ipv4 string, ipv6Prefix string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.UEIPv4Address = ipv4
	s.UEIPv6Prefix = ipv6Prefix
	s.UpdatedAt = time.Now()
}

// SetUPFInfo sets UPF information
func (s *PDUSession) SetUPFInfo(nodeID, n4Address string, teidUplink, teidDownlink uint32) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.UPFNodeID = nodeID
	s.UPFN4Address = n4Address
	s.UPFTEIDUplink = teidUplink
	s.UPFTEIDDownlink = teidDownlink
	s.UpdatedAt = time.Now()
}

// SetGNBInfo sets gNB information
func (s *PDUSession) SetGNBInfo(teidUplink uint32, n3Address string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.GNBTEIDUplink = teidUplink
	s.GNBN3Address = n3Address
	s.UpdatedAt = time.Now()
}

// SetSessionAMBR sets the session AMBR
func (s *PDUSession) SetSessionAMBR(uplink, downlink uint64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.SessionAMBR = BitRate{
		Uplink:   uplink,
		Downlink: downlink,
	}
	s.UpdatedAt = time.Now()
}
