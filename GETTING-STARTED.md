# Getting Started with 5G Network Development

## Prerequisites

### Required Software

1. **Container Runtime**
   - Docker 24.0+ or Podman 4.0+
   
2. **Kubernetes**
   - Option A: Local development
     - kind 0.20+
     - k3d 5.6+
     - minikube 1.32+
   - Option B: Cloud
     - EKS, GKE, or AKS

3. **Development Tools**
   - Go 1.22+
   - Node.js 20+ (for WebUI)
   - Python 3.11+ (for NWDAF)
   - Rust 1.75+ (optional, for UPF)
   - clang/LLVM 15+ (for eBPF)
   
4. **CLI Tools**
   - kubectl
   - helm 3.12+
   - git
   - make
   - jq

5. **Database Clients**
   - ClickHouse client
   - psql (PostgreSQL, if used for NRF)

### System Requirements

- **CPU:** 8+ cores (16+ recommended for full stack)
- **RAM:** 16 GB minimum (32 GB recommended)
- **Disk:** 50 GB free space
- **OS:** Linux (Ubuntu 22.04/24.04, RHEL 9, or similar)

## Quick Start

### 1. Clone Repository

```bash
git clone https://github.com/your-org/5g-network.git
cd 5g-network
```

### 2. Set Up Local Kubernetes Cluster

#### Using kind (Kubernetes in Docker)

```bash
# Create cluster with custom config for 5G
cat <<EOF > kind-config.yaml
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
  - role: control-plane
    extraPortMappings:
      - containerPort: 38412  # AMF NGAP
        hostPort: 38412
        protocol: SCTP
      - containerPort: 2152   # GTP-U
        hostPort: 2152
        protocol: UDP
  - role: worker
  - role: worker
  - role: worker
EOF

kind create cluster --name 5g-network --config kind-config.yaml
```

#### Using k3d

```bash
k3d cluster create 5g-network \
  --agents 3 \
  --port "38412:38412@loadbalancer" \
  --port "2152:2152/udp@loadbalancer"
```

### 3. Install Dependencies

#### Install ClickHouse

```bash
helm repo add clickhouse https://clickhouse.github.io/clickhouse-kubernetes
helm repo update

helm install clickhouse clickhouse/clickhouse \
  --namespace databases \
  --create-namespace \
  --set persistence.size=20Gi \
  --set replicaCount=3
```

Wait for ClickHouse to be ready:
```bash
kubectl wait --for=condition=ready pod -l app=clickhouse -n databases --timeout=300s
```

#### Install Victoria Metrics

```bash
helm repo add vm https://victoriametrics.github.io/helm-charts/
helm repo update

helm install victoria-metrics vm/victoria-metrics-cluster \
  --namespace observability \
  --create-namespace \
  --set vmselect.replicaCount=2 \
  --set vminsert.replicaCount=2 \
  --set vmstorage.replicaCount=2 \
  --set vmstorage.persistentVolume.size=50Gi
```

#### Install OpenTelemetry Collector

```bash
helm repo add open-telemetry https://open-telemetry.github.io/opentelemetry-helm-charts
helm repo update

helm install otel-collector open-telemetry/opentelemetry-collector \
  --namespace observability \
  --set mode=deployment \
  --set config.exporters.otlp.endpoint=tempo:4317
```

#### Install Grafana + Tempo (for distributed tracing)

```bash
# Tempo
helm repo add grafana https://grafana.github.io/helm-charts
helm install tempo grafana/tempo \
  --namespace observability \
  --set tempo.storage.trace.backend=local

# Grafana
helm install grafana grafana/grafana \
  --namespace observability \
  --set adminPassword=admin \
  --set service.type=LoadBalancer \
  --set datasources."datasources\.yaml".apiVersion=1 \
  --set datasources."datasources\.yaml".datasources[0].name=Tempo \
  --set datasources."datasources\.yaml".datasources[0].type=tempo \
  --set datasources."datasources\.yaml".datasources[0].url=http://tempo:3100
```

#### Install Loki (for log aggregation)

```bash
helm install loki grafana/loki-stack \
  --namespace observability \
  --set grafana.enabled=false \
  --set prometheus.enabled=false \
  --set loki.persistence.enabled=true \
  --set loki.persistence.size=10Gi
```

### 4. Set Up Development Environment

#### Go Development

