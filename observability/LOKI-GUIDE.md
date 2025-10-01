# Loki Log Aggregation Guide

## ✅ Loki is Now Fully Configured!

All 5G Network Functions are now sending logs to Loki for centralized log aggregation and analysis.

## 🚀 Quick Start

### Access Logs in Grafana

1. **Open Grafana**: http://localhost:3001 (admin/admin)
2. **Go to Dashboards**: Click the menu icon → Dashboards
3. **Open "5G Network Logs"**: Click on the dashboard
4. **Set Time Range**: Last 15 minutes (top-right)

### Or Use Explore

1. **Open Grafana**: http://localhost:3001
2. **Click Explore** (compass icon on left sidebar)
3. **Select Datasource**: Loki
4. **Run LogQL Queries** (see below)

## 📊 LogQL Query Examples

### Basic Queries

```logql
# All 5G network logs
{job=~"5g-.*"}

# Logs from specific NF
{nf_type="NRF"}
{nf_type="AMF"}
{nf_type="SMF"}
{nf_type="UPF"}

# Logs by job
{job="5g-nrf"}
{job="5g-amf"}
```

### Filter by Log Level

```logql
# Error logs from all NFs
{job=~"5g-.*"} | json | level="error"

# Info logs from NRF
{nf_type="NRF"} | json | level="info"

# Warning logs
{job=~"5g-.*"} | json | level="warn"
```

### Search Log Content

```logql
# Logs containing "registered"
{job=~"5g-.*"} |= "registered"

# Logs containing "error" (case insensitive)
{job=~"5g-.*"} |~ "(?i)error"

# NRF logs about heartbeat
{nf_type="NRF"} |= "heartbeat"

# AMF registration logs
{nf_type="AMF"} |= "Registration"

# SMF session logs
{nf_type="SMF"} |= "session"
```

### Advanced Queries

```logql
# Count error logs per NF
sum by (nf_type) (count_over_time({job=~"5g-.*"} | json | level="error" [5m]))

# Log rate by NF type
sum by (nf_type) (rate({job=~"5g-.*"}[1m]))

# HTTP request logs
{job=~"5g-.*"} | json | msg=~"HTTP request.*"

# Authentication logs
{job=~"5g-.*"} |= "authentication" or "Authentication"
```

### Registration Flow Logs

```logql
# Complete UE registration flow
{nf_type=~"AMF|AUSF|UDM|UDR"} |= "imsi-001010000000001"

# NRF registration activity
{nf_type="NRF"} |= "NF registered"

# Authentication flow
{nf_type=~"AUSF|UDM"} |= "authentication"
```

### Session Management Logs

```logql
# PDU session logs
{nf_type=~"SMF|UPF"} |= "PDU" or "session"

# PFCP messages
{nf_type=~"SMF|UPF"} |= "PFCP"

# GTP-U packets
{nf_type="UPF"} |= "GTP-U"
```

## 📈 Metrics from Logs

Loki can generate metrics from logs:

```logql
# Rate of logs per NF
rate({job=~"5g-.*"}[5m])

# Count of errors over time
count_over_time({job=~"5g-.*"} | json | level="error" [5m])

# HTTP request rate
rate({job=~"5g-.*"} |= "HTTP request" [1m])

# Registration attempts
count_over_time({nf_type="AMF"} |= "Registration" [5m])
```

## 🔍 Troubleshooting

### Check if Logs are Being Collected

```bash
# Check Promtail is running
docker ps | grep promtail

# Check Promtail logs
docker logs promtail --tail 50

# Query Loki directly
curl 'http://localhost:3100/loki/api/v1/labels' | jq

# Check what NF types have logs
curl 'http://localhost:3100/loki/api/v1/label/nf_type/values' | jq

# Query recent logs
curl 'http://localhost:3100/loki/api/v1/query_range?query={job="5g-nrf"}&limit=10' | jq
```

### Check Log Files

```bash
# List all NF log files
ls -lh /tmp/*.log

# Tail NRF logs
tail -f /tmp/nrf.log

# Check log format (should be JSON)
head -5 /tmp/nrf.log
```

### Restart Promtail

```bash
cd /home/silby/5G/observability
docker compose restart promtail
```

## 📝 Log File Locations

All NF logs are written to `/tmp/` and collected by Promtail:

- `/tmp/nrf.log` → `{nf_type="NRF"}`
- `/tmp/udr.log` → `{nf_type="UDR"}`
- `/tmp/udm.log` → `{nf_type="UDM"}`
- `/tmp/ausf.log` → `{nf_type="AUSF"}`
- `/tmp/amf.log` → `{nf_type="AMF"}`
- `/tmp/smf.log` → `{nf_type="SMF"}`
- `/tmp/upf.log` → `{nf_type="UPF"}`

## 🎯 Common Use Cases

### Debug UE Registration Issues

```logql
# View entire registration flow for a specific IMSI
{job=~"5g-.*"} |= "imsi-001010000000001" | json

# Check for errors in registration
{nf_type=~"AMF|AUSF|UDM|UDR"} | json | level="error"
```

### Monitor NRF Activity

```logql
# NRF registrations
{nf_type="NRF"} |= "NF registered"

# NRF heartbeats
{nf_type="NRF"} |= "heartbeat"

# NRF discovery requests
{nf_type="NRF"} |= "discovery"
```

### Track PDU Sessions

```logql
# SMF session creation
{nf_type="SMF"} |= "PDU session created"

# UPF session establishment
{nf_type="UPF"} |= "PFCP session"

# Session errors
{nf_type=~"SMF|UPF"} | json | level="error"
```

### Performance Analysis

```logql
# HTTP request durations
{job=~"5g-.*"} | json | msg=~"HTTP request.*duration.*"

# Slow requests (>100ms)
{job=~"5g-.*"} | json | duration > 0.1

# Database query times (UDR)
{nf_type="UDR"} |= "duration"
```

## 🎊 Your Loki Setup is Complete!

- ✅ All NF logs are being collected
- ✅ Structured JSON parsing configured
- ✅ Labels for easy filtering (nf_type, level, job)
- ✅ Grafana dashboard created
- ✅ Timestamps preserved from log entries

Access your logs now: **http://localhost:3001** → Dashboards → "5G Network Logs"

