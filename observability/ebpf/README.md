# eBPF Tracing Programs

This directory contains eBPF programs for kernel-level distributed tracing of the 5G Core Network.

## ⚠️ Important: eBPF is Optional

**The 5G network works perfectly without eBPF!**

eBPF provides advanced kernel-level observability, but you can develop and run everything with application-level tracing.

## Quick Status Check

```bash
# Check if eBPF can compile
make check

# Try to compile (will skip if headers missing)
make all
```

## Requirements

To compile eBPF programs, you need:

### System Requirements
- Linux kernel 5.8+ (check with `uname -r`)
- Kernel headers installed
- Clang 10+ (check with `clang --version`)
- libbpf development files

### Install Dependencies

**Ubuntu/Debian:**
```bash
sudo apt-get update
sudo apt-get install -y \
    clang \
    llvm \
    libbpf-dev \
    linux-headers-$(uname -r) \
    bpftool \
    make
```

**Fedora/RHEL:**
```bash
sudo dnf install -y \
    clang \
    llvm \
    libbpf-devel \
    kernel-devel \
    bpftool \
    make
```

## Compilation

```bash
# Clean previous builds
make clean

# Compile eBPF programs
make all

# Generate vmlinux.h from running kernel
make vmlinux
```

## Files

- `trace_http.c` - HTTP request/response tracing with W3C trace context
- `loader.go` - Go loader that attaches eBPF programs and exports to OpenTelemetry
- `vmlinux.h` - Kernel type definitions (minimal placeholder, regenerate for your kernel)
- `Makefile` - Build configuration

## What eBPF Provides

When properly compiled and loaded:

1. **Kernel-level HTTP tracing** - Capture requests at kernel level
2. **W3C Trace Context** - Extract and propagate traceparent headers
3. **Network packet inspection** - TCP send/recv tracing
4. **Zero application overhead** - Tracing happens in kernel
5. **Automatic context propagation** - No code changes needed

## Development Without eBPF

You can still get comprehensive observability:

- ✅ Application-level OpenTelemetry traces
- ✅ Structured logging (zap)
- ✅ Victoria Metrics
- ✅ Grafana dashboards
- ✅ Distributed tracing (via OpenTelemetry SDK)

The only difference is tracing happens at application level instead of kernel level.

## Usage

Once compiled, the eBPF programs are loaded by network functions:

```go
// In NF main.go
import "github.com/your-org/5g-network/observability/ebpf"

ebpfTracer, err := ebpf.NewEBPFTracer(&ebpf.Config{
    NFName:   "amf",
    NFBinary: binaryPath,
}, logger)

if err := ebpfTracer.Load(ctx); err != nil {
    // Falls back to application-level tracing
    logger.Warn("eBPF not available, using application tracing")
}
```

## Troubleshooting

### "kernel headers not found"
```bash
# Install for your kernel version
sudo apt-get install linux-headers-$(uname -r)
```

### "clang not found"
```bash
sudo apt-get install clang llvm
```

### "libbpf not found"
```bash
sudo apt-get install libbpf-dev
```

### Compilation errors
The Makefile is designed to fail gracefully. If compilation fails, you can still use the 5G network without eBPF.

## Migration Path

1. **Now**: Develop without eBPF (application-level tracing)
2. **Later**: Install kernel headers when needed
3. **Production**: Enable eBPF for advanced observability

No code changes required - just compile and enable!

## More Information

See `/home/silby/5G/EBPF-SETUP.md` for detailed setup guide.
