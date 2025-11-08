package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage configuration",
	Long:  "Get or set configuration values such as max-retries, backoff-base, job-timeout",
}

var configGetCmd = &cobra.Command{
	Use:   "get [key]",
	Short: "Get configuration value",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		key := args[0]
		value, err := db.GetConfig(key)
		if err != nil {
			return fmt.Errorf("failed to get config: %w", err)
		}
		if value == "" {
			return fmt.Errorf("config key not found: %s", key)
		}
		fmt.Printf("%s = %s\n", key, value)
		return nil
	},
}

var configSetCmd = &cobra.Command{
	Use:   "set [key] [value]",
	Short: "Set configuration value",
	Long:  "Update configuration values (max-retries, backoff-base, job-timeout)",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		key := args[0]
		value := args[1]
		validKeys := map[string]bool{
			"max-retries":  true,
			"backoff-base": true,
			"job-timeout":  true,
		}

		if !validKeys[key] {
			return fmt.Errorf("invalid config key. Valid keys: %s",
				strings.Join([]string{"max-retries", "backoff-base", "job-timeout"}, ", "))
		}

		if err := db.SetConfig(key, value); err != nil {
			return fmt.Errorf("failed to set config: %w", err)
		}
		fmt.Printf("✓ Config updated: %s = %s\n", key, value)
		return nil
	},
}

func init() {
	configCmd.AddCommand(configGetCmd)
	configCmd.AddCommand(configSetCmd)
	rootCmd.AddCommand(configCmd)
}
