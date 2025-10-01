package context

import (
	"net"
	"sync"
	"time"
)

// UPFSession represents a PDU session in the UPF
type UPFSession struct {
	SEID         uint64 // F-SEID (Session Endpoint Identifier)
	SMFSEID      uint64 // SMF's F-SEID
	UEAddress    net.IP // UE IP address
	GNBTEID      uint32 // gNB Tunnel Endpoint ID (N3)
	UPFTEID      uint32 // UPF Tunnel Endpoint ID (N3)
	GNBAddress   net.IP // gNB IP address
	DNN          string // Data Network Name
	PDRs         []PDR  // Packet Detection Rules
	FARs         []FAR  // Forwarding Action Rules
	QERs         []QER  // QoS Enforcement Rules
	CreatedAt    time.Time
	LastActivity time.Time
}

// PDR represents a Packet Detection Rule (3GPP TS 29.244)
type PDR struct {
	PDRID              uint16 // PDR ID
	Precedence         uint32 // Rule precedence
	PDI                PDI    // Packet Detection Information
	OuterHeaderRemoval uint8  // 0=None, 1=GTP-U/UDP/IPv4
	FARID              uint32 // Forwarding Action Rule ID
	QERID              uint32 // QoS Enforcement Rule ID
}

// PDI represents Packet Detection Information
type PDI struct {
	SourceInterface uint8  // 0=Access (N3), 1=Core (N6), 2=SGi-LAN, 3=CP-function
	NetworkInstance string // DNN/APN
	FTEID           *FTEID // F-TEID for GTP-U
	UEIPAddress     net.IP // UE IP address
	SDFFilter       string // Service Data Flow filter
}

// FAR represents a Forwarding Action Rule (3GPP TS 29.244)
type FAR struct {
	FARID                 uint32 // FAR ID
	ApplyAction           uint8  // 0=DROP, 1=FORW, 2=BUFF, 3=NOCP, 4=DUPL
	ForwardingParameters  *ForwardingParameters
	DuplicatingParameters *DuplicatingParameters
}

// ForwardingParameters for FAR
type ForwardingParameters struct {
	DestinationInterface uint8  // 0=Access (N3), 1=Core (N6)
	NetworkInstance      string // DNN
	OuterHeaderCreation  *OuterHeaderCreation
	ForwardingPolicy     string
}

// OuterHeaderCreation for GTP-U encapsulation
type OuterHeaderCreation struct {
	Description uint16 // Bit flags: GTP-U/UDP/IPv4
	TEID        uint32 // Tunnel Endpoint ID
	IPv4Address net.IP // Peer IP address
	IPv6Address net.IP
	Port        uint16
}

// DuplicatingParameters for traffic duplication
type DuplicatingParameters struct {
	DestinationInterface uint8
	OuterHeaderCreation  *OuterHeaderCreation
}

// QER represents a QoS Enforcement Rule (3GPP TS 29.244)
type QER struct {
	QERID      uint32 // QER ID
	QFI        uint8  // QoS Flow Identifier
	MBR        *MBR   // Maximum Bit Rate
	GBR        *GBR   // Guaranteed Bit Rate
	PacketRate *PacketRate
	GateStatus uint8 // 0=OPEN, 1=CLOSED
}

// MBR represents Maximum Bit Rate
type MBR struct {
	Uplink   uint64 // bps
	Downlink uint64 // bps
}

// GBR represents Guaranteed Bit Rate
type GBR struct {
	Uplink   uint64 // bps
	Downlink uint64 // bps
}

// PacketRate for rate limiting
type PacketRate struct {
	Uplink   uint64 // packets per second
	Downlink uint64 // packets per second
}

// FTEID represents Fully Qualified Tunnel Endpoint Identifier
type FTEID struct {
	TEID        uint32
	IPv4Address net.IP
	IPv6Address net.IP
	ChooseID    uint8 // 0=Use provided, 1=UPF allocates
}

// UPFContext manages all UPF sessions
type UPFContext struct {
	sessions map[uint64]*UPFSession // Key: SEID
	mu       sync.RWMutex
	teidPool *TEIDPool
}

// TEIDPool manages TEID allocation
type TEIDPool struct {
	nextTEID uint32
	used     map[uint32]bool
	mu       sync.Mutex
}

// NewUPFContext creates a new UPF context
func NewUPFContext() *UPFContext {
	return &UPFContext{
		sessions: make(map[uint64]*UPFSession),
		teidPool: &TEIDPool{
			nextTEID: 1,
			used:     make(map[uint32]bool),
		},
	}
}

// CreateSession creates a new PDU session
func (c *UPFContext) CreateSession(seid uint64) *UPFSession {
	c.mu.Lock()
	defer c.mu.Unlock()

	session := &UPFSession{
		SEID:         seid,
		PDRs:         make([]PDR, 0),
		FARs:         make([]FAR, 0),
		QERs:         make([]QER, 0),
		CreatedAt:    time.Now(),
		LastActivity: time.Now(),
	}

	c.sessions[seid] = session
	return session
}

// GetSession retrieves a session by SEID
func (c *UPFContext) GetSession(seid uint64) (*UPFSession, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	session, exists := c.sessions[seid]
	return session, exists
}

// DeleteSession removes a session
func (c *UPFContext) DeleteSession(seid uint64) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if session, exists := c.sessions[seid]; exists {
		// Release TEIDs
		c.teidPool.Release(session.UPFTEID)
		delete(c.sessions, seid)
	}
}

// GetAllSessions returns all active sessions
func (c *UPFContext) GetAllSessions() []*UPFSession {
	c.mu.RLock()
	defer c.mu.RUnlock()

	sessions := make([]*UPFSession, 0, len(c.sessions))
	for _, session := range c.sessions {
		sessions = append(sessions, session)
	}
	return sessions
}

// UpdateActivity updates the last activity time
func (c *UPFContext) UpdateActivity(seid uint64) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if session, exists := c.sessions[seid]; exists {
		session.LastActivity = time.Now()
	}
}

// AllocateTEID allocates a new TEID from the pool
func (c *UPFContext) AllocateTEID() uint32 {
	return c.teidPool.Allocate()
}

// AllocateTEID allocates a new TEID
func (p *TEIDPool) Allocate() uint32 {
	p.mu.Lock()
	defer p.mu.Unlock()

	for p.used[p.nextTEID] {
		p.nextTEID++
		if p.nextTEID == 0 {
			p.nextTEID = 1 // Skip 0
		}
	}

	teid := p.nextTEID
	p.used[teid] = true
	p.nextTEID++

	return teid
}

// Release releases a TEID
func (p *TEIDPool) Release(teid uint32) {
	p.mu.Lock()
	defer p.mu.Unlock()

	delete(p.used, teid)
}

// GetStats returns UPF statistics
func (c *UPFContext) GetStats() map[string]interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return map[string]interface{}{
		"total_sessions":  len(c.sessions),
		"active_sessions": len(c.sessions), // TODO: Filter by activity
	}
}
