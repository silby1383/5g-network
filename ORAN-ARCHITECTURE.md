# O-RAN Architecture Implementation Guide

## Overview

This document details the O-RAN (Open Radio Access Network) compliant architecture for the 5G network project, including simulated radio interfaces and RAN Intelligent Controller (RIC) implementation.

## O-RAN Alliance Standards

Our implementation follows O-RAN Alliance specifications:
- **O-RAN Architecture:** O-RAN.WG1.O-RAN-Architecture-Description
- **E2 Interface:** O-RAN.WG3.E2AP
- **A1 Interface:** O-RAN.WG2.A1-AP
- **O1 Interface:** O-RAN.WG1.O1-Interface
- **Fronthaul:** O-RAN.WG4.CUS (Control, User, and Synchronization planes)

## Architecture Components

```
┌─────────────────────────────────────────────────────────────────────┐
│                         Non-RT RIC                                   │
│              (Non-Real-Time RAN Intelligent Controller)             │
│  ┌────────────┐  ┌────────────┐  ┌────────────┐                    │
│  │   rApp:    │  │   rApp:    │  │   rApp:    │                    │
│  │  Traffic   │  │   Energy   │  │    QoS     │                    │
│  │ Steering   │  │  Savings   │  │Optimization│                    │
│  └────────────┘  └────────────┘  └────────────┘                    │
│         │                │                │                          │
│         └────────────────┴────────────────┘                          │
│                          │ A1                                        │
└──────────────────────────┼───────────────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────────────┐
│                        Near-RT RIC                                   │
│           (Near-Real-Time RAN Intelligent Controller)               │
│                      (<1 second latency)                            │
│  ┌────────────┐  ┌────────────┐  ┌────────────┐                    │
│  │   xApp:    │  │   xApp:    │  │   xApp:    │                    │
│  │   Traffic  │  │  Mobility  │  │   Anomaly  │                    │
│  │ Prediction │  │   Mgmt     │  │ Detection  │                    │
│  └────────────┘  └────────────┘  └────────────┘                    │
│         │                │                │                          │
│         └────────────────┴────────────────┘                          │
│                 E2 Service Models                                    │
│         ┌──────────┬──────────┬──────────┐                          │
│         │   KPM    │    RC    │    NI    │                          │
│         │(Metrics) │(Control) │ (Node    │                          │
│         │          │          │  Info)   │                          │
│         └──────────┴──────────┴──────────┘                          │
│                          │ E2                                        │
└──────────────────────────┼───────────────────────────────────────────┘
                           │
           ┌───────────────┴───────────────┐
           │                               │
           ▼                               ▼
┌───────────────────┐           ┌───────────────────┐
│       O-CU        │           │       O-DU        │
│  (Central Unit)   │   F1      │ (Distributed Unit)│
│  ┌─────┐ ┌─────┐ │◄─────────►│                   │
│  │CU-CP│ │CU-UP│ │           │   L2: MAC, RLC    │
│  │PDCP │ │PDCP │ │           │   L1-High         │
│  │ RRC │ │SDAP │ │           │                   │
│  └─────┘ └─────┘ │           └───────────────────┘
└───────────────────┘                     │
        │ N2                              │ Fronthaul (eCPRI)
        │ (to AMF)                        │
        │                                 ▼
        │                       ┌───────────────────┐
        │                       │       O-RU        │
        │ N3                    │   (Radio Unit)    │
        │ (to UPF)              │   **SIMULATED**   │
        ▼                       │                   │
                                │   RF Processing   │
                                │   L1-Low (PHY)    │
                                └───────────────────┘
                                         │
                                         ▼
                                    UE (simulated)
```

## Component Breakdown

### 1. O-RAN Central Unit (O-CU)

The O-CU is split into two logical functions:

#### CU-CP (Control Plane)
- **Responsibilities:**
  - RRC (Radio Resource Control)
  - PDCP (Packet Data Convergence Protocol) for control plane
  - Connection management
  - Mobility management
  - UE context management
  
- **Interfaces:**
  - N2: O-CU-CP ↔ AMF
  - E1: O-CU-CP ↔ O-CU-UP
  - F1-C: O-CU-CP ↔ O-DU (control plane)
  - E2: O-CU-CP ↔ Near-RT RIC

#### CU-UP (User Plane)
- **Responsibilities:**
  - SDAP (Service Data Adaptation Protocol)
  - PDCP (Packet Data Convergence Protocol) for user plane
  - QoS flow management
  - Header compression
  
