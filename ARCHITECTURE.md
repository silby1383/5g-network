# 5G Network Architecture & Implementation Plan

## Executive Summary

This document outlines the architecture and implementation strategy for a cloud-native, 3GPP-compliant 5G network with full observability, distributed tracing, and a comprehensive management WebUI.

## Table of Contents

1. [System Architecture](#system-architecture)
2. [Network Functions (NFs)](#network-functions-nfs)
3. [Technology Stack](#technology-stack)
4. [AI Agent Development Strategy](#ai-agent-development-strategy)
5. [Observability & Tracing](#observability--tracing)
6. [Data Architecture](#data-architecture)
7. [Implementation Roadmap](#implementation-roadmap)

## System Architecture

### High-Level Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                        Management Plane                          │
│  ┌────────────────────────────────────────────────────────────┐ │
│  │  WebUI (Next.js + React) - Network Management Console      │ │
│  │  - NF Deployment & Lifecycle Management                    │ │
│  │  - Subscriber Management                                    │ │
│  │  - Policy Management                                        │ │
│  │  - Real-time Metrics & Tracing Visualization               │ │
│  └────────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                      Control Plane (5GC)                         │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐       │
│  │   AMF    │  │   SMF    │  │   AUSF   │  │   UDM    │       │
│  │ (Access  │  │ (Session │  │  (Auth)  │  │  (Unified│       │
│  │ Mobility)│  │  Mgmt)   │  │          │  │   Data)  │       │
│  └──────────┘  └──────────┘  └──────────┘  └──────────┘       │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐       │
│  │   PCF    │  │   NRF    │  │   NSSF   │  │   UDR    │       │
│  │ (Policy) │  │(Discovery)│  │ (Slice)  │  │  (Data   │       │
│  │          │  │          │  │          │  │Repository)│       │
│  └──────────┘  └──────────┘  └──────────┘  └──────────┘       │
│  ┌──────────┐  ┌──────────┐                                    │
│  │   NEF    │  │  NWDAF   │                                    │
│  │(Exposure)│  │(Analytics)│                                    │
│  └──────────┘  └──────────┘                                    │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                          Data Plane                              │
│  ┌────────────────────────────────────────────────────────────┐ │
│  │                  UPF (User Plane Function)                  │ │
│  │  - High-performance packet processing                       │ │
│  │  - QoS enforcement                                          │ │
│  │  - Traffic routing & forwarding                            │ │
│  └────────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                         RAN (Radio Access Network)               │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐                      │
│  │  gNodeB  │  │ CU (Ctrl)│  │ DU (Data)│                      │
│  │  (5G BS) │  │   Unit   │  │   Unit)  │                      │
│  └──────────┘  └──────────┘  └──────────┘                      │
└─────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│                    Observability & Data Layer                    │
│  ┌──────────────┐ ┌──────────────┐ ┌──────────────┐           │
│  │ ClickHouse   │ │   Victoria   │ │    eBPF      │           │
│  │ (Subscriber  │ │   Metrics    │ │  Tracing     │           │
│  │   & UDR)     │ │  (Metrics)   │ │  (Traces)    │           │
│  └──────────────┘ └──────────────┘ └──────────────┘           │
│  ┌────────────────────────────────────────────────────────────┐ │
│  │  OpenTelemetry Collector (Trace Context Propagation)       │ │
│  └────────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────┘
```

## Network Functions (NFs)

### Core Network Functions

#### 1. AMF (Access and Mobility Management Function)
- **Responsibilities:**
  - Registration management
  - Connection management
  - Reachability management
  - Mobility management
  - Access authentication/authorization
  - NAS signaling termination
- **Interfaces:** N1 (UE), N2 (RAN), N8 (UDM), N11 (SMF), N12 (AUSF), N14 (AMF)
- **Language:** Go (high performance, excellent concurrency)
- **Key Libraries:** 
  - Free5GC AMF as reference
  - SCTP support for N2
  - NAS protocol implementation

#### 2. SMF (Session Management Function)
- **Responsibilities:**
  - Session establishment, modification, release
  - UE IP address allocation
  - DHCP functions
  - QoS control
  - Policy enforcement
- **Interfaces:** N4 (UPF), N7 (PCF), N10 (UDM), N11 (AMF)
- **Language:** Go
- **Key Features:**
  - PDU session management
  - PFCP protocol for UPF communication

#### 3. UPF (User Plane Function)
- **Responsibilities:**
  - Packet routing & forwarding
  - QoS handling
  - Traffic measurement
  - Lawful intercept
- **Interfaces:** N3 (RAN), N4 (SMF), N6 (Data Network), N9 (UPF)
- **Language:** Go (simulated data plane)
- **Key Features:**
  - **Simulated packet processing** (Go-based for initial development)
  - GTP-U encapsulation/decapsulation
  - QoS enforcement simulation
  - **Migration path:** Can be replaced with eBPF/XDP implementation later
  - Supports same PFCP interface for compatibility

#### 4. AUSF (Authentication Server Function)
- **Responsibilities:**
  - Authentication credential processing
  - EAP authentication
  - 5G AKA authentication
- **Interfaces:** N12 (AMF), N13 (UDM)
- **Language:** Go
- **Key Features:**
  - Cryptographic operations
  - SUCI/SUPI handling

#### 5. UDM (Unified Data Management)
- **Responsibilities:**
  - User identification handling
  - Access authorization
  - Subscription management
  - Service/session continuity
- **Interfaces:** N8 (AMF), N10 (SMF), N13 (AUSF), N35/N36 (UDR)
- **Language:** Go
- **Backend:** ClickHouse for subscriber data

#### 6. UDR (Unified Data Repository)
- **Responsibilities:**
  - Subscriber data storage
  - Policy data storage
  - Application data storage
- **Interfaces:** N35 (UDM), N36 (PCF), N37 (NEF)
- **Language:** Go with ClickHouse adapter
- **Database:** ClickHouse

#### 7. PCF (Policy Control Function)
- **Responsibilities:**
  - Policy rule provisioning
  - QoS decisions
  - Charging control
  - UE policy provisioning
- **Interfaces:** N5 (AF), N7 (SMF), N15 (AMF), N36 (UDR)
- **Language:** Go
- **Key Features:**
  - Policy decision engine
  - Dynamic policy updates

#### 8. NRF (Network Repository Function)
- **Responsibilities:**
  - NF service registration
  - NF service discovery
  - NF profile management
- **Interfaces:** All NFs (Service-based interface)
- **Language:** Go
- **Key Features:**
  - Service mesh integration
  - Health checking

#### 9. NSSF (Network Slice Selection Function)
- **Responsibilities:**
  - Network slice selection
  - Allowed NSSAI determination
- **Interfaces:** N22 (AMF)
- **Language:** Go
- **Key Features:**
  - Slice policy management

#### 10. NEF (Network Exposure Function)
- **Responsibilities:**
  - External API exposure
  - Event exposure
  - Capability exposure
- **Interfaces:** N33 (AF), N29/N30 (SMF), N37 (UDR)
- **Language:** Go
- **Key Features:**
  - RESTful API gateway
  - OAuth2/OpenID Connect

#### 11. NWDAF (Network Data Analytics Function)
- **Responsibilities:**
  - Network analytics
  - ML-based predictions
  - Data collection from NFs
- **Language:** Python + Go
- **Key Features:**
  - Integration with Victoria Metrics
  - ML models for network optimization

### RAN Components

#### 12. gNodeB (5G Base Station) - Basic CU/DU Split
- **Components:**
  - **Central Unit (CU)** - RRC, PDCP (control and user plane)
  - **Distributed Unit (DU)** - RLC, MAC, High PHY
  - **Radio Unit (RU)** - Low PHY, RF processing (**simulated**)
- **Language:** Go (for simulated implementation)
- **Key Features:**
  - **CU/DU split architecture**
  - Interfaces:
    - **N2:** CU ↔ AMF (NGAP)
    - **N3:** CU ↔ UPF (GTP-U)
    - **F1:** DU ↔ CU split interface
    - **Fronthaul:** DU ↔ RU (**simulated**)
  - **Simulated radio interface** for testing without physical RF
  - Virtual RF environment with channel modeling
  - UE attachment and mobility simulation

## Technology Stack

### Core Technologies

#### Container Orchestration
- **Kubernetes** (K8s)
  - Helm charts for each NF
  - StatefulSets for stateful NFs
  - Deployments for stateless NFs
  - HPA (Horizontal Pod Autoscaler)
  - Custom Resource Definitions (CRDs) for 5G resources

#### Service Mesh
- **Istio** or **Linkerd**
  - mTLS between NFs
  - Traffic management
  - Circuit breaking
  - Observability integration

#### Programming Languages
- **Go**: Primary language for all NFs (including simulated UPF)
  - High performance
  - Excellent concurrency
  - Strong networking libraries
  - Simulated data plane for initial development
- **TypeScript/Next.js**: WebUI
- **Python**: NWDAF analytics, orchestration, RIC xApps/rApps
- **Future (Optional):** eBPF/C or Rust for production data plane

#### Communication Protocols
- **HTTP/2 + JSON** (SBI - Service Based Interface)
- **gRPC** (Internal microservices)
- **SCTP** (N2 interface)
- **GTP-U** (User plane tunneling)
- **PFCP** (N4 interface)

### Data Layer

#### ClickHouse
- **Use Cases:**
  - Subscriber database (UDR)
  - Policy data
  - Call Detail Records (CDR)
  - Session records
- **Schema Design:**
  - Subscriber profiles
  - Authentication vectors
  - Subscription data
  - Network slice subscriptions

#### Victoria Metrics
- **Use Cases:**
  - Real-time metrics collection
  - Time-series data
  - Performance monitoring
  - Capacity planning
- **Metrics:**
  - NF CPU/Memory/Network
  - Registration success/failure rates
  - Session establishment times
  - Throughput per UPF
  - Active sessions per SMF

### Observability

#### eBPF Tracing
- **Components:**
  - **eBPF programs** in each NF container
  - Kernel-level tracing without code changes
  - Network packet tracing
  - System call tracing

#### OpenTelemetry
- **Instrumentation:**
  - All NFs instrumented with OTEL SDK
  - Trace context propagation via HTTP headers (W3C Trace Context)
  - Spans for each SBI request/response
  - Parent-child span relationships

#### Trace Context Propagation
```
HTTP Header: traceparent: 00-{trace-id}-{span-id}-{flags}

Example flow:
AMF -> UDM -> UDR
├─ Span: AMF.RegisterUE (trace-id: abc123)
   ├─ Span: UDM.GetSubscriberData (parent: AMF.RegisterUE)
      ├─ Span: UDR.QueryClickHouse (parent: UDM.GetSubscriberData)
```

- **Storage:** Jaeger or Tempo (with ClickHouse backend)
- **Visualization:** Grafana + Tempo/Jaeger UI

#### Logging
- **Stack:** Fluent Bit -> Loki -> Grafana
- **Structured logging** with correlation IDs

### Management WebUI

#### Frontend Stack
- **Framework:** Next.js 14+ (App Router)
- **UI Library:** Shadcn UI + Radix UI
- **Styling:** Tailwind CSS
- **State Management:** Zustand + TanStack Query
- **Real-time:** WebSockets for live updates
- **Visualization:** 
  - D3.js for network topology
  - Recharts for metrics
  - React Flow for call flow visualization

#### Backend API
- **Framework:** Next.js API routes or separate Go service
- **API Style:** RESTful + GraphQL (for complex queries)
- **Authentication:** JWT + RBAC
- **Real-time:** Server-Sent Events (SSE) or WebSockets

#### WebUI Features
1. **Dashboard**
   - Network status overview
   - Active subscribers
   - Session statistics
   - Alerts & notifications

2. **NF Management**
   - Deploy/scale/stop NFs
   - Configuration management
   - Health monitoring
   - Log viewer

3. **Subscriber Management**
   - Add/edit/delete subscribers
   - Subscription profile management
   - Device management
   - Location tracking

4. **Policy Management**
   - QoS policies
   - Charging rules
   - Slice policies
   - Access control

5. **Network Slicing**
   - Create/manage slices
   - Slice resource allocation
   - Slice performance monitoring

6. **Observability**
   - Metrics dashboards (Victoria Metrics)
   - Trace visualization (call flows)
   - Log aggregation
   - Alerts management

7. **Topology View**
   - Interactive network map
   - NF relationships
   - Traffic flow visualization

## AI Agent Development Strategy

### Multi-Agent Architecture

Each AI agent is responsible for developing one or more related Network Functions. This approach ensures:
- Domain expertise per agent
- Parallel development
- Clear separation of concerns
- Easier testing and integration

### Agent Breakdown

#### Agent 1: Core Control Plane Agent
**Responsibilities:**
- AMF (Access and Mobility Management)
- AUSF (Authentication Server)

**Tasks:**
1. Implement 3GPP TS 23.502 registration procedures
2. NAS message handling
3. SCTP/NGAP protocol implementation
4. Authentication flow (5G-AKA)
5. Mobility management state machines
6. Integration with UDM and SMF

**Deliverables:**
- AMF service (Go)
- AUSF service (Go)
- Docker images
- Helm charts
- Unit tests (>80% coverage)
- Integration tests
- OpenAPI specs

#### Agent 2: Session Management Agent
**Responsibilities:**
- SMF (Session Management Function)
- PCF (Policy Control Function)

**Tasks:**
1. PDU session establishment/modification/release
2. PFCP protocol implementation (N4)
3. Policy decision point integration
4. QoS management
5. Charging control

**Deliverables:**
- SMF service (Go)
- PCF service (Go)
- PFCP library
- Docker images
- Helm charts
- Tests

#### Agent 3: Data Plane Agent
**Responsibilities:**
- UPF (User Plane Function) - Simulated Implementation

**Tasks:**
1. **Simulated packet processing in Go** (with clean interfaces for future eBPF upgrade)
2. GTP-U encapsulation/decapsulation (simulated)
3. QoS enforcement (simulated)
4. Traffic steering simulation
5. N4 (PFCP) session management
6. Metrics export
7. **Design interfaces** to allow future swap to eBPF/XDP implementation

**Deliverables:**
- UPF service (Go with simulated data plane)
- Functional tests (control plane correctness)
- Docker images
- Helm charts
- Tests
- **Note:** Performance targets relaxed for simulated version; focus on correctness

#### Agent 4: Data Management Agent
**Responsibilities:**
- UDM (Unified Data Management)
- UDR (Unified Data Repository)
- ClickHouse schema design

**Tasks:**
1. Subscriber data management
2. ClickHouse integration
3. Data model design (3GPP compliant)
4. CRUD APIs for subscriber data
5. Authentication vector generation
6. Data replication and backup

**Deliverables:**
- UDM service (Go)
- UDR service (Go)
- ClickHouse schemas and migrations
- Subscriber data API
- Docker images
- Helm charts
- Tests

#### Agent 5: Service Discovery & Exposure Agent
**Responsibilities:**
- NRF (Network Repository Function)
- NEF (Network Exposure Function)
- NSSF (Network Slice Selection)

**Tasks:**
1. Service registration/discovery
2. NF profile management
3. External API gateway
4. OAuth2 authentication
5. Network slice selection logic
6. Kubernetes service discovery integration

**Deliverables:**
- NRF service (Go)
- NEF service (Go)
- NSSF service (Go)
- API gateway
- Docker images
- Helm charts
- Tests

#### Agent 6: Analytics & Intelligence Agent
**Responsibilities:**
- NWDAF (Network Data Analytics Function)
- Victoria Metrics integration
- ML models

**Tasks:**
1. Data collection from all NFs
2. Victoria Metrics adapter
3. Network analytics engine
4. ML models for:
   - Load prediction
   - Anomaly detection
   - QoS optimization
5. Real-time analytics API

**Deliverables:**
- NWDAF service (Python + Go)
- ML models
- Analytics dashboard
- Docker images
- Helm charts
- Tests

#### Agent 7: RAN Agent
**Responsibilities:**
- gNodeB with CU/DU split (simulated implementation)
- Simulated radio interface

**Tasks:**
1. **gNodeB Implementation:**
   - Central Unit (CU) for RRC and PDCP
   - Distributed Unit (DU) for RLC, MAC, High PHY
   - Radio Unit (RU) - simulated
2. **Interfaces:**
   - N2 interface (NGAP to AMF)
   - N3 interface (GTP-U to UPF)
   - F1 interface (DU to CU split)
   - Fronthaul (DU to RU) - simulated
3. **Simulated Radio Interface:**
   - Virtual RF environment
   - Channel modeling (path loss, fading)
   - UE attachment simulation
   - Mobility simulation
4. UE simulator integration
5. Basic radio resource management

**Deliverables:**
- CU service (Go)
- DU service (Go)
- RU simulator (Go)
- UE simulator with radio simulation
- Docker images
- Helm charts
- Tests

#### Agent 8: Observability Agent
**Responsibilities:**
- eBPF-based distributed tracing infrastructure
- OpenTelemetry instrumentation with eBPF
- Trace context propagation across all NFs

**Tasks:**
1. **eBPF Tracing Programs:**
   - HTTP request/response tracing at kernel level
   - Function entry/exit tracing (uprobe/kprobe)
   - Network packet tracing
   - System call tracing
   - Trace context extraction from headers
2. **OpenTelemetry Integration:**
   - eBPF → OpenTelemetry exporter
   - Application-level OTEL instrumentation
   - Unified trace collection
3. **Trace Context Propagation:**
   - W3C Trace Context standard
   - Automatic context injection in eBPF
   - Cross-NF span correlation
4. OpenTelemetry Collector configuration
5. Integration with Jaeger/Tempo
6. Custom eBPF dashboards in Grafana

**Deliverables:**
- eBPF tracing programs (C)
- eBPF loader (Go)
- OTEL instrumentation package (Go)
- Trace propagation middleware
- eBPF → OTEL bridge
- Grafana dashboards
- Documentation

#### Agent 9: WebUI Frontend Agent
**Responsibilities:**
- Management WebUI (Next.js)
- Frontend components and pages

**Tasks:**
1. Next.js application setup
2. Component library (Shadcn UI)
3. Pages:
   - Dashboard
   - NF management
   - Subscriber management
   - Policy management
   - Observability views
   - Network topology
4. Real-time updates (WebSockets)
5. Responsive design
6. Authentication UI

**Deliverables:**
- Next.js application
- Component library
- E2E tests (Playwright)
- Docker image
- Deployment config

#### Agent 10: WebUI Backend Agent
**Responsibilities:**
- Backend API for WebUI
- Integration with all NFs

**Tasks:**
1. REST API + GraphQL API
2. Authentication/authorization
3. WebSocket server
4. Integration with:
   - Kubernetes API (NF lifecycle)
   - ClickHouse (subscriber data)
   - Victoria Metrics (metrics)
   - Jaeger/Tempo (traces)
   - All NFs (control operations)
5. RBAC implementation

**Deliverables:**
- API service (Go)
- GraphQL schema
- Authentication service
- Docker image
- Helm chart
- API documentation (OpenAPI)

#### Agent 11: Infrastructure & DevOps Agent
**Responsibilities:**
- Kubernetes manifests
- CI/CD pipelines
- Helm charts
- Deployment automation

**Tasks:**
1. Kubernetes cluster setup (kind/k3s for dev, production configs)
2. Helm chart for entire 5G deployment
3. CI/CD pipelines (GitHub Actions/GitLab CI)
4. Infrastructure as Code (Terraform for cloud resources)
5. Monitoring stack deployment (Victoria Metrics, Grafana, Loki)
6. ClickHouse deployment and clustering
7. Network policies
8. Ingress/Egress configuration

**Deliverables:**
- Helm charts for all components
- K8s manifests
- CI/CD pipelines
- Terraform modules
- Deployment scripts
- Documentation

#### Agent 12: Testing & Integration Agent
**Responsibilities:**
- Integration testing
- End-to-end testing
- Performance testing

**Tasks:**
1. Integration test suite
2. E2E test scenarios (3GPP test cases)
3. Performance benchmarks
4. Load testing (K6/Locust)
5. Chaos engineering tests
6. Test automation framework

**Deliverables:**
- Integration test suite
- E2E test scenarios
- Performance test reports
- Load testing scripts
- CI integration

### Agent Coordination

#### Shared Responsibilities
1. **Common Libraries:**
   - NAS message encoding/decoding
   - SBI client/server (HTTP/2)
   - PFCP protocol
   - GTP protocol
   - OpenTelemetry instrumentation
   - ClickHouse client
   - Victoria Metrics client

2. **Interface Definitions:**
   - OpenAPI specifications for all SBI interfaces
   - Protocol buffer definitions for gRPC
   - Database schemas

3. **Development Standards:**
   - Code style guides
   - Testing standards
   - Documentation requirements
   - Git workflow

#### Development Workflow

```
Phase 1: Foundation (Weeks 1-4)
├─ Agent 11: Set up infrastructure
├─ Agent 8: Set up observability stack
├─ All Agents: Define interfaces (OpenAPI specs)
└─ Common library development

Phase 2: Core NFs (Weeks 5-12)
├─ Agent 1: AMF + AUSF
├─ Agent 4: UDM + UDR + ClickHouse
├─ Agent 5: NRF (needed by all NFs)
└─ Integration testing starts

Phase 3: Session & Data Plane (Weeks 13-20)
├─ Agent 2: SMF + PCF
├─ Agent 3: UPF
├─ Agent 7: Basic gNodeB/RAN
└─ E2E registration + session establishment tests

Phase 4: Advanced Features (Weeks 21-28)
├─ Agent 5: NEF + NSSF
├─ Agent 6: NWDAF + Analytics
├─ Agent 8: Advanced eBPF tracing
└─ Performance optimization

Phase 5: Management UI (Weeks 29-36)
├─ Agent 9: Frontend
├─ Agent 10: Backend API
└─ Integration with all NFs

Phase 6: Testing & Hardening (Weeks 37-44)
├─ Agent 12: Comprehensive testing
├─ All Agents: Bug fixes and optimization
├─ Security hardening
└─ Documentation finalization

Phase 7: Production Readiness (Weeks 45-48)
├─ Production deployment configs
├─ Disaster recovery procedures
├─ Runbooks and operational docs
└─ Final performance validation
```

## Data Architecture

### ClickHouse Schema

```sql
-- Subscriber Database
CREATE TABLE subscribers (
    supi String,  -- Subscriber Permanent Identifier
    suci String,  -- Subscriber Concealed Identifier
    imsi String,
    msisdn String,
    imei String,
    subscription_profile_id String,
    access_restrictions Array(String),
    subscriber_status Enum('ACTIVE', 'SUSPENDED', 'DELETED'),
    created_at DateTime,
    updated_at DateTime,
    INDEX idx_supi supi TYPE bloom_filter GRANULARITY 4,
    INDEX idx_imsi imsi TYPE bloom_filter GRANULARITY 4
) ENGINE = ReplicatedMergeTree('/clickhouse/tables/{shard}/subscribers', '{replica}')
ORDER BY (supi)
PARTITION BY toYYYYMM(created_at);

-- Authentication Vectors
CREATE TABLE auth_vectors (
    supi String,
    rand FixedString(16),
    autn FixedString(16),
    xres_star FixedString(16),
    kseaf FixedString(32),
    generated_at DateTime,
    expires_at DateTime,
    used Bool DEFAULT false
) ENGINE = ReplicatedMergeTree('/clickhouse/tables/{shard}/auth_vectors', '{replica}')
ORDER BY (supi, generated_at)
TTL expires_at + INTERVAL 1 DAY;

-- Session Records
CREATE TABLE pdu_sessions (
    session_id String,
    supi String,
    dnn String,  -- Data Network Name
    snssai String,  -- Single Network Slice Selection Assistance Information
    pdu_session_type Enum('IPv4', 'IPv6', 'IPv4v6', 'Ethernet', 'Unstructured'),
    ue_ipv4 IPv4,
    ue_ipv6 IPv6,
    upf_id String,
    smf_id String,
    qos_profile_id String,
    session_ambr_uplink UInt64,
    session_ambr_downlink UInt64,
    created_at DateTime,
    closed_at Nullable(DateTime),
    duration_seconds UInt32 MATERIALIZED dateDiff('second', created_at, closed_at),
    INDEX idx_supi supi TYPE bloom_filter GRANULARITY 4,
    INDEX idx_session session_id TYPE bloom_filter GRANULARITY 4
) ENGINE = ReplicatedMergeTree('/clickhouse/tables/{shard}/pdu_sessions', '{replica}')
ORDER BY (created_at, session_id)
PARTITION BY toYYYYMM(created_at);

-- Policy Data
CREATE TABLE policies (
    policy_id String,
    policy_type Enum('QoS', 'Charging', 'Access', 'Slice'),
    snssai String,
    dnn String,
    policy_rules String,  -- JSON
    priority UInt8,
    active Bool DEFAULT true,
    created_at DateTime,
    updated_at DateTime
) ENGINE = ReplicatedMergeTree('/clickhouse/tables/{shard}/policies', '{replica}')
ORDER BY (policy_type, priority, policy_id);

-- Network Slice Subscriptions
CREATE TABLE slice_subscriptions (
    supi String,
    snssai String,
    default_indicator Bool,
    slice_priority UInt8,
    access_types Array(Enum('3GPP', 'NON_3GPP')),
    created_at DateTime
) ENGINE = ReplicatedMergeTree('/clickhouse/tables/{shard}/slice_subscriptions', '{replica}')
ORDER BY (supi, snssai);

-- Call Detail Records (CDR)
CREATE TABLE cdrs (
    cdr_id String,
    supi String,
    imsi String,
    msisdn String,
    session_id String,
    call_type Enum('Voice', 'Data', 'SMS', 'Video'),
    start_time DateTime,
    end_time DateTime,
    duration_seconds UInt32,
    data_volume_uplink UInt64,
    data_volume_downlink UInt64,
    charging_id String,
    cost Decimal(10, 4),
    currency String,
    serving_network String,
    location String
) ENGINE = ReplicatedMergeTree('/clickhouse/tables/{shard}/cdrs', '{replica}')
ORDER BY (start_time, supi)
PARTITION BY toYYYYMM(start_time);

-- Trace Spans (for long-term storage)
CREATE TABLE trace_spans (
    trace_id String,
    span_id String,
    parent_span_id String,
    operation_name String,
    nf_type String,
    nf_instance_id String,
    start_time DateTime64(9),
    duration_ns UInt64,
    status_code UInt16,
    tags Map(String, String),
    INDEX idx_trace trace_id TYPE bloom_filter GRANULARITY 4
) ENGINE = MergeTree()
ORDER BY (start_time, trace_id, span_id)
PARTITION BY toYYYYMM(start_time)
TTL start_time + INTERVAL 90 DAY;
```

### Victoria Metrics Configuration

```yaml
# Metrics to collect from each NF

# General NF metrics
- nf_cpu_usage{nf_type, nf_instance_id}
- nf_memory_usage{nf_type, nf_instance_id}
- nf_network_bytes_sent{nf_type, nf_instance_id, interface}
- nf_network_bytes_received{nf_type, nf_instance_id, interface}
- nf_http_requests_total{nf_type, nf_instance_id, method, endpoint, status}
- nf_http_request_duration_seconds{nf_type, nf_instance_id, method, endpoint}

# AMF specific
- amf_registered_ues{amf_instance_id}
- amf_registration_requests_total{amf_instance_id, result}
- amf_registration_duration_seconds{amf_instance_id}
- amf_handovers_total{amf_instance_id, result}

# SMF specific
- smf_active_sessions{smf_instance_id, dnn, snssai}
- smf_session_establishment_requests_total{smf_instance_id, result}
- smf_session_establishment_duration_seconds{smf_instance_id}
- smf_qos_flows_active{smf_instance_id}

# UPF specific
- upf_throughput_bps{upf_instance_id, direction}
- upf_packets_processed_total{upf_instance_id}
- upf_packets_dropped_total{upf_instance_id, reason}
- upf_active_pdu_sessions{upf_instance_id}
- upf_gtp_tunnels_active{upf_instance_id}

# UDM/UDR specific
- udm_subscriber_queries_total{udm_instance_id, query_type, result}
- udm_query_duration_seconds{udm_instance_id, query_type}
- udr_clickhouse_query_duration_seconds{udr_instance_id, query_type}

# NRF specific
- nrf_registered_nfs{nrf_instance_id, nf_type}
- nrf_discovery_requests_total{nrf_instance_id, nf_type}
- nrf_nf_heartbeats_total{nrf_instance_id, nf_type}

# NWDAF specific
- nwdaf_predictions_total{nwdaf_instance_id, prediction_type}
- nwdaf_anomalies_detected_total{nwdaf_instance_id, anomaly_type}
```

## Implementation Roadmap

### Month 1-2: Foundation
**Weeks 1-4:**
- [ ] Project structure setup
- [ ] Kubernetes cluster (development)
- [ ] CI/CD pipeline skeleton
- [ ] ClickHouse deployment
- [ ] Victoria Metrics deployment
- [ ] OpenTelemetry Collector deployment
- [ ] Common libraries:
  - [ ] SBI client/server framework
  - [ ] OpenTelemetry instrumentation
  - [ ] ClickHouse client wrapper
  - [ ] Victoria Metrics client
- [ ] OpenAPI specifications for all NF SBIs
- [ ] eBPF development environment

**Weeks 5-8:**
- [ ] NRF (Network Repository Function) - Basic version
- [ ] UDR (Unified Data Repository) + ClickHouse integration
- [ ] eBPF tracing framework
- [ ] Development documentation
- [ ] Code review process

### Month 3-4: Core Network Functions
**Weeks 9-12:**
- [ ] AMF (Access and Mobility Management)
  - [ ] Registration management
  - [ ] NAS signaling
  - [ ] NGAP (N2) interface
- [ ] AUSF (Authentication Server Function)
  - [ ] 5G-AKA authentication
  - [ ] EAP-AKA' authentication
- [ ] UDM (Unified Data Management)
  - [ ] Integration with UDR/ClickHouse
  - [ ] Subscription data management

**Weeks 13-16:**
- [ ] Integration: AMF + AUSF + UDM + UDR + NRF
- [ ] Test: UE registration flow
- [ ] OpenTelemetry instrumentation for above NFs
- [ ] eBPF probes for N1/N2 interfaces

### Month 5-6: Session Management & Data Plane
**Weeks 17-20:**
- [ ] SMF (Session Management Function)
  - [ ] PDU session management
  - [ ] QoS handling
  - [ ] PFCP (N4) interface
- [ ] PCF (Policy Control Function)
  - [ ] Policy decision point
  - [ ] QoS policies

**Weeks 21-24:**
- [ ] UPF (User Plane Function)
  - [ ] GTP-U encapsulation/decapsulation
  - [ ] Packet routing
  - [ ] QoS enforcement
  - [ ] eBPF/XDP data plane
  - [ ] N4 (PFCP) support
- [ ] Integration: SMF + UPF + AMF
- [ ] Test: End-to-end PDU session establishment
- [ ] Performance benchmarking (target: 10 Gbps)

### Month 7-8: RAN and Additional NFs
**Weeks 25-28:**
- [ ] Basic gNodeB (simulated)
  - [ ] N2 interface (NGAP)
  - [ ] N3 interface (GTP-U)
- [ ] UE Simulator
- [ ] NSSF (Network Slice Selection Function)
- [ ] NEF (Network Exposure Function)
  - [ ] External API gateway
  - [ ] OAuth2/OIDC

**Weeks 29-32:**
- [ ] NWDAF (Network Data Analytics Function)
  - [ ] Data collection from Victoria Metrics
  - [ ] Basic analytics
  - [ ] ML models (load prediction, anomaly detection)
- [ ] Network slicing implementation
- [ ] Multi-UE, multi-session testing

### Month 9-10: Management WebUI
**Weeks 33-36:**
- [ ] WebUI Backend API
  - [ ] REST API
  - [ ] GraphQL API
  - [ ] Authentication/Authorization (JWT + RBAC)
  - [ ] WebSocket server for real-time updates
  - [ ] Integration with K8s API
  - [ ] Integration with all NFs

**Weeks 37-40:**
- [ ] WebUI Frontend
  - [ ] Next.js application setup
  - [ ] Dashboard page
  - [ ] NF management pages
  - [ ] Subscriber management pages
  - [ ] Policy management pages
  - [ ] Network topology visualization
  - [ ] Observability pages:
    - [ ] Metrics dashboards
    - [ ] Trace visualization
    - [ ] Log viewer
  - [ ] Real-time updates via WebSockets

### Month 11: Advanced Observability
**Weeks 41-44:**
- [ ] Advanced eBPF tracing
  - [ ] Cross-NF trace correlation
  - [ ] Latency analysis
  - [ ] Bottleneck detection
- [ ] Custom Grafana dashboards
- [ ] Alerting rules
- [ ] Log aggregation and analysis
- [ ] Distributed tracing E2E validation

### Month 12: Testing, Optimization & Hardening
**Weeks 45-48:**
- [ ] Comprehensive integration testing
- [ ] 3GPP compliance testing
- [ ] Performance optimization
- [ ] Security hardening
  - [ ] mTLS between all NFs
  - [ ] Network policies
  - [ ] Secret management (Vault)
  - [ ] Security scanning
- [ ] Load testing (1000+ concurrent UEs)
- [ ] Chaos engineering tests
- [ ] Documentation:
  - [ ] Deployment guide
  - [ ] Operations runbook
  - [ ] API documentation
  - [ ] Architecture decision records

## Development Guidelines

### Code Quality Standards
- **Test Coverage:** Minimum 80% for all services
- **Linting:** Golangci-lint for Go, ESLint for TypeScript
- **Code Review:** All PRs require 2 approvals
- **Documentation:** GoDoc for all public APIs, TSDoc for TypeScript

### 3GPP Compliance
All implementations must adhere to:
- **TS 23.501:** System Architecture
- **TS 23.502:** Procedures
- **TS 23.503:** Policy and Charging
- **TS 29.500:** Technical Realization of Service Based Architecture
- **TS 29.501-29.574:** Specific NF SBI specifications
- **TS 38.300:** NR and NG-RAN Overall Description
- **TS 38.401:** NG-RAN Architecture

### Security Best Practices
- OAuth2/OIDC for external APIs
- mTLS for inter-NF communication
- Kubernetes Network Policies
- RBAC for all APIs
- Secret management (HashiCorp Vault or Kubernetes Secrets)
- Regular security audits
- CVE scanning in CI/CD

### Performance Targets
- **AMF:** 10,000 registrations/second
- **SMF:** 5,000 session establishments/second
- **UPF:** 10+ Gbps throughput (single instance)
- **UDM/UDR:** <10ms p99 latency for subscriber queries
- **NRF:** <5ms p99 latency for discovery requests

### Scalability Requirements
- **Horizontal scaling:** All control plane NFs
- **Stateless design:** Where possible
- **Database sharding:** ClickHouse for >10M subscribers
- **Caching:** Redis for frequently accessed data
- **Load balancing:** Kubernetes Service + Istio

## Repository Structure

```
5g-network/
├── nf/                          # Network Functions
│   ├── amf/                     # Access and Mobility Management
│   ├── smf/                     # Session Management
│   ├── upf/                     # User Plane Function
│   ├── ausf/                    # Authentication Server
│   ├── udm/                     # Unified Data Management
│   ├── udr/                     # Unified Data Repository
│   ├── pcf/                     # Policy Control
│   ├── nrf/                     # Network Repository
│   ├── nssf/                    # Network Slice Selection
│   ├── nef/                     # Network Exposure
│   ├── nwdaf/                   # Network Data Analytics
│   └── gnb/                     # gNodeB (RAN)
├── common/                      # Shared libraries
│   ├── sbi/                     # Service Based Interface framework
│   ├── nas/                     # NAS message encoding/decoding
│   ├── pfcp/                    # PFCP protocol
│   ├── gtp/                     # GTP protocol
│   ├── ngap/                    # NGAP protocol
│   ├── otel/                    # OpenTelemetry instrumentation
│   ├── db/                      # Database clients
│   └── metrics/                 # Metrics clients
├── webui/                       # Management WebUI
│   ├── frontend/                # Next.js application
│   └── backend/                 # Backend API
├── observability/               # Observability stack
│   ├── ebpf/                    # eBPF programs
│   ├── otel-collector/          # OpenTelemetry Collector config
│   └── dashboards/              # Grafana dashboards
├── deploy/                      # Deployment configurations
│   ├── helm/                    # Helm charts
│   ├── k8s/                     # Kubernetes manifests
│   └── terraform/               # Infrastructure as Code
├── test/                        # Testing
│   ├── integration/             # Integration tests
│   ├── e2e/                     # End-to-end tests
│   └── performance/             # Performance tests
├── docs/                        # Documentation
│   ├── architecture/            # Architecture docs
│   ├── api/                     # API documentation
│   └── operations/              # Operations guides
└── tools/                       # Development tools
    ├── ue-simulator/            # UE simulator
    └── traffic-generator/       # Traffic generator
```

## Next Steps

1. **Set up development environment**
   - Provision Kubernetes cluster (kind/k3s for local, EKS/GKE/AKS for cloud)
   - Install observability stack

2. **Create common libraries**
   - SBI framework
   - Protocol implementations
   - OpenTelemetry instrumentation

3. **Begin NF development** (following agent breakdown)
   - Start with NRF (required by all NFs)
   - Parallel development of AMF, AUSF, UDM, UDR

4. **Establish CI/CD pipeline**
   - Automated testing
   - Container image building
   - Deployment to staging

5. **Iterate and integrate**
   - Frequent integration testing
   - Performance benchmarking
   - Security scanning

## Conclusion

This architecture provides a comprehensive, cloud-native, 3GPP-compliant 5G network implementation with:
- ✅ Full network function coverage
- ✅ Scalable containerized deployment
- ✅ Advanced observability with eBPF and OpenTelemetry
- ✅ ClickHouse for subscriber data
- ✅ Victoria Metrics for metrics
- ✅ Comprehensive management WebUI
- ✅ Production-grade security and performance

The multi-agent development approach enables parallel development while maintaining clear ownership and interfaces.
