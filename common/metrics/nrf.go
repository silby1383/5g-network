package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// NRF-specific metrics
var (
	// NF Registration metrics
	RegisteredNFsTotal = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "nrf_registered_nfs_total",
			Help: "Total number of registered NFs by type",
		},
		[]string{"nf_type"},
	)

	NFRegistrations = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "nrf_nf_registrations_total",
			Help: "Total number of NF registrations",
		},
		[]string{"nf_type", "status"},
	)

	NFDeregistrations = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "nrf_nf_deregistrations_total",
			Help: "Total number of NF deregistrations",
		},
		[]string{"nf_type"},
	)

	// Discovery metrics
	DiscoveryRequests = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "nrf_discovery_requests_total",
			Help: "Total number of NF discovery requests",
		},
		[]string{"target_nf_type", "status"},
	)

	// Subscription metrics
	ActiveSubscriptions = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "nrf_active_subscriptions",
			Help: "Number of active subscriptions",
		},
	)

	// Heartbeat metrics
	HeartbeatsReceived = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "nrf_heartbeats_received_total",
			Help: "Total number of heartbeats received",
		},
		[]string{"nf_type"},
	)
)

// SetRegisteredNFs sets the count of registered NFs by type
func SetRegisteredNFs(nfType string, count int) {
	RegisteredNFsTotal.WithLabelValues(nfType).Set(float64(count))
}

// RecordNFRegistration records an NF registration attempt
func RecordNFRegistration(nfType, status string) {
	NFRegistrations.WithLabelValues(nfType, status).Inc()
}

// RecordNFDeregistration records an NF deregistration
func RecordNFDeregistration(nfType string) {
	NFDeregistrations.WithLabelValues(nfType).Inc()
}

// RecordDiscoveryRequest records a discovery request
func RecordDiscoveryRequest(targetNFType, status string) {
	DiscoveryRequests.WithLabelValues(targetNFType, status).Inc()
}

// SetActiveSubscriptions sets the number of active subscriptions
func SetActiveSubscriptions(count int) {
	ActiveSubscriptions.Set(float64(count))
}

// RecordHeartbeat records a heartbeat reception
func RecordHeartbeat(nfType string) {
	HeartbeatsReceived.WithLabelValues(nfType).Inc()
}
