package repository

import (
	"encoding/json"
	"time"
)

// SubscriberData represents complete subscriber information (TS 29.505)
type SubscriberData struct {
	SUPI     string `json:"supi"`
	SUPIType string `json:"supiType"` // "imsi" or "nai"

	// PLMN
	PLMNIDmcc string `json:"plmnId.mcc"`
	PLMNIDmnc string `json:"plmnId.mnc"`

	// Status
	SubscriberStatus string `json:"gpsis,omitempty"` // ACTIVE, INACTIVE, SUSPENDED
	MSISDN           string `json:"msisdn,omitempty"`

	// UE-AMBR (Aggregate Maximum Bit Rate)
	SubscribedUeAmbrUplink   uint64 `json:"subscribedUeAmbr.uplink,string"`
	SubscribedUeAmbrDownlink uint64 `json:"subscribedUeAmbr.downlink,string"`

	// Network Slicing
	NSSAI              []SNSSAI `json:"nssai,omitempty"`
	DefaultSingleNSSAI *SNSSAI  `json:"defaultSingleNssai,omitempty"`

	// DNN Configurations
	DNNConfigurations map[string]*DNNConfiguration `json:"dnnConfigurations,omitempty"`

	// Roaming
	RoamingAllowed bool     `json:"roamingAllowed"`
	RoamingAreas   []string `json:"roamingAreas,omitempty"`

	// Security
	OPCKey               string `json:"opcKey,omitempty"`
	AuthenticationMethod string `json:"authenticationMethod,omitempty"`

	// Metadata
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// SNSSAI represents Single Network Slice Selection Assistance Information
type SNSSAI struct {
	SST int    `json:"sst"`          // Slice/Service Type (0-255)
	SD  string `json:"sd,omitempty"` // Slice Differentiator
}

// DNNConfiguration represents DNN-specific configuration
type DNNConfiguration struct {
	PDUSessionTypes     []string `json:"pduSessionTypes"`
	SscModes            []int    `json:"sscModes"`
	IwkEpsInd           bool     `json:"iwkEpsInd,omitempty"`
	SessionAMBRUplink   uint64   `json:"sessionAmbr.uplink,string"`
	SessionAMBRDownlink uint64   `json:"sessionAmbr.downlink,string"`
	FiveQI              int      `json:"5qi"`
	StaticIPAddress     string   `json:"staticIpAddress,omitempty"`
	StaticIPv6Prefix    string   `json:"staticIpv6Prefix,omitempty"`
}

// AuthenticationSubscription represents authentication subscription data (TS 29.503)
type AuthenticationSubscription struct {
	SUPI                 string `json:"supi"`
	AuthenticationMethod string `json:"authenticationMethod"` // 5G_AKA, EAP_AKA_PRIME

	// Permanent Key
	PermanentKey   string `json:"permanentKey,omitempty"` // K (hex encoded)
	PermanentKeyID uint8  `json:"permanentKeyId,omitempty"`

	// Algorithm
	EncAlgorithm string `json:"encAlgorithm,omitempty"` // milenage, tuak
	EncOPC       string `json:"encOpc,omitempty"`       // OPc
	EncOP        string `json:"encTopcKey,omitempty"`   // OP

	// SQN (Sequence Number)
	SQN       uint64 `json:"sequenceNumber,string"`
	SQNScheme string `json:"sqnScheme,omitempty"`

	// AMF (Authentication Management Field)
	AuthenticationManagementField string `json:"authenticationManagementField,omitempty"`

	// Metadata
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// SessionManagementSubscriptionData represents SM subscription data
type SessionManagementSubscriptionData struct {
	SUPI string `json:"supi"`
	DNN  string `json:"dnn"`

	// Session AMBR
	SessionAMBRUplink   uint64 `json:"sessionAmbr.uplink,string"`
	SessionAMBRDownlink uint64 `json:"sessionAmbr.downlink,string"`

	// QoS
	Default5QI       int `json:"default5qi"`
	ARPPriorityLevel int `json:"arpPriorityLevel"`

	// SSC Mode
	SSCModes       []int `json:"allowedSscModes"`
	DefaultSSCMode int   `json:"defaultSscMode"`

	// PDU Session Type
	PDUSessionTypes       []string `json:"pduSessionTypes"`
	DefaultPDUSessionType string   `json:"defaultPduSessionType"`

	// Static IP
	StaticIPAddress  string `json:"staticIpAddress,omitempty"`
	StaticIPv6Prefix string `json:"staticIpv6Prefix,omitempty"`

	// Charging
	ChargingCharacteristics string `json:"chargingCharacteristics,omitempty"`

	// Metadata
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// SDMSubscription represents a subscription for data change notifications
type SDMSubscription struct {
	SubscriptionID        string    `json:"subscriptionId"`
	NFInstanceID          string    `json:"nfInstanceId"`
	CallbackURI           string    `json:"callbackReference"`
	MonitoredResourceURIs []string  `json:"monitoredResourceUris"`
	SingleNSSAI           *SNSSAI   `json:"singleNssai,omitempty"`
	DNN                   string    `json:"dnn,omitempty"`
	Expiry                time.Time `json:"expires,omitempty"`
	CreatedAt             time.Time `json:"createdAt"`
}

// PolicyData represents policy data for a subscriber
type PolicyData struct {
	SUPI                 string          `json:"supi"`
	SubscriberPolicies   json.RawMessage `json:"subscriberPolicies,omitempty"`
	SubscribedDefaultQoS json.RawMessage `json:"subscribedDefaultQos,omitempty"`
	CreatedAt            time.Time       `json:"createdAt"`
	UpdatedAt            time.Time       `json:"updatedAt"`
}

// AuthenticationVector represents a 5G authentication vector
type AuthenticationVector struct {
	RAND     string `json:"rand"`     // Random challenge (128 bits, hex)
	AUTN     string `json:"autn"`     // Authentication token (128 bits, hex)
	XRES     string `json:"xres"`     // Expected response (64-128 bits, hex)
	XRESStar string `json:"xresStar"` // XRES* for 5G (128 bits, hex)
	KAUSF    string `json:"kausf"`    // Key for AUSF (256 bits, hex)
}

// ConfirmationData represents 5G-AKA confirmation data
type ConfirmationData struct {
	KSEAF string `json:"kseaf"` // Session key (256 bits, hex)
}

// AuthEvent represents an authentication event for auditing
type AuthEvent struct {
	SUPI           string    `json:"supi"`
	Success        bool      `json:"success"`
	AuthMethod     string    `json:"authMethod"`
	ServingNetwork string    `json:"servingNetwork"`
	Timestamp      time.Time `json:"timestamp"`
	FailureReason  string    `json:"failureReason,omitempty"`
}

// MarshalJSON custom marshaling for SNSSAI arrays
func (s *SubscriberData) MarshalNSSAI() (string, error) {
	if len(s.NSSAI) == 0 {
		return "[]", nil
	}
	data, err := json.Marshal(s.NSSAI)
	return string(data), err
}

// UnmarshalNSSAI custom unmarshaling for SNSSAI arrays
func (s *SubscriberData) UnmarshalNSSAI(data string) error {
	if data == "" || data == "[]" {
		s.NSSAI = []SNSSAI{}
		return nil
	}
	return json.Unmarshal([]byte(data), &s.NSSAI)
}

// MarshalDNNConfigurations marshals DNN configurations to JSON string
func (s *SubscriberData) MarshalDNNConfigurations() (string, error) {
	if len(s.DNNConfigurations) == 0 {
		return "{}", nil
	}
	data, err := json.Marshal(s.DNNConfigurations)
	return string(data), err
}

// UnmarshalDNNConfigurations unmarshals DNN configurations from JSON string
func (s *SubscriberData) UnmarshalDNNConfigurations(data string) error {
	if data == "" || data == "{}" {
		s.DNNConfigurations = make(map[string]*DNNConfiguration)
		return nil
	}
	return json.Unmarshal([]byte(data), &s.DNNConfigurations)
}
