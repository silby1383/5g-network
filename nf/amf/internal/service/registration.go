package service

import (
	"context"
	"fmt"

	"github.com/your-org/5g-network/nf/amf/internal/client"
	"github.com/your-org/5g-network/nf/amf/internal/config"
	amfcontext "github.com/your-org/5g-network/nf/amf/internal/context"
	"go.uber.org/zap"
)

// RegistrationService handles UE registration procedures
type RegistrationService struct {
	config         *config.Config
	ausfClient     *client.AUSFClient
	contextManager *amfcontext.UEContextManager
	logger         *zap.Logger
}

// NewRegistrationService creates a new registration service
func NewRegistrationService(
	cfg *config.Config,
	ausfClient *client.AUSFClient,
	contextManager *amfcontext.UEContextManager,
	logger *zap.Logger,
) *RegistrationService {
	return &RegistrationService{
		config:         cfg,
		ausfClient:     ausfClient,
		contextManager: contextManager,
		logger:         logger,
	}
}

// RegistrationRequest represents a UE registration request
type RegistrationRequest struct {
	SUPI             string              `json:"supi"`
	RegistrationType string              `json:"registrationType"` // "INITIAL", "MOBILITY", "PERIODIC"
	FollowOnRequest  bool                `json:"followOnRequest"`
	RequestedNSSAI   []amfcontext.SNSSAI `json:"requestedNssai,omitempty"`
}

// RegistrationResponse represents a registration response
type RegistrationResponse struct {
	Result          string                          `json:"result"` // "SUCCESS", "FAILURE"
	SUPI            string                          `json:"supi"`
	GUAMI           string                          `json:"guami"`
	AllowedNSSAI    []amfcontext.SNSSAI             `json:"allowedNssai,omitempty"`
	ConfiguredNSSAI []amfcontext.SNSSAI             `json:"configuredNssai,omitempty"`
	TAI             amfcontext.TrackingAreaIdentity `json:"tai"`
	T3512           int                             `json:"t3512"` // Periodic registration timer
	Reason          string                          `json:"reason,omitempty"`
}

// AuthenticationRequest represents an authentication request
type AuthenticationRequest struct {
	SUPI string `json:"supi"`
}

// AuthenticationResponse represents an authentication response
type AuthenticationResponse struct {
	AuthType  string `json:"authType"`
	AuthCtxID string `json:"authCtxId"`
	RAND      string `json:"rand"`
	AUTN      string `json:"autn"`
}

// AuthenticationConfirmRequest represents an authentication confirmation
type AuthenticationConfirmRequest struct {
	AuthCtxID string `json:"authCtxId"`
	RES       string `json:"resStar"`
}

// AuthenticationConfirmResponse represents confirmation response
type AuthenticationConfirmResponse struct {
	Result string `json:"result"` // "SUCCESS", "FAILURE"
	SUPI   string `json:"supi,omitempty"`
	KSEAF  string `json:"kseaf,omitempty"`
}

// InitiateAuthentication initiates UE authentication
func (s *RegistrationService) InitiateAuthentication(ctx context.Context, req *AuthenticationRequest) (*AuthenticationResponse, error) {
	s.logger.Info("Initiating UE authentication",
		zap.String("supi", req.SUPI),
	)

	// Get or create UE context
	ueCtx := s.contextManager.GetOrCreateContext(req.SUPI)

	// Build serving network name
	servingNetworkName := fmt.Sprintf("5G:mnc%s.mcc%s.3gppnetwork.org",
		s.config.PLMN.MNC, s.config.PLMN.MCC)

	// Request authentication from AUSF
	ausfReq := &client.UEAuthenticationRequest{
		SUPI:               req.SUPI,
		ServingNetworkName: servingNetworkName,
	}

	ausfResp, err := s.ausfClient.InitiateAuthentication(ctx, ausfReq)
	if err != nil {
		return nil, fmt.Errorf("failed to initiate authentication with AUSF: %w", err)
	}

	// Store authentication context temporarily
	ueCtx.UpdateConnectionState(amfcontext.ConnectionStateConnected)

	s.logger.Info("Authentication initiated via AUSF",
		zap.String("supi", req.SUPI),
		zap.String("auth_ctx_id", ausfResp.AuthCtxID),
		zap.String("auth_type", ausfResp.AuthType),
	)

	// Return challenge to UE
	return &AuthenticationResponse{
		AuthType:  ausfResp.AuthType,
		AuthCtxID: ausfResp.AuthCtxID,
		RAND:      ausfResp.Var5gAuthData.RAND,
		AUTN:      ausfResp.Var5gAuthData.AUTN,
	}, nil
}

