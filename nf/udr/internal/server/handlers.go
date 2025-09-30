package server

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/your-org/5g-network/nf/udr/internal/repository"
	"go.uber.org/zap"
)

// handleGetAMData handles GET request for Access and Mobility subscription data
// TS 29.505, Clause 5.2.3.2
func (s *UDRServer) handleGetAMData(w http.ResponseWriter, r *http.Request) {
	supi := chi.URLParam(r, "supi")

	subscriber, err := s.repository.GetSubscriber(r.Context(), supi)
	if err != nil {
		s.respondError(w, http.StatusNotFound, "subscriber not found", err)
		return
	}

	// Extract AM data
	amData := map[string]interface{}{
		"supi": subscriber.SUPI,
		"subscribedUeAmbr": map[string]interface{}{
			"uplink":   subscriber.SubscribedUeAmbrUplink,
			"downlink": subscriber.SubscribedUeAmbrDownlink,
		},
		"nssai":            subscriber.NSSAI,
		"subscriberStatus": subscriber.SubscriberStatus,
		"roamingAllowed":   subscriber.RoamingAllowed,
	}

	s.respondJSON(w, http.StatusOK, amData)
}

// handleUpdateAMData handles PUT request to update AM data
func (s *UDRServer) handleUpdateAMData(w http.ResponseWriter, r *http.Request) {
	supi := chi.URLParam(r, "supi")

	var data repository.SubscriberData
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		s.respondError(w, http.StatusBadRequest, "invalid request body", err)
		return
	}

	data.SUPI = supi
	err := s.repository.UpdateSubscriber(r.Context(), supi, &data)
	if err != nil {
		s.respondError(w, http.StatusInternalServerError, "failed to update subscriber", err)
		return
	}

	s.respondJSON(w, http.StatusOK, &data)
}

// handleGetSMData handles GET request for Session Management subscription data
// TS 29.505, Clause 5.2.3.3
func (s *UDRServer) handleGetSMData(w http.ResponseWriter, r *http.Request) {
	supi := chi.URLParam(r, "supi")
	dnn := r.URL.Query().Get("dnn")

	if dnn == "" {
		// Return all SM data for subscriber
		subscriber, err := s.repository.GetSubscriber(r.Context(), supi)
		if err != nil {
			s.respondError(w, http.StatusNotFound, "subscriber not found", err)
			return
		}

		s.respondJSON(w, http.StatusOK, subscriber.DNNConfigurations)
		return
	}

	// Return specific DNN data
	smData, err := s.repository.GetSMSubscription(r.Context(), supi, dnn)
	if err != nil {
		s.respondError(w, http.StatusNotFound, "SM data not found", err)
		return
	}

	s.respondJSON(w, http.StatusOK, smData)
}

// handleUpdateSMData handles PUT request to update SM data
func (s *UDRServer) handleUpdateSMData(w http.ResponseWriter, r *http.Request) {
	supi := chi.URLParam(r, "supi")
	dnn := r.URL.Query().Get("dnn")

	var data repository.SessionManagementSubscriptionData
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		s.respondError(w, http.StatusBadRequest, "invalid request body", err)
		return
	}

	data.SUPI = supi
	data.DNN = dnn

	err := s.repository.UpdateSMSubscription(r.Context(), supi, dnn, &data)
	if err != nil {
		s.respondError(w, http.StatusInternalServerError, "failed to update SM data", err)
		return
	}

	s.respondJSON(w, http.StatusOK, &data)
}

// handleGetAuthSubscription handles GET request for authentication subscription
// TS 29.503, Clause 5.2.3.2.2
func (s *UDRServer) handleGetAuthSubscription(w http.ResponseWriter, r *http.Request) {
	supi := chi.URLParam(r, "supi")

	authSub, err := s.repository.GetAuthenticationSubscription(r.Context(), supi)
	if err != nil {
		s.respondError(w, http.StatusNotFound, "authentication subscription not found", err)
		return
	}

	// Don't expose sensitive keys in response
	response := map[string]interface{}{
		"authenticationMethod":          authSub.AuthenticationMethod,
		"encAlgorithm":                  authSub.EncAlgorithm,
		"authenticationManagementField": authSub.AuthenticationManagementField,
		"sqnScheme":                     authSub.SQNScheme,
		// Note: permanentKey, encOpc should only be used internally
	}

	s.respondJSON(w, http.StatusOK, response)
}

