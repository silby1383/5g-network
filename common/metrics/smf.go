package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// SMF-specific metrics
var (
	// PDU Session metrics
	ActivePDUSessions = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "smf_active_pdu_sessions",
			Help: "Number of active PDU sessions",
		},
	)

	PDUSessionEstablishments = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "smf_pdu_session_establishments_total",
			Help: "Total number of PDU session establishment attempts",
		},
		[]string{"result", "dnn"},
	)

	PDUSessionModifications = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "smf_pdu_session_modifications_total",
			Help: "Total number of PDU session modifications",
		},
		[]string{"result"},
	)

	PDUSessionReleases = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "smf_pdu_session_releases_total",
			Help: "Total number of PDU session releases",
		},
		[]string{"reason"},
	)

	// PFCP metrics (SMF side - client)
	SMFPFCPSessionsActive = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "smf_pfcp_sessions_active",
			Help: "Number of active PFCP sessions",
		},
	)

	SMFPFCPMessages = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "smf_pfcp_messages_total",
			Help: "Total number of PFCP messages",
		},
		[]string{"type", "direction"},
	)

	// QoS Flow metrics
	ActiveQoSFlows = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "smf_active_qos_flows",
			Help: "Number of active QoS flows",
		},
	)
)

// SetActivePDUSessions sets the number of active PDU sessions
func SetActivePDUSessions(count int) {
	ActivePDUSessions.Set(float64(count))
}

// RecordPDUSessionEstablishment records a PDU session establishment
func RecordPDUSessionEstablishment(result, dnn string) {
	PDUSessionEstablishments.WithLabelValues(result, dnn).Inc()
}

// RecordPDUSessionModification records a PDU session modification
func RecordPDUSessionModification(result string) {
	PDUSessionModifications.WithLabelValues(result).Inc()
}

// RecordPDUSessionRelease records a PDU session release
func RecordPDUSessionRelease(reason string) {
	PDUSessionReleases.WithLabelValues(reason).Inc()
}

// SetSMFPFCPSessionsActive sets the number of active PFCP sessions
func SetSMFPFCPSessionsActive(count int) {
	SMFPFCPSessionsActive.Set(float64(count))
}

// RecordSMFPFCPMessage records a PFCP message
func RecordSMFPFCPMessage(msgType, direction string) {
	SMFPFCPMessages.WithLabelValues(msgType, direction).Inc()
}

// SetActiveQoSFlows sets the number of active QoS flows
func SetActiveQoSFlows(count int) {
	ActiveQoSFlows.Set(float64(count))
}
