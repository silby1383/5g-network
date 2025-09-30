# AI Agent Development Guide

## Overview

This document provides detailed instructions for each AI agent responsible for developing specific components of the 5G network system.

## Agent Development Template

Each agent should follow this standardized approach:

### 1. Initial Setup
```bash
# Clone repository
git clone <repo-url>
cd 5g-network

# Create agent branch
git checkout -b agent-<number>-<component-name>

# Set up development environment
make setup-dev-env
```

### 2. Development Checklist

For each component, the agent must complete:

- [ ] **Requirements Analysis**
  - Review 3GPP specifications
  - Identify all interfaces and protocols
  - Define data models
  
- [ ] **Architecture Design**
  - Create component architecture diagram
  - Define internal modules
  - Specify dependencies
  
- [ ] **API Design**
  - Create OpenAPI specification
  - Define all endpoints
  - Document request/response schemas
  
- [ ] **Implementation**
  - Set up project structure
  - Implement core functionality
  - Add error handling
  - Implement observability (metrics, traces, logs)
  
- [ ] **Testing**
  - Unit tests (>80% coverage)
  - Integration tests
  - Protocol compliance tests
  
- [ ] **Documentation**
  - Code documentation (GoDoc/TSDoc)
  - API documentation
  - Deployment guide
  - Troubleshooting guide
  
- [ ] **Containerization**
  - Dockerfile
  - Multi-stage builds
  - Security scanning
  
- [ ] **Kubernetes Deployment**
  - Helm chart
  - Resource limits
  - Health checks
  - Service definitions
  
- [ ] **CI/CD**
  - GitHub Actions workflow
  - Automated testing
  - Image building and pushing
  - Security scanning

---

## Agent 1: Core Control Plane (AMF + AUSF)

### Scope
- AMF (Access and Mobility Management Function)
- AUSF (Authentication Server Function)

### 3GPP Specifications
- **TS 23.502:** Procedures for the 5G System
- **TS 29.518:** AMF Services (Namf)
- **TS 29.509:** AUSF Services (Nausf)
- **TS 24.501:** Non-Access-Stratum (NAS) protocol
- **TS 38.413:** NG Application Protocol (NGAP)

### Key Interfaces

#### AMF Interfaces
- **N1:** UE ↔ AMF (NAS messages over HTTP/2)
- **N2:** RAN ↔ AMF (NGAP over SCTP)
- **N8:** AMF ↔ UDM
- **N11:** AMF ↔ SMF
- **N12:** AMF ↔ AUSF
- **N14:** AMF ↔ AMF (inter-AMF)
- **N22:** AMF ↔ NSSF

#### AUSF Interfaces
- **N12:** AUSF ↔ AMF
- **N13:** AUSF ↔ UDM

### Implementation Tasks

#### AMF Implementation

**Package Structure:**
```
nf/amf/
├── cmd/
│   └── main.go                 # Entry point
├── internal/
│   ├── context/
│   │   ├── amf_context.go      # AMF context
│   │   ├── ue_context.go       # UE context management
│   │   └── ran_context.go      # RAN context
│   ├── gmm/
│   │   ├── handler.go          # GMM (GPRS Mobility Management) handlers
│   │   ├── registration.go     # Registration procedures
│   │   ├── deregistration.go   # Deregistration procedures
│   │   └── mobility.go         # Mobility management
│   ├── nas/
│   │   ├── encoder.go          # NAS message encoding
│   │   ├── decoder.go          # NAS message decoding
│   │   └── security.go         # NAS security
│   ├── ngap/
│   │   ├── handler.go          # NGAP message handlers
│   │   ├── sctp.go             # SCTP server
│   │   └── messages.go         # NGAP message types
│   ├── sbi/
│   │   ├── server.go           # HTTP/2 SBI server
│   │   ├── consumer.go         # SBI client (to other NFs)
│   │   └── routes.go           # Route definitions
│   ├── event/
│   │   └── exposure.go         # Event exposure service
│   └── metrics/
│       └── metrics.go          # Prometheus metrics
├── pkg/
│   └── api/
│       └── openapi/            # OpenAPI specs
├── config/
│   └── config.yaml             # Configuration
├── test/
│   ├── unit/                   # Unit tests
│   └── integration/            # Integration tests
├── Dockerfile
└── README.md
```

**Core Functionality:**

1. **Registration Management (TS 23.502, Clause 4.2.2)**
   ```go
   // Implement registration state machine
   type RegistrationState int
   
   const (
       RM_DEREGISTERED RegistrationState = iota
       RM_REGISTERED
   )
   
   type UEContext struct {
       SUPI                 string
       SUCI                 string
       GUTI                 string
       RegistrationState    RegistrationState
       SecurityContext      *SecurityContext
       PDUSessions          map[string]*PDUSession
       // ... more fields
   }
   
   // Registration procedure
   func (amf *AMF) HandleRegistrationRequest(req *RegistrationRequest) error {
       // 1. Decode NAS message
       // 2. Identity verification (may invoke AUSF)
       // 3. Authentication (invoke AUSF via N12)
       // 4. Security mode procedure
       // 5. Subscription data retrieval (UDM via N8)
       // 6. Initial context setup with RAN
       // 7. Registration accept
   }
   ```

