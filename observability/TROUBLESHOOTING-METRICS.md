# Metrics Troubleshooting Guide

## Issue: Metrics showing 0 counts in Grafana

### Root Cause

Docker container isolation prevents vmagent (running in Docker) from reaching NF metrics endpoints (running on host):

```
NFs (host)           → Metrics on localhost:9090-9098 ✅
vmagent (Docker)     → Cannot reach host's localhost ❌
```

### Why This Happens

1. NFs run directly on the host, exposing metrics on `0.0.0.0:9090-9098`
2. vmagent runs in a Docker container
3. Docker networking isolates containers from host's localhost
4. Even with `host.docker.internal` or host IP, firewall/routing blocks access

### Solutions

#### Option 1: Run VictoriaMetrics Stack Natively (Recommended)

Install VictoriaMetrics and vmagent as native binaries on the host:

```bash
# Download VictoriaMetrics
cd /tmp
wget https://github.com/VictoriaMetrics/VictoriaMetrics/releases/download/v1.96.0/victoria-metrics-linux-amd64-v1.96.0.tar.gz
tar xzf victoria-metrics-linux-amd64-v1.96.0.tar.gz
sudo mv victoria-metrics-prod /usr/local/bin/victoriametrics

# Download vmagent
wget https://github.com/VictoriaMetrics/VictoriaMetrics/releases/download/v1.96.0/vmutils-linux-amd64-v1.96.0.tar.gz
tar xzf vmutils-linux-amd64-v1.96.0.tar.gz
sudo mv vmagent-prod /usr/local/bin/vmagent

# Start VictoriaMetrics
victoriametrics -storageDataPath=/var/lib/victoriametrics -retentionPeriod=12 &

# Start vmagent
vmagent -promscrape.config=/home/silby/5G/observability/victoriametrics/prometheus.yml \
        -remoteWrite.url=http://localhost:8428/api/v1/write &
```

#### Option 2: Use Docker Host Network Mode

**Note**: This approach has known issues on some systems.

```yaml
# In docker-compose.yml
vmagent:
  image: victoriametrics/vmagent:latest
  network_mode: "host"  # Share host's network namespace
  volumes:
    - ./victoriametrics/prometheus.yml:/etc/prometheus/prometheus.yml
  command:
    - "--promscrape.config=/etc/prometheus/prometheus.yml"
    - "--remoteWrite.url=http://localhost:8428/api/v1/write"
```

Update prometheus.yml targets to use `localhost`:
```yaml
- targets: ['localhost:9090']  # Instead of host.docker.internal
```

#### Option 3: Reverse Proxy with Port Publishing

Expose metrics ports from host to Docker network:

```yaml
# In docker-compose.yml, add to each service:
extra_hosts:
  - "metrics-host:192.168.1.15"  # Your host IP
```

Or use socat to forward ports into Docker network.

#### Option 4: Run NFs in Docker Too

The most consistent approach - containerize everything:

```yaml
services:
  nrf:
    build: ./nf/nrf
    ports:
      - "8080:8080"
      - "9090:9090"
    networks:
      - 5g-network
```

This ensures everything is in the same network namespace.

### Current Workaround

For immediate results, let's run vmagent natively:

```bash
# Stop Docker vmagent
cd /home/silby/5G/observability
docker compose stop vmagent

# Update prometheus.yml targets to localhost
sed -i 's/host\.docker\.internal/localhost/g' victoriametrics/prometheus.yml
sed -i 's/192\.168\.1\.15/localhost/g' victoriametrics/prometheus.yml

# Start vmagent natively
vmagent -promscrape.config=$PWD/victoriametrics/prometheus.yml \
        -remoteWrite.url=http://localhost:8428/api/v1/write \
        -httpListenAddr=:8429 &

# Wait 30 seconds and check metrics
sleep 30
curl 'http://localhost:8428/api/v1/query?query=service_up'
```

### Verification

After applying the solution:

```bash
# Check vmagent can reach metrics
curl http://localhost:9090/metrics | head

# Check VictoriaMetrics has data
curl 'http://localhost:8428/api/v1/query?query=service_up' | jq

# View in Grafana
# http://localhost:3001 → Explore → service_up
```

### Long-term Solution

Containerize all NFs for consistent networking:
- All services in same Docker Compose stack
- Shared network namespace
- Service discovery via Docker DNS
- No host/container networking confusion

This is the recommended production approach.

