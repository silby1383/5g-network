package server

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/your-org/5g-network/nf/upf/internal/config"
	upfcontext "github.com/your-org/5g-network/nf/upf/internal/context"
	"github.com/your-org/5g-network/nf/upf/internal/gtpu"
	"go.uber.org/zap"
)

// Server represents the UPF admin/monitoring HTTP server
type Server struct {
	config      *config.Config
	router      *chi.Mux
	httpServer  *http.Server
	upfContext  *upfcontext.UPFContext
	gtpuHandler *gtpu.GTPUHandler
	logger      *zap.Logger
}

// NewServer creates a new UPF server
func NewServer(cfg *config.Config, upfCtx *upfcontext.UPFContext, gtpuHandler *gtpu.GTPUHandler, logger *zap.Logger) *Server {
	s := &Server{
		config:      cfg,
		router:      chi.NewRouter(),
		upfContext:  upfCtx,
		gtpuHandler: gtpuHandler,
		logger:      logger,
	}

	s.setupRoutes()
	return s
}

// setupRoutes configures the API routes
func (s *Server) setupRoutes() {
	s.router.Use(middleware.RequestID)
	s.router.Use(middleware.RealIP)
	s.router.Use(middleware.Logger)
	s.router.Use(middleware.Recoverer)
	s.router.Use(middleware.Timeout(60 * time.Second))

	s.router.Get("/health", s.handleHealthCheck)
	s.router.Get("/ready", s.handleReadinessCheck)
	s.router.Get("/status", s.handleStatus)
	s.router.Get("/sessions", s.handleGetSessions)
	s.router.Get("/stats", s.handleGetStats)
}

// Start starts the HTTP server
func (s *Server) Start() error {
	addr := ":9096" // Admin port

	s.httpServer = &http.Server{
		Addr:         addr,
		Handler:      s.router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	s.logger.Info("Starting UPF admin server", zap.String("address", addr))
	return s.httpServer.ListenAndServe()
}

// Stop gracefully stops the server
func (s *Server) Stop(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}

// handleHealthCheck handles health check requests
func (s *Server) handleHealthCheck(w http.ResponseWriter, r *http.Request) {
	s.respondJSON(w, http.StatusOK, map[string]string{
		"status": "healthy",
	})
}

// handleReadinessCheck handles readiness check requests
func (s *Server) handleReadinessCheck(w http.ResponseWriter, r *http.Request) {
	s.respondJSON(w, http.StatusOK, map[string]string{
		"status": "ready",
	})
}

// handleStatus handles status requests
func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	s.respondJSON(w, http.StatusOK, map[string]interface{}{
		"upf_instance_id": s.config.NF.InstanceID,
		"upf_name":        s.config.NF.Name,
		"node_id":         s.config.PFCP.NodeID,
		"stats":           s.upfContext.GetStats(),
		"version":         "1.0.0",
	})
}

// handleGetSessions returns all active sessions
func (s *Server) handleGetSessions(w http.ResponseWriter, r *http.Request) {
	sessions := s.upfContext.GetAllSessions()

	sessionList := make([]map[string]interface{}, 0, len(sessions))
	for _, session := range sessions {
		sessionList = append(sessionList, map[string]interface{}{
			"seid":          session.SEID,
			"ue_address":    session.UEAddress.String(),
			"upf_teid":      session.UPFTEID,
			"gnb_teid":      session.GNBTEID,
			"dnn":           session.DNN,
			"created_at":    session.CreatedAt,
			"last_activity": session.LastActivity,
		})
	}

	s.respondJSON(w, http.StatusOK, map[string]interface{}{
		"sessions": sessionList,
		"count":    len(sessionList),
	})
}

// handleGetStats returns GTP-U statistics
func (s *Server) handleGetStats(w http.ResponseWriter, r *http.Request) {
	gtpuStats := s.gtpuHandler.GetStats()
	upfStats := s.upfContext.GetStats()

	s.respondJSON(w, http.StatusOK, map[string]interface{}{
		"gtpu": map[string]interface{}{
			"uplink_packets":   gtpuStats.UplinkPackets,
			"downlink_packets": gtpuStats.DownlinkPackets,
			"uplink_bytes":     gtpuStats.UplinkBytes,
			"downlink_bytes":   gtpuStats.DownlinkBytes,
			"dropped_packets":  gtpuStats.DroppedPackets,
		},
		"sessions": upfStats,
	})
}

// respondJSON writes a JSON response
func (s *Server) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if data != nil {
		if err := json.NewEncoder(w).Encode(data); err != nil {
			s.logger.Error("Failed to encode JSON response", zap.Error(err))
		}
	}
}
