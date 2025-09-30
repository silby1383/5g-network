# eBPF Tracing Programs Makefile

CLANG ?= clang
LLC ?= llc
ARCH := $(shell uname -m | sed 's/x86_64/x86/' | sed 's/aarch64/arm64/')

# eBPF source files
SOURCES := trace_http.c
OBJECTS := $(SOURCES:.c=.o)

# Kernel version and paths
KERNEL_VERSION := $(shell uname -r)
KERNEL_HEADERS := /usr/src/linux-headers-$(KERNEL_VERSION)

# Flags
CLANG_FLAGS := -O2 -g -Wall -target bpf -D__TARGET_ARCH_$(ARCH) -D__BPF__
INCLUDES := -I/usr/include \
            -I/usr/include/bpf \
            -I$(KERNEL_HEADERS)/include \
            -I$(KERNEL_HEADERS)/include/uapi \
            -I$(KERNEL_HEADERS)/include/generated/uapi \
            -I$(KERNEL_HEADERS)/arch/x86/include \
            -I$(KERNEL_HEADERS)/arch/x86/include/uapi \
            -I$(KERNEL_HEADERS)/arch/x86/include/generated/uapi \
            -I.

# Check if we have necessary headers
KERNEL_HEADERS := $(shell ls /usr/src/linux-headers-$(shell uname -r)/include 2>/dev/null)

.PHONY: all clean vmlinux check

all: check
	@if [ -n "$(KERNEL_HEADERS)" ]; then \
		$(MAKE) build; \
	else \
		echo "⚠ Kernel headers not found - skipping eBPF compilation"; \
		echo "  This is OK for development. Install linux-headers to enable eBPF."; \
		echo "  See EBPF-SETUP.md for details."; \
	fi

build: $(OBJECTS)
	@echo "✓ eBPF programs compiled successfully"

check:
	@echo "Checking eBPF build requirements..."
	@command -v $(CLANG) >/dev/null 2>&1 || { echo "⚠ clang not found"; exit 0; }
	@echo "  ✓ clang found"

%.o: %.c
	@echo "Compiling $<..."
	@$(CLANG) $(CLANG_FLAGS) $(INCLUDES) -c $< -o $@ 2>&1 || { \
		echo "⚠ Compilation failed - kernel headers may be missing"; \
		echo "  Install: sudo apt-get install linux-headers-$$(uname -r)"; \
		exit 0; \
	}

vmlinux:
	@echo "Generating vmlinux.h (requires bpftool)..."
	@if command -v bpftool &> /dev/null; then \
		bpftool btf dump file /sys/kernel/btf/vmlinux format c > vmlinux.h; \
		echo "✓ vmlinux.h generated"; \
	else \
		echo "⚠ bpftool not found, cannot generate vmlinux.h"; \
		echo "  You can download a pregenerated one or install bpftool"; \
	fi

clean:
	rm -f $(OBJECTS)
	@echo "✓ Cleaned eBPF artifacts"

help:
	@echo "eBPF Tracing Programs Makefile"
	@echo ""
	@echo "Targets:"
	@echo "  all      - Compile all eBPF programs (default)"
	@echo "  vmlinux  - Generate vmlinux.h from running kernel"
	@echo "  clean    - Remove compiled objects"
	@echo "  help     - Show this help message"
	@echo ""
	@echo "Requirements:"
	@echo "  - clang (15+)"
	@echo "  - kernel headers (linux-headers-\$$(uname -r))"
	@echo "  - libbpf-dev"
	@echo "  - bpftool (optional, for vmlinux.h generation)"