```bash
# Install dependencies
cd 5g-network
go mod download

# Install development tools
make install-dev-tools
```

This will install:
- golangci-lint
- mockgen
- protoc-gen-go
- go test coverage tools

#### eBPF Development

```bash
# Install eBPF toolchain (Ubuntu/Debian)
sudo apt-get update
sudo apt-get install -y \
  clang \
  llvm \
  libbpf-dev \
  linux-headers-$(uname -r) \
  bpftool

# Or on RHEL/Fedora
sudo dnf install -y \
  clang \
  llvm \
  libbpf-devel \
  kernel-devel \
  bpftool
```

#### Node.js Development (for WebUI)

```bash
cd webui/frontend
npm install

# Or with pnpm (faster)
pnpm install
```

### 5. Initialize ClickHouse Database

```bash
# Port forward to ClickHouse
kubectl port-forward -n databases svc/clickhouse 9000:9000 &

# Run schema migration
export CLICKHOUSE_HOST=localhost
export CLICKHOUSE_PORT=9000
make db-migrate
```

This creates all necessary tables for subscribers, policies, sessions, etc.

### 6. Build and Deploy Core Network Functions

#### Build All Images

```bash
# Build all NF images
make build-all

# Or build specific NF
make build-amf
make build-smf
```

#### Deploy to Kubernetes

```bash
# Create namespace for 5G core
kubectl create namespace 5gc

# Deploy NRF (Network Repository Function) first
helm install nrf deploy/helm/nrf \
  --namespace 5gc

# Wait for NRF to be ready
kubectl wait --for=condition=ready pod -l app=nrf -n 5gc --timeout=300s

# Deploy other NFs
make deploy-core
```

This deploys:
- NRF (Network Repository)
- UDR (Unified Data Repository)
- UDM (Unified Data Management)
- AUSF (Authentication Server)
- AMF (Access and Mobility Management)
- SMF (Session Management)
- PCF (Policy Control)
- NSSF (Network Slice Selection)
- NEF (Network Exposure)

### 7. Deploy User Plane

```bash
# Deploy UPF (User Plane Function)
helm install upf deploy/helm/upf \
  --namespace 5gc \
  --set dataplane.type=ebpf \
  --set resources.limits.cpu=4000m \
  --set resources.limits.memory=8Gi
```

### 8. Deploy Management WebUI

```bash
# Build WebUI images
cd webui
make build

# Deploy WebUI backend
helm install webui-backend deploy/helm/webui-backend \
  --namespace 5gc

# Deploy WebUI frontend
helm install webui-frontend deploy/helm/webui-frontend \
  --namespace 5gc \
  --set ingress.enabled=true \
  --set ingress.host=5g-ui.local
```

Access WebUI:
```bash
# Get LoadBalancer IP or NodePort
kubectl get svc -n 5gc webui-frontend

# Or port forward
kubectl port-forward -n 5gc svc/webui-frontend 3000:3000

# Open browser to http://localhost:3000
```

### 9. Verify Deployment

```bash
# Check all pods are running
kubectl get pods -n 5gc

# Expected output:
# NAME                          READY   STATUS    RESTARTS   AGE
# nrf-0                         1/1     Running   0          5m
# udr-0                         1/1     Running   0          4m
# udm-0                         1/1     Running   0          4m
# ausf-0                        1/1     Running   0          4m
# amf-0                         1/1     Running   0          3m
# amf-1                         1/1     Running   0          3m
# amf-2                         1/1     Running   0          3m
# smf-0                         1/1     Running   0          3m
# smf-1                         1/1     Running   0          3m
# pcf-0                         1/1     Running   0          3m
# upf-0                         1/1     Running   0          2m
# webui-backend-0               1/1     Running   0          1m
# webui-frontend-0              1/1     Running   0          1m

# Check NF registration with NRF
kubectl exec -n 5gc nrf-0 -- curl http://localhost:8080/nnrf-nfm/v1/nf-instances | jq
```

### 10. Load Test Data

```bash
# Create test subscribers in ClickHouse
make load-test-subscribers

# This creates:
# - 1000 test subscribers
# - Default subscription profiles
# - Authentication vectors
# - Network slice subscriptions
```

## Development Workflow

### Building a Network Function

Example: Developing AMF

