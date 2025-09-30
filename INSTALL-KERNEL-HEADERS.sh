#!/bin/bash
# Install Kernel Headers for eBPF
# This script installs all necessary dependencies for eBPF compilation

set -e

echo "======================================"
echo "Installing Kernel Headers for eBPF"
echo "======================================"
echo

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

# Check if running as root or with sudo
if [[ $EUID -ne 0 ]]; then
   echo -e "${YELLOW}This script needs sudo privileges.${NC}"
   echo "Please run: sudo $0"
   exit 1
fi

# Get kernel version
KERNEL_VERSION=$(uname -r)
echo -e "${GREEN}Detected kernel version: ${KERNEL_VERSION}${NC}"
echo

# Update package list
echo -e "${GREEN}Step 1: Updating package list...${NC}"
apt-get update

# Install kernel headers
echo
echo -e "${GREEN}Step 2: Installing kernel headers...${NC}"
apt-get install -y linux-headers-${KERNEL_VERSION} || {
    echo -e "${RED}Failed to install kernel headers for ${KERNEL_VERSION}${NC}"
    echo "Available versions:"
    apt-cache search linux-headers | grep ${KERNEL_VERSION%%-*}
    exit 1
}

# Install eBPF development tools
echo
echo -e "${GREEN}Step 3: Installing eBPF development tools...${NC}"
apt-get install -y \
    clang \
    llvm \
    libbpf-dev \
    bpftool \
    make \
    pkg-config \
    gcc \
    libc6-dev

# Verify installation
echo
echo -e "${GREEN}Step 4: Verifying installation...${NC}"
echo

# Check kernel headers
if [ -d "/usr/src/linux-headers-${KERNEL_VERSION}/include" ]; then
    echo -e "  ${GREEN}✓${NC} Kernel headers installed"
    echo "    Location: /usr/src/linux-headers-${KERNEL_VERSION}"
else
    echo -e "  ${RED}✗${NC} Kernel headers NOT found"
fi

# Check clang
if command -v clang &> /dev/null; then
    CLANG_VERSION=$(clang --version | head -1)
    echo -e "  ${GREEN}✓${NC} Clang installed: ${CLANG_VERSION}"
else
    echo -e "  ${RED}✗${NC} Clang NOT found"
fi

# Check llvm
if command -v llc &> /dev/null; then
    echo -e "  ${GREEN}✓${NC} LLVM installed"
else
    echo -e "  ${YELLOW}⚠${NC} LLVM not found"
fi

# Check libbpf
if [ -f "/usr/include/bpf/bpf.h" ]; then
    echo -e "  ${GREEN}✓${NC} libbpf-dev installed"
else
    echo -e "  ${RED}✗${NC} libbpf-dev NOT found"
fi

# Check bpftool
if command -v bpftool &> /dev/null; then
    echo -e "  ${GREEN}✓${NC} bpftool installed"
else
    echo -e "  ${YELLOW}⚠${NC} bpftool not found"
fi

# Check BPF filesystem
if [ -d "/sys/fs/bpf" ]; then
    echo -e "  ${GREEN}✓${NC} BPF filesystem available"
else
    echo -e "  ${YELLOW}⚠${NC} BPF filesystem not mounted"
fi

# Check BTF support
if [ -f "/sys/kernel/btf/vmlinux" ]; then
    echo -e "  ${GREEN}✓${NC} BTF (BPF Type Format) available"
else
    echo -e "  ${YELLOW}⚠${NC} BTF not available (kernel may be too old)"
fi

echo
echo "======================================"
echo -e "${GREEN}✓ Installation complete!${NC}"
echo "======================================"
echo
echo "Next steps:"
echo "  1. Go to eBPF directory:"
echo "     cd /home/silby/5G/observability/ebpf"
echo
echo "  2. Compile eBPF programs:"
echo "     make clean"
echo "     make all"
echo
echo "  3. Should see:"
echo "     ✓ eBPF programs compiled successfully"
echo
echo "If you see errors, check EBPF-SETUP.md for troubleshooting."
echo
