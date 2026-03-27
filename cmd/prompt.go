package cmd

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/mwunsch/mansplain/internal/llm"
	"github.com/spf13/cobra"
)

var promptCmd = &cobra.Command{
	Use:   "prompt [file | -]",
	Short: "Emit the LLM prompt without calling an API",
	Long: `Assemble and print the full prompt that would be sent to the LLM,
without actually making an API call. Useful for agents that want to
generate man pages using their own model context.

  mansplain prompt --name rg
  mansplain prompt --system-only
  mansplain prompt README.md --name mytool
  mansplain prompt --from-help "jq --help"`,
	Args: cobra.MaximumNArgs(1),
	RunE: runPrompt,
}

var (
	promptFromHelp   string
	promptToolName   string
	promptSection    int
	promptSystemOnly bool
)

func init() {
	promptCmd.Flags().StringVar(&promptFromHelp, "from-help", "", "run a command and capture its output")
	promptCmd.Flags().StringVar(&promptToolName, "name", "", "tool name")
	promptCmd.Flags().IntVar(&promptSection, "section", 1, "man page section number")
	promptCmd.Flags().BoolVar(&promptSystemOnly, "system-only", false, "emit only the system prompt")

	rootCmd.AddCommand(promptCmd)
}

func runPrompt(cmd *cobra.Command, args []string) error {
	llm.SetSystemPrompt(generatePrompt)

	if promptSystemOnly {
		fmt.Print(llm.SystemPrompt())
		return nil
	}

	var sources []llm.Source

	// Positional arg: file or -
	if len(args) == 1 {
		if args[0] == "-" {
			content, err := io.ReadAll(os.Stdin)
			if err != nil {
				return fmt.Errorf("reading stdin: %w", err)
			}
			sources = append(sources, llm.Source{Type: "stdin", Content: string(content)})
		} else {
			content, err := os.ReadFile(args[0])
			if err != nil {
				return fmt.Errorf("reading %s: %w", args[0], err)
			}
			sources = append(sources, llm.Source{Type: "readme", Content: string(content)})
		}
	}

	if promptFromHelp != "" {
		helpText, err := captureHelp(promptFromHelp)
		if err != nil {
			return fmt.Errorf("capturing help output: %w", err)
		}
		sources = append(sources, llm.Source{Type: "help", Content: helpText})

		if promptToolName == "" {
			promptToolName = inferToolName(promptFromHelp)
		}
	}

	// If name given but no sources, try --help
	if len(sources) == 0 && promptToolName != "" {
		helpText, err := captureHelp(promptToolName + " --help")
		if err != nil {
			return fmt.Errorf("no source material provided and %q --help failed: %w", promptToolName, err)
		}
		sources = append(sources, llm.Source{Type: "help", Content: helpText})
	}

	if len(sources) == 0 {
		return fmt.Errorf("no source material; provide a file, use --from-help, --name, or -")
	}

	if promptToolName == "" {
		promptToolName = inferToolNameFromSources(sources)
	}
	if promptToolName == "" {
		return fmt.Errorf("could not infer tool name; use --name to specify it")
	}

	// Print system prompt, then user prompt
	fmt.Println("=== SYSTEM PROMPT ===")
	fmt.Println(llm.SystemPrompt())
	fmt.Println()
	fmt.Println("=== USER PROMPT ===")
	fmt.Print(llm.BuildUserPrompt(llm.GenerateRequest{
		Sources: sources,
		Name:    promptToolName,
		Section: promptSection,
	}))

	// Also print as a hint for how many tokens this is
	systemLen := len(strings.Fields(llm.SystemPrompt()))
	userLen := len(strings.Fields(llm.BuildUserPrompt(llm.GenerateRequest{
		Sources: sources,
		Name:    promptToolName,
		Section: promptSection,
	})))
	fmt.Fprintf(os.Stderr, "\n~%d words (system: ~%d, user: ~%d)\n", systemLen+userLen, systemLen, userLen)

	return nil
}
