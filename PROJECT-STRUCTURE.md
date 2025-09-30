# Project Structure and File Organization

## Overview

This document defines the complete directory structure for the 5G network project, ensuring consistency across all components.

## Root Directory Structure

```
5g-network/
├── .github/                      # GitHub Actions workflows, issue templates
│   ├── workflows/
│   │   ├── ci.yml                # Continuous Integration
│   │   ├── build-images.yml      # Docker image building
│   │   ├── deploy-dev.yml        # Deploy to dev environment
│   │   └── security-scan.yml     # Security scanning
│   └── ISSUE_TEMPLATE/
├── api/                          # API definitions
│   ├── openapi/                  # OpenAPI 3.0 specifications
│   │   ├── amf.yaml
│   │   ├── smf.yaml
│   │   ├── ausf.yaml
│   │   └── ...
│   └── proto/                    # Protocol Buffer definitions
│       ├── common.proto
│       └── internal.proto
├── cmd/                          # Main applications (if monorepo approach)
│   └── <nf-name>/
│       └── main.go
├── common/                       # Shared libraries
│   ├── sbi/                      # Service Based Interface framework
│   │   ├── client.go
│   │   ├── server.go
│   │   ├── middleware/
│   │   │   ├── auth.go
│   │   │   ├── tracing.go
│   │   │   └── logging.go
│   │   └── models/
│   ├── nas/                      # NAS protocol
│   │   ├── encoder.go
│   │   ├── decoder.go
│   │   ├── messages/
│   │   └── security/
│   ├── pfcp/                     # PFCP protocol
│   │   ├── client.go
│   │   ├── server.go
│   │   ├── messages.go
│   │   └── ie/                   # Information Elements
│   ├── gtp/                      # GTP-U protocol
│   │   ├── gtpu.go
│   │   ├── header.go
│   │   └── tunnel.go
│   ├── ngap/                     # NGAP protocol
│   │   ├── encoder.go
│   │   ├── decoder.go
│   │   └── messages/
│   ├── otel/                     # OpenTelemetry utilities
│   │   ├── tracer.go
│   │   ├── metrics.go
│   │   └── propagation.go
│   ├── db/                       # Database clients
│   │   ├── clickhouse/
│   │   │   ├── client.go
│   │   │   └── models.go
│   │   └── postgres/
│   │       └── client.go
│   ├── metrics/                  # Victoria Metrics client
│   │   └── client.go
│   ├── config/                   # Configuration utilities
│   │   ├── loader.go
│   │   └── validator.go
│   ├── logger/                   # Structured logging
│   │   └── logger.go
│   ├── nrf/                      # NRF client (used by all NFs)
│   │   ├── client.go
│   │   ├── discovery.go
│   │   └── registration.go
│   └── utils/
│       ├── crypto.go
│       ├── uuid.go
│       └── conversion.go
├── nf/                           # Network Functions
│   ├── amf/                      # Access and Mobility Management
│   │   ├── cmd/
│   │   │   └── main.go
│   │   ├── internal/
│   │   │   ├── context/
│   │   │   ├── gmm/
│   │   │   ├── nas/
│   │   │   ├── ngap/
│   │   │   ├── sbi/
│   │   │   └── metrics/
│   │   ├── pkg/
│   │   │   └── api/
│   │   ├── config/
│   │   │   ├── config.yaml
│   │   │   └── config.go
│   │   ├── test/
│   │   │   ├── unit/
│   │   │   └── integration/
│   │   ├── Dockerfile
│   │   ├── Makefile
│   │   ├── go.mod
│   │   ├── go.sum
│   │   └── README.md
│   ├── smf/                      # Session Management
│   │   └── [similar structure to amf]
│   ├── upf/                      # User Plane Function
│   │   ├── cmd/
│   │   ├── internal/
│   │   │   ├── context/
│   │   │   ├── pfcp/
│   │   │   ├── dataplane/
│   │   │   │   ├── ebpf/
│   │   │   │   ├── xdp_loader.go
│   │   │   │   └── stats.go
│   │   │   ├── gtpu/
│   │   │   ├── qos/
│   │   │   └── metrics/
│   │   ├── ebpf/                 # eBPF C source files
│   │   │   ├── gtp_decap.c
│   │   │   ├── gtp_encap.c
│   │   │   ├── qos.c
│   │   │   ├── classifier.c
│   │   │   └── Makefile
│   │   ├── config/
│   │   ├── test/
│   │   │   ├── unit/
│   │   │   ├── integration/
│   │   │   └── performance/      # Performance benchmarks
│   │   ├── Dockerfile
│   │   └── README.md
│   ├── ausf/                     # Authentication Server
│   ├── udm/                      # Unified Data Management
│   ├── udr/                      # Unified Data Repository
│   ├── pcf/                      # Policy Control
│   ├── nrf/                      # Network Repository
│   ├── nssf/                     # Network Slice Selection
│   ├── nef/                      # Network Exposure
│   ├── nwdaf/                    # Network Data Analytics
│   │   ├── cmd/
│   │   ├── internal/
│   │   │   ├── analytics/
│   │   │   ├── ml/               # Machine learning models
│   │   │   │   ├── load_prediction/
│   │   │   │   ├── anomaly_detection/
│   │   │   │   └── qos_optimization/
│   │   │   └── metrics/
│   │   ├── python/               # Python ML components
│   │   │   ├── models/
│   │   │   ├── training/
│   │   │   └── inference/
│   │   └── requirements.txt
│   └── gnb/                      # gNodeB (RAN)
│       ├── cmd/
│       ├── internal/
│       │   ├── cu/               # Central Unit
│       │   ├── du/               # Distributed Unit
│       │   ├── ngap/
│       │   ├── gtpu/
│       │   └── rrc/
│       └── config/
├── webui/                        # Management WebUI
│   ├── frontend/                 # Next.js application
│   │   ├── app/                  # Next.js 14 app directory
│   │   │   ├── layout.tsx
│   │   │   ├── page.tsx          # Dashboard
│   │   │   ├── dashboard/
│   │   │   ├── nf-management/
│   │   │   │   ├── page.tsx
│   │   │   │   └── [nf-type]/
│   │   │   ├── subscribers/
│   │   │   │   ├── page.tsx
│   │   │   │   ├── new/
│   │   │   │   └── [id]/
│   │   │   ├── policies/
│   │   │   ├── slices/
│   │   │   ├── observability/
│   │   │   │   ├── metrics/
│   │   │   │   ├── traces/
│   │   │   │   └── logs/
│   │   │   ├── topology/
│   │   │   └── api/              # API routes
│   │   ├── components/
│   │   │   ├── ui/               # Shadcn UI components
│   │   │   ├── dashboard/
│   │   │   ├── nf/
│   │   │   ├── subscriber/
│   │   │   ├── topology/
│   │   │   └── charts/
│   │   ├── lib/
│   │   │   ├── api-client.ts
│   │   │   ├── utils.ts
│   │   │   └── validation.ts
│   │   ├── hooks/
│   │   │   ├── use-nf-status.ts
│   │   │   ├── use-metrics.ts
│   │   │   └── use-websocket.ts
│   │   ├── store/                # Zustand stores
│   │   │   ├── nf-store.ts
│   │   │   ├── subscriber-store.ts
│   │   │   └── auth-store.ts
│   │   ├── styles/
│   │   ├── public/
│   │   ├── package.json
│   │   ├── tsconfig.json
│   │   ├── tailwind.config.ts
│   │   ├── next.config.js
│   │   ├── Dockerfile
│   │   └── README.md
│   └── backend/                  # Backend API (Go)
│       ├── cmd/
│       │   └── main.go
│       ├── internal/
│       │   ├── api/
│       │   │   ├── rest/         # REST API handlers
│       │   │   ├── graphql/      # GraphQL resolvers
│       │   │   └── websocket/    # WebSocket handlers
│       │   ├── auth/
│       │   │   ├── jwt.go
│       │   │   ├── rbac.go
│       │   │   └── middleware.go
│       │   ├── k8s/              # Kubernetes API client
│       │   │   ├── client.go
│       │   │   ├── nf_lifecycle.go
│       │   │   └── deployment.go
│       │   ├── nf/               # NF integration clients
│       │   │   ├── amf_client.go
│       │   │   ├── smf_client.go
│       │   │   └── ...
│       │   ├── subscriber/
│       │   │   ├── service.go
│       │   │   └── repository.go
│       │   └── metrics/
│       ├── config/
│       ├── test/
│       ├── Dockerfile
│       └── README.md
├── observability/                # Observability components
│   ├── ebpf/                     # eBPF tracing programs
│   │   ├── trace_nf.c            # Trace NF function calls
│   │   ├── trace_http.c          # Trace HTTP requests
│   │   ├── trace_context.c       # Extract trace context
│   │   ├── loader/               # eBPF loader (Go)
│   │   └── Makefile
│   ├── otel-collector/
│   │   ├── config.yaml           # OpenTelemetry Collector config
│   │   └── pipelines/
│   ├── dashboards/               # Grafana dashboards
│   │   ├── 5g-core-overview.json
│   │   ├── amf-metrics.json
│   │   ├── smf-metrics.json
│   │   ├── upf-performance.json
│   │   ├── subscriber-analytics.json
│   │   ├── call-flow-traces.json
│   │   └── clickhouse-performance.json
│   ├── alerts/                   # Alerting rules
│   │   ├── prometheus-rules.yaml
│   │   └── victoriametrics-rules.yaml
│   └── exporters/
│       └── custom-exporters/
├── deploy/                       # Deployment configurations
│   ├── helm/                     # Helm charts
│   │   ├── 5g-core/              # Umbrella chart
│   │   │   ├── Chart.yaml
│   │   │   ├── values.yaml
│   │   │   ├── templates/
│   │   │   └── charts/           # Subcharts
│   │   ├── amf/
│   │   │   ├── Chart.yaml
│   │   │   ├── values.yaml
│   │   │   ├── templates/
│   │   │   │   ├── deployment.yaml
│   │   │   │   ├── service.yaml
│   │   │   │   ├── configmap.yaml
│   │   │   │   ├── secret.yaml
│   │   │   │   ├── servicemonitor.yaml
│   │   │   │   ├── hpa.yaml
│   │   │   │   └── networkpolicy.yaml
│   │   │   └── README.md
│   │   ├── smf/
│   │   ├── upf/
│   │   ├── [other NFs]/
│   │   ├── webui/
│   │   ├── clickhouse/
│   │   ├── victoria-metrics/
│   │   └── observability/
│   ├── k8s/                      # Raw Kubernetes manifests
│   │   ├── namespaces/
│   │   ├── base/
│   │   └── overlays/
│   │       ├── dev/
│   │       ├── staging/
│   │       └── production/
│   └── terraform/                # Infrastructure as Code
│       ├── aws/
│       │   ├── eks/
│       │   ├── vpc/
│       │   └── rds/
│       ├── gcp/
│       │   └── gke/
│       └── azure/
│           └── aks/
├── test/                         # Cross-component tests
│   ├── integration/
│   │   ├── registration_test.go
│   │   ├── session_test.go
│   │   ├── mobility_test.go
│   │   └── slicing_test.go
│   ├── e2e/
│   │   ├── scenarios/
│   │   │   ├── basic_flow.yaml
│   │   │   ├── multi_ue.yaml
│   │   │   └── handover.yaml
│   │   └── runner/
│   ├── performance/
│   │   ├── load_test.js         # K6 scripts
│   │   ├── throughput_test.go
│   │   └── latency_test.go
│   └── compliance/
│       └── 3gpp_conformance/
├── tools/                        # Development and testing tools
│   ├── ue-simulator/
│   │   ├── cmd/
│   │   ├── internal/
│   │   │   ├── ue/
│   │   │   ├── nas/
│   │   │   └── scenarios/
│   │   ├── config/
│   │   └── README.md
│   ├── traffic-generator/
│   │   ├── cmd/
│   │   └── internal/
│   ├── pcap-analyzer/
│   │   └── analyze.py
│   ├── db-migration/
│   │   ├── clickhouse/
│   │   │   ├── migrations/
│   │   │   │   ├── 001_initial_schema.sql
│   │   │   │   ├── 002_add_slicing.sql
│   │   │   │   └── ...
│   │   │   └── migrate.go
│   │   └── postgres/
│   ├── code-generator/
│   │   ├── openapi-gen.sh
│   │   └── proto-gen.sh
│   └── dev-env/
│       ├── kind-config.yaml
│       ├── k3d-config.yaml
│       └── setup.sh
├── docs/                         # Documentation
│   ├── architecture/
│   │   ├── overview.md
│   │   ├── control-plane.md
│   │   ├── data-plane.md
│   │   ├── observability.md
│   │   └── security.md
│   ├── api/
│   │   ├── sbi-reference.md
│   │   └── webui-api.md
│   ├── development/
│   │   ├── setup.md
│   │   ├── coding-standards.md
│   │   ├── testing.md
│   │   └── ci-cd.md
│   ├── operations/
│   │   ├── deployment.md
│   │   ├── scaling.md
│   │   ├── monitoring.md
│   │   ├── troubleshooting.md
│   │   └── disaster-recovery.md
│   ├── 3gpp/
│   │   ├── procedures.md
│   │   └── compliance.md
│   └── images/
├── scripts/                      # Utility scripts
│   ├── build-all.sh
│   ├── deploy-dev.sh
│   ├── test-all.sh
│   ├── load-test-data.sh
│   ├── backup-clickhouse.sh
│   └── generate-certs.sh
├── config/                       # Global configurations
│   ├── dev/
│   ├── staging/
│   └── production/
├── .gitignore
├── .golangci.yml                 # Go linter config
├── .editorconfig
├── Makefile                      # Root Makefile
├── go.work                       # Go workspace (if using modules)
├── LICENSE
├── README.md
├── ARCHITECTURE.md               # This architecture doc
├── GETTING-STARTED.md
├── CONTRIBUTING.md
└── CHANGELOG.md
```

