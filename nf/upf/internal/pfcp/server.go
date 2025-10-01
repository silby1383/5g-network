package pfcp

import (
	"context"
	"encoding/binary"
	"fmt"
	"net"
	"time"

	"github.com/your-org/5g-network/nf/upf/internal/config"
	upfcontext "github.com/your-org/5g-network/nf/upf/internal/context"
	"go.uber.org/zap"
)

// PFCP Message Types (3GPP TS 29.244)
const (
	PFCP_HEARTBEAT_REQUEST              = 1
	PFCP_HEARTBEAT_RESPONSE             = 2
	PFCP_ASSOCIATION_SETUP_REQUEST      = 5
	PFCP_ASSOCIATION_SETUP_RESPONSE     = 6
	PFCP_SESSION_ESTABLISHMENT_REQUEST  = 50
	PFCP_SESSION_ESTABLISHMENT_RESPONSE = 51
	PFCP_SESSION_MODIFICATION_REQUEST   = 52
	PFCP_SESSION_MODIFICATION_RESPONSE  = 53
	PFCP_SESSION_DELETION_REQUEST       = 54
	PFCP_SESSION_DELETION_RESPONSE      = 55
)

// PFCPServer handles PFCP protocol on N4 interface
type PFCPServer struct {
	config      *config.Config
	conn        *net.UDPConn
	upfContext  *upfcontext.UPFContext
	logger      *zap.Logger
	smfAddr     *net.UDPAddr
	sequenceNum uint32
}

// PFCPHeader represents PFCP message header
type PFCPHeader struct {
	Version        uint8
	MessageType    uint8
	MessageLength  uint16
	SEID           uint64
	SequenceNumber uint32
}

// NewPFCPServer creates a new PFCP server
func NewPFCPServer(cfg *config.Config, upfCtx *upfcontext.UPFContext, logger *zap.Logger) *PFCPServer {
	return &PFCPServer{
		config:      cfg,
		upfContext:  upfCtx,
		logger:      logger,
		sequenceNum: 1,
	}
}

// Start starts the PFCP server
func (s *PFCPServer) Start(ctx context.Context) error {
	addr, err := net.ResolveUDPAddr("udp", s.config.GetPFCPAddress())
	if err != nil {
		return fmt.Errorf("failed to resolve PFCP address: %w", err)
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen on PFCP port: %w", err)
	}
	s.conn = conn

	s.logger.Info("PFCP server started",
		zap.String("address", s.config.GetPFCPAddress()),
		zap.String("node_id", s.config.PFCP.NodeID))

	// Handle incoming messages
	go s.handleMessages(ctx)

	// Send periodic heartbeats
	go s.sendHeartbeats(ctx)

	<-ctx.Done()
	return conn.Close()
}

// handleMessages processes incoming PFCP messages
func (s *PFCPServer) handleMessages(ctx context.Context) {
	buffer := make([]byte, 65535)

	for {
		select {
		case <-ctx.Done():
			return
		default:
			n, addr, err := s.conn.ReadFromUDP(buffer)
			if err != nil {
				s.logger.Error("Failed to read PFCP message", zap.Error(err))
				continue
			}

			// Parse header
			if n < 16 {
				s.logger.Warn("PFCP message too short", zap.Int("length", n))
				continue
			}

			header := s.parsePFCPHeader(buffer[:n])
			s.logger.Debug("Received PFCP message",
				zap.Uint8("type", header.MessageType),
				zap.Uint64("seid", header.SEID),
				zap.String("from", addr.String()))

			// Handle message based on type
			s.handleMessage(header, buffer[:n], addr)
		}
	}
}

// parsePFCPHeader parses PFCP message header
func (s *PFCPServer) parsePFCPHeader(data []byte) *PFCPHeader {
	header := &PFCPHeader{
		Version:       (data[0] >> 5) & 0x07,
		MessageType:   data[1],
		MessageLength: binary.BigEndian.Uint16(data[2:4]),
	}

	// Check if S flag is set (SEID present)
	if (data[0] & 0x01) == 1 {
		header.SEID = binary.BigEndian.Uint64(data[4:12])
		header.SequenceNumber = binary.BigEndian.Uint32(data[12:15]) >> 8
	} else {
		header.SequenceNumber = binary.BigEndian.Uint32(data[4:7]) >> 8
	}

	return header
}

