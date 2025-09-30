package service

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"fmt"

	"github.com/your-org/5g-network/nf/udm/internal/client"
	"github.com/your-org/5g-network/nf/udm/internal/crypto"
	"go.uber.org/zap"
)

// AuthenticationService handles UE authentication operations
type AuthenticationService struct {
	udrClient *client.UDRClient
	logger    *zap.Logger
}

// NewAuthenticationService creates a new authentication service
func NewAuthenticationService(udrClient *client.UDRClient, logger *zap.Logger) *AuthenticationService {
	return &AuthenticationService{
		udrClient: udrClient,
		logger:    logger,
	}
}

// AuthenticationInfo represents authentication information request
type AuthenticationInfo struct {
	SUPI                  string `json:"supi"`
	ServingNetworkName    string `json:"servingNetworkName"`
	ResynchronizationInfo *struct {
		RAND []byte `json:"rand"`
		AUTS []byte `json:"auts"`
	} `json:"resynchronizationInfo,omitempty"`
}

// AuthenticationInfoResult represents the authentication response
type AuthenticationInfoResult struct {
	AuthType             string       `json:"authType"` // "5G_AKA" or "EAP_AKA_PRIME"
	AuthenticationVector *AVType5GAKA `json:"authenticationVector,omitempty"`
}

// AVType5GAKA represents a 5G AKA authentication vector
type AVType5GAKA struct {
	RAND  string `json:"rand"`  // Random challenge (hex)
	AUTN  string `json:"autn"`  // Authentication token (hex)
	HXRES string `json:"hxres"` // Expected response (hex)
	KAUSF string `json:"kausf"` // Key for AUSF (hex)
}

// GenerateAuthData generates authentication vectors for a UE
func (s *AuthenticationService) GenerateAuthData(ctx context.Context, authInfo *AuthenticationInfo) (*AuthenticationInfoResult, error) {
	s.logger.Info("Generating authentication data",
		zap.String("supi", authInfo.SUPI),
		zap.String("serving_network", authInfo.ServingNetworkName),
	)

	// Get authentication subscription from UDR
	authSub, err := s.udrClient.GetAuthenticationSubscription(ctx, authInfo.SUPI)
	if err != nil {
		return nil, fmt.Errorf("failed to get authentication subscription: %w", err)
	}

	// Parse permanent key (K)
	k, err := crypto.HexToBytes(authSub.PermanentKey)
	if err != nil {
		return nil, fmt.Errorf("invalid permanent key: %w", err)
	}

	// Parse OPc
	var opc []byte
	if authSub.EncOPC != "" {
		opc, err = crypto.HexToBytes(authSub.EncOPC)
		if err != nil {
			return nil, fmt.Errorf("invalid OPc: %w", err)
		}
	} else if authSub.EncOP != "" {
		// Compute OPc from OP
		op, err := crypto.HexToBytes(authSub.EncOP)
		if err != nil {
			return nil, fmt.Errorf("invalid OP: %w", err)
		}
		opc, err = crypto.ComputeOPc(k, op)
		if err != nil {
			return nil, fmt.Errorf("failed to compute OPc: %w", err)
		}
	} else {
		return nil, fmt.Errorf("neither OPc nor OP provided")
	}

	// Generate random RAND
	randBytes := make([]byte, 16)
	if _, err := rand.Read(randBytes); err != nil {
		return nil, fmt.Errorf("failed to generate RAND: %w", err)
	}

	// Get and increment SQN from UDR
	sqnValue, err := s.udrClient.IncrementSQN(ctx, authInfo.SUPI)
	if err != nil {
		return nil, fmt.Errorf("failed to increment SQN: %w", err)
	}

	// Convert SQN to bytes (48 bits)
	sqnBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(sqnBytes, sqnValue)
	sqn := sqnBytes[2:8] // Take lower 48 bits

	// Parse AMF
	amf, err := crypto.HexToBytes(authSub.AuthenticationManagementField)
	if err != nil {
		// Default AMF value
		amf = []byte{0x80, 0x00}
	}

	// Generate authentication vector using MILENAGE
	av, err := crypto.GenerateAuthVector(k, opc, randBytes, sqn, amf)
	if err != nil {
		return nil, fmt.Errorf("failed to generate auth vector: %w", err)
	}

	// Derive KAUSF (for 5G)
	// KAUSF = KDF(CK || IK, SN name, SQN ⊕ AK)
	// Simplified version for now - in production, use proper KDF
	kausf := make([]byte, 32)
	copy(kausf[:16], av.CK)
	copy(kausf[16:], av.IK)

	// Compute HXRES* (hash of XRES for 5G)
	// Simplified: In production, use SHA-256(RAND || XRES || serving network name)
	hxres := av.XRES

	s.logger.Info("Generated authentication vector",
		zap.String("supi", authInfo.SUPI),
		zap.String("auth_method", authSub.AuthenticationMethod),
	)

	return &AuthenticationInfoResult{
		AuthType: "5G_AKA",
		AuthenticationVector: &AVType5GAKA{
			RAND:  crypto.BytesToHex(av.RAND),
			AUTN:  crypto.BytesToHex(av.AUTN),
			HXRES: crypto.BytesToHex(hxres),
			KAUSF: crypto.BytesToHex(kausf),
		},
	}, nil
}

// ConfirmAuth confirms authentication result
func (s *AuthenticationService) ConfirmAuth(ctx context.Context, supi string, authEvent interface{}) error {
	s.logger.Info("Confirming authentication", zap.String("supi", supi))
	// In production, store authentication event in UDR
	return nil
}