```bash
# Navigate to AMF directory
cd nf/amf

# Run tests
go test ./... -v

# Run with coverage
go test ./... -cover -coverprofile=coverage.out

# View coverage
go tool cover -html=coverage.out

# Build locally
go build -o bin/amf ./cmd/main.go

# Run locally (connects to K8s services)
./bin/amf --config ./config/config.yaml

# Build Docker image
docker build -t 5g/amf:dev .

# Load image into kind cluster
kind load docker-image 5g/amf:dev --name 5g-network

# Deploy to K8s
helm upgrade --install amf deploy/helm/amf \
  --namespace 5gc \
  --set image.tag=dev
```

### Running Integration Tests

```bash
# Run all integration tests
make test-integration

# Run specific NF integration test
make test-integration-amf

# Run E2E tests (full registration + session establishment)
make test-e2e
```

### Debugging

#### View Logs

```bash
# View logs from specific NF
kubectl logs -n 5gc amf-0 --follow

# View logs from all AMF instances
kubectl logs -n 5gc -l app=amf --follow

# Filter logs by level
kubectl logs -n 5gc amf-0 | jq 'select(.level == "error")'
```

#### Port Forwarding for Local Development

```bash
# Forward NRF
kubectl port-forward -n 5gc svc/nrf 8080:8080 &

# Forward UDM
kubectl port-forward -n 5gc svc/udm 8081:8080 &

# Forward ClickHouse
kubectl port-forward -n databases svc/clickhouse 9000:9000 &

# Now you can run AMF locally and it will connect to K8s services
cd nf/amf
go run cmd/main.go --config config/dev-config.yaml
```

#### Exec into Pod

```bash
# Get shell in AMF pod
kubectl exec -it -n 5gc amf-0 -- /bin/sh

# Inside pod, test connectivity
curl http://nrf:8080/nnrf-nfm/v1/nf-instances
```

### Observability

#### Metrics

```bash
# Port forward Grafana
kubectl port-forward -n observability svc/grafana 3000:80

# Open http://localhost:3000
# Default credentials: admin/admin
```

Import dashboards from `observability/dashboards/`:
- 5G Core Overview
- AMF Metrics
- SMF Metrics
- UPF Performance
- ClickHouse Performance

#### Distributed Tracing

```bash
# Port forward Grafana (includes Tempo)
kubectl port-forward -n observability svc/grafana 3000:80

# In Grafana:
# 1. Go to Explore
# 2. Select Tempo datasource
# 3. Search for traces by service name (amf, smf, etc.)
# 4. View call flow visualization
```

#### Logs

```bash
# Port forward Grafana (includes Loki)
kubectl port-forward -n observability svc/grafana 3000:80

# In Grafana:
# 1. Go to Explore
# 2. Select Loki datasource
# 3. Use LogQL queries:
#    {namespace="5gc", app="amf"} |= "error"
#    {namespace="5gc"} |= "registration"
```

## Testing with UE Simulator

### Deploy UE Simulator

```bash
helm install ue-simulator deploy/helm/ue-simulator \
  --namespace 5gc \
  --set ue.count=10 \
  --set ue.autoRegister=true
```

### Manual UE Registration

```bash
# Exec into UE simulator
kubectl exec -it -n 5gc ue-simulator-0 -- /bin/bash

# Register UE
./uesim register \
  --imsi 001010000000001 \
  --key 465B5CE8B199B49FAA5F0A2EE238A6BC \
  --opc E8ED289DEBA952E4283B54E88E6183CA \
  --amf-addr amf.5gc.svc.cluster.local:38412

# Establish PDU session
./uesim session create \
  --imsi 001010000000001 \
  --dnn internet

# Send data
./uesim data send \
  --imsi 001010000000001 \
  --dest 8.8.8.8 \
  --size 1024
```

### Automated Testing

```bash
# Run automated test scenarios
./uesim test run --scenario registration_100_ues
./uesim test run --scenario session_establishment
./uesim test run --scenario mobility_handover
./uesim test run --scenario load_1000_ues
```

## Common Tasks

### Add a New Subscriber

```bash
# Via WebUI: http://localhost:3000/subscribers/new

# Or via CLI:
kubectl exec -n databases clickhouse-0 -- clickhouse-client --query "
INSERT INTO subscribers (supi, imsi, msisdn, subscription_profile_id, subscriber_status, created_at, updated_at)
VALUES ('imsi-001010000000100', '001010000000100', '1234567890', 'default', 'ACTIVE', now(), now())
"
```