// ConfirmAuthentication confirms UE authentication
func (s *RegistrationService) ConfirmAuthentication(ctx context.Context, req *AuthenticationConfirmRequest) (*AuthenticationConfirmResponse, error) {
	s.logger.Info("Confirming UE authentication",
		zap.String("auth_ctx_id", req.AuthCtxID),
	)

	// Confirm with AUSF
	ausfResp, err := s.ausfClient.ConfirmAuthentication(ctx, req.AuthCtxID, req.RES)
	if err != nil {
		return nil, fmt.Errorf("failed to confirm authentication with AUSF: %w", err)
	}

	if ausfResp.AuthResult != "AUTHENTICATION_SUCCESS" {
		s.logger.Warn("Authentication failed",
			zap.String("auth_ctx_id", req.AuthCtxID),
			zap.String("result", ausfResp.AuthResult),
		)

		return &AuthenticationConfirmResponse{
			Result: "FAILURE",
		}, nil
	}

	// Get UE context
	ueCtx, exists := s.contextManager.GetContext(ausfResp.SUPI)
	if !exists {
		return nil, fmt.Errorf("UE context not found for SUPI: %s", ausfResp.SUPI)
	}

	// Establish security context with KSEAF from AUSF
	secCtx := &amfcontext.SecurityContext{
		KSEAF:                  ausfResp.KSEAF,
		NASSecurityEstablished: true,
		IntegrityAlgorithm:     s.config.Security.IntegrityOrder[0],
		CipheringAlgorithm:     s.config.Security.CipheringOrder[0],
	}
	ueCtx.SetSecurityContext(secCtx)

	s.logger.Info("Authentication successful",
		zap.String("supi", ausfResp.SUPI),
		zap.String("auth_ctx_id", req.AuthCtxID),
	)

	return &AuthenticationConfirmResponse{
		Result: "SUCCESS",
		SUPI:   ausfResp.SUPI,
		KSEAF:  ausfResp.KSEAF,
	}, nil
}

// RegisterUE handles UE registration
func (s *RegistrationService) RegisterUE(ctx context.Context, req *RegistrationRequest) (*RegistrationResponse, error) {
	s.logger.Info("Processing UE registration",
		zap.String("supi", req.SUPI),
		zap.String("type", req.RegistrationType),
	)

	// Get UE context
	ueCtx, exists := s.contextManager.GetContext(req.SUPI)
	if !exists {
		return &RegistrationResponse{
			Result: "FAILURE",
			Reason: "UE not authenticated",
		}, nil
	}

	// Check if security context is established
	if ueCtx.SecurityContext == nil || !ueCtx.SecurityContext.NASSecurityEstablished {
		return &RegistrationResponse{
			Result: "FAILURE",
			Reason: "Security context not established",
		}, nil
	}

	// Determine allowed NSSAI (simplified - accept all requested)
	allowedNSSAI := req.RequestedNSSAI
	if len(allowedNSSAI) == 0 {
		// Use default from config
		allowedNSSAI = make([]amfcontext.SNSSAI, len(s.config.AMF.SupportedSNSSAI))
		for i, snssai := range s.config.AMF.SupportedSNSSAI {
			allowedNSSAI[i] = amfcontext.SNSSAI{
				SST: snssai.SST,
				SD:  snssai.SD,
			}
		}
	}

	// Update UE context
	ueCtx.AllowedNSSAI = allowedNSSAI
	ueCtx.ConfiguredNSSAI = allowedNSSAI
	ueCtx.GUAMI = s.config.GetGUAMI()
	ueCtx.AMFRegionID = s.config.AMF.RegionID
	ueCtx.AMFSetID = s.config.AMF.SetID
	ueCtx.AMFPointer = s.config.AMF.Pointer
	ueCtx.TAI = amfcontext.TrackingAreaIdentity{
		PLMNID: amfcontext.PLMNID{
			MCC: s.config.PLMN.MCC,
			MNC: s.config.PLMN.MNC,
		},
		TAC: s.config.PLMN.TAC,
	}
	ueCtx.UpdateRegistrationState(amfcontext.RegistrationStateRegistered)

	s.logger.Info("UE registered successfully",
		zap.String("supi", req.SUPI),
		zap.String("guami", ueCtx.GUAMI),
	)

	return &RegistrationResponse{
		Result:          "SUCCESS",
		SUPI:            req.SUPI,
		GUAMI:           ueCtx.GUAMI,
		AllowedNSSAI:    allowedNSSAI,
		ConfiguredNSSAI: allowedNSSAI,
		TAI:             ueCtx.TAI,
		T3512:           s.config.Timers.T3512,
	}, nil
}

// DeregisterUE handles UE deregistration
func (s *RegistrationService) DeregisterUE(ctx context.Context, supi string) error {
	s.logger.Info("Processing UE deregistration",
		zap.String("supi", supi),
	)

	ueCtx, exists := s.contextManager.GetContext(supi)
	if !exists {
		return fmt.Errorf("UE context not found")
	}

	// Update state
	ueCtx.UpdateRegistrationState(amfcontext.RegistrationStateDeregistered)
	ueCtx.UpdateConnectionState(amfcontext.ConnectionStateIdle)

	// Remove context
	s.contextManager.RemoveContext(supi)

	s.logger.Info("UE deregistered",
		zap.String("supi", supi),
	)

	return nil
}

// GetRegistrationStats returns registration statistics
func (s *RegistrationService) GetRegistrationStats() map[string]interface{} {
	return map[string]interface{}{
		"total_contexts": len(s.contextManager.GetAllContexts()),
		"registered_ues": s.contextManager.GetRegisteredCount(),
		"connected_ues":  s.contextManager.GetConnectedCount(),
	}
}
