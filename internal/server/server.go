package server

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"os"
	"path"
	"path/filepath"
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
	srv := &http.Server{
		Handler: with404(s.dir, http.FileServer(http.Dir(s.dir))),
	}
	go func() {
		<-ctx.Done()
		_ = srv.Close()
	}()
	_ = srv.Serve(s.listener)
}

// with404 wraps next and serves 404.html from dir when the requested path
// does not exist, matching the behavior of static hosts like GitHub Pages.
// Falls back to http.NotFound if 404.html is missing from the build directory.
func with404(dir string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqPath := filepath.Join(
			dir,
			filepath.FromSlash(path.Clean("/"+r.URL.Path)),
		)

		info, err := os.Stat(reqPath)
		if err == nil && info.IsDir() {
			_, err = os.Stat(filepath.Join(reqPath, "index.html"))
		}

		if errors.Is(err, fs.ErrNotExist) {
			page, readErr := os.ReadFile(filepath.Join(dir, "404.html"))
			if readErr == nil {
				w.Header().Set("Content-Type", "text/html; charset=utf-8")
				w.WriteHeader(http.StatusNotFound)
				_, _ = w.Write(page)
				return
			}
			http.NotFound(w, r)
			return
		}

		next.ServeHTTP(w, r)
	})
}
