package server

import (
	"net"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewServerError(t *testing.T) {
	// Bind a port, then try to bind the same port again — should fail.
	ln, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = ln.Close() }()
	port := ln.Addr().(*net.TCPAddr).Port

	_, err = New(t.TempDir(), port)
	if err == nil {
		t.Error("expected error when port is already in use, got nil")
	}
}

func TestServe(t *testing.T) {
	tests := []struct {
		name       string
		setup      func(dir string)
		path       string
		wantStatus int
		wantBody   string
	}{
		{
			name: "serves index.html",
			setup: func(dir string) {
				_ = os.WriteFile(
					filepath.Join(dir, "index.html"),
					[]byte("<h1>hello</h1>"),
					0o644,
				)
			},
			path:       "/",
			wantStatus: http.StatusOK,
			wantBody:   "<h1>hello</h1>",
		},
		{
			name:       "returns 404 for missing file",
			setup:      func(dir string) {},
			path:       "/missing",
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			tt.setup(dir)

			srv, err := New(dir, 0) // port 0 = random available port
			if err != nil {
				t.Fatalf("New() error = %v", err)
			}

			go srv.Start(t.Context())
			time.Sleep(50 * time.Millisecond)

			resp, err := http.Get("http://" + srv.Addr() + tt.path)
			if err != nil {
				t.Fatalf("GET %s: %v", tt.path, err)
			}
			defer func() { _ = resp.Body.Close() }()

			if resp.StatusCode != tt.wantStatus {
				t.Errorf("status = %d, want %d", resp.StatusCode, tt.wantStatus)
			}
		})
	}
}