2. **NGAP (N2) Interface**
   ```go
   // SCTP server for NGAP
   type NGAPServer struct {
       listener *sctp.SCTPListener
       ranContexts map[string]*RANContext
   }
   
   func (s *NGAPServer) HandleNGSetupRequest(req *NGSetupRequest) {
       // Process NG setup from gNodeB
   }
   
   func (s *NGAPServer) HandleInitialUEMessage(msg *InitialUEMessage) {
       // Extract NAS message and forward to GMM handler
   }
   ```

3. **NAS Security (TS 33.501)**
   ```go
   type SecurityContext struct {
       Kseaf        []byte
       Kamf         []byte
       NASCountUL   uint32
       NASCountDL   uint32
       Kgnb         []byte
       // ... more security keys
   }
   
   func (sc *SecurityContext) EncryptNASMessage(msg []byte) ([]byte, error) {
       // Implement NAS encryption (AES-128)
   }
   
   func (sc *SecurityContext) VerifyNASMAC(msg []byte, mac []byte) bool {
       // Implement NAS integrity protection (CMAC)
   }
   ```

4. **Mobility Management**
   ```go
   func (amf *AMF) HandleHandoverRequest(req *HandoverRequest) error {
       // Inter-AMF or intra-AMF handover
       // Update UE context, coordinate with target RAN/AMF
   }
   ```

5. **Integration with Other NFs**
   ```go
   // UDM client
   func (amf *AMF) GetAuthenticationInfo(suci string) (*AuthInfo, error) {
       // Call UDM via N8 (Nudm_UEAuthentication)
   }
   
   // AUSF client
   func (amf *AMF) AuthenticateUE(suci string) (*AuthResult, error) {
       // Call AUSF via N12 (Nausf_UEAuthentication)
   }
   
   // SMF client
   func (amf *AMF) CreateSMContext(req *CreateSMContextRequest) error {
       // Call SMF via N11 (Nsmf_PDUSession)
   }
   ```

6. **OpenTelemetry Instrumentation**
   ```go
   import "go.opentelemetry.io/otel"
   
   func (amf *AMF) HandleRegistrationRequest(ctx context.Context, req *RegistrationRequest) error {
       ctx, span := otel.Tracer("amf").Start(ctx, "AMF.HandleRegistrationRequest")
       defer span.End()
       
       span.SetAttributes(
           attribute.String("suci", req.SUCI),
           attribute.String("registration_type", req.Type),
       )
       
       // Propagate context to downstream calls
       authInfo, err := amf.AuthenticateUE(ctx, req.SUCI)
       // ...
   }
   ```

#### AUSF Implementation

**Package Structure:**
```
nf/ausf/
├── cmd/
│   └── main.go
├── internal/
│   ├── context/
│   │   └── ausf_context.go
│   ├── auth/
│   │   ├── aka.go              # 5G-AKA authentication
│   │   ├── eap.go              # EAP-AKA' authentication
│   │   └── crypto.go           # Cryptographic operations
│   ├── sbi/
│   │   ├── server.go
│   │   └── consumer.go
│   └── metrics/
│       └── metrics.go
├── config/
│   └── config.yaml
├── test/
└── Dockerfile
```

**Core Functionality:**

1. **5G-AKA Authentication (TS 33.501)**
   ```go
   type AKAAuthenticator struct {
       udmClient *UDMClient
   }
   
   func (a *AKAAuthenticator) Authenticate(suci string) (*AuthResult, error) {
       // 1. De-conceal SUCI to get SUPI (via UDM)
       supi, err := a.udmClient.DeconcelIdentity(suci)
       
       // 2. Get authentication vector from UDM
       authVector, err := a.udmClient.GetAuthVector(supi)
       
       // 3. Compute authentication challenge
       // RAND, AUTN from auth vector
       
       // 4. Derive Kseaf (anchor key)
       kseaf := deriveKseaf(authVector.Kausf)
       
       return &AuthResult{
           RAND: authVector.RAND,
           AUTN: authVector.AUTN,
           XRES_STAR: authVector.XRES_STAR,
           Kseaf: kseaf,
       }, nil
   }
   
   func (a *AKAAuthenticator) ConfirmAuth(xres []byte) bool {
       // Verify XRES from UE against XRES_STAR
       return verifyXRES(xres, expectedXRES_STAR)
   }
   ```

2. **EAP-AKA' (TS 33.402)**
   ```go
   func (a *AKAAuthenticator) EAPAuthenticate(identity string) (*EAPResponse, error) {
       // Implement EAP-AKA' for non-3GPP access
   }
   ```

3. **SBI Interface**
   ```go
   // Nausf_UEAuthentication service
   func (s *SBIServer) UEAuthenticationPost(c *gin.Context) {
       var req UEAuthenticationRequest
       c.BindJSON(&req)
       
       authResult, err := s.ausf.Authenticate(req.SUCI)
       // Return authentication vectors
       c.JSON(200, authResult)
   }
   ```

### Testing Requirements

