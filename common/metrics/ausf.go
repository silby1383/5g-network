package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// AUSF-specific metrics
var (
	// Authentication metrics
	AuthenticationAttempts = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ausf_authentication_attempts_total",
			Help: "Total number of authentication attempts",
		},
		[]string{"method", "result"},
	)

	AuthenticationDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "ausf_authentication_duration_seconds",
			Help:    "Authentication duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method"},
	)

	// 5G-AKA specific
	AKAVectorGenerations = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ausf_aka_vector_generations_total",
			Help: "Total number of AKA vector generations",
		},
		[]string{"result"},
	)

	// Active contexts
	ActiveAuthContexts = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "ausf_active_auth_contexts",
			Help: "Number of active authentication contexts",
		},
	)
)

// RecordAuthenticationAttempt records an authentication attempt
func RecordAuthenticationAttempt(method, result string) {
	AuthenticationAttempts.WithLabelValues(method, result).Inc()
}

// RecordAuthenticationDuration records authentication duration
func RecordAuthenticationDuration(method string, duration float64) {
	AuthenticationDuration.WithLabelValues(method).Observe(duration)
}

// RecordAKAVectorGeneration records an AKA vector generation
func RecordAKAVectorGeneration(result string) {
	AKAVectorGenerations.WithLabelValues(result).Inc()
}

// SetActiveAuthContexts sets the number of active auth contexts
func SetActiveAuthContexts(count int) {
	ActiveAuthContexts.Set(float64(count))
}
