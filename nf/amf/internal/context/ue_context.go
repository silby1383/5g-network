package context

import (
	"sync"
	"time"
)

// UEContext represents a UE (User Equipment) context in AMF
type UEContext struct {
	// UE Identity
	SUPI string // Subscription Permanent Identifier
	SUCI string // Subscription Concealed Identifier
	GPSI string // Generic Public Subscription Identifier
	PEI  string // Permanent Equipment Identifier

	// Registration State
	RegistrationState RegistrationState
	ConnectionState   ConnectionState

	// Location
	TAI TrackingAreaIdentity

	// Security
	SecurityContext *SecurityContext

	// Network Slicing
	AllowedNSSAI    []SNSSAI
	ConfiguredNSSAI []SNSSAI

	// AMF Context
	GUAMI       string // Globally Unique AMF Identifier
	AMFRegionID uint8
	AMFSetID    uint16
	AMFPointer  uint8

	// Timestamps
	CreatedAt      time.Time
	RegisteredAt   time.Time
	LastActivityAt time.Time

	// Session Info
	PDUSessions map[uint8]*PDUSessionInfo // Session ID -> Session Info

	mu sync.RWMutex
}

// RegistrationState represents UE registration state
type RegistrationState string

const (
	RegistrationStateDeregistered RegistrationState = "DEREGISTERED"
	RegistrationStateRegistered   RegistrationState = "REGISTERED"
)

// ConnectionState represents UE connection state
type ConnectionState string

const (
	ConnectionStateIdle      ConnectionState = "IDLE"
	ConnectionStateConnected ConnectionState = "CONNECTED"
)

// TrackingAreaIdentity represents Tracking Area Identity
type TrackingAreaIdentity struct {
	PLMNID PLMNID `json:"plmnId"`
	TAC    string `json:"tac"` // Tracking Area Code
}

// PLMNID represents Public Land Mobile Network ID
type PLMNID struct {
	MCC string `json:"mcc"` // Mobile Country Code
	MNC string `json:"mnc"` // Mobile Network Code
}

// SNSSAI represents Single Network Slice Selection Assistance Information
type SNSSAI struct {
	SST uint8  `json:"sst"`          // Slice/Service Type
	SD  string `json:"sd,omitempty"` // Slice Differentiator
}

// SecurityContext represents UE security context
type SecurityContext struct {
	// Keys
	KAMF    string // AMF key
	KSEAF   string // SEAF key (from AUSF)
	KNASenc string // NAS encryption key
	KNASint string // NAS integrity key

	// Security algorithms
	IntegrityAlgorithm string
	CipheringAlgorithm string

	// Counters
	UplinkNASCount   uint32
	DownlinkNASCount uint32

	// State
	NASSecurityEstablished bool
	ASSecurityEstablished  bool
}

// PDUSessionInfo represents PDU session information
type PDUSessionInfo struct {
	SessionID     uint8
	DNN           string
	SNSSAI        SNSSAI
	SessionAMBR   SessionAMBR
	SMFInstanceID string
	State         PDUSessionState
	CreatedAt     time.Time
}

// SessionAMBR represents Session Aggregate Maximum Bit Rate
type SessionAMBR struct {
	Uplink   uint64 // bps
	Downlink uint64 // bps
}

// PDUSessionState represents PDU session state
type PDUSessionState string

const (
	PDUSessionStateActive   PDUSessionState = "ACTIVE"
	PDUSessionStateInactive PDUSessionState = "INACTIVE"
	PDUSessionStateReleased PDUSessionState = "RELEASED"
)

// NewUEContext creates a new UE context
func NewUEContext(supi string) *UEContext {
	return &UEContext{
		SUPI:              supi,
		RegistrationState: RegistrationStateDeregistered,
		ConnectionState:   ConnectionStateIdle,
		PDUSessions:       make(map[uint8]*PDUSessionInfo),
		CreatedAt:         time.Now(),
	}
}