#### Unit Tests
```go
// Test registration procedure
func TestAMFRegistration(t *testing.T) {
    // Mock UDM, AUSF, RAN
    mockUDM := &MockUDM{}
    mockAUSF := &MockAUSF{}
    
    amf := NewAMF(mockUDM, mockAUSF)
    
    req := &RegistrationRequest{
        SUCI: "suci-0-001-01-0000-0-0-1234567890",
        Type: INITIAL_REGISTRATION,
    }
    
    err := amf.HandleRegistrationRequest(context.Background(), req)
    assert.NoError(t, err)
    
    // Verify UE context created
    ueCtx := amf.GetUEContext(req.SUCI)
    assert.NotNil(t, ueCtx)
    assert.Equal(t, RM_REGISTERED, ueCtx.RegistrationState)
}
```

#### Integration Tests
```go
// Test AMF + AUSF + UDM integration
func TestRegistrationFlow(t *testing.T) {
    // Start real AUSF and UDM instances
    ausf := startAUSF()
    udm := startUDM()
    amf := startAMF()
    
    defer func() {
        ausf.Stop()
        udm.Stop()
        amf.Stop()
    }()
    
    // Simulate UE registration
    client := NewAMFClient()
    resp, err := client.Register("suci-0-001-01-0000-0-0-1234567890")
    
    assert.NoError(t, err)
    assert.Equal(t, "REGISTERED", resp.Status)
}
```

### Configuration

```yaml
# config/config.yaml
amf:
  name: amf-1
  instance_id: "3fa85f64-5717-4562-b3fc-2c963f66afa6"
  
  # SBI interface
  sbi:
    scheme: https
    bind_addr: 0.0.0.0
    port: 8080
    tls:
      cert: /etc/amf/certs/amf.crt
      key: /etc/amf/certs/amf.key
  
  # NGAP (N2) interface
  ngap:
    bind_addr: 0.0.0.0
    port: 38412
  
  # NRF connection
  nrf:
    url: https://nrf.5gc.svc.cluster.local:8080
  
  # PLMN configuration
  plmn:
    mcc: "001"
    mnc: "01"
  
  # TAC (Tracking Area Code) list
  tac_list:
    - 1
    - 2
    - 3
  
  # Security
  security:
    integrity_algorithms:
      - NIA1
      - NIA2
      - NIA3
    encryption_algorithms:
      - NEA1
      - NEA2
      - NEA3
  
  # Observability
  observability:
    metrics:
      enabled: true
      port: 9090
    tracing:
      enabled: true
      exporter: otlp
      endpoint: otel-collector.observability.svc.cluster.local:4317
    logging:
      level: info
      format: json

ausf:
  name: ausf-1
  instance_id: "4fa85f64-5717-4562-b3fc-2c963f66afa7"
  
  sbi:
    scheme: https
    bind_addr: 0.0.0.0
    port: 8080
    tls:
      cert: /etc/ausf/certs/ausf.crt
      key: /etc/ausf/certs/ausf.key
  
  nrf:
    url: https://nrf.5gc.svc.cluster.local:8080
  
  # UDM connection (for auth vectors)
  udm:
    url: https://udm.5gc.svc.cluster.local:8080
  
  observability:
    metrics:
      enabled: true
      port: 9090
    tracing:
      enabled: true
      exporter: otlp
      endpoint: otel-collector.observability.svc.cluster.local:4317
```

### Dockerfile

```dockerfile
# AMF Dockerfile
FROM golang:1.22-alpine AS builder

WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o amf ./nf/amf/cmd

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /build/amf .
COPY --from=builder /build/nf/amf/config/config.yaml ./config/

# Non-root user
RUN addgroup -g 1000 amf && \
    adduser -D -u 1000 -G amf amf && \
    chown -R amf:amf /app

USER amf

EXPOSE 8080 38412 9090

CMD ["./amf", "--config", "./config/config.yaml"]
```

### Helm Chart

```yaml
# deploy/helm/amf/values.yaml
replicaCount: 3

image:
  repository: 5g/amf
  tag: latest
  pullPolicy: IfNotPresent

service:
  sbi:
    type: ClusterIP
    port: 8080
  ngap:
    type: LoadBalancer
    port: 38412
  metrics:
    port: 9090

resources:
  requests:
    cpu: 500m
    memory: 512Mi
  limits:
    cpu: 2000m
    memory: 2Gi

autoscaling:
  enabled: true
  minReplicas: 3
  maxReplicas: 10
  targetCPUUtilizationPercentage: 70

config:
  plmn:
    mcc: "001"
    mnc: "01"
  
  nrf:
    url: https://nrf.5gc.svc.cluster.local:8080

# OpenTelemetry
opentelemetry:
  enabled: true
  collectorEndpoint: otel-collector.observability.svc.cluster.local:4317

# Service mesh (Istio)
serviceMesh:
  enabled: true
  mtls:
    mode: STRICT
```

### Deliverables Checklist

