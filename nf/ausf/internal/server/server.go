package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/your-org/5g-network/nf/ausf/internal/config"
	"github.com/your-org/5g-network/nf/ausf/internal/service"
	"go.uber.org/zap"
)

// AUSFServer represents the AUSF HTTP server
type AUSFServer struct {
	config *config.Config
	router *chi.Mux
	server *http.Server
	logger *zap.Logger

	// Services
	authService *service.AuthenticationService
}

// NewServer creates a new AUSF server
func NewServer(
	cfg *config.Config,
	authService *service.AuthenticationService,
	logger *zap.Logger,
) *AUSFServer {
	s := &AUSFServer{
		config:      cfg,
		router:      chi.NewRouter(),
		logger:      logger,
		authService: authService,
	}

	s.setupMiddleware()
	s.setupRoutes()

	return s
}

// setupMiddleware configures HTTP middleware
func (s *AUSFServer) setupMiddleware() {
	s.router.Use(middleware.RequestID)
	s.router.Use(middleware.RealIP)
	s.router.Use(s.loggingMiddleware)
	s.router.Use(middleware.Recoverer)
	s.router.Use(middleware.Timeout(60 * time.Second))
}

// setupRoutes configures HTTP routes
func (s *AUSFServer) setupRoutes() {
	// Health and status
	s.router.Get("/health", s.handleHealth)
	s.router.Get("/ready", s.handleReady)
	s.router.Get("/status", s.handleStatus)

	// Nausf_UEAuthentication service (TS 29.509)
	s.router.Route("/nausf-auth/v1", func(r chi.Router) {
		// UE authentication initiation
		r.Post("/ue-authentications", s.handleUEAuthenticationRequest)

		// 5G-AKA confirmation
		r.Put("/ue-authentications/{authCtxId}/5g-aka-confirmation", s.handleConfirm5gAkaAuth)

		// EAP session (future)
		// r.Post("/ue-authentications/{authCtxId}/eap-session", s.handleEAPSession)
	})

	// Admin endpoints
	s.router.Route("/admin", func(r chi.Router) {
		r.Get("/stats", s.handleGetStats)
		r.Get("/test/auth-context/{authCtxId}", s.handleGetAuthContext) // Test only!
	})
}

// Start starts the HTTP server
func (s *AUSFServer) Start() error {
	addr := fmt.Sprintf("%s:%d", s.config.SBI.BindAddress, s.config.SBI.Port)

	s.server = &http.Server{
		Addr:         addr,
		Handler:      s.router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	s.logger.Info("Starting AUSF HTTP server", zap.String("address", addr))

	if s.config.SBI.TLS.Enabled {
		return s.server.ListenAndServeTLS(s.config.SBI.TLS.CertFile, s.config.SBI.TLS.KeyFile)
	}

	return s.server.ListenAndServe()
}

// Stop gracefully stops the HTTP server
func (s *AUSFServer) Stop(ctx context.Context) error {
	s.logger.Info("Stopping AUSF HTTP server")

	if s.server != nil {
		return s.server.Shutdown(ctx)
	}

	return nil
}

// Middleware

func (s *AUSFServer) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
		next.ServeHTTP(ww, r)

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

// Helper functions

func (s *AUSFServer) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		s.logger.Error("Failed to encode JSON response", zap.Error(err))
	}
}

func (s *AUSFServer) respondError(w http.ResponseWriter, status int, message string, err error) {
	s.logger.Error(message, zap.Error(err))

	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(status)

	response := map[string]interface{}{
		"status": status,
		"title":  message,
	}

	if err != nil {
		response["detail"] = err.Error()
	}

	json.NewEncoder(w).Encode(response)
}

// Health check handlers

func (s *AUSFServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	s.respondJSON(w, http.StatusOK, map[string]string{
		"status": "healthy",
	})
}

func (s *AUSFServer) handleReady(w http.ResponseWriter, r *http.Request) {
	// Check if AUSF is ready to serve requests
	s.respondJSON(w, http.StatusOK, map[string]string{
		"status": "ready",
	})
}

func (s *AUSFServer) handleStatus(w http.ResponseWriter, r *http.Request) {
	stats := s.authService.GetStats()

	s.respondJSON(w, http.StatusOK, map[string]interface{}{
		"service": "AUSF",
		"version": "1.0.0",
		"stats":   stats,
	})
}