// handleUpdateAuthSubscription handles PUT request to update authentication subscription
func (s *UDRServer) handleUpdateAuthSubscription(w http.ResponseWriter, r *http.Request) {
	supi := chi.URLParam(r, "supi")

	var data repository.AuthenticationSubscription
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		s.respondError(w, http.StatusBadRequest, "invalid request body", err)
		return
	}

	data.SUPI = supi
	err := s.repository.UpdateAuthenticationSubscription(r.Context(), supi, &data)
	if err != nil {
		s.respondError(w, http.StatusInternalServerError, "failed to update auth subscription", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// handleIncrementSQN handles PATCH request to increment SQN
// TS 29.503, Clause 5.2.3.2.4
func (s *UDRServer) handleIncrementSQN(w http.ResponseWriter, r *http.Request) {
	supi := chi.URLParam(r, "supi")

	newSQN, err := s.repository.IncrementSQN(r.Context(), supi)
	if err != nil {
		s.respondError(w, http.StatusInternalServerError, "failed to increment SQN", err)
		return
	}

	s.logger.Debug("SQN incremented",
		zap.String("supi", supi),
		zap.Uint64("new_sqn", newSQN),
	)

	s.respondJSON(w, http.StatusOK, map[string]interface{}{
		"sqn": newSQN,
	})
}

// handleGetPolicyData handles GET request for policy data
// TS 29.519
func (s *UDRServer) handleGetPolicyData(w http.ResponseWriter, r *http.Request) {
	supi := chi.URLParam(r, "supi")

	policyData, err := s.repository.GetPolicyData(r.Context(), supi)
	if err != nil {
		s.respondError(w, http.StatusNotFound, "policy data not found", err)
		return
	}

	s.respondJSON(w, http.StatusOK, policyData)
}

// handleUpdatePolicyData handles PUT request to update policy data
func (s *UDRServer) handleUpdatePolicyData(w http.ResponseWriter, r *http.Request) {
	supi := chi.URLParam(r, "supi")

	var data repository.PolicyData
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		s.respondError(w, http.StatusBadRequest, "invalid request body", err)
		return
	}

	data.SUPI = supi
	err := s.repository.UpdatePolicyData(r.Context(), supi, &data)
	if err != nil {
		s.respondError(w, http.StatusInternalServerError, "failed to update policy data", err)
		return
	}

	s.respondJSON(w, http.StatusOK, &data)
}

// handleGetSubscriptions handles GET request for SDM subscriptions
func (s *UDRServer) handleGetSubscriptions(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement SDM subscriptions listing
	s.respondJSON(w, http.StatusOK, []interface{}{})
}

// handleCreateSubscription handles POST request to create SDM subscription
func (s *UDRServer) handleCreateSubscription(w http.ResponseWriter, r *http.Request) {
	var subscription repository.SDMSubscription
	if err := json.NewDecoder(r.Body).Decode(&subscription); err != nil {
		s.respondError(w, http.StatusBadRequest, "invalid request body", err)
		return
	}

	err := s.repository.CreateSDMSubscription(r.Context(), &subscription)
	if err != nil {
		s.respondError(w, http.StatusInternalServerError, "failed to create subscription", err)
		return
	}

	s.respondJSON(w, http.StatusCreated, &subscription)
}

// handleDeleteSubscription handles DELETE request to remove SDM subscription
func (s *UDRServer) handleDeleteSubscription(w http.ResponseWriter, r *http.Request) {
	subscriptionID := chi.URLParam(r, "subscriptionId")

	err := s.repository.DeleteSDMSubscription(r.Context(), subscriptionID)
	if err != nil {
		s.respondError(w, http.StatusNotFound, "subscription not found", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Administrative Handlers

// handleListSubscribers handles GET request to list all subscribers
func (s *UDRServer) handleListSubscribers(w http.ResponseWriter, r *http.Request) {
	// Parse pagination parameters
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 100 // default
	offset := 0

	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}

	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil {
			offset = o
		}
	}

	subscribers, err := s.repository.ListSubscribers(r.Context(), limit, offset)
	if err != nil {
		s.respondError(w, http.StatusInternalServerError, "failed to list subscribers", err)
		return
	}

	s.respondJSON(w, http.StatusOK, map[string]interface{}{
		"subscribers": subscribers,
		"total":       len(subscribers),
		"limit":       limit,
		"offset":      offset,
	})
}

// handleCreateSubscriber handles POST request to create a new subscriber
func (s *UDRServer) handleCreateSubscriber(w http.ResponseWriter, r *http.Request) {
	var data repository.SubscriberData
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		s.respondError(w, http.StatusBadRequest, "invalid request body", err)
		return
	}

	err := s.repository.CreateSubscriber(r.Context(), &data)
	if err != nil {
		s.respondError(w, http.StatusConflict, "failed to create subscriber", err)
		return
	}

	s.logger.Info("Subscriber created via admin API", zap.String("supi", data.SUPI))
	s.respondJSON(w, http.StatusCreated, &data)
}

// handleGetSubscriber handles GET request for a specific subscriber
func (s *UDRServer) handleGetSubscriber(w http.ResponseWriter, r *http.Request) {
	supi := chi.URLParam(r, "supi")

	subscriber, err := s.repository.GetSubscriber(r.Context(), supi)
	if err != nil {
		s.respondError(w, http.StatusNotFound, "subscriber not found", err)
		return
	}

	s.respondJSON(w, http.StatusOK, subscriber)
}

// handlePutSubscriber handles PUT request to update a subscriber
func (s *UDRServer) handlePutSubscriber(w http.ResponseWriter, r *http.Request) {
	supi := chi.URLParam(r, "supi")

	var data repository.SubscriberData
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		s.respondError(w, http.StatusBadRequest, "invalid request body", err)
		return
	}

	data.SUPI = supi
	err := s.repository.UpdateSubscriber(r.Context(), supi, &data)
	if err != nil {
		s.respondError(w, http.StatusInternalServerError, "failed to update subscriber", err)
		return
	}

	s.respondJSON(w, http.StatusOK, &data)
}

