// Package cliutil provides shared CLI utilities.
package cliutil

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"time"

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

// startServer attempts to start the serve command in the background.
func (r *Runner) startServer() error {
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	// Build args for serve command
	args := []string{"serve"}
	if r.socketFlag != "" {
		args = append(args, "--socket", r.socketFlag)
	}

	//nolint:gosec,noctx // execPath from os.Executable() is safe; background server needs no context
	cmd := exec.Command(execPath, args...)
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}

	return nil
}

// waitForServer waits for the server to become available.
func (r *Runner) waitForServer(maxWait time.Duration) bool {
	deadline := time.Now().Add(maxWait)
	for time.Now().Before(deadline) {
		client := r.Client()
		_, err := client.Call("status", nil)
		if err == nil {
			return true
		}
		time.Sleep(200 * time.Millisecond)
	}
	return false
}

// ensureServer ensures the server is running, starting it if necessary.
func (r *Runner) ensureServer() error {
	// Check if server is already running
	client := r.Client()
	_, err := client.Call("status", nil)
	if err == nil {
		return nil // Server is running
	}

	// Server not running, try to start it
	if err.Code == ipc.ErrCodeServerNotRunning {
		fmt.Fprintln(os.Stderr, "Server not running, starting...")
		if startErr := r.startServer(); startErr != nil {
			return startErr
		}

		// Wait for server to be ready
		if !r.waitForServer(10 * time.Second) {
			return fmt.Errorf("server failed to start within timeout")
		}
		fmt.Fprintln(os.Stderr, "Server started successfully")
		return nil
	}

	return fmt.Errorf("failed to connect to server: %s", err.Message)
}

// CallDirect executes an RPC call without auto-starting the server.
// Use this for commands like "status" that check if the server is running.
func (r *Runner) CallDirect(method string, params any) any {
	client := r.Client()
	result, err := client.Call(method, params)
	if err != nil {
		r.handleError(err)
	}
	return result
}

// Call executes an RPC call and returns the result or exits on error.
// Automatically starts the server if it's not running.
func (r *Runner) Call(method string, params any) any {
	// Ensure server is running (auto-start if needed)
	if err := r.ensureServer(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	client := r.Client()
	result, err := client.Call(method, params)
	if err != nil {
		r.handleError(err)
	}
	return result
}

// handleError handles RPC errors with user-friendly messages.
func (r *Runner) handleError(err *ipc.ErrorObject) {
	switch err.Code {
	case ipc.ErrCodeServerNotRunning:
		fmt.Fprintln(os.Stderr, "Error: Server is not running")
		fmt.Fprintln(os.Stderr, "Run: agent-telegram serve")
	case ipc.ErrCodeNotAuthorized:
		fmt.Fprintln(os.Stderr, "Error: Not authorized")
		fmt.Fprintln(os.Stderr, "Please login first: agent-telegram login")
	case ipc.ErrCodeNotInitialized:
		fmt.Fprintln(os.Stderr, "Error: Client not initialized")
		fmt.Fprintln(os.Stderr, "The server may still be starting up. Please try again in a few seconds.")
	default:
		fmt.Fprintf(os.Stderr, "Error: %s\n", err.Message)
	}
	os.Exit(1)
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

// ExtractInt64 safely extracts an int64 from a map (handles both int64 and float64).
func ExtractInt64(m map[string]any, key string) int64 {
	if v, ok := m[key].(int64); ok {
		return v
	}
	return int64(ExtractFloat64(m, key))
}

// ToMap converts any value to a map[string]any safely.
func ToMap(result any) (map[string]any, bool) {
	m, ok := result.(map[string]any)
	return m, ok
}
