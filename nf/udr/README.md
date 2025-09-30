# UDR - Unified Data Repository

## Overview

The UDR (Unified Data Repository) is a critical 5G core network function that provides centralized data storage for subscriber information, authentication credentials, session management data, and policy data.

## 3GPP Specifications

- **TS 29.504:** Unified Data Repository Services
- **TS 29.505:** Subscription Data Management
- **TS 29.503:** Authentication Server Services (auth data)
- **TS 29.519:** Policy Data Management

## Features

✅ **Subscriber Data Management**
- Complete subscriber profiles (SUPI, MSISDN, PLMN, etc.)
- Access and Mobility (AM) subscription data
- Session Management (SM) subscription data
- Network slicing support (S-NSSAI)

✅ **Authentication Data**
- 5G-AKA authentication credentials
- Permanent key (K) storage
- OPc/OP management
- SQN (Sequence Number) tracking with atomic increment
- Milenage algorithm support

✅ **Session Management**
- DNN (Data Network Name) configurations
- QoS profiles (5QI, ARP)
- PDU session types (IPv4, IPv6, IPv4v6, Ethernet)
- SSC modes
- Static IP allocation

✅ **Policy Data**
- Subscriber-specific policies
- QoS policies
- Charging characteristics

✅ **Scalability**
- ClickHouse backend for horizontal scaling
- Supports millions of subscribers
- Optimized queries with proper indexing
- ReplacingMergeTree for updates

## Architecture

```
nf/udr/
├── cmd/
│   └── main.go                  # Entry point
├── config/
│   └── udr.yaml                 # Configuration
├── internal/
│   ├── clickhouse/
│   │   ├── client.go            # ClickHouse client
│   │   └── schema.sql           # Database schema
│   ├── config/
│   │   └── config.go            # Configuration management
│   ├── repository/
│   │   ├── models.go            # Data models
│   │   └── repository.go        # Repository implementation
│   └── server/
│       ├── handlers.go          # API handlers
│       └── server.go            # HTTP server
└── README.md
```

## API Endpoints

### Health & Status
- `GET /health` - Health check
- `GET /ready` - Readiness check
- `GET /status` - Service status and statistics

### Subscription Data (3GPP TS 29.505)
- `GET /nudr-dr/v1/subscription-data/{supi}/provisioned-data/am-data` - Get AM data
- `PUT /nudr-dr/v1/subscription-data/{supi}/provisioned-data/am-data` - Update AM data
- `GET /nudr-dr/v1/subscription-data/{supi}/provisioned-data/sm-data` - Get SM data
- `PUT /nudr-dr/v1/subscription-data/{supi}/provisioned-data/sm-data` - Update SM data

### Authentication Data (3GPP TS 29.503)
- `GET /nudr-dr/v1/subscription-data/{supi}/authentication-data/authentication-subscription` - Get auth subscription
- `PUT /nudr-dr/v1/subscription-data/{supi}/authentication-data/authentication-subscription` - Update auth subscription
- `PATCH /nudr-dr/v1/subscription-data/{supi}/authentication-data/authentication-subscription/sqn` - Increment SQN

### Policy Data (3GPP TS 29.519)
- `GET /nudr-dr/v1/policy-data/ues/{supi}/sm-data` - Get policy data
- `PUT /nudr-dr/v1/policy-data/ues/{supi}/sm-data` - Update policy data

### Administrative Endpoints
- `GET /admin/subscribers` - List all subscribers (with pagination)
- `POST /admin/subscribers` - Create subscriber
- `GET /admin/subscribers/{supi}` - Get subscriber details
- `PUT /admin/subscribers/{supi}` - Update subscriber
- `DELETE /admin/subscribers/{supi}` - Delete subscriber
- `GET /admin/stats` - Get repository statistics

## Quick Start

### Prerequisites
- Go 1.21+
- ClickHouse 23.x+
- Network connectivity to ClickHouse

### Installation

```bash
# Build
make build-udr

# Or manually
go build -o bin/udr ./nf/udr/cmd
```

### ClickHouse Setup

1. **Install ClickHouse:**
```bash
# Docker (easiest)
docker run -d --name clickhouse \
  -p 9000:9000 \
  -p 8123:8123 \
  clickhouse/clickhouse-server

# Or use package manager
# Ubuntu/Debian:
sudo apt-get install -y clickhouse-server clickhouse-client
sudo service clickhouse-server start
```

2. **Initialize Schema:**
```bash
# Using UDR binary
./bin/udr --init-schema --config nf/udr/config/udr.yaml

# Or manually with clickhouse-client
clickhouse-client < nf/udr/internal/clickhouse/schema.sql
```

### Running UDR

```bash
# Default configuration
./bin/udr

# Custom configuration
./bin/udr --config /path/to/udr.yaml --log-level debug

# Initialize schema on first run
./bin/udr --init-schema
```

