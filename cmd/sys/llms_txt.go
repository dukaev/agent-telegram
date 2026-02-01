// Package sys provides system commands.
package sys

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// LLMsTxtCmd represents the llms-txt command.
var LLMsTxtCmd = &cobra.Command{
	Use:     "llms-txt",
	Short:   "Generate full CLI documentation for LLMs",
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
		printLLMsDocumentation(rootCmd)
	}
}

// printLLMsDocumentation generates full documentation for all commands.
func printLLMsDocumentation(rootCmd *cobra.Command) {
	fmt.Println("# agent-telegram CLI Documentation")
	fmt.Println()
	fmt.Println("Telegram IPC agent CLI - A command-line tool for interacting with Telegram via IPC server.")
	fmt.Println()
	fmt.Println("## Global Flags")
	fmt.Println()
	printFlags(rootCmd.PersistentFlags())
	fmt.Println()
	fmt.Println("---")
	fmt.Println()
	fmt.Println("## Commands")
	fmt.Println()

	// Collect commands by group
	groups := make(map[string][]*cobra.Command)
	for _, cmd := range rootCmd.Commands() {
		if !cmd.IsAvailableCommand() || cmd.IsAdditionalHelpTopicCommand() {
			continue
		}
		groupID := cmd.GroupID
		if groupID == "" {
			groupID = "other"
		}
		groups[groupID] = append(groups[groupID], cmd)
	}

	// Define group order and titles
	groupOrder := []string{"server", "auth", "message", "chat", "get", "other"}
	groupTitles := map[string]string{
		"server":  "Server Commands",
		"auth":    "Authentication Commands",
		"message": "Message Commands",
		"chat":    "Chat Commands",
		"get":     "Get Commands",
		"other":   "Other Commands",
	}

	for _, groupID := range groupOrder {
		cmds, ok := groups[groupID]
		if !ok || len(cmds) == 0 {
			continue
		}

		fmt.Printf("### %s\n\n", groupTitles[groupID])
		for _, cmd := range cmds {
			printCommandDoc(cmd, 0)
		}
	}
}

// printCommandDoc prints detailed documentation for a command and its subcommands.
func printCommandDoc(cmd *cobra.Command, depth int) {
	indent := strings.Repeat("  ", depth)
	headerLevel := strings.Repeat("#", 4+depth)
	if depth > 2 {
		headerLevel = "######" // Max header level
	}

	// Command header
	fmt.Printf("%s%s `%s`\n\n", indent, headerLevel, cmd.CommandPath())

	// Short description
	if cmd.Short != "" {
		fmt.Printf("%s%s\n\n", indent, cmd.Short)
	}

	// Long description
	if cmd.Long != "" && cmd.Long != cmd.Short {
		fmt.Printf("%s**Description:**\n\n", indent)
		// Indent each line of the long description
		for _, line := range strings.Split(cmd.Long, "\n") {
			fmt.Printf("%s%s\n", indent, line)
		}
		fmt.Println()
	}

	// Usage
	fmt.Printf("%s**Usage:**\n\n", indent)
	fmt.Printf("%s```\n", indent)
	fmt.Printf("%s%s\n", indent, cmd.UseLine())
	fmt.Printf("%s```\n\n", indent)

	// Aliases
	if len(cmd.Aliases) > 0 {
		fmt.Printf("%s**Aliases:** %s\n\n", indent, strings.Join(cmd.Aliases, ", "))
	}

	// Arguments
	if cmd.Args != nil {
		argsDesc := getArgsDescription(cmd)
		if argsDesc != "" {
			fmt.Printf("%s**Arguments:** %s\n\n", indent, argsDesc)
		}
	}

	// Flags (use Flags() directly to avoid merge conflicts with shorthand flags)
	flags := cmd.Flags()
	if flags.HasFlags() {
		fmt.Printf("%s**Flags:**\n\n", indent)
		printFlagsWithIndent(flags, indent)
		fmt.Println()
	}

	// Example
	if cmd.Example != "" {
		fmt.Printf("%s**Example:**\n\n", indent)
		fmt.Printf("%s```\n", indent)
		for _, line := range strings.Split(cmd.Example, "\n") {
			fmt.Printf("%s%s\n", indent, line)
		}
		fmt.Printf("%s```\n\n", indent)
	}

	// Separator
	fmt.Printf("%s---\n\n", indent)

	// Subcommands
	for _, subcmd := range cmd.Commands() {
		if !subcmd.IsAvailableCommand() || subcmd.IsAdditionalHelpTopicCommand() {
			continue
		}
		printCommandDoc(subcmd, depth+1)
	}
}

// printFlags prints all flags in a readable format.
func printFlags(flags *pflag.FlagSet) {
	printFlagsWithIndent(flags, "")
}

// printFlagsWithIndent prints all flags with a given indent.
func printFlagsWithIndent(flags *pflag.FlagSet, indent string) {
	flags.VisitAll(func(f *pflag.Flag) {
		// Build flag string
		flagStr := ""
		if f.Shorthand != "" {
			flagStr = fmt.Sprintf("-%s, --%s", f.Shorthand, f.Name)
		} else {
			flagStr = fmt.Sprintf("--%s", f.Name)
		}

		// Add type
		flagType := f.Value.Type()
		if flagType != "bool" {
			flagStr += fmt.Sprintf(" <%s>", flagType)
		}

		// Add default if present and not empty
		defaultVal := f.DefValue
		defaultStr := ""
		if defaultVal != "" && defaultVal != "false" && defaultVal != "0" && defaultVal != "[]" {
			defaultStr = fmt.Sprintf(" (default: %s)", defaultVal)
		}

		// Print flag line
		fmt.Printf("%s- `%s`%s\n", indent, flagStr, defaultStr)
		if f.Usage != "" {
			fmt.Printf("%s  %s\n", indent, f.Usage)
		}
	})
}

// getArgsDescription returns a description of the command's expected arguments.
func getArgsDescription(cmd *cobra.Command) string {
	// Extract from Use string
	use := cmd.Use
	if idx := strings.Index(use, " "); idx != -1 {
		args := use[idx+1:]
		if args != "" {
			return "`" + args + "`"
		}
	}
	return ""
}
