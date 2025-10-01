package service

import (
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/your-org/5g-network/nf/smf/internal/config"
	"github.com/your-org/5g-network/nf/smf/internal/context"
	"github.com/your-org/5g-network/nf/smf/internal/n4"
	"go.uber.org/zap"
)

// SessionService handles PDU session management
// 3GPP TS 23.502 - Procedures for the 5G System
// 3GPP TS 29.502 - Session Management Services
type SessionService struct {
	config     *config.Config
	smfContext *context.SMFContext
	pfcpClient *n4.PFCPClient
	logger     *zap.Logger
	ueIPPool   *IPPool
}

// NewSessionService creates a new session service
func NewSessionService(
	cfg *config.Config,
	smfContext *context.SMFContext,
	pfcpClient *n4.PFCPClient,
	logger *zap.Logger,
) (*SessionService, error) {
	// Initialize UE IP pool
	ipPool, err := NewIPPool(cfg.SMF.UESubnet.IPv4)
	if err != nil {
		return nil, fmt.Errorf("failed to create IP pool: %w", err)
	}

	return &SessionService{
		config:     cfg,
		smfContext: smfContext,
		pfcpClient: pfcpClient,
		logger:     logger,
		ueIPPool:   ipPool,
	}, nil
}

// CreateSessionRequest represents a PDU session creation request from AMF
type CreateSessionRequest struct {
	SUPI           string         `json:"supi"`
	PDUSessionID   uint8          `json:"pduSessionId"`
	DNN            string         `json:"dnn"`
	SNSSAI         context.SNSSAI `json:"snssai"`
	PDUSessionType string         `json:"pduSessionType"`

	// From gNB (via AMF)
	GNBN3Address  string `json:"gnbN3Address"`
	GNBTEIDUplink uint32 `json:"gnbTeidUplink"`
}

// CreateSessionResponse represents a PDU session creation response
type CreateSessionResponse struct {
	Result        string          `json:"result"` // "SUCCESS", "FAILURE"
	SUPI          string          `json:"supi"`
	PDUSessionID  uint8           `json:"pduSessionId"`
	UEIPv4Address string          `json:"ueIpv4Address,omitempty"`
	SessionAMBR   context.BitRate `json:"sessionAmbr"`
	QoSFlows      []QoSFlowInfo   `json:"qosFlows"`

	// For gNB (via AMF)
	UPFN3Address    string `json:"upfN3Address"`
	UPFTEIDDownlink uint32 `json:"upfTeidDownlink"`

	Reason string `json:"reason,omitempty"`
}

// QoSFlowInfo represents QoS flow information
type QoSFlowInfo struct {
	QFI      uint8 `json:"qfi"`
	FiveQI   uint8 `json:"fiveQI"`
	Priority uint8 `json:"priority"`
}

// UpdateSessionRequest represents a PDU session update request
type UpdateSessionRequest struct {
	SUPI             string        `json:"supi"`
	PDUSessionID     uint8         `json:"pduSessionId"`
	QoSFlowsToAdd    []QoSFlowInfo `json:"qosFlowsToAdd,omitempty"`
	QoSFlowsToRemove []uint8       `json:"qosFlowsToRemove,omitempty"`
}

// UpdateSessionResponse represents a PDU session update response
type UpdateSessionResponse struct {
	Result       string `json:"result"`
	SUPI         string `json:"supi"`
	PDUSessionID uint8  `json:"pduSessionId"`
	Reason       string `json:"reason,omitempty"`
}

// ReleaseSessionRequest represents a PDU session release request
type ReleaseSessionRequest struct {
	SUPI         string `json:"supi"`
	PDUSessionID uint8  `json:"pduSessionId"`
	Cause        string `json:"cause,omitempty"`
}

// ReleaseSessionResponse represents a PDU session release response
type ReleaseSessionResponse struct {
	Result       string `json:"result"`
	SUPI         string `json:"supi"`
	PDUSessionID uint8  `json:"pduSessionId"`
	Reason       string `json:"reason,omitempty"`
}

