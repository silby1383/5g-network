package ebpf

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/cilium/ebpf"
	"github.com/cilium/ebpf/link"
	"github.com/cilium/ebpf/perf"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

//go:generate go run github.com/cilium/ebpf/cmd/bpf2go -cc clang -cflags "-O2 -g -Wall -Werror" tracehttp trace_http.c -- -I/usr/include/bpf

// HTTPEvent represents an HTTP request/response captured by eBPF
type HTTPEvent struct {
	TimestampNS   uint64
	PID           uint32
	TID           uint32
	Method        [16]byte
	Path          [128]byte
	Traceparent   [55]byte
	StatusCode    uint16
	DurationNS    uint64
	ContentLength uint32
}

// TraceContext represents W3C trace context
type TraceContext struct {
	TraceID [16]byte
	SpanID  [8]byte
	Flags   uint8
}

// EBPFTracer manages eBPF-based tracing for a network function
type EBPFTracer struct {
	nfName     string
	nfBinary   string
	collection *ebpf.Collection
	links      []link.Link
	reader     *perf.Reader
	logger     *zap.Logger
	tracer     trace.Tracer
	eventChan  chan *HTTPEvent
	stopChan   chan struct{}
}

// Config holds eBPF tracer configuration
type Config struct {
	NFName    string   // Network function name (e.g., "amf", "smf")
	NFBinary  string   // Path to NF binary
	Functions []string // Functions to trace
}

// NewEBPFTracer creates a new eBPF tracer for a network function
func NewEBPFTracer(config *Config, logger *zap.Logger) (*EBPFTracer, error) {
	return &EBPFTracer{
		nfName:    config.NFName,
		nfBinary:  config.NFBinary,
		logger:    logger,
		tracer:    otel.Tracer("ebpf-tracer"),
		eventChan: make(chan *HTTPEvent, 10000),
		stopChan:  make(chan struct{}),
	}, nil
}

// Load loads and attaches eBPF programs
func (t *EBPFTracer) Load(ctx context.Context) error {
	ctx, span := t.tracer.Start(ctx, "EBPFTracer.Load")
	defer span.End()

	t.logger.Info("Loading eBPF programs", zap.String("nf", t.nfName))

	// Load compiled eBPF object
	spec, err := loadTracehttp()
	if err != nil {
		return fmt.Errorf("failed to load eBPF spec: %w", err)
	}

	// Load into kernel
	coll, err := ebpf.NewCollection(spec)
	if err != nil {
		return fmt.Errorf("failed to create eBPF collection: %w", err)
	}
	t.collection = coll

	// Attach HTTP handler start uprobe
	if err := t.attachHTTPHandlerStart(); err != nil {
		t.logger.Warn("Failed to attach HTTP handler start probe", zap.Error(err))
		// Continue even if this fails - some NFs might not have HTTP handlers
	}

	// Attach HTTP handler end uprobe
	if err := t.attachHTTPHandlerEnd(); err != nil {
		t.logger.Warn("Failed to attach HTTP handler end probe", zap.Error(err))
	}

	// Attach TCP send/recv kprobes for network-level tracing
	if err := t.attachNetworkProbes(); err != nil {
		t.logger.Warn("Failed to attach network probes", zap.Error(err))
	}

	// Create perf event reader
	rd, err := perf.NewReader(t.collection.Maps["http_events"], 4096*os.Getpagesize())
	if err != nil {
		return fmt.Errorf("failed to create perf reader: %w", err)
	}
	t.reader = rd

	// Start event processing
	go t.processEvents()

	span.SetAttributes(
		attribute.String("nf_name", t.nfName),
		attribute.String("nf_binary", t.nfBinary),
	)

	t.logger.Info("eBPF programs loaded successfully")
	return nil
}

