# Final Project Configuration

## Overview

The 5G network project has been configured with a **balanced approach** that combines:
1. **Simulated UPF data plane** for faster development
2. **Basic gNodeB with CU/DU split** for standard architecture
3. **eBPF-based OpenTelemetry tracing** for deep observability

This configuration provides the best mix of development speed, architectural correctness, and production-ready observability.

## Final Architecture Decisions

### 1. UPF Data Plane: **Simulated in Go**

**Implementation:**
- Go-based simulated packet processing
- Clean `DataPlane` interface for future upgrades
- Full PFCP protocol compliance
- GTP-U encapsulation/decapsulation (simulated)
- QoS enforcement (simulated)

**Benefits:**
- ✅ Faster development (no eBPF complexity)
- ✅ Easier debugging
- ✅ Focus on control plane correctness
- ✅ Migration path to eBPF/XDP when needed

### 2. RAN: **Basic gNodeB with CU/DU Split + Simulated Radio**

**Components:**
- **CU (Central Unit):** RRC, PDCP
- **DU (Distributed Unit):** RLC, MAC, High PHY
- **RU (Radio Unit):** Low PHY, RF processing (**simulated**)

**Interfaces:**
- **N2:** CU ↔ AMF (NGAP over SCTP)
- **N3:** CU ↔ UPF (GTP-U)
- **F1:** DU ↔ CU (F1AP over SCTP)
- **Fronthaul:** DU ↔ RU (simulated)

**Simulated Radio:**
- Virtual RF environment
- Channel modeling (path loss, fading)
- UE attachment simulation
- Mobility and handover simulation

**Benefits:**
- ✅ Standard 3GPP CU/DU split architecture
- ✅ No physical RF equipment needed
- ✅ Full RAN protocol stack testable
- ✅ Reproducible test conditions
- ✅ Can upgrade to real RU hardware later

**Why Not O-RAN:**
- Simpler to implement initially
- O-RAN adds complexity (RIC, xApps, rApps, E2 interface)
- Can migrate to O-RAN later if needed (ORAN-ARCHITECTURE.md available for reference)

### 3. Observability: **eBPF-based OpenTelemetry Tracing**

**Architecture:**
```
┌──────────────────────────────────────────────────────┐
│             Network Functions (All NFs)              │
│     (AMF, SMF, UPF, UDM, etc. + gNodeB)             │
└──────────┬───────────────────────────────────────────┘
           │
           │ HTTP/gRPC with W3C Trace Context
           │
┌──────────▼───────────────────────────────────────────┐
│              eBPF Tracing Layer                      │
│  ┌────────────────────────────────────────────────┐ │
│  │  eBPF Programs (kernel-level):                 │ │
│  │  - HTTP request/response capture               │ │
│  │  - Function entry/exit (uprobe/kprobe)         │ │
│  │  - Network packet tracing                      │ │
│  │  - Trace context extraction                    │ │
│  └────────────────────────────────────────────────┘ │
└──────────┬───────────────────────────────────────────┘
           │
           │ eBPF perf events
           │
┌──────────▼───────────────────────────────────────────┐
│          eBPF Exporter (Go)                          │
│  - Reads eBPF maps                                   │
│  - Converts to OpenTelemetry spans                   │
│  - Correlates with application traces                │
└──────────┬───────────────────────────────────────────┘
           │
           │ OTLP
           │
┌──────────▼───────────────────────────────────────────┐
│       OpenTelemetry Collector                        │
│  - Receives eBPF traces                              │
│  - Receives application traces                       │
│  - Merges and correlates                             │
└──────────┬───────────────────────────────────────────┘
           │
           │
┌──────────▼───────────────────────────────────────────┐
│        Tempo/Jaeger (Trace Storage)                  │
│        Grafana (Visualization)                       │
└──────────────────────────────────────────────────────┘
```

**Implementation:**

1. **eBPF Programs (C):**
   ```c
   // Trace HTTP requests at kernel level
   SEC("uprobe/http_handler")
   int trace_http_request(struct pt_regs *ctx) {
       // Extract request details
       // Extract W3C traceparent header
       // Store in eBPF map
   }
   ```

