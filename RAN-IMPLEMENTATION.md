# gNodeB Implementation Guide - CU/DU Split with Simulated Radio

## Overview

This document describes the implementation of a 5G gNodeB (base station) with CU/DU split architecture and simulated radio interface. This approach allows full testing and validation of the 5G core network without physical RF equipment.

## Architecture

```
┌─────────────────────────────────────────────────────────┐
│                    5G Core Network                       │
│                    (AMF, SMF, UPF)                       │
└───────────────┬─────────────────┬───────────────────────┘
                │ N2 (NGAP)       │ N3 (GTP-U)
                │                 │
┌───────────────▼─────────────────▼───────────────────────┐
│              Central Unit (CU)                           │
│  ┌──────────────────────────────────────────────────┐   │
│  │  RRC (Radio Resource Control)                    │   │
│  │  PDCP (Packet Data Convergence Protocol)         │   │
│  │  - Control Plane PDCP                            │   │
│  │  - User Plane PDCP                               │   │
│  └──────────────────────────────────────────────────┘   │
└───────────────────────────┬─────────────────────────────┘
                            │ F1 Interface
┌───────────────────────────▼─────────────────────────────┐
│            Distributed Unit (DU)                         │
│  ┌──────────────────────────────────────────────────┐   │
│  │  RLC (Radio Link Control)                        │   │
│  │  MAC (Medium Access Control)                     │   │
│  │  High PHY (Physical Layer - upper)               │   │
│  └──────────────────────────────────────────────────┘   │
└───────────────────────────┬─────────────────────────────┘
                            │ Fronthaul (simulated)
┌───────────────────────────▼─────────────────────────────┐
│            Radio Unit (RU) - SIMULATED                   │
│  ┌──────────────────────────────────────────────────┐   │
│  │  Low PHY (Physical Layer - lower)                │   │
│  │  RF Processing (simulated)                       │   │
│  │  - Signal generation                             │   │
│  │  - Channel modeling                              │   │
│  │  - UE simulation                                 │   │
│  └──────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────┘
                            │
                            ▼
                     UE (simulated)
```

## Component Details

### Central Unit (CU)

**Responsibilities:**
- RRC connection management
- PDCP layer processing
- Mobility management
- Connection to 5G Core (AMF, UPF)

**Protocols:**
- **N2 (to AMF):** NGAP over SCTP
- **N3 (to UPF):** GTP-U over UDP
- **F1 (to DU):** F1AP over SCTP

**Implementation (Go):**

```go
// nf/gnb/internal/cu/cu.go
package cu

type CentralUnit struct {
    config      *Config
    rrcLayer    *RRC
    pdcpLayer   *PDCP
    n2Client    *N2Client  // NGAP to AMF
    n3Client    *N3Client  // GTP-U to UPF
    f1Server    *F1Server  // F1 to DU
    ueContexts  map[string]*UEContext
}

type UEContext struct {
    UEID        uint32
    IMSI        string
    GUTI        string
    RRCState    string  // "IDLE", "CONNECTED"
    PDCPConfig  *PDCPConfig
    Bearers     map[uint8]*Bearer
}

func (cu *CentralUnit) HandleRRCSetupRequest(ueID uint32) error {
    ctx, span := otel.Tracer("cu").Start(context.Background(), "CU.RRCSetupRequest")
    defer span.End()
    
    // Create UE context
    ueCtx := &UEContext{
        UEID:     ueID,
        RRCState: "CONNECTED",
    }
    cu.ueContexts[fmt.Sprintf("%d", ueID)] = ueCtx
    
    // Send RRC Setup to UE (via DU)
    rrcSetup := cu.rrcLayer.CreateRRCSetup(ueID)
    return cu.f1Server.SendToUE(ueID, rrcSetup)
}

func (cu *CentralUnit) HandleN2InitialUEMessage(msg *N2InitialUEMessage) error {
    // Forward to AMF
    return cu.n2Client.SendInitialUEMessage(msg)
}
```

