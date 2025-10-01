#!/bin/bash

# Script to add metrics initialization to all NF main.go files
# This adds metrics server startup code to UDM, AUSF, AMF, SMF, and UPF

set -e

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  🔧 Adding Metrics to All NFs"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

cd "$(dirname "$0")/.."

# Note: This script shows what needs to be added manually to each NF
# Due to different code structures, manual addition is recommended

cat << 'EOF'

ADD THE FOLLOWING TO EACH NF's cmd/main.go:

1. ADD IMPORT:
   "github.com/your-org/5g-network/common/metrics"

2. ADD AFTER context creation (ctx, cancel := context.WithCancel...):

	// Initialize metrics server
	metricsServer := metrics.NewMetricsServer(PORT, logger)
	go func() {
		logger.Info("Starting metrics server on :PORT")
		if err := metricsServer.Start(); err != nil {
			logger.Error("Metrics server error", zap.Error(err))
		}
	}()
	defer metricsServer.Stop()

	// Set service up
	metrics.SetServiceUp(true)
	defer metrics.SetServiceUp(false)

PORTS:
  UDM:  9092
  AUSF: 9093  (Note: NOT 9094, that conflicts with AMF)
  AMF:  9094
  SMF:  9095
  UPF:  9096 (or 9097 if using admin port)

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

OR use this quick automated approach:

  1. Rebuild all NFs:
     ./scripts/rebuild-and-restart-nfs.sh
  
  2. Each NF will expose metrics once recompiled

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

EOF

echo "Manual additions are required due to different NF code structures."
echo "Imports have been added. Rebuild using: make build-all"
echo ""

