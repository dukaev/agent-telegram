package cliutil

import "strings"

// ShellArg quotes one argument for copy-paste-safe POSIX shell commands.
func ShellArg(value string) string {
	if value == "" {
		return "''"
	}
	if strings.ContainsAny(value, " \t\n\"'\\$`") {
		return "'" + strings.ReplaceAll(value, "'", "'\\''") + "'"
	}
	return value
}
