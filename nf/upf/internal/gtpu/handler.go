package gtpu

import (
	"context"
	"encoding/binary"
	"fmt"
	"net"

	"github.com/your-org/5g-network/nf/upf/internal/config"
	upfcontext "github.com/your-org/5g-network/nf/upf/internal/context"
	"go.uber.org/zap"
)

// GTP-U Message Types (3GPP TS 29.281)
const (
	GTPU_ECHO_REQUEST     = 1
	GTPU_ECHO_RESPONSE    = 2
	GTPU_ERROR_INDICATION = 26
	GTPU_END_MARKER       = 254
	GTPU_G_PDU            = 255
)

// GTPUHandler handles GTP-U protocol on N3 interface
type GTPUHandler struct {
	config     *config.Config
	n3Conn     *net.UDPConn
	n6Conn     *net.UDPConn
	upfContext *upfcontext.UPFContext
	logger     *zap.Logger
	stats      *GTPUStats
}

// GTPUStats holds GTP-U statistics
type GTPUStats struct {
	UplinkPackets   uint64
	DownlinkPackets uint64
	UplinkBytes     uint64
	DownlinkBytes   uint64
	DroppedPackets  uint64
}

// GTPUHeader represents GTP-U header (simplified)
type GTPUHeader struct {
	Flags          uint8
	MessageType    uint8
	Length         uint16
	TEID           uint32
	SequenceNumber uint16
	NPDU           uint8
	NextExtHeader  uint8
}

// NewGTPUHandler creates a new GTP-U handler
func NewGTPUHandler(cfg *config.Config, upfCtx *upfcontext.UPFContext, logger *zap.Logger) *GTPUHandler {
	return &GTPUHandler{
		config:     cfg,
		upfContext: upfCtx,
		logger:     logger,
		stats:      &GTPUStats{},
	}
}

// Start starts the GTP-U handler
func (h *GTPUHandler) Start(ctx context.Context) error {
	// Start N3 listener (gNB -> UPF)
	if err := h.startN3Listener(ctx); err != nil {
		return err
	}

	// Start N6 listener (Data Network -> UPF)
	if err := h.startN6Listener(ctx); err != nil {
		return err
	}

	<-ctx.Done()
	return nil
}

// startN3Listener starts N3 interface listener
func (h *GTPUHandler) startN3Listener(ctx context.Context) error {
	addr, err := net.ResolveUDPAddr("udp", h.config.GetN3Address())
	if err != nil {
		return fmt.Errorf("failed to resolve N3 address: %w", err)
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen on N3: %w", err)
	}
	h.n3Conn = conn

	h.logger.Info("N3 (GTP-U) interface started", zap.String("address", h.config.GetN3Address()))

	go h.handleN3Traffic(ctx)
	return nil
}

// startN6Listener starts N6 interface listener
func (h *GTPUHandler) startN6Listener(ctx context.Context) error {
	// For development, use raw socket or TUN/TAP interface
	// Simplified: Listen on a UDP port for testing
	addr, err := net.ResolveUDPAddr("udp", "0.0.0.0:2153")
	if err != nil {
		return fmt.Errorf("failed to resolve N6 address: %w", err)
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen on N6: %w", err)
	}
	h.n6Conn = conn

	h.logger.Info("N6 (Data Network) interface started", zap.String("address", "0.0.0.0:2153"))

	go h.handleN6Traffic(ctx)
	return nil
}

// handleN3Traffic processes uplink traffic from gNB
func (h *GTPUHandler) handleN3Traffic(ctx context.Context) {
	buffer := make([]byte, h.config.Forwarding.BufferSize)

	for {
		select {
		case <-ctx.Done():
			return
		default:
			n, addr, err := h.n3Conn.ReadFromUDP(buffer)
			if err != nil {
				h.logger.Error("Failed to read from N3", zap.Error(err))
				continue
			}

			// Parse GTP-U header
			if n < 8 {
				h.logger.Warn("GTP-U packet too short", zap.Int("length", n))
				continue
			}

			header := h.parseGTPUHeader(buffer[:n])

			// Handle based on message type
			switch header.MessageType {
			case GTPU_ECHO_REQUEST:
				h.handleEchoRequest(addr)
			case GTPU_G_PDU:
				h.handleUplinkPacket(header, buffer[8:n], addr)
			default:
				h.logger.Debug("Unsupported GTP-U message type", zap.Uint8("type", header.MessageType))
			}
		}
	}
}

// handleN6Traffic processes downlink traffic from data network
func (h *GTPUHandler) handleN6Traffic(ctx context.Context) {
	buffer := make([]byte, h.config.Forwarding.BufferSize)

	for {
		select {
		case <-ctx.Done():
			return
		default:
			n, addr, err := h.n6Conn.ReadFromUDP(buffer)
			if err != nil {
				h.logger.Error("Failed to read from N6", zap.Error(err))
				continue
			}

			// Find session based on destination IP (UE IP)
			h.handleDownlinkPacket(buffer[:n], addr)
		}
	}
}

// parseGTPUHeader parses GTP-U header
func (h *GTPUHandler) parseGTPUHeader(data []byte) *GTPUHeader {
	header := &GTPUHeader{
		Flags:       data[0],
		MessageType: data[1],
		Length:      binary.BigEndian.Uint16(data[2:4]),
		TEID:        binary.BigEndian.Uint32(data[4:8]),
	}

	// Check for optional fields (S, PN, E flags)
	if (data[0] & 0x07) != 0 {
		if len(data) >= 12 {
			header.SequenceNumber = binary.BigEndian.Uint16(data[8:10])
			header.NPDU = data[10]
			header.NextExtHeader = data[11]
		}
	}

	return header
}