// attachHTTPHandlerStart attaches uprobe to HTTP handler start
func (t *EBPFTracer) attachHTTPHandlerStart() error {
	prog := t.collection.Programs["trace_http_request_start"]
	if prog == nil {
		return fmt.Errorf("program trace_http_request_start not found")
	}

	// Find the HTTP handler function symbol in the binary
	// Common patterns: HandleHTTPRequest, ServeHTTP, http.(*ServeMux).ServeHTTP
	symbols := []string{
		"HandleHTTPRequest",
		"ServeHTTP",
		"main.handleRequest",
	}

	for _, symbol := range symbols {
		l, err := link.Uprobe(t.nfBinary, symbol, prog, nil)
		if err != nil {
			continue
		}
		t.links = append(t.links, l)
		t.logger.Info("Attached HTTP handler start probe", zap.String("symbol", symbol))
		return nil
	}

	return fmt.Errorf("failed to attach to any HTTP handler symbol")
}

// attachHTTPHandlerEnd attaches uprobe to HTTP handler end
func (t *EBPFTracer) attachHTTPHandlerEnd() error {
	prog := t.collection.Programs["trace_http_request_end"]
	if prog == nil {
		return fmt.Errorf("program trace_http_request_end not found")
	}

	symbols := []string{
		"HandleHTTPRequest",
		"ServeHTTP",
		"main.handleRequest",
	}

	for _, symbol := range symbols {
		l, err := link.Uretprobe(t.nfBinary, symbol, prog, nil)
		if err != nil {
			continue
		}
		t.links = append(t.links, l)
		t.logger.Info("Attached HTTP handler end probe", zap.String("symbol", symbol))
		return nil
	}

	return fmt.Errorf("failed to attach to any HTTP handler symbol")
}

// attachNetworkProbes attaches kprobes for network tracing
func (t *EBPFTracer) attachNetworkProbes() error {
	// Attach to tcp_sendmsg
	if prog := t.collection.Programs["trace_tcp_sendmsg"]; prog != nil {
		l, err := link.Kprobe("tcp_sendmsg", prog, nil)
		if err != nil {
			return fmt.Errorf("failed to attach tcp_sendmsg: %w", err)
		}
		t.links = append(t.links, l)
		t.logger.Info("Attached tcp_sendmsg kprobe")
	}

	// Attach to tcp_recvmsg
	if prog := t.collection.Programs["trace_tcp_recvmsg"]; prog != nil {
		l, err := link.Kprobe("tcp_recvmsg", prog, nil)
		if err != nil {
			return fmt.Errorf("failed to attach tcp_recvmsg: %w", err)
		}
		t.links = append(t.links, l)
		t.logger.Info("Attached tcp_recvmsg kprobe")
	}

	return nil
}

// processEvents reads events from the perf buffer and processes them
func (t *EBPFTracer) processEvents() {
	t.logger.Info("Starting eBPF event processing")

	for {
		select {
		case <-t.stopChan:
			t.logger.Info("Stopping eBPF event processing")
			return
		default:
		}

		record, err := t.reader.Read()
		if err != nil {
			if perf.IsClosed(err) {
				return
			}
			t.logger.Error("Error reading from perf buffer", zap.Error(err))
			continue
		}

		if record.LostSamples > 0 {
			t.logger.Warn("Lost perf samples", zap.Uint64("count", record.LostSamples))
		}

		// Parse event
		var event HTTPEvent
		if err := binary.Read(bytes.NewReader(record.RawSample), binary.LittleEndian, &event); err != nil {
			t.logger.Error("Error parsing event", zap.Error(err))
			continue
		}

		// Send to channel for processing
		select {
		case t.eventChan <- &event:
		default:
			t.logger.Warn("Event channel full, dropping event")
		}

		// Export to OpenTelemetry
		t.exportToOTel(&event)
	}
}

