// Package cliutil provides shared CLI utilities.
package cliutil

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"sync"
	"time"

	"github.com/spf13/cobra"

	"agent-telegram/internal/ipc"
	"agent-telegram/internal/paths"
)

// RPCClient defines the interface for RPC calls.
type RPCClient interface {
	Call(method string, params any) (any, *ipc.ErrorObject)
}

var (
	cliLogger     *slog.Logger
	cliLoggerOnce sync.Once
)

// getCLILogger returns a logger that writes to ~/.agent-telegram/cli.log.
func getCLILogger() *slog.Logger {
	cliLoggerOnce.Do(func() {
		logPath, err := paths.CLILogFilePath()
		if err != nil {
			cliLogger = slog.New(slog.NewTextHandler(os.Stderr, nil))
			return
		}
		//nolint:gosec // logPath is from trusted CLILogFilePath()
		f, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
		if err != nil {
			cliLogger = slog.New(slog.NewTextHandler(os.Stderr, nil))
			return
		}
		cliLogger = slog.New(slog.NewJSONHandler(f, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		}))
	})
	return cliLogger
}

// Runner handles common command execution logic.
type Runner struct {
	socketFlag    string
	jsonOutput    bool
	quiet         bool
	lastDuration  time.Duration
	outputFormat  OutputFormat
	idKey         string
	fieldSelector *FieldSelector
	filterExprs   FilterExpressions
	dryRun        bool
}

// NewRunner creates a new command runner with the given socket flag and JSON output setting.
func NewRunner(socketFlag string, jsonOutput bool) *Runner {
	return &Runner{
		socketFlag: socketFlag,
		jsonOutput: jsonOutput,
	}
}

// NewRunnerFromCmd creates a runner from a cobra command.
// It extracts the socket, quiet, text, output, fields, filter, and dry-run flags.
// The jsonOutput parameter is deprecated (JSON is now the default) and ignored.
func NewRunnerFromCmd(cmd *cobra.Command, _ bool) *Runner {
	socketPath, _ := cmd.Flags().GetString("socket")
	quiet, _ := cmd.Flags().GetBool("quiet")
	globalText, _ := cmd.Flags().GetBool("text")
	outputFlag, _ := cmd.Flags().GetString("output")
	fieldsFlag, _ := cmd.Flags().GetStringSlice("fields")
	filterFlag, _ := cmd.Flags().GetStringSlice("filter")
	dryRun, _ := cmd.Flags().GetBool("dry-run")

	format := ParseOutputFormat(outputFlag, globalText)

	var filterExprs FilterExpressions
	if len(filterFlag) > 0 {
		var err error
		filterExprs, err = ParseFilterExpressions(filterFlag)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	}

	return &Runner{
		socketFlag:    socketPath,
		jsonOutput:    format == OutputJSON,
		quiet:         quiet,
		outputFormat:  format,
		fieldSelector: NewFieldSelector(fieldsFlag),
		filterExprs:   filterExprs,
		dryRun:        dryRun,
	}
}