// CreateSession handles PDU session creation
func (s *SessionService) CreateSession(req *CreateSessionRequest) (*CreateSessionResponse, error) {
	s.logger.Info("Creating PDU session",
		zap.String("supi", req.SUPI),
		zap.Uint8("pdu_session_id", req.PDUSessionID),
		zap.String("dnn", req.DNN),
		zap.Int("sst", req.SNSSAI.SST),
		zap.String("sd", req.SNSSAI.SD),
	)

	// 1. Create PDU session context
	session := context.NewPDUSession(req.SUPI, req.PDUSessionID, req.DNN, req.SNSSAI)
	session.SetGNBInfo(req.GNBTEIDUplink, req.GNBN3Address)

	// 2. Allocate UE IP address
	ueIP, err := s.ueIPPool.Allocate()
	if err != nil {
		return &CreateSessionResponse{
			Result: "FAILURE",
			Reason: fmt.Sprintf("failed to allocate UE IP: %v", err),
		}, err
	}
	session.SetUEIPAddress(ueIP, "")

	// 3. Set Session AMBR (from policy or default)
	session.SetSessionAMBR(1000000000, 2000000000) // 1 Gbps UL, 2 Gbps DL

	// 4. Add default QoS flow (QFI=1, 5QI=9 for internet)
	defaultQoSFlow := &context.QoSFlow{
		QFI:       1,
		FiveQI:    9, // Non-GBR, internet
		Priority:  10,
		CreatedAt: time.Now(),
	}
	session.AddQoSFlow(defaultQoSFlow)

	// 5. Get UPF information
	upfNodeID, upfN4Addr := s.smfContext.GetUPFInfo()

	// 6. Generate SEID for PFCP session
	seid := n4.GenerateSEID(req.SUPI, req.PDUSessionID)

	// 7. Build PFCP Session Establishment Request
	pfcpReq := s.buildPFCPEstablishmentRequest(session, seid, upfNodeID)

	// 8. Send PFCP Session Establishment to UPF
	session.UpdateState(context.PDUSessionStateActivePending)

	pfcpResp, err := s.pfcpClient.EstablishSession(pfcpReq)
	if err != nil {
		s.logger.Error("PFCP session establishment failed", zap.Error(err))
		s.ueIPPool.Release(ueIP)
		return &CreateSessionResponse{
			Result: "FAILURE",
			Reason: fmt.Sprintf("PFCP establishment failed: %v", err),
		}, err
	}

	// 9. Validate PFCP response
	if err := n4.ValidatePFCPResponse(pfcpResp.Cause); err != nil {
		s.logger.Error("PFCP response invalid", zap.Error(err))
		s.ueIPPool.Release(ueIP)
		return &CreateSessionResponse{
			Result: "FAILURE",
			Reason: fmt.Sprintf("PFCP response invalid: %v", err),
		}, err
	}

	// 10. Update session with UPF information
	session.SetUPFInfo(
		upfNodeID,
		upfN4Addr,
		pfcpResp.UPFTEID.TEID,
		pfcpResp.UPFTEID.TEID, // Use same TEID for simplicity
	)

	// 11. Update session state to active
	session.UpdateState(context.PDUSessionStateActive)

	// 12. Add session to SMF context
	if err := s.smfContext.AddSession(session); err != nil {
		s.logger.Error("Failed to add session to context", zap.Error(err))
		s.ueIPPool.Release(ueIP)
		return &CreateSessionResponse{
			Result: "FAILURE",
			Reason: fmt.Sprintf("failed to add session: %v", err),
		}, err
	}

	s.logger.Info("PDU session created successfully",
		zap.String("supi", req.SUPI),
		zap.Uint8("pdu_session_id", req.PDUSessionID),
		zap.String("ue_ip", ueIP),
		zap.Uint32("upf_teid", pfcpResp.UPFTEID.TEID),
	)

	// 13. Build response
	return &CreateSessionResponse{
		Result:        "SUCCESS",
		SUPI:          req.SUPI,
		PDUSessionID:  req.PDUSessionID,
		UEIPv4Address: ueIP,
		SessionAMBR:   session.SessionAMBR,
		QoSFlows: []QoSFlowInfo{
			{
				QFI:      uint8(defaultQoSFlow.QFI),
				FiveQI:   defaultQoSFlow.FiveQI,
				Priority: defaultQoSFlow.Priority,
			},
		},
		UPFN3Address:    pfcpResp.UPFTEID.IPv4,
		UPFTEIDDownlink: pfcpResp.UPFTEID.TEID,
	}, nil
}

// ReleaseSession handles PDU session release
func (s *SessionService) ReleaseSession(req *ReleaseSessionRequest) (*ReleaseSessionResponse, error) {
	s.logger.Info("Releasing PDU session",
		zap.String("supi", req.SUPI),
		zap.Uint8("pdu_session_id", req.PDUSessionID),
		zap.String("cause", req.Cause),
	)

	// 1. Get session from context
	session, err := s.smfContext.GetSession(req.SUPI, req.PDUSessionID)
	if err != nil {
		return &ReleaseSessionResponse{
			Result: "FAILURE",
			Reason: fmt.Sprintf("session not found: %v", err),
		}, err
	}

	// 2. Update session state
	session.UpdateState(context.PDUSessionStateReleasing)

	// 3. Generate SEID
	seid := n4.GenerateSEID(req.SUPI, req.PDUSessionID)

	// 4. Send PFCP Session Deletion to UPF
	pfcpReq := &n4.SessionDeletionRequest{
		SEID: seid,
	}

	pfcpResp, err := s.pfcpClient.DeleteSession(pfcpReq)
	if err != nil {
		s.logger.Error("PFCP session deletion failed", zap.Error(err))
		// Continue with local cleanup
	} else if err := n4.ValidatePFCPResponse(pfcpResp.Cause); err != nil {
		s.logger.Error("PFCP deletion response invalid", zap.Error(err))
	}

	// 5. Release UE IP address
	s.ueIPPool.Release(session.UEIPv4Address)

	// 6. Remove session from context
	if err := s.smfContext.RemoveSession(req.SUPI, req.PDUSessionID); err != nil {
		s.logger.Error("Failed to remove session from context", zap.Error(err))
	}

	s.logger.Info("PDU session released successfully",
		zap.String("supi", req.SUPI),
		zap.Uint8("pdu_session_id", req.PDUSessionID),
	)

	return &ReleaseSessionResponse{
		Result:       "SUCCESS",
		SUPI:         req.SUPI,
		PDUSessionID: req.PDUSessionID,
	}, nil
}