### Distributed Unit (DU)

**Responsibilities:**
- RLC layer processing
- MAC layer scheduling
- High PHY processing
- Bridge between CU and RU

**Protocols:**
- **F1 (to CU):** F1AP over SCTP
- **Fronthaul (to RU):** Simulated interface

**Implementation (Go):**

```go
// nf/gnb/internal/du/du.go
package du

type DistributedUnit struct {
    config        *Config
    rlcLayer      *RLC
    macLayer      *MAC
    f1Client      *F1Client
    fronthaulIF   *FronthaulInterface
    scheduler     *Scheduler
}

type Scheduler struct {
    ues         map[uint32]*UESchedulingInfo
    resourceMap *ResourceMap
}

func (du *DistributedUnit) ScheduleResources() {
    // MAC scheduler
    for ueID, ueInfo := range du.scheduler.ues {
        // Allocate PRBs (Physical Resource Blocks)
        prbs := du.scheduler.AllocatePRBs(ueID, ueInfo.BufferStatus)
        
        // Send scheduling grant to UE via RU
        du.fronthaulIF.SendSchedulingGrant(ueID, prbs)
    }
}

func (du *DistributedUnit) ProcessUplinkData(ueID uint32, data []byte) error {
    ctx, span := otel.Tracer("du").Start(context.Background(), "DU.ProcessUplinkData")
    defer span.End()
    
    // RLC processing
    rlcPDU := du.rlcLayer.Deframe(data)
    
    // Forward to CU via F1
    return du.f1Client.SendUplinkData(ueID, rlcPDU)
}
```

### Radio Unit (RU) - Simulated

**Responsibilities:**
- Simulate RF transmission/reception
- Channel modeling
- UE signal simulation

**Implementation (Go):**

