package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"github.com/your-org/5g-network/nf/ausf/internal/client"
	"go.uber.org/zap"
)

// AuthenticationService handles UE authentication operations
type AuthenticationService struct {
	udmClient *client.UDMClient
	contexts  map[string]*AuthenticationContext // authCtxId -> context
	mu        sync.RWMutex
	logger    *zap.Logger
}

// NewAuthenticationService creates a new authentication service
func NewAuthenticationService(udmClient *client.UDMClient, logger *zap.Logger) *AuthenticationService {
	return &AuthenticationService{
		udmClient: udmClient,
		contexts:  make(map[string]*AuthenticationContext),
		logger:    logger,
	}
}

// AuthenticationContext represents an ongoing authentication session
type AuthenticationContext struct {
	AuthCtxID          string
	SUPI               string
	ServingNetworkName string
	AuthType           string // "5G_AKA" or "EAP_AKA_PRIME"
	RAND               string
	AUTN               string
	HXRES              string
	KAUSF              string
	KSEAF              string // Derived from KAUSF
	CreatedAt          time.Time
	ExpiresAt          time.Time
}

// UEAuthenticationRequest represents authentication initiation request from AMF
type UEAuthenticationRequest struct {
	SUPI               string `json:"supiOrSuci"`
	ServingNetworkName string `json:"servingNetworkName"`
	ResynchronizationInfo *struct {
		RAND string `json:"rand"`
		AUTS string `json:"auts"`
	} `json:"resynchronizationInfo,omitempty"`
}

// UEAuthenticationResponse represents authentication response to AMF
type UEAuthenticationResponse struct {
	AuthType            string                  `json:"authType"`
	Var5gAuthData       *Var5gAuthData         `json:"_5gAuthData,omitempty"`
	Links               map[string]interface{} `json:"_links"`
}

// Var5gAuthData represents 5G authentication data
type Var5gAuthData struct {
	RAND  string `json:"rand"`
	AUTN  string `json:"autn"`
}

// ConfirmationData represents authentication confirmation from AMF
type ConfirmationData struct {
	RES string `json:"resStar"` // RES* from UE
}

// ConfirmationDataResponse represents authentication confirmation response
type ConfirmationDataResponse struct {
	AuthResult string  `json:"authResult"` // "AUTHENTICATION_SUCCESS" or "AUTHENTICATION_FAILURE"
	SUPI       string  `json:"supi,omitempty"`
	KSEAF      string  `json:"kseaf,omitempty"`
}

// UEAuthenticationCtx initiates authentication for a UE
func (s *AuthenticationService) UEAuthenticationCtx(ctx context.Context, req *UEAuthenticationRequest) (*UEAuthenticationResponse, error) {
	s.logger.Info("Initiating UE authentication",
		zap.String("supi", req.SUPI),
		zap.String("serving_network", req.ServingNetworkName),
	)

	// Request authentication vector from UDM
	authInfo := &client.AuthenticationInfo{
		SUPI:               req.SUPI,
		ServingNetworkName: req.ServingNetworkName,
		ResynchronizationInfo: req.ResynchronizationInfo,
	}

	authResult, err := s.udmClient.GenerateAuthData(ctx, authInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to get auth data from UDM: %w", err)
	}

	if authResult.AuthenticationVector == nil {
		return nil, fmt.Errorf("no authentication vector received from UDM")
	}

	// Generate authentication context ID
	authCtxID := s.generateAuthCtxID()

	// Derive KSEAF from KAUSF
	// KSEAF = KDF(KAUSF, serving network name)
	// Simplified: In production, use proper 3GPP KDF
	kseaf := s.deriveKSEAF(authResult.AuthenticationVector.KAUSF, req.ServingNetworkName)

	// Store authentication context
	authCtx := &AuthenticationContext{
		AuthCtxID:          authCtxID,
		SUPI:               req.SUPI,
		ServingNetworkName: req.ServingNetworkName,
		AuthType:           authResult.AuthType,
		RAND:               authResult.AuthenticationVector.RAND,
		AUTN:               authResult.AuthenticationVector.AUTN,
		HXRES:              authResult.AuthenticationVector.HXRES,
		KAUSF:              authResult.AuthenticationVector.KAUSF,
		KSEAF:              kseaf,
		CreatedAt:          time.Now(),
		ExpiresAt:          time.Now().Add(5 * time.Minute),
	}

	s.mu.Lock()
	s.contexts[authCtxID] = authCtx
	s.mu.Unlock()

	s.logger.Info("Authentication context created",
		zap.String("supi", req.SUPI),
		zap.String("auth_ctx_id", authCtxID),
		zap.String("auth_type", authResult.AuthType),
	)

	// Build response
	response := &UEAuthenticationResponse{
		AuthType: authResult.AuthType,
		Var5gAuthData: &Var5gAuthData{
			RAND: authResult.AuthenticationVector.RAND,
			AUTN: authResult.AuthenticationVector.AUTN,
		},
		Links: map[string]interface{}{
			"5g-aka": map[string]string{
				"href": fmt.Sprintf("/nausf-auth/v1/ue-authentications/%s/5g-aka-confirmation", authCtxID),
			},
		},
	}

	return response, nil
}

