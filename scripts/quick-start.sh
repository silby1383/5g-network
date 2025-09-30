#!/bin/bash
# Quick Start Script for 5G Network
# This script sets up everything needed to run the 5G network locally

set -e

echo "========================================="
echo " 5G Network - Quick Start Script"
echo "========================================="
echo

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo -e "${GREEN}This script will:${NC}"
echo "  1. Create a local Kubernetes cluster (kind)"
echo "  2. Deploy infrastructure (ClickHouse, Victoria Metrics, etc.)"
echo "  3. Deploy 5G Core Network Functions"
echo "  4. Deploy Management WebUI"
echo "  5. Load test data"
echo

read -p "Continue? (y/n) " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    exit 1
fi

echo -e "${GREEN}Step 1: Creating Kubernetes cluster...${NC}"
if kind get clusters 2>/dev/null | grep -q "5g-network"; then
    echo "  Cluster already exists"
else
    # Check if kind config exists, otherwise use default
    if [ -f "deploy/kind/config.yaml" ]; then
        kind create cluster --name 5g-network --config deploy/kind/config.yaml
    else
        echo "  Creating cluster with default configuration..."
        kind create cluster --name 5g-network
    fi
    echo "  ✓ Cluster created"
fi
echo

echo -e "${GREEN}Step 2: Deploying ClickHouse...${NC}"
kubectl create namespace databases --dry-run=client -o yaml | kubectl apply -f -
helm repo add clickhouse https://clickhouse.github.io/clickhouse-kubernetes || true
helm repo update
helm upgrade --install clickhouse clickhouse/clickhouse \
    --namespace databases \
    --set persistence.size=10Gi \
    --set replicaCount=1 \
    --wait --timeout 5m
echo "  ✓ ClickHouse deployed"
echo

echo -e "${GREEN}Step 3: Deploying Victoria Metrics...${NC}"
kubectl create namespace observability --dry-run=client -o yaml | kubectl apply -f -
helm repo add vm https://victoriametrics.github.io/helm-charts/ || true
helm repo update
helm upgrade --install victoria-metrics vm/victoria-metrics-single \
    --namespace observability \
    --set server.persistentVolume.size=10Gi \
    --wait --timeout 5m
echo "  ✓ Victoria Metrics deployed"
echo

echo -e "${GREEN}Step 4: Deploying OpenTelemetry Collector...${NC}"
helm repo add open-telemetry https://open-telemetry.github.io/opentelemetry-helm-charts || true
helm repo update
helm upgrade --install otel-collector open-telemetry/opentelemetry-collector \
    --namespace observability \
    --set mode=deployment \
    --wait --timeout 5m
echo "  ✓ OpenTelemetry Collector deployed"
echo

echo -e "${GREEN}Step 5: Deploying Grafana + Tempo...${NC}"
helm repo add grafana https://grafana.github.io/helm-charts || true
helm repo update
helm upgrade --install tempo grafana/tempo \
    --namespace observability \
    --wait --timeout 5m
helm upgrade --install grafana grafana/grafana \
    --namespace observability \
    --set adminPassword=admin \
    --set service.type=NodePort \
    --set service.nodePort=30080 \
    --wait --timeout 5m
echo "  ✓ Grafana + Tempo deployed"
echo

echo -e "${GREEN}Step 6: Initializing ClickHouse database...${NC}"
# Check if clickhouse-client is available
if command -v clickhouse-client &> /dev/null; then
    kubectl port-forward -n databases svc/clickhouse 9000:9000 > /dev/null 2>&1 &
    PF_PID=$!
    sleep 5

    # Create database schema
    clickhouse-client --host localhost --port 9000 --query "CREATE DATABASE IF NOT EXISTS 5gcore" || true

    # Create tables
    clickhouse-client --host localhost --port 9000 --database 5gcore --multiquery << 'EOF'
CREATE TABLE IF NOT EXISTS subscribers (
    supi String,
    imsi String,
    msisdn String,
    subscription_profile_id String,
    subscriber_status Enum('ACTIVE' = 1, 'SUSPENDED' = 2, 'DELETED' = 3),
    created_at DateTime,
    updated_at DateTime
) ENGINE = MergeTree()
ORDER BY (supi)
PARTITION BY toYYYYMM(created_at);