// handleDeleteSubscriber handles DELETE request to remove a subscriber
func (s *UDRServer) handleDeleteSubscriber(w http.ResponseWriter, r *http.Request) {
	supi := chi.URLParam(r, "supi")

	err := s.repository.DeleteSubscriber(r.Context(), supi)
	if err != nil {
		s.respondError(w, http.StatusNotFound, "subscriber not found", err)
		return
	}

	s.logger.Info("Subscriber deleted via admin API", zap.String("supi", supi))
	w.WriteHeader(http.StatusNoContent)
}

// handleGetStats handles GET request for repository statistics
func (s *UDRServer) handleGetStats(w http.ResponseWriter, r *http.Request) {
	stats, err := s.repository.GetStats(r.Context())
	if err != nil {
		s.respondError(w, http.StatusInternalServerError, "failed to get stats", err)
		return
	}

	s.respondJSON(w, http.StatusOK, stats)
}

// handleCreateAuthSubscription handles POST request to create authentication subscription
func (s *UDRServer) handleCreateAuthSubscription(w http.ResponseWriter, r *http.Request) {
	var data repository.AuthenticationSubscription
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		s.respondError(w, http.StatusBadRequest, "invalid request body", err)
		return
	}

	err := s.repository.CreateAuthenticationSubscription(r.Context(), &data)
	if err != nil {
		s.respondError(w, http.StatusConflict, "failed to create auth subscription", err)
		return
	}

	s.logger.Info("Authentication subscription created via admin API", zap.String("supi", data.SUPI))
	s.respondJSON(w, http.StatusCreated, &data)
}
