package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidateCommand(t *testing.T) {
	tests := []struct {
		name    string
		yaml    string
		wantOut string
		wantErr bool
	}{
		{
			name:    "valid config prints ok",
			yaml:    "base_url: example.com\nmodules:\n  r:\n    git_url: https://g.com/r\n",
			wantOut: "Config is valid",
		},
		{
			name:    "missing base_url prints error",
			yaml:    "modules:\n  r:\n    git_url: https://g.com/r\n",
			wantOut: "base_url",
			wantErr: true,
		},
		{
			name:    "missing git_url prints module name",
			yaml:    "base_url: example.com\nmodules:\n  router:\n    description: test\n",
			wantOut: "router",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			cfgPath := filepath.Join(dir, "mpkg.yaml")
			_ = os.WriteFile(cfgPath, []byte(tt.yaml), 0o644)

			outBuf := &bytes.Buffer{}
			root := NewRootCmd()
			root.SetOut(outBuf)
			root.SetErr(outBuf)
			root.SetArgs([]string{"validate", "--config", cfgPath})

			err := root.Execute()
			out := outBuf.String()

			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !strings.Contains(out, tt.wantOut) {
				t.Errorf("output = %q, want to contain %q", out, tt.wantOut)
			}
		})
	}
}