// UpdateRegistrationState updates the UE registration state
func (ue *UEContext) UpdateRegistrationState(state RegistrationState) {
	ue.mu.Lock()
	defer ue.mu.Unlock()

	ue.RegistrationState = state
	if state == RegistrationStateRegistered {
		ue.RegisteredAt = time.Now()
	}
	ue.LastActivityAt = time.Now()
}

// UpdateConnectionState updates the UE connection state
func (ue *UEContext) UpdateConnectionState(state ConnectionState) {
	ue.mu.Lock()
	defer ue.mu.Unlock()

	ue.ConnectionState = state
	ue.LastActivityAt = time.Now()
}

// SetSecurityContext sets the security context
func (ue *UEContext) SetSecurityContext(sc *SecurityContext) {
	ue.mu.Lock()
	defer ue.mu.Unlock()

	ue.SecurityContext = sc
	ue.LastActivityAt = time.Now()
}

// AddPDUSession adds a PDU session
func (ue *UEContext) AddPDUSession(session *PDUSessionInfo) {
	ue.mu.Lock()
	defer ue.mu.Unlock()

	ue.PDUSessions[session.SessionID] = session
	ue.LastActivityAt = time.Now()
}

// RemovePDUSession removes a PDU session
func (ue *UEContext) RemovePDUSession(sessionID uint8) {
	ue.mu.Lock()
	defer ue.mu.Unlock()

	delete(ue.PDUSessions, sessionID)
	ue.LastActivityAt = time.Now()
}

// GetPDUSession retrieves a PDU session
func (ue *UEContext) GetPDUSession(sessionID uint8) (*PDUSessionInfo, bool) {
	ue.mu.RLock()
	defer ue.mu.RUnlock()

	session, exists := ue.PDUSessions[sessionID]
	return session, exists
}

// IsRegistered checks if UE is registered
func (ue *UEContext) IsRegistered() bool {
	ue.mu.RLock()
	defer ue.mu.RUnlock()

	return ue.RegistrationState == RegistrationStateRegistered
}

// IsConnected checks if UE is connected
func (ue *UEContext) IsConnected() bool {
	ue.mu.RLock()
	defer ue.mu.RUnlock()

	return ue.ConnectionState == ConnectionStateConnected
}

// UEContextManager manages all UE contexts
type UEContextManager struct {
	contexts map[string]*UEContext // SUPI -> UE Context
	mu       sync.RWMutex
}

// NewUEContextManager creates a new UE context manager
func NewUEContextManager() *UEContextManager {
	return &UEContextManager{
		contexts: make(map[string]*UEContext),
	}
}

// CreateContext creates a new UE context
func (m *UEContextManager) CreateContext(supi string) *UEContext {
	m.mu.Lock()
	defer m.mu.Unlock()

	ctx := NewUEContext(supi)
	m.contexts[supi] = ctx
	return ctx
}

// GetContext retrieves a UE context by SUPI
func (m *UEContextManager) GetContext(supi string) (*UEContext, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	ctx, exists := m.contexts[supi]
	return ctx, exists
}

// GetOrCreateContext gets an existing context or creates a new one
func (m *UEContextManager) GetOrCreateContext(supi string) *UEContext {
	// Try to get first
	if ctx, exists := m.GetContext(supi); exists {
		return ctx
	}

	// Create new
	return m.CreateContext(supi)
}

// RemoveContext removes a UE context
func (m *UEContextManager) RemoveContext(supi string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.contexts, supi)
}

// GetAllContexts returns all UE contexts
func (m *UEContextManager) GetAllContexts() []*UEContext {
	m.mu.RLock()
	defer m.mu.RUnlock()

	contexts := make([]*UEContext, 0, len(m.contexts))
	for _, ctx := range m.contexts {
		contexts = append(contexts, ctx)
	}
	return contexts
}

// GetRegisteredCount returns the number of registered UEs
func (m *UEContextManager) GetRegisteredCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	count := 0
	for _, ctx := range m.contexts {
		if ctx.IsRegistered() {
			count++
		}
	}
	return count
}

// GetConnectedCount returns the number of connected UEs
func (m *UEContextManager) GetConnectedCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	count := 0
	for _, ctx := range m.contexts {
		if ctx.IsConnected() {
			count++
		}
	}
	return count
}
