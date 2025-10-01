package server

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/your-org/5g-network/common/metrics"
	"github.com/your-org/5g-network/nf/nrf/internal/repository"
	"go.uber.org/zap"
)

// handleNFRegister handles NF registration (PUT /nf-instances/{nfInstanceId})
// TS 29.510, Clause 5.2.2.2.1
func (s *NRFServer) handleNFRegister(w http.ResponseWriter, r *http.Request) {
	nfInstanceID := chi.URLParam(r, "nfInstanceId")

	// Parse request body
	var profile repository.NFProfile
	if err := json.NewDecoder(r.Body).Decode(&profile); err != nil {
		s.respondError(w, http.StatusBadRequest, "invalid request body", err)
		return
	}

	// Set NF instance ID from URL
	profile.NFInstanceID = nfInstanceID

	// Register NF
	err := s.repository.Register(r.Context(), &profile)
	if err != nil {
		s.respondError(w, http.StatusConflict, "registration failed", err)
		metrics.RecordNFRegistration("unknown", "failed")
		return
	}

	// Record successful registration
	metrics.RecordNFRegistration(string(profile.NFType), "success")
	stats, _ := s.repository.GetStats(r.Context())
	metrics.SetRegisteredNFs(string(profile.NFType), stats.NFsByType[string(profile.NFType)])

	// Return registered profile
	s.respondJSON(w, http.StatusCreated, &profile)

	s.logger.Info("NF registered",
		zap.String("nf_instance_id", nfInstanceID),
		zap.String("nf_type", string(profile.NFType)),
	)
}

// handleNFUpdate handles NF profile update (PATCH /nf-instances/{nfInstanceId})
// TS 29.510, Clause 5.2.2.2.2
func (s *NRFServer) handleNFUpdate(w http.ResponseWriter, r *http.Request) {
	nfInstanceID := chi.URLParam(r, "nfInstanceId")

	// Parse request body
	var profile repository.NFProfile
	if err := json.NewDecoder(r.Body).Decode(&profile); err != nil {
		s.respondError(w, http.StatusBadRequest, "invalid request body", err)
		return
	}

	// Update NF
	err := s.repository.Update(r.Context(), nfInstanceID, &profile)
	if err != nil {
		s.respondError(w, http.StatusNotFound, "update failed", err)
		return
	}

	// Return updated profile
	s.respondJSON(w, http.StatusOK, &profile)

	s.logger.Info("NF profile updated",
		zap.String("nf_instance_id", nfInstanceID),
	)
}

// handleNFDeregister handles NF deregistration (DELETE /nf-instances/{nfInstanceId})
// TS 29.510, Clause 5.2.2.2.3
func (s *NRFServer) handleNFDeregister(w http.ResponseWriter, r *http.Request) {
	nfInstanceID := chi.URLParam(r, "nfInstanceId")

	// Deregister NF
	err := s.repository.Deregister(r.Context(), nfInstanceID)
	if err != nil {
		s.respondError(w, http.StatusNotFound, "deregistration failed", err)
		metrics.RecordNFDeregistration("failed")
		return
	}

	// Record successful deregistration
	metrics.RecordNFDeregistration("unknown") // We don't have the NF type here
	stats, _ := s.repository.GetStats(r.Context())
	// Update all NF type counts
	for nfType, count := range stats.NFsByType {
		metrics.SetRegisteredNFs(nfType, count)
	}

	w.WriteHeader(http.StatusNoContent)

	s.logger.Info("NF deregistered",
		zap.String("nf_instance_id", nfInstanceID),
	)
}

// handleNFGet handles getting an NF profile (GET /nf-instances/{nfInstanceId})
func (s *NRFServer) handleNFGet(w http.ResponseWriter, r *http.Request) {
	nfInstanceID := chi.URLParam(r, "nfInstanceId")

	// Get NF profile
	profile, err := s.repository.Get(r.Context(), nfInstanceID)
	if err != nil {
		s.respondError(w, http.StatusNotFound, "NF not found", err)
		return
	}

	s.respondJSON(w, http.StatusOK, profile)
}

// handleNFList handles listing all NF profiles (GET /nf-instances)
func (s *NRFServer) handleNFList(w http.ResponseWriter, r *http.Request) {
	// Get all NF profiles
	profiles, err := s.repository.GetAll(r.Context())
	if err != nil {
		s.respondError(w, http.StatusInternalServerError, "failed to get profiles", err)
		return
	}

	s.respondJSON(w, http.StatusOK, map[string]interface{}{
		"nfInstances": profiles,
		"totalCount":  len(profiles),
	})
}

// handleHeartbeat handles NF heartbeat (PATCH /nf-instances/{nfInstanceId}/heartbeat)
// TS 29.510, Clause 5.2.2.2.4
func (s *NRFServer) handleHeartbeat(w http.ResponseWriter, r *http.Request) {
	nfInstanceID := chi.URLParam(r, "nfInstanceId")

	// Get NF profile for metrics
	profile, _ := s.repository.Get(r.Context(), nfInstanceID)

	// Update heartbeat
	err := s.repository.UpdateHeartbeat(r.Context(), nfInstanceID)
	if err != nil {
		s.respondError(w, http.StatusNotFound, "heartbeat failed", err)
		return
	}

	if profile != nil {
		metrics.RecordHeartbeat(string(profile.NFType))
	}

	// Return patch result with updated heartbeat time
	s.respondJSON(w, http.StatusNoContent, nil)

	s.logger.Debug("Heartbeat received",
		zap.String("nf_instance_id", nfInstanceID),
	)
}

