package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

const (
	version     = "0.2.0"
	defaultPort = 8080
)

// NewRootCmd builds and returns the root Cobra command.
func NewRootCmd() *cobra.Command {
	var cfgFile string

	root := &cobra.Command{
		Use:   "mpkg",
		Short: "Generate static HTML for Go vanity import paths",
	}
	root.PersistentFlags().
		StringVarP(&cfgFile, "config", "c", "mpkg.yaml", "Path to config file")

	root.AddCommand(newVersionCmd())
	root.AddCommand(newValidateCmd(&cfgFile))
	root.AddCommand(newGenerateCmd(&cfgFile))
	root.AddCommand(newServeCmd(&cfgFile))
	root.AddCommand(newWatchCmd(&cfgFile))
	root.CompletionOptions.DisableDefaultCmd = true
	return root
}

// Execute runs the CLI.
func Execute() {
	if err := NewRootCmd().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print mpkg version",
		Run: func(cmd *cobra.Command, args []string) {
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), "mpkg version "+version)
		},
	}
}
