package cmd

import (
	"fmt"
	"os"

	"github.com/Truelist-io-Email-Validation/truelist-cli/internal/config"
	"github.com/Truelist-io-Email-Validation/truelist-cli/internal/output"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configSetCmd)
}

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage CLI configuration",
}

var configSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set a configuration value",
	Long: `Set a configuration value. Supported keys:

  api-key    Your Truelist API key

Example:
  truelist config set api-key tk_live_abc123`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		key, value := args[0], args[1]

		switch key {
		case "api-key":
			cfg, err := config.Load()
			if err != nil {
				output.PrintError(os.Stderr, err.Error())
				return err
			}
			cfg.APIKey = value
			if err := config.Save(cfg); err != nil {
				output.PrintError(os.Stderr, err.Error())
				return err
			}

			fp, _ := config.FilePath()
			fmt.Printf("API key saved to %s\n", fp)
			return nil

		default:
			err := fmt.Errorf("unknown config key: %s (supported: api-key)", key)
			output.PrintError(os.Stderr, err.Error())
			return err
		}
	},
}