// handleMessage routes messages to appropriate handlers
func (s *PFCPServer) handleMessage(header *PFCPHeader, data []byte, addr *net.UDPAddr) {
	switch header.MessageType {
	case PFCP_HEARTBEAT_REQUEST:
		s.handleHeartbeatRequest(header, addr)
	case PFCP_ASSOCIATION_SETUP_REQUEST:
		s.handleAssociationSetupRequest(header, data, addr)
	case PFCP_SESSION_ESTABLISHMENT_REQUEST:
		s.handleSessionEstablishmentRequest(header, data, addr)
	case PFCP_SESSION_MODIFICATION_REQUEST:
		s.handleSessionModificationRequest(header, data, addr)
	case PFCP_SESSION_DELETION_REQUEST:
		s.handleSessionDeletionRequest(header, data, addr)
	default:
		s.logger.Warn("Unsupported PFCP message type", zap.Uint8("type", header.MessageType))
	}
}

// handleHeartbeatRequest handles PFCP heartbeat request
func (s *PFCPServer) handleHeartbeatRequest(header *PFCPHeader, addr *net.UDPAddr) {
	response := s.buildHeartbeatResponse(header.SequenceNumber)
	s.sendResponse(response, addr)
	s.logger.Debug("Sent heartbeat response", zap.String("to", addr.String()))
}

// handleAssociationSetupRequest handles PFCP association setup
func (s *PFCPServer) handleAssociationSetupRequest(header *PFCPHeader, data []byte, addr *net.UDPAddr) {
	s.smfAddr = addr
	response := s.buildAssociationSetupResponse(header.SequenceNumber)
	s.sendResponse(response, addr)
	s.logger.Info("PFCP association established", zap.String("smf", addr.String()))
}

// handleSessionEstablishmentRequest handles session establishment
func (s *PFCPServer) handleSessionEstablishmentRequest(header *PFCPHeader, data []byte, addr *net.UDPAddr) {
	// Create new session
	session := s.upfContext.CreateSession(header.SEID)

	// Allocate UPF F-TEID for N3
	session.UPFTEID = s.upfContext.AllocateTEID()

	s.logger.Info("PFCP session established",
		zap.Uint64("seid", header.SEID),
		zap.Uint32("upf_teid", session.UPFTEID))

	// Build and send response
	response := s.buildSessionEstablishmentResponse(header.SequenceNumber, header.SEID, session.UPFTEID)
	s.sendResponse(response, addr)
}

// handleSessionModificationRequest handles session modification
func (s *PFCPServer) handleSessionModificationRequest(header *PFCPHeader, data []byte, addr *net.UDPAddr) {
	_, exists := s.upfContext.GetSession(header.SEID)
	if !exists {
		s.logger.Error("Session not found", zap.Uint64("seid", header.SEID))
		return
	}

	// Update session (simplified - full implementation would parse IEs)
	s.upfContext.UpdateActivity(header.SEID)

	s.logger.Info("PFCP session modified", zap.Uint64("seid", header.SEID))

	response := s.buildSessionModificationResponse(header.SequenceNumber, header.SEID)
	s.sendResponse(response, addr)
}

// handleSessionDeletionRequest handles session deletion
func (s *PFCPServer) handleSessionDeletionRequest(header *PFCPHeader, data []byte, addr *net.UDPAddr) {
	s.upfContext.DeleteSession(header.SEID)

	s.logger.Info("PFCP session deleted", zap.Uint64("seid", header.SEID))

	response := s.buildSessionDeletionResponse(header.SequenceNumber, header.SEID)
	s.sendResponse(response, addr)
}

// sendHeartbeats sends periodic heartbeats to SMF
func (s *PFCPServer) sendHeartbeats(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if s.smfAddr != nil {
				request := s.buildHeartbeatRequest()
				s.sendResponse(request, s.smfAddr)
			}
		}
	}
}

// Helper functions to build PFCP messages (simplified)
func (s *PFCPServer) buildHeartbeatResponse(seqNum uint32) []byte {
	msg := make([]byte, 12)
	msg[0] = 0x20 // Version 1, no S flag
	msg[1] = PFCP_HEARTBEAT_RESPONSE
	binary.BigEndian.PutUint16(msg[2:4], 8) // Length
	msg[4] = byte(seqNum >> 16)
	msg[5] = byte(seqNum >> 8)
	msg[6] = byte(seqNum)
	return msg
}

