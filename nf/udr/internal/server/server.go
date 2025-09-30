package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/your-org/5g-network/nf/udr/internal/config"
	"github.com/your-org/5g-network/nf/udr/internal/repository"
	"go.uber.org/zap"
)

// UDRServer represents the UDR HTTP server
type UDRServer struct {
	config     *config.Config
	repository repository.Repository
	router     *chi.Mux
	httpServer *http.Server
	logger     *zap.Logger
}

// NewUDRServer creates a new UDR server instance
func NewUDRServer(cfg *config.Config, repo repository.Repository, logger *zap.Logger) (*UDRServer, error) {
	server := &UDRServer{
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
func (s *UDRServer) setupRoutes() {
	// Middleware
	s.router.Use(middleware.RequestID)
	s.router.Use(middleware.RealIP)
	s.router.Use(s.loggingMiddleware)
	s.router.Use(middleware.Recoverer)
	s.router.Use(middleware.Timeout(60 * time.Second))

	// Health endpoints
	s.router.Get("/health", s.handleHealth)
	s.router.Get("/ready", s.handleReady)
	s.router.Get("/status", s.handleStatus)

	// Data Repository Service (TS 29.504)
	s.router.Route("/nudr-dr/v1", func(r chi.Router) {
		// Subscription Data (TS 29.505)
		r.Route("/subscription-data", func(r chi.Router) {
			// Access and Mobility Data
			r.Get("/{supi}/provisioned-data/am-data", s.handleGetAMData)
			r.Put("/{supi}/provisioned-data/am-data", s.handleUpdateAMData)

			// Session Management Data
			r.Get("/{supi}/provisioned-data/sm-data", s.handleGetSMData)
			r.Put("/{supi}/provisioned-data/sm-data", s.handleUpdateSMData)

			// Authentication Subscription
			r.Get("/{supi}/authentication-data/authentication-subscription", s.handleGetAuthSubscription)
			r.Put("/{supi}/authentication-data/authentication-subscription", s.handleUpdateAuthSubscription)
			r.Patch("/{supi}/authentication-data/authentication-subscription/sqn", s.handleIncrementSQN)
		})

		// Policy Data (TS 29.519)
		r.Route("/policy-data", func(r chi.Router) {
			r.Get("/ues/{supi}/sm-data", s.handleGetPolicyData)
			r.Put("/ues/{supi}/sm-data", s.handleUpdatePolicyData)
		})

		// Exposure Data
		r.Route("/exposure-data", func(r chi.Router) {
			r.Get("/subs-to-notify", s.handleGetSubscriptions)
			r.Post("/subs-to-notify", s.handleCreateSubscription)
			r.Delete("/subs-to-notify/{subscriptionId}", s.handleDeleteSubscription)
		})
	})

	// Administrative endpoints
	s.router.Route("/admin", func(r chi.Router) {
		r.Get("/subscribers", s.handleListSubscribers)
		r.Post("/subscribers", s.handleCreateSubscriber)
		r.Get("/subscribers/{supi}", s.handleGetSubscriber)
		r.Put("/subscribers/{supi}", s.handlePutSubscriber)
		r.Delete("/subscribers/{supi}", s.handleDeleteSubscriber)
		r.Get("/stats", s.handleGetStats)
	})
}

// Start starts the HTTP server
func (s *UDRServer) Start(ctx context.Context) error {
	addr := fmt.Sprintf("%s:%d", s.config.SBI.BindAddress, s.config.SBI.Port)

	s.httpServer = &http.Server{
		Addr:         addr,
		Handler:      s.router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	s.logger.Info("Starting UDR HTTP server", zap.String("address", addr))

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
func (s *UDRServer) Stop(ctx context.Context) error {
	s.logger.Info("Stopping UDR server")

	if s.httpServer != nil {
		return s.httpServer.Shutdown(ctx)
	}

	return nil
}

// loggingMiddleware logs HTTP requests
func (s *UDRServer) loggingMiddleware(next http.Handler) http.Handler {
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
func (s *UDRServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"healthy"}`))
}

// handleReady handles readiness check requests
func (s *UDRServer) handleReady(w http.ResponseWriter, r *http.Request) {
	// Check if repository is ready
	if err := s.repository.Ping(r.Context()); err != nil {
		s.respondError(w, http.StatusServiceUnavailable, "repository unavailable", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ready"}`))
}

// handleStatus handles status requests
func (s *UDRServer) handleStatus(w http.ResponseWriter, r *http.Request) {
	stats, err := s.repository.GetStats(r.Context())
	if err != nil {
		s.respondError(w, http.StatusInternalServerError, "failed to get stats", err)
		return
	}

	s.respondJSON(w, http.StatusOK, map[string]interface{}{
		"udr_instance_id": s.config.NF.InstanceID,
		"udr_name":        s.config.NF.Name,
		"version":         "1.0.0",
		"stats":           stats,
	})
}

// respondJSON writes a JSON response
func (s *UDRServer) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	// In production, use proper JSON marshaling
	fmt.Fprintf(w, "%+v", data)
}

// respondError writes an error response
func (s *UDRServer) respondError(w http.ResponseWriter, status int, message string, err error) {
	s.logger.Error(message, zap.Error(err))

	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(status)

	fmt.Fprintf(w, `{"status":%d,"title":"%s","detail":"%s"}`, status, message, err.Error())
}
