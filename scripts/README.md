# Test Scripts

Collection of test scripts for the 5G Core Network implementation.

## Available Scripts

### 1. `test-5g-aka-auth.sh` - 5G-AKA Authentication Flow Test

Tests the complete 5G-AKA authentication flow through all network functions.

**What it tests:**
- AMF → AUSF: Authentication initiation
- AUSF → UDM: Authentication vector generation
- UDM → UDR: Authentication subscription retrieval
- UDR → ClickHouse: Subscriber data query
- Full MILENAGE algorithm execution
- KSEAF key derivation
- Authentication confirmation

**Prerequisites:**
- All required services must be running:
  - NRF (port 8080)
  - UDR (port 8081)
  - UDM (port 8082)
  - AUSF (port 8083)
  - ClickHouse
- Subscriber must exist in ClickHouse with authentication credentials
- `jq` must be installed

**Usage:**

```bash
# Test default subscriber (imsi-001010000000001)
./scripts/test-5g-aka-auth.sh

# Test specific subscriber
./scripts/test-5g-aka-auth.sh imsi-001010000000002

# With custom AUSF URL
AUSF_URL=http://ausf.example.com:8083 ./scripts/test-5g-aka-auth.sh

# With custom serving network
SERVING_NETWORK="5G:mnc002.mcc002.3gppnetwork.org" ./scripts/test-5g-aka-auth.sh
```

**Environment Variables:**
- `AUSF_URL` - AUSF base URL (default: http://localhost:8083)
- `SERVING_NETWORK` - Serving network name (default: 5G:mnc001.mcc001.3gppnetwork.org)

**Example Output:**

```
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  🔐 5G-AKA AUTHENTICATION FLOW TEST
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

SUPI:            imsi-001010000000001
Serving Network: 5G:mnc001.mcc001.3gppnetwork.org
AUSF URL:        http://localhost:8083

✓ AUSF is healthy
✓ Authentication initiated successfully
✓ Retrieved authentication context
✓ Authentication confirmed successfully!

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  ✅ 5G-AKA AUTHENTICATION SUCCESSFUL!
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Authentication Results:
  Result: AUTHENTICATION_SUCCESS
  SUPI:   imsi-001010000000001
  KSEAF:  78f7152cfe8684c39a8f46cddde668a7667de4e9900607a4c56ecc9fac91728a
```

**Exit Codes:**
- `0` - Success (authentication succeeded)
- `1` - Failure (authentication failed or service unavailable)

**Testing Flow:**

The script executes the following steps:

1. **Health Check**: Verifies AUSF is running
2. **Authentication Initiation**: 
   - Sends POST request to `/nausf-auth/v1/ue-authentications`
   - Receives `authCtxId`, `RAND`, and `AUTN`
3. **Get Test Context** (simulates UE):
   - Retrieves `HXRES*` from test endpoint
   - In production, UE would compute `RES*` from `RAND` using its keys
4. **Confirm Authentication**:
   - Sends PUT request with `RES*`
   - Receives authentication result and `KSEAF`

**3GPP Compliance:**
- TS 33.501 - Security architecture and procedures
- TS 29.509 - AUSF services
- TS 35.205-208 - MILENAGE algorithm

## Creating New Test Scripts

When creating new test scripts:

1. Make scripts executable: `chmod +x scripts/your-script.sh`
2. Add shebang: `#!/bin/bash`
3. Use `set -e` for error handling
4. Include help/usage information
5. Provide meaningful exit codes
6. Add colorized output for clarity
7. Validate prerequisites before running
8. Document in this README

## Installation

```bash
# Make all scripts executable
chmod +x scripts/*.sh

# Install jq if not already installed
sudo apt-get install jq
```

## Troubleshooting

**Script fails with "AUSF is not reachable":**
- Check if AUSF is running: `curl http://localhost:8083/health`
- Start AUSF: `./bin/ausf --config nf/ausf/config/ausf.yaml`

**Script fails with "Authentication initiation failed":**
- Check UDM is running: `curl http://localhost:8082/health`
- Check UDR is running: `curl http://localhost:8081/health`
- Verify subscriber exists in ClickHouse

**Script fails with "jq: command not found":**
- Install jq: `sudo apt-get install jq`

**Authentication fails (AUTHENTICATION_FAILURE):**
- Verify subscriber has authentication credentials in UDR
- Check UDR logs: `tail -f /tmp/udr.log`
- Check UDM logs: `tail -f /tmp/udm.log`
- Check AUSF logs: `tail -f /tmp/ausf.log`

## Related Documentation

- [Main README](../README.md)
- [AUSF README](../nf/ausf/README.md)
- [UDM README](../nf/udm/README.md)
- [UDR README](../nf/udr/README.md)
