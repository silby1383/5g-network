package simulated

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/your-org/5g-network/common/dataplane"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// SimulatedDataPlane implements a simulated UPF data plane in Go
type SimulatedDataPlane struct {
	config   *dataplane.Config
	sessions map[uint64]*SessionRules
	stats    *dataplane.Stats
	logger   *zap.Logger
	tracer   trace.Tracer
	mu       sync.RWMutex

	// Processing workers
	workers    int
	packetChan chan *dataplane.Packet
	stopChan   chan struct{}
}

// SessionRules holds all rules for a PFCP session
type SessionRules struct {
	SessionID uint64
	PDRs      map[uint16]*dataplane.PDR
	FARs      map[uint16]*dataplane.FAR
	QERs      map[uint16]*dataplane.QER
	URRs      map[uint32]*dataplane.URR

	// Statistics
	PacketsProcessed uint64
	BytesProcessed   uint64
	CreatedAt        time.Time
}

// NewSimulatedDataPlane creates a new simulated data plane
func NewSimulatedDataPlane(logger *zap.Logger) *SimulatedDataPlane {
	return &SimulatedDataPlane{
		sessions: make(map[uint64]*SessionRules),
		stats: &dataplane.Stats{
			Errors:    make(map[string]uint64),
			Timestamp: time.Now(),
		},
		logger:     logger,
		tracer:     otel.Tracer("upf-dataplane"),
		packetChan: make(chan *dataplane.Packet, 10000),
		stopChan:   make(chan struct{}),
	}
}

// Initialize the data plane
func (s *SimulatedDataPlane) Initialize(ctx context.Context, config *dataplane.Config) error {
	ctx, span := s.tracer.Start(ctx, "SimulatedDataPlane.Initialize")
	defer span.End()

	s.config = config
	s.workers = config.Workers
	if s.workers == 0 {
		s.workers = 4
	}

	// Start packet processing workers
	for i := 0; i < s.workers; i++ {
		go s.packetWorker(i)
	}

	s.logger.Info("Simulated data plane initialized",
		zap.Int("workers", s.workers),
		zap.String("n3_interface", config.N3Interface),
		zap.String("n6_interface", config.N6Interface),
	)

	span.SetAttributes(
		attribute.Int("workers", s.workers),
		attribute.String("type", "simulated"),
	)

	return nil
}

// InstallPDR installs a Packet Detection Rule
func (s *SimulatedDataPlane) InstallPDR(ctx context.Context, sessionID uint64, pdr *dataplane.PDR) error {
	ctx, span := s.tracer.Start(ctx, "SimulatedDataPlane.InstallPDR")
	defer span.End()

	s.mu.Lock()
	defer s.mu.Unlock()

	// Get or create session
	session, exists := s.sessions[sessionID]
	if !exists {
		session = &SessionRules{
			SessionID: sessionID,
			PDRs:      make(map[uint16]*dataplane.PDR),
			FARs:      make(map[uint16]*dataplane.FAR),
			QERs:      make(map[uint16]*dataplane.QER),
			URRs:      make(map[uint32]*dataplane.URR),
			CreatedAt: time.Now(),
		}
		s.sessions[sessionID] = session
		s.stats.TotalSessions++
		s.stats.ActiveSessions++
	}

	// Install PDR
	session.PDRs[pdr.PDRID] = pdr

	s.logger.Debug("PDR installed",
		zap.Uint64("session_id", sessionID),
		zap.Uint16("pdr_id", pdr.PDRID),
		zap.Uint32("precedence", pdr.Precedence),
	)

	span.SetAttributes(
		attribute.Int64("session_id", int64(sessionID)),
		attribute.Int("pdr_id", int(pdr.PDRID)),
		attribute.Int("precedence", int(pdr.Precedence)),
	)

	return nil
}

// InstallFAR installs a Forwarding Action Rule
func (s *SimulatedDataPlane) InstallFAR(ctx context.Context, sessionID uint64, far *dataplane.FAR) error {
	ctx, span := s.tracer.Start(ctx, "SimulatedDataPlane.InstallFAR")
	defer span.End()

	s.mu.Lock()
	defer s.mu.Unlock()

	session, exists := s.sessions[sessionID]
	if !exists {
		return fmt.Errorf("session %d not found", sessionID)
	}

	session.FARs[far.FARID] = far

	s.logger.Debug("FAR installed",
		zap.Uint64("session_id", sessionID),
		zap.Uint16("far_id", far.FARID),
		zap.Uint8("apply_action", far.ApplyAction),
	)

	span.SetAttributes(
		attribute.Int64("session_id", int64(sessionID)),
		attribute.Int("far_id", int(far.FARID)),
	)

	return nil
}