### Scale Network Functions

```bash
# Scale AMF to 5 replicas
kubectl scale -n 5gc statefulset/amf --replicas=5

# Or via Helm
helm upgrade amf deploy/helm/amf \
  --namespace 5gc \
  --set replicaCount=5

# Enable autoscaling
kubectl autoscale -n 5gc deployment/amf \
  --min=3 --max=10 --cpu-percent=70
```

### Update Configuration

```bash
# Edit ConfigMap
kubectl edit configmap -n 5gc amf-config

# Restart pods to pick up new config
kubectl rollout restart -n 5gc statefulset/amf
```

### Backup ClickHouse Data

```bash
# Create backup
kubectl exec -n databases clickhouse-0 -- clickhouse-client --query "BACKUP DATABASE default TO Disk('backups', '2025-09-30')"

# Restore backup
kubectl exec -n databases clickhouse-0 -- clickhouse-client --query "RESTORE DATABASE default FROM Disk('backups', '2025-09-30')"
```

## Troubleshooting

### NF Not Registering with NRF

```bash
# Check NRF is accessible
kubectl exec -n 5gc amf-0 -- nslookup nrf.5gc.svc.cluster.local

# Check NRF logs
kubectl logs -n 5gc nrf-0

# Test NRF API
kubectl exec -n 5gc amf-0 -- curl http://nrf:8080/nnrf-nfm/v1/nf-instances
```

### UE Registration Failing

```bash
# Check AMF logs
kubectl logs -n 5gc amf-0 | grep registration

# Check AUSF logs
kubectl logs -n 5gc ausf-0

# Check UDM logs
kubectl logs -n 5gc udm-0

# Verify subscriber exists in ClickHouse
kubectl exec -n databases clickhouse-0 -- clickhouse-client --query "
SELECT * FROM subscribers WHERE imsi='001010000000001'
"

# Check traces in Grafana for full call flow
```

### PDU Session Establishment Failing

```bash
# Check SMF logs
kubectl logs -n 5gc smf-0 | grep session

# Check UPF logs
kubectl logs -n 5gc upf-0

# Check PFCP connectivity (SMF to UPF)
kubectl exec -n 5gc smf-0 -- ping upf-0.upf.5gc.svc.cluster.local

# Verify UPF has capacity
kubectl exec -n 5gc upf-0 -- curl http://localhost:9090/metrics | grep upf_active_sessions
```

### Performance Issues

```bash
# Check CPU/Memory usage
kubectl top pods -n 5gc

# Check UPF throughput
kubectl exec -n 5gc upf-0 -- curl http://localhost:9090/metrics | grep upf_throughput_bps

# Check ClickHouse query performance
kubectl exec -n databases clickhouse-0 -- clickhouse-client --query "
SELECT
    query,
    query_duration_ms,
    read_rows,
    read_bytes
FROM system.query_log
WHERE type = 2  -- Finished queries
ORDER BY query_duration_ms DESC
LIMIT 10
"

# Check Victoria Metrics ingestion rate
kubectl port-forward -n observability svc/victoria-metrics-vmselect 8481:8481
curl http://localhost:8481/select/0/prometheus/api/v1/query?query=rate(vm_rows_inserted_total[5m])
```

## Next Steps

1. **Explore WebUI**: http://localhost:3000
   - View network topology
   - Manage subscribers
   - Monitor metrics and traces

2. **Run E2E Tests**: Validate full call flows
   ```bash
   make test-e2e-all
   ```

3. **Load Testing**: Test with high UE count
   ```bash
   make load-test-1000-ues
   ```

4. **Deploy to Production**: Follow production deployment guide in `docs/operations/production-deployment.md`

5. **Contribute**: See `CONTRIBUTING.md` for development guidelines

## Resources

- **Documentation**: `/docs`
- **API Specs**: `/api/openapi`
- **Helm Charts**: `/deploy/helm`
- **Examples**: `/examples`
- **3GPP Specs**: https://www.3gpp.org/DynaReport/23-series.htm

## Support

- **Issues**: GitHub Issues
- **Discussions**: GitHub Discussions
- **Slack**: #5g-network channel
- **Email**: 5g-dev@your-org.com

