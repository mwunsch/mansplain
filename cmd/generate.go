package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/mwunsch/mansplain/internal/config"
	"github.com/mwunsch/mansplain/internal/llm"
	"github.com/mwunsch/mansplain/internal/ui"
	"github.com/spf13/cobra"
)

var generatePrompt string

// SetGeneratePrompt sets the system prompt for the generate command.
func SetGeneratePrompt(prompt string) {
	generatePrompt = prompt
}

var generateCmd = &cobra.Command{
	Use:   "generate [file | -]",
	Short: "Generate a man page from source material via LLM",
	Long: `Generate a well-formed mdoc(7) man page using an LLM.

Source material can come from files, --help output, or stdin:

  mansplain generate README.md --name mytool
  mansplain generate --from-help "rg --help" -o rg.1
  mansplain generate --name jq
  curl --help | mansplain generate - --name curl

Use --section for non-command man pages:

  mansplain generate config.toml --name myapp.conf --section 5
  mansplain generate ARCHITECTURE.md --name myframework --section 7

Use - to read from stdin. If only --name is given, runs <name> --help.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runGenerate,
}

var (
	fromHelp string
	toolName string
	section  int
	output   string
	dryRun   bool
)

func init() {
	generateCmd.Flags().StringVar(&fromHelp, "from-help", "", "run a command and capture its output")
	generateCmd.Flags().StringVar(&toolName, "name", "", "tool name (inferred if possible)")
	generateCmd.Flags().IntVar(&section, "section", 1, "man page section number")
	generateCmd.Flags().StringVarP(&output, "output", "o", "", "output file (default: stdout)")
	generateCmd.Flags().BoolVar(&dryRun, "dry-run", false, "print the prompt instead of calling the API")

	rootCmd.AddCommand(generateCmd)
}

func runGenerate(cmd *cobra.Command, args []string) error {
	var sources []llm.Source

	// Positional arg: file path or - for stdin
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

	// --from-help: run a command and capture output
	if fromHelp != "" {
		helpText, err := captureHelp(fromHelp)
		if err != nil {
			return fmt.Errorf("capturing help output: %w", err)
		}
		sources = append(sources, llm.Source{Type: "help", Content: helpText})

		if toolName == "" {
			toolName = inferToolName(fromHelp)
		}
	}

	// If we have a name but no sources, try running <name> --help
	if len(sources) == 0 && toolName != "" {
		helpText, err := captureHelp(toolName + " --help")
		if err != nil {
			return fmt.Errorf("no source material provided and %q --help failed: %w", toolName, err)
		}
		sources = append(sources, llm.Source{Type: "help", Content: helpText})
	}

	if len(sources) == 0 {
		return fmt.Errorf("no source material; provide a file, use --from-help, --name, or -")
	}

	// Infer tool name from source content if not given
	if toolName == "" {
		toolName = inferToolNameFromSources(sources)
	}

	if toolName == "" {
		return fmt.Errorf("could not infer tool name; use --name to specify it")
	}

	// Assemble prompt
	llm.SetSystemPrompt(generatePrompt)

	req := llm.GenerateRequest{
		Sources: sources,
		Name:    toolName,
		Section: section,
	}

	if dryRun {
		fmt.Println("=== SYSTEM PROMPT ===")
		fmt.Println(llm.SystemPrompt())
		fmt.Println()
		fmt.Println("=== USER PROMPT ===")
		fmt.Print(llm.BuildUserPrompt(req))
		return nil
	}

	cfg, err := config.Resolve(config.Flags{
		BaseURL: flagBaseURL,
		APIKey:  flagAPIKey,
		Model:   flagModel,
	})
	if err != nil {
		return err
	}

	ui.ProviderBanner(cfg.BaseURL, cfg.Model)

	client := llm.NewClient(llm.Config{
		APIURL: cfg.BaseURL,
		APIKey: cfg.APIKey,
		Model:  cfg.Model,
	})

	var result *llm.GenerateResult
	start := time.Now()

	err = ui.RunGeneration(
		fmt.Sprintf("Generating %s(%d)...", toolName, section),
		func(ctx context.Context) error {
			var genErr error
			result, genErr = client.Generate(ctx, req)
			return genErr
		},
	)

	elapsed := time.Since(start)

	if err != nil {
		return err
	}

	if output != "" {
		if err := os.WriteFile(output, []byte(result.Content), 0644); err != nil {
			return fmt.Errorf("writing output: %w", err)
		}
		ui.Success(fmt.Sprintf("Wrote %s", output))
	} else {
		fmt.Print(result.Content)
	}

	ui.TokenUsage(result.Usage.PromptTokens, result.Usage.CompletionTokens, result.Usage.TotalTokens, elapsed)

	return nil
}

func captureHelp(command string) (string, error) {
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return "", fmt.Errorf("empty command")
	}
	cmd := exec.Command(parts[0], parts[1:]...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		if len(out) > 0 {
			return string(out), nil
		}
		return "", fmt.Errorf("running %q: %w", command, err)
	}
	return string(out), nil
}

func inferToolNameFromSources(sources []llm.Source) string {
	for _, src := range sources {
		if name := inferToolNameFromText(src.Content); name != "" {
			return name
		}
	}
	return ""
}

func inferToolNameFromText(text string) string {
	lines := strings.Split(text, "\n")
	limit := 30
	if len(lines) < limit {
		limit = len(lines)
	}

	for i, line := range lines[:limit] {
		trimmed := strings.TrimSpace(line)

		for _, prefix := range []string{"Usage:", "usage:", "USAGE:"} {
			if strings.HasPrefix(trimmed, prefix) {
				rest := strings.TrimSpace(strings.TrimPrefix(trimmed, prefix))
				if rest != "" {
					fields := strings.Fields(rest)
					if len(fields) > 0 {
						return extractBaseName(fields[0])
					}
				}
				if i+1 < len(lines) {
					nextFields := strings.Fields(lines[i+1])
					if len(nextFields) > 0 {
						return extractBaseName(nextFields[0])
					}
				}
			}
		}

		if i < 3 && trimmed != "" {
			fields := strings.Fields(trimmed)
			if len(fields) >= 2 && (fields[1] == "-" || fields[1] == "--" || fields[1] == "[command]" || strings.HasPrefix(fields[1], "[")) {
				return extractBaseName(fields[0])
			}
		}
	}
	return ""
}

func extractBaseName(s string) string {
	if idx := strings.LastIndex(s, "/"); idx >= 0 {
		s = s[idx+1:]
	}
	s = strings.TrimRight(s, ":.")
	return s
}

func inferToolName(command string) string {
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return ""
	}
	name := parts[0]
	if idx := strings.LastIndex(name, "/"); idx >= 0 {
		name = name[idx+1:]
	}
	return name
}
