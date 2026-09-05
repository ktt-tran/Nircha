// Package server wires the HTTP routes and middleware together and owns
// the underlying http.Server's lifecycle (start, graceful shutdown).
package server

import (
	"context"
	"net/http"
	"time"

	"search-engine/searcher/config"
	"search-engine/searcher/handlers"
)

type Server struct {
	httpServer *http.Server
}

func NewServer(handler *handlers.Handler, cfg *config.Config) *Server {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/health", handler.HandleHealth)
	mux.HandleFunc("GET /api/search", handler.HandleSearch)
	mux.HandleFunc("GET /api/suggest", handler.HandleSuggest)

	// Order: Logging wraps everything so it always logs the final status
	// and total duration, even on a panic or a CORS-blocked request.
	// Recover sits inside Logging so a caught panic's 500 still gets
	// logged with the right status. CORS sits innermost, closest to the
	// actual handlers, but since it sets headers before calling the next
	// handler (rather than after), those headers are already attached to
	// the response even if something deeper panics - so error responses
	// remain readable by browser JS, not just success responses.
	wrapped := handlers.Chain(mux,
		handlers.Logging,
		handlers.Recover,
		handlers.CORS(cfg.CORSAllowedOrigins),
	)

	httpServer := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: wrapped,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      cfg.SearchTimeout + 5*time.Second,
		IdleTimeout:       60 * time.Second,
	}

	return &Server{httpServer: httpServer}
}

// Start begins serving and blocks until the server stops. A normal
// shutdown (via Shutdown) surfaces as a nil error, not
// http.ErrServerClosed.
func (s *Server) Start() error {
	err := s.httpServer.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

// Addr returns the address the server is configured to listen on.
func (s *Server) Addr() string {
	return s.httpServer.Addr
}

// Shutdown gracefully stops the server: it stops accepting new connections
// and waits for in-flight requests to finish, up to ctx's deadline.
func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}