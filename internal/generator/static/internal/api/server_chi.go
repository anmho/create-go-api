package api

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// ChiServer encapsulates the HTTP server using Chi
type ChiServer struct {
	router *chi.Mux
}

// NewChiServer creates a new HTTP server
func NewChiServer() *ChiServer {
	s := &ChiServer{
		router: chi.NewRouter(),
	}

	s.setupMiddleware()

	return s
}

// setupMiddleware configures Chi middleware
func (s *ChiServer) setupMiddleware() {
	s.router.Use(middleware.RequestID)
	s.router.Use(middleware.Logger)
	s.router.Use(middleware.Recoverer)
	s.router.Use(middleware.Heartbeat("/health"))
}

// Handler returns the HTTP handler
func (s *ChiServer) Handler() http.Handler {
	return s.router
}

// ServeHTTP implements http.Handler
func (s *ChiServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

// Shutdown gracefully shuts down the server (no-op for Chi, handled by http.Server)
func (s *ChiServer) Shutdown(ctx context.Context) error {
	// Chi doesn't have its own shutdown logic, this is handled by http.Server
	return nil
}

// Router returns the Chi router
func (s *ChiServer) Router() *chi.Mux {
	return s.router
}