2. **eBPF Loader (Go):**
   ```go
   func LoadeBPFTracer(nfName string) (*eBPFTracer, error) {
       // Load compiled eBPF program
       // Attach to NF process
       // Read perf events
       // Export to OpenTelemetry
   }
   ```

3. **Trace Context Propagation:**
   - W3C Trace Context standard
   - Automatic injection via eBPF
   - Cross-NF correlation
   - End-to-end visibility

**Benefits:**
- ✅ **Kernel-level visibility** without code changes
- ✅ **Automatic trace context** extraction
- ✅ **Complete call flows** across all NFs
- ✅ **Performance insights** (latency at every layer)
- ✅ **Production-ready** observability from day one

## What Each Agent Builds

### Agent 1: AMF + AUSF
- Registration procedures
- Authentication flows
- NAS security
- **eBPF tracing** automatically enabled

### Agent 2: SMF + PCF
- Session management
- Policy control
- **eBPF tracing** automatically enabled

### Agent 3: UPF (Simulated Data Plane)
- **Simulated** packet processing in Go
- PFCP server
- GTP-U tunnel management
- Clean interface for future eBPF/XDP upgrade
- **eBPF tracing** for control plane

### Agent 4: UDM + UDR
- Subscriber management
- ClickHouse integration
- **eBPF tracing** automatically enabled

### Agent 5: NRF + NEF + NSSF
- Service discovery
- API exposure
- Slice selection
- **eBPF tracing** automatically enabled

### Agent 6: NWDAF
- Analytics engine
- ML models
- **eBPF tracing** automatically enabled

### Agent 7: gNodeB (CU/DU/RU)
- **CU:** RRC, PDCP
- **DU:** RLC, MAC, scheduler
- **RU:** **Simulated** radio interface
- Channel modeling
- UE simulation
- **eBPF tracing** for CU and DU

### Agent 8: eBPF-based Observability
- **eBPF programs** for all NFs
- eBPF → OpenTelemetry bridge
- Trace context propagation
- Grafana dashboards
- **This is the critical component!**

### Agent 9-12: WebUI, Infrastructure, Testing
- Management interface
- Kubernetes deployment
- Comprehensive testing

## Complete Technology Stack

| Layer | Technology | Details |
|-------|-----------|---------|
| **Control Plane NFs** | Go 1.22+ | AMF, SMF, UDM, UDR, PCF, NRF, NSSF, NEF, NWDAF |
| **Data Plane** | Go (simulated) | UPF with simulated packet processing |
| **RAN** | Go | CU, DU, RU (simulated radio) |
| **Tracing** | eBPF + C + Go | Kernel-level distributed tracing |
| **Observability** | OpenTelemetry | Unified trace collection |
| **Metrics** | Victoria Metrics | Time-series metrics storage |
| **Subscriber DB** | ClickHouse | Subscriber and policy data |
| **WebUI** | Next.js 14 + React | Management console |
| **Orchestration** | Kubernetes | Container orchestration |

## Development Timeline (48 weeks)

### Phase 1 (Weeks 1-4): Foundation
- Kubernetes setup
- ClickHouse, Victoria Metrics deployment
- **eBPF development environment**
- Common libraries

### Phase 2 (Weeks 5-12): Core NFs
- AMF, AUSF, UDM, UDR, NRF
- Basic UE registration
- **Application-level OTEL** instrumentation

### Phase 3 (Weeks 13-20): Session & Data Plane
- SMF, PCF
- **Simulated UPF**
- PDU session establishment

### Phase 4 (Weeks 21-28): RAN & Advanced NFs
- **gNodeB (CU/DU/RU simulated)**
- NSSF, NEF
- Network slicing

### Phase 5 (Weeks 29-32): Analytics & **eBPF Tracing**
- NWDAF
- **eBPF programs for all NFs**
- **Trace context propagation**
- **Complete observability**

### Phase 6 (Weeks 33-40): Management WebUI
- Frontend and backend
- Full network control

### Phase 7 (Weeks 41-48): Testing & Hardening
- Comprehensive testing
- Performance optimization
- Security hardening