// handleUplinkPacket processes uplink data (N3 -> N6)
func (h *GTPUHandler) handleUplinkPacket(header *GTPUHeader, payload []byte, srcAddr *net.UDPAddr) {
	// Find session by TEID
	var session *upfcontext.UPFSession
	for _, s := range h.upfContext.GetAllSessions() {
		if s.UPFTEID == header.TEID {
			session = s
			break
		}
	}

	if session == nil {
		h.logger.Warn("No session found for TEID", zap.Uint32("teid", header.TEID))
		h.stats.DroppedPackets++
		return
	}

	// Update activity
	h.upfContext.UpdateActivity(session.SEID)

	// Extract IP packet from GTP-U payload
	ipPacket := payload

	// Apply QoS enforcement (simplified)
	if !h.applyQoS(session, ipPacket, true) {
		h.stats.DroppedPackets++
		return
	}

	// Forward to N6 (Data Network)
	h.forwardToN6(ipPacket, session)

	h.stats.UplinkPackets++
	h.stats.UplinkBytes += uint64(len(ipPacket))

	h.logger.Debug("Uplink packet forwarded",
		zap.Uint32("teid", header.TEID),
		zap.Int("size", len(ipPacket)),
		zap.String("ue_ip", session.UEAddress.String()))
}

// handleDownlinkPacket processes downlink data (N6 -> N3)
func (h *GTPUHandler) handleDownlinkPacket(ipPacket []byte, srcAddr *net.UDPAddr) {
	// Extract destination IP (UE IP) from IP header
	if len(ipPacket) < 20 {
		return
	}

	dstIP := net.IP(ipPacket[16:20])

	// Find session by UE IP
	var session *upfcontext.UPFSession
	for _, s := range h.upfContext.GetAllSessions() {
		if s.UEAddress.Equal(dstIP) {
			session = s
			break
		}
	}

	if session == nil {
		h.logger.Debug("No session found for UE IP", zap.String("ip", dstIP.String()))
		h.stats.DroppedPackets++
		return
	}

	// Apply QoS enforcement
	if !h.applyQoS(session, ipPacket, false) {
		h.stats.DroppedPackets++
		return
	}

	// Encapsulate in GTP-U and forward to gNB
	h.forwardToN3(ipPacket, session)

	h.stats.DownlinkPackets++
	h.stats.DownlinkBytes += uint64(len(ipPacket))

	h.logger.Debug("Downlink packet forwarded",
		zap.Uint32("gnb_teid", session.GNBTEID),
		zap.Int("size", len(ipPacket)),
		zap.String("ue_ip", session.UEAddress.String()))
}

// forwardToN6 forwards packet to data network
func (h *GTPUHandler) forwardToN6(ipPacket []byte, session *upfcontext.UPFSession) {
	// In development: forward to localhost or drop
	// In production: use TUN/TAP or raw sockets
	h.logger.Debug("Packet forwarded to N6", zap.Int("size", len(ipPacket)))
}

// forwardToN3 encapsulates and forwards packet to gNB
func (h *GTPUHandler) forwardToN3(ipPacket []byte, session *upfcontext.UPFSession) {
	// Build GTP-U header
	gtpuPacket := make([]byte, 8+len(ipPacket))

	// GTP-U header
	gtpuPacket[0] = 0x30 // Version 1, PT=1, no optional fields
	gtpuPacket[1] = GTPU_G_PDU
	binary.BigEndian.PutUint16(gtpuPacket[2:4], uint16(len(ipPacket)))
	binary.BigEndian.PutUint32(gtpuPacket[4:8], session.GNBTEID)

	// Copy IP packet
	copy(gtpuPacket[8:], ipPacket)

	// Send to gNB
	if session.GNBAddress != nil {
		gnbAddr := &net.UDPAddr{
			IP:   session.GNBAddress,
			Port: h.config.N3.Port,
		}

		_, err := h.n3Conn.WriteToUDP(gtpuPacket, gnbAddr)
		if err != nil {
			h.logger.Error("Failed to send to gNB", zap.Error(err))
		}
	}
}

// applyQoS applies QoS enforcement
func (h *GTPUHandler) applyQoS(session *upfcontext.UPFSession, packet []byte, uplink bool) bool {
	// Simplified QoS: check against MBR
	for _, qer := range session.QERs {
		if qer.GateStatus == 1 { // Closed
			return false
		}

		// Rate limiting would go here
		// For now, accept all packets
	}
	return true
}

// handleEchoRequest handles GTP-U echo request
func (h *GTPUHandler) handleEchoRequest(addr *net.UDPAddr) {
	response := make([]byte, 8)
	response[0] = 0x30
	response[1] = GTPU_ECHO_RESPONSE
	binary.BigEndian.PutUint16(response[2:4], 4)
	// Recovery IE would go here

	h.n3Conn.WriteToUDP(response, addr)
	h.logger.Debug("Sent GTP-U echo response", zap.String("to", addr.String()))
}

// GetStats returns GTP-U statistics
func (h *GTPUHandler) GetStats() *GTPUStats {
	return h.stats
}
