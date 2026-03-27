package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"os/exec"
	"strings"

	"github.com/mwunsch/mansplain/internal/ui"
	"github.com/spf13/cobra"
)

var installCmd = &cobra.Command{
	Use:   "install <file>",
	Short: "Install a man page so man(1) can find it",
	Long: `Copy a man page to your local man directory.

  mansplain install rg.1
  mansplain install --dry-run rg.1
  mansplain generate --name rg -o rg.1 && mansplain install rg.1

Installs to ~/.local/share/man/man<section>/ by default.
The section number is inferred from the file extension (.1 = section 1).`,
	Args: cobra.ExactArgs(1),
	RunE: runInstall,
}

var installDryRun bool

func init() {
	installCmd.Flags().BoolVar(&installDryRun, "dry-run", false, "show where the file would go")
	rootCmd.AddCommand(installCmd)
}

func runInstall(cmd *cobra.Command, args []string) error {
	src := args[0]

	// Infer section from extension
	ext := filepath.Ext(src)
	if ext == "" || len(ext) != 2 || ext[1] < '1' || ext[1] > '9' {
		return fmt.Errorf("cannot infer section from %q; expected extension like .1 through .9", src)
	}
	section := string(ext[1])
	base := filepath.Base(src)

	// Determine target directory
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("finding home directory: %w", err)
	}
	dir := filepath.Join(home, ".local", "share", "man", "man"+section)
	dest := filepath.Join(dir, base)

	if installDryRun {
		fmt.Printf("%s -> %s\n", src, dest)
		return nil
	}

	// Read source
	content, err := os.ReadFile(src)
	if err != nil {
		return fmt.Errorf("reading %s: %w", src, err)
	}

	// Create directory
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("creating %s: %w", dir, err)
	}

	// Write
	if err := os.WriteFile(dest, content, 0644); err != nil {
		return fmt.Errorf("writing %s: %w", dest, err)
	}

	// Update man database if available
	if makewhatis, err := exec.LookPath("makewhatis"); err == nil {
		mandir := filepath.Join(home, ".local", "share", "man")
		exec.Command(makewhatis, mandir).Run()
	} else if mandb, err := exec.LookPath("mandb"); err == nil {
		mandir := filepath.Join(home, ".local", "share", "man")
		exec.Command(mandb, "-q", mandir).Run()
	}

	ui.Success(fmt.Sprintf("Installed %s", dest))

	// Check if the man directory is in MANPATH
	manpath := os.Getenv("MANPATH")
	mandir := filepath.Join(home, ".local", "share", "man")
	if manpath != "" && !strings.Contains(manpath, mandir) {
		ui.Warning(fmt.Sprintf("%s is not in $MANPATH; add it to your shell profile", mandir))
	}

	return nil
}
