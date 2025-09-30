# Quick Development Start Guide

## TL;DR - Start in 5 Minutes

```bash
# 1. Read the implementation guide
cat AI-AGENT-GUIDE.md

# 2. Start with NRF
mkdir -p nf/nrf/{cmd,internal/{config,server,repository,discovery}}

# 3. Copy the entry point pattern
cp nf/amf/cmd/main.go nf/nrf/cmd/main.go

# 4. Edit and customize for NRF
# (See AI-AGENT-GUIDE.md for details)

# 5. Build
make build-nrf

# 6. Run
./bin/nrf
```

## Recommended Path for New Developers

### Week 1: NRF (Network Repository Function)

**Why start here?** NRF is the foundation - all other NFs register with it.

**What you'll build:**
- NF registration endpoint
- NF discovery service
- Health monitoring
- Subscription management

**Steps:**

1. **Read documentation (30 min)**
   ```bash
   # Essential reading
   grep -A 200 "Agent 2: NRF" AI-AGENT-GUIDE.md
   ```

2. **Create structure (5 min)**
   ```bash
   mkdir -p nf/nrf/cmd
   mkdir -p nf/nrf/internal/config
   mkdir -p nf/nrf/internal/server
   mkdir -p nf/nrf/internal/repository
   mkdir -p nf/nrf/internal/discovery
   ```

3. **Implement main.go (30 min)**
   - Copy pattern from `nf/amf/cmd/main.go`
   - Initialize logger
   - Load config
   - Start HTTP server
   - Handle shutdown

4. **Implement config (15 min)**
   - Create `internal/config/config.go`
   - Define configuration struct
   - Add YAML loading

5. **Implement HTTP server (1 hour)**
   - Create `internal/server/server.go`
   - Add registration endpoint: `POST /nnrf-nfm/v1/nf-instances/:nfInstanceId`
   - Add discovery endpoint: `GET /nnrf-disc/v1/nf-instances`
   - Add health check: `GET /health`

6. **Implement repository (1 hour)**
   - Create `internal/repository/repository.go`
   - Store NF profiles in memory (map)
   - CRUD operations

7. **Test it (30 min)**
   ```bash
   make build-nrf
   ./bin/nrf --config config/nrf.yaml
   
   # In another terminal, test:
   curl http://localhost:8080/health
   ```

**Total time:** ~4 hours for basic working NRF

### Week 2: AMF (Access and Mobility Management)

**What you'll build:**
- UE registration handling
- NGAP interface to gNodeB
- NAS message processing

**Pattern:** Similar to NRF but with:
- SCTP server for NGAP
- UE context management
- Integration with NRF for discovery

### Week 3: SMF (Session Management Function)

**What you'll build:**
- PDU session creation
- PFCP client to UPF
- IP address allocation

### Week 4+: Other Components

Pick based on interest:
- UPF (if you like data plane)
- AUSF/UDM (if you like security)
- gNodeB (if you like radio)
- WebUI (if you like frontend)

## Development Workflow

### Daily Workflow

```bash
# Morning
git pull
cd /home/silby/5G

# During development
make fmt              # Format code
make lint            # Check for issues
make build-<nf>      # Build your component
./bin/<nf>           # Test locally

# Before commit
make test-unit       # Run tests
make verify          # Full check
git commit -m "feat: your feature"
```

### Code Style

Follow patterns in existing code:

1. **Main entry point:**
   ```go
   // See: nf/amf/cmd/main.go
   package main
   
   import (
       "context"
       "go.uber.org/zap"
   )
   
   func main() {
       logger, _ := zap.NewProduction()
       defer logger.Sync()
       // ...
   }
   ```

2. **HTTP Server:**
   ```go
   // Use chi router
   r := chi.NewRouter()
   r.Get("/health", handleHealth)
   r.Post("/register", handleRegister)
   http.ListenAndServe(":8080", r)
   ```

