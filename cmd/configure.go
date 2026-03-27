package cmd

import (
	"fmt"

	"github.com/mwunsch/mansplain/internal/config"
	"github.com/mwunsch/mansplain/internal/ui"
	"github.com/spf13/cobra"
)

var configureCmd = &cobra.Command{
	Use:   "configure",
	Short: "Set up the LLM connection",
	RunE:  runConfigure,
}

func init() {
	rootCmd.AddCommand(configureCmd)
}

func runConfigure(cmd *cobra.Command, args []string) error {
	existing, err := config.Load()
	if err != nil {
		return err
	}

	result, err := ui.RunConfigureForm(existing)
	if err != nil {
		return err
	}
	if result == nil {
		return nil // user cancelled
	}

	if err := config.Save(result); err != nil {
		return err
	}

	ui.Success(fmt.Sprintf("Saved to %s", config.ConfigPath()))
	return nil
}
