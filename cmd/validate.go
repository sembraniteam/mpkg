package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"go.sembraniteam.com/mpkg/internal/config"
)

func newValidateCmd(cfgFile *string) *cobra.Command {
	return &cobra.Command{
		Use:   "validate",
		Short: "Validate mpkg.yaml without generating output",
		RunE: func(cmd *cobra.Command, args []string) error {
			_, err := config.Load(*cfgFile)
			if err != nil {
				return fmt.Errorf("✗ %w", err)
			}
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), "✓ Config is valid")
			return nil
		},
	}
}
