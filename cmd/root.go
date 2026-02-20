package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "truelist",
	Short: "Truelist CLI — email validation from your terminal",
	Long: `Truelist CLI is the official command-line tool for Truelist.io email validation.

Validate single emails, bulk CSV files, or pipe from stdin.
Get started by setting your API key:

  truelist config set api-key YOUR_API_KEY`,
	SilenceUsage:  true,
	SilenceErrors: true,
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
