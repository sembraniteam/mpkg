package server

import (
	"context"
	"fmt"
	"net"
	"net/http"
)

// Server serves a static directory over HTTP.
type Server struct {
	listener net.Listener
	dir      string
}

// New creates a Server bound to an available port.
// Pass port=0 to let the OS pick a free port.
func New(dir string, port int) (*Server, error) {
	addr := fmt.Sprintf(":%d", port)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("listen on %s: %w", addr, err)
	}
	return &Server{listener: ln, dir: dir}, nil
}

// Addr returns the bound address (e.g. "127.0.0.1:8080").
func (s *Server) Addr() string {
	return s.listener.Addr().String()
}

// Start serves requests until ctx is canceled.
func (s *Server) Start(ctx context.Context) {
	srv := &http.Server{Handler: http.FileServer(http.Dir(s.dir))}
	go func() {
		<-ctx.Done()
		_ = srv.Close()
	}()
	_ = srv.Serve(s.listener)
}
