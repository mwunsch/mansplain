package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var version = "dev"

// SetVersion sets the version string (injected from main via ldflags).
func SetVersion(v string) {
	version = v
}

var flagVersion bool

var rootCmd = &cobra.Command{
	Use:   "mansplain",
	Short: "generate mdoc(7) man pages from source material",
	Long: `mansplain generates mdoc(7) man pages from source material.

Generate man pages with an LLM, or convert ronn-format markdown
to mdoc deterministically without one.

  mansplain generate --name rg
  mansplain generate README.md --name mytool
  mansplain generate config.toml --name myapp.conf --section 5
  mansplain convert man/tool.1.md -o man/tool.1`,
	SilenceUsage:  true,
	SilenceErrors: true,
	Run: func(cmd *cobra.Command, args []string) {
		if flagVersion {
			fmt.Printf("mansplain %s\n", version)
			return
		}
		cmd.Help()
	},
}

var (
	flagBaseURL string
	flagAPIKey  string
	flagModel   string
)

func init() {
	rootCmd.Flags().BoolVarP(&flagVersion, "version", "v", false, "print version")
	rootCmd.PersistentFlags().StringVar(&flagBaseURL, "base-url", "", "OpenAI-compatible API base URL")
	rootCmd.PersistentFlags().StringVar(&flagAPIKey, "api-key", "", "API key")
	rootCmd.PersistentFlags().StringVar(&flagModel, "model", "", "LLM model to use")
}

func Execute() error {
	return rootCmd.Execute()
}
