package server

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/your-org/5g-network/nf/ausf/internal/service"
	"go.uber.org/zap"
)

// handleUEAuthenticationRequest handles POST request to initiate UE authentication
// TS 29.509, Clause 5.2.2.2.2
func (s *AUSFServer) handleUEAuthenticationRequest(w http.ResponseWriter, r *http.Request) {
	var req service.UEAuthenticationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.respondError(w, http.StatusBadRequest, "invalid request body", err)
		return
	}

	s.logger.Info("Received UE authentication request",
		zap.String("supi", req.SUPI),
		zap.String("serving_network", req.ServingNetworkName),
	)

	response, err := s.authService.UEAuthenticationCtx(r.Context(), &req)
	if err != nil {
		s.respondError(w, http.StatusInternalServerError, "failed to initiate authentication", err)
		return
	}

	s.logger.Info("UE authentication initiated",
		zap.String("supi", req.SUPI),
		zap.String("auth_type", response.AuthType),
	)

	s.respondJSON(w, http.StatusCreated, response)
}

// handleConfirm5gAkaAuth handles PUT request to confirm 5G-AKA authentication
// TS 29.509, Clause 5.2.2.2.3
func (s *AUSFServer) handleConfirm5gAkaAuth(w http.ResponseWriter, r *http.Request) {
	authCtxID := chi.URLParam(r, "authCtxId")

	var confirmData service.ConfirmationData
	if err := json.NewDecoder(r.Body).Decode(&confirmData); err != nil {
		s.respondError(w, http.StatusBadRequest, "invalid request body", err)
		return
	}

	s.logger.Info("Received 5G-AKA confirmation",
		zap.String("auth_ctx_id", authCtxID),
	)

	response, err := s.authService.Confirm5gAkaAuth(r.Context(), authCtxID, &confirmData)
	if err != nil {
		s.respondError(w, http.StatusNotFound, "failed to confirm authentication", err)
		return
	}

	s.logger.Info("5G-AKA authentication confirmed",
		zap.String("auth_ctx_id", authCtxID),
		zap.String("result", response.AuthResult),
	)

	s.respondJSON(w, http.StatusOK, response)
}

// handleGetStats handles GET request for statistics
func (s *AUSFServer) handleGetStats(w http.ResponseWriter, r *http.Request) {
	stats := s.authService.GetStats()
	
	s.respondJSON(w, http.StatusOK, map[string]interface{}{
		"service":    "AUSF",
		"version":    "1.0.0",
		"auth_stats": stats,
	})
}

// handleGetAuthContext handles GET request for auth context (TEST ONLY!)
// This is for testing without a real UE - NOT FOR PRODUCTION
func (s *AUSFServer) handleGetAuthContext(w http.ResponseWriter, r *http.Request) {
	authCtxID := chi.URLParam(r, "authCtxId")

	authCtx, err := s.authService.GetAuthContext(authCtxID)
	if err != nil {
		s.respondError(w, http.StatusNotFound, "authentication context not found", err)
		return
	}

	// Return context including HXRES* for testing
	s.respondJSON(w, http.StatusOK, map[string]interface{}{
		"authCtxId":          authCtx.AuthCtxID,
		"supi":               authCtx.SUPI,
		"authType":           authCtx.AuthType,
		"rand":               authCtx.RAND,
		"autn":               authCtx.AUTN,
		"hxres":              authCtx.HXRES, // For testing!
		"servingNetworkName": authCtx.ServingNetworkName,
	})
}
