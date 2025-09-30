package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/your-org/5g-network/nf/nrf/internal/config"
	"github.com/your-org/5g-network/nf/nrf/internal/repository"
	"go.uber.org/zap"
)

// NRFServer represents the NRF HTTP server
type NRFServer struct {
	config     *config.Config
	repository repository.Repository
	router     *chi.Mux
	httpServer *http.Server
	logger     *zap.Logger
}

// NewNRFServer creates a new NRF server instance
func NewNRFServer(cfg *config.Config, logger *zap.Logger) (*NRFServer, error) {
	// Create repository
	repo := repository.NewMemoryRepository(logger)

	server := &NRFServer{
		config:     cfg,
		repository: repo,
		router:     chi.NewRouter(),
		logger:     logger,
	}

	// Setup routes
	server.setupRoutes()

	return server, nil
}

// setupRoutes configures HTTP routes
func (s *NRFServer) setupRoutes() {
	// Middleware
	s.router.Use(middleware.RequestID)
	s.router.Use(middleware.RealIP)
	s.router.Use(s.loggingMiddleware)
	s.router.Use(middleware.Recoverer)
	s.router.Use(middleware.Timeout(60 * time.Second))

	// Health check
	s.router.Get("/health", s.handleHealth)
	s.router.Get("/ready", s.handleReady)

	// NF Management Service (TS 29.510, Clause 5.2.2)
	s.router.Route("/nnrf-nfm/v1", func(r chi.Router) {
		// NF Instance Management
		r.Put("/nf-instances/{nfInstanceId}", s.handleNFRegister)
		r.Patch("/nf-instances/{nfInstanceId}", s.handleNFUpdate)
		r.Delete("/nf-instances/{nfInstanceId}", s.handleNFDeregister)
		r.Get("/nf-instances/{nfInstanceId}", s.handleNFGet)
		r.Get("/nf-instances", s.handleNFList)

		// Heartbeat (Keep-alive)
		r.Put("/nf-instances/{nfInstanceId}/heartbeat", s.handleHeartbeat)

		// Subscriptions
		r.Post("/subscriptions", s.handleSubscribe)
		r.Delete("/subscriptions/{subscriptionId}", s.handleUnsubscribe)
		r.Get("/subscriptions/{subscriptionId}", s.handleGetSubscription)
	})

	// NF Discovery Service (TS 29.510, Clause 5.2.3)
	s.router.Route("/nnrf-disc/v1", func(r chi.Router) {
		r.Get("/nf-instances", s.handleNFDiscover)
	})

	// Status endpoint
	s.router.Get("/status", s.handleStatus)
}

// Start starts the HTTP server
func (s *NRFServer) Start(ctx context.Context) error {
	addr := fmt.Sprintf("%s:%d", s.config.SBI.BindAddress, s.config.SBI.Port)

	s.httpServer = &http.Server{
		Addr:         addr,
		Handler:      s.router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	s.logger.Info("Starting HTTP server", zap.String("address", addr))

	// Start server
	if s.config.SBI.TLS.Enabled {
		return s.httpServer.ListenAndServeTLS(
			s.config.SBI.TLS.CertFile,
			s.config.SBI.TLS.KeyFile,
		)
	}

	return s.httpServer.ListenAndServe()
}

// Stop stops the HTTP server gracefully
func (s *NRFServer) Stop(ctx context.Context) error {
	s.logger.Info("Stopping NRF server")

	// Close repository
	if memRepo, ok := s.repository.(*repository.MemoryRepository); ok {
		memRepo.Close()
	}

	if s.httpServer != nil {
		return s.httpServer.Shutdown(ctx)
	}

	return nil
}

// loggingMiddleware logs HTTP requests
func (s *NRFServer) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Wrap response writer to capture status code
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

		// Call next handler
		next.ServeHTTP(ww, r)

		// Log request
		s.logger.Info("HTTP request",
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
			zap.Int("status", ww.Status()),
			zap.Duration("duration", time.Since(start)),
			zap.String("remote_addr", r.RemoteAddr),
			zap.String("request_id", middleware.GetReqID(r.Context())),
		)
	})
}

// handleHealth handles health check requests
func (s *NRFServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"healthy"}`))
}

// handleReady handles readiness check requests
func (s *NRFServer) handleReady(w http.ResponseWriter, r *http.Request) {
	// Check if repository is ready
	_, err := s.repository.GetStats(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte(`{"status":"not ready","error":"repository unavailable"}`))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ready"}`))
}

// handleStatus handles status requests
func (s *NRFServer) handleStatus(w http.ResponseWriter, r *http.Request) {
	stats, err := s.repository.GetStats(r.Context())
	if err != nil {
		s.respondError(w, http.StatusInternalServerError, "failed to get stats", err)
		return
	}

	s.respondJSON(w, http.StatusOK, map[string]interface{}{
		"nrf_instance_id": s.config.NF.InstanceID,
		"nrf_name":        s.config.NF.Name,
		"version":         "1.0.0",
		"stats":           stats,
	})
}

// respondJSON writes a JSON response
func (s *NRFServer) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	// In production, use proper JSON marshaling
	// For now, simple response
	fmt.Fprintf(w, "%+v", data)
}

// respondError writes an error response
func (s *NRFServer) respondError(w http.ResponseWriter, status int, message string, err error) {
	s.logger.Error(message, zap.Error(err))

	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(status)

	fmt.Fprintf(w, `{"status":%d,"title":"%s","detail":"%s"}`, status, message, err.Error())
}
