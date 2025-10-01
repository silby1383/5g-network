# 5G Network Observability Stack

Professional observability solution for the 5G Core Network using VictoriaMetrics, Loki, and Grafana.

## 🏗️ Architecture

```
┌─────────────┐
│  5G NFs     │
│ (7 services)│
└──────┬──────┘
       │
       ├──► VictoriaMetrics (Metrics)
       │    Port: 8428
       │
       ├──► Loki (Logs)
       │    Port: 3100
       │
       └──► Prometheus Exporter
            Ports: 9090-9096

         ┌──────────────┐
         │   vmagent    │ ◄─── Scrapes metrics
         │  (scraper)   │
         └──────┬───────┘
                │
         ┌──────▼───────┐
         │ VictoriaMetrics│
         │  (storage)   │
         └──────────────┘

         ┌──────────────┐
         │   Promtail   │ ◄─── Collects logs
         │ (log shipper)│
         └──────┬───────┘
                │
         ┌──────▼───────┐
         │     Loki     │
         │  (log aggr.) │
         └──────────────┘

         ┌──────────────┐
         │   Grafana    │ ◄─── Visualizes
         │  (dashboard) │
         └──────────────┘
```

## 🚀 Quick Start

### Start Observability Stack

```bash
cd observability
docker-compose up -d
```

### Access Dashboards

- **Grafana**: http://localhost:3001
  - Username: `admin`
  - Password: `admin`
- **VictoriaMetrics UI**: http://localhost:8428/vmui
- **Loki**: http://localhost:3100/ready
- **Alertmanager**: http://localhost:9093

### Stop Stack

```bash
cd observability
docker-compose down
```

## 📊 Dashboards

### Network Overview Dashboard
- **URL**: http://localhost:3001/d/5g-overview
- **Features**:
  - NF status visualization
  - Request rates across all NFs
  - Real-time log streaming
  - Network health metrics

### Per-NF Dashboards
Each Network Function has a dedicated dashboard:

1. **NRF Dashboard** - Service discovery metrics
2. **UDR Dashboard** - Database performance
3. **UDM Dashboard** - Subscriber operations
4. **AUSF Dashboard** - Authentication metrics
5. **AMF Dashboard** - Registration & mobility
6. **SMF Dashboard** - Session management
7. **UPF Dashboard** - User plane throughput

### Call Flow Dashboard
- End-to-end flow visualization
- Request tracing across NFs
- Latency analysis
- Error tracking

## 📈 Metrics Available

### System Metrics
- CPU usage per NF
- Memory consumption
- HTTP request rates
- Response times

### 5G-Specific Metrics
- NRF registrations
- UE registrations (AMF)
- PDU sessions (SMF)
- GTP-U tunnels (UPF)
- Authentication attempts (AUSF)
- Database queries (UDR)

### Example Queries

```promql
# Total HTTP requests across all NFs
sum(rate(http_requests_total[5m])) by (nf_type)

# Active PDU sessions
sum(pdu_sessions_active) by (smf_instance)

# Authentication success rate
rate(auth_attempts_total{result="success"}[5m]) / 
rate(auth_attempts_total[5m])

# UPF throughput
sum(rate(gtpu_bytes_total[5m])) by (direction)
```

## 📝 Log Queries

### LogQL Examples

```logql
# All errors across network
{nf_type=~".+"} |= "error" | json

# AMF registration events
{nf_type="AMF"} |= "registration" | json

# SMF session establishment
{nf_type="SMF"} |= "session" |= "established"

# High-level errors only
{nf_type=~".+"} | json | level="error"
```

## 🔔 Alerting

### Alert Rules
Configured alerts include:

- **NF Down**: Network function unavailable
- **High Error Rate**: >5% error rate
- **Memory Pressure**: >80% memory usage
- **Slow Responses**: P95 latency >1s
- **Session Failures**: PDU session failures
- **Authentication Failures**: Auth failure spike

### Alert Destinations
Configure in `alertmanager/config.yml`:
- Webhook endpoints
- Email notifications
- Slack integration
- PagerDuty

## 🔧 Configuration

### Adding New Metrics

1. Add scrape target in `victoriametrics/prometheus.yml`
2. Implement Prometheus exporter in NF
3. Expose metrics endpoint

### Custom Dashboards

1. Create JSON in `grafana/dashboards/`
2. Reload Grafana or restart container
3. Access via Grafana UI

### Log Retention

Configure in `loki/loki-config.yml`:
```yaml
limits_config:
  retention_period: 168h  # 7 days
```

### Metrics Retention

Configure in `docker-compose.yml`:
```yaml
command:
  - "--retentionPeriod=12"  # 12 months
```

## 📦 Data Persistence

Volumes for data persistence:
- `vm-data`: VictoriaMetrics time-series data
- `loki-data`: Loki log data
- `grafana-data`: Grafana dashboards and settings

## 🐛 Troubleshooting

### Check Container Logs
```bash
docker-compose logs -f victoriametrics
docker-compose logs -f loki
docker-compose logs -f grafana
```

### Verify Metrics Collection
```bash
curl http://localhost:8428/api/v1/targets
```

### Verify Log Collection
```bash
curl http://localhost:3100/ready
```

### Grafana Not Loading
1. Check Grafana logs
2. Verify datasource connectivity
3. Check dashboard provisioning

## 📚 Resources

- [VictoriaMetrics Docs](https://docs.victoriametrics.com/)
- [Loki Docs](https://grafana.com/docs/loki/latest/)
- [Grafana Docs](https://grafana.com/docs/grafana/latest/)
- [PromQL Guide](https://prometheus.io/docs/prometheus/latest/querying/basics/)
- [LogQL Guide](https://grafana.com/docs/loki/latest/logql/)

## 🎯 Next Steps

1. ✅ Configure metrics exporters in NFs
2. ✅ Create custom dashboards
3. ✅ Set up alerting rules
4. ✅ Integrate with external monitoring
5. ✅ Add distributed tracing

