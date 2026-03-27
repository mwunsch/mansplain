package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var scaffoldCmd = &cobra.Command{
	Use:   "scaffold",
	Short: "Generate an mdoc template offline (no LLM needed)",
	RunE: func(cmd *cobra.Command, args []string) error {
		return fmt.Errorf("scaffold not yet implemented")
	},
}

func init() {
	rootCmd.AddCommand(scaffoldCmd)
}