// InstallQER installs a QoS Enforcement Rule
func (s *SimulatedDataPlane) InstallQER(ctx context.Context, sessionID uint64, qer *dataplane.QER) error {
	ctx, span := s.tracer.Start(ctx, "SimulatedDataPlane.InstallQER")
	defer span.End()

	s.mu.Lock()
	defer s.mu.Unlock()

	session, exists := s.sessions[sessionID]
	if !exists {
		return fmt.Errorf("session %d not found", sessionID)
	}

	session.QERs[qer.QERID] = qer

	s.logger.Debug("QER installed",
		zap.Uint64("session_id", sessionID),
		zap.Uint16("qer_id", qer.QERID),
		zap.Uint8("qfi", qer.QFI),
	)

	return nil
}

// InstallURR installs a Usage Reporting Rule
func (s *SimulatedDataPlane) InstallURR(ctx context.Context, sessionID uint64, urr *dataplane.URR) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	session, exists := s.sessions[sessionID]
	if !exists {
		return fmt.Errorf("session %d not found", sessionID)
	}

	session.URRs[urr.URRID] = urr
	return nil
}

// RemovePDR removes a PDR
func (s *SimulatedDataPlane) RemovePDR(ctx context.Context, sessionID uint64, pdrID uint16) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if session, exists := s.sessions[sessionID]; exists {
		delete(session.PDRs, pdrID)
	}
	return nil
}

// RemoveFAR removes a FAR
func (s *SimulatedDataPlane) RemoveFAR(ctx context.Context, sessionID uint64, farID uint16) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if session, exists := s.sessions[sessionID]; exists {
		delete(session.FARs, farID)
	}
	return nil
}

// RemoveQER removes a QER
func (s *SimulatedDataPlane) RemoveQER(ctx context.Context, sessionID uint64, qerID uint16) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if session, exists := s.sessions[sessionID]; exists {
		delete(session.QERs, qerID)
	}
	return nil
}

// RemoveURR removes a URR
func (s *SimulatedDataPlane) RemoveURR(ctx context.Context, sessionID uint64, urrID uint32) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if session, exists := s.sessions[sessionID]; exists {
		delete(session.URRs, urrID)
	}
	return nil
}

// RemoveSession removes an entire session
func (s *SimulatedDataPlane) RemoveSession(ctx context.Context, sessionID uint64) error {
	ctx, span := s.tracer.Start(ctx, "SimulatedDataPlane.RemoveSession")
	defer span.End()

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.sessions[sessionID]; exists {
		delete(s.sessions, sessionID)
		s.stats.ActiveSessions--

		s.logger.Info("Session removed",
			zap.Uint64("session_id", sessionID),
		)
	}

	return nil
}

// ProcessPacket simulates packet processing
func (s *SimulatedDataPlane) ProcessPacket(ctx context.Context, packet *dataplane.Packet) error {
	select {
	case s.packetChan <- packet:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	default:
		s.incrementError(dataplane.ErrQueueFull)
		return fmt.Errorf("packet queue full")
	}
}

// packetWorker processes packets from the queue
func (s *SimulatedDataPlane) packetWorker(id int) {
	s.logger.Info("Packet worker started", zap.Int("worker_id", id))

	for {
		select {
		case packet := <-s.packetChan:
			s.processPacketInternal(packet)
		case <-s.stopChan:
			s.logger.Info("Packet worker stopped", zap.Int("worker_id", id))
			return
		}
	}
}

// processPacketInternal handles the actual packet processing logic
func (s *SimulatedDataPlane) processPacketInternal(packet *dataplane.Packet) {
	ctx, span := s.tracer.Start(context.Background(), "SimulatedDataPlane.processPacket")
	defer span.End()

	s.mu.RLock()
	defer s.mu.RUnlock()

	// Match packet against all PDRs to find matching session
	var matchedSession *SessionRules
	var matchedPDR *dataplane.PDR
	var matchedFAR *dataplane.FAR

	for _, session := range s.sessions {
		for _, pdr := range session.PDRs {
			if s.matchPDR(pdr, packet) {
				matchedSession = session
				matchedPDR = pdr
				// Get associated FAR
				if far, exists := session.FARs[pdr.FARID]; exists {
					matchedFAR = far
				}
				break
			}
		}
		if matchedSession != nil {
			break
		}
	}

	if matchedSession == nil {
		// No matching PDR found - drop packet
		s.stats.PacketsDropped++
		s.incrementError(dataplane.ErrSessionNotFound)
		span.SetAttributes(attribute.String("action", "drop"))
		return
	}

	// Update statistics
	s.stats.PacketsProcessed++
	s.stats.BytesProcessed += uint64(len(packet.Data))
	matchedSession.PacketsProcessed++
	matchedSession.BytesProcessed += uint64(len(packet.Data))

	// Apply FAR action
	if matchedFAR != nil {
		s.applyFAR(ctx, matchedFAR, packet, matchedPDR, span)
	}

	// Apply QER if present
	for _, qerID := range matchedPDR.QERID {
		if qer, exists := matchedSession.QERs[qerID]; exists {
			s.applyQER(qer, packet)
		}
	}

	span.SetAttributes(
		attribute.Int64("session_id", int64(matchedSession.SessionID)),
		attribute.Int("pdr_id", int(matchedPDR.PDRID)),
		attribute.Int("packet_size", len(packet.Data)),
	)
}

