package cmd

import (
	"testing"
)

func TestPortFromAddr(t *testing.T) {
	tests := []struct {
		addr string
		want string
	}{
		{addr: "127.0.0.1:8080", want: "8080"},
		{addr: "[::]:0", want: "0"},
		{addr: "nocolon", want: "nocolon"},
	}
	for _, tt := range tests {
		got := portFromAddr(tt.addr)
		if got != tt.want {
			t.Errorf("portFromAddr(%q) = %q, want %q", tt.addr, got, tt.want)
		}
	}
}

func TestOpenURLArgs(t *testing.T) {
	// Test command selection logic without spawning an OS process.
	cmd, args := openURLArgs("http://example.com")
	if cmd == "" {
		t.Error("expected non-empty command")
	}
	if len(args) == 0 {
		t.Error("expected at least one argument")
	}
	if args[len(args)-1] != "http://example.com" {
		t.Errorf("expected URL as last arg, got %v", args)
	}
}