- **Interfaces:**
  - N3: O-CU-UP ↔ UPF
  - E1: O-CU-UP ↔ O-CU-CP
  - F1-U: O-CU-UP ↔ O-DU (user plane)
  - E2: O-CU-UP ↔ Near-RT RIC

### 2. O-RAN Distributed Unit (O-DU)

- **Responsibilities:**
  - MAC (Medium Access Control)
  - RLC (Radio Link Control)
  - High PHY (Physical Layer - upper)
  - Scheduling decisions
  - HARQ (Hybrid ARQ)
  
- **Interfaces:**
  - F1: O-DU ↔ O-CU
  - Fronthaul: O-DU ↔ O-RU (eCPRI over Ethernet)
  - E2: O-DU ↔ Near-RT RIC
  - O1: O-DU ↔ SMO (Service Management and Orchestration)

### 3. O-RAN Radio Unit (O-RU) - SIMULATED

- **Responsibilities (in real deployment):**
  - Low PHY processing
  - RF transmission/reception
  - Beamforming
  - Digital-to-analog conversion
  
- **Simulation Approach:**
  - Simulate RF channel characteristics
  - Virtual UE attachment
  - Synthetic signal processing
  - Configurable propagation models
  
- **Interfaces:**
  - Fronthaul: O-RU ↔ O-DU (eCPRI)
  - O1: O-RU ↔ SMO

### 4. RAN Intelligent Controller (RIC)

#### Near-RT RIC (Near-Real-Time)
- **Latency Requirement:** 10ms - 1 second
- **Responsibilities:**
  - Real-time RAN optimization
  - Dynamic resource allocation
  - Interference management
  - Mobility optimization
  - Load balancing
  
- **xApps (Near-RT RIC Applications):**
  - Traffic Steering xApp
  - QoS Prediction xApp
  - Mobility Management xApp
  - Anomaly Detection xApp
  - Energy Savings xApp

- **E2 Service Models:**
  - **KPM (Key Performance Metrics):** Collect RAN metrics
  - **RC (RAN Control):** Control RAN behavior
  - **NI (Node Information):** Retrieve node information

#### Non-RT RIC (Non-Real-Time)
- **Latency Requirement:** >1 second
- **Responsibilities:**
  - Long-term optimization
  - ML model training
  - Policy management
  - Configuration management
  
- **rApps (Non-RT RIC Applications):**
  - Traffic Prediction rApp (ML-based)
  - Network Slicing rApp
  - Energy Optimization rApp
  - QoS Optimization rApp

## Open Interfaces

### E2 Interface (O-DU/O-CU ↔ Near-RT RIC)

**Purpose:** Enable near-real-time control and monitoring of RAN functions

**E2 Service Models:**

1. **E2SM-KPM (Key Performance Metrics)**
   - Collect RAN performance metrics
   - Examples: Throughput, PRB utilization, UE count, latency

2. **E2SM-RC (RAN Control)**
   - Control RAN behavior
   - Examples: Handover decisions, QoS adjustments, resource allocation

3. **E2SM-NI (Node Information)**
   - Retrieve RAN node information
   - Examples: Cell configuration, capabilities

**Implementation:**
```go
// E2 Interface - KPM Service Model
type E2Client struct {
    ricEndpoint string
    conn        *grpc.ClientConn
}

type KPMReport struct {
    Timestamp      time.Time
    CellID         string
    PRBUtilization float64
    ActiveUEs      uint32
    Throughput     uint64
    Latency        time.Duration
}

func (c *E2Client) SendKPMReport(report *KPMReport) error {
    // Send metrics to Near-RT RIC
    msg := &e2ap.E2APMessage{
        ProcedureCode: e2ap.PROCEDURE_CODE_RIC_INDICATION,
        ServiceModel:  e2sm.SERVICE_MODEL_KPM,
        Payload:       encodeKPMReport(report),
    }
    
    return c.send(msg)
}

// E2 Interface - RC Service Model
type RCControlRequest struct {
    TargetCellID   string
    ControlAction  string  // "HANDOVER", "QOS_UPDATE", "RESOURCE_ALLOCATION"
    Parameters     map[string]interface{}
}

func (c *E2Client) HandleRCControlRequest(req *RCControlRequest) error {
    switch req.ControlAction {
    case "HANDOVER":
        return c.triggerHandover(req.Parameters)
    case "QOS_UPDATE":
        return c.updateQoS(req.Parameters)
    case "RESOURCE_ALLOCATION":
        return c.allocateResources(req.Parameters)
    }
    return nil
}
```