func (s *PFCPServer) buildHeartbeatRequest() []byte {
	msg := make([]byte, 12)
	msg[0] = 0x20
	msg[1] = PFCP_HEARTBEAT_REQUEST
	binary.BigEndian.PutUint16(msg[2:4], 8)
	seqNum := s.sequenceNum
	s.sequenceNum++
	msg[4] = byte(seqNum >> 16)
	msg[5] = byte(seqNum >> 8)
	msg[6] = byte(seqNum)
	return msg
}

func (s *PFCPServer) buildAssociationSetupResponse(seqNum uint32) []byte {
	msg := make([]byte, 20)
	msg[0] = 0x20
	msg[1] = PFCP_ASSOCIATION_SETUP_RESPONSE
	binary.BigEndian.PutUint16(msg[2:4], 16)
	msg[4] = byte(seqNum >> 16)
	msg[5] = byte(seqNum >> 8)
	msg[6] = byte(seqNum)
	// Add Cause IE: Request accepted
	msg[8] = 0x00
	msg[9] = 0x13 // Cause IE type
	binary.BigEndian.PutUint16(msg[10:12], 1)
	msg[12] = 0x01 // Cause: Request accepted
	return msg
}

func (s *PFCPServer) buildSessionEstablishmentResponse(seqNum uint32, seid uint64, teid uint32) []byte {
	msg := make([]byte, 32)
	msg[0] = 0x21 // Version 1, S flag set
	msg[1] = PFCP_SESSION_ESTABLISHMENT_RESPONSE
	binary.BigEndian.PutUint16(msg[2:4], 28)
	binary.BigEndian.PutUint64(msg[4:12], seid)
	msg[12] = byte(seqNum >> 16)
	msg[13] = byte(seqNum >> 8)
	msg[14] = byte(seqNum)
	// Add Cause IE
	msg[16] = 0x00
	msg[17] = 0x13
	binary.BigEndian.PutUint16(msg[18:20], 1)
	msg[20] = 0x01 // Accepted
	// Add F-TEID IE (simplified)
	msg[22] = 0x00
	msg[23] = 0x15 // F-TEID type
	binary.BigEndian.PutUint16(msg[24:26], 5)
	msg[26] = 0x01 // CH flag
	binary.BigEndian.PutUint32(msg[27:31], teid)
	return msg
}

func (s *PFCPServer) buildSessionModificationResponse(seqNum uint32, seid uint64) []byte {
	msg := make([]byte, 24)
	msg[0] = 0x21
	msg[1] = PFCP_SESSION_MODIFICATION_RESPONSE
	binary.BigEndian.PutUint16(msg[2:4], 20)
	binary.BigEndian.PutUint64(msg[4:12], seid)
	msg[12] = byte(seqNum >> 16)
	msg[13] = byte(seqNum >> 8)
	msg[14] = byte(seqNum)
	// Cause
	msg[16] = 0x00
	msg[17] = 0x13
	binary.BigEndian.PutUint16(msg[18:20], 1)
	msg[20] = 0x01
	return msg
}

func (s *PFCPServer) buildSessionDeletionResponse(seqNum uint32, seid uint64) []byte {
	msg := make([]byte, 24)
	msg[0] = 0x21
	msg[1] = PFCP_SESSION_DELETION_RESPONSE
	binary.BigEndian.PutUint16(msg[2:4], 20)
	binary.BigEndian.PutUint64(msg[4:12], seid)
	msg[12] = byte(seqNum >> 16)
	msg[13] = byte(seqNum >> 8)
	msg[14] = byte(seqNum)
	// Cause
	msg[16] = 0x00
	msg[17] = 0x13
	binary.BigEndian.PutUint16(msg[18:20], 1)
	msg[20] = 0x01
	return msg
}

func (s *PFCPServer) sendResponse(msg []byte, addr *net.UDPAddr) {
	_, err := s.conn.WriteToUDP(msg, addr)
	if err != nil {
		s.logger.Error("Failed to send PFCP response", zap.Error(err))
	}
}
