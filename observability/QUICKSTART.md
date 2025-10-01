# 5G Network Observability - Quick Start Guide

## 🚀 Getting Started in 3 Steps

### Step 1: Start the Observability Stack

```bash
cd observability
./start-observability.sh
```

Or manually:
```bash
docker compose up -d
```

### Step 2: Rebuild NFs with Metrics

```bash
cd ..
make build-nrf
# Rebuild other NFs as needed
```

### Step 3: Restart NFs and Verify

```bash
# Kill existing NFs
pkill nrf

# Start with metrics
./bin/nrf --config nf/nrf/config/nrf.yaml

# Verify metrics endpoint
curl http://localhost:9090/metrics
```

## 📊 Access Dashboards

- **Grafana**: http://localhost:3001 (admin/admin)
- **VictoriaMetrics**: http://localhost:8428/vmui
- **Loki**: http://localhost:3100/ready

## 📈 View Metrics

### Network Overview Dashboard
http://localhost:3001/d/5g-overview

### Explore Metrics (PromQL)
http://localhost:3001/explore

Example queries:
```promql
# Total HTTP requests across all NFs
sum(rate(http_requests_total[5m])) by (job)

# Active PDU sessions
smf_active_pdu_sessions

# UPF throughput
sum(rate(upf_gtpu_bytes_total[5m])) by (direction) * 8
```

### Explore Logs (LogQL)
Switch datasource to "Loki", then:

```logql
# All errors
{nf_type=~".+"} |= "error"

# AMF registration events
{nf_type="AMF"} |= "registration"

# High-level errors only
{nf_type=~".+"} | json | level="error"
```

## 🔧 Metrics Endpoints

| NF   | Port | URL |
|------|------|-----|
| NRF  | 9090 | http://localhost:9090/metrics |
| UDR  | 9091 | http://localhost:9091/metrics |
| UDM  | 9092 | http://localhost:9092/metrics |
| AUSF | 9094 | http://localhost:9094/metrics |
| AMF  | 9094 | http://localhost:9094/metrics |
| SMF  | 9095 | http://localhost:9095/metrics |
| UPF  | 9096 | http://localhost:9096/metrics |

## 📝 Key Metrics by NF

### NRF (Service Discovery)
- `nrf_registered_nfs_total` - Total registered NFs by type
- `nrf_nf_registrations_total` - Registration attempts
- `nrf_discovery_requests_total` - Discovery requests
- `nrf_heartbeats_received_total` - Heartbeats received

### AMF (Access & Mobility)
- `amf_registered_ues_total` - Active UE registrations
- `amf_registration_attempts_total` - Registration attempts
- `amf_authentication_requests_total` - Auth requests
- `amf_active_connections` - Active connections

### SMF (Session Management)
- `smf_active_pdu_sessions` - Active PDU sessions
- `smf_pdu_session_establishments_total` - Session establishments
- `smf_pfcp_sessions_active` - Active PFCP sessions
- `smf_active_qos_flows` - Active QoS flows

### UPF (User Plane)
- `upf_gtpu_packets_total` - GTP-U packets (uplink/downlink)
- `upf_gtpu_bytes_total` - GTP-U bytes (uplink/downlink)
- `upf_uplink_throughput_bps` - Uplink throughput
- `upf_downlink_throughput_bps` - Downlink throughput
- `upf_active_sessions` - Active sessions

### AUSF (Authentication)
- `ausf_authentication_attempts_total` - Auth attempts
- `ausf_aka_vector_generations_total` - AKA vector generations
- `ausf_active_auth_contexts` - Active contexts

### UDM (User Data Management)
- `udm_vector_generations_total` - Vector generations
- `udm_sdm_requests_total` - SDM requests
- `udm_active_ue_contexts` - Active UE contexts

### UDR (Data Repository)
- `udr_subscriber_queries_total` - Subscriber queries
- `udr_database_query_duration_seconds` - DB query duration
- `udr_auth_subscription_queries_total` - Auth queries

## 🎯 Common Tasks

### Check if VictoriaMetrics is scraping
```bash
curl http://localhost:8428/api/v1/targets | jq
```

### Check Loki status
```bash
curl http://localhost:3100/ready
```

### View container logs
```bash
docker compose logs -f grafana
docker compose logs -f victoriametrics
docker compose logs -f loki
```

### Stop observability stack
```bash
./stop-observability.sh
```

## 🐛 Troubleshooting

### Metrics not showing in Grafana
1. Check if NFs are exposing metrics: `curl http://localhost:9090/metrics`
2. Check VictoriaMetrics targets: `curl http://localhost:8428/api/v1/targets`
3. Verify vmagent logs: `docker compose logs vmagent`

### Logs not showing in Grafana
1. Check Promtail logs: `docker compose logs promtail`
2. Verify log files exist: `ls -la /tmp/*.log`
3. Check Loki status: `curl http://localhost:3100/ready`

### Container not starting
1. Check logs: `docker compose logs <service-name>`
2. Verify ports are free: `netstat -tuln | grep <port>`
3. Restart: `docker compose restart <service-name>`

## 📚 Next Steps

1. Create custom dashboards for each NF
2. Set up alerting rules
3. Configure external alert receivers (Slack, email, etc.)
4. Add distributed tracing with Jaeger
5. Create SLA monitoring dashboards
6. Set up long-term metrics retention

## 🔗 Resources

- [PromQL Cheat Sheet](https://promlabs.com/promql-cheat-sheet/)
- [LogQL Guide](https://grafana.com/docs/loki/latest/logql/)
- [VictoriaMetrics Docs](https://docs.victoriametrics.com/)
- [Grafana Dashboard Best Practices](https://grafana.com/docs/grafana/latest/dashboards/)

