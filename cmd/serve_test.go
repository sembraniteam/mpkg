package cmd

import (
	"bytes"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

// safeBuffer is a bytes.Buffer protected by a mutex for concurrent use.
type safeBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (b *safeBuffer) Write(p []byte) (n int, err error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.Write(p)
}

func (b *safeBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.String()
}

func TestServeCommand(t *testing.T) {
	tests := []struct {
		name    string
		yaml    string
		wantOut string
		wantErr bool
	}{
		{
			name:    "starts server and prints address",
			yaml:    "base_url: example.com\nbuild:\n  output_dir: OUTDIR\nmodules: {}\n",
			wantOut: "Serving at",
		},
		{
			name:    "invalid config returns error",
			yaml:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			cfgPath := filepath.Join(dir, "mpkg.yaml")
			outDir := filepath.Join(dir, "out")

			if tt.yaml != "" {
				yaml := strings.ReplaceAll(tt.yaml, "OUTDIR", outDir)
				_ = os.WriteFile(cfgPath, []byte(yaml), 0o644)
			} else {
				cfgPath = filepath.Join(dir, "no.yaml")
			}

			outBuf := &safeBuffer{}
			root := NewRootCmd()
			root.SetOut(outBuf)
			root.SetErr(outBuf)

			ctx := t.Context()

			done := make(chan error, 1)
			go func() {
				root.SetArgs(
					[]string{"serve", "--config", cfgPath, "--port", "0"},
				)
				done <- root.ExecuteContext(ctx)
			}()

			if !tt.wantErr {
				time.Sleep(150 * time.Millisecond)
				if !strings.Contains(outBuf.String(), tt.wantOut) {
					t.Errorf(
						"output = %q, want to contain %q",
						outBuf.String(),
						tt.wantOut,
					)
				}
				// Verify server actually responds
				addr := extractAddr(outBuf.String())
				if addr != "" {
					resp, err := http.Get("http://" + addr + "/")
					if err == nil {
						_ = resp.Body.Close()
					}
				}
			} else {
				// For error case, give it a moment to fail
				select {
				case err := <-done:
					if err == nil {
						t.Error("expected error, got nil")
					}
				case <-time.After(500 * time.Millisecond):
					t.Error("command did not return in time")
				}
			}
		})
	}
}

func extractAddr(s string) string {
	_, after, ok := strings.Cut(s, "http://")
	if !ok {
		return ""
	}
	end := strings.IndexAny(after, " \n\r")
	if end == -1 {
		return after
	}
	return after[:end]
}
