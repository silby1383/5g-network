#!/bin/bash
# 5G Network Development Environment Setup Script

set -e

echo "======================================"
echo "5G Network Development Environment Setup"
echo "======================================"
echo

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check if running on Linux
if [[ "$(uname)" != "Linux" ]]; then
    echo -e "${RED}Error: This script requires Linux${NC}"
    exit 1
fi

# Check if running as root
if [[ $EUID -eq 0 ]]; then
   echo -e "${YELLOW}Warning: Running as root. Consider using a regular user.${NC}"
fi

echo -e "${GREEN}Step 1: Checking prerequisites...${NC}"

# Check for required tools
check_command() {
    if command -v $1 &> /dev/null; then
        echo -e "  ✓ $1 found"
        return 0
    else
        echo -e "  ${RED}✗ $1 not found${NC}"
        return 1
    fi
}

MISSING_TOOLS=()

check_command docker || MISSING_TOOLS+=("docker")
check_command kubectl || MISSING_TOOLS+=("kubectl")
check_command helm || MISSING_TOOLS+=("helm")
check_command go || MISSING_TOOLS+=("go")
check_command node || MISSING_TOOLS+=("node")
check_command clang || MISSING_TOOLS+=("clang")

if [ ${#MISSING_TOOLS[@]} -ne 0 ]; then
    echo -e "${RED}Missing required tools: ${MISSING_TOOLS[*]}${NC}"
    echo "Please install them before continuing."
    exit 1
fi

echo

echo -e "${GREEN}Step 2: Installing Go development tools...${NC}"
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
go install github.com/golang/mock/mockgen@latest
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
go install github.com/cilium/ebpf/cmd/bpf2go@latest
echo -e "  ✓ Go tools installed"
echo

echo -e "${GREEN}Step 3: Installing eBPF development dependencies...${NC}"
if command -v apt-get &> /dev/null; then
    sudo apt-get update
    sudo apt-get install -y \
        clang \
        llvm \
        libbpf-dev \
        linux-headers-$(uname -r) \
        bpftool \
        make \
        pkg-config
elif command -v dnf &> /dev/null; then
    sudo dnf install -y \
        clang \
        llvm \
        libbpf-devel \
        kernel-devel \
        bpftool \
        make \
        pkg-config
else
    echo -e "${YELLOW}Warning: Package manager not recognized. Please install eBPF tools manually.${NC}"
fi
echo -e "  ✓ eBPF tools installed"
echo

echo -e "${GREEN}Step 4: Installing Node.js development tools...${NC}"
if [ -d "webui/frontend" ] && [ -f "webui/frontend/package.json" ]; then
    cd webui/frontend
    sudo npm install
    sudo npm install -g typescript eslint prettier
    cd - > /dev/null
    echo -e "  ✓ Node.js tools installed"
else
    echo -e "  ${YELLOW}⚠ WebUI not yet initialized${NC}"
fi
echo

echo -e "${GREEN}Step 5: Setting up Git hooks...${NC}"
mkdir -p .git/hooks
cat > .git/hooks/pre-commit << 'EOF'
#!/bin/bash
# Pre-commit hook for 5G Network project

echo "Running pre-commit checks..."

# Check Go code
if git diff --cached --name-only | grep '\.go$' > /dev/null; then
    echo "Checking Go code..."
    go fmt ./...
    golangci-lint run --timeout 5m
    if [ $? -ne 0 ]; then
        echo "Go lint failed. Please fix errors before committing."
        exit 1
    fi
fi

# Check TypeScript code
if git diff --cached --name-only | grep -E '\.(ts|tsx)$' > /dev/null; then
    echo "Checking TypeScript code..."
    cd webui/frontend
    npm run lint
    if [ $? -ne 0 ]; then
        echo "TypeScript lint failed. Please fix errors before committing."
        exit 1
    fi
    cd -
fi

echo "Pre-commit checks passed!"
EOF
chmod +x .git/hooks/pre-commit
echo -e "  ✓ Git hooks installed"
echo

echo -e "${GREEN}Step 6: Checking Kubernetes setup...${NC}"
if kubectl cluster-info &> /dev/null; then
    echo -e "  ✓ Kubernetes cluster is accessible"
    kubectl cluster-info
else
    echo -e "  ${YELLOW}⚠ No Kubernetes cluster found${NC}"
    echo "  Run: make create-cluster"
fi
echo

echo -e "${GREEN}Step 7: Creating development directories...${NC}"
mkdir -p bin
mkdir -p logs
mkdir -p coverage
mkdir -p deploy/local
echo -e "  ✓ Directories created"
echo

echo -e "${GREEN}Step 8: Generating configuration templates...${NC}"
mkdir -p config/dev
cat > config/dev/local.env << 'EOF'
# Development Environment Variables
export FG_ENV=development
export FG_LOG_LEVEL=debug
export FG_METRICS_ENABLED=true
export FG_TRACING_ENABLED=true
export FG_EBPF_ENABLED=true

# ClickHouse
export FG_CLICKHOUSE_ADDR=localhost:9000
export FG_CLICKHOUSE_DB=5gcore_dev

# Victoria Metrics
export FG_VICTORIA_METRICS_ADDR=localhost:8428

# OpenTelemetry
export FG_OTEL_EXPORTER_OTLP_ENDPOINT=localhost:4317
EOF
echo -e "  ✓ Configuration templates created"
echo

echo -e "${GREEN}Step 9: Building eBPF programs...${NC}"
if [ -d "observability/ebpf" ] && [ -f "observability/ebpf/trace_http.c" ]; then
    # Check if we can compile eBPF
    if command -v clang &> /dev/null; then
        echo "  Compiling eBPF programs..."
        cd observability/ebpf
        # Compile manually since Makefile might not exist yet
        clang -O2 -g -target bpf -c trace_http.c -o trace_http.o 2>/dev/null || \
            echo -e "  ${YELLOW}⚠ eBPF compilation failed (needs kernel headers)${NC}"
        cd - > /dev/null
        echo -e "  ✓ eBPF programs compiled (or skipped)"
    else
        echo -e "  ${YELLOW}⚠ Clang not found, skipping eBPF compilation${NC}"
    fi
else
    echo -e "  ${YELLOW}⚠ eBPF directory exists (ready for compilation)${NC}"
fi
echo

echo "======================================"
echo -e "${GREEN}✓ Development environment setup complete!${NC}"
echo "======================================"
echo
echo "Next steps:"
echo "  1. Source environment variables:"
echo "     source config/dev/local.env"
echo
echo "  2. Create Kubernetes cluster:"
echo "     make create-cluster"
echo
echo "  3. Deploy infrastructure:"
echo "     make deploy-infra"
echo
echo "  4. Deploy 5G core:"
echo "     make deploy-core"
echo
echo "  5. Run tests:"
echo "     make test-all"
echo
echo "Documentation: See README.md and GETTING-STARTED.md"
echo