### Configuration

Edit `nf/udr/config/udr.yaml`:

```yaml
clickhouse:
  addresses:
    - localhost:9000    # ClickHouse address
  database: udr
  username: default
  password: ""

sbi:
  scheme: http
  bind_address: 0.0.0.0
  port: 8081

nrf:
  url: http://localhost:8080
  enabled: true
```

## Usage Examples

### Create a Subscriber

```bash
curl -X POST http://localhost:8081/admin/subscribers \
  -H "Content-Type: application/json" \
  -d '{
    "supi": "imsi-001010000000001",
    "supiType": "imsi",
    "plmnId.mcc": "001",
    "plmnId.mnc": "01",
    "subscriberStatus": "ACTIVE",
    "msisdn": "1234567890",
    "subscribedUeAmbr.uplink": "100000000",
    "subscribedUeAmbr.downlink": "200000000",
    "nssai": [
      {"sst": 1, "sd": "000001"}
    ],
    "roamingAllowed": true
  }'
```

### Get Subscriber

```bash
curl http://localhost:8081/admin/subscribers/imsi-001010000000001
```

### Get Authentication Data (for UDM/AUSF)

```bash
curl http://localhost:8081/nudr-dr/v1/subscription-data/imsi-001010000000001/authentication-data/authentication-subscription
```

### Increment SQN (during authentication)

```bash
curl -X PATCH http://localhost:8081/nudr-dr/v1/subscription-data/imsi-001010000000001/authentication-data/authentication-subscription/sqn
```

### Get Statistics

```bash
curl http://localhost:8081/admin/stats
```

## Database Schema

The UDR uses ClickHouse with the following tables:

- **subscribers** - Main subscriber data
- **authentication_subscription** - Authentication credentials
- **session_management_subscription** - SM subscription data per DNN
- **sdm_subscriptions** - Subscriptions for data change notifications
- **policy_data** - Policy data for PCF

All tables use `ReplacingMergeTree` engine for efficient updates.

## Integration with Other NFs

### UDM (Unified Data Management)
UDM uses UDR to:
- Retrieve subscriber data
- Get authentication credentials
- Generate authentication vectors
- Manage SQN for replay protection

### AUSF (Authentication Server)
AUSF calls UDR via UDM to:
- Get authentication subscription data
- Verify authentication credentials

### AMF (Access and Mobility Management)
AMF calls UDM which uses UDR for:
- Subscriber registration
- Access authorization
- Mobility restrictions

### PCF (Policy Control Function)
PCF uses UDR to:
- Retrieve subscriber policies
- Get QoS policies
- Apply charging rules

## Performance

**Targets:**
- < 10ms p99 latency for subscriber queries
- 10,000+ queries per second
- Support for 10M+ subscribers

**Optimizations:**
- ClickHouse columnar storage
- Proper indexing on SUPI
- Connection pooling
- Efficient data models

## Deployment

### Docker

```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /build
COPY . .
RUN go build -o udr ./nf/udr/cmd

FROM alpine:latest
RUN apk --no-cache add ca-certificates
COPY --from=builder /build/udr /app/udr
COPY nf/udr/config/udr.yaml /app/config/
WORKDIR /app
CMD ["./udr", "--config", "./config/udr.yaml"]
```

### Kubernetes

Use the Helm charts in `deploy/helm/5g-core/` with UDR enabled.

## Testing

```bash
# Run tests
make test-udr

# Or manually
go test -v ./nf/udr/...

# With coverage
go test -race -coverprofile=coverage.out ./nf/udr/...
go tool cover -html=coverage.out
```

## Troubleshooting

### ClickHouse Connection Failed
- Ensure ClickHouse is running: `docker ps` or `service clickhouse-server status`
- Check firewall: port 9000 must be accessible
- Verify credentials in `udr.yaml`

### Schema Not Found
- Initialize schema: `./bin/udr --init-schema`
- Or manually: `clickhouse-client < internal/clickhouse/schema.sql`

### High Latency
- Check ClickHouse server resources
- Review query performance with ClickHouse's query log
- Consider adding more ClickHouse nodes for sharding

## Development

### Adding New Data Types

1. Update `schema.sql` with new table
2. Add model in `models.go`
3. Implement repository methods in `repository.go`
4. Add API handlers in `handlers.go`
5. Update routes in `server.go`

### Code Structure

- **clickhouse/** - Database client and schema
- **config/** - Configuration management
- **repository/** - Data access layer
- **server/** - HTTP server and API handlers

## License

Part of the 5G Network project.

## References

- [3GPP TS 29.504](https://www.3gpp.org/ftp/Specs/archive/29_series/29.504/) - Nudr Services
- [3GPP TS 29.505](https://www.3gpp.org/ftp/Specs/archive/29_series/29.505/) - Subscription Data Management
- [ClickHouse Documentation](https://clickhouse.com/docs/en/)
