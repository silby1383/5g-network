package service

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

// UECMService handles UE Context Management (Nudm_UECM)
type UECMService struct {
	contexts map[string]*UEContext // supi -> UE context
	mu       sync.RWMutex
	logger   *zap.Logger
}

// NewUECMService creates a new UECM service
func NewUECMService(logger *zap.Logger) *UECMService {
	return &UECMService{
		contexts: make(map[string]*UEContext),
		logger:   logger,
	}
}

// UEContext represents UE context information
type UEContext struct {
	SUPI               string    `json:"supi"`
	AMFInstanceID      string    `json:"amfInstanceId,omitempty"`
	GUAMI              *GUAMI    `json:"guami,omitempty"`
	PEI                string    `json:"pei,omitempty"` // Permanent Equipment Identifier
	UDMGroupID         string    `json:"udmGroupId,omitempty"`
	RoutingIndicator   string    `json:"routingIndicator,omitempty"`
	RegistrationTime   time.Time `json:"registrationTime,omitempty"`
	DeregistrationTime time.Time `json:"deregistrationTime,omitempty"`
	PurgeFlag          bool      `json:"purgeFlag,omitempty"`
	IratChangeAllowed  bool      `json:"iratChangeAllowed,omitempty"`
}

// GUAMI represents Globally Unique AMF Identifier
type GUAMI struct {
	PlmnID      PlmnID `json:"plmnId"`
	AMFRegionID string `json:"amfRegionId"`
	AMFSetID    string `json:"amfSetId"`
	AMFPointer  string `json:"amfPointer"`
}

// PlmnID represents PLMN identifier
type PlmnID struct {
	MCC string `json:"mcc"`
	MNC string `json:"mnc"`
}

// AMF3GPPAccessRegistration represents AMF registration for 3GPP access
type AMF3GPPAccessRegistration struct {
	AMFInstanceID          string        `json:"amfInstanceId"`
	DeregestrationReason   string        `json:"deregCallbackUri,omitempty"`
	GUAMI                  *GUAMI        `json:"guami,omitempty"`
	RATType                string        `json:"ratType"` // NR, EUTRA
	InitialRegistrationInd bool          `json:"initialRegistrationInd,omitempty"`
	BackupAMFInfo          []interface{} `json:"backupAmfInfo,omitempty"`
}

// RegisterAMF3GPPAccess registers AMF context for 3GPP access
func (s *UECMService) RegisterAMF3GPPAccess(ctx context.Context, supi string, registration *AMF3GPPAccessRegistration) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.logger.Info("Registering AMF context",
		zap.String("supi", supi),
		zap.String("amf_instance_id", registration.AMFInstanceID),
		zap.String("rat_type", registration.RATType),
	)

	// Create or update UE context
	ueContext, exists := s.contexts[supi]
	if !exists {
		ueContext = &UEContext{
			SUPI: supi,
		}
		s.contexts[supi] = ueContext
	}

	// Update context with AMF information
	ueContext.AMFInstanceID = registration.AMFInstanceID
	ueContext.GUAMI = registration.GUAMI
	ueContext.RegistrationTime = time.Now()
	ueContext.PurgeFlag = false

	s.logger.Info("AMF context registered",
		zap.String("supi", supi),
		zap.String("amf_instance_id", registration.AMFInstanceID),
	)

	return nil
}

// UpdateAMF3GPPAccess updates AMF context
func (s *UECMService) UpdateAMF3GPPAccess(ctx context.Context, supi string, updates map[string]interface{}) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.logger.Info("Updating AMF context",
		zap.String("supi", supi),
	)

	_, exists := s.contexts[supi]
	if !exists {
		return fmt.Errorf("UE context not found for SUPI: %s", supi)
	}

	// Update context fields
	// In production, parse and apply updates from the map

	s.logger.Debug("AMF context updated",
		zap.String("supi", supi),
	)

	return nil
}

// DeregisterAMF3GPPAccess deregisters AMF context
func (s *UECMService) DeregisterAMF3GPPAccess(ctx context.Context, supi string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.logger.Info("Deregistering AMF context",
		zap.String("supi", supi),
	)

	ueContext, exists := s.contexts[supi]
	if !exists {
		return fmt.Errorf("UE context not found for SUPI: %s", supi)
	}

	// Mark as deregistered
	ueContext.DeregistrationTime = time.Now()
	ueContext.AMFInstanceID = ""

	// Optionally delete the context
	// delete(s.contexts, supi)

	s.logger.Info("AMF context deregistered",
		zap.String("supi", supi),
	)

	return nil
}

// Get3GPPRegistration retrieves AMF registration information
func (s *UECMService) Get3GPPRegistration(ctx context.Context, supi string) (*AMF3GPPAccessRegistration, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	s.logger.Debug("Getting AMF registration",
		zap.String("supi", supi),
	)

	ueContext, exists := s.contexts[supi]
	if !exists {
		return nil, fmt.Errorf("UE context not found for SUPI: %s", supi)
	}

	if ueContext.AMFInstanceID == "" {
		return nil, fmt.Errorf("no AMF registration found for SUPI: %s", supi)
	}

	registration := &AMF3GPPAccessRegistration{
		AMFInstanceID: ueContext.AMFInstanceID,
		GUAMI:         ueContext.GUAMI,
		RATType:       "NR",
	}

	return registration, nil
}

// GetUEContext retrieves UE context
func (s *UECMService) GetUEContext(ctx context.Context, supi string) (*UEContext, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	ueContext, exists := s.contexts[supi]
	if !exists {
		return nil, fmt.Errorf("UE context not found for SUPI: %s", supi)
	}

	return ueContext, nil
}

// GetStats returns UECM statistics
func (s *UECMService) GetStats() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return map[string]interface{}{
		"total_contexts":      len(s.contexts),
		"registered_contexts": s.countRegistered(),
	}
}

func (s *UECMService) countRegistered() int {
	count := 0
	for _, ctx := range s.contexts {
		if ctx.AMFInstanceID != "" {
			count++
		}
	}
	return count
}
