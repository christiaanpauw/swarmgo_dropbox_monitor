package web

import (
	"context"
	"net/http"

	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/container"
	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/lifecycle"
)

// Server represents the web server
type Server struct {
	*lifecycle.BaseComponent
	container *container.Container
	server    *http.Server
}

// NewServer creates a new web server
func NewServer(c *container.Container) *Server {
	return &Server{
		BaseComponent: lifecycle.NewBaseComponent("WebServer"),
		container:    c,
		server:      &http.Server{Addr: ":8080"},
	}
}

// Start starts the web server
func (s *Server) Start(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	// Start container first
	if err := s.container.Start(ctx); err != nil {
		return err
	}

	// Set up routes
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleIndex)
	mux.HandleFunc("/health", s.handleHealth)
	s.server.Handler = mux

	// Start server
	go func() {
		if err := s.server.ListenAndServe(); err != http.ErrServerClosed {
			s.SetState(lifecycle.StateFailed)
		}
	}()

	s.SetState(lifecycle.StateRunning)
	return nil
}

// Stop stops the web server
func (s *Server) Stop(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	// Shutdown server
	if err := s.server.Shutdown(ctx); err != nil {
		return err
	}

	// Stop container
	if err := s.container.Stop(ctx); err != nil {
		return err
	}

	s.SetState(lifecycle.StateStopped)
	return nil
}

// Health checks the health of the web server
func (s *Server) Health(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	if s.State() == lifecycle.StateFailed {
		return lifecycle.ErrNotRunning
	}

	if s.State() == lifecycle.StateStopped {
		return lifecycle.ErrNotRunning
	}

	return s.container.Health(ctx)
}

// handleIndex handles the index page
func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Welcome to Dropbox Monitor"))
}

// handleHealth handles the health check endpoint
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	if err := s.Health(r.Context()); err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte(err.Error()))
		return
	}
	w.Write([]byte("OK"))
}