// matchPDR checks if a packet matches a PDR
func (s *SimulatedDataPlane) matchPDR(pdr *dataplane.PDR, packet *dataplane.Packet) bool {
	if pdr.PDI == nil {
		return false
	}

	// Match on interface
	if pdr.PDI.SourceInterface != "" {
		if (pdr.PDI.SourceInterface == "ACCESS" && packet.Interface != "N3") ||
			(pdr.PDI.SourceInterface == "CORE" && packet.Interface != "N6") {
			return false
		}
	}

	// Match on UE IP
	if pdr.PDI.UEIPAddress != nil {
		if pdr.PDI.UEIPAddress.IPv4 != nil && !pdr.PDI.UEIPAddress.IPv4.Equal(packet.DstIP) {
			return false
		}
	}

	// Match on F-TEID (for GTP-U packets)
	if pdr.PDI.LocalFTEID != nil {
		if pdr.PDI.LocalFTEID.TEID != packet.TEID {
			return false
		}
	}

	// Match on QFI
	if pdr.PDI.QFI != 0 {
		// Would need to extract QFI from packet headers
	}

	return true
}

// applyFAR applies forwarding actions
func (s *SimulatedDataPlane) applyFAR(ctx context.Context, far *dataplane.FAR, packet *dataplane.Packet, pdr *dataplane.PDR, span trace.Span) {
	// DROP action
	if far.ApplyAction&0x01 != 0 {
		s.stats.PacketsDropped++
		span.SetAttributes(attribute.String("action", "drop"))
		return
	}

	// FORWARD action
	if far.ApplyAction&0x02 != 0 {
		s.stats.PacketsForwarded++

		// Simulate GTP-U encapsulation/decapsulation
		if far.ForwardingParameters != nil {
			if far.ForwardingParameters.OuterHeaderCreation != nil {
				// Encapsulate in GTP-U
				s.logger.Debug("Simulating GTP-U encapsulation",
					zap.Uint32("teid", far.ForwardingParameters.OuterHeaderCreation.TEID),
					zap.String("dst_ip", far.ForwardingParameters.OuterHeaderCreation.IPv4.String()),
				)
			}

			if pdr.OuterHeaderRemoval != nil {
				// Decapsulate GTP-U
				s.logger.Debug("Simulating GTP-U decapsulation",
					zap.Uint32("teid", packet.TEID),
				)
			}
		}

		span.SetAttributes(
			attribute.String("action", "forward"),
			attribute.String("destination", far.ForwardingParameters.DestinationInterface),
		)
		return
	}

	// BUFFER action
	if far.ApplyAction&0x04 != 0 {
		s.stats.PacketsBuffered++
		span.SetAttributes(attribute.String("action", "buffer"))
		return
	}
}

// applyQER applies QoS enforcement
func (s *SimulatedDataPlane) applyQER(qer *dataplane.QER, packet *dataplane.Packet) {
	// Check gate status
	if qer.GateStatus == 1 { // CLOSED
		// Would drop packet in real implementation
		return
	}

	// Simulate rate limiting (simplified)
	if qer.MBR != nil {
		// Would enforce maximum bit rate here
		s.logger.Debug("Simulating MBR enforcement",
			zap.Uint64("mbr_uplink", qer.MBR.Uplink),
			zap.Uint64("mbr_downlink", qer.MBR.Downlink),
		)
	}

	if qer.GBR != nil {
		// Would enforce guaranteed bit rate here
		s.logger.Debug("Simulating GBR enforcement",
			zap.Uint64("gbr_uplink", qer.GBR.Uplink),
			zap.Uint64("gbr_downlink", qer.GBR.Downlink),
		)
	}
}

// GetStats returns current statistics
func (s *SimulatedDataPlane) GetStats(ctx context.Context) (*dataplane.Stats, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Create a copy of stats
	stats := *s.stats
	stats.Timestamp = time.Now()
	stats.ActiveTunnels = stats.ActiveSessions // Simplified

	return &stats, nil
}

// Shutdown stops the data plane
func (s *SimulatedDataPlane) Shutdown(ctx context.Context) error {
	ctx, span := s.tracer.Start(ctx, "SimulatedDataPlane.Shutdown")
	defer span.End()

	s.logger.Info("Shutting down simulated data plane")

	// Stop all workers
	close(s.stopChan)

	// Wait for workers to finish
	time.Sleep(100 * time.Millisecond)

	s.logger.Info("Simulated data plane stopped")
	return nil
}

// incrementError safely increments error counter
func (s *SimulatedDataPlane) incrementError(errType string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.stats.Errors[errType]++
}
