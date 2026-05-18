package cmd

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"

	"github.com/spf13/cobra"
	"go.sembraniteam.com/mpkg/internal/config"
	"go.sembraniteam.com/mpkg/internal/generator"
	"go.sembraniteam.com/mpkg/internal/server"
)

func newServeCmd(cfgFile *string) *cobra.Command {
	var (
		port        int
		openBrowser bool
	)

	c := &cobra.Command{
		Use:   "serve",
		Short: "Preview generated site with a local HTTP server",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load(*cfgFile)
			if err != nil {
				return err
			}
			if err := generator.Build(cfg); err != nil {
				return err
			}
			srv, err := server.New(cfg.Build.OutputDir, port)
			if err != nil {
				return fmt.Errorf("start server: %w", err)
			}
			url := "http://localhost:" + portFromAddr(srv.Addr())
			_, _ = fmt.Fprintf(
				cmd.OutOrStdout(),
				"Serving at %s\nPress Ctrl+C to stop.\n",
				url,
			)
			if openBrowser {
				openURL(url)
			}
			ctx := cmd.Context()
			if ctx == nil {
				ctx = context.Background()
			}
			srv.Start(ctx)
			return nil
		},
	}

	c.Flags().IntVarP(&port, "port", "p", defaultPort, "HTTP server port")
	c.Flags().
		BoolVar(&openBrowser, "open", false, "Open browser after server starts")
	return c
}

func portFromAddr(addr string) string {
	for i := len(addr) - 1; i >= 0; i-- {
		if addr[i] == ':' {
			return addr[i+1:]
		}
	}
	return addr
}

func openURL(url string) {
	c, args := openURLArgs(url)
	_ = exec.Command(c, args...).Start()
}

// openURLArgs returns the OS command and arguments needed to open a URL.
// Extracted for testing without spawning an OS process.
func openURLArgs(url string) (string, []string) {
	switch runtime.GOOS {
	case "darwin":
		return "open", []string{url}
	case "windows":
		return "cmd", []string{"/c", "start", url}
	default:
		return "xdg-open", []string{url}
	}
}
