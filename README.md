# 5G Core Network

A cloud-native, 3GPP-compliant 5G Core Network implementation with comprehensive observability, built for scalability and production readiness.

## 🌟 Features

- **Full 5G Core Network Functions**
  - AMF (Access and Mobility Management)
  - SMF (Session Management)
  - UPF (User Plane Function) with simulated data plane
  - AUSF, UDM, UDR (Authentication and Data Management)
  - PCF (Policy Control)
  - NRF (Network Repository)
  - NSSF (Network Slice Selection)
  - NEF (Network Exposure)
  - NWDAF (Network Data Analytics)

- **gNodeB with CU/DU/RU Split**
  - 3GPP-compliant F1 interface
  - Simulated radio interface
  - Support for multiple cells and UEs

- **Advanced Observability**
  - eBPF-based distributed tracing with W3C trace context propagation
  - OpenTelemetry integration across all network functions
  - Victoria Metrics for high-performance metrics storage
  - Grafana dashboards with Tempo for trace visualization
  - ClickHouse for subscriber and session data

- **Cloud-Native Architecture**
  - Kubernetes-native deployment
  - Horizontal auto-scaling
  - Service mesh ready
  - Health checks and readiness probes

- **Management WebUI**
  - Real-time network monitoring
  - Subscriber management
  - Session tracking
  - Analytics dashboards

## 🚀 Quick Start

### Prerequisites

- Linux (kernel 5.8+ for eBPF support)
- Docker (20.10+)
- Kubernetes cluster or kind (0.20+)
- Go (1.21+)
- Node.js (18+)
- Clang/LLVM (15+) for eBPF

### Option 1: Quick Start Script

The fastest way to get started:

```bash
# Clone the repository
git clone https://github.com/your-org/5g-network.git
cd 5g-network

# Run the quick start script
make quick-start
```

This will:
1. Create a local Kubernetes cluster
2. Deploy all infrastructure (ClickHouse, Victoria Metrics, Grafana, etc.)
3. Build and deploy all 5G network functions
4. Load test subscriber data

### Option 2: Manual Setup

1. **Set up development environment:**

```bash
make setup
```

2. **Create Kubernetes cluster:**

```bash
make create-cluster
```

3. **Build Docker images:**

```bash
make docker-build-all
```

4. **Load images into cluster:**

```bash
make load-images
```

5. **Deploy infrastructure:**

```bash
make deploy-infra
```

6. **Deploy 5G core:**

```bash
make deploy-core
```

## 📚 Documentation

- **[GETTING-STARTED.md](GETTING-STARTED.md)** - Detailed getting started guide
- **[ARCHITECTURE.md](ARCHITECTURE.md)** - System architecture overview
- **[AI-AGENT-GUIDE.md](AI-AGENT-GUIDE.md)** - Development guide for AI agents
- **[RAN-IMPLEMENTATION.md](RAN-IMPLEMENTATION.md)** - gNodeB implementation details
- **[ROADMAP.md](ROADMAP.md)** - 48-week development timeline
- **[PROJECT-SUMMARY.md](PROJECT-SUMMARY.md)** - Project overview and navigation

## 🏗️ Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                        Management WebUI                          │
│                      (Next.js + TypeScript)                      │
└─────────────────────────────────────────────────────────────────┘
                              ▲
                              │ HTTP/REST
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                        5G Control Plane                          │
│  ┌────────┐  ┌────────┐  ┌────────┐  ┌────────┐  ┌────────┐   │
│  │  AMF   │  │  SMF   │  │  AUSF  │  │  UDM   │  │  PCF   │   │
│  └────────┘  └────────┘  └────────┘  └────────┘  └────────┘   │
│  ┌────────┐  ┌────────┐  ┌────────┐  ┌────────┐  ┌────────┐   │
│  │  UDR   │  │  NRF   │  │ NSSF   │  │  NEF   │  │ NWDAF  │   │
│  └────────┘  └────────┘  └────────┘  └────────┘  └────────┘   │
└─────────────────────────────────────────────────────────────────┘
                              ▲
                              │ N2/N3
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                           gNodeB                                 │
│  ┌──────────┐      F1      ┌──────────┐    Fronthaul  ┌───────┐│
│  │    CU    │ ◄──────────► │    DU    │ ◄───────────► │   RU  ││
│  │(Control) │              │ (Baseband)│               │ (RF)  ││
│  └──────────┘              └──────────┘               └───────┘│
└─────────────────────────────────────────────────────────────────┘
                              ▲
                              │ Radio Interface (Simulated)
                              ▼
                         [ UE Devices ]

