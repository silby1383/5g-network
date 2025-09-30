package server

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/your-org/5g-network/nf/udm/internal/client"
	"github.com/your-org/5g-network/nf/udm/internal/service"
	"go.uber.org/zap"
)

// Authentication Service Handlers (Nudm_UEAuthentication)

func (s *UDMServer) handleGenerateAuthData(w http.ResponseWriter, r *http.Request) {
	supi := chi.URLParam(r, "supi")

	var authInfo service.AuthenticationInfo
	if err := json.NewDecoder(r.Body).Decode(&authInfo); err != nil {
		s.respondError(w, http.StatusBadRequest, "invalid request body", err)
		return
	}

	authInfo.SUPI = supi

	result, err := s.authService.GenerateAuthData(r.Context(), &authInfo)
	if err != nil {
		s.respondError(w, http.StatusInternalServerError, "failed to generate auth data", err)
		return
	}

	s.logger.Info("Generated authentication data", zap.String("supi", supi))
	s.respondJSON(w, http.StatusOK, result)
}

func (s *UDMServer) handleConfirmAuth(w http.ResponseWriter, r *http.Request) {
	supi := chi.URLParam(r, "supi")

	var authEvent map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&authEvent); err != nil {
		s.respondError(w, http.StatusBadRequest, "invalid request body", err)
		return
	}

	if err := s.authService.ConfirmAuth(r.Context(), supi, authEvent); err != nil {
		s.respondError(w, http.StatusInternalServerError, "failed to confirm auth", err)
		return
	}

	s.respondJSON(w, http.StatusCreated, map[string]string{
		"status": "confirmed",
	})
}

// Subscriber Data Management Handlers (Nudm_SDM)

func (s *UDMServer) handleGetAMData(w http.ResponseWriter, r *http.Request) {
	supi := chi.URLParam(r, "supi")
	plmnIDStr := r.URL.Query().Get("plmn-id")

	var plmnID *client.PLMNID
	if plmnIDStr != "" {
		// Parse PLMN ID (format: mcc-mnc)
		// For simplicity, using default PLMN
		plmnID = &client.PLMNID{
			MCC: s.config.PLMN.MCC,
			MNC: s.config.PLMN.MNC,
		}
	}

	amData, err := s.sdmService.GetAMData(r.Context(), supi, plmnID)
	if err != nil {
		s.respondError(w, http.StatusNotFound, "failed to get AM data", err)
		return
	}

	s.logger.Debug("Retrieved AM data", zap.String("supi", supi))
	s.respondJSON(w, http.StatusOK, amData)
}

func (s *UDMServer) handleGetSMData(w http.ResponseWriter, r *http.Request) {
	supi := chi.URLParam(r, "supi")
	dnn := r.URL.Query().Get("dnn")

	plmnID := &client.PLMNID{
		MCC: s.config.PLMN.MCC,
		MNC: s.config.PLMN.MNC,
	}

	smData, err := s.sdmService.GetSMData(r.Context(), supi, plmnID, dnn)
	if err != nil {
		s.respondError(w, http.StatusNotFound, "failed to get SM data", err)
		return
	}

	s.logger.Debug("Retrieved SM data", zap.String("supi", supi), zap.String("dnn", dnn))
	s.respondJSON(w, http.StatusOK, smData)
}

func (s *UDMServer) handleGetSMDataWithPlmn(w http.ResponseWriter, r *http.Request) {
	supi := chi.URLParam(r, "supi")
	servingPlmnID := chi.URLParam(r, "servingPlmnId")
	dnn := r.URL.Query().Get("dnn")

	s.logger.Debug("Getting SM data with PLMN",
		zap.String("supi", supi),
		zap.String("serving_plmn_id", servingPlmnID),
		zap.String("dnn", dnn),
	)

	// Parse serving PLMN ID
	plmnID := &client.PLMNID{
		MCC: s.config.PLMN.MCC,
		MNC: s.config.PLMN.MNC,
	}

	smData, err := s.sdmService.GetSMData(r.Context(), supi, plmnID, dnn)
	if err != nil {
		s.respondError(w, http.StatusNotFound, "failed to get SM data", err)
		return
	}

	s.respondJSON(w, http.StatusOK, smData)
}