- [ ] AMF service implementation
- [ ] AUSF service implementation
- [ ] NAS protocol encoder/decoder
- [ ] NGAP protocol implementation
- [ ] SCTP server implementation
- [ ] 5G-AKA authentication
- [ ] EAP-AKA' authentication
- [ ] SBI client/server
- [ ] OpenAPI specifications
- [ ] Unit tests (>80% coverage)
- [ ] Integration tests
- [ ] Docker images
- [ ] Helm charts
- [ ] Configuration examples
- [ ] Documentation (README, API docs)

---

## Agent 2: Session Management (SMF + PCF)

### Scope
- SMF (Session Management Function)
- PCF (Policy Control Function)

### 3GPP Specifications
- **TS 23.502:** Procedures for the 5G System (Session Management)
- **TS 23.503:** Policy and Charging Control Framework
- **TS 29.502:** SMF Services (Nsmf)
- **TS 29.507:** PCF Services (Npcf)
- **TS 29.244:** PFCP Protocol

### Key Interfaces

#### SMF Interfaces
- **N4:** SMF ↔ UPF (PFCP protocol)
- **N7:** SMF ↔ PCF
- **N10:** SMF ↔ UDM
- **N11:** SMF ↔ AMF

#### PCF Interfaces
- **N5:** PCF ↔ AF (Application Function)
- **N7:** PCF ↔ SMF
- **N15:** PCF ↔ AMF
- **N36:** PCF ↔ UDR

### Implementation Tasks

#### SMF Implementation

**Package Structure:**
```
nf/smf/
├── cmd/
│   └── main.go
├── internal/
│   ├── context/
│   │   ├── smf_context.go
│   │   ├── ue_context.go
│   │   └── pdu_session.go
│   ├── pfcp/
│   │   ├── client.go          # PFCP client (to UPF)
│   │   ├── session.go         # PFCP session management
│   │   └── messages.go        # PFCP message encoding/decoding
│   ├── pdusession/
│   │   ├── establishment.go   # Session establishment
│   │   ├── modification.go    # Session modification
│   │   ├── release.go         # Session release
│   │   └── qos.go             # QoS flow management
│   ├── sbi/
│   │   ├── server.go
│   │   ├── consumer.go
│   │   └── routes.go
│   ├── ipam/
│   │   └── ip_allocator.go    # UE IP address allocation
│   └── metrics/
│       └── metrics.go
├── config/
│   └── config.yaml
├── test/
└── Dockerfile
```

**Core Functionality:**

1. **PDU Session Establishment (TS 23.502, Clause 4.3.2)**
   ```go
   type PDUSession struct {
       SessionID        uint8
       SUPI             string
       DNN              string
       SNSSAI           string
       SessionType      string  // IPv4, IPv6, IPv4v6, Ethernet
       UEIP             net.IP
       UPFEndpoint      string
       QoSFlows         map[uint8]*QoSFlow
       SessionAMBR      *AMBR
       PFCPSessionID    uint64
   }
   
   func (smf *SMF) CreatePDUSession(ctx context.Context, req *CreatePDUSessionRequest) (*PDUSession, error) {
       ctx, span := otel.Tracer("smf").Start(ctx, "SMF.CreatePDUSession")
       defer span.End()
       
       // 1. Validate request
       // 2. Get subscription data from UDM (N10)
       // 3. Get policy from PCF (N7)
       // 4. Select UPF
       // 5. Allocate UE IP address
       // 6. Create PFCP session with UPF (N4)
       // 7. Store session context
       // 8. Return session info to AMF
   }
   ```

2. **PFCP (N4) Protocol**
   ```go
   type PFCPClient struct {
       conn *net.UDPConn
       upfEndpoint *net.UDPAddr
   }
   
   func (c *PFCPClient) SendSessionEstablishmentRequest(req *PFCPSessionEstablishmentRequest) error {
       // Create PFCP session on UPF
       // Install packet detection rules (PDR)
       // Install forwarding action rules (FAR)
       // Install QoS enforcement rules (QER)
   }
   
   type PDR struct {
       PDRID           uint16
       Precedence      uint32
       PDI             *PacketDetectionInfo
       OuterHeaderRemoval bool
       FARID           uint16
   }
   
   type FAR struct {
       FARID           uint16
       ApplyAction     uint8  // Forward, Drop, Buffer
       ForwardingParameters *ForwardingParameters
   }
   
   type QER struct {
       QERID           uint16
       QFI             uint8
       MBR             *Bitrate
       GBR             *Bitrate
   }
   ```

3. **QoS Flow Management**
   ```go
   type QoSFlow struct {
       QFI             uint8   // QoS Flow Identifier
       FiveQI          uint8   // 5G QoS Identifier
       Priority        uint8
       PacketDelayBudget uint16 // in ms
       PacketErrorRate float64
       AveragingWindow uint32
       MBR             *Bitrate  // Maximum Bit Rate
       GBR             *Bitrate  // Guaranteed Bit Rate (for GBR flows)
   }
   
   func (smf *SMF) EstablishQoSFlow(sessionID uint8, qosProfile *QoSProfile) (*QoSFlow, error) {
       // Create QoS flow based on policy from PCF
   }
   ```

