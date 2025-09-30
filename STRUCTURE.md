# 5G Network Project Structure

This document provides an overview of the project directory structure.

```
5G/
├── README.md                          # Main project documentation
├── GETTING-STARTED.md                 # Quick start guide
├── ARCHITECTURE.md                    # System architecture
├── AI-AGENT-GUIDE.md                  # AI agent development guide
├── RAN-IMPLEMENTATION.md              # gNodeB implementation details
├── ROADMAP.md                         # 48-week development timeline
├── PROJECT-SUMMARY.md                 # Project overview
├── FINAL-UPDATES.md                   # Latest configuration summary
├── Makefile                           # Build automation
├── go.mod                             # Go module definition
├── go.sum                             # Go module checksums
├── .gitignore                         # Git ignore rules
│
├── common/                            # Shared libraries and interfaces
│   ├── dataplane/                     # Data plane interface (UPF)
│   │   └── interface.go               # Clean abstraction for simulated/eBPF swap
│   ├── f1/                            # F1 interface (CU-DU)
│   │   └── interface.go               # F1AP message types and structures
│   ├── ngap/                          # NGAP interface (AMF-gNodeB)
│   ├── pfcp/                          # PFCP interface (SMF-UPF)
│   ├── sbi/                           # Service-based interface (NF-NF)
│   ├── nas/                           # NAS messages (UE-AMF)
│   ├── logging/                       # Logging utilities
│   ├── metrics/                       # Metrics collection
│   └── tracing/                       # Tracing helpers
│
├── nf/                                # Network Functions
│   ├── nrf/                           # Network Repository Function
│   │   ├── cmd/                       # Main application
│   │   │   └── main.go
│   │   ├── internal/                  # Internal packages
│   │   │   ├── config/                # Configuration
│   │   │   ├── server/                # HTTP server
│   │   │   ├── repository/            # NF registration repository
│   │   │   └── discovery/             # NF discovery
│   │   ├── Dockerfile                 # Container image
│   │   └── README.md
│   │
│   ├── amf/                           # Access and Mobility Management
│   │   ├── cmd/
│   │   │   └── main.go                # AMF entry point with eBPF
│   │   ├── internal/
│   │   │   ├── config/
│   │   │   ├── server/                # HTTP/NGAP servers
│   │   │   ├── context/               # UE and NGAP contexts
│   │   │   ├── gmm/                   # 5GMM (mobility management)
│   │   │   ├── nas/                   # NAS message handling
│   │   │   ├── ngap/                  # NGAP procedures
│   │   │   └── sbi/                   # SBI client/server
│   │   └── Dockerfile
│   │
│   ├── smf/                           # Session Management Function
│   │   ├── cmd/
│   │   ├── internal/
│   │   │   ├── gsm/                   # 5GSM (session management)
│   │   │   ├── pfcp/                  # PFCP client
│   │   │   └── pdu_session/           # PDU session management
│   │   └── Dockerfile
│   │
│   ├── upf/                           # User Plane Function
│   │   ├── cmd/
│   │   ├── internal/
│   │   │   ├── dataplane/             # Data plane implementations
│   │   │   │   └── simulated/         # Simulated Go-based data plane
│   │   │   │       ├── simulated.go   # Main implementation
│   │   │   │       ├── gtp.go         # GTP-U handling
│   │   │   │       ├── qos.go         # QoS enforcement
│   │   │   │       └── stats.go       # Statistics
│   │   │   └── pfcp/                  # PFCP server
│   │   └── Dockerfile
│   │
│   ├── ausf/                          # Authentication Server Function
│   ├── udm/                           # Unified Data Management
│   ├── udr/                           # Unified Data Repository
│   ├── pcf/                           # Policy Control Function
│   ├── nssf/                          # Network Slice Selection
│   ├── nef/                           # Network Exposure Function
│   ├── nwdaf/                         # Network Data Analytics
│   │   ├── cmd/
│   │   ├── internal/
│   │   └── ml_models/                 # Python ML models
│   │
│   └── gnb/                           # gNodeB
│       ├── cmd/
│       │   ├── cu/                    # Central Unit
│       │   ├── du/                    # Distributed Unit
│       │   └── ru/                    # Radio Unit (simulator)
│       ├── internal/
│       │   ├── cu/
│       │   │   └── cu.go              # CU implementation
│       │   ├── du/                    # DU implementation
│       │   ├── ru/                    # RU simulator
│       │   ├── f1/                    # F1 interface
│       │   ├── rrc/                   # RRC layer
│       │   ├── pdcp/                  # PDCP layer
│       │   ├── rlc/                   # RLC layer
│       │   ├── mac/                   # MAC layer
│       │   └── phy/                   # PHY layer (simulated)
│       └── Dockerfile
│
├── observability/                     # Observability infrastructure
│   ├── ebpf/                          # eBPF tracing programs
│   │   ├── trace_http.c               # HTTP tracing with W3C context
│   │   ├── trace_network.c            # Network-level tracing
│   │   ├── loader.go                  # eBPF program loader
│   │   ├── Makefile                   # eBPF build
│   │   └── vmlinux.h                  # Kernel types
│   │
│   ├── otel/                          # OpenTelemetry configuration
│   │   ├── collector.yaml             # OTel Collector config
│   │   └── instrumentation.go         # Instrumentation helpers
│   │
│   └── dashboards/                    # Grafana dashboards
│       ├── 5g-overview.json
│       ├── amf-metrics.json
│       ├── smf-metrics.json
│       ├── upf-metrics.json
│       └── tracing.json
│
├── webui/                             # Management WebUI
│   ├── frontend/                      # Next.js frontend
│   │   ├── app/                       # Next.js 13+ app directory
│   │   ├── components/                # React components
│   │   ├── lib/                       # Utilities
│   │   ├── styles/                    # Tailwind CSS
│   │   ├── package.json
│   │   └── tsconfig.json
│   │
│   └── backend/                       # Backend API (optional)
│       └── api/
│
├── deploy/                            # Deployment configurations
│   ├── helm/                          # Helm charts
│   │   ├── 5g-core/                   # Main 5G core chart
│   │   │   ├── Chart.yaml
│   │   │   ├── values.yaml            # Production-ready values
│   │   │   └── templates/             # K8s manifests
│   │   ├── clickhouse/
│   │   ├── victoria-metrics/
│   │   ├── otel-collector/
│   │   ├── grafana/
│   │   └── tempo/
│   │
│   ├── kind/                          # kind cluster config
│   │   └── config.yaml
│   │
│   └── local/                         # Local development
│       └── docker-compose.yaml
│
├── scripts/                           # Automation scripts
│   ├── setup-dev-env.sh               # Development setup
│   ├── quick-start.sh                 # Quick start script
│   ├── load-test-data.sh              # Load test subscribers
│   └── demo.sh                        # Demo scenario
│
├── config/                            # Configuration files
│   ├── dev/                           # Development configs
│   │   └── local.env
│   ├── staging/                       # Staging configs
│   └── production/                    # Production configs
│
├── test/                              # Tests
│   ├── unit/                          # Unit tests (alongside code)
│   ├── integration/                   # Integration tests
│   ├── e2e/                           # End-to-end tests
│   └── fixtures/                      # Test data
│
├── tools/                             # Development tools
│   ├── ue-simulator/                  # UE simulator for testing
│   ├── traffic-generator/             # Traffic generation
│   └── dev-env/                       # Dev environment utilities
│
├── docs/                              # Additional documentation
│   ├── api/                           # API documentation
│   ├── deployment/                    # Deployment guides
│   └── development/                   # Development guides
│
├── bin/                               # Compiled binaries (gitignored)
├── coverage/                          # Test coverage (gitignored)
└── logs/                              # Log files (gitignored)
```

