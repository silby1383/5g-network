package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// AMF-specific metrics
var (
	// UE Registration metrics
	RegisteredUEs = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "amf_registered_ues_total",
			Help: "Total number of registered UEs",
		},
	)

	RegistrationAttempts = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "amf_registration_attempts_total",
			Help: "Total number of UE registration attempts",
		},
		[]string{"result"},
	)

	// Authentication metrics
	AuthenticationRequests = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "amf_authentication_requests_total",
			Help: "Total number of authentication requests",
		},
		[]string{"result"},
	)

	// Mobility metrics
	HandoverAttempts = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "amf_handover_attempts_total",
			Help: "Total number of handover attempts",
		},
		[]string{"result"},
	)

	// Connection metrics
	ActiveConnections = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "amf_active_connections",
			Help: "Number of active UE connections",
		},
	)
)

// SetRegisteredUEs sets the count of registered UEs
func SetRegisteredUEs(count int) {
	RegisteredUEs.Set(float64(count))
}

// RecordRegistrationAttempt records a registration attempt
func RecordRegistrationAttempt(result string) {
	RegistrationAttempts.WithLabelValues(result).Inc()
}

// RecordAuthenticationRequest records an authentication request
func RecordAuthenticationRequest(result string) {
	AuthenticationRequests.WithLabelValues(result).Inc()
}

// RecordHandoverAttempt records a handover attempt
func RecordHandoverAttempt(result string) {
	HandoverAttempts.WithLabelValues(result).Inc()
}

// SetActiveConnections sets the number of active connections
func SetActiveConnections(count int) {
	ActiveConnections.Set(float64(count))
}
