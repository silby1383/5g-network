package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// UDR-specific metrics
var (
	// Subscriber data metrics
	SubscriberQueries = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "udr_subscriber_queries_total",
			Help: "Total number of subscriber data queries",
		},
		[]string{"type", "result"},
	)

	SubscriberUpdates = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "udr_subscriber_updates_total",
			Help: "Total number of subscriber data updates",
		},
		[]string{"type", "result"},
	)

	// Database metrics
	DatabaseQueryDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "udr_database_query_duration_seconds",
			Help:    "Database query duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"operation"},
	)

	DatabaseErrors = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "udr_database_errors_total",
			Help: "Total number of database errors",
		},
		[]string{"operation"},
	)

	// Authentication data
	AuthSubscriptionQueries = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "udr_auth_subscription_queries_total",
			Help: "Total number of authentication subscription queries",
		},
		[]string{"result"},
	)

	// SDM subscriptions
	ActiveSDMSubscriptions = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "udr_active_sdm_subscriptions",
			Help: "Number of active SDM subscriptions",
		},
	)
)

// RecordSubscriberQuery records a subscriber query
func RecordSubscriberQuery(queryType, result string) {
	SubscriberQueries.WithLabelValues(queryType, result).Inc()
}

// RecordSubscriberUpdate records a subscriber update
func RecordSubscriberUpdate(updateType, result string) {
	SubscriberUpdates.WithLabelValues(updateType, result).Inc()
}

// RecordDatabaseQuery records a database query
func RecordDatabaseQuery(operation string, duration float64) {
	DatabaseQueryDuration.WithLabelValues(operation).Observe(duration)
}

// RecordDatabaseError records a database error
func RecordDatabaseError(operation string) {
	DatabaseErrors.WithLabelValues(operation).Inc()
}

// RecordAuthSubscriptionQuery records an auth subscription query
func RecordAuthSubscriptionQuery(result string) {
	AuthSubscriptionQueries.WithLabelValues(result).Inc()
}

// SetActiveSDMSubscriptions sets the number of active SDM subscriptions
func SetActiveSDMSubscriptions(count int) {
	ActiveSDMSubscriptions.Set(float64(count))
}
