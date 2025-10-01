package server

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/your-org/5g-network/common/metrics"
	"github.com/your-org/5g-network/nf/smf/internal/service"
	"go.uber.org/zap"
)

// handleHealthCheck handles GET /health
func (s *SMFServer) handleHealthCheck(w http.ResponseWriter, r *http.Request) {
	s.respondJSON(w, http.StatusOK, map[string]string{
		"status": "healthy",
	})
}

// handleReadinessCheck handles GET /ready
func (s *SMFServer) handleReadinessCheck(w http.ResponseWriter, r *http.Request) {
	s.respondJSON(w, http.StatusOK, map[string]string{
		"status": "ready",
	})
}

// handleStatus handles GET /status
func (s *SMFServer) handleStatus(w http.ResponseWriter, r *http.Request) {
	stats := s.sessionService.GetSessionStatistics()

	s.respondJSON(w, http.StatusOK, map[string]interface{}{
		"service": "SMF",
		"version": "1.0.0",
		"name":    s.config.SMF.Name,
		"stats":   stats,
	})
}

// handleCreateSMContext handles POST /nsmf-pdusession/v1/sm-contexts
// TS 29.502, Clause 5.2.2.2.1
func (s *SMFServer) handleCreateSMContext(w http.ResponseWriter, r *http.Request) {
	var req service.CreateSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.respondError(w, http.StatusBadRequest, "invalid request body", err)
		return
	}

	resp, err := s.sessionService.CreateSession(&req)
	if err != nil {
		s.respondError(w, http.StatusInternalServerError, "failed to create session", err)
		metrics.RecordPDUSessionEstablishment("initial", "failed")
		return
	}

	if resp.Result != "SUCCESS" {
		s.respondError(w, http.StatusBadRequest, resp.Reason, nil)
		metrics.RecordPDUSessionEstablishment("initial", "failed")
		return
	}

	// Record successful PDU session establishment
	metrics.RecordPDUSessionEstablishment("initial", "success")
	stats := s.sessionService.GetSessionStatistics()
	if activeSessions, ok := stats["active_sessions"].(int); ok {
		metrics.SetActivePDUSessions(activeSessions)
	}
	if activeQoS, ok := stats["active_qos_flows"].(int); ok {
		metrics.SetActiveQoSFlows(activeQoS)
	}

	s.logger.Info("PDU session created via API",
		zap.String("supi", resp.SUPI),
		zap.Uint8("pdu_session_id", resp.PDUSessionID),
		zap.String("ue_ip", resp.UEIPv4Address),
	)

	s.respondJSON(w, http.StatusCreated, resp)
}

// handleUpdateSMContext handles PUT /nsmf-pdusession/v1/sm-contexts/{smContextRef}/modify
// TS 29.502, Clause 5.2.2.3.1
func (s *SMFServer) handleUpdateSMContext(w http.ResponseWriter, r *http.Request) {
	smContextRef := chi.URLParam(r, "smContextRef")

	var req service.UpdateSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.respondError(w, http.StatusBadRequest, "invalid request body", err)
		return
	}

	s.logger.Info("PDU session update requested",
		zap.String("sm_context_ref", smContextRef),
		zap.String("supi", req.SUPI),
	)

	// TODO: Implement session update logic
	s.respondJSON(w, http.StatusOK, map[string]interface{}{
		"result": "SUCCESS",
		"supi":   req.SUPI,
	})
}

// handleReleaseSMContext handles POST /nsmf-pdusession/v1/sm-contexts/{smContextRef}/release
// TS 29.502, Clause 5.2.2.4.1
func (s *SMFServer) handleReleaseSMContext(w http.ResponseWriter, r *http.Request) {
	smContextRef := chi.URLParam(r, "smContextRef")

	var req service.ReleaseSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.respondError(w, http.StatusBadRequest, "invalid request body", err)
		return
	}

	resp, err := s.sessionService.ReleaseSession(&req)
	if err != nil {
		s.respondError(w, http.StatusInternalServerError, "failed to release session", err)
		return
	}

	s.logger.Info("PDU session released via API",
		zap.String("sm_context_ref", smContextRef),
		zap.String("supi", resp.SUPI),
		zap.Uint8("pdu_session_id", resp.PDUSessionID),
	)

	s.respondJSON(w, http.StatusOK, resp)
}

// handleGetSMContext handles GET /nsmf-pdusession/v1/sm-contexts/{smContextRef}
func (s *SMFServer) handleGetSMContext(w http.ResponseWriter, r *http.Request) {
	smContextRef := chi.URLParam(r, "smContextRef")

	s.logger.Info("PDU session context retrieval requested",
		zap.String("sm_context_ref", smContextRef),
	)

	// TODO: Parse smContextRef to get SUPI and PDU Session ID
	// For now, return placeholder
	s.respondJSON(w, http.StatusOK, map[string]interface{}{
		"smContextRef": smContextRef,
		"status":       "active",
	})
}

// handleListSessions handles GET /admin/sessions
func (s *SMFServer) handleListSessions(w http.ResponseWriter, r *http.Request) {
	stats := s.sessionService.GetSessionStatistics()

	s.respondJSON(w, http.StatusOK, stats)
}

// handleGetSessionsBySUPI handles GET /admin/sessions/{supi}
func (s *SMFServer) handleGetSessionsBySUPI(w http.ResponseWriter, r *http.Request) {
	supi := chi.URLParam(r, "supi")

	s.logger.Info("Retrieving sessions for SUPI",
		zap.String("supi", supi),
	)

	// TODO: Implement session retrieval by SUPI
	s.respondJSON(w, http.StatusOK, map[string]interface{}{
		"supi":     supi,
		"sessions": []interface{}{},
	})
}

// handleGetStats handles GET /admin/stats
func (s *SMFServer) handleGetStats(w http.ResponseWriter, r *http.Request) {
	stats := s.sessionService.GetSessionStatistics()

	s.respondJSON(w, http.StatusOK, stats)
}

// respondJSON sends a JSON response
func (s *SMFServer) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		s.logger.Error("Failed to encode JSON response", zap.Error(err))
	}
}

// respondError sends an error response
func (s *SMFServer) respondError(w http.ResponseWriter, status int, message string, err error) {
	response := map[string]interface{}{
		"status": status,
		"title":  message,
	}

	if err != nil {
		response["detail"] = err.Error()
	}

	s.respondJSON(w, status, response)

	if err != nil {
		s.logger.Error(message, zap.Error(err), zap.Int("status", status))
	}
}