### A1 Interface (Non-RT RIC ↔ Near-RT RIC)

**Purpose:** Policy-based guidance from Non-RT RIC to Near-RT RIC

**Implementation:**
```go
type A1Policy struct {
    PolicyID   string
    PolicyType string  // "QOS", "ENERGY_SAVINGS", "TRAFFIC_STEERING"
    Parameters map[string]interface{}
}

func (ric *NearRTRIC) ReceiveA1Policy(policy *A1Policy) error {
    // Apply policy from Non-RT RIC
    switch policy.PolicyType {
    case "QOS":
        return ric.applyQoSPolicy(policy.Parameters)
    case "ENERGY_SAVINGS":
        return ric.applyEnergySavingsPolicy(policy.Parameters)
    case "TRAFFIC_STEERING":
        return ric.applyTrafficSteeringPolicy(policy.Parameters)
    }
    return nil
}
```

### F1 Interface (O-DU ↔ O-CU)

**Purpose:** Split between distributed and centralized RAN processing

**F1-C (Control Plane):**
- UE context management
- RRC message transfer
- Paging

**F1-U (User Plane):**
- User data tunneling (GTP-U)

**Implementation:**
```go
// F1 Interface
type F1Client struct {
    cuEndpoint string
    conn       net.Conn
}

// F1-C: UE Context Setup Request
type F1UEContextSetupRequest struct {
    gNBDUID    uint32
    gNBCUUEF1APID uint32
    CRNTI      uint16
    ServCellID uint32
    DRBs       []*DRBToBeSetup
}

func (f *F1Client) SendUEContextSetupRequest(req *F1UEContextSetupRequest) error {
    msg := encodeF1APMessage(F1AP_UE_CONTEXT_SETUP_REQUEST, req)
    return f.send(msg)
}

// F1-U: User data transfer
func (f *F1Client) SendUserData(ueID uint32, data []byte) error {
    // Encapsulate in GTP-U and send via F1-U
    gtpPacket := encapsulateGTPU(ueID, data)
    return f.sendUserPlane(gtpPacket)
}
```

### Fronthaul Interface (O-DU ↔ O-RU)

**Purpose:** eCPRI (enhanced Common Public Radio Interface) over Ethernet

**Planes:**
- Control Plane: Configuration and management
- User Plane: IQ data transfer
- Synchronization Plane: Timing synchronization

**Simulation Approach:**
```go
// Simulated Fronthaul
type SimulatedFronthaul struct {
    odu  *ODU
    oru  *ORU
    channel chan *IQData
}

type IQData struct {
    Timestamp  time.Time
    Samples    []complex64  // I/Q samples
    CenterFreq float64
    Bandwidth  float64
}

func (f *SimulatedFronthaul) SendIQData(data *IQData) error {
    // Simulate propagation delay
    time.Sleep(100 * time.Microsecond)
    
    // Apply channel model (optional)
    processedData := f.applyChannelModel(data)
    
    f.channel <- processedData
    return nil
}

func (f *SimulatedFronthaul) applyChannelModel(data *IQData) *IQData {
    // Simulate fading, noise, etc.
    // For now, just pass through
    return data
}
```

## RIC xApps and rApps

### Near-RT RIC xApps

#### 1. Traffic Steering xApp
```go
// xapp/traffic_steering/main.go
package main

type TrafficSteeringXApp struct {
    e2Client *E2Client
    metrics  *MetricsCollector
}

func (x *TrafficSteeringXApp) Run() {
    ticker := time.NewTicker(100 * time.Millisecond)
    defer ticker.Stop()
    
    for range ticker.C {
        // Collect metrics from O-DU/O-CU
        metrics := x.metrics.GetCurrentMetrics()
        
        // Make steering decisions based on load
        if metrics.CellLoad > 0.8 {
            // Trigger handover for some UEs to neighbor cell
            x.triggerLoadBalancing(metrics.CellID)
        }
    }
}

func (x *TrafficSteeringXApp) triggerLoadBalancing(cellID string) error {
    // Send RC control message to O-DU
    req := &RCControlRequest{
        TargetCellID:  cellID,
        ControlAction: "HANDOVER",
        Parameters: map[string]interface{}{
            "target_cell": findLightestLoadedNeighbor(cellID),
            "ue_count":    10,  // Move 10 UEs
        },
    }
    
    return x.e2Client.SendRCControlRequest(req)
}
```

