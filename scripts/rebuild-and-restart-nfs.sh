#!/bin/bash

# Rebuild and Restart All NFs with Metrics Support
# This script rebuilds all NFs and restarts them to enable metrics collection

set -e

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  🔄 REBUILDING AND RESTARTING ALL NFS WITH METRICS"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

cd "$(dirname "$0")/.."

# Kill all running NFs
echo "🛑 Stopping all running NFs..."
pkill -f "bin/nrf" || true
pkill -f "bin/udr" || true
pkill -f "bin/udm" || true
pkill -f "bin/ausf" || true
pkill -f "bin/amf" || true
pkill -f "bin/smf" || true
pkill -f "bin/upf" || true

sleep 2

# Rebuild all NFs
echo ""
echo "🔨 Rebuilding all NFs..."
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

echo "  Building NRF..."
make build-nrf

echo "  Building UDR..."
make build-udr

echo "  Building UDM..."
make build-udm

echo "  Building AUSF..."
make build-ausf

echo "  Building AMF..."
make build-amf

echo "  Building SMF..."
make build-smf

echo "  Building UPF..."
make build-upf

# Start all NFs with metrics
echo ""
echo "🚀 Starting all NFs..."
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

echo "  Starting NRF (metrics on :9090)..."
nohup ./bin/nrf --config nf/nrf/config/nrf.yaml > /tmp/nrf.log 2>&1 &
sleep 1

echo "  Starting UDR (metrics on :9091)..."
nohup ./bin/udr --config nf/udr/config/udr.yaml > /tmp/udr.log 2>&1 &
sleep 1

echo "  Starting UDM (metrics on :9092)..."
nohup ./bin/udm --config nf/udm/config/udm.yaml > /tmp/udm.log 2>&1 &
sleep 1

echo "  Starting AUSF (metrics on :9094)..."
nohup ./bin/ausf --config nf/ausf/config/ausf.yaml > /tmp/ausf.log 2>&1 &
sleep 1

echo "  Starting AMF (metrics on :9094)..."
nohup ./bin/amf --config nf/amf/config/amf.yaml > /tmp/amf.log 2>&1 &
sleep 1

echo "  Starting SMF (metrics on :9095)..."
nohup ./bin/smf --config nf/smf/config/smf.yaml > /tmp/smf.log 2>&1 &
sleep 1

echo "  Starting UPF (metrics on :9096)..."
nohup ./bin/upf --config nf/upf/config/upf.yaml > /tmp/upf.log 2>&1 &
sleep 2

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  ✅ ALL NFS RESTARTED WITH METRICS SUPPORT"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

# Verify NF processes
echo "📊 Running processes:"
ps aux | grep -E "bin/(nrf|udr|udm|ausf|amf|smf|upf)" | grep -v grep

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  🔍 VERIFY METRICS ENDPOINTS"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "Test metrics endpoints:"
echo "  curl http://localhost:9090/metrics | head  # NRF"
echo "  curl http://localhost:9091/metrics | head  # UDR"
echo "  curl http://localhost:9092/metrics | head  # UDM"
echo "  curl http://localhost:9094/metrics | head  # AUSF (or AMF)"
echo "  curl http://localhost:9095/metrics | head  # SMF"
echo "  curl http://localhost:9096/metrics | head  # UPF"
echo ""
echo "View in Grafana:"
echo "  http://localhost:3001/explore"
echo ""
echo "Check VictoriaMetrics targets:"
echo "  curl http://localhost:8428/api/v1/targets | jq"
echo ""

