package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/mwunsch/mansplain/internal/convert"
	"github.com/mwunsch/mansplain/internal/ui"
	"github.com/spf13/cobra"
)

var convertCmd = &cobra.Command{
	Use:   "convert [file | -]",
	Short: "Convert ronn-format markdown to mdoc",
	Long: `Convert a ronn-format(7) markdown file to mdoc(7) source.

No LLM required. The conversion is deterministic.

  mansplain convert man/tool.1.md -o man/tool.1
  mansplain convert - < man/tool.1.md
  mansplain convert man/tool.1.md | mansplain lint -

The input format follows ronn-format(7) conventions:
  # name(section) -- description    title line
  ## SECTION NAME                   section headings
  ` + "`" + `--flag` + "`" + ` <arg>                        flags and arguments
  * ` + "`" + `--option` + "`" + ` <value>:              definition lists
    Description text`,
	Args: cobra.ExactArgs(1),
	RunE: runConvert,
}

var convertOutput string

func init() {
	convertCmd.Flags().StringVarP(&convertOutput, "output", "o", "", "output file (default: stdout)")
	rootCmd.AddCommand(convertCmd)
}

func runConvert(cmd *cobra.Command, args []string) error {
	var input []byte
	var err error

	if args[0] == "-" {
		input, err = io.ReadAll(os.Stdin)
		if err != nil {
			return fmt.Errorf("reading stdin: %w", err)
		}
	} else {
		input, err = os.ReadFile(args[0])
		if err != nil {
			return fmt.Errorf("reading %s: %w", args[0], err)
		}
	}

	result, err := convert.Convert(input)
	if err != nil {
		return err
	}

	if convertOutput != "" {
		if err := os.WriteFile(convertOutput, result, 0644); err != nil {
			return fmt.Errorf("writing output: %w", err)
		}
		ui.Success(fmt.Sprintf("Wrote %s", convertOutput))
		return nil
	}

	os.Stdout.Write(result)
	return nil
}
