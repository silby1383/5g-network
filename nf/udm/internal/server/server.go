package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/your-org/5g-network/nf/udm/internal/config"
	"github.com/your-org/5g-network/nf/udm/internal/service"
	"go.uber.org/zap"
)

// UDMServer represents the UDM HTTP server
type UDMServer struct {
	config *config.Config
	router *chi.Mux
	server *http.Server
	logger *zap.Logger

	// Services
	authService *service.AuthenticationService
	sdmService  *service.SDMService
	uecmService *service.UECMService
}

// NewServer creates a new UDM server
func NewServer(
	cfg *config.Config,
	authService *service.AuthenticationService,
	sdmService *service.SDMService,
	uecmService *service.UECMService,
	logger *zap.Logger,
) *UDMServer {
	s := &UDMServer{
		config:      cfg,
		router:      chi.NewRouter(),
		logger:      logger,
		authService: authService,
		sdmService:  sdmService,
		uecmService: uecmService,
	}

	s.setupMiddleware()
	s.setupRoutes()

	return s
}

// setupMiddleware configures HTTP middleware
func (s *UDMServer) setupMiddleware() {
	s.router.Use(middleware.RequestID)
	s.router.Use(middleware.RealIP)
	s.router.Use(s.loggingMiddleware)
	s.router.Use(middleware.Recoverer)
	s.router.Use(middleware.Timeout(60 * time.Second))
}

// setupRoutes configures HTTP routes
func (s *UDMServer) setupRoutes() {
	// Health and status
	s.router.Get("/health", s.handleHealth)
	s.router.Get("/ready", s.handleReady)
	s.router.Get("/status", s.handleStatus)

	// Nudm_UEAuthentication service (TS 29.503)
	s.router.Route("/nudm-ueau/v1", func(r chi.Router) {
		r.Post("/supi/{supi}/security-information/generate-auth-data", s.handleGenerateAuthData)
		r.Post("/supi/{supi}/auth-events", s.handleConfirmAuth)
	})

	// Nudm_SDM service (TS 29.503)
	s.router.Route("/nudm-sdm/v1", func(r chi.Router) {
		// Access and Mobility subscription data
		r.Get("/supi/{supi}/am-data", s.handleGetAMData)

		// Session Management subscription data
		r.Get("/supi/{supi}/sm-data", s.handleGetSMData)
		r.Get("/supi/{supi}/{servingPlmnId}/sm-data", s.handleGetSMDataWithPlmn)

		// Subscriptions
		r.Post("/supi/{supi}/sdm-subscriptions", s.handleSubscribeSDM)
		r.Delete("/supi/{supi}/sdm-subscriptions/{subscriptionId}", s.handleUnsubscribeSDM)
	})

	// Nudm_UECM service (TS 29.503)
	s.router.Route("/nudm-uecm/v1", func(r chi.Router) {
		// 3GPP access registration
		r.Put("/supi/{supi}/registrations/amf-3gpp-access", s.handleRegisterAMF3GPP)
		r.Patch("/supi/{supi}/registrations/amf-3gpp-access", s.handleUpdateAMF3GPP)
		r.Get("/supi/{supi}/registrations/amf-3gpp-access", s.handleGetAMF3GPP)
		r.Delete("/supi/{supi}/registrations/amf-3gpp-access", s.handleDeregisterAMF3GPP)

		// UE context
		r.Get("/supi/{supi}/ue-context", s.handleGetUEContext)
	})

	// Admin endpoints
	s.router.Route("/admin", func(r chi.Router) {
		r.Get("/stats", s.handleGetStats)
	})
}

// Start starts the HTTP server
func (s *UDMServer) Start() error {
	addr := fmt.Sprintf("%s:%d", s.config.SBI.BindAddress, s.config.SBI.Port)

	s.server = &http.Server{
		Addr:         addr,
		Handler:      s.router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	s.logger.Info("Starting UDM HTTP server", zap.String("address", addr))

	if s.config.SBI.TLS.Enabled {
		return s.server.ListenAndServeTLS(s.config.SBI.TLS.CertFile, s.config.SBI.TLS.KeyFile)
	}

	return s.server.ListenAndServe()
}

// Stop gracefully stops the HTTP server
func (s *UDMServer) Stop(ctx context.Context) error {
	s.logger.Info("Stopping UDM HTTP server")

	if s.server != nil {
		return s.server.Shutdown(ctx)
	}

	return nil
}

// Middleware

func (s *UDMServer) loggingMiddleware(next http.Handler) http.Handler {
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

func (s *UDMServer) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		s.logger.Error("Failed to encode JSON response", zap.Error(err))
	}
}

func (s *UDMServer) respondError(w http.ResponseWriter, status int, message string, err error) {
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

func (s *UDMServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	s.respondJSON(w, http.StatusOK, map[string]string{
		"status": "healthy",
	})
}

func (s *UDMServer) handleReady(w http.ResponseWriter, r *http.Request) {
	// Check if UDM is ready to serve requests
	// Could check UDR connectivity, etc.
	s.respondJSON(w, http.StatusOK, map[string]string{
		"status": "ready",
	})
}

func (s *UDMServer) handleStatus(w http.ResponseWriter, r *http.Request) {
	stats := s.uecmService.GetStats()

	s.respondJSON(w, http.StatusOK, map[string]interface{}{
		"service": "UDM",
		"version": "1.0.0",
		"stats":   stats,
	})
}