// SetIDKey sets the ID key used for --output ids mode.
func (r *Runner) SetIDKey(key string) {
	r.idKey = key
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
// Uses a lock file to prevent race conditions when multiple CLI processes
// try to start the server simultaneously.
func (r *Runner) ensureServer() error {
	// Check if server is already running
	client := r.Client()
	_, err := client.Call("status", nil)
	if err == nil {
		return nil // Server is running
	}

	// Server not running, try to start it
	if err.Code != ipc.ErrCodeServerNotRunning {
		return fmt.Errorf("failed to connect to server: %s", err.Message)
	}

	// Acquire lock to prevent multiple processes from starting server simultaneously
	lockPath, lockErr := paths.LockFilePath()
	if lockErr != nil {
		// If we can't get lock path, try to start anyway
		return r.startServerWithWait()
	}

	lock := paths.NewLockFile(lockPath)
	acquired, lockErr := lock.TryLock()
	if lockErr != nil {
		// If lock fails, try to start anyway
		return r.startServerWithWait()
	}

	if !acquired {
		// Another process is starting the server, just wait for it
		r.Log("Another process is starting the server, waiting...")
		if !r.waitForServer(15 * time.Second) {
			return fmt.Errorf("server failed to start within timeout")
		}
		return nil
	}

	// We have the lock, check again if server started while we were acquiring lock
	_, err = client.Call("status", nil)
	if err == nil {
		_ = lock.Unlock()
		return nil // Server started by another process
	}

	// Start the server - release lock immediately so serve can acquire it
	r.Log("Server not running, starting...")
	if startErr := r.startServer(); startErr != nil {
		_ = lock.Unlock()
		return startErr
	}

	// Release lock immediately so serve can acquire it
	_ = lock.Unlock()

	// Wait for server to be ready (Telegram auth can take time)
	if !r.waitForServer(30 * time.Second) {
		return fmt.Errorf("server failed to start within timeout")
	}

	r.Log("Server started successfully")
	return nil
}

// startServerWithWait starts the server and waits for it to be ready.
func (r *Runner) startServerWithWait() error {
	r.Log("Server not running, starting...")
	if startErr := r.startServer(); startErr != nil {
		return startErr
	}

	if !r.waitForServer(30 * time.Second) {
		return fmt.Errorf("server failed to start within timeout")
	}
	r.Log("Server started successfully")
	return nil
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

	log := getCLILogger()
	start := time.Now()

	client := r.Client()
	result, err := client.Call(method, params)

	duration := time.Since(start)
	r.lastDuration = duration
	if err != nil {
		log.Info("cli: call",
			"method", method,
			"params", truncateAny(params),
			"duration_ms", duration.Milliseconds(),
			"error_code", err.Code,
			"error", err.Message,
		)
		r.handleError(err)
	}

	log.Info("cli: call",
		"method", method,
		"params", truncateAny(params),
		"duration_ms", duration.Milliseconds(),
		"status", "ok",
	)

	// Apply --filter expressions (early: modifies data for all formatters)
	if len(r.filterExprs) > 0 {
		result = r.filterExprs.Apply(result)
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
// If --dry-run is set, prints a preview and exits without making the call.
func (r *Runner) CallWithParams(method string, params map[string]any) any {
	if r.dryRun {
		r.printDryRun(method, params)
		os.Exit(0)
	}
	return r.Call(method, params)
}

// printDryRun prints a dry-run summary of the action that would be performed.
func (r *Runner) printDryRun(method string, params map[string]any) {
	if r.outputFormat == OutputJSON {
		summary := map[string]any{
			"dry_run": true,
			"method":  method,
			"params":  params,
		}
		r.PrintJSON(summary)
		return
	}

	fmt.Fprintln(os.Stderr, "DRY RUN — would execute:")
	fmt.Fprintf(os.Stderr, "  Method: %s\n", method)
	if len(params) > 0 {
		fmt.Fprintln(os.Stderr, "  Params:")
		for k, v := range params {
			fmt.Fprintf(os.Stderr, "    %s: %v\n", k, v)
		}
	}
	fmt.Fprintln(os.Stderr, "\nNo changes made.")
}

// PrintResult prints the result in the configured output format.
// JSON and IDs output goes to stdout. Human-readable output uses the formatter.
// --fields is applied before JSON/IDs formatting (not text).
func (r *Runner) PrintResult(result any, formatter func(any)) {
	switch r.outputFormat {
	case OutputJSON:
		if r.fieldSelector != nil {
			result = r.fieldSelector.Apply(result)
		}
		r.PrintJSON(result)
	case OutputIDs:
		if r.fieldSelector != nil {
			result = r.fieldSelector.Apply(result)
		}
		printIDs(result, r.idKey)
	case OutputText:
		if formatter == nil {
			r.PrintJSON(result)
			return
		}
		if r.quiet {
			return
		}
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
// Output goes to stderr so stdout remains clean for data.
func FormatSuccess(result any, action string) {
	r, ok := result.(map[string]any)
	if !ok {
		fmt.Fprintf(os.Stderr, "%s succeeded!\n", action)
		return
	}

	fmt.Fprintf(os.Stderr, "%s sent successfully!\n", action)
	if id, ok := r["id"].(float64); ok {
		fmt.Fprintf(os.Stderr, "  ID: %d\n", int64(id))
	}
	if peer, ok := r["peer"].(string); ok {
		fmt.Fprintf(os.Stderr, "  Peer: %s\n", peer)
	}
}

// Fatal prints an error message to stderr and exits with code 1.
func (r *Runner) Fatal(msg string) {
	fmt.Fprintf(os.Stderr, "Error: %s\n", msg)
	os.Exit(1)
}

// LastDuration returns the duration of the last Call().
func (r *Runner) LastDuration() time.Duration {
	return r.lastDuration
}

// IsQuiet returns true if quiet mode is enabled.
func (r *Runner) IsQuiet() bool {
	return r.quiet
}

// Logf prints a message to stderr if not in quiet mode.
func (r *Runner) Logf(format string, args ...any) {
	if !r.quiet {
		fmt.Fprintf(os.Stderr, format, args...)
	}
}

// Log prints a message to stderr if not in quiet mode.
func (r *Runner) Log(msg string) {
	if !r.quiet {
		fmt.Fprintln(os.Stderr, msg)
	}
}

const maxLogSize = 1024

// truncateAny returns a truncated JSON string representation of a value.
func truncateAny(v any) string {
	data, err := json.Marshal(v)
	if err != nil {
		return fmt.Sprintf("%v", v)
	}
	s := string(data)
	if len(s) > maxLogSize {
		return s[:maxLogSize] + "..."
	}
	return s
}

// ExtractString safely extracts a string from a map.
func ExtractString(m map[string]any, key string) string {
	if val, ok := m[key].(string); ok {
		return val
	}
	return ""
}

// ExtractFloat64 safely extracts a float64 from a map (handles float64 and json.Number).
func ExtractFloat64(m map[string]any, key string) float64 {
	switch v := m[key].(type) {
	case float64:
		return v
	case json.Number:
		f, _ := v.Float64()
		return f
	}
	return 0
}

// ExtractInt64 safely extracts an int64 from a map (handles int64, float64, and json.Number).
func ExtractInt64(m map[string]any, key string) int64 {
	switch v := m[key].(type) {
	case int64:
		return v
	case json.Number:
		n, _ := v.Int64()
		return n
	case float64:
		return int64(v)
	}
	return 0
}

// ToMap converts any value to a map[string]any safely.
func ToMap(result any) (map[string]any, bool) {
	m, ok := result.(map[string]any)
	return m, ok
}
