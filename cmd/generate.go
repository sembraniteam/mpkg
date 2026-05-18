package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"go.sembraniteam.com/mpkg/internal/config"
	"go.sembraniteam.com/mpkg/internal/generator"
)

func newGenerateCmd(cfgFile *string) *cobra.Command {
	var (
		outputDir string
		clean     bool
		quiet     bool
	)

	c := &cobra.Command{
		Use:   "generate",
		Short: "Build static HTML into the output directory",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load(*cfgFile)
			if err != nil {
				return err
			}
			if outputDir != "" {
				cfg.Build.OutputDir = outputDir
			}
			if clean {
				if err := os.RemoveAll(cfg.Build.OutputDir); err != nil {
					return fmt.Errorf("clean output dir: %w", err)
				}
			}
			start := time.Now()
			if err := generator.Build(cfg); err != nil {
				return err
			}
			if !quiet {
				_, _ = fmt.Fprintf(
					cmd.OutOrStdout(),
					"✓ Generated %d package(s) → %s (%s)\n",
					len(
						cfg.Modules,
					),
					cfg.Build.OutputDir,
					time.Since(start).Round(time.Millisecond),
				)
			}
			return nil
		},
	}

	c.Flags().
		StringVarP(&outputDir, "output", "o", "", "Override output directory")
	c.Flags().
		BoolVar(&clean, "clean", false, "Remove output directory before generating")
	c.Flags().
		BoolVarP(&quiet, "quiet", "q", false, "Suppress output except errors")
	return c
}