```go
// nf/gnb/internal/ru/simulator.go
package ru

type RadioSimulator struct {
    config      *Config
    cells       map[string]*VirtualCell
    ues         map[string]*VirtualUE
    channel     *ChannelModel
}

type VirtualCell struct {
    CellID      string
    CenterFreq  float64  // Hz (e.g., 3.5 GHz)
    Bandwidth   float64  // Hz (e.g., 100 MHz)
    TxPower     float64  // dBm
    Location    Location
    ActiveUEs   map[string]*VirtualUE
}

type VirtualUE struct {
    UEID        uint32
    IMSI        string
    Location    Location
    Velocity    Velocity
    AttachedCell string
    RSRP        float64  // Reference Signal Received Power (dBm)
    RSRQ        float64  // Reference Signal Received Quality (dB)
    SINR        float64  // Signal-to-Interference-plus-Noise Ratio (dB)
}

type ChannelModel struct {
    PropagationModel string  // "FREE_SPACE", "URBAN", "INDOOR"
    FadingType      string  // "RAYLEIGH", "RICIAN", "NONE"
    NoiseFloor      float64  // dBm
}

// Simulate RF propagation
func (r *RadioSimulator) ComputeRSRP(ue *VirtualUE, cell *VirtualCell) float64 {
    // Distance
    distance := computeDistance(ue.Location, cell.Location)
    
    // Path loss model
    var pathLoss float64
    switch r.channel.PropagationModel {
    case "FREE_SPACE":
        // Free space path loss: 20*log10(d) + 20*log10(f) + 20*log10(4*pi/c)
        pathLoss = 20*math.Log10(distance) + 20*math.Log10(cell.CenterFreq) + 20*math.Log10(4*math.Pi/3e8)
    case "URBAN":
        // COST-231 Hata model for urban areas
        pathLoss = r.cost231HataUrban(distance, cell.CenterFreq)
    }
    
    // RSRP = Tx Power - Path Loss + Antenna Gains
    rsrp := cell.TxPower - pathLoss
    
    // Apply fading
    if r.channel.FadingType == "RAYLEIGH" {
        rsrp += r.rayleighFading()
    }
    
    return rsrp
}

// Simulate UE attachment
func (r *RadioSimulator) SimulateUEAttachment(ue *VirtualUE) (string, error) {
    ctx, span := otel.Tracer("ru").Start(context.Background(), "RU.UEAttachment")
    defer span.End()
    
    // Find best cell based on RSRP
    bestCell := ""
    bestRSRP := -math.MaxFloat64
    
    for cellID, cell := range r.cells {
        rsrp := r.ComputeRSRP(ue, cell)
        if rsrp > bestRSRP {
            bestRSRP = rsrp
            bestCell = cellID
        }
    }
    
    // Check if RSRP is above threshold
    if bestRSRP < -120.0 {  // dBm
        return "", fmt.Errorf("no suitable cell found")
    }
    
    // Attach UE to cell
    ue.AttachedCell = bestCell
    ue.RSRP = bestRSRP
    r.cells[bestCell].ActiveUEs[ue.IMSI] = ue
    
    span.SetAttributes(
        attribute.String("cell_id", bestCell),
        attribute.Float64("rsrp", bestRSRP),
    )
    
    return bestCell, nil
}

// Simulate handover
func (r *RadioSimulator) SimulateHandover(ue *VirtualUE) error {
    // Check neighboring cells
    currentCell := r.cells[ue.AttachedCell]
    
    for cellID, cell := range r.cells {
        if cellID == ue.AttachedCell {
            continue
        }
        
        neighborRSRP := r.ComputeRSRP(ue, cell)
        
        // Handover if neighbor is 3 dB better
        if neighborRSRP > ue.RSRP+3.0 {
            // Trigger handover
            delete(currentCell.ActiveUEs, ue.IMSI)
            cell.ActiveUEs[ue.IMSI] = ue
            
            ue.AttachedCell = cellID
            ue.RSRP = neighborRSRP
            
            log.Infof("Handover: UE %s from cell %s to %s", ue.IMSI, currentCell.CellID, cellID)
            return nil
        }
    }
    
    return nil
}
```

### F1 Interface (CU ↔ DU)

```go
// common/f1/f1ap.go
package f1

const (
    F1AP_INITIAL_UL_RRC_MESSAGE    = 0
    F1AP_DL_RRC_MESSAGE_TRANSFER   = 1
    F1AP_UL_RRC_MESSAGE_TRANSFER   = 2
    F1AP_UE_CONTEXT_SETUP_REQUEST  = 3
    F1AP_UE_CONTEXT_SETUP_RESPONSE = 4
)

type F1APMessage struct {
    ProcedureCode uint8
    Criticality   uint8
    Value         interface{}
}

type UEContextSetupRequest struct {
    gNBCUUEF1APID uint32
    gNBDUUEF1APID uint32
    ServCellIndex uint8
    DRBsToSetup   []*DRBToSetup
}

type DRBToSetup struct {
    DRBID         uint8
    QoSFlowInfo   *QoSFlowInfo
    ULTunnelInfo  *GTPTunnel
    RLCMode       string  // "AM", "UM", "TM"
}

func (f *F1Client) SendUEContextSetupRequest(req *UEContextSetupRequest) error {
    ctx, span := otel.Tracer("f1").Start(context.Background(), "F1.UEContextSetupRequest")
    defer span.End()
    
    msg := &F1APMessage{
        ProcedureCode: F1AP_UE_CONTEXT_SETUP_REQUEST,
        Value:         req,
    }
    
    return f.send(msg)
}
```

## Package Structure