4. **UE IP Address Management**
   ```go
   type IPAMAllocator struct {
       ipPools map[string]*IPPool
   }
   
   type IPPool struct {
       DNN         string
       CIDR        *net.IPNet
       Allocated   map[string]net.IP  // SUPI -> IP
   }
   
   func (ipam *IPAMAllocator) AllocateIP(supi string, dnn string, sessionType string) (net.IP, error) {
       pool := ipam.ipPools[dnn]
       ip := pool.GetNextAvailableIP()
       pool.Allocated[supi] = ip
       return ip, nil
   }
   ```

#### PCF Implementation

**Package Structure:**
```
nf/pcf/
├── cmd/
│   └── main.go
├── internal/
│   ├── context/
│   │   └── pcf_context.go
│   ├── policy/
│   │   ├── decision.go        # Policy decision point
│   │   ├── rules.go           # Policy rules engine
│   │   ├── qos.go             # QoS policies
│   │   ├── charging.go        # Charging policies
│   │   └── access.go          # Access control policies
│   ├── sbi/
│   │   ├── server.go
│   │   └── consumer.go
│   └── metrics/
│       └── metrics.go
├── config/
│   └── config.yaml
├── test/
└── Dockerfile
```

**Core Functionality:**

1. **Policy Decision Engine**
   ```go
   type PolicyDecisionEngine struct {
       udrClient *UDRClient
       rules     map[string]*PolicyRule
   }
   
   type PolicyRule struct {
       RuleID          string
       Precedence      uint32
       FlowDescription []string
       QoSData         *QoSData
       ChargingData    *ChargingData
       Conditions      []string
   }
   
   func (pde *PolicyDecisionEngine) GetPolicyDecision(ctx context.Context, req *PolicyRequest) (*PolicyDecision, error) {
       // 1. Retrieve policy data from UDR
       // 2. Evaluate policy rules
       // 3. Apply conditions (time, location, usage)
       // 4. Make QoS decision
       // 5. Make charging decision
       
       return &PolicyDecision{
           QoSFlows: qosFlows,
           ChargingRules: chargingRules,
           SessionAMBR: sessionAMBR,
       }, nil
   }
   ```

2. **QoS Policy**
   ```go
   func (pcf *PCF) DetermineQoS(supi string, dnn string, snssai string) (*QoSProfile, error) {
       // Based on subscription, DNN, slice
       // Return 5QI, ARP, bit rates
       
       return &QoSProfile{
           FiveQI:  9,  // Non-GBR
           Priority: 8,
           ARP: &AllocationRetentionPriority{
               PriorityLevel: 8,
               PreemptionCapability: MAY_PREEMPT,
               PreemptionVulnerability: NOT_PREEMPTABLE,
           },
       }, nil
   }
   ```

3. **Charging Control**
   ```go
   type ChargingRule struct {
       RuleID      string
       RatingGroup uint32
       ServiceID   uint32
       ReportingLevel string  // PCC_RULE_LEVEL or RATING_GROUP_LEVEL
       MeteringMethod string  // DURATION, VOLUME, EVENT
   }
   ```

### Configuration

```yaml
# SMF config
smf:
  name: smf-1
  instance_id: "5fa85f64-5717-4562-b3fc-2c963f66afa8"
  
  sbi:
    scheme: https
    bind_addr: 0.0.0.0
    port: 8080
  
  pfcp:
    bind_addr: 0.0.0.0
    port: 8805
  
  nrf:
    url: https://nrf.5gc.svc.cluster.local:8080
  
  # UPF selection
  upf:
    selection_mode: round_robin  # or least_loaded
    upf_list:
      - node_id: upf-1.5gc.svc.cluster.local
        n3_addr: 192.168.1.10
        n4_addr: 192.168.2.10
      - node_id: upf-2.5gc.svc.cluster.local
        n3_addr: 192.168.1.11
        n4_addr: 192.168.2.11
  
  # UE IP pools
  ip_pools:
    - dnn: internet
      cidr: 10.60.0.0/16
    - dnn: ims
      cidr: 10.61.0.0/16
  
  # DNN configurations
  dnn_list:
    - dnn: internet
      dns:
        ipv4: 8.8.8.8
        ipv6: 2001:4860:4860::8888
    - dnn: ims
      dns:
        ipv4: 10.10.10.1

# PCF config
pcf:
  name: pcf-1
  instance_id: "6fa85f64-5717-4562-b3fc-2c963f66afa9"
  
  sbi:
    scheme: https
    bind_addr: 0.0.0.0
    port: 8080
  
  nrf:
    url: https://nrf.5gc.svc.cluster.local:8080
  
  # Default policies
  default_qos:
    five_qi: 9
    priority: 8
  
  # Policy rules
  policy_rules:
    - rule_id: default-internet
      dnn: internet
      qos:
        five_qi: 9
        priority: 8
        mbr_uplink: 100000000    # 100 Mbps
        mbr_downlink: 200000000  # 200 Mbps
```

### Deliverables Checklist

- [ ] SMF service implementation
- [ ] PCF service implementation
- [ ] PFCP protocol implementation
- [ ] PDU session management
- [ ] QoS flow management
- [ ] UE IP address allocation (IPAM)
- [ ] Policy decision engine
- [ ] Integration with UPF, PCF, UDM, AMF
- [ ] Unit tests (>80% coverage)
- [ ] Integration tests
- [ ] Docker images
- [ ] Helm charts
- [ ] Documentation