// buildPFCPEstablishmentRequest builds PFCP Session Establishment Request
func (s *SessionService) buildPFCPEstablishmentRequest(
	session *context.PDUSession,
	seid uint64,
	upfNodeID string,
) *n4.SessionEstablishmentRequest {
	// Build PDRs (Packet Detection Rules)
	pdrs := []n4.PDR{
		// PDR for uplink (from UE to DN)
		{
			PDRID:      1,
			Precedence: 100,
			PDI: n4.PDI{
				SourceInterface: "ACCESS",
				FTEID: &n4.FTEID{
					TEID: session.GNBTEIDUplink,
					IPv4: session.GNBN3Address,
				},
				UEIPAddress:     session.UEIPv4Address,
				NetworkInstance: session.DNN,
			},
			OuterHeaderRemoval: true,
			FARID:              1,
			QERID:              1,
		},
		// PDR for downlink (from DN to UE)
		{
			PDRID:      2,
			Precedence: 100,
			PDI: n4.PDI{
				SourceInterface: "CORE",
				UEIPAddress:     session.UEIPv4Address,
				NetworkInstance: session.DNN,
			},
			FARID: 2,
			QERID: 1,
		},
	}

	// Build FARs (Forwarding Action Rules)
	fars := []n4.FAR{
		// FAR for uplink
		{
			FARID:       1,
			ApplyAction: "FORWARD",
			ForwardingParameters: &n4.ForwardingParameters{
				DestinationInterface: "CORE",
				NetworkInstance:      session.DNN,
			},
		},
		// FAR for downlink
		{
			FARID:       2,
			ApplyAction: "FORWARD",
			ForwardingParameters: &n4.ForwardingParameters{
				DestinationInterface: "ACCESS",
				NetworkInstance:      session.DNN,
				OuterHeaderCreation: &n4.OuterHeaderCreation{
					TEID: session.GNBTEIDUplink,
					IPv4: session.GNBN3Address,
				},
			},
		},
	}

	// Build QERs (QoS Enforcement Rules)
	qers := []n4.QER{
		{
			QERID:       1,
			QFI:         1,
			MBRUplink:   session.SessionAMBR.Uplink,
			MBRDownlink: session.SessionAMBR.Downlink,
		},
	}

	return &n4.SessionEstablishmentRequest{
		NodeID:        upfNodeID,
		SEID:          seid,
		UEIPv4Address: session.UEIPv4Address,
		DNN:           session.DNN,
		PDRs:          pdrs,
		FARs:          fars,
		QERs:          qers,
	}
}

// GetSessionStatistics returns session statistics
func (s *SessionService) GetSessionStatistics() map[string]interface{} {
	stats := s.smfContext.GetStatistics()
	return map[string]interface{}{
		"total_sessions":    stats.TotalSessions,
		"active_sessions":   stats.ActiveSessions,
		"released_sessions": stats.ReleasedSessions,
		"allocated_ue_ips":  s.ueIPPool.AllocatedCount(),
	}
}

// IPPool manages UE IP address allocation
type IPPool struct {
	subnet    *net.IPNet
	allocated map[string]bool
	mu        sync.Mutex
}

// NewIPPool creates a new IP pool
func NewIPPool(cidr string) (*IPPool, error) {
	_, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, fmt.Errorf("invalid CIDR: %w", err)
	}

	return &IPPool{
		subnet:    ipNet,
		allocated: make(map[string]bool),
	}, nil
}

// Allocate allocates a new IP address
func (p *IPPool) Allocate() (string, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Simple allocation - increment from network base
	ip := make(net.IP, len(p.subnet.IP))
	copy(ip, p.subnet.IP)

	// Start from .1 (skip .0)
	ip[len(ip)-1]++

	for p.subnet.Contains(ip) {
		ipStr := ip.String()
		if !p.allocated[ipStr] {
			p.allocated[ipStr] = true
			return ipStr, nil
		}

		// Increment IP
		for i := len(ip) - 1; i >= 0; i-- {
			ip[i]++
			if ip[i] != 0 {
				break
			}
		}
	}

	return "", fmt.Errorf("IP pool exhausted")
}

// Release releases an IP address back to the pool
func (p *IPPool) Release(ip string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	delete(p.allocated, ip)
}

// AllocatedCount returns the number of allocated IPs
func (p *IPPool) AllocatedCount() int {
	p.mu.Lock()
	defer p.mu.Unlock()

	return len(p.allocated)
}
