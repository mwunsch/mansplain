package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/mwunsch/mansplain/internal/ui"
	"github.com/spf13/cobra"
)

var lintCmd = &cobra.Command{
	Use:   "lint [file | -]",
	Short: "Validate man page structure and completeness",
	Long: `Check an mdoc(7) file for structural problems.

Runs mandoc -Tlint if available, then checks for required sections
(NAME, SYNOPSIS, DESCRIPTION) and recommended sections (OPTIONS,
EXAMPLES, EXIT STATUS, SEE ALSO).

  mansplain lint page.1
  mansplain generate --name rg | mansplain lint -`,
	Args: cobra.ExactArgs(1),
	RunE: runLint,
}

func init() {
	rootCmd.AddCommand(lintCmd)
}

var requiredSections = []string{"NAME", "SYNOPSIS", "DESCRIPTION"}
var recommendedSections = []string{"OPTIONS", "EXAMPLES", "EXIT STATUS", "SEE ALSO"}

func runLint(cmd *cobra.Command, args []string) error {
	// Read source
	var content []byte
	var err error
	var filename string

	if args[0] == "-" {
		content, err = os.ReadFile("/dev/stdin")
		filename = "<stdin>"
	} else {
		filename = args[0]
		content, err = os.ReadFile(filename)
	}
	if err != nil {
		return fmt.Errorf("reading %s: %w", filename, err)
	}

	hasErrors := false

	// Run mandoc -Tlint if available
	if mandocPath, err := exec.LookPath("mandoc"); err == nil {
		mandoc := exec.Command(mandocPath, "-Tlint")
		mandoc.Stdin = strings.NewReader(string(content))
		out, _ := mandoc.CombinedOutput()
		if len(out) > 0 {
			// Filter out mandocdb warnings (system-level, not our problem)
			scanner := bufio.NewScanner(strings.NewReader(string(out)))
			for scanner.Scan() {
				line := scanner.Text()
				if strings.Contains(line, "mandoc_mandocdb") || strings.Contains(line, "mandoc.db") {
					continue
				}
				if strings.Contains(line, "referenced manual not found") {
					continue
				}
				if strings.Contains(line, "ERROR") {
					ui.Error(line)
					hasErrors = true
				} else if strings.Contains(line, "WARNING") {
					ui.Warning(line)
				} else if strings.Contains(line, "STYLE") {
					ui.Info("style", line)
				} else {
					ui.Info("mandoc", line)
				}
			}
		}
		// Don't treat mandoc's exit code as definitive —
		// it exits non-zero for warnings too. We track errors
		// from the output parsing above.
	}

	// Infer section from .Dt line for section-aware checks
	pageSection := inferPageSection(string(content))

	// Check for required and recommended sections
	sections := findSections(string(content))

	required := requiredSections
	recommended := recommendedSections
	if pageSection == 7 {
		// Section 7 pages don't need SYNOPSIS
		required = []string{"NAME", "DESCRIPTION"}
		recommended = []string{"EXAMPLES", "SEE ALSO"}
	} else if pageSection == 5 {
		// Section 5 pages: SYNOPSIS shows a file path, not flags
		recommended = []string{"EXAMPLES", "SEE ALSO"}
	}

	for _, req := range required {
		if !sections[req] {
			ui.Error(fmt.Sprintf("missing required section: %s", req))
			hasErrors = true
		}
	}

	for _, rec := range recommended {
		if !sections[rec] {
			ui.Warning(fmt.Sprintf("missing recommended section: %s", rec))
		}
	}

	// Check for .Dd header
	if !strings.Contains(string(content), ".Dd") {
		ui.Error("missing .Dd (document date)")
		hasErrors = true
	}

	// Check for .Dt header
	if !strings.Contains(string(content), ".Dt") {
		ui.Error("missing .Dt (document title)")
		hasErrors = true
	}

	// Check for raw troff macros that should be mdoc instead
	rawTroff := []string{".ft", ".sp", ".in", ".br", ".nf", ".fi", ".RS", ".RE"}
	for _, macro := range rawTroff {
		if containsMacro(string(content), macro) {
			ui.Warning(fmt.Sprintf("raw troff macro %s found; use mdoc equivalents", macro))
		}
	}

	if hasErrors {
		return fmt.Errorf("lint failed")
	}

	ui.Success(fmt.Sprintf("%s: ok", filename))
	return nil
}

// inferPageSection extracts the man page section number from the .Dt line.
func inferPageSection(content string) int {
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, ".Dt ") {
			fields := strings.Fields(line)
			if len(fields) >= 3 {
				if s := fields[2]; len(s) == 1 && s[0] >= '1' && s[0] <= '9' {
					return int(s[0] - '0')
				}
			}
		}
	}
	return 1 // default to section 1
}

// findSections scans for .Sh lines and returns a set of section names.
func findSections(content string) map[string]bool {
	sections := make(map[string]bool)
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, ".Sh ") {
			name := strings.TrimPrefix(line, ".Sh ")
			// Section names may be quoted
			name = strings.Trim(name, "\"")
			sections[name] = true
		}
	}
	return sections
}

// containsMacro checks if a macro appears at the start of a line.
func containsMacro(content, macro string) bool {
	for _, line := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == macro || strings.HasPrefix(trimmed, macro+" ") {
			return true
		}
	}
	return false
}
