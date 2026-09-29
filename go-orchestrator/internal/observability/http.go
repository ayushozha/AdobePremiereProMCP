package observability

import (
	"context"
	"errors"
	"net"
	"net/http"
	"time"

	"github.com/ayushozha/AdobePremiereProMCP/go-orchestrator/internal/health"
)

// Server owns a separate local HTTP listener for metrics and health.
type Server struct {
	listener net.Listener
	http     *http.Server
}

// NewServer binds before MCP starts, so configuration and port conflicts are
// visible startup errors instead of silently disabling observability.
func NewServer(address string, metrics *Metrics, checker *health.Checker) (*Server, error) {
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return nil, err
	}
	mux := http.NewServeMux()
	mux.Handle("GET /metrics", metrics.Handler())
	mux.Handle("/", health.NewHTTPHandler(checker))
	return &Server{
		listener: listener,
		http:     &http.Server{Handler: mux, ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 10 * time.Second, WriteTimeout: 10 * time.Second, IdleTimeout: 30 * time.Second},
	}, nil
}

// Serve runs until the context is canceled or the listener fails.
func (s *Server) Serve(ctx context.Context) error {
	stopped := make(chan error, 1)
	go func() { stopped <- s.http.Serve(s.listener) }()
	select {
	case err := <-stopped:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := s.http.Shutdown(shutdownCtx); err != nil {
			_ = s.http.Close()
			return err
		}
		err := <-stopped
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	}
}

// Close releases the listener even if Serve was never started.
func (s *Server) Close() error {
	_ = s.http.Close()
	return s.listener.Close()
}