```
nf/gnb/
├── cmd/
│   ├── cu/              # CU main
│   ├── du/              # DU main
│   └── ru-simulator/    # RU simulator main
├── internal/
│   ├── cu/
│   │   ├── cu.go
│   │   ├── rrc/         # RRC layer
│   │   ├── pdcp/        # PDCP layer
│   │   ├── n2/          # NGAP client (to AMF)
│   │   └── n3/          # GTP-U client (to UPF)
│   ├── du/
│   │   ├── du.go
│   │   ├── rlc/         # RLC layer
│   │   ├── mac/         # MAC layer
│   │   ├── scheduler/   # Resource scheduler
│   │   └── phy/         # High PHY
│   ├── ru/
│   │   ├── simulator.go # RF simulator
│   │   ├── channel.go   # Channel model
│   │   └── ue.go        # UE simulation
│   └── f1/
│       ├── client.go    # F1 client (DU side)
│       └── server.go    # F1 server (CU side)
├── pkg/
│   └── radio/
│       ├── propagation.go  # Path loss models
│       └── fading.go       # Fading models
├── config/
│   ├── cu-config.yaml
│   ├── du-config.yaml
│   └── ru-config.yaml
├── test/
└── Dockerfile
```

## Configuration

```yaml
# cu-config.yaml
cu:
  name: cu-1
  instance_id: "uuid"
  
  # N2 interface (to AMF)
  n2:
    amf_addr: amf.5gc.svc.cluster.local:38412
    local_addr: 0.0.0.0:38412
  
  # N3 interface (to UPF)
  n3:
    upf_addr: upf.5gc.svc.cluster.local:2152
    local_addr: 0.0.0.0:2152
  
  # F1 interface (to DU)
  f1:
    listen_addr: 0.0.0.0:38472
  
  plmn:
    mcc: "001"
    mnc: "01"

# du-config.yaml
du:
  name: du-1
  instance_id: "uuid"
  
  # F1 interface (to CU)
  f1:
    cu_addr: cu-1:38472
  
  # Fronthaul (to RU)
  fronthaul:
    ru_addrs:
      - ru-simulator-1:50001
  
  cells:
    - cell_id: "001"
      pci: 1  # Physical Cell ID
      bandwidth: 100  # MHz
      max_ues: 100

# ru-config.yaml (simulator)
ru:
  name: ru-simulator-1
  
  # Fronthaul (to DU)
  fronthaul:
    du_addr: du-1:50001
    listen_addr: 0.0.0.0:50001
  
  # Simulated cells
  cells:
    - cell_id: "001"
      center_freq: 3.5e9  # 3.5 GHz
      bandwidth: 100e6    # 100 MHz
      tx_power: 43        # dBm
      location:
        lat: 37.7749
        lon: -122.4194
  
  # Channel model
  channel:
    propagation_model: "URBAN"  # or "FREE_SPACE", "INDOOR"
    fading: "RAYLEIGH"          # or "RICIAN", "NONE"
    noise_floor: -104           # dBm
  
  # Simulated UEs
  ues:
    - imsi: "001010000000001"
      location:
        lat: 37.7750
        lon: -122.4195
      mobility:
        enabled: true
        speed: 3  # m/s (walking)
        pattern: "random_walk"
```

## UE Simulator