## Naming Conventions

### Directories
- Use lowercase with hyphens: `nf-management`, `data-plane`
- Network functions use abbreviations: `amf/`, `smf/`, `upf/`

### Go Files
- Use snake_case: `ue_context.go`, `session_management.go`
- Test files: `*_test.go`
- Main entry points: `main.go`

### TypeScript/React Files
- Components: PascalCase with `.tsx`: `Dashboard.tsx`, `NfStatus.tsx`
- Utilities: camelCase with `.ts`: `apiClient.ts`, `formatters.ts`
- Hooks: `use-` prefix: `use-metrics.ts`

### Configuration Files
- Use lowercase with extension: `config.yaml`, `values.yaml`
- Environment-specific: `config.dev.yaml`, `config.prod.yaml`

## Import Paths

### Go Modules

```go
// Common libraries
import (
    "github.com/your-org/5g-network/common/sbi"
    "github.com/your-org/5g-network/common/nas"
    "github.com/your-org/5g-network/common/otel"
)

// Within a NF
import (
    "github.com/your-org/5g-network/nf/amf/internal/context"
    "github.com/your-org/5g-network/nf/amf/internal/gmm"
)
```

### TypeScript

```typescript
// Absolute imports (configured in tsconfig.json)
import { Button } from '@/components/ui/button'
import { useNfStatus } from '@/hooks/use-nf-status'
import { apiClient } from '@/lib/api-client'
```

