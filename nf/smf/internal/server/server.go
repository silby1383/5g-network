package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/your-org/5g-network/nf/smf/internal/config"
	"github.com/your-org/5g-network/nf/smf/internal/service"
	"go.uber.org/zap"
)

// SMFServer represents the SMF HTTP server
type SMFServer struct {
	config         *config.Config
	router         *chi.Mux
	server         *http.Server
	logger         *zap.Logger
	sessionService *service.SessionService
}

// NewSMFServer creates a new SMF HTTP server
func NewSMFServer(
	cfg *config.Config,
	sessionService *service.SessionService,
	logger *zap.Logger,
) *SMFServer {
	s := &SMFServer{
		config:         cfg,
		router:         chi.NewRouter(),
		logger:         logger,
		sessionService: sessionService,
	}

	s.setupRoutes()

	s.server = &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.SBI.IPv4, cfg.SBI.Port),
		Handler:      s.router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return s
}

// setupRoutes configures the API routes
func (s *SMFServer) setupRoutes() {
	s.router.Use(middleware.RequestID)
	s.router.Use(middleware.RealIP)
	s.router.Use(s.loggingMiddleware)
	s.router.Use(middleware.Recoverer)
	s.router.Use(middleware.Timeout(60 * time.Second))

	// Health & monitoring
	s.router.Get("/health", s.handleHealthCheck)
	s.router.Get("/ready", s.handleReadinessCheck)
	s.router.Get("/status", s.handleStatus)

	// 3GPP TS 29.502 - Nsmf_PDUSession API
	s.router.Route("/nsmf-pdusession/v1", func(r chi.Router) {
		// SM Contexts (PDU Sessions)
		r.Post("/sm-contexts", s.handleCreateSMContext)
		r.Put("/sm-contexts/{smContextRef}/modify", s.handleUpdateSMContext)
		r.Post("/sm-contexts/{smContextRef}/release", s.handleReleaseSMContext)
		r.Get("/sm-contexts/{smContextRef}", s.handleGetSMContext)
	})

	// Admin endpoints
	s.router.Route("/admin", func(r chi.Router) {
		r.Get("/sessions", s.handleListSessions)
		r.Get("/sessions/{supi}", s.handleGetSessionsBySUPI)
		r.Get("/stats", s.handleGetStats)
	})
}

// Start starts the HTTP server
func (s *SMFServer) Start() error {
	s.logger.Info("Starting SMF HTTP server",
		zap.String("address", s.server.Addr),
	)

	if s.config.SBI.TLS.Enabled {
		return s.server.ListenAndServeTLS(s.config.SBI.TLS.Cert, s.config.SBI.TLS.Key)
	}

	return s.server.ListenAndServe()
}

// Stop stops the HTTP server
func (s *SMFServer) Stop(ctx context.Context) error {
	s.logger.Info("Stopping SMF HTTP server")

	if s.server != nil {
		return s.server.Shutdown(ctx)
	}

	return nil
}

// loggingMiddleware logs HTTP requests
func (s *SMFServer) loggingMiddleware(next http.Handler) http.Handler {
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
		)
	})
}
