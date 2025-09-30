package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/your-org/5g-network/nf/amf/internal/config"
	amfcontext "github.com/your-org/5g-network/nf/amf/internal/context"
	"github.com/your-org/5g-network/nf/amf/internal/service"
	"go.uber.org/zap"
)

// AMFServer represents the AMF HTTP server
type AMFServer struct {
	config             *config.Config
	router             *chi.Mux
	server             *http.Server
	logger             *zap.Logger

	// Services
	registrationService *service.RegistrationService
	contextManager      *amfcontext.UEContextManager
}

// NewServer creates a new AMF server
func NewServer(
	cfg *config.Config,
	registrationService *service.RegistrationService,
	contextManager *amfcontext.UEContextManager,
	logger *zap.Logger,
) *AMFServer {
	s := &AMFServer{
		config:              cfg,
		router:              chi.NewRouter(),
		logger:              logger,
		registrationService: registrationService,
		contextManager:      contextManager,
	}

	s.setupMiddleware()
	s.setupRoutes()

	return s
}

// setupMiddleware configures HTTP middleware
func (s *AMFServer) setupMiddleware() {
	s.router.Use(middleware.RequestID)
	s.router.Use(middleware.RealIP)
	s.router.Use(s.loggingMiddleware)
	s.router.Use(middleware.Recoverer)
	s.router.Use(middleware.Timeout(60 * time.Second))
}

// setupRoutes configures HTTP routes
func (s *AMFServer) setupRoutes() {
	// Health and status
	s.router.Get("/health", s.handleHealth)
	s.router.Get("/ready", s.handleReady)
	s.router.Get("/status", s.handleStatus)

	// Namf_Communication service (TS 29.518)
	s.router.Route("/namf-comm/v1", func(r chi.Router) {
		// UE Context Management
		r.Get("/ue-contexts/{ueContextId}", s.handleGetUEContext)
		r.Post("/ue-contexts/{ueContextId}/release", s.handleReleaseUEContext)
		
		// N1 Message Transfer
		r.Post("/ue-contexts/{ueContextId}/n1-n2-messages", s.handleN1N2Transfer)
	})

	// UE Authentication (AMF-specific, not in 3GPP but useful for testing)
	s.router.Route("/namf-auth/v1", func(r chi.Router) {
		r.Post("/authenticate", s.handleAuthenticationRequest)
		r.Put("/authenticate/{authCtxId}/confirm", s.handleAuthenticationConfirm)
	})

	// UE Registration (AMF-specific, not in 3GPP but useful for testing)
	s.router.Route("/namf-reg/v1", func(r chi.Router) {
		r.Post("/register", s.handleRegistrationRequest)
		r.Delete("/ue-contexts/{supi}", s.handleDeregistration)
	})

	// Admin endpoints
	s.router.Route("/admin", func(r chi.Router) {
		r.Get("/ue-contexts", s.handleListUEContexts)
		r.Get("/stats", s.handleGetStats)
	})
}

// Start starts the HTTP server
func (s *AMFServer) Start() error {
	addr := fmt.Sprintf("%s:%d", s.config.SBI.BindAddress, s.config.SBI.Port)

	s.server = &http.Server{
		Addr:         addr,
		Handler:      s.router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	s.logger.Info("Starting AMF HTTP server", zap.String("address", addr))

	if s.config.SBI.TLS.Enabled {
		return s.server.ListenAndServeTLS(s.config.SBI.TLS.CertFile, s.config.SBI.TLS.KeyFile)
	}

	return s.server.ListenAndServe()
}

// Stop gracefully stops the HTTP server
func (s *AMFServer) Stop(ctx context.Context) error {
	s.logger.Info("Stopping AMF HTTP server")
	
	if s.server != nil {
		return s.server.Shutdown(ctx)
	}
	
	return nil
}

// Middleware

func (s *AMFServer) loggingMiddleware(next http.Handler) http.Handler {
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

func (s *AMFServer) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		s.logger.Error("Failed to encode JSON response", zap.Error(err))
	}
}

func (s *AMFServer) respondError(w http.ResponseWriter, status int, message string, err error) {
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

func (s *AMFServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	s.respondJSON(w, http.StatusOK, map[string]string{
		"status": "healthy",
	})
}

func (s *AMFServer) handleReady(w http.ResponseWriter, r *http.Request) {
	s.respondJSON(w, http.StatusOK, map[string]string{
		"status": "ready",
	})
}

func (s *AMFServer) handleStatus(w http.ResponseWriter, r *http.Request) {
	stats := s.registrationService.GetRegistrationStats()
	
	s.respondJSON(w, http.StatusOK, map[string]interface{}{
		"service": "AMF",
		"version": "1.0.0",
		"guami":   s.config.GetGUAMI(),
		"stats":   stats,
	})
}