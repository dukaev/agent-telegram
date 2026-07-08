package sys

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"

	"github.com/spf13/cobra"

	"agent-telegram/internal/cliutil"
	"agent-telegram/internal/docs"
)

var docsTarget string

// DocsCmd groups documentation generation helpers.
var DocsCmd = &cobra.Command{
	GroupID: "server",
	Use:     "docs",
	Short:   "Generate and check documentation from code",
	Long: `Generate and check documentation from the Cobra command tree,
registered IPC methods, operation schemas, and global flags.`,
}

// DocsGenerateCmd updates generated documentation blocks in README.md.
var DocsGenerateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Update generated documentation blocks",
	Run:   runDocsGenerate,
}

// DocsCheckCmd verifies generated documentation blocks are current.
var DocsCheckCmd = &cobra.Command{
	Use:   "check",
	Short: "Check generated documentation blocks",
	Run:   runDocsCheck,
}

// AddDocsCommand adds docs generation commands.
func AddDocsCommand(rootCmd *cobra.Command) {
	rootCmd.AddCommand(DocsCmd)
	DocsCmd.AddCommand(DocsGenerateCmd, DocsCheckCmd)
	for _, cmd := range []*cobra.Command{DocsGenerateCmd, DocsCheckCmd} {
		cmd.Flags().StringVar(&docsTarget, "target", "README.md", "Markdown file with generated blocks")
	}
}

func runDocsGenerate(cmd *cobra.Command, _ []string) {
	runner := cliutil.NewRunnerFromCmd(cmd, true)
	runner.SetAction("docs_generate")

	//nolint:gosec // docsTarget is an explicit local CLI target path.
	current, err := os.ReadFile(docsTarget)
	if err != nil {
		runner.Fatal(err.Error())
	}
	updated, err := docs.UpdateReadme(cmd.Root(), current)
	if err != nil {
		runner.Fatal(err.Error())
	}
	changed := !bytes.Equal(current, updated)
	if changed {
		if err := os.WriteFile(docsTarget, updated, docsFileMode(docsTarget)); err != nil {
			runner.Fatal(err.Error())
		}
	}
	runner.PrintJSON(map[string]any{
		"ok":      true,
		"target":  docsTarget,
		"changed": changed,
	})
}

func docsFileMode(path string) fs.FileMode {
	info, err := os.Stat(path)
	if err == nil {
		return info.Mode().Perm()
	}
	return 0o600
}

func runDocsCheck(cmd *cobra.Command, _ []string) {
	runner := cliutil.NewRunnerFromCmd(cmd, true)
	runner.SetAction("docs_check")

	//nolint:gosec // docsTarget is an explicit local CLI target path.
	current, err := os.ReadFile(docsTarget)
	if err != nil {
		runner.Fatal(err.Error())
	}
	updated, err := docs.UpdateReadme(cmd.Root(), current)
	if err != nil {
		runner.Fatal(err.Error())
	}
	if !bytes.Equal(current, updated) {
		runner.PrintJSON(map[string]any{
			"ok":      false,
			"target":  docsTarget,
			"message": fmt.Sprintf("%s is out of date; run: agent-telegram docs generate --target %s", docsTarget, docsTarget),
		})
		cliutil.Exit(1)
	}
	runner.PrintJSON(map[string]any{
		"ok":     true,
		"target": docsTarget,
	})
}
