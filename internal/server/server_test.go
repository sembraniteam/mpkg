package server

import (
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewServerError(t *testing.T) {
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
		{
			name: "serves custom 404.html for missing file",
			setup: func(dir string) {
				_ = os.WriteFile(
					filepath.Join(dir, "404.html"),
					[]byte("<h1>not found</h1>"),
					0o644,
				)
			},
			path:       "/missing",
			wantStatus: http.StatusNotFound,
			wantBody:   "<h1>not found</h1>",
		},
		{
			name: "returns 404 for directory without index.html",
			setup: func(dir string) {
				_ = os.MkdirAll(filepath.Join(dir, "somedir"), 0o755)
				_ = os.WriteFile(
					filepath.Join(dir, "404.html"),
					[]byte("<h1>not found</h1>"),
					0o644,
				)
			},
			path:       "/somedir",
			wantStatus: http.StatusNotFound,
			wantBody:   "<h1>not found</h1>",
		},
		{
			name: "path traversal returns 404",
			setup: func(dir string) {
				_ = os.WriteFile(
					filepath.Join(dir, "404.html"),
					[]byte("<h1>not found</h1>"),
					0o644,
				)
			},
			path:       "/../../etc/passwd",
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			tt.setup(dir)

			srv, err := New(dir, 0)
			if err != nil {
				t.Fatalf("New() error = %v", err)
			}

			go srv.Start(t.Context())

			resp, err := http.Get("http://" + srv.Addr() + tt.path)
			if err != nil {
				t.Fatalf("GET %s: %v", tt.path, err)
			}
			defer func() { _ = resp.Body.Close() }()

			if resp.StatusCode != tt.wantStatus {
				t.Errorf("status = %d, want %d", resp.StatusCode, tt.wantStatus)
			}

			if tt.wantBody != "" {
				body, _ := io.ReadAll(resp.Body)
				if !strings.Contains(string(body), tt.wantBody) {
					t.Errorf(
						"body does not contain %q, got: %s",
						tt.wantBody,
						body,
					)
				}
			}
		})
	}
}