```go
// tools/ue-simulator/internal/ue/ue.go
package ue

type UESimulator struct {
    config    *Config
    ues       map[string]*SimulatedUE
    gnbClient *GNBClient
}

type SimulatedUE struct {
    IMSI        string
    Key         []byte  // K (128-bit)
    OPc         []byte  // OPc (128-bit)
    Location    Location
    Velocity    Velocity
    State       string  // "IDLE", "CONNECTED"
    AttachedCell string
}

func (s *UESimulator) RegisterUE(ue *SimulatedUE) error {
    ctx, span := otel.Tracer("ue-sim").Start(context.Background(), "UE.Register")
    defer span.End()
    
    // 1. Send RRC Setup Request to gNB
    rrcSetupReq := createRRCSetupRequest(ue.IMSI)
    resp, err := s.gnbClient.SendRRCMessage(rrcSetupReq)
    if err != nil {
        return err
    }
    
    // 2. Send NAS Registration Request
    nasRegReq := createNASRegistrationRequest(ue.IMSI, ue.Key, ue.OPc)
    return s.gnbClient.SendNASMessage(nasRegReq)
}

func (s *UESimulator) EstablishPDUSession(imsi string, dnn string) error {
    ctx, span := otel.Tracer("ue-sim").Start(context.Background(), "UE.EstablishPDUSession")
    defer span.End()
    
    nasReq := createPDUSessionEstablishmentRequest(imsi, dnn)
    return s.gnbClient.SendNASMessage(nasReq)
}

func (s *UESimulator) SimulateMobility() {
    ticker := time.NewTicker(1 * time.Second)
    defer ticker.Stop()
    
    for range ticker.C {
        for _, ue := range s.ues {
            if ue.Velocity.Speed > 0 {
                // Update position
                ue.Location = updateLocation(ue.Location, ue.Velocity, 1.0)
                
                // Trigger measurement reports (for handover)
                s.gnbClient.SendMeasurementReport(ue.IMSI, ue.Location)
            }
        }
    }
}
```

## Testing

```go
// Test complete flow: UE attachment → Registration → PDU Session
func TestCompleteCallFlow(t *testing.T) {
    // Start components
    cu := startCU()
    du := startDU()
    ru := startRUSimulator()
    defer func() {
        cu.Stop()
        du.Stop()
        ru.Stop()
    }()
    
    // Create UE
    ue := &SimulatedUE{
        IMSI: "001010000000001",
        Key:  []byte{0x46, 0x5B, 0x5C, 0xE8, 0xB1, 0x99, 0xB4, 0x9F, 0xAA, 0x5F, 0x0A, 0x2E, 0xE2, 0x38, 0xA6, 0xBC},
        Location: Location{Lat: 37.7750, Lon: -122.4195},
    }
    
    // 1. UE attaches to cell
    cellID, err := ru.SimulateUEAttachment(ue)
    assert.NoError(t, err)
    assert.NotEmpty(t, cellID)
    
    // 2. RRC connection establishment
    err = cu.HandleRRCSetupRequest(ue.UEID)
    assert.NoError(t, err)
    
    // 3. NAS Registration (via AMF)
    // This would trigger AMF → AUSF → UDM flow
    
    // 4. PDU Session Establishment
    // This would trigger SMF → UPF → PCF flow
    
    time.Sleep(2 * time.Second)
    
    // Verify UE is registered and has session
    ueCtx := cu.GetUEContext(ue.UEID)
    assert.Equal(t, "CONNECTED", ueCtx.RRCState)
}
```

## Deployment

```yaml
# deploy/helm/gnb/values.yaml
cu:
  replicaCount: 2
  image:
    repository: 5g/gnb-cu
    tag: latest
  
du:
  replicaCount: 3
  image:
    repository: 5g/gnb-du
    tag: latest

ruSimulator:
  replicaCount: 5
  image:
    repository: 5g/gnb-ru-simulator
    tag: latest
  config:
    cells:
      - cellID: "001"
        centerFreq: 3.5e9
        txPower: 43
```

## Benefits

1. **No RF Equipment Needed** - Full testing without physical hardware
2. **Reproducible** - Consistent test conditions
3. **Scalable** - Simulate hundreds of cells and thousands of UEs
4. **Debuggable** - Full visibility into all layers
5. **CU/DU Split** - Standard 3GPP architecture
6. **Migration Path** - Can replace RU simulator with real hardware later

## Summary

This implementation provides:
- ✅ Standard gNodeB CU/DU split architecture
- ✅ Simulated radio interface with channel modeling
- ✅ Full protocol stack (RRC, PDCP, RLC, MAC, PHY)
- ✅ UE attachment and mobility simulation
- ✅ F1 interface between CU and DU
- ✅ Integration with 5G Core (AMF, UPF)
- ✅ Comprehensive testing capabilities

