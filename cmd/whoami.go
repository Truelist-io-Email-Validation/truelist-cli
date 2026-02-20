package cmd

import (
	"context"
	"os"

	"github.com/Truelist-io-Email-Validation/truelist-cli/internal/client"
	"github.com/Truelist-io-Email-Validation/truelist-cli/internal/config"
	"github.com/Truelist-io-Email-Validation/truelist-cli/internal/output"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(whoamiCmd)
}

var whoamiCmd = &cobra.Command{
	Use:   "whoami",
	Short: "Display current account information",
	Long:  "Check your API key and display account details including email, plan, and remaining credits.",
	RunE: func(cmd *cobra.Command, args []string) error {
		apiKey, err := config.GetAPIKey()
		if err != nil {
			output.PrintError(os.Stderr, err.Error())
			return err
		}

		c := client.New(apiKey)
		info, err := c.Whoami(context.Background())
		if err != nil {
			output.PrintError(os.Stderr, err.Error())
			return err
		}

		output.PrintAccountInfo(os.Stdout, info)
		return nil
	},
}
