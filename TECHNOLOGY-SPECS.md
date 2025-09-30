# Technology Specifications and Interface Details

## Table of Contents

1. [Technology Stack Details](#technology-stack-details)
2. [Communication Protocols](#communication-protocols)
3. [ClickHouse Implementation](#clickhouse-implementation)
4. [Victoria Metrics Implementation](#victoria-metrics-implementation)
5. [eBPF Tracing Architecture](#ebpf-tracing-architecture)
6. [Trace Context Propagation](#trace-context-propagation)
7. [Security Architecture](#security-architecture)
8. [Performance Specifications](#performance-specifications)

---

## Technology Stack Details

### Programming Languages

#### Go 1.22+
**Usage:** All control plane Network Functions

**Rationale:**
- Excellent performance and concurrency (goroutines)
- Strong standard library for networking
- Native HTTP/2 support
- Easy cross-compilation
- Small binary sizes
- Fast startup times

**Dependencies:**
```go
// Common dependencies across all NFs
require (
    github.com/gin-gonic/gin v1.9.1              // HTTP framework
    github.com/sirupsen/logrus v1.9.3            // Logging
    go.opentelemetry.io/otel v1.21.0             // Tracing
    go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin v0.46.1
    github.com/prometheus/client_golang v1.17.0  // Metrics
    github.com/ClickHouse/clickhouse-go/v2 v2.16.0
    github.com/spf13/viper v1.18.2               // Configuration
    google.golang.org/grpc v1.60.0               // gRPC
    google.golang.org/protobuf v1.31.0           // Protocol Buffers
)
```

#### TypeScript + Next.js 14
**Usage:** Management WebUI

**Technology Stack:**
```json
{
  "dependencies": {
    "next": "^14.1.0",
    "react": "^18.2.0",
    "react-dom": "^18.2.0",
    "@tanstack/react-query": "^5.17.0",
    "zustand": "^4.4.7",
    "@radix-ui/react-*": "latest",
    "tailwindcss": "^3.4.0",
    "zod": "^3.22.4",
    "socket.io-client": "^4.6.0",
    "d3": "^7.8.5",
    "recharts": "^2.10.3",
    "react-flow-renderer": "^10.3.17"
  },
  "devDependencies": {
    "typescript": "^5.3.3",
    "@types/react": "^18.2.48",
    "eslint": "^8.56.0",
    "prettier": "^3.1.1",
    "playwright": "^1.40.1"
  }
}
```

#### Python 3.11+
**Usage:** NWDAF ML components, tooling

**Stack:**
```python
# requirements.txt
numpy==1.26.3
pandas==2.1.4
scikit-learn==1.3.2
tensorflow==2.15.0  # or PyTorch
prometheus-client==0.19.0
clickhouse-driver==0.2.6
grpcio==1.60.0
fastapi==0.108.0
uvicorn==0.25.0
```

#### C + eBPF
**Usage:** UPF data plane, tracing

**Stack:**
- libbpf 1.3+
- BCC (BPF Compiler Collection) for development
- Clang/LLVM 15+ for compilation

### Container Runtime

#### Docker
**Base Images:**
- Go NFs: `golang:1.22-alpine` (builder), `alpine:latest` (runtime)
- Python: `python:3.11-slim`
- WebUI: `node:20-alpine` (builder), `nginx:alpine` (runtime)

**Multi-stage Build Example:**
```dockerfile
# Builder stage
FROM golang:1.22-alpine AS builder
RUN apk add --no-cache git make gcc musl-dev
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-extldflags "-static"' -o nf-binary ./cmd

# Runtime stage
FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /app
COPY --from=builder /build/nf-binary .
COPY --from=builder /build/config ./config
RUN addgroup -g 1000 nf && adduser -D -u 1000 -G nf nf
USER nf
ENTRYPOINT ["./nf-binary"]
```

### Kubernetes

#### Version
- Kubernetes 1.28+
- Helm 3.14+

#### Custom Resource Definitions (CRDs)

```yaml
# Example: PDU Session CRD
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: pdusessions.5g.example.com
spec:
  group: 5g.example.com
  versions:
    - name: v1
      served: true
      storage: true
      schema:
        openAPIV3Schema:
          type: object
          properties:
            spec:
              type: object
              properties:
                supi:
                  type: string
                dnn:
                  type: string
                snssai:
                  type: object
                  properties:
                    sst:
                      type: integer
                    sd:
                      type: string
                sessionType:
                  type: string
                  enum: [IPv4, IPv6, IPv4v6, Ethernet]
            status:
              type: object
              properties:
                state:
                  type: string
                  enum: [Establishing, Active, Modifying, Releasing]
                ueIP:
                  type: string
                upfID:
                  type: string
  scope: Namespaced
  names:
    plural: pdusessions
    singular: pdusession
    kind: PDUSession
    shortNames:
      - pdu
```

---

## Communication Protocols

### Service Based Interface (SBI)

#### HTTP/2 + JSON
**Standard:** 3GPP TS 29.500

**Implementation:**
```go
// Server setup with HTTP/2
import (
    "net/http"
    "golang.org/x/net/http2"
    "golang.org/x/net/http2/h2c"
)

func NewSBIServer(config *Config) *http.Server {
    handler := setupRoutes()
    
    // Wrap with OpenTelemetry
    handler = otelhttp.NewHandler(handler, "sbi-server")
    
    server := &http.Server{
        Addr:    config.SBIAddress,
        Handler: h2c.NewHandler(handler, &http2.Server{}),
    }
    
    return server
}

// Request/Response with trace context
func (c *NFClient) SendRequest(ctx context.Context, method, url string, body interface{}) (*Response, error) {
    ctx, span := otel.Tracer("nf-client").Start(ctx, fmt.Sprintf("SBI.%s", method))
    defer span.End()
    
    // Create request
    req, err := http.NewRequestWithContext(ctx, method, url, marshalBody(body))
    if err != nil {
        return nil, err
    }
    
    // Add 3GPP headers
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Accept", "application/json")
    
    // OpenTelemetry automatically injects traceparent header
    
    // Send request
    resp, err := c.httpClient.Do(req)
    if err != nil {
        span.RecordError(err)
        return nil, err
    }
    defer resp.Body.Close()
    
    // Record response status
    span.SetAttributes(attribute.Int("http.status_code", resp.StatusCode))
    
    return parseResponse(resp)
}
```

**Message Format (3GPP TS 29.518 - AMF Example):**
```json
{
  "nfInstanceId": "3fa85f64-5717-4562-b3fc-2c963f66afa6",
  "nfType": "AMF",
  "nfStatus": "REGISTERED",
  "ipv4Addresses": ["10.0.1.10"],
  "plmnList": [
    {
      "mcc": "001",
      "mnc": "01"
    }
  ],
  "sNssais": [
    {
      "sst": 1,
      "sd": "000001"
    }
  ],
  "amfInfo": {
    "amfRegionId": "01",
    "amfSetId": "001",
    "guamiList": [
      {
        "plmnId": {
          "mcc": "001",
          "mnc": "01"
        },
        "amfId": "010001"
      }
    ]
  }
}
```

### PFCP (Packet Forwarding Control Protocol)

**Standard:** 3GPP TS 29.244

**Transport:** UDP port 8805

**Implementation:**
```go
package pfcp

import (
    "net"
    "encoding/binary"
)

// PFCP Header
type Header struct {
    Version     uint8
    MP          bool  // Message Priority
    S           bool  // SEID present
    MessageType uint8
    Length      uint16
    SEID        uint64  // Only if S=1
    SequenceNumber uint32
}

// PFCP Session Establishment Request
type SessionEstablishmentRequest struct {
    Header          Header
    NodeID          *IE  // Information Element
    CPFSEID         *IE  // CP F-SEID
    CreatePDR       []*IE
    CreateFAR       []*IE
    CreateURR       []*IE
    CreateQER       []*IE
    CreateBAR       *IE
}

// Send PFCP message
func (c *PFCPClient) SendSessionEstablishmentRequest(req *SessionEstablishmentRequest) (*SessionEstablishmentResponse, error) {
    // Encode message
    buf := encodeMessage(req)
    
    // Send via UDP
    _, err := c.conn.WriteToUDP(buf, c.upfAddr)
    if err != nil {
        return nil, err
    }
    
    // Wait for response
    respBuf := make([]byte, 4096)
    n, _, err := c.conn.ReadFromUDP(respBuf)
    if err != nil {
        return nil, err
    }
    
    // Decode response
    resp := decodeSessionEstablishmentResponse(respBuf[:n])
    return resp, nil
}

// PDR (Packet Detection Rule)
type PDR struct {
    PDRID              uint16
    Precedence         uint32
    PDI                *PacketDetectionInfo
    OuterHeaderRemoval *OuterHeaderRemoval
    FARID              uint16
    URRID              []uint32
    QERID              []uint16
}

// FAR (Forwarding Action Rule)
type FAR struct {
    FARID               uint16
    ApplyAction         uint8  // Forward=0x2, Drop=0x1, Buffer=0x4
    ForwardingParameters *ForwardingParameters
    BARID               uint16
}

// QER (QoS Enforcement Rule)
type QER struct {
    QERID               uint16
    QFI                 uint8
    GateStatus          uint8
    MBR                 *MBR
    GBR                 *GBR
    PacketRate          *PacketRate
}
```

### GTP-U (GPRS Tunneling Protocol - User Plane)

**Standard:** 3GPP TS 29.281

**Transport:** UDP port 2152

**Header Format:**
```
 0                   1                   2                   3
 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|  Ver  | PT| *|E|S|PN| Message Type  |          Length           |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|                Tunnel Endpoint Identifier (TEID)               |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|          Sequence Number          |   N-PDU Number  |Extension |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
```

**Go Implementation:**
```go
type GTPUHeader struct {
    Version      uint8
    ProtocolType uint8
    MessageType  uint8
    Length       uint16
    TEID         uint32
    SequenceNumber uint16
    NPDUNumber   uint8
    NextExtensionHeader uint8
}

func (h *GTPUHeader) Encode() []byte {
    buf := make([]byte, 12)
    
    // First byte: Version(3) | PT(1) | Reserved(1) | E(1) | S(1) | PN(1)
    buf[0] = (h.Version << 5) | (h.ProtocolType << 4)
    
    // Message Type
    buf[1] = h.MessageType
    
    // Length
    binary.BigEndian.PutUint16(buf[2:4], h.Length)
    
    // TEID
    binary.BigEndian.PutUint32(buf[4:8], h.TEID)
    
    // Sequence Number
    binary.BigEndian.PutUint16(buf[8:10], h.SequenceNumber)
    
    // N-PDU Number
    buf[10] = h.NPDUNumber
    
    // Next Extension Header
    buf[11] = h.NextExtensionHeader
    
    return buf
}

// Encapsulate IP packet in GTP-U
func EncapsulateGTPU(teid uint32, ipPacket []byte) []byte {
    header := &GTPUHeader{
        Version:      1,
        ProtocolType: 1,
        MessageType:  0xFF,  // G-PDU
        Length:       uint16(len(ipPacket)),
        TEID:         teid,
        SequenceNumber: 0,
    }
    
    buf := header.Encode()
    buf = append(buf, ipPacket...)
    
    return buf
}
```

### NGAP (NG Application Protocol)

**Standard:** 3GPP TS 38.413

**Transport:** SCTP port 38412

**Implementation:**
```go
import (
    "github.com/ishidawataru/sctp"
)

// NGAP Server
type NGAPServer struct {
    listener *sctp.SCTPListener
    ranContexts map[uint32]*RANContext  // Stream ID -> RAN
}

func NewNGAPServer(addr string) (*NGAPServer, error) {
    laddr, _ := sctp.ResolveSCTPAddr("sctp", addr)
    listener, err := sctp.ListenSCTP("sctp", laddr)
    if err != nil {
        return nil, err
    }
    
    return &NGAPServer{
        listener: listener,
        ranContexts: make(map[uint32]*RANContext),
    }, nil
}

func (s *NGAPServer) Run() error {
    for {
        conn, err := s.listener.AcceptSCTP()
        if err != nil {
            return err
        }
        
        go s.handleConnection(conn)
    }
}

func (s *NGAPServer) handleConnection(conn *sctp.SCTPConn) {
    defer conn.Close()
    
    buf := make([]byte, 65536)
    for {
        n, info, err := conn.SCTPRead(buf)
        if err != nil {
            return
        }
        
        // Parse NGAP message
        msg, err := ngap.Decode(buf[:n])
        if err != nil {
            continue
        }
        
        // Handle based on message type
        s.handleNGAPMessage(info.Stream, msg)
    }
}

// NGAP message types
const (
    NGSetupRequest          = 21
    NGSetupResponse         = 22
    InitialUEMessage        = 15
    DownlinkNASTransport    = 4
    UplinkNASTransport      = 46
    InitialContextSetupRequest = 14
    PDUSessionResourceSetupRequest = 29
    HandoverRequired        = 11
)
```

---

## ClickHouse Implementation

### Deployment Architecture

```yaml
# Clustered ClickHouse with replication
clickhouse:
  shards: 3
  replicas: 2
  
  # Each shard has 2 replicas
  # Shard 1: clickhouse-0-0, clickhouse-0-1
  # Shard 2: clickhouse-1-0, clickhouse-1-1
  # Shard 3: clickhouse-2-0, clickhouse-2-1
```

### Go Client Integration

```go
package clickhouse

import (
    "context"
    "crypto/tls"
    "github.com/ClickHouse/clickhouse-go/v2"
    "github.com/ClickHouse/clickhouse-go/v2/lib/driver"
)

type Client struct {
    conn driver.Conn
}

func NewClient(config *Config) (*Client, error) {
    conn, err := clickhouse.Open(&clickhouse.Options{
        Addr: []string{
            "clickhouse-0.clickhouse:9000",
            "clickhouse-1.clickhouse:9000",
            "clickhouse-2.clickhouse:9000",
        },
        Auth: clickhouse.Auth{
            Database: "5gcore",
            Username: config.Username,
            Password: config.Password,
        },
        TLS: &tls.Config{
            InsecureSkipVerify: false,
        },
        Settings: clickhouse.Settings{
            "max_execution_time": 60,
        },
        DialTimeout:      time.Duration(10) * time.Second,
        MaxOpenConns:     10,
        MaxIdleConns:     5,
        ConnMaxLifetime:  time.Hour,
        ConnOpenStrategy: clickhouse.ConnOpenRoundRobin,
    })
    
    if err != nil {
        return nil, err
    }
    
    return &Client{conn: conn}, nil
}

// Subscriber operations
type Subscriber struct {
    SUPI                 string
    IMSI                 string
    MSISDN               string
    SubscriptionProfileID string
    SubscriberStatus     string
    CreatedAt            time.Time
    UpdatedAt            time.Time
}

func (c *Client) GetSubscriber(ctx context.Context, supi string) (*Subscriber, error) {
    var sub Subscriber
    
    err := c.conn.QueryRow(ctx, `
        SELECT supi, imsi, msisdn, subscription_profile_id, subscriber_status, created_at, updated_at
        FROM subscribers
        WHERE supi = ?
    `, supi).Scan(
        &sub.SUPI,
        &sub.IMSI,
        &sub.MSISDN,
        &sub.SubscriptionProfileID,
        &sub.SubscriberStatus,
        &sub.CreatedAt,
        &sub.UpdatedAt,
    )
    
    if err != nil {
        return nil, err
    }
    
    return &sub, nil
}

func (c *Client) CreatePDUSession(ctx context.Context, session *PDUSession) error {
    batch, err := c.conn.PrepareBatch(ctx, `
        INSERT INTO pdu_sessions (
            session_id, supi, dnn, snssai, pdu_session_type,
            ue_ipv4, upf_id, smf_id, created_at
        )
    `)
    if err != nil {
        return err
    }
    
    err = batch.Append(
        session.SessionID,
        session.SUPI,
        session.DNN,
        session.SNSSAI,
        session.Type,
        session.UEIP,
        session.UPFID,
        session.SMFID,
        time.Now(),
    )
    if err != nil {
        return err
    }
    
    return batch.Send()
}

// Bulk insert for performance
func (c *Client) BulkInsertCDRs(ctx context.Context, cdrs []*CDR) error {
    batch, err := c.conn.PrepareBatch(ctx, "INSERT INTO cdrs")
    if err != nil {
        return err
    }
    
    for _, cdr := range cdrs {
        err = batch.Append(
            cdr.CDRID,
            cdr.SUPI,
            cdr.SessionID,
            cdr.StartTime,
            cdr.EndTime,
            cdr.DataVolumeUplink,
            cdr.DataVolumeDownlink,
            // ... more fields
        )
        if err != nil {
            return err
        }
    }
    
    return batch.Send()
}
```

### Materialized Views for Analytics

```sql
-- Create materialized view for session statistics
CREATE MATERIALIZED VIEW session_stats_hourly
ENGINE = SummingMergeTree()
PARTITION BY toYYYYMMDD(hour)
ORDER BY (hour, dnn, snssai)
AS
SELECT
    toStartOfHour(created_at) AS hour,
    dnn,
    snssai,
    count() AS session_count,
    sum(duration_seconds) AS total_duration,
    avg(duration_seconds) AS avg_duration
FROM pdu_sessions
GROUP BY hour, dnn, snssai;

-- Query for dashboard
SELECT
    hour,
    dnn,
    snssai,
    session_count,
    avg_duration
FROM session_stats_hourly
WHERE hour >= now() - INTERVAL 24 HOUR
ORDER BY hour DESC;
```

---

## Victoria Metrics Implementation

### Architecture

```
Victoria Metrics Cluster:
├── vminsert (2 replicas) - Ingestion
├── vmstorage (3 replicas) - Storage
└── vmselect (2 replicas) - Querying
```

### Metrics Export from NFs

```go
package metrics

import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
    "github.com/prometheus/client_golang/prometheus/promhttp"
)

type AMFMetrics struct {
    RegisteredUEs prometheus.Gauge
    RegistrationRequests *prometheus.CounterVec
    RegistrationDuration *prometheus.HistogramVec
    Handovers *prometheus.CounterVec
}

func NewAMFMetrics(instanceID string) *AMFMetrics {
    return &AMFMetrics{
        RegisteredUEs: promauto.NewGauge(prometheus.GaugeOpts{
            Name: "amf_registered_ues",
            Help: "Number of currently registered UEs",
            ConstLabels: prometheus.Labels{
                "amf_instance_id": instanceID,
            },
        }),
        
        RegistrationRequests: promauto.NewCounterVec(
            prometheus.CounterOpts{
                Name: "amf_registration_requests_total",
                Help: "Total number of registration requests",
                ConstLabels: prometheus.Labels{
                    "amf_instance_id": instanceID,
                },
            },
            []string{"result"},  // success, failure
        ),
        
        RegistrationDuration: promauto.NewHistogramVec(
            prometheus.HistogramOpts{
                Name: "amf_registration_duration_seconds",
                Help: "Registration procedure duration",
                Buckets: prometheus.ExponentialBuckets(0.001, 2, 15), // 1ms to ~16s
                ConstLabels: prometheus.Labels{
                    "amf_instance_id": instanceID,
                },
            },
            []string{"registration_type"},  // initial, periodic, mobility
        ),
        
        Handovers: promauto.NewCounterVec(
            prometheus.CounterOpts{
                Name: "amf_handovers_total",
                Help: "Total number of handovers",
                ConstLabels: prometheus.Labels{
                    "amf_instance_id": instanceID,
                },
            },
            []string{"result", "handover_type"},
        ),
    }
}

// Usage in AMF
func (amf *AMF) HandleRegistrationRequest(ctx context.Context, req *RegistrationRequest) error {
    timer := prometheus.NewTimer(amf.metrics.RegistrationDuration.WithLabelValues(req.Type))
    defer timer.ObserveDuration()
    
    err := amf.processRegistration(ctx, req)
    
    if err != nil {
        amf.metrics.RegistrationRequests.WithLabelValues("failure").Inc()
        return err
    }
    
    amf.metrics.RegistrationRequests.WithLabelValues("success").Inc()
    amf.metrics.RegisteredUEs.Inc()
    
    return nil
}
```

### VictoriaMetrics Query Examples

```promql
# Registration rate per second
rate(amf_registration_requests_total{result="success"}[5m])

# P99 registration latency
histogram_quantile(0.99, rate(amf_registration_duration_seconds_bucket[5m]))

# Active sessions per DNN
sum(smf_active_sessions) by (dnn)

# UPF throughput
sum(rate(upf_bytes_processed_total[1m])) by (upf_instance_id) * 8

# Error rate
sum(rate(nf_http_requests_total{status=~"5.."}[5m])) by (nf_type)
/ 
sum(rate(nf_http_requests_total[5m])) by (nf_type)
```

### Remote Write Configuration

```yaml
# Prometheus remote write to Victoria Metrics
remote_write:
  - url: http://vminsert.observability.svc.cluster.local:8480/insert/0/prometheus/
    queue_config:
      max_samples_per_send: 10000
      batch_send_deadline: 5s
      max_shards: 30
```

---

## eBPF Tracing Architecture

### eBPF Programs

#### HTTP Request Tracing

```c
// ebpf/trace_http.c
#include <linux/bpf.h>
#include <linux/ptrace.h>
#include <linux/tcp.h>

#define TRACEPARENT_LEN 55

struct http_event {
    __u32 pid;
    __u32 tid;
    __u64 timestamp_ns;
    char method[8];
    char path[128];
    char traceparent[TRACEPARENT_LEN];
    __u16 status_code;
    __u64 duration_ns;
};

// Map to store events
BPF_PERF_OUTPUT(http_events);

// Map to correlate request/response
BPF_HASH(active_requests, __u64, struct http_event);

// Trace HTTP request start
int trace_http_request(struct pt_regs *ctx) {
    __u64 id = bpf_get_current_pid_tgid();
    __u64 ts = bpf_ktime_get_ns();
    
    struct http_event event = {};
    event.pid = id >> 32;
    event.tid = id;
    event.timestamp_ns = ts;
    
    // Read HTTP method and path from user space
    bpf_probe_read_user_str(&event.method, sizeof(event.method), 
                           (void *)PT_REGS_PARM1(ctx));
    bpf_probe_read_user_str(&event.path, sizeof(event.path), 
                           (void *)PT_REGS_PARM2(ctx));
    
    // Extract traceparent header
    bpf_probe_read_user_str(&event.traceparent, sizeof(event.traceparent),
                           (void *)PT_REGS_PARM3(ctx));
    
    active_requests.update(&id, &event);
    
    return 0;
}

// Trace HTTP response
int trace_http_response(struct pt_regs *ctx) {
    __u64 id = bpf_get_current_pid_tgid();
    __u64 ts = bpf_ktime_get_ns();
    
    struct http_event *event = active_requests.lookup(&id);
    if (!event) {
        return 0;
    }
    
    event->status_code = (__u16)PT_REGS_PARM1(ctx);
    event->duration_ns = ts - event->timestamp_ns;
    
    // Send event to user space
    http_events.perf_submit(ctx, event, sizeof(*event));
    
    active_requests.delete(&id);
    
    return 0;
}
```

#### Network Packet Tracing

```c
// ebpf/trace_packets.c
#include <linux/bpf.h>
#include <linux/if_ether.h>
#include <linux/ip.h>
#include <linux/udp.h>

struct packet_event {
    __u64 timestamp_ns;
    __u32 src_ip;
    __u32 dst_ip;
    __u16 src_port;
    __u16 dst_port;
    __u8 protocol;
    __u32 packet_size;
    char nf_type[16];
    __u32 teid;  // For GTP-U packets
};

BPF_PERF_OUTPUT(packet_events);

SEC("xdp_packet_trace")
int trace_packet(struct xdp_md *ctx) {
    void *data_end = (void *)(long)ctx->data_end;
    void *data = (void *)(long)ctx->data;
    
    struct ethhdr *eth = data;
    if ((void *)(eth + 1) > data_end)
        return XDP_PASS;
    
    if (eth->h_proto != htons(ETH_P_IP))
        return XDP_PASS;
    
    struct iphdr *ip = (void *)(eth + 1);
    if ((void *)(ip + 1) > data_end)
        return XDP_PASS;
    
    struct packet_event event = {};
    event.timestamp_ns = bpf_ktime_get_ns();
    event.src_ip = ip->saddr;
    event.dst_ip = ip->daddr;
    event.protocol = ip->protocol;
    event.packet_size = ntohs(ip->tot_len);
    
    if (ip->protocol == IPPROTO_UDP) {
        struct udphdr *udp = (void *)(ip + 1);
        if ((void *)(udp + 1) > data_end)
            return XDP_PASS;
        
        event.src_port = ntohs(udp->source);
        event.dst_port = ntohs(udp->dest);
        
        // Check if GTP-U (port 2152)
        if (udp->dest == htons(2152)) {
            // Extract TEID from GTP-U header
            __u32 *teid_ptr = (void *)(udp + 1) + 4;
            if ((void *)(teid_ptr + 1) > data_end)
                return XDP_PASS;
            
            event.teid = ntohl(*teid_ptr);
        }
    } else if (ip->protocol == IPPROTO_SCTP) {
        // NGAP traffic (port 38412)
        // Parse SCTP header
    }
    
    packet_events.perf_submit(ctx, &event, sizeof(event));
    
    return XDP_PASS;
}
```

### Go eBPF Loader

```go
package ebpf

import (
    "github.com/cilium/ebpf"
    "github.com/cilium/ebpf/link"
    "github.com/cilium/ebpf/perf"
)

type Tracer struct {
    collection *ebpf.Collection
    readers    []*perf.Reader
    eventChan  chan *TraceEvent
}

func NewTracer(nfType string) (*Tracer, error) {
    // Load compiled eBPF programs
    spec, err := ebpf.LoadCollectionSpec("trace_http.o")
    if err != nil {
        return nil, err
    }
    
    // Load into kernel
    coll, err := ebpf.NewCollection(spec)
    if err != nil {
        return nil, err
    }
    
    // Attach to tracepoints/kprobes
    prog := coll.Programs["trace_http_request"]
    _, err = link.Uprobe("/usr/local/bin/amf", "HandleHTTPRequest", prog, nil)
    if err != nil {
        return nil, err
    }
    
    // Create perf event reader
    rd, err := perf.NewReader(coll.Maps["http_events"], 4096)
    if err != nil {
        return nil, err
    }
    
    tracer := &Tracer{
        collection: coll,
        readers:    []*perf.Reader{rd},
        eventChan:  make(chan *TraceEvent, 1000),
    }
    
    go tracer.readEvents()
    
    return tracer, nil
}

func (t *Tracer) readEvents() {
    for _, rd := range t.readers {
        go func(reader *perf.Reader) {
            for {
                record, err := reader.Read()
                if err != nil {
                    continue
                }
                
                event := parseEvent(record.RawSample)
                
                // Forward to OpenTelemetry
                t.forwardToOTel(event)
            }
        }(rd)
    }
}

func (t *Tracer) forwardToOTel(event *HTTPEvent) {
    // Parse traceparent header
    traceID, spanID, _ := parseTraceparent(event.Traceparent)
    
    // Create span in OTEL
    ctx := trace.ContextWithRemoteSpanContext(
        context.Background(),
        trace.NewSpanContext(trace.SpanContextConfig{
            TraceID: traceID,
            SpanID:  spanID,
        }),
    )
    
    _, span := otel.Tracer("ebpf").Start(ctx, fmt.Sprintf("HTTP %s %s", event.Method, event.Path))
    span.SetAttributes(
        attribute.String("http.method", event.Method),
        attribute.String("http.path", event.Path),
        attribute.Int("http.status_code", int(event.StatusCode)),
        attribute.Int64("http.duration_ns", int64(event.DurationNS)),
    )
    span.End()
}
```

---

## Trace Context Propagation

### W3C Trace Context Standard

**Header Format:**
```
traceparent: 00-{trace-id}-{parent-id}-{trace-flags}
tracestate: vendor1=value1,vendor2=value2

Example:
traceparent: 00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01
```

### Implementation

```go
package tracing

import (
    "context"
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/propagation"
    "go.opentelemetry.io/otel/trace"
)

// Initialize OpenTelemetry
func InitTracing(serviceName, nfType, instanceID string) error {
    // Create trace provider
    tp, err := tracesdk.NewTracerProvider(
        tracesdk.WithBatcher(otlptracegrpc.New(context.Background())),
        tracesdk.WithResource(resource.NewWithAttributes(
            semconv.SchemaURL,
            semconv.ServiceNameKey.String(serviceName),
            attribute.String("nf.type", nfType),
            attribute.String("nf.instance_id", instanceID),
        )),
    )
    if err != nil {
        return err
    }
    
    otel.SetTracerProvider(tp)
    
    // Set W3C Trace Context propagator
    otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
        propagation.TraceContext{},
        propagation.Baggage{},
    ))
    
    return nil
}

// HTTP middleware for trace propagation
func TraceMiddleware(nfType string) gin.HandlerFunc {
    return func(c *gin.Context) {
        // Extract trace context from headers
        ctx := otel.GetTextMapPropagator().Extract(
            c.Request.Context(),
            propagation.HeaderCarrier(c.Request.Header),
        )
        
        // Start span
        ctx, span := otel.Tracer(nfType).Start(
            ctx,
            fmt.Sprintf("%s %s", c.Request.Method, c.Request.URL.Path),
            trace.WithSpanKind(trace.SpanKindServer),
        )
        defer span.End()
        
        // Add attributes
        span.SetAttributes(
            attribute.String("http.method", c.Request.Method),
            attribute.String("http.url", c.Request.URL.String()),
            attribute.String("http.target", c.Request.URL.Path),
        )
        
        // Set context for downstream handlers
        c.Request = c.Request.WithContext(ctx)
        
        c.Next()
        
        // Record response status
        span.SetAttributes(attribute.Int("http.status_code", c.Writer.Status()))
        
        if c.Writer.Status() >= 400 {
            span.SetStatus(codes.Error, "HTTP error")
        }
    }
}

// Example call flow with trace propagation
func (amf *AMF) HandleRegistration(ctx context.Context, req *RegistrationRequest) error {
    // This span is child of HTTP handler span
    ctx, span := otel.Tracer("amf").Start(ctx, "AMF.HandleRegistration")
    defer span.End()
    
    span.SetAttributes(
        attribute.String("suci", req.SUCI),
        attribute.String("registration_type", req.Type),
    )
    
    // Call AUSF - trace context automatically propagated via HTTP headers
    authResult, err := amf.ausfClient.Authenticate(ctx, req.SUCI)
    if err != nil {
        span.RecordError(err)
        return err
    }
    
    // Call UDM - trace context propagated
    subData, err := amf.udmClient.GetSubscriberData(ctx, authResult.SUPI)
    if err != nil {
        span.RecordError(err)
        return err
    }
    
    // All these calls are correlated in distributed trace
    
    return nil
}
```

### Trace Visualization

The resulting trace shows complete call flow:

```
Trace ID: 4bf92f3577b34da6a3ce929d0e0e4736

HTTP POST /namf-comm/v1/ue-contexts                    [AMF]      200ms
└─ AMF.HandleRegistration                              [AMF]      195ms
   ├─ AUSF.Authenticate                                [AUSF]      50ms
   │  └─ UDM.GetAuthVector                             [UDM]       30ms
   │     └─ ClickHouse.Query                           [UDR]       10ms
   ├─ UDM.GetSubscriberData                            [UDM]       40ms
   │  └─ ClickHouse.Query                              [UDR]       15ms
   ├─ PCF.GetAMPolicy                                  [PCF]       30ms
   └─ NSSF.SelectSlice                                 [NSSF]      20ms
```

---

## Security Architecture

### mTLS Between NFs

```go
// Generate certificates (use cert-manager in K8s)
// or tools/generate-certs.sh

// Server configuration
func NewSecureSBIServer(config *Config) (*http.Server, error) {
    cert, err := tls.LoadX509KeyPair(config.CertFile, config.KeyFile)
    if err != nil {
        return nil, err
    }
    
    caCert, err := ioutil.ReadFile(config.CAFile)
    if err != nil {
        return nil, err
    }
    
    caCertPool := x509.NewCertPool()
    caCertPool.AppendCertsFromPEM(caCert)
    
    tlsConfig := &tls.Config{
        Certificates: []tls.Certificate{cert},
        ClientAuth:   tls.RequireAndVerifyClientCert,
        ClientCAs:    caCertPool,
        MinVersion:   tls.VersionTLS13,
    }
    
    server := &http.Server{
        Addr:      config.SBIAddress,
        TLSConfig: tlsConfig,
        Handler:   setupRoutes(),
    }
    
    return server, nil
}

// Client configuration
func NewSecureSBIClient(config *Config) (*http.Client, error) {
    cert, err := tls.LoadX509KeyPair(config.CertFile, config.KeyFile)
    if err != nil {
        return nil, err
    }
    
    caCert, err := ioutil.ReadFile(config.CAFile)
    if err != nil {
        return nil, err
    }
    
    caCertPool := x509.NewCertPool()
    caCertPool.AppendCertsFromPEM(caCert)
    
    tlsConfig := &tls.Config{
        Certificates: []tls.Certificate{cert},
        RootCAs:      caCertPool,
        MinVersion:   tls.VersionTLS13,
    }
    
    transport := &http.Transport{
        TLSClientConfig: tlsConfig,
    }
    
    client := &http.Client{
        Transport: transport,
        Timeout:   30 * time.Second,
    }
    
    return client, nil
}
```

### Authentication & Authorization (WebUI)

```go
// JWT-based authentication
type AuthService struct {
    secretKey []byte
    rbac      *RBACService
}

func (s *AuthService) GenerateToken(user *User) (string, error) {
    claims := jwt.MapClaims{
        "sub":   user.ID,
        "email": user.Email,
        "roles": user.Roles,
        "exp":   time.Now().Add(24 * time.Hour).Unix(),
        "iat":   time.Now().Unix(),
    }
    
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString(s.secretKey)
}

func (s *AuthService) ValidateToken(tokenString string) (*User, error) {
    token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
        return s.secretKey, nil
    })
    
    if err != nil || !token.Valid {
        return nil, err
    }
    
    claims := token.Claims.(jwt.MapClaims)
    
    user := &User{
        ID:    claims["sub"].(string),
        Email: claims["email"].(string),
        Roles: claims["roles"].([]string),
    }
    
    return user, nil
}

// RBAC middleware
func (s *AuthService) RequireRole(roles ...string) gin.HandlerFunc {
    return func(c *gin.Context) {
        user := c.MustGet("user").(*User)
        
        for _, role := range roles {
            if s.rbac.UserHasRole(user, role) {
                c.Next()
                return
            }
        }
        
        c.JSON(403, gin.H{"error": "Insufficient permissions"})
        c.Abort()
    }
}

// Usage
router.GET("/subscribers", authMiddleware, requireRole("admin", "operator"), listSubscribers)
```

---

## Performance Specifications

### Target Metrics

| Component | Metric | Target |
|-----------|--------|--------|
| AMF | Registrations/sec | 10,000 |
| AMF | Registration latency (p99) | <100ms |
| SMF | Sessions/sec | 5,000 |
| SMF | Session setup latency (p99) | <200ms |
| UPF | Throughput | 10+ Gbps |
| UPF | Packet processing latency | <1ms |
| UPF | Concurrent sessions | 100,000+ |
| UDM/UDR | Query latency (p99) | <10ms |
| NRF | Discovery latency (p99) | <5ms |
| ClickHouse | Write throughput | 1M+ rows/sec |
| ClickHouse | Query latency (simple) | <100ms |

### Optimization Techniques

1. **Connection Pooling**
2. **Caching** (Redis for frequently accessed data)
3. **Async Processing** (goroutines, channels)
4. **Batch Operations** (ClickHouse bulk inserts)
5. **eBPF/XDP** for UPF data plane
6. **HTTP/2** multiplexing
7. **gRPC** for internal communication
8. **Database indexing** (ClickHouse bloom filters)

---

This specification provides the foundation for implementing a production-grade, high-performance 5G network system with comprehensive observability and scalability.

