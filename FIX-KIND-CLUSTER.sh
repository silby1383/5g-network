#!/bin/bash
# Fix Kind Cluster Creation Issues
# This script cleans up and recreates the kind cluster

set -e

echo "======================================"
echo "Kind Cluster - Cleanup and Recreate"
echo "======================================"
echo

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

# Step 1: Check prerequisites
echo -e "${GREEN}Step 1: Checking prerequisites...${NC}"

if ! command -v kind &> /dev/null; then
    echo -e "${RED}✗ kind is not installed${NC}"
    echo "Install kind: https://kind.sigs.k8s.io/docs/user/quick-start/#installation"
    exit 1
fi
echo "  ✓ kind is installed"

if ! command -v docker &> /dev/null; then
    echo -e "${RED}✗ docker is not installed${NC}"
    exit 1
fi

if ! docker ps &> /dev/null; then
    echo -e "${RED}✗ docker is not running or permission denied${NC}"
    echo "Try: sudo systemctl start docker"
    echo "Or: sudo usermod -aG docker $USER && newgrp docker"
    exit 1
fi
echo "  ✓ docker is running"

if ! command -v kubectl &> /dev/null; then
    echo -e "${YELLOW}⚠ kubectl is not installed${NC}"
    echo "Install kubectl: https://kubernetes.io/docs/tasks/tools/"
fi
echo

# Step 2: Clean up existing cluster
echo -e "${GREEN}Step 2: Cleaning up existing cluster...${NC}"

# Delete cluster if exists
if kind get clusters 2>/dev/null | grep -q "5g-network"; then
    echo "  Deleting existing 5g-network cluster..."
    kind delete cluster --name 5g-network
    echo "  ✓ Cluster deleted"
else
    echo "  No existing cluster found"
fi

# Clean up any leftover containers
echo "  Cleaning up containers..."
docker ps -aq --filter "name=5g-network" | xargs -r docker rm -f 2>/dev/null || true

# Clean up networks
echo "  Cleaning up networks..."
docker network prune -f > /dev/null 2>&1

echo "  ✓ Cleanup complete"
echo

# Step 3: Create new cluster
echo -e "${GREEN}Step 3: Creating new kind cluster...${NC}"

# Try with config first, fall back to simple if it fails
if [ -f "deploy/kind/config.yaml" ]; then
    echo "  Attempting creation with custom config..."
    if kind create cluster --name 5g-network --config deploy/kind/config.yaml 2>&1; then
        echo -e "  ${GREEN}✓ Cluster created with custom config${NC}"
    else
        echo -e "  ${YELLOW}⚠ Custom config failed, trying simple creation...${NC}"
        kind create cluster --name 5g-network
        echo -e "  ${GREEN}✓ Cluster created with default config${NC}"
    fi
else
    echo "  Creating cluster with default config..."
    kind create cluster --name 5g-network
    echo -e "  ${GREEN}✓ Cluster created${NC}"
fi
echo

# Step 4: Verify cluster
echo -e "${GREEN}Step 4: Verifying cluster...${NC}"

# Wait a moment for cluster to be ready
sleep 5

# Check cluster status
if kubectl cluster-info --context kind-5g-network &> /dev/null; then
    echo -e "  ${GREEN}✓ Cluster is running${NC}"
    echo
    kubectl cluster-info --context kind-5g-network
else
    echo -e "  ${RED}✗ Cluster verification failed${NC}"
    exit 1
fi
echo

# Step 5: Create namespaces
echo -e "${GREEN}Step 5: Creating namespaces...${NC}"
kubectl create namespace 5gc --dry-run=client -o yaml | kubectl apply -f -
kubectl create namespace databases --dry-run=client -o yaml | kubectl apply -f -
kubectl create namespace observability --dry-run=client -o yaml | kubectl apply -f -
echo -e "  ${GREEN}✓ Namespaces created${NC}"
echo

echo "======================================"
echo -e "${GREEN}✓ Kind cluster is ready!${NC}"
echo "======================================"
echo
echo "Cluster name: 5g-network"
echo "Context: kind-5g-network"
echo
echo "Next steps:"
echo "  1. Deploy infrastructure:"
echo "     make deploy-infra"
echo
echo "  2. Deploy 5G core:"
echo "     make deploy-core"
echo
echo "  3. Check status:"
echo "     kubectl get nodes"
echo "     kubectl get pods -A"
echo