#### 2. QoS Prediction xApp
```go
type QoSPredictionXApp struct {
    e2Client *E2Client
    model    *MLModel
}

func (x *QoSPredictionXApp) Run() {
    // Subscribe to KPM reports
    x.e2Client.SubscribeKPM(func(report *KPMReport) {
        // Predict future QoS based on current metrics
        prediction := x.model.Predict(report)
        
        // If degradation predicted, take preemptive action
        if prediction.PredictedLatency > threshold {
            x.adjustQoS(report.CellID, prediction)
        }
    })
}
```

#### 3. Anomaly Detection xApp
```go
type AnomalyDetectionXApp struct {
    e2Client *E2Client
    detector *AnomalyDetector
}

func (x *AnomalyDetectionXApp) Run() {
    x.e2Client.SubscribeKPM(func(report *KPMReport) {
        // Detect anomalies in metrics
        if x.detector.IsAnomaly(report) {
            // Alert and potentially take corrective action
            x.handleAnomaly(report)
        }
    })
}
```

### Non-RT RIC rApps

#### 1. Traffic Prediction rApp (Python with ML)
```python
# rapp/traffic_prediction/main.py
import tensorflow as tf
import numpy as np
from a1_client import A1Client

class TrafficPredictionRApp:
    def __init__(self):
        self.model = self.load_model()
        self.a1_client = A1Client()
    
    def run(self):
        while True:
            # Collect historical data
            historical_data = self.collect_historical_data()
            
            # Train/update ML model
            self.train_model(historical_data)
            
            # Make predictions for next hour
            predictions = self.predict_traffic(horizon_hours=1)
            
            # Send policy to Near-RT RIC via A1
            policy = self.generate_policy(predictions)
            self.a1_client.send_policy(policy)
            
            time.sleep(3600)  # Run every hour
    
    def predict_traffic(self, horizon_hours):
        # Use LSTM model to predict traffic
        return self.model.predict(self.get_recent_features())
    
    def generate_policy(self, predictions):
        # Generate A1 policy based on predictions
        return {
            "policy_id": "traffic_pred_001",
            "policy_type": "RESOURCE_ALLOCATION",
            "parameters": {
                "predicted_load": predictions.tolist(),
                "recommended_resources": self.calc_resources(predictions)
            }
        }
```

#### 2. Energy Optimization rApp
```python
class EnergyOptimizationRApp:
    def run(self):
        while True:
            # Analyze energy consumption patterns
            energy_data = self.collect_energy_data()
            
            # Identify opportunities for savings
            savings_opportunities = self.analyze_savings(energy_data)
            
            # Generate policy to put underutilized cells to sleep
            if savings_opportunities:
                policy = {
                    "policy_id": "energy_savings_001",
                    "policy_type": "ENERGY_SAVINGS",
                    "parameters": {
                        "cells_to_sleep": savings_opportunities["cells"],
                        "sleep_schedule": savings_opportunities["schedule"]
                    }
                }
                self.a1_client.send_policy(policy)
            
            time.sleep(1800)  # Run every 30 minutes
```

## Simulated Radio Interface

### Virtual RF Environment

```go
// internal/radio/simulator.go
package radio

type RadioSimulator struct {
    cells      map[string]*VirtualCell
    ues        map[string]*VirtualUE
    channel    *ChannelModel
}

type VirtualCell struct {
    CellID        string
    CenterFreq    float64  // Hz
    Bandwidth     float64  // Hz
    TxPower       float64  // dBm
    Location      Location
    ActiveUEs     map[string]*VirtualUE
}

type VirtualUE struct {
    UEID          string
    IMSI          string
    Location      Location
    Velocity      Velocity
    AttachedCell  string
    RSRP          float64  // Reference Signal Received Power (dBm)
    RSRQ          float64  // Reference Signal Received Quality (dB)
    SINR          float64  // Signal-to-Interference-plus-Noise Ratio (dB)
}

type ChannelModel struct {
    PropagationModel string  // "FREE_SPACE", "PATH_LOSS", "URBAN", "RURAL"
    FadingType       string  // "NONE", "RAYLEIGH", "RICIAN"
}

func (r *RadioSimulator) ComputeRSRP(ue *VirtualUE, cell *VirtualCell) float64 {
    // Distance between UE and cell
    distance := computeDistance(ue.Location, cell.Location)
    
    // Path loss (simplified model)
    pathLoss := computePathLoss(cell.CenterFreq, distance, r.channel.PropagationModel)
    
    // RSRP = TxPower - PathLoss + Gains
    rsrp := cell.TxPower - pathLoss
    
    // Apply fading if configured
    if r.channel.FadingType != "NONE" {
        rsrp += r.applyFading(r.channel.FadingType)
    }
    
    return rsrp
}

func (r *RadioSimulator) SimulateUEAttachment(ue *VirtualUE) string {
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
    
    // Update UE state
    ue.AttachedCell = bestCell
    ue.RSRP = bestRSRP
    
    // Attach UE to cell
    r.cells[bestCell].ActiveUEs[ue.UEID] = ue
    
    return bestCell
}

func (r *RadioSimulator) SimulateHandover(ue *VirtualUE, targetCell string) error {
    // Detach from current cell
    if ue.AttachedCell != "" {
        delete(r.cells[ue.AttachedCell].ActiveUEs, ue.UEID)
    }
    
    // Attach to target cell
    r.cells[targetCell].ActiveUEs[ue.UEID] = ue
    ue.AttachedCell = targetCell
    
    // Update signal quality
    ue.RSRP = r.ComputeRSRP(ue, r.cells[targetCell])
    
    return nil
}
```