---

## Agent 3: Data Plane (UPF) - Simulated Implementation

### Scope
- UPF (User Plane Function) with **simulated data plane**

### 3GPP Specifications
- **TS 23.501:** System Architecture (User Plane Function)
- **TS 29.244:** PFCP Protocol
- **TS 29.281:** GTP-U Protocol

### Key Interfaces
- **N3:** UPF ↔ gNodeB (GTP-U)
- **N4:** UPF ↔ SMF (PFCP)
- **N6:** UPF ↔ Data Network
- **N9:** UPF ↔ UPF (for multi-UPF scenarios)

### Implementation Strategy

**Simulated Data Plane Approach:**
- Implement packet processing logic in Go
- Simulate GTP-U encapsulation/decapsulation
- Track packets in memory (no actual forwarding initially)
- **Clean interface design** to allow future swap to eBPF/XDP
- Focus on **control plane correctness** and PFCP compliance

**Migration Path:**
- Phase 1: Simulated data plane (current)
- Phase 2 (Optional): Replace with eBPF/XDP for production performance
- Same PFCP interface ensures compatibility

### Implementation Tasks

**Package Structure:**
```
nf/upf/
├── cmd/
│   └── main.go
├── internal/
│   ├── context/
│   │   ├── upf_context.go
│   │   └── pfcp_session.go
│   ├── pfcp/
│   │   ├── server.go          # PFCP server
│   │   ├── handler.go
│   │   └── session.go
│   ├── dataplane/
│   │   ├── interface.go       # Data plane interface (for future swap)
│   │   ├── simulated/         # Simulated implementation
│   │   │   ├── processor.go  # Simulated packet processor
│   │   │   ├── gtp.go        # GTP encap/decap simulation
│   │   │   ├── qos.go        # QoS enforcement simulation
│   │   │   └── stats.go      # Statistics tracking
│   │   └── future_ebpf/       # Placeholder for eBPF implementation
│   ├── gtpu/
│   │   ├── handler.go         # GTP-U packet handling
│   │   └── tunnel.go          # GTP tunnel management
│   ├── qos/
│   │   └── enforcer.go        # QoS enforcement
│   └── metrics/
│       └── metrics.go
├── config/
│   └── config.yaml
├── test/
└── Dockerfile
```

**Core Functionality:**

1. **Data Plane Interface (for future extensibility)**
   ```go
   // internal/dataplane/interface.go
   package dataplane
   
   // DataPlane interface allows swapping implementations
   type DataPlane interface {
       // Initialize the data plane
       Initialize(config *Config) error
       
       // Install forwarding rules from PFCP session
       InstallPDR(sessionID uint64, pdr *PDR) error
       InstallFAR(sessionID uint64, far *FAR) error
       InstallQER(sessionID uint64, qer *QER) error
       
       // Remove rules
       RemovePDR(sessionID uint64, pdrID uint16) error
       RemoveFAR(sessionID uint64, farID uint16) error
       RemoveQER(sessionID uint64, qerID uint16) error
       
       // Process a packet (for simulation)
       ProcessPacket(packet *Packet) error
       
       // Get statistics
       GetStats() (*DataPlaneStats, error)
       
       // Shutdown
       Shutdown() error
   }
   
   type DataPlaneStats struct {
       PacketsProcessed uint64
       PacketsDropped   uint64
       BytesProcessed   uint64
       ActiveSessions   uint32
   }
   ```

2. **Simulated Data Plane Implementation**
   ```go
   // internal/dataplane/simulated/processor.go
   package simulated
   
   import (
       "sync"
       "github.com/your-org/5g-network/nf/upf/internal/dataplane"
   )
   
   type SimulatedDataPlane struct {
       sessions map[uint64]*SessionRules
       stats    *dataplane.DataPlaneStats
       mu       sync.RWMutex
   }
   
   type SessionRules struct {
       PDRs map[uint16]*dataplane.PDR
       FARs map[uint16]*dataplane.FAR
       QERs map[uint16]*dataplane.QER
   }
   
   func NewSimulatedDataPlane() *SimulatedDataPlane {
       return &SimulatedDataPlane{
           sessions: make(map[uint64]*SessionRules),
           stats:    &dataplane.DataPlaneStats{},
       }
   }
   
   func (s *SimulatedDataPlane) Initialize(config *dataplane.Config) error {
       log.Info("Initializing simulated data plane")
       return nil
   }
   
   func (s *SimulatedDataPlane) InstallPDR(sessionID uint64, pdr *dataplane.PDR) error {
       s.mu.Lock()
       defer s.mu.Unlock()
       
       if _, exists := s.sessions[sessionID]; !exists {
           s.sessions[sessionID] = &SessionRules{
               PDRs: make(map[uint16]*dataplane.PDR),
               FARs: make(map[uint16]*dataplane.FAR),
               QERs: make(map[uint16]*dataplane.QER),
           }
       }
       
       s.sessions[sessionID].PDRs[pdr.PDRID] = pdr
       log.Infof("Installed PDR %d for session %d", pdr.PDRID, sessionID)
       return nil
   }
   
   func (s *SimulatedDataPlane) ProcessPacket(packet *dataplane.Packet) error {
       s.mu.RLock()
       defer s.mu.RUnlock()
       
       // Simulate packet processing
       // 1. Match against PDRs
       // 2. Apply FAR (forward/drop/buffer)
       // 3. Apply QER (QoS enforcement)
       // 4. Update statistics
       
       s.stats.PacketsProcessed++
       s.stats.BytesProcessed += uint64(len(packet.Data))
       
       // Log packet for debugging
       log.Debugf("Processed packet: src=%s dst=%s size=%d", 
           packet.SrcIP, packet.DstIP, len(packet.Data))
       
       return nil
   }
   
   func (s *SimulatedDataPlane) GetStats() (*dataplane.DataPlaneStats, error) {
       s.mu.RLock()
       defer s.mu.RUnlock()
       
       stats := *s.stats
       stats.ActiveSessions = uint32(len(s.sessions))
       return &stats, nil
   }
   ```