func (s *UDMServer) handleSubscribeSDM(w http.ResponseWriter, r *http.Request) {
	supi := chi.URLParam(r, "supi")

	var subscription struct {
		CallbackReference     string   `json:"callbackReference"`
		MonitoredResourceUris []string `json:"monitoredResourceUris"`
	}

	if err := json.NewDecoder(r.Body).Decode(&subscription); err != nil {
		s.respondError(w, http.StatusBadRequest, "invalid request body", err)
		return
	}

	subscriptionID, err := s.sdmService.SubscribeToDataChanges(r.Context(), supi, subscription.CallbackReference)
	if err != nil {
		s.respondError(w, http.StatusInternalServerError, "failed to create subscription", err)
		return
	}

	s.respondJSON(w, http.StatusCreated, map[string]string{
		"subscriptionId":    subscriptionID,
		"callbackReference": subscription.CallbackReference,
	})
}

func (s *UDMServer) handleUnsubscribeSDM(w http.ResponseWriter, r *http.Request) {
	supi := chi.URLParam(r, "supi")
	subscriptionID := chi.URLParam(r, "subscriptionId")

	if err := s.sdmService.UnsubscribeFromDataChanges(r.Context(), subscriptionID); err != nil {
		s.respondError(w, http.StatusNotFound, "failed to delete subscription", err)
		return
	}

	s.logger.Info("SDM subscription deleted",
		zap.String("supi", supi),
		zap.String("subscription_id", subscriptionID),
	)

	w.WriteHeader(http.StatusNoContent)
}

// UE Context Management Handlers (Nudm_UECM)

func (s *UDMServer) handleRegisterAMF3GPP(w http.ResponseWriter, r *http.Request) {
	supi := chi.URLParam(r, "supi")

	var registration service.AMF3GPPAccessRegistration
	if err := json.NewDecoder(r.Body).Decode(&registration); err != nil {
		s.respondError(w, http.StatusBadRequest, "invalid request body", err)
		return
	}

	if err := s.uecmService.RegisterAMF3GPPAccess(r.Context(), supi, &registration); err != nil {
		s.respondError(w, http.StatusInternalServerError, "failed to register AMF", err)
		return
	}

	s.logger.Info("AMF registered",
		zap.String("supi", supi),
		zap.String("amf_instance_id", registration.AMFInstanceID),
	)

	s.respondJSON(w, http.StatusCreated, &registration)
}

func (s *UDMServer) handleUpdateAMF3GPP(w http.ResponseWriter, r *http.Request) {
	supi := chi.URLParam(r, "supi")

	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		s.respondError(w, http.StatusBadRequest, "invalid request body", err)
		return
	}

	if err := s.uecmService.UpdateAMF3GPPAccess(r.Context(), supi, updates); err != nil {
		s.respondError(w, http.StatusNotFound, "failed to update AMF registration", err)
		return
	}

	s.logger.Info("AMF registration updated", zap.String("supi", supi))
	w.WriteHeader(http.StatusNoContent)
}

func (s *UDMServer) handleGetAMF3GPP(w http.ResponseWriter, r *http.Request) {
	supi := chi.URLParam(r, "supi")

	registration, err := s.uecmService.Get3GPPRegistration(r.Context(), supi)
	if err != nil {
		s.respondError(w, http.StatusNotFound, "AMF registration not found", err)
		return
	}

	s.logger.Debug("Retrieved AMF registration", zap.String("supi", supi))
	s.respondJSON(w, http.StatusOK, registration)
}

func (s *UDMServer) handleDeregisterAMF3GPP(w http.ResponseWriter, r *http.Request) {
	supi := chi.URLParam(r, "supi")

	if err := s.uecmService.DeregisterAMF3GPPAccess(r.Context(), supi); err != nil {
		s.respondError(w, http.StatusNotFound, "failed to deregister AMF", err)
		return
	}

	s.logger.Info("AMF deregistered", zap.String("supi", supi))
	w.WriteHeader(http.StatusNoContent)
}

func (s *UDMServer) handleGetUEContext(w http.ResponseWriter, r *http.Request) {
	supi := chi.URLParam(r, "supi")

	ueContext, err := s.uecmService.GetUEContext(r.Context(), supi)
	if err != nil {
		s.respondError(w, http.StatusNotFound, "UE context not found", err)
		return
	}

	s.logger.Debug("Retrieved UE context", zap.String("supi", supi))
	s.respondJSON(w, http.StatusOK, ueContext)
}

// Admin Handlers

func (s *UDMServer) handleGetStats(w http.ResponseWriter, r *http.Request) {
	stats := s.uecmService.GetStats()

	s.respondJSON(w, http.StatusOK, map[string]interface{}{
		"service":    "UDM",
		"version":    "1.0.0",
		"uecm_stats": stats,
	})
}
