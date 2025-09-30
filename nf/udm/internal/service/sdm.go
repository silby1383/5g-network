package service

import (
	"context"
	"fmt"

	"github.com/your-org/5g-network/nf/udm/internal/client"
	"go.uber.org/zap"
)

// SDMService handles Subscriber Data Management (Nudm_SDM)
type SDMService struct {
	udrClient *client.UDRClient
	logger    *zap.Logger
}

// NewSDMService creates a new SDM service
func NewSDMService(udrClient *client.UDRClient, logger *zap.Logger) *SDMService {
	return &SDMService{
		udrClient: udrClient,
		logger:    logger,
	}
}

// AccessAndMobilitySubscriptionData represents AM subscription data (TS 29.503)
type AccessAndMobilitySubscriptionData struct {
	GPSIS                  []string                `json:"gpsis,omitempty"`
	SubscribedUeAMBR       *AMBR                   `json:"subscribedUeAmbr,omitempty"`
	NSSAI                  *NSSAI                  `json:"nssai,omitempty"`
	RatRestrictions        []string                `json:"ratRestrictions,omitempty"`
	ForbiddenAreas         []interface{}           `json:"forbiddenAreas,omitempty"`
	ServiceAreaRestriction *ServiceAreaRestriction `json:"serviceAreaRestriction,omitempty"`
}

// AMBR represents Aggregate Maximum Bit Rate
type AMBR struct {
	Uplink   string `json:"uplink"`   // e.g., "1000000000" (1 Gbps)
	Downlink string `json:"downlink"` // e.g., "2000000000" (2 Gbps)
}

// NSSAI represents Network Slice Selection Assistance Information
type NSSAI struct {
	DefaultSingleNSSAIs []client.SNSSAI `json:"defaultSingleNssais,omitempty"`
	SingleNSSAIs        []client.SNSSAI `json:"singleNssais,omitempty"`
}

// ServiceAreaRestriction represents service area restrictions
type ServiceAreaRestriction struct {
	RestrictionType string        `json:"restrictionType,omitempty"`
	Areas           []interface{} `json:"areas,omitempty"`
}

// SessionManagementSubscriptionData represents SM subscription data (TS 29.503)
type SessionManagementSubscriptionData struct {
	SingleNSSAI       client.SNSSAI                `json:"singleNssai"`
	DnnConfigurations map[string]*DnnConfiguration `json:"dnnConfigurations,omitempty"`
}

// DnnConfiguration represents DNN configuration
type DnnConfiguration struct {
	PduSessionTypes *PduSessionTypes `json:"pduSessionTypes,omitempty"`
	SscModes        *SscModes        `json:"sscModes,omitempty"`
	SessionAMBR     *AMBR            `json:"sessionAmbr,omitempty"`
	Var5gQosProfile *Var5gQosProfile `json:"5gQosProfile,omitempty"`
	StaticIPAddress []string         `json:"staticIpAddress,omitempty"`
}

// PduSessionTypes represents PDU session types
type PduSessionTypes struct {
	DefaultSessionType  string   `json:"defaultSessionType"`
	AllowedSessionTypes []string `json:"allowedSessionTypes,omitempty"`
}

// SscModes represents SSC modes
type SscModes struct {
	DefaultSscMode  string   `json:"defaultSscMode"`
	AllowedSscModes []string `json:"allowedSscModes,omitempty"`
}

// Var5gQosProfile represents 5G QoS profile
type Var5gQosProfile struct {
	Var5qi        int  `json:"5qi"`
	PriorityLevel int  `json:"priorityLevel,omitempty"`
	ARP           *ARP `json:"arp,omitempty"`
}

// ARP represents Allocation and Retention Priority
type ARP struct {
	PriorityLevel int    `json:"priorityLevel"`
	PreemptCap    string `json:"preemptCap,omitempty"`
	PreemptVuln   string `json:"preemptVuln,omitempty"`
}

// GetAMData retrieves Access and Mobility subscription data
func (s *SDMService) GetAMData(ctx context.Context, supi string, plmnID *client.PLMNID) (*AccessAndMobilitySubscriptionData, error) {
	s.logger.Info("Getting AM subscription data",
		zap.String("supi", supi),
	)

	// Get subscriber data from UDR
	subData, err := s.udrClient.GetSubscriberData(ctx, supi)
	if err != nil {
		return nil, fmt.Errorf("failed to get subscriber data: %w", err)
	}

	// Convert to AM subscription data
	amData := &AccessAndMobilitySubscriptionData{
		SubscribedUeAMBR: &AMBR{
			Uplink:   fmt.Sprintf("%d", subData.SubscribedUeAmbrUplink),
			Downlink: fmt.Sprintf("%d", subData.SubscribedUeAmbrDownlink),
		},
	}

	// Add NSSAI if available
	if len(subData.NSSAI) > 0 {
		amData.NSSAI = &NSSAI{
			SingleNSSAIs: subData.NSSAI,
		}
		if len(subData.NSSAI) > 0 {
			amData.NSSAI.DefaultSingleNSSAIs = []client.SNSSAI{subData.NSSAI[0]}
		}
	}

	s.logger.Debug("Retrieved AM subscription data",
		zap.String("supi", supi),
	)

	return amData, nil
}