## Implementation Package Structure

```
nf/oran/
├── cmd/
│   ├── ocu/             # O-CU main
│   ├── odu/             # O-DU main
│   ├── oru/             # O-RU simulator main
│   ├── near-rt-ric/     # Near-RT RIC main
│   └── non-rt-ric/      # Non-RT RIC main
├── internal/
│   ├── ocu/
│   │   ├── cucp/        # CU-CP implementation
│   │   ├── cuup/        # CU-UP implementation
│   │   ├── rrc/         # RRC protocol
│   │   ├── pdcp/        # PDCP protocol
│   │   └── sdap/        # SDAP protocol
│   ├── odu/
│   │   ├── mac/         # MAC layer
│   │   ├── rlc/         # RLC layer
│   │   ├── scheduler/   # Resource scheduler
│   │   └── phy/         # High PHY
│   ├── oru/
│   │   └── simulator/   # RF simulator
│   ├── ric/
│   │   ├── near-rt/
│   │   │   ├── e2/      # E2 interface
│   │   │   ├── xapps/   # xApp framework
│   │   │   └── sdl/     # Shared Data Layer
│   │   └── non-rt/
│   │       ├── a1/      # A1 interface
│   │       └── rapps/   # rApp framework (Python)
│   ├── interfaces/
│   │   ├── e2/          # E2 protocol
│   │   ├── a1/          # A1 protocol
│   │   ├── f1/          # F1 protocol
│   │   └── fronthaul/   # Fronthaul (eCPRI)
│   └── radio/
│       └── simulator.go # Radio environment simulator
├── pkg/
│   └── e2sm/            # E2 Service Models (KPM, RC, NI)
├── xapps/               # Sample xApps
│   ├── traffic-steering/
│   ├── qos-prediction/
│   └── anomaly-detection/
├── rapps/               # Sample rApps (Python)
│   ├── traffic-prediction/
│   └── energy-optimization/
├── config/
├── test/
└── Dockerfile
```

## Configuration Example