## Build Artifacts

All build artifacts should be in `.gitignore`:
- `bin/` - Compiled binaries
- `dist/` - Built assets
- `coverage/` - Test coverage reports
- `*.log` - Log files
- `.next/` - Next.js build
- `node_modules/` - NPM packages

## Configuration Management

### Environment Variables

Use a consistent prefix: `FG_` (Five G)

```bash
# NF identification
FG_NF_NAME=amf-1
FG_NF_TYPE=AMF
FG_NF_INSTANCE_ID=uuid

# Network
FG_SBI_ADDR=0.0.0.0:8080
FG_NGAP_ADDR=0.0.0.0:38412

# Dependencies
FG_NRF_URL=https://nrf:8080
FG_CLICKHOUSE_URL=tcp://clickhouse:9000

# Observability
FG_OTEL_ENDPOINT=otel-collector:4317
FG_METRICS_PORT=9090
FG_LOG_LEVEL=info
```

### Secrets Management

Secrets should never be in Git. Use:
- Kubernetes Secrets
- HashiCorp Vault
- Cloud provider secret managers (AWS Secrets Manager, GCP Secret Manager)

## Summary

This structure provides:
- ✅ Clear separation of concerns
- ✅ Consistent organization across all components
- ✅ Scalable for adding new NFs
- ✅ Easy navigation for developers
- ✅ CI/CD friendly
- ✅ Follows Go, TypeScript, and Kubernetes best practices