3. **GTP-U Simulation**
   ```go
   // internal/dataplane/simulated/gtp.go
   package simulated
   
   import (
       "encoding/binary"
       "net"
   )
   
   type GTPUHeader struct {
       Version      uint8
       ProtocolType uint8
       MessageType  uint8
       Length       uint16
       TEID         uint32
       SequenceNumber uint16
   }
   
   // Simulate GTP-U encapsulation
   func (s *SimulatedDataPlane) EncapsulateGTPU(teid uint32, innerPacket []byte) []byte {
       header := &GTPUHeader{
           Version:      1,
           ProtocolType: 1,
           MessageType:  0xFF, // G-PDU
           Length:       uint16(len(innerPacket)),
           TEID:         teid,
       }
       
       gtpPacket := encodeGTPUHeader(header)
       gtpPacket = append(gtpPacket, innerPacket...)
       
       log.Debugf("Encapsulated packet in GTP-U: TEID=%d, size=%d", teid, len(gtpPacket))
       return gtpPacket
   }
   
   // Simulate GTP-U decapsulation
   func (s *SimulatedDataPlane) DecapsulateGTPU(gtpPacket []byte) (uint32, []byte, error) {
       if len(gtpPacket) < 12 {
           return 0, nil, fmt.Errorf("packet too short")
       }
       
       header := decodeGTPUHeader(gtpPacket[:12])
       innerPacket := gtpPacket[12:]
       
       log.Debugf("Decapsulated GTP-U packet: TEID=%d, inner size=%d", header.TEID, len(innerPacket))
       return header.TEID, innerPacket, nil
   }
   ```

4. **PFCP Session Management (Go)**
   ```go
   type PFCPSession struct {
       SessionID       uint64
       SEID            uint64  // SMF's SEID
       LocalSEID       uint64  // UPF's SEID
       PDRs            map[uint16]*PDR
       FARs            map[uint16]*FAR
       QERs            map[uint16]*QER
       GTPUTunnels     map[uint32]*GTPUTunnel  // TEID -> Tunnel
   }
   
   func (upf *UPF) HandleSessionEstablishmentRequest(req *PFCPSessionEstablishmentRequest) error {
       // 1. Create session context
       session := &PFCPSession{
           SessionID: req.SessionID,
           SEID:      req.CPFSEID.SEID,
           LocalSEID: upf.generateSEID(),
       }
       
       // 2. Install PDRs, FARs, QERs in data plane (eBPF maps)
       for _, pdr := range req.CreatePDR {
           upf.installPDR(session, pdr)
       }
       
       for _, far := range req.CreateFAR {
           upf.installFAR(session, far)
       }
       
       // 3. Create GTP-U tunnel
       teid := upf.allocateTEID()
       tunnel := &GTPUTunnel{
           TEID: teid,
           PeerAddr: req.FARs[0].ForwardingParameters.NetworkInstance,
       }
       
       // 4. Program eBPF map
       upf.updateBPFMap(teid, session)
       
       return nil
   }
   ```

3. **QoS Enforcement**
   ```go
   type QoSEnforcer struct {
       rateLimiters map[uint8]*rate.Limiter  // QFI -> RateLimiter
   }
   
   func (qos *QoSEnforcer) EnforceQoS(qfi uint8, packetSize int) bool {
       limiter, exists := qos.rateLimiters[qfi]
       if !exists {
           return true  // No rate limit
       }
       
       return limiter.AllowN(time.Now(), packetSize)
   }
   ```

4. **Performance Monitoring**
   ```go
   type UPFMetrics struct {
       PacketsProcessed   prometheus.Counter
       PacketsDropped     prometheus.Counter
       BytesProcessed     prometheus.Counter
       ThroughputBps      prometheus.Gauge
       ActiveSessions     prometheus.Gauge
       GTPUTunnels        prometheus.Gauge
   }
   ```

### Performance Requirements (Simulated Version)