// handleNFDiscover handles NF discovery (GET /nnrf-disc/v1/nf-instances)
// TS 29.510, Clause 5.2.3.2.2
func (s *NRFServer) handleNFDiscover(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	query := &repository.DiscoveryQuery{}

	// Extract common query parameters
	if nfType := r.URL.Query().Get("target-nf-type"); nfType != "" {
		query.NFType = repository.NFType(nfType)
	}

	if nfID := r.URL.Query().Get("target-nf-instance-id"); nfID != "" {
		query.TargetNFID = nfID
	}

	if requesterFQDN := r.URL.Query().Get("requester-nf-fqdn"); requesterFQDN != "" {
		query.RequesterFQDN = requesterFQDN
	}

	// PLMN ID
	if mcc := r.URL.Query().Get("requester-plmn-mcc"); mcc != "" {
		if mnc := r.URL.Query().Get("requester-plmn-mnc"); mnc != "" {
			query.PLMNID = &repository.PLMNID{
				MCC: mcc,
				MNC: mnc,
			}
		}
	}

	// AMF-specific parameters
	if amfRegionID := r.URL.Query().Get("target-amf-region-id"); amfRegionID != "" {
		query.AMFRegionID = amfRegionID
	}

	if amfSetID := r.URL.Query().Get("target-amf-set-id"); amfSetID != "" {
		query.AMFSetID = amfSetID
	}

	// SMF-specific parameters
	if dnn := r.URL.Query().Get("dnn"); dnn != "" {
		query.DNN = dnn
	}

	// TAI
	if tac := r.URL.Query().Get("tai-tac"); tac != "" {
		if mcc := r.URL.Query().Get("tai-plmn-mcc"); mcc != "" {
			if mnc := r.URL.Query().Get("tai-plmn-mnc"); mnc != "" {
				query.TAI = &repository.TAI{
					PLMNID: repository.PLMNID{MCC: mcc, MNC: mnc},
					TAC:    tac,
				}
			}
		}
	}

	// Perform discovery
	profiles, err := s.repository.Discover(r.Context(), query)
	if err != nil {
		s.respondError(w, http.StatusInternalServerError, "discovery failed", err)
		metrics.RecordDiscoveryRequest(string(query.NFType), "failed")
		return
	}

	// Record successful discovery
	metrics.RecordDiscoveryRequest(string(query.NFType), "success")

	// Return results
	s.respondJSON(w, http.StatusOK, map[string]interface{}{
		"validityPeriod": 3600, // seconds
		"nfInstances":    profiles,
		"searchId":       uuid.New().String(),
	})

	s.logger.Info("NF discovery",
		zap.String("target_nf_type", string(query.NFType)),
		zap.Int("results_count", len(profiles)),
	)
}

// handleSubscribe handles subscription creation (POST /subscriptions)
// TS 29.510, Clause 5.2.2.3.1
func (s *NRFServer) handleSubscribe(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	var subscription repository.Subscription
	if err := json.NewDecoder(r.Body).Decode(&subscription); err != nil {
		s.respondError(w, http.StatusBadRequest, "invalid request body", err)
		return
	}

	// Generate subscription ID if not provided
	if subscription.SubscriptionID == "" {
		subscription.SubscriptionID = uuid.New().String()
	}

	// Set validity time if not provided (default: 24 hours)
	if subscription.ValidityTime.IsZero() {
		subscription.ValidityTime = time.Now().Add(24 * time.Hour)
	}

	// Create subscription
	err := s.repository.Subscribe(r.Context(), &subscription)
	if err != nil {
		s.respondError(w, http.StatusInternalServerError, "subscription failed", err)
		return
	}

	// Return subscription
	s.respondJSON(w, http.StatusCreated, &subscription)

	s.logger.Info("Subscription created",
		zap.String("subscription_id", subscription.SubscriptionID),
		zap.String("callback_uri", subscription.CallbackURI),
	)
}

// handleUnsubscribe handles subscription deletion (DELETE /subscriptions/{subscriptionId})
// TS 29.510, Clause 5.2.2.3.2
func (s *NRFServer) handleUnsubscribe(w http.ResponseWriter, r *http.Request) {
	subscriptionID := chi.URLParam(r, "subscriptionId")

	// Delete subscription
	err := s.repository.Unsubscribe(r.Context(), subscriptionID)
	if err != nil {
		s.respondError(w, http.StatusNotFound, "unsubscribe failed", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)

	s.logger.Info("Subscription removed",
		zap.String("subscription_id", subscriptionID),
	)
}

// handleGetSubscription handles getting a subscription (GET /subscriptions/{subscriptionId})
func (s *NRFServer) handleGetSubscription(w http.ResponseWriter, r *http.Request) {
	subscriptionID := chi.URLParam(r, "subscriptionId")

	// Get subscription
	subscription, err := s.repository.GetSubscription(r.Context(), subscriptionID)
	if err != nil {
		s.respondError(w, http.StatusNotFound, "subscription not found", err)
		return
	}

	s.respondJSON(w, http.StatusOK, subscription)
}
