package f1

import (
	"context"
	"net"
)

// F1AP message types (3GPP TS 38.473)
const (
	F1AP_RESET                            = 0
	F1AP_RESET_ACKNOWLEDGE                = 1
	F1AP_F1_SETUP_REQUEST                 = 2
	F1AP_F1_SETUP_RESPONSE                = 3
	F1AP_F1_SETUP_FAILURE                 = 4
	F1AP_GNB_DU_CONFIGURATION_UPDATE      = 5
	F1AP_GNB_DU_CONFIGURATION_UPDATE_ACK  = 6
	F1AP_GNB_CU_CONFIGURATION_UPDATE      = 7
	F1AP_GNB_CU_CONFIGURATION_UPDATE_ACK  = 8
	F1AP_UE_CONTEXT_SETUP_REQUEST         = 9
	F1AP_UE_CONTEXT_SETUP_RESPONSE        = 10
	F1AP_UE_CONTEXT_SETUP_FAILURE         = 11
	F1AP_UE_CONTEXT_RELEASE_COMMAND       = 12
	F1AP_UE_CONTEXT_RELEASE_COMPLETE      = 13
	F1AP_UE_CONTEXT_MODIFICATION_REQUEST  = 14
	F1AP_UE_CONTEXT_MODIFICATION_RESPONSE = 15
	F1AP_UE_CONTEXT_MODIFICATION_REQUIRED = 16
	F1AP_INITIAL_UL_RRC_MESSAGE_TRANSFER  = 17
	F1AP_DL_RRC_MESSAGE_TRANSFER          = 18
	F1AP_UL_RRC_MESSAGE_TRANSFER          = 19
)

// F1Interface defines the F1 interface between CU and DU
type F1Interface interface {
	// Setup procedures
	SendF1SetupRequest(ctx context.Context, req *F1SetupRequest) (*F1SetupResponse, error)
	SendF1SetupResponse(ctx context.Context, resp *F1SetupResponse) error

	// UE Context Management
	SendUEContextSetupRequest(ctx context.Context, req *UEContextSetupRequest) (*UEContextSetupResponse, error)
	SendUEContextReleaseCommand(ctx context.Context, cmd *UEContextReleaseCommand) error
	SendUEContextModificationRequest(ctx context.Context, req *UEContextModificationRequest) (*UEContextModificationResponse, error)

	// RRC Message Transfer
	SendInitialULRRCMessageTransfer(ctx context.Context, msg *InitialULRRCMessage) error
	SendDLRRCMessageTransfer(ctx context.Context, msg *DLRRCMessage) error
	SendULRRCMessageTransfer(ctx context.Context, msg *ULRRCMessage) error

	// Configuration Update
	SendDUConfigurationUpdate(ctx context.Context, update *DUConfigurationUpdate) error
	SendCUConfigurationUpdate(ctx context.Context, update *CUConfigurationUpdate) error
}

// F1SetupRequest - DU -> CU
type F1SetupRequest struct {
	TransactionID    uint8
	GNBDUID          uint64
	GNBDUName        string
	ServedCellsToAdd []*ServedCell
	GNBDURRCVersion  *RRCVersion
}

// F1SetupResponse - CU -> DU
type F1SetupResponse struct {
	TransactionID   uint8
	GNBCUNAME       string
	CellsToActivate []*CellToActivate
	GNBCURRCVersion *RRCVersion
}

// ServedCell information
type ServedCell struct {
	ServedCellIndex uint8
	ServedCellInfo  *ServedCellInfo
	GNBDUSYSINFO    *SystemInformation
}

// ServedCellInfo contains cell configuration
type ServedCellInfo struct {
	NRCGI                          *NRCGI
	NRPCI                          uint16 // NR Physical Cell ID
	FiveGSTAC                      []byte // 5GS Tracking Area Code
	ConfiguredEPS_TAC              []byte
	ServedPLMNs                    []*ServedPLMN
	NRModeInfo                     *NRModeInfo
	MeasurementTimingConfiguration []byte
}

// NRCGI (NR Cell Global Identifier)
type NRCGI struct {
	PLMNID   *PLMNID
	NRCellID uint64 // 36 bits
}

// PLMNID
type PLMNID struct {
	MCC string // Mobile Country Code (3 digits)
	MNC string // Mobile Network Code (2-3 digits)
}

// ServedPLMN
type ServedPLMN struct {
	PLMNID           *PLMNID
	SliceSupportList []*SliceSupport
}