┌─────────────────────────────────────────────────────────────────┐
│                         Data Plane                               │
│  ┌────────────────────────────────────────────────────────────┐ │
│  │  UPF (Simulated Go-based Data Plane)                       │ │
│  │  • GTP-U Encapsulation/Decapsulation                       │ │
│  │  • QoS Enforcement                                          │ │
│  │  • Usage Reporting                                          │ │
│  │  • Migration path to eBPF/XDP                              │ │
│  └────────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│                      Observability Stack                         │
│  ┌──────────────┐  ┌────────────────┐  ┌──────────────────┐    │
│  │ eBPF Tracer  │  │ Victoria       │  │   ClickHouse     │    │
│  │ + OpenTelemetry│ │   Metrics     │  │   (Subscriber    │    │
│  │   (Traces)   │  │   (Metrics)    │  │    Repository)   │    │
│  └──────────────┘  └────────────────┘  └──────────────────┘    │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │            Grafana + Tempo (Visualization)                │  │
│  └──────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
```

## 🔧 Development

### Building Components

```bash
# Build all network functions
make build

# Build specific NF
make build-amf
make build-smf
make build-upf

# Build gNodeB components
make build-gnb-cu
make build-gnb-du
make build-gnb-ru
```

### Running Tests

```bash
# Run all tests
make test

# Unit tests
make test-unit

# Integration tests
make test-integration

# End-to-end tests
make test-e2e

# Generate coverage report
make test-coverage
```

### Code Quality

```bash
# Format code
make fmt

# Run linters
make lint

# Run go vet
make vet

# Full verification
make verify
```

## 📊 Accessing Services

After deployment, access the services:

### Grafana (Monitoring & Tracing)
```bash
# Port-forward Grafana
make grafana-port-forward
# Open http://localhost:3000
# Default credentials: admin/admin
```

### Management WebUI
```bash
# Port-forward WebUI
make webui-port-forward
# Open http://localhost:8080
```

### ClickHouse (Subscriber Database)
```bash
# Open ClickHouse shell
make clickhouse-shell

# Query subscribers
SELECT * FROM subscribers LIMIT 10;
```

### Logs
```bash
# View AMF logs
make logs-amf

# View SMF logs
make logs-smf

# View UPF logs
make logs-upf
```

## 🐛 Troubleshooting

### Check cluster status
```bash
make status
```

### Check pod logs
```bash
kubectl logs -n 5gc -l app=amf --tail=100
```

### Check eBPF tracing
```bash
# Check if eBPF programs are loaded
sudo bpftool prog list

# Check trace events
kubectl logs -n 5gc -l app=ebpf-tracer
```

### Restart specific NF
```bash
kubectl rollout restart deployment/amf -n 5gc
```

## 🧪 Testing with UE Simulator

```bash
# Run UE simulator
./bin/ue-simulator --config config/ue-simulator.yaml

# Simulate registration
./bin/ue-simulator register --imsi 001010000000001

# Simulate PDU session establishment
./bin/ue-simulator pdu-session --imsi 001010000000001 --dnn internet
```

## 📈 Performance Targets

### Simulated Mode (Current)
- Registration throughput: 1,000 registrations/sec
- PDU session setup: 500 sessions/sec
- Concurrent UEs: 10,000
- End-to-end latency: < 100ms

### Production Mode (Future with eBPF/XDP)
- Registration throughput: 10,000+ registrations/sec
- PDU session setup: 5,000+ sessions/sec
- Concurrent UEs: 100,000+
- Data plane throughput: 100 Gbps+

## 🛣️ Roadmap

See [ROADMAP.md](ROADMAP.md) for the complete 48-week development timeline.

**Current Phase:** Infrastructure & Control Plane (Weeks 1-16)
- ✅ Project setup and architecture
- ✅ Common libraries
- ✅ NRF implementation
- 🔄 AMF implementation (in progress)
- 📅 SMF implementation (upcoming)

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

Please ensure:
- All tests pass (`make test`)
- Code is formatted (`make fmt`)
- Linters pass (`make lint`)
- Documentation is updated

## 📝 License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- 3GPP for the 5G specifications
- The open source community
- eBPF and OpenTelemetry projects

## 📧 Contact

For questions or support, please open an issue or contact the maintainers.

---

**Built with ❤️ for the future of mobile networks**