package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// UDM-specific metrics
var (
	// Authentication vector generation
	VectorGenerations = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "udm_vector_generations_total",
			Help: "Total number of authentication vector generations",
		},
		[]string{"result"},
	)

	VectorGenerationDuration = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "udm_vector_generation_duration_seconds",
			Help:    "Vector generation duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
	)

	// Subscriber data management
	SDMRequests = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "udm_sdm_requests_total",
			Help: "Total number of SDM requests",
		},
		[]string{"type", "result"},
	)

	// UE context management
	UEContextRegistrations = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "udm_ue_context_registrations_total",
			Help: "Total number of UE context registrations",
		},
		[]string{"result"},
	)

	ActiveUEContexts = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "udm_active_ue_contexts",
			Help: "Number of active UE contexts",
		},
	)

	// SQN management
	SQNIncrements = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "udm_sqn_increments_total",
			Help: "Total number of SQN increments",
		},
	)
)

// RecordVectorGeneration records a vector generation
func RecordVectorGeneration(result string) {
	VectorGenerations.WithLabelValues(result).Inc()
}

// RecordVectorGenerationDuration records vector generation duration
func RecordVectorGenerationDuration(duration float64) {
	VectorGenerationDuration.Observe(duration)
}

// RecordSDMRequest records an SDM request
func RecordSDMRequest(requestType, result string) {
	SDMRequests.WithLabelValues(requestType, result).Inc()
}

// RecordUEContextRegistration records a UE context registration
func RecordUEContextRegistration(result string) {
	UEContextRegistrations.WithLabelValues(result).Inc()
}

// SetActiveUEContexts sets the number of active UE contexts
func SetActiveUEContexts(count int) {
	ActiveUEContexts.Set(float64(count))
}

// RecordSQNIncrement records an SQN increment
func RecordSQNIncrement() {
	SQNIncrements.Inc()
}
