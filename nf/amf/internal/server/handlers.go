package server

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/your-org/5g-network/nf/amf/internal/service"
	"go.uber.org/zap"
)

// handleAuthenticationRequest handles POST request to initiate UE authentication
func (s *AMFServer) handleAuthenticationRequest(w http.ResponseWriter, r *http.Request) {
	var req service.AuthenticationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.respondError(w, http.StatusBadRequest, "invalid request body", err)
		return
	}

	s.logger.Info("Received authentication request",
		zap.String("supi", req.SUPI),
	)

	response, err := s.registrationService.InitiateAuthentication(r.Context(), &req)
	if err != nil {
		s.respondError(w, http.StatusInternalServerError, "failed to initiate authentication", err)
		return
	}

	s.logger.Info("Authentication initiated",
		zap.String("supi", req.SUPI),
		zap.String("auth_ctx_id", response.AuthCtxID),
	)

	s.respondJSON(w, http.StatusCreated, response)
}

// handleAuthenticationConfirm handles PUT request to confirm UE authentication
func (s *AMFServer) handleAuthenticationConfirm(w http.ResponseWriter, r *http.Request) {
	authCtxID := chi.URLParam(r, "authCtxId")

	var req service.AuthenticationConfirmRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.respondError(w, http.StatusBadRequest, "invalid request body", err)
		return
	}
	req.AuthCtxID = authCtxID

	s.logger.Info("Received authentication confirmation",
		zap.String("auth_ctx_id", authCtxID),
	)

	response, err := s.registrationService.ConfirmAuthentication(r.Context(), &req)
	if err != nil {
		s.respondError(w, http.StatusInternalServerError, "failed to confirm authentication", err)
		return
	}

	s.logger.Info("Authentication confirmed",
		zap.String("auth_ctx_id", authCtxID),
		zap.String("result", response.Result),
	)

	s.respondJSON(w, http.StatusOK, response)
}

// handleRegistrationRequest handles POST request for UE registration
func (s *AMFServer) handleRegistrationRequest(w http.ResponseWriter, r *http.Request) {
	var req service.RegistrationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.respondError(w, http.StatusBadRequest, "invalid request body", err)
		return
	}

	s.logger.Info("Received registration request",
		zap.String("supi", req.SUPI),
		zap.String("type", req.RegistrationType),
	)

	response, err := s.registrationService.RegisterUE(r.Context(), &req)
	if err != nil {
		s.respondError(w, http.StatusInternalServerError, "failed to register UE", err)
		return
	}

	if response.Result != "SUCCESS" {
		s.logger.Warn("Registration failed",
			zap.String("supi", req.SUPI),
			zap.String("reason", response.Reason),
		)
		s.respondJSON(w, http.StatusForbidden, response)
		return
	}

	s.logger.Info("UE registered successfully",
		zap.String("supi", req.SUPI),
		zap.String("guami", response.GUAMI),
	)

	s.respondJSON(w, http.StatusCreated, response)
}

// handleDeregistration handles DELETE request for UE deregistration
func (s *AMFServer) handleDeregistration(w http.ResponseWriter, r *http.Request) {
	supi := chi.URLParam(r, "supi")

	s.logger.Info("Received deregistration request",
		zap.String("supi", supi),
	)

	err := s.registrationService.DeregisterUE(r.Context(), supi)
	if err != nil {
		s.respondError(w, http.StatusNotFound, "failed to deregister UE", err)
		return
	}

	s.logger.Info("UE deregistered",
		zap.String("supi", supi),
	)

	w.WriteHeader(http.StatusNoContent)
}

// handleGetUEContext handles GET request for UE context
func (s *AMFServer) handleGetUEContext(w http.ResponseWriter, r *http.Request) {
	ueContextID := chi.URLParam(r, "ueContextId")

	// For simplicity, ueContextId == SUPI
	ueCtx, exists := s.contextManager.GetContext(ueContextID)
	if !exists {
		s.respondError(w, http.StatusNotFound, "UE context not found", nil)
		return
	}

	s.respondJSON(w, http.StatusOK, map[string]interface{}{
		"supi":              ueCtx.SUPI,
		"registrationState": ueCtx.RegistrationState,
		"connectionState":   ueCtx.ConnectionState,
		"guami":             ueCtx.GUAMI,
		"tai":               ueCtx.TAI,
		"allowedNssai":      ueCtx.AllowedNSSAI,
	})
}

// handleReleaseUEContext handles POST request to release UE context
func (s *AMFServer) handleReleaseUEContext(w http.ResponseWriter, r *http.Request) {
	ueContextID := chi.URLParam(r, "ueContextId")

	s.logger.Info("Releasing UE context",
		zap.String("ue_context_id", ueContextID),
	)

	// For simplicity, ueContextId == SUPI
	err := s.registrationService.DeregisterUE(r.Context(), ueContextID)
	if err != nil {
		s.respondError(w, http.StatusNotFound, "failed to release UE context", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// handleN1N2Transfer handles POST request for N1/N2 message transfer
func (s *AMFServer) handleN1N2Transfer(w http.ResponseWriter, r *http.Request) {
	ueContextID := chi.URLParam(r, "ueContextId")

	s.logger.Info("N1/N2 message transfer",
		zap.String("ue_context_id", ueContextID),
	)

	// Simplified - just acknowledge
	s.respondJSON(w, http.StatusOK, map[string]string{
		"status": "accepted",
	})
}

// handleListUEContexts handles GET request for listing all UE contexts
func (s *AMFServer) handleListUEContexts(w http.ResponseWriter, r *http.Request) {
	contexts := s.contextManager.GetAllContexts()

	ueList := make([]map[string]interface{}, 0, len(contexts))
	for _, ctx := range contexts {
		ueList = append(ueList, map[string]interface{}{
			"supi":              ctx.SUPI,
			"registrationState": ctx.RegistrationState,
			"connectionState":   ctx.ConnectionState,
			"guami":             ctx.GUAMI,
			"registeredAt":      ctx.RegisteredAt,
			"lastActivityAt":    ctx.LastActivityAt,
		})
	}

	s.respondJSON(w, http.StatusOK, map[string]interface{}{
		"total": len(ueList),
		"ues":   ueList,
	})
}

// handleGetStats handles GET request for statistics
func (s *AMFServer) handleGetStats(w http.ResponseWriter, r *http.Request) {
	stats := s.registrationService.GetRegistrationStats()
	
	s.respondJSON(w, http.StatusOK, map[string]interface{}{
		"service":          "AMF",
		"version":          "1.0.0",
		"guami":            s.config.GetGUAMI(),
		"plmn":             map[string]string{
			"mcc": s.config.PLMN.MCC,
			"mnc": s.config.PLMN.MNC,
			"tac": s.config.PLMN.TAC,
		},
		"registration_stats": stats,
	})
}