3. **Configuration:**
   ```go
   type Config struct {
       LogLevel string
       SBI      SBIConfig
   }
   ```

4. **Error handling:**
   ```go
   if err != nil {
       logger.Error("operation failed", zap.Error(err))
       return fmt.Errorf("context: %w", err)
   }
   ```

## Key Files to Reference

### For Structure
- `nf/amf/cmd/main.go` - Entry point pattern
- `nf/gnb/internal/cu/cu.go` - Component structure

### For Interfaces
- `common/dataplane/interface.go` - Clean interface design
- `common/f1/interface.go` - Protocol definitions

### For Implementation
- `nf/upf/internal/dataplane/simulated/simulated.go` - Full implementation example
- `observability/ebpf/loader.go` - Advanced Go patterns

## Testing Strategy

### Unit Tests

```go
// nf/nrf/internal/repository/repository_test.go
package repository

import (
    "testing"
    "github.com/stretchr/testify/assert"
)

func TestRegisterNF(t *testing.T) {
    repo := NewRepository()
    nf := &NFProfile{
        NFInstanceID: "test-123",
        NFType: "AMF",
    }
    
    err := repo.Register(nf)
    assert.NoError(t, err)
    
    retrieved, err := repo.Get("test-123")
    assert.NoError(t, err)
    assert.Equal(t, nf.NFInstanceID, retrieved.NFInstanceID)
}
```

Run with: `make test-unit`

### Integration Tests

Create in `test/integration/nrf_test.go`:
```go
// Test NRF with real HTTP server
func TestNRFIntegration(t *testing.T) {
    // Start NRF
    // Make HTTP requests
    // Verify responses
}
```

Run with: `make test-integration`

## Common Commands

```bash
# Build specific NF
make build-nrf
make build-amf
make build-smf

# Build all
make build

# Run tests
make test-unit
make test-integration

# Code quality
make fmt
make lint
make verify

# Docker
make docker-build-nrf
make docker-build-all

# Kubernetes
make create-cluster
make deploy-core
make status

# Clean
make clean
make clean-all
```

## Getting Help

### Documentation
1. `AI-AGENT-GUIDE.md` - Detailed implementation guide
2. `ARCHITECTURE.md` - System design
3. `RAN-IMPLEMENTATION.md` - gNodeB specifics

### 3GPP Specs
- [3GPP Portal](https://www.3gpp.org/specifications)
- TS 23.501 - Architecture
- TS 29.510 - NRF
- TS 29.518 - AMF
- TS 29.502 - SMF

### Code Examples
- Existing files in `common/` and `nf/`
- AI-AGENT-GUIDE.md has code snippets

## Tips for Success

1. **Start Simple** - Get basic functionality working first
2. **Follow Patterns** - Use existing code as reference
3. **Test Early** - Write tests as you go
4. **Read Specs** - 3GPP specs are your friend
5. **Iterate** - Don't try to implement everything at once

## What's Already Done

✅ Project structure
✅ Build system (Makefile)
✅ Go modules configured
✅ WebUI foundation
✅ Example implementations
✅ Deployment configs
✅ Documentation

## What You Need to Do

🔲 Implement NRF
🔲 Implement AMF  
🔲 Implement SMF
🔲 Implement UPF (simulated data plane exists)
🔲 Implement AUSF/UDM/UDR
🔲 Implement PCF/NSSF/NEF
🔲 Implement NWDAF
🔲 Complete gNodeB (CU started)
🔲 Complete WebUI (structure exists)
🔲 Write tests
🔲 Create documentation

## Next Steps Right Now

```bash
# 1. Read the guide
cat AI-AGENT-GUIDE.md | less

# 2. Pick NRF section
grep -A 200 "Agent 2: NRF" AI-AGENT-GUIDE.md

# 3. Start coding
mkdir -p nf/nrf/cmd
vim nf/nrf/cmd/main.go

# 4. Build and test
make build-nrf
./bin/nrf
```

---

**You have everything you need. Start building!** 🚀

