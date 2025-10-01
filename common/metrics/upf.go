package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// UPF-specific metrics
var (
	// GTP-U metrics
	GTPUPackets = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "upf_gtpu_packets_total",
			Help: "Total number of GTP-U packets",
		},
		[]string{"direction"}, // uplink, downlink
	)

	GTPUBytes = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "upf_gtpu_bytes_total",
			Help: "Total number of GTP-U bytes",
		},
		[]string{"direction"},
	)

	GTPUPacketsDropped = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "upf_gtpu_packets_dropped_total",
			Help: "Total number of dropped GTP-U packets",
		},
		[]string{"reason"},
	)

	// Session metrics
	UPFActiveSessions = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "upf_active_sessions",
			Help: "Number of active UPF sessions",
		},
	)

	// PFCP metrics (UPF side - server)
	UPFPFCPSessionEstablishments = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "upf_pfcp_session_establishments_total",
			Help: "Total number of PFCP session establishments",
		},
		[]string{"result"},
	)

	UPFPFCPMessages = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "upf_pfcp_messages_total",
			Help: "Total number of PFCP messages",
		},
		[]string{"type"},
	)

	// QoS metrics
	QoSViolations = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "upf_qos_violations_total",
			Help: "Total number of QoS violations",
		},
		[]string{"qfi"},
	)

	// Throughput
	UplinkThroughput = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "upf_uplink_throughput_bps",
			Help: "Current uplink throughput in bits per second",
		},
	)

	DownlinkThroughput = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "upf_downlink_throughput_bps",
			Help: "Current downlink throughput in bits per second",
		},
	)
)

// RecordGTPUPacket records a GTP-U packet
func RecordGTPUPacket(direction string, bytes int) {
	GTPUPackets.WithLabelValues(direction).Inc()
	GTPUBytes.WithLabelValues(direction).Add(float64(bytes))
}

// RecordGTPUPacketDropped records a dropped GTP-U packet
func RecordGTPUPacketDropped(reason string) {
	GTPUPacketsDropped.WithLabelValues(reason).Inc()
}

// SetUPFActiveSessions sets the number of active sessions
func SetUPFActiveSessions(count int) {
	UPFActiveSessions.Set(float64(count))
}

// RecordUPFPFCPSessionEstablishment records a PFCP session establishment
func RecordUPFPFCPSessionEstablishment(result string) {
	UPFPFCPSessionEstablishments.WithLabelValues(result).Inc()
}

// RecordUPFPFCPMessage records a PFCP message
func RecordUPFPFCPMessage(msgType string) {
	UPFPFCPMessages.WithLabelValues(msgType).Inc()
}

// RecordQoSViolation records a QoS violation
func RecordQoSViolation(qfi string) {
	QoSViolations.WithLabelValues(qfi).Inc()
}

// SetUplinkThroughput sets the uplink throughput
func SetUplinkThroughput(bps float64) {
	UplinkThroughput.Set(bps)
}

// SetDownlinkThroughput sets the downlink throughput
func SetDownlinkThroughput(bps float64) {
	DownlinkThroughput.Set(bps)
}
