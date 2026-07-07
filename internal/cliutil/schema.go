// Package cliutil provides shared CLI utilities.
package cliutil

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"agent-telegram/internal/operations"
)

// commandMethods maps cobra commands to their IPC method names.
var commandMethods = map[*cobra.Command]string{}

// RegisterMethod associates a cobra command with its IPC method name.
// It also wraps the command's Args validator to skip validation when --schema is set.
func RegisterMethod(cmd *cobra.Command, method string) {
	commandMethods[cmd] = method
	if cmd.Args != nil {
		original := cmd.Args
		cmd.Args = func(c *cobra.Command, args []string) error {
			if s, _ := c.Flags().GetBool("schema"); s {
				return nil
			}
			return original(c, args)
		}
	}
}

// ResultMethodNames returns all IPC methods with a registered result schema.
func ResultMethodNames() []string {
	return operations.Methods()
}

// CommandMethodNames returns all IPC methods associated with CLI commands.
func CommandMethodNames() []string {
	seen := make(map[string]struct{}, len(commandMethods))
	for _, method := range commandMethods {
		seen[method] = struct{}{}
	}
	methods := make([]string, 0, len(seen))
	for method := range seen {
		methods = append(methods, method)
	}
	return methods
}

// HasResultSchema reports whether a method has a result schema.
func HasResultSchema(method string) bool {
	return operations.HasSchema(method)
}

// PrintCommandSchema outputs the result schema for a command and exits.
// Called from Execute() before cobra's validation to bypass required flag/arg checks.
func PrintCommandSchema(cmd *cobra.Command) {
	method, ok := commandMethods[cmd]
	if !ok {
		fmt.Fprintln(os.Stderr, "Error: no schema available for this command")
		Exit(1)
	}
	if !printSchema(method) {
		fmt.Fprintf(os.Stderr, "Error: no schema for method %q\n", method)
		Exit(1)
	}
	Exit(0)
}

// printSchema outputs the JSON schema for a method. Returns true if method was found.
func printSchema(method string) bool {
	op, ok := operations.Get(method)
	if !ok {
		return false
	}
	manifest := operations.ManifestFor(op)
	output := map[string]any{
		"method":               method,
		"summary":              manifest.Summary,
		"safety":               manifest.Safety,
		"idempotent":           manifest.Idempotent,
		"retryable":            manifest.Retryable,
		"requiresConfirmation": manifest.RequiresConfirmation,
		"inputSchema":          manifest.InputSchema,
		"outputSchema":         manifest.OutputSchema,
		"schema":               manifest.OutputSchema,
	}
	data, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		Exit(1)
	}
	fmt.Println(string(data))
	return true
}
