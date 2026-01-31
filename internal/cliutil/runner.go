// Package cliutil provides shared CLI utilities.
package cliutil

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"agent-telegram/internal/ipc"
)

// RPCClient defines the interface for RPC calls.
type RPCClient interface {
	Call(method string, params any) (any, *ipc.ErrorObject)
}

// Runner handles common command execution logic.
type Runner struct {
	socketFlag string
	jsonOutput bool
}

// NewRunner creates a new command runner with the given socket flag and JSON output setting.
func NewRunner(socketFlag string, jsonOutput bool) *Runner {
	return &Runner{
		socketFlag: socketFlag,
		jsonOutput: jsonOutput,
	}
}

// NewRunnerFromCmd creates a runner from a cobra command.
// It extracts the socket flag from the command's persistent flags.
func NewRunnerFromCmd(cmd *cobra.Command, jsonOutput bool) *Runner {
	socketPath, _ := cmd.Flags().GetString("socket")
	return &Runner{
		socketFlag: socketPath,
		jsonOutput: jsonOutput,
	}
}

// NewRunnerFromRoot creates a runner from a root command with socket flag.
func NewRunnerFromRoot(rootCmd *cobra.Command, jsonOutput bool) *Runner {
	socketPath, _ := rootCmd.Flags().GetString("socket")
	return &Runner{
		socketFlag: socketPath,
		jsonOutput: jsonOutput,
	}
}

// Client creates a new IPC client.
func (r *Runner) Client() RPCClient {
	return ipc.NewClient(r.socketFlag)
}

// Call executes an RPC call and returns the result or exits on error.
func (r *Runner) Call(method string, params any) any {
	client := r.Client()
	result, err := client.Call(method, params)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err.Message)
		os.Exit(1)
	}
	return result
}

// CallWithParams executes an RPC call with parameters.
func (r *Runner) CallWithParams(method string, params map[string]any) any {
	return r.Call(method, params)
}

// PrintResult prints the result in JSON or human-readable format.
func (r *Runner) PrintResult(result any, formatter func(any)) {
	switch {
	case r.jsonOutput || formatter == nil:
		r.PrintJSON(result)
	default:
		formatter(result)
	}
}

// PrintJSON prints the result as JSON.
func (r *Runner) PrintJSON(result any) {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(string(data))
}

// MustParseInt64 parses an int64 from a string or exits on error.
func (r *Runner) MustParseInt64(s string) int64 {
	var value int64
	_, err := fmt.Sscanf(s, "%d", &value)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: invalid number: %v\n", err)
		os.Exit(1)
	}
	return value
}

// MustParseInt parses an int from a string or exits on error.
func (r *Runner) MustParseInt(s string) int {
	var value int
	_, err := fmt.Sscanf(s, "%d", &value)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: invalid number: %v\n", err)
		os.Exit(1)
	}
	return value
}

// FormatSuccess formats a success message with common fields.
func FormatSuccess(result any, action string) {
	r, ok := result.(map[string]any)
	if !ok {
		fmt.Printf("%s succeeded!\n", action)
		return
	}

	fmt.Printf("%s sent successfully!\n", action)
	if id, ok := r["id"].(float64); ok {
		fmt.Printf("  ID: %d\n", int64(id))
	}
	if peer, ok := r["peer"].(string); ok {
		fmt.Printf("  Peer: %s\n", peer)
	}
}

// ExtractString safely extracts a string from a map.
func ExtractString(m map[string]any, key string) string {
	if val, ok := m[key].(string); ok {
		return val
	}
	return ""
}

// ExtractFloat64 safely extracts a float64 from a map.
func ExtractFloat64(m map[string]any, key string) float64 {
	if val, ok := m[key].(float64); ok {
		return val
	}
	return 0
}

// ExtractInt64 safely extracts an int64 from a map.
func ExtractInt64(m map[string]any, key string) int64 {
	return int64(ExtractFloat64(m, key))
}

// ToMap converts any value to a map[string]any safely.
func ToMap(result any) (map[string]any, bool) {
	m, ok := result.(map[string]any)
	return m, ok
}