## Key Design Principles

### 1. **Clean Separation of Concerns**
- Each NF is self-contained
- Common functionality in `common/`
- Clear interfaces between components

### 2. **3GPP Compliance**
- All interfaces follow 3GPP specifications
- Message structures match standards
- Proper procedure flows

### 3. **Observability First**
- eBPF tracing at kernel level
- OpenTelemetry for distributed tracing
- Comprehensive metrics and logging

### 4. **Cloud-Native**
- Containerized components
- Kubernetes deployment
- Horizontal scaling
- Health checks and readiness

### 5. **Migration-Friendly**
- Clean abstractions (e.g., `dataplane.Interface`)
- Simulated implementations can be swapped
- eBPF/XDP data plane ready

## Navigation Guide

- **Starting Development?** → See `GETTING-STARTED.md`
- **Understanding Architecture?** → See `ARCHITECTURE.md`
- **AI Agent Development?** → See `AI-AGENT-GUIDE.md`
- **gNodeB Details?** → See `RAN-IMPLEMENTATION.md`
- **Timeline?** → See `ROADMAP.md`
- **Quick Reference?** → See `PROJECT-SUMMARY.md`

## Component Ownership (AI Agents)

- **Agent 1**: Infrastructure & Common Libraries
- **Agent 2**: NRF
- **Agent 3**: UPF (Simulated Data Plane)
- **Agent 4**: AUSF, UDM, UDR
- **Agent 5**: PCF, NSSF, NEF
- **Agent 6**: NWDAF
- **Agent 7**: gNodeB (CU/DU/RU)
- **Agent 8**: eBPF Observability
- **Agent 9**: AMF
- **Agent 10**: SMF
- **Agent 11**: Management WebUI