// Confirm5gAkaAuth confirms 5G-AKA authentication
func (s *AuthenticationService) Confirm5gAkaAuth(ctx context.Context, authCtxID string, confirmData *ConfirmationData) (*ConfirmationDataResponse, error) {
	s.logger.Info("Confirming 5G-AKA authentication",
		zap.String("auth_ctx_id", authCtxID),
	)

	// Retrieve authentication context
	s.mu.RLock()
	authCtx, exists := s.contexts[authCtxID]
	s.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("authentication context not found: %s", authCtxID)
	}

	// Check if context expired
	if time.Now().After(authCtx.ExpiresAt) {
		s.mu.Lock()
		delete(s.contexts, authCtxID)
		s.mu.Unlock()
		return nil, fmt.Errorf("authentication context expired")
	}

	// Verify RES* matches HXRES*
	// In production, compute HRES* from RES* and compare with stored HXRES*
	// Simplified: Direct comparison
	authSuccess := confirmData.RES == authCtx.HXRES

	var response *ConfirmationDataResponse
	if authSuccess {
		s.logger.Info("Authentication successful",
			zap.String("supi", authCtx.SUPI),
			zap.String("auth_ctx_id", authCtxID),
		)

		response = &ConfirmationDataResponse{
			AuthResult: "AUTHENTICATION_SUCCESS",
			SUPI:       authCtx.SUPI,
			KSEAF:      authCtx.KSEAF,
		}

		// Notify UDM of successful authentication
		authEvent := map[string]interface{}{
			"nfInstanceId":       "ausf-1", // Should use actual instance ID
			"success":            true,
			"timeStamp":          time.Now().Format(time.RFC3339),
			"authType":           authCtx.AuthType,
			"servingNetworkName": authCtx.ServingNetworkName,
		}
		
		if err := s.udmClient.ConfirmAuth(ctx, authCtx.SUPI, authEvent); err != nil {
			s.logger.Error("Failed to confirm auth with UDM", zap.Error(err))
			// Continue anyway - authentication was successful
		}
	} else {
		s.logger.Warn("Authentication failed",
			zap.String("supi", authCtx.SUPI),
			zap.String("auth_ctx_id", authCtxID),
		)

		response = &ConfirmationDataResponse{
			AuthResult: "AUTHENTICATION_FAILURE",
		}
	}

	// Clean up authentication context
	s.mu.Lock()
	delete(s.contexts, authCtxID)
	s.mu.Unlock()

	return response, nil
}

// GetAuthContext retrieves an authentication context
func (s *AuthenticationService) GetAuthContext(authCtxID string) (*AuthenticationContext, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	authCtx, exists := s.contexts[authCtxID]
	if !exists {
		return nil, fmt.Errorf("authentication context not found")
	}

	return authCtx, nil
}

// GetStats returns authentication statistics
func (s *AuthenticationService) GetStats() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return map[string]interface{}{
		"active_contexts": len(s.contexts),
	}
}

// generateAuthCtxID generates a unique authentication context ID
func (s *AuthenticationService) generateAuthCtxID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// deriveKSEAF derives KSEAF from KAUSF
// KSEAF = KDF(KAUSF, serving network name)
// Simplified implementation - in production use 3GPP KDF
func (s *AuthenticationService) deriveKSEAF(kausfHex, servingNetworkName string) string {
	kausf, _ := hex.DecodeString(kausfHex)
	
	// Simple KDF using SHA-256
	h := sha256.New()
	h.Write(kausf)
	h.Write([]byte(servingNetworkName))
	h.Write([]byte("KSEAF"))
	
	kseaf := h.Sum(nil)
	return hex.EncodeToString(kseaf)
}

// CleanupExpiredContexts removes expired authentication contexts
func (s *AuthenticationService) CleanupExpiredContexts() {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	for id, ctx := range s.contexts {
		if now.After(ctx.ExpiresAt) {
			delete(s.contexts, id)
			s.logger.Debug("Removed expired auth context", zap.String("auth_ctx_id", id))
		}
	}
}
