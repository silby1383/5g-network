#!/bin/bash

# Start 5G Network Observability Stack
# Uses VictoriaMetrics, Loki, and Grafana

set -e

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  🚀 Starting 5G Network Observability Stack"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

cd "$(dirname "$0")"

# Check if docker is available
if ! command -v docker &> /dev/null; then
    echo "❌ Docker is not installed"
    exit 1
fi

# Start containers
echo "📦 Starting containers..."
if command -v docker-compose &> /dev/null; then
    docker-compose up -d
elif docker compose version &> /dev/null; then
    docker compose up -d
else
    echo "❌ Docker Compose not found"
    exit 1
fi

echo ""
echo "⏳ Waiting for services to be ready..."
sleep 10

# Check services
echo ""
echo "📊 Service Status:"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
docker ps --filter "name=5g-" --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  ✅ Observability Stack Started!"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "🌐 Access URLs:"
echo "  • Grafana:         http://localhost:3001 (admin/admin)"
echo "  • VictoriaMetrics: http://localhost:8428/vmui"
echo "  • Loki:            http://localhost:3100/ready"
echo "  • Alertmanager:    http://localhost:9093"
echo ""
echo "📊 Quick Links:"
echo "  • Network Overview: http://localhost:3001/d/5g-overview"
echo "  • Explore Metrics:  http://localhost:3001/explore"
echo "  • Explore Logs:     http://localhost:3001/explore?orgId=1&left=%5B%22now-1h%22,%22now%22,%22Loki%22%5D"
echo ""
echo "🛑 To stop: ./stop-observability.sh or 'docker compose down'"
echo ""