```yaml
# O-CU Configuration
ocu:
  name: ocu-1
  instance_id: "uuid"
  
  cucp:
    sbi:
      addr: 0.0.0.0:8080
    n2:
      addr: 0.0.0.0:38412  # to AMF
    e1:
      addr: 0.0.0.0:38462  # to CU-UP
    f1c:
      addr: 0.0.0.0:38472  # to O-DU
    e2:
      ric_addr: near-rt-ric:36421
  
  cuup:
    n3:
      addr: 0.0.0.0:2152   # to UPF (GTP-U)
    f1u:
      addr: 0.0.0.0:2153   # to O-DU
    e2:
      ric_addr: near-rt-ric:36421

# O-DU Configuration
odu:
  name: odu-1
  instance_id: "uuid"
  
  f1:
    cu_addr: ocu-1:38472
  
  fronthaul:
    ru_addrs:
      - oru-1:50001
      - oru-2:50001
  
  e2:
    ric_addr: near-rt-ric:36421
  
  cells:
    - cell_id: "001"
      pci: 1  # Physical Cell ID
      center_freq: 3.5e9  # 3.5 GHz
      bandwidth: 100e6    # 100 MHz
      max_ues: 100

# O-RU Simulator Configuration
oru:
  name: oru-1
  instance_id: "uuid"
  
  fronthaul:
    odu_addr: odu-1:50001
  
  radio:
    simulation_mode: true
    cells:
      - cell_id: "001"
        location:
          lat: 37.7749
          lon: -122.4194
        tx_power: 43  # dBm
        center_freq: 3.5e9
        bandwidth: 100e6
    
    channel_model:
      propagation: "URBAN"
      fading: "RAYLEIGH"
      
  ues:
    - imsi: "001010000000001"
      location:
        lat: 37.7750
        lon: -122.4195
      mobility:
        enabled: true
        speed: 3  # m/s (walking)

# Near-RT RIC Configuration
near_rt_ric:
  name: near-rt-ric-1
  
  e2:
    listen_addr: 0.0.0.0:36421
  
  a1:
    non_rt_ric_addr: non-rt-ric:8080
  
  xapps:
    - name: traffic-steering
      enabled: true
      interval: 100ms
    - name: qos-prediction
      enabled: true
      interval: 1s
    - name: anomaly-detection
      enabled: true
      interval: 5s

# Non-RT RIC Configuration
non_rt_ric:
  name: non-rt-ric-1
  
  a1:
    listen_addr: 0.0.0.0:8080
  
  rapps:
    - name: traffic-prediction
      enabled: true
      interval: 1h
      ml_model: /models/traffic_lstm.h5
    - name: energy-optimization
      enabled: true
      interval: 30m
```

## Deployment

### Helm Chart Values (O-RAN Components)

```yaml
# deploy/helm/oran/values.yaml
ocu:
  replicaCount: 2
  image:
    repository: 5g/ocu
    tag: latest
  resources:
    requests:
      cpu: 500m
      memory: 512Mi

odu:
  replicaCount: 3
  image:
    repository: 5g/odu
    tag: latest
  resources:
    requests:
      cpu: 1000m
      memory: 1Gi

oru:
  # Simulated O-RUs
  replicaCount: 5
  image:
    repository: 5g/oru-simulator
    tag: latest
  resources:
    requests:
      cpu: 250m
      memory: 256Mi

nearRtRic:
  replicaCount: 2
  image:
    repository: 5g/near-rt-ric
    tag: latest
  resources:
    requests:
      cpu: 1000m
      memory: 2Gi

nonRtRic:
  replicaCount: 1
  image:
    repository: 5g/non-rt-ric
    tag: latest
  resources:
    requests:
      cpu: 500m
      memory: 1Gi
```

## Testing

### Integration Tests

```go
func TestORAN_E2EFlow(t *testing.T) {
    // 1. Start Near-RT RIC
    ric := startNearRTRIC()
    defer ric.Stop()
    
    // 2. Start O-CU
    ocu := startOCU()
    defer ocu.Stop()
    
    // 3. Start O-DU
    odu := startODU()
    defer odu.Stop()
    
    // 4. Start O-RU simulator
    oru := startORU()
    defer oru.Stop()
    
    // 5. Simulate UE attachment
    ue := &VirtualUE{
        IMSI: "001010000000001",
    }
    
    err := oru.AttachUE(ue)
    assert.NoError(t, err)
    
    // 6. Verify E2 messages sent to RIC
    time.Sleep(1 * time.Second)
    metrics := ric.GetReceivedMetrics()
    assert.Greater(t, len(metrics), 0)
    
    // 7. Trigger xApp action (handover)
    err = ric.TriggerHandover(ue.IMSI, "target-cell-002")
    assert.NoError(t, err)
    
    // 8. Verify handover completed
    time.Sleep(500 * time.Millisecond)
    assert.Equal(t, "target-cell-002", ue.AttachedCell)
}
```

## Benefits of O-RAN Architecture

1. **Openness:** Standardized interfaces enable multi-vendor deployment
2. **Intelligence:** RIC enables AI/ML-driven optimization
3. **Flexibility:** Virtualized components can run on COTS hardware
4. **Innovation:** xApps/rApps allow rapid feature deployment
5. **Testability:** Simulated radio interface enables comprehensive testing without RF equipment

## Summary

This O-RAN implementation provides:
- ✅ O-RAN Alliance compliant architecture
- ✅ E2, A1, F1, and Fronthaul interfaces
- ✅ Near-RT and Non-RT RIC with sample xApps/rApps
- ✅ Simulated radio interface for testing
- ✅ Clean separation between O-CU, O-DU, and O-RU
- ✅ Migration path to real RF hardware

The simulated approach allows full 5G network testing and development without physical radio equipment.