## Key Benefits of This Configuration

### Development Speed
- ✅ **Simulated UPF** removes eBPF complexity for data plane
- ✅ **Basic gNodeB** simpler than O-RAN
- ✅ **Simulated radio** no RF equipment needed

### Production-Ready Observability
- ✅ **eBPF tracing** from day one
- ✅ Kernel-level visibility
- ✅ Complete trace context propagation
- ✅ No code instrumentation required

### 3GPP Compliance
- ✅ All protocols standard-compliant
- ✅ Real PFCP, NGAP, F1AP, GTP-U
- ✅ Full 5G Core functionality

### Migration Path
- ✅ **UPF:** Swap to eBPF/XDP data plane later
- ✅ **RAN:** Add real RU hardware when available
- ✅ **O-RAN:** Can implement later if desired (architecture documented)

## Comparison: Different Approaches

| Aspect | This Config | Full eBPF | O-RAN |
|--------|------------|-----------|-------|
| **UPF** | Simulated (Go) | eBPF/XDP | Simulated |
| **RAN** | CU/DU split | CU/DU split | O-CU/O-DU/O-RU + RIC |
| **Tracing** | **eBPF-based** | **eBPF-based** | App-level |
| **Complexity** | **Medium** | High | High |
| **Dev Speed** | **Fast** | Slow | Slow |
| **Observability** | **Excellent** | **Excellent** | Good |
| **Production Ready** | **Yes** | **Yes** | Yes |

## Testing Capabilities

With this configuration, you can test:
- ✅ **Full UE registration** (simulated UE → gNodeB → AMF → AUSF → UDM)
- ✅ **PDU session establishment** (UE → gNodeB → AMF → SMF → UPF)
- ✅ **Data flows** (simulated packet processing)
- ✅ **Handovers** (simulated mobility)
- ✅ **Network slicing** (multiple S-NSSAIs)
- ✅ **Complete call flows** with **eBPF traces**
- ✅ **1000+ simulated UEs**
- ✅ **All without physical equipment**

## Documentation

### Core Documents
- ✅ **[ARCHITECTURE.md](ARCHITECTURE.md)** - Updated for this config
- ✅ **[RAN-IMPLEMENTATION.md](RAN-IMPLEMENTATION.md)** - gNodeB CU/DU/RU details
- ✅ **[AI-AGENT-GUIDE.md](AI-AGENT-GUIDE.md)** - Agent responsibilities updated
- ✅ **[ROADMAP.md](ROADMAP.md)** - 48-week timeline with eBPF tracing
- ✅ **[TECHNOLOGY-SPECS.md](TECHNOLOGY-SPECS.md)** - Full tech stack

### Reference Documents
- 📚 **[ORAN-ARCHITECTURE.md](ORAN-ARCHITECTURE.md)** - Available if you want O-RAN later
- 📚 **[UPDATES.md](UPDATES.md)** - Previous iteration changes

## Summary

This configuration provides:

✅ **Practical Development**
- Simulated UPF and radio
- No complex eBPF for data plane initially
- Standard gNodeB architecture

✅ **Production Observability**
- eBPF-based distributed tracing
- Kernel-level visibility
- Complete trace context propagation
- Deep insights into all NFs

✅ **Full 5G Functionality**
- All 12 Network Functions
- 3GPP compliant protocols
- Complete call flows
- Network slicing

✅ **Clear Migration Path**
- Upgrade UPF to eBPF/XDP when needed
- Add real RU hardware when available
- Implement O-RAN if desired

## Next Steps

1. **Review Documentation**
   - Read RAN-IMPLEMENTATION.md for gNodeB details
   - Understand eBPF tracing architecture in ARCHITECTURE.md

2. **Begin Development**
   - Follow ROADMAP.md timeline
   - Start with Phase 1: Foundation
   - **Week 31-32:** Focus on eBPF tracing implementation

3. **Build and Test**
   ```bash
   make setup-dev-env
   make create-cluster
   make deploy-dev
   make test-e2e
   ```

**You now have the optimal configuration: fast development + production-ready observability!** 🚀