**Current Phase (Simulated):**
- **Focus:** Functional correctness, not raw performance
- **Sessions:** Support 1,000+ concurrent sessions
- **Validation:** PFCP protocol compliance
- **Testing:** Control plane operations

**Future Phase (eBPF/XDP - Optional):**
- **Throughput:** 10+ Gbps per UPF instance
- **Latency:** <1ms packet processing
- **Sessions:** 100,000+ concurrent sessions
- **Packet rate:** 10+ Mpps (million packets per second)

### Testing

```go
// Functional test for simulated data plane
func TestSimulatedDataPlane(t *testing.T) {
    dp := simulated.NewSimulatedDataPlane()
    err := dp.Initialize(&dataplane.Config{})
    assert.NoError(t, err)
    
    // Test PDR installation
    pdr := &dataplane.PDR{
        PDRID: 1,
        Precedence: 100,
    }
    err = dp.InstallPDR(12345, pdr)
    assert.NoError(t, err)
    
    // Test packet processing
    packet := &dataplane.Packet{
        Data: []byte("test packet"),
        SrcIP: "10.0.0.1",
        DstIP: "10.0.0.2",
    }
    err = dp.ProcessPacket(packet)
    assert.NoError(t, err)
    
    // Check stats
    stats, err := dp.GetStats()
    assert.NoError(t, err)
    assert.Equal(t, uint64(1), stats.PacketsProcessed)
}

// Integration test with PFCP
func TestPFCPSessionEstablishment(t *testing.T) {
    upf := NewUPF(simulated.NewSimulatedDataPlane())
    
    // Receive PFCP session establishment request
    req := &PFCPSessionEstablishmentRequest{
        SessionID: 12345,
        CreatePDR: []*PDR{{PDRID: 1}},
        CreateFAR: []*FAR{{FARID: 1}},
        CreateQER: []*QER{{QERID: 1}},
    }
    
    resp, err := upf.HandleSessionEstablishmentRequest(req)
    assert.NoError(t, err)
    assert.Equal(t, PFCP_CAUSE_REQUEST_ACCEPTED, resp.Cause)
}
```

### Deliverables Checklist

- [ ] UPF service implementation with **simulated data plane**
- [ ] **Clean data plane interface** for future implementations
- [ ] PFCP server (full protocol compliance)
- [ ] GTP-U tunnel management (simulated)
- [ ] QoS enforcement (simulated)
- [ ] Packet classification and forwarding (simulated)
- [ ] N3, N4, N6, N9 interfaces
- [ ] **Functional tests** (PFCP protocol validation)
- [ ] Unit tests (>80% coverage)
- [ ] Integration tests with SMF
- [ ] Docker image
- [ ] Helm chart
- [ ] Documentation with migration guide for eBPF/XDP

---

## Coordination Between Agents

### Shared Resources

All agents should collaborate on:

1. **Common Libraries Repository**
   - Location: `/5g-network/common/`
   - Managed via: Shared PR reviews
   - Examples:
     - SBI framework
     - Protocol encoders/decoders
     - OpenTelemetry instrumentation
     - ClickHouse/Victoria Metrics clients

2. **OpenAPI Specifications**
   - Location: `/5g-network/api/openapi/`
   - Process:
     - Each agent creates OpenAPI spec for their NF
     - PRs reviewed by all agents
     - Versioned (semantic versioning)
   
3. **Protocol Buffer Definitions**
   - For internal gRPC communication
   - Shared code generation

### Integration Points

Each agent must:

1. **Register with NRF**
   - All NFs must register on startup
   - Use common NRF client library

2. **Implement Health Checks**
   - Kubernetes readiness/liveness probes
   - `/health` endpoint

3. **Export Metrics**
   - Victoria Metrics format
   - Common metric naming convention

4. **Implement Distributed Tracing**
   - OpenTelemetry SDK
   - Trace context propagation

5. **Structured Logging**
   - JSON format
   - Common fields (timestamp, level, nf_type, nf_instance_id)

### Development Workflow

1. **Branch Strategy**
   ```
   main
   ├── agent-1-amf-ausf
   ├── agent-2-smf-pcf
   ├── agent-3-upf
   ...
   ```

2. **PR Process**
   - Create PR from agent branch to main
   - Automated CI checks (tests, linting)
   - Code review by at least 1 other agent
   - Merge after approval

3. **Integration Testing**
   - Weekly integration test runs
   - All agents' code tested together
   - Issues tracked and assigned

### Communication

- **Daily standups** (async via Slack/Discord)
- **Weekly sync meetings**
- **Shared documentation** (Confluence/Notion)
- **Issue tracker** (GitHub Issues/Jira)

---

## Summary

This guide provides a template for each AI agent to follow. Key principles:

1. **3GPP Compliance:** Adhere to specifications
2. **Observability:** Metrics, traces, logs
3. **Testing:** Comprehensive unit and integration tests
4. **Documentation:** Code, API, deployment docs
5. **Containerization:** Docker + Kubernetes ready
6. **Performance:** Meet or exceed targets
7. **Collaboration:** Shared libraries, APIs, integration

Each agent should adapt this template to their specific NFs while maintaining consistency across the project.
