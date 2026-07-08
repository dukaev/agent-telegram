// Package sys provides system commands.
package sys

import (
	"fmt"

	"github.com/spf13/cobra"

	"agent-telegram/internal/docs"
)

// LLMsTxtCmd represents the llms-txt command.
var LLMsTxtCmd = &cobra.Command{
	Use:   "llms-txt",
	Short: "Generate full CLI documentation for LLMs",
	Long: `Generate comprehensive documentation of all commands, subcommands,
flags, and arguments in a format suitable for LLMs.

This outputs detailed information about every command including:
- Command name and aliases
- Short and long descriptions
- All flags with types and defaults
- Usage examples
- Argument specifications

Example:
  agent-telegram llms-txt > llms.txt`,
}

// AddLLMsTxtCommand adds the llms-txt command to the root command.
func AddLLMsTxtCommand(rootCmd *cobra.Command) {
	rootCmd.AddCommand(LLMsTxtCmd)

	LLMsTxtCmd.Run = func(_ *cobra.Command, _ []string) {
		fmt.Print(docs.GenerateLLMMarkdown(rootCmd))
	}
}