// SliceSupport (S-NSSAI)
type SliceSupport struct {
	SST uint8  // Slice/Service Type
	SD  []byte // Slice Differentiator (3 bytes)
}

// NRModeInfo (FDD or TDD)
type NRModeInfo struct {
	FDD *FDDInfo
	TDD *TDDInfo
}

// FDDInfo
type FDDInfo struct {
	ULARFCN                 uint32
	DLARFCN                 uint32
	ULTransmissionBandwidth uint16
	DLTransmissionBandwidth uint16
}

// TDDInfo
type TDDInfo struct {
	NRARFCN               uint32
	TransmissionBandwidth uint16
}

// SystemInformation
type SystemInformation struct {
	SIB1 []byte // System Information Block 1
}

// CellToActivate
type CellToActivate struct {
	NRCGI *NRCGI
}

// RRCVersion
type RRCVersion struct {
	Latest   []byte
	Extended []byte
}

// UEContextSetupRequest - CU -> DU
type UEContextSetupRequest struct {
	GNBCUUEF1APID uint32
	GNBDUUEF1APID uint32 // Optional
	SpCell        *SpCell
	SRBsToBeSetup []*SRBToBeSetup
	DRBsToBeSetup []*DRBToBeSetup
	CUtoDURRCInfo *CUtoDURRCInformation
}

// UEContextSetupResponse - DU -> CU
type UEContextSetupResponse struct {
	GNBCUUEF1APID     uint32
	GNBDUUEF1APID     uint32
	DUtoCURRCInfo     *DUtoCURRCInformation
	CellstoActivate   []*CellsActivated
	SRBsSetup         []*SRBSetup
	DRBsSetup         []*DRBSetup
	SRBsFailedToSetup []*SRBFailedToSetup
	DRBsFailedToSetup []*DRBFailedToSetup
}

// UEContextReleaseCommand - CU -> DU
type UEContextReleaseCommand struct {
	GNBCUUEF1APID uint32
	GNBDUUEF1APID uint32
	Cause         *Cause
	RRCContainer  []byte
}

// UEContextModificationRequest - CU -> DU
type UEContextModificationRequest struct {
	GNBCUUEF1APID    uint32
	GNBDUUEF1APID    uint32
	SRBsToBeSetup    []*SRBToBeSetup
	DRBsToBeSetup    []*DRBToBeSetup
	DRBsToBeModified []*DRBToBeModified
	DRBsToBeReleased []uint8
}

// UEContextModificationResponse - DU -> CU
type UEContextModificationResponse struct {
	GNBCUUEF1APID      uint32
	GNBDUUEF1APID      uint32
	DRBsModified       []*DRBModified
	DRBsFailedToModify []*DRBFailedToModify
}

// SpCell (Special Cell)
type SpCell struct {
	ServCellIndex uint8
	ServCellID    *NRCGI
	ServCellULCfg *CellULConfiguration
}

// CellULConfiguration
type CellULConfiguration struct {
	CellULConfigured bool
}

// SRBToBeSetup (Signaling Radio Bearer)
type SRBToBeSetup struct {
	SRBID                 uint8 // 1, 2, or 3
	DuplicationIndication bool
}

// DRBToBeSetup (Data Radio Bearer)
type DRBToBeSetup struct {
	DRBID                 uint8
	QoSInfo               *QoSFlowLevelQoSParameters
	ULUPTNLInfo           []*UPTransportLayerInformation
	RLCMode               string // "AM", "UM", "TM"
	ULConfiguration       *ULConfiguration
	DuplicationIndication bool
}

// DRBToBeModified
type DRBToBeModified struct {
	DRBID       uint8
	QoSInfo     *QoSFlowLevelQoSParameters
	ULUPTNLInfo []*UPTransportLayerInformation
}

// SRBSetup
type SRBSetup struct {
	SRBID uint8
}

// DRBSetup
type DRBSetup struct {
	DRBID       uint8
	DLUPTNLInfo []*UPTransportLayerInformation
}

// DRBModified
type DRBModified struct {
	DRBID       uint8
	DLUPTNLInfo []*UPTransportLayerInformation
}

// SRBFailedToSetup
type SRBFailedToSetup struct {
	SRBID uint8
	Cause *Cause
}

// DRBFailedToSetup
type DRBFailedToSetup struct {
	DRBID uint8
	Cause *Cause
}

// DRBFailedToModify
type DRBFailedToModify struct {
	DRBID uint8
	Cause *Cause
}

