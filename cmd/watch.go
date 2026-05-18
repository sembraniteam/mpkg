package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"go.sembraniteam.com/mpkg/internal/config"
	"go.sembraniteam.com/mpkg/internal/generator"
	"go.sembraniteam.com/mpkg/internal/server"
	"go.sembraniteam.com/mpkg/internal/watcher"
)

func newWatchCmd(cfgFile *string) *cobra.Command {
	var (
		port        int
		openBrowser bool
	)

	c := &cobra.Command{
		Use:   "watch",
		Short: "Serve and auto-regenerate when mpkg.yaml changes",
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
				"Watching %s for changes...\nServing at %s\nPress Ctrl+C to stop.\n",
				*cfgFile,
				url,
			)

			if openBrowser {
				openURL(url)
			}

			ctx := cmd.Context()
			go srv.Start(ctx)

			return watcher.Watch(ctx, *cfgFile, func() {
				newCfg, err := config.Load(*cfgFile)
				if err != nil {
					_, _ = fmt.Fprintf(
						cmd.ErrOrStderr(),
						"reload error: %v\n",
						err,
					)
					return
				}
				if err := generator.Build(newCfg); err != nil {
					_, _ = fmt.Fprintf(
						cmd.ErrOrStderr(),
						"rebuild error: %v\n",
						err,
					)
					return
				}
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), "↺ Rebuilt")
			})
		},
	}

	c.Flags().IntVarP(&port, "port", "p", defaultPort, "HTTP server port")
	c.Flags().
		BoolVar(&openBrowser, "open", false, "Open browser after server starts")
	return c
}
