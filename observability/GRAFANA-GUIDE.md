# Grafana Metrics Guide

## ✅ Metrics Are Now Working!

All Network Functions are now instrumented and metrics are being collected by VictoriaMetrics.

## 🚀 Quick Start

1. **Open Grafana**: http://localhost:3001
2. **Login**: admin / admin
3. **Go to Explore**: Click compass icon (🧭) on left sidebar
4. **Select Datasource**: VictoriaMetrics
5. **Set Time Range**: "Last 15 minutes" (top-right corner)

## 📊 Working Metrics by NF

### NRF (Network Repository Function)
```promql
# Registered NFs
nrf_registered_nfs_total

# Heartbeats received
nrf_heartbeats_received_total

# Heartbeat rate (per second)
rate(nrf_heartbeats_received_total[1m])

# Registrations
nrf_registrations_total

# Discoveries
nrf_discovery_requests_total
```

### AUSF (Authentication Server Function)
```promql
# Authentication attempts
ausf_authentication_attempts_total

# Success rate
rate(ausf_authentication_attempts_total{result="success"}[5m])

# AKA vector generations
ausf_aka_vector_generations_total

# Authentication latency (p99)
histogram_quantile(0.99, ausf_authentication_duration_seconds_bucket)
```

### AMF (Access and Mobility Management Function)
```promql
# Registration attempts
amf_registration_attempts_total

# Registered UEs
amf_registered_ues_total

# Registration rate
rate(amf_registration_attempts_total[5m])

# Active connections
amf_active_connections

# Authentication requests
amf_authentication_requests_total
```

### UDM (Unified Data Management)
```promql
# Vector generations
udm_vector_generations_total

# Vector generation latency
udm_vector_generation_duration_seconds

# SDM requests
udm_sdm_requests_total

# Active UE contexts
udm_active_ue_contexts
```

### UDR (Unified Data Repository)
```promql
# Subscriber queries
udr_subscriber_queries_total

# Auth subscription queries
udr_auth_subscription_queries_total

# Database query latency
udr_database_query_duration_seconds

# Active SDM subscriptions
udr_active_sdm_subscriptions
```

### SMF (Session Management Function)
```promql
# PDU session establishments
smf_pdu_session_establishments_total

# Active PDU sessions
smf_active_pdu_sessions

# Session releases
smf_pdu_session_releases_total

# PFCP sessions
smf_pfcp_sessions_active

# Active QoS flows
smf_active_qos_flows

# PFCP messages
smf_pfcp_messages_total
```

### UPF (User Plane Function)
```promql
# GTP-U packets
upf_gtpu_packets_total

# GTP-U bytes
upf_gtpu_bytes_total

# Throughput (Mbps)
rate(upf_gtpu_bytes_total[5m]) * 8 / 1000000

# Active sessions
upf_active_sessions

# PFCP sessions
upf_pfcp_session_establishments_total

# QoS violations
upf_qos_violations_total
```

## 📈 Useful Dashboard Queries

### Network Overview
```promql
# All services health (should show 1 for each)
service_up

# Total HTTP request rate
sum(rate(http_requests_total[5m])) by (job)

# Error rate %
sum(rate(http_requests_total{code=~"5.."}[5m])) / 
sum(rate(http_requests_total[5m])) * 100
```

### Authentication Flow
```promql
# End-to-end authentication rate
rate(amf_authentication_requests_total{result="success"}[5m])

# Authentication success %
sum(rate(ausf_authentication_attempts_total{result="success"}[5m])) /
sum(rate(ausf_authentication_attempts_total[5m])) * 100

# Average auth latency
rate(ausf_authentication_duration_seconds_sum[5m]) /
rate(ausf_authentication_duration_seconds_count[5m])
```

### Session Management
```promql
# Active PDU sessions
smf_active_pdu_sessions

# Session establishment rate
rate(smf_pdu_session_establishments_total{result="initial"}[5m])

# Session success rate %
sum(rate(smf_pdu_session_establishments_total{result="initial"}[5m])) /
sum(rate(smf_pdu_session_establishments_total[5m])) * 100
```

### UPF Data Plane
```promql
# Uplink throughput (Mbps)
rate(upf_gtpu_bytes_total{direction="uplink"}[5m]) * 8 / 1000000

# Downlink throughput (Mbps)
rate(upf_gtpu_bytes_total{direction="downlink"}[5m]) * 8 / 1000000

# Total throughput
sum(rate(upf_gtpu_bytes_total[5m])) * 8 / 1000000

# Packet rate
sum(rate(upf_gtpu_packets_total[5m])) by (direction)
```

## 🔍 Troubleshooting

### Metrics Show Zero
1. **Check Time Range**: Set to "Last 15 minutes" (not "Last 6 hours")
2. **Run Test Scripts**: Generate traffic to increment counters
   ```bash
   cd /home/silby/5G
   bash scripts/test-ue-registration.sh
   bash scripts/test-pdu-session.sh
   ```
3. **Check NFs Are Running**: `ps aux | grep "bin/"`
4. **Verify Endpoints**: `curl http://localhost:9090/metrics`

### Metrics Not Updating
1. **Wait 30-60 seconds**: VictoriaMetrics scrapes every 15 seconds
2. **Refresh Grafana**: Click refresh button or reload page
3. **Check VictoriaMetrics**: 
   ```bash
   curl 'http://localhost:8428/api/v1/query?query=service_up'
   ```

### No Data in Grafana
1. **Check Datasource**: Must be "VictoriaMetrics" (not Prometheus)
2. **Check URL**: http://localhost:8428 in datasource settings
3. **Test Query**: Try simple query like `service_up` first

## 🎯 Expected Values After Tests

After running `test-ue-registration.sh`:
- `amf_registration_attempts_total{result="success"}` > 0
- `ausf_authentication_attempts_total{result="success"}` > 0
- `udm_vector_generations_total{result="success"}` > 0
- `udr_auth_subscription_queries_total{result="success"}` > 0

After running `test-pdu-session.sh`:
- `smf_pdu_session_establishments_total{result="initial"}` > 0
- `smf_pdu_session_releases_total` > 0

Always visible:
- `service_up` = 1 for all NFs
- `nrf_heartbeats_received_total` - continuously increasing
- `nrf_registered_nfs_total` = 6 or 7

## 📚 Documentation

- **Full Metrics List**: `observability/METRICS-SUMMARY.md`
- **Troubleshooting**: `observability/TROUBLESHOOTING-METRICS.md`
- **Setup Guide**: `observability/README.md`
- **Quick Reference**: `observability/QUICKSTART.md`

Your 5G network now has complete metrics visibility! 🎊