// QoSFlowLevelQoSParameters
type QoSFlowLevelQoSParameters struct {
	QoSCharacteristics               *QoSCharacteristics
	NGRANAllocationRetentionPriority *AllocationRetentionPriority
	GBRQoSFlowInfo                   *GBRQoSFlowInformation
	ReflectiveQoSAttribute           bool
}

// QoSCharacteristics
type QoSCharacteristics struct {
	NonDynamic5QI *NonDynamic5QIDescriptor
	Dynamic5QI    *Dynamic5QIDescriptor
}

// NonDynamic5QIDescriptor
type NonDynamic5QIDescriptor struct {
	FiveQI             uint8
	QoSPriorityLevel   uint8
	AveragingWindow    uint16
	MaxDataBurstVolume uint32
}

// Dynamic5QIDescriptor
type Dynamic5QIDescriptor struct {
	QoSPriorityLevel   uint8
	PacketDelayBudget  uint16
	PacketErrorRate    *PacketErrorRate
	AveragingWindow    uint16
	MaxDataBurstVolume uint32
}

// PacketErrorRate
type PacketErrorRate struct {
	Scalar   uint8
	Exponent uint8
}

// AllocationRetentionPriority
type AllocationRetentionPriority struct {
	PriorityLevel           uint8
	PreemptionCapability    string // "SHALL_NOT_TRIGGER_PREEMPTION", "MAY_TRIGGER_PREEMPTION"
	PreemptionVulnerability string // "NOT_PREEMPTABLE", "PREEMPTABLE"
}

// GBRQoSFlowInformation
type GBRQoSFlowInformation struct {
	MaxFlowBitRateDL        uint64
	MaxFlowBitRateUL        uint64
	GuaranteedFlowBitRateDL uint64
	GuaranteedFlowBitRateUL uint64
	MaxPacketLossRateDL     uint16
	MaxPacketLossRateUL     uint16
}

// UPTransportLayerInformation (GTP-U tunnel info)
type UPTransportLayerInformation struct {
	GTPTunnel *GTPTunnel
}

// GTPTunnel
type GTPTunnel struct {
	TransportLayerAddress net.IP
	GTPTEID               uint32
}

// ULConfiguration
type ULConfiguration struct {
	ULUEConfiguration string // "NO_DATA", "SHARED", "ONLY"
}

// CUtoDURRCInformation
type CUtoDURRCInformation struct {
	CGConfigInfo    []byte
	UECapabilityRAT []byte
	MeasConfig      []byte
}

// DUtoCURRCInformation
type DUtoCURRCInformation struct {
	CellGroupConfig   []byte
	MeasGapConfig     []byte
	RequestedP_MaxFR1 uint8
}

// CellsActivated
type CellsActivated struct {
	NRCGI *NRCGI
}

// InitialULRRCMessage - DU -> CU
type InitialULRRCMessage struct {
	GNBDUUEF1APID      uint32
	NRCGI              *NRCGI
	CRNTI              uint16 // Cell Radio Network Temporary Identifier
	RRCContainer       []byte // RRC Setup Request
	DUtoCURRCContainer []byte
}

// DLRRCMessage - CU -> DU
type DLRRCMessage struct {
	GNBCUUEF1APID uint32
	GNBDUUEF1APID uint32
	SRBID         uint8
	RRCContainer  []byte
}

// ULRRCMessage - DU -> CU
type ULRRCMessage struct {
	GNBCUUEF1APID uint32
	GNBDUUEF1APID uint32
	SRBID         uint8
	RRCContainer  []byte
}

// DUConfigurationUpdate - DU -> CU
type DUConfigurationUpdate struct {
	TransactionID       uint8
	ServedCellsToAdd    []*ServedCell
	ServedCellsToModify []*ServedCell
	ServedCellsToDelete []*NRCGI
}

// CUConfigurationUpdate - CU -> DU
type CUConfigurationUpdate struct {
	TransactionID     uint8
	CellsToActivate   []*CellToActivate
	CellsToDeactivate []*NRCGI
}

// Cause
type Cause struct {
	RadioNetwork *CauseRadioNetwork
	Transport    *CauseTransport
	Protocol     *CauseProtocol
	Misc         *CauseMisc
}

// CauseRadioNetwork
type CauseRadioNetwork struct {
	Value string
}

// CauseTransport
type CauseTransport struct {
	Value string
}

// CauseProtocol
type CauseProtocol struct {
	Value string
}

// CauseMisc
type CauseMisc struct {
	Value string
}
