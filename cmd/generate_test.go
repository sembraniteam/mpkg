package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerateCommand(t *testing.T) {
	validYAML := func(outDir string) string {
		return "base_url: example.com\nbuild:\n  output_dir: " + outDir + "\nmodules:\n  r:\n    git_url: https://g.com/r\n"
	}

	tests := []struct {
		name      string
		setupYAML func(dir, outDir string) string
		extraArgs []string
		wantFiles []string
		wantOut   string
		wantErr   bool
	}{
		{
			name:      "generates files and prints summary",
			setupYAML: func(dir, outDir string) string { return validYAML(outDir) },
			wantFiles: []string{"index.html", "r/index.html", "404.html"},
			wantOut:   "Generated",
		},
		{
			name:      "--clean removes output dir first",
			setupYAML: func(dir, outDir string) string { return validYAML(outDir) },
			extraArgs: []string{"--clean"},
			wantFiles: []string{"index.html"},
			wantOut:   "Generated",
		},
		{
			name:      "missing config returns error",
			setupYAML: nil,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			outDir := filepath.Join(dir, "out")
			cfgPath := filepath.Join(dir, "mpkg.yaml")

			if tt.setupYAML != nil {
				yaml := tt.setupYAML(dir, outDir)
				_ = os.WriteFile(cfgPath, []byte(yaml), 0o644)
			} else {
				cfgPath = filepath.Join(dir, "nonexistent.yaml")
			}

			outBuf := &bytes.Buffer{}
			root := NewRootCmd()
			root.SetOut(outBuf)
			root.SetErr(outBuf)
			args := append(
				[]string{"generate", "--config", cfgPath},
				tt.extraArgs...)
			root.SetArgs(args)

			err := root.Execute()
			if (err != nil) != tt.wantErr {
				t.Fatalf("Execute() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}

			out := outBuf.String()
			if tt.wantOut != "" && !strings.Contains(out, tt.wantOut) {
				t.Errorf("output = %q, want to contain %q", out, tt.wantOut)
			}
			for _, f := range tt.wantFiles {
				if _, err := os.Stat(filepath.Join(outDir, f)); err != nil {
					t.Errorf("expected file %q: %v", f, err)
				}
			}
		})
	}
}
