package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var previewCmd = &cobra.Command{
	Use:   "preview [file]",
	Short: "Render a man page to the terminal",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return fmt.Errorf("preview not yet implemented")
	},
}

func init() {
	rootCmd.AddCommand(previewCmd)
}
