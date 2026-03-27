package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var installCmd = &cobra.Command{
	Use:   "install [file]",
	Short: "Install a man page to the appropriate system location",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return fmt.Errorf("install not yet implemented")
	},
}

func init() {
	rootCmd.AddCommand(installCmd)
}
