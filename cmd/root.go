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

var rootCmd = &cobra.Command{
	Use:   "mansplain",
	Short: "generate man pages from --help output and READMEs",
	Long: `mansplain generates mdoc(7) man pages using an LLM.

Feed it --help output, a README, or just a tool name and it produces
idiomatic, well-structured man pages ready for man(1).

  mansplain generate --name rg
  mansplain generate --from-help "jq --help" -o jq.1
  curl --help | mansplain generate - --name curl
  mansplain generate README.md --name mytool`,
	SilenceUsage:  true,
	SilenceErrors: true,
}

var (
	flagBaseURL string
	flagAPIKey  string
	flagModel   string
)

func init() {
	rootCmd.PersistentFlags().StringVar(&flagBaseURL, "base-url", "", "OpenAI-compatible API base URL")
	rootCmd.PersistentFlags().StringVar(&flagAPIKey, "api-key", "", "API key")
	rootCmd.PersistentFlags().StringVar(&flagModel, "model", "", "LLM model to use")

	rootCmd.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Print the version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("mansplain %s\n", version)
		},
	})
}

func Execute() error {
	return rootCmd.Execute()
}
