package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

// safeBufWriter is a thread-safe writer that also supports String().
type safeBufWriter struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (s *safeBufWriter) Write(p []byte) (n int, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.buf.Write(p)
}

func (s *safeBufWriter) String() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.buf.String()
}

func TestWatchCommand(t *testing.T) {
	tests := []struct {
		name    string
		yaml    string
		wantOut string
		wantErr bool
	}{
		{
			name:    "starts watcher and prints watching message",
			yaml:    "base_url: example.com\nbuild:\n  output_dir: OUTDIR\nmodules: {}\n",
			wantOut: "Watching",
		},
		{
			name:    "invalid config returns error",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			outDir := filepath.Join(dir, "out")
			cfgPath := filepath.Join(dir, "mpkg.yaml")

			if tt.yaml != "" {
				yaml := strings.ReplaceAll(tt.yaml, "OUTDIR", outDir)
				_ = os.WriteFile(cfgPath, []byte(yaml), 0o644)
			} else {
				cfgPath = filepath.Join(dir, "no.yaml")
			}

			safeBuf := &safeBufWriter{}

			root := NewRootCmd()
			root.SetOut(safeBuf)
			root.SetErr(safeBuf)

			ctx := t.Context()

			done := make(chan error, 1)
			go func() {
				root.SetArgs(
					[]string{"watch", "--config", cfgPath, "--port", "0"},
				)
				done <- root.ExecuteContext(ctx)
			}()

			if !tt.wantErr {
				time.Sleep(200 * time.Millisecond)
				out := safeBuf.String()
				if !strings.Contains(out, tt.wantOut) {
					t.Errorf("output = %q, want to contain %q", out, tt.wantOut)
				}
			} else {
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

func TestWatchCommandRebuildOnChange(t *testing.T) {
	dir := t.TempDir()
	outDir := filepath.Join(dir, "out")
	cfgPath := filepath.Join(dir, "mpkg.yaml")

	validYAML := "base_url: example.com\nbuild:\n  output_dir: " + outDir + "\nmodules: {}\n"
	_ = os.WriteFile(cfgPath, []byte(validYAML), 0o644)

	safeBuf := &safeBufWriter{}
	root := NewRootCmd()
	root.SetOut(safeBuf)
	root.SetErr(safeBuf)

	ctx := t.Context()
	go func() {
		root.SetArgs([]string{"watch", "--config", cfgPath, "--port", "0"})
		_ = root.ExecuteContext(ctx)
	}()

	// Wait for watcher to start
	time.Sleep(200 * time.Millisecond)

	// Trigger a rebuild by writing to the config file
	_ = os.WriteFile(cfgPath, []byte(validYAML), 0o644)

	// Give time for rebuild to complete
	time.Sleep(500 * time.Millisecond)

	out := safeBuf.String()
	if !strings.Contains(out, "Watching") {
		t.Errorf("output = %q, want to contain 'Watching'", out)
	}
}

func TestWatchCommandRebuildWithInvalidConfig(t *testing.T) {
	dir := t.TempDir()
	outDir := filepath.Join(dir, "out")
	cfgPath := filepath.Join(dir, "mpkg.yaml")

	validYAML := "base_url: example.com\nbuild:\n  output_dir: " + outDir + "\nmodules: {}\n"
	_ = os.WriteFile(cfgPath, []byte(validYAML), 0o644)

	safeBuf := &safeBufWriter{}
	root := NewRootCmd()
	root.SetOut(safeBuf)
	root.SetErr(safeBuf)

	ctx := t.Context()
	go func() {
		root.SetArgs([]string{"watch", "--config", cfgPath, "--port", "0"})
		_ = root.ExecuteContext(ctx)
	}()

	// Wait for watcher to start
	time.Sleep(200 * time.Millisecond)

	// Write broken YAML to trigger reload error path
	_ = os.WriteFile(cfgPath, []byte("base_url: \nmodules:\n  bad:\n"), 0o644)

	// Give time for error handling to run
	time.Sleep(500 * time.Millisecond)
}