// GetSMData retrieves Session Management subscription data
func (s *SDMService) GetSMData(ctx context.Context, supi string, plmnID *client.PLMNID, dnn string) (*SessionManagementSubscriptionData, error) {
	s.logger.Info("Getting SM subscription data",
		zap.String("supi", supi),
		zap.String("dnn", dnn),
	)

	// Get subscriber data from UDR
	subData, err := s.udrClient.GetSubscriberData(ctx, supi)
	if err != nil {
		return nil, fmt.Errorf("failed to get subscriber data: %w", err)
	}

	// Get SM data from UDR if DNN is specified
	var smData *client.SessionManagementSubscriptionData
	if dnn != "" {
		smData, err = s.udrClient.GetSessionManagementData(ctx, supi, dnn)
		if err != nil {
			s.logger.Warn("Failed to get SM data from UDR, using defaults",
				zap.String("supi", supi),
				zap.String("dnn", dnn),
				zap.Error(err),
			)
		}
	}

	// Build SM subscription data
	smSubData := &SessionManagementSubscriptionData{
		DnnConfigurations: make(map[string]*DnnConfiguration),
	}

	// Add default S-NSSAI if available
	if len(subData.NSSAI) > 0 {
		smSubData.SingleNSSAI = subData.NSSAI[0]
	}

	// Add DNN configuration
	dnnConfig := &DnnConfiguration{
		PduSessionTypes: &PduSessionTypes{
			DefaultSessionType:  "IPV4",
			AllowedSessionTypes: []string{"IPV4", "IPV6", "IPV4V6"},
		},
		SscModes: &SscModes{
			DefaultSscMode:  "SSC_MODE_1",
			AllowedSscModes: []string{"SSC_MODE_1", "SSC_MODE_2", "SSC_MODE_3"},
		},
		SessionAMBR: &AMBR{
			Uplink:   fmt.Sprintf("%d", subData.SubscribedUeAmbrUplink),
			Downlink: fmt.Sprintf("%d", subData.SubscribedUeAmbrDownlink),
		},
		Var5gQosProfile: &Var5gQosProfile{
			Var5qi:        9, // Default 5QI for internet
			PriorityLevel: 8,
			ARP: &ARP{
				PriorityLevel: 8,
				PreemptCap:    "NOT_PREEMPT",
				PreemptVuln:   "NOT_PREEMPTABLE",
			},
		},
	}

	// Override with UDR data if available
	if smData != nil {
		if smData.SessionAmbrUplink > 0 {
			dnnConfig.SessionAMBR.Uplink = fmt.Sprintf("%d", smData.SessionAmbrUplink)
		}
		if smData.SessionAmbrDownlink > 0 {
			dnnConfig.SessionAMBR.Downlink = fmt.Sprintf("%d", smData.SessionAmbrDownlink)
		}
		if smData.Default5QI > 0 {
			dnnConfig.Var5gQosProfile.Var5qi = int(smData.Default5QI)
		}
		if smData.DefaultPDUSessionType != "" {
			dnnConfig.PduSessionTypes.DefaultSessionType = smData.DefaultPDUSessionType
		}
	}

	if dnn == "" {
		dnn = "internet" // Default DNN
	}
	smSubData.DnnConfigurations[dnn] = dnnConfig

	s.logger.Debug("Retrieved SM subscription data",
		zap.String("supi", supi),
		zap.String("dnn", dnn),
	)

	return smSubData, nil
}

// SubscribeToDataChanges subscribes to data change notifications
func (s *SDMService) SubscribeToDataChanges(ctx context.Context, supi string, callbackURI string) (string, error) {
	s.logger.Info("Creating SDM subscription",
		zap.String("supi", supi),
		zap.String("callback_uri", callbackURI),
	)

	// In production, create subscription in UDR
	subscriptionID := fmt.Sprintf("sdm-sub-%s", supi)

	return subscriptionID, nil
}

// UnsubscribeFromDataChanges unsubscribes from data change notifications
func (s *SDMService) UnsubscribeFromDataChanges(ctx context.Context, subscriptionID string) error {
	s.logger.Info("Deleting SDM subscription",
		zap.String("subscription_id", subscriptionID),
	)

	// In production, delete subscription from UDR
	return nil
}