// exportToOTel converts eBPF event to OpenTelemetry span
func (t *EBPFTracer) exportToOTel(event *HTTPEvent) {
	// Parse method and path
	method := string(bytes.TrimRight(event.Method[:], "\x00"))
	path := string(bytes.TrimRight(event.Path[:], "\x00"))
	traceparent := string(bytes.TrimRight(event.Traceparent[:], "\x00"))

	// Parse W3C trace context from traceparent header
	var traceID trace.TraceID
	var spanID trace.SpanID
	if traceparent != "" {
		// Parse: 00-{trace-id}-{span-id}-{flags}
		// For simplicity, using a basic parser
		// Production would use proper W3C trace context parser
		copy(traceID[:], traceparent[3:35])
		copy(spanID[:], traceparent[36:52])
	}

	// Create span context
	spanContext := trace.NewSpanContext(trace.SpanContextConfig{
		TraceID:    traceID,
		SpanID:     spanID,
		TraceFlags: trace.FlagsSampled,
	})

	ctx := trace.ContextWithSpanContext(context.Background(), spanContext)

	// Create span
	_, span := t.tracer.Start(ctx, fmt.Sprintf("HTTP %s %s", method, path),
		trace.WithTimestamp(nsToTime(event.TimestampNS)),
		trace.WithSpanKind(trace.SpanKindServer),
	)

	// Add attributes
	span.SetAttributes(
		attribute.String("nf.name", t.nfName),
		attribute.String("http.method", method),
		attribute.String("http.target", path),
		attribute.Int("http.status_code", int(event.StatusCode)),
		attribute.Int64("http.duration_ns", int64(event.DurationNS)),
		attribute.Int("http.content_length", int(event.ContentLength)),
		attribute.Int("process.pid", int(event.PID)),
		attribute.Int("thread.id", int(event.TID)),
		attribute.String("ebpf.source", "kernel"),
	)

	// End span with calculated duration
	span.End(trace.WithTimestamp(nsToTime(event.TimestampNS + event.DurationNS)))

	t.logger.Debug("eBPF event exported to OpenTelemetry",
		zap.String("method", method),
		zap.String("path", path),
		zap.Uint16("status", event.StatusCode),
		zap.Uint64("duration_ns", event.DurationNS),
	)
}

// Close closes the eBPF tracer and cleans up resources
func (t *EBPFTracer) Close() error {
	t.logger.Info("Closing eBPF tracer")

	// Stop event processing
	close(t.stopChan)

	// Close perf reader
	if t.reader != nil {
		if err := t.reader.Close(); err != nil {
			t.logger.Error("Error closing perf reader", zap.Error(err))
		}
	}

	// Detach all probes
	for _, l := range t.links {
		if err := l.Close(); err != nil {
			t.logger.Error("Error closing link", zap.Error(err))
		}
	}

	// Close collection
	if t.collection != nil {
		if err := t.collection.Close(); err != nil {
			t.logger.Error("Error closing eBPF collection", zap.Error(err))
		}
	}

	t.logger.Info("eBPF tracer closed")
	return nil
}

// GetEventChannel returns the channel for receiving HTTP events
func (t *EBPFTracer) GetEventChannel() <-chan *HTTPEvent {
	return t.eventChan
}

// Helper function to convert nanoseconds to time.Time
func nsToTime(ns uint64) time.Time {
	return time.Unix(0, int64(ns))
}

// AttachToProcess attaches eBPF programs to a running process
func AttachToProcess(pid int, config *Config, logger *zap.Logger) (*EBPFTracer, error) {
	// Find binary path from PID
	binaryPath, err := filepath.EvalSymlinks(fmt.Sprintf("/proc/%d/exe", pid))
	if err != nil {
		return nil, fmt.Errorf("failed to find binary for PID %d: %w", pid, err)
	}

	config.NFBinary = binaryPath

	tracer, err := NewEBPFTracer(config, logger)
	if err != nil {
		return nil, err
	}

	if err := tracer.Load(context.Background()); err != nil {
		return nil, err
	}

	return tracer, nil
}
