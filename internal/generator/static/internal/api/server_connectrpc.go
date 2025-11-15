package api

import (
	"context"
	"net/http"

	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

// ConnectRPCServer encapsulates the gRPC server using ConnectRPC
type ConnectRPCServer struct {
	mux *http.ServeMux
}

// NewConnectRPCServer creates a new gRPC server
func NewConnectRPCServer() *ConnectRPCServer {
	return &ConnectRPCServer{
		mux: http.NewServeMux(),
	}
}

// Handler returns the HTTP handler for the gRPC server
func (s *ConnectRPCServer) Handler() http.Handler {
	return h2c.NewHandler(s.mux, &http2.Server{})
}

// ServeHTTP implements http.Handler
func (s *ConnectRPCServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.Handler().ServeHTTP(w, r)
}

// Shutdown gracefully shuts down the server (no-op for ConnectRPC, handled by http.Server)
func (s *ConnectRPCServer) Shutdown(ctx context.Context) error {
	// ConnectRPC doesn't have its own shutdown logic, this is handled by http.Server
	return nil
}

// Mux returns the HTTP mux
func (s *ConnectRPCServer) Mux() *http.ServeMux {
	return s.mux
}

