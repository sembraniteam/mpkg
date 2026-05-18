package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func TestVersionCommand(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantOut string
	}{
		{
			name:    "version prints version string",
			args:    []string{"version"},
			wantOut: "mpkg version",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			root := NewRootCmd()
			root.SetOut(buf)
			root.SetArgs(tt.args)
			if err := root.Execute(); err != nil {
				t.Fatalf("Execute() error = %v", err)
			}
			if !strings.Contains(buf.String(), tt.wantOut) {
				t.Errorf(
					"output = %q, want to contain %q",
					buf.String(),
					tt.wantOut,
				)
			}
		})
	}
}