CREATE TABLE IF NOT EXISTS pdu_sessions (
    session_id String,
    supi String,
    dnn String,
    snssai String,
    pdu_session_type Enum('IPv4' = 1, 'IPv6' = 2, 'IPv4v6' = 3),
    ue_ipv4 IPv4,
    upf_id String,
    smf_id String,
    created_at DateTime,
    closed_at Nullable(DateTime)
) ENGINE = MergeTree()
ORDER BY (created_at, session_id)
PARTITION BY toYYYYMM(created_at);
EOF

    kill $PF_PID 2>/dev/null || true
    echo "  ✓ Database initialized"
else
    echo -e "  ${YELLOW}⚠ clickhouse-client not found, skipping database initialization${NC}"
    echo "  You can initialize it later using kubectl exec"
fi
echo

echo -e "${GREEN}Step 7: Building Docker images...${NC}"
if [ -f "Makefile" ]; then
    make docker-build-all || echo -e "  ${YELLOW}⚠ Some images failed to build${NC}"
    echo "  ✓ Image build attempted"
else
    echo -e "  ${YELLOW}⚠ Makefile not found, skipping image build${NC}"
fi
echo

echo -e "${GREEN}Step 8: Loading images into kind cluster...${NC}"
# Load images if they exist
for img in nrf amf smf upf ausf udm udr pcf; do
    if docker image inspect docker.io/5gnetwork/$img:latest &> /dev/null; then
        kind load docker-image docker.io/5gnetwork/$img:latest --name 5g-network 2>/dev/null || \
            echo -e "  ${YELLOW}⚠ Failed to load $img${NC}"
    else
        echo -e "  ${YELLOW}⚠ Image $img:latest not found, skipping${NC}"
    fi
done
echo "  ✓ Image loading attempted"
echo

echo -e "${GREEN}Step 9: Deploying 5G Core Network...${NC}"
kubectl create namespace 5gc --dry-run=client -o yaml | kubectl apply -f -
helm upgrade --install 5g-core deploy/helm/5g-core \
    --namespace 5gc \
    --wait --timeout 10m
echo "  ✓ 5G Core deployed"
echo

echo -e "${GREEN}Step 10: Loading test subscribers...${NC}"
if command -v clickhouse-client &> /dev/null; then
    kubectl port-forward -n databases svc/clickhouse 9000:9000 > /dev/null 2>&1 &
    PF_PID=$!
    sleep 5

    clickhouse-client --host localhost --port 9000 --database 5gcore --multiquery << 'EOF'
INSERT INTO subscribers VALUES
    ('imsi-001010000000001', '001010000000001', '1234567890', 'default', 'ACTIVE', now(), now()),
    ('imsi-001010000000002', '001010000000002', '1234567891', 'default', 'ACTIVE', now(), now()),
    ('imsi-001010000000003', '001010000000003', '1234567892', 'default', 'ACTIVE', now(), now()),
    ('imsi-001010000000004', '001010000000004', '1234567893', 'default', 'ACTIVE', now(), now()),
    ('imsi-001010000000005', '001010000000005', '1234567894', 'default', 'ACTIVE', now(), now());
EOF

    kill $PF_PID 2>/dev/null || true
    echo "  ✓ Test subscribers loaded"
else
    echo -e "  ${YELLOW}⚠ clickhouse-client not found, skipping test data${NC}"
fi
echo

echo "========================================="
echo -e "${GREEN}✓ 5G Network is now running!${NC}"
echo "========================================="
echo
echo "Access points:"
echo "  • Grafana: http://localhost:30080 (admin/admin)"
echo "  • WebUI: kubectl port-forward -n 5gc svc/webui-frontend 3000:3000"
echo
echo "Check status:"
echo "  kubectl get pods -n 5gc"
echo "  kubectl get pods -n databases"
echo "  kubectl get pods -n observability"
echo
echo "View logs:"
echo "  kubectl logs -n 5gc -l app=amf --follow"
echo "  kubectl logs -n 5gc -l app=smf --follow"
echo
echo "Run tests:"
echo "  make test-e2e"
echo
echo "Stop everything:"
echo "  kind delete cluster --name 5g-network"
echo
