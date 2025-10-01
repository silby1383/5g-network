package metrics

import (
	"fmt"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

// Common metrics that all NFs should expose
var (
	// HTTP Request metrics
	HTTPRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	HTTPRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request latency in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)

	// Service health
	ServiceUp = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "service_up",
			Help: "Whether the service is up (1 = up, 0 = down)",
		},
	)

	// NRF Registration
	NRFRegistered = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "nrf_registered",
			Help: "Whether registered with NRF (1 = registered, 0 = not registered)",
		},
	)

	NRFHeartbeatFailures = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "nrf_heartbeat_failures_total",
			Help: "Total number of NRF heartbeat failures",
		},
	)
)

// MetricsServer represents a Prometheus metrics HTTP server
type MetricsServer struct {
	port   int
	server *http.Server
	logger *zap.Logger
}

// NewMetricsServer creates a new metrics server
func NewMetricsServer(port int, logger *zap.Logger) *MetricsServer {
	return &MetricsServer{
		port:   port,
		logger: logger,
	}
}

// Start starts the metrics HTTP server
func (m *MetricsServer) Start() error {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())

	// Health endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	m.server = &http.Server{
		Addr:         fmt.Sprintf(":%d", m.port),
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	m.logger.Info("Starting metrics server", zap.Int("port", m.port))
	return m.server.ListenAndServe()
}

// Stop gracefully stops the metrics server
func (m *MetricsServer) Stop() error {
	if m.server != nil {
		return m.server.Close()
	}
	return nil
}

// RecordHTTPRequest records an HTTP request
func RecordHTTPRequest(method, path, status string, duration float64) {
	HTTPRequestsTotal.WithLabelValues(method, path, status).Inc()
	HTTPRequestDuration.WithLabelValues(method, path).Observe(duration)
}

// SetServiceUp sets the service health status
func SetServiceUp(up bool) {
	if up {
		ServiceUp.Set(1)
	} else {
		ServiceUp.Set(0)
	}
}

// SetNRFRegistered sets NRF registration status
func SetNRFRegistered(registered bool) {
	if registered {
		NRFRegistered.Set(1)
	} else {
		NRFRegistered.Set(0)
	}
}

// RecordNRFHeartbeatFailure records an NRF heartbeat failure
func RecordNRFHeartbeatFailure() {
	NRFHeartbeatFailures.Inc()
}
