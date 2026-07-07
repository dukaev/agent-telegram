// Package cliutil provides shared CLI utilities.
package cliutil

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"sync"
	"time"

	"github.com/spf13/cobra"

	"agent-telegram/internal/ipc"
	"agent-telegram/internal/observability"
	"agent-telegram/internal/operations"
	"agent-telegram/internal/paths"
)

// RPCClient defines the interface for RPC calls.
type RPCClient interface {
	Call(method string, params any) (any, *ipc.ErrorObject)
}

type traceRPCClient interface {
	CallWithTrace(method string, params any, traceID string) (any, *ipc.ErrorObject)
}

var (
	cliLoggerMu sync.Mutex
	cliLoggers  = map[string]*slog.Logger{}
)

// getCLILogger returns a logger that writes to the instance-scoped CLI log.
func getCLILogger(socketPath string) *slog.Logger {
	cliLoggerMu.Lock()
	defer cliLoggerMu.Unlock()

	if logger, ok := cliLoggers[socketPath]; ok {
		return logger
	}

	var logger *slog.Logger
	logPath, err := paths.CLILogFilePathForSocket(socketPath)
	if err != nil {
		logger = slog.New(slog.NewTextHandler(os.Stderr, nil))
		cliLoggers[socketPath] = logger
		return logger
	}
	//nolint:gosec // logPath is from trusted CLILogFilePathForSocket()
	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		logger = slog.New(slog.NewTextHandler(os.Stderr, nil))
		cliLoggers[socketPath] = logger
		return logger
	}
	logger = slog.New(slog.NewJSONHandler(f, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	cliLoggers[socketPath] = logger
	return logger
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
	traceID       string
	outputBudget  OutputBudgetOptions
	receipt       bool
	lastMethod    string
	lastSafety    string
}

// NewRunner creates a new command runner with the given socket flag and JSON output setting.
func NewRunner(socketFlag string, jsonOutput bool) *Runner {
	return &Runner{
		socketFlag: socketFlag,
		jsonOutput: jsonOutput,
		traceID:    observability.NewTraceID(),
		outputBudget: OutputBudgetOptions{
			Verbosity: VerbosityFull,
		},
	}
}

// NewRunnerFromCmd creates a runner from a cobra command.
// It extracts the socket, quiet, output, filter, and dry-run flags.
// The jsonOutput parameter is deprecated (JSON is now the default) and ignored.
func NewRunnerFromCmd(cmd *cobra.Command, _ bool) *Runner {
	socketPath, _ := cmd.Flags().GetString("socket")
	quiet, _ := cmd.Flags().GetBool("quiet")
	globalText, _ := cmd.Flags().GetBool("text")
	outputFlag, _ := cmd.Flags().GetString("output")
	fieldsFlag, _ := cmd.Flags().GetStringSlice("fields")
	filterFlag, _ := cmd.Flags().GetStringSlice("filter")
	verbosityFlag, _ := cmd.Flags().GetString("verbosity")
	maxItems, _ := cmd.Flags().GetInt("max-items")
	maxTextChars, _ := cmd.Flags().GetInt("max-text-chars")
	includeFields, _ := cmd.Flags().GetStringSlice("include")
	omitFields, _ := cmd.Flags().GetStringSlice("omit")
	summary, _ := cmd.Flags().GetBool("summary")
	receipt, _ := cmd.Flags().GetBool("receipt")
	dryRun, _ := cmd.Flags().GetBool("dry-run")

	format := ParseOutputFormat(outputFlag, globalText)

	var filterExprs FilterExpressions
	if len(filterFlag) > 0 {
		var err error
		filterExprs, err = ParseFilterExpressions(filterFlag)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			Exit(1)
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
		traceID:       observability.NewTraceID(),
		outputBudget: OutputBudgetOptions{
			Verbosity:    ParseVerbosity(verbosityFlag),
			MaxItems:     maxItems,
			MaxTextChars: maxTextChars,
			Include:      includeFields,
			Omit:         omitFields,
			Summary:      summary,
		},
		receipt: receipt,
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
		traceID:    observability.NewTraceID(),
		outputBudget: OutputBudgetOptions{
			Verbosity: VerbosityFull,
		},
	}
}

// Client creates a new IPC client.
func (r *Runner) Client() RPCClient {
	return ipc.NewClient(r.socketFlag)
}

// waitForServer waits for the server to become available.
func (r *Runner) waitForServer(maxWait time.Duration) bool {
	deadline := time.Now().Add(maxWait)
	for time.Now().Before(deadline) {
		client := r.Client()
		status, err := client.Call("status", nil)
		if err == nil && serverReady(status) {
			return true
		}
		time.Sleep(200 * time.Millisecond)
	}
	return false
}

// ensureServer ensures the server is running.
func (r *Runner) ensureServer() error {
	// Check if server is already running
	client := r.Client()
	status, err := client.Call("status", nil)
	if err == nil {
		if serverReady(status) {
			return nil // Server is running and Telegram is ready
		}
		r.Log("Server is running, waiting for Telegram client...")
		if !r.waitForServer(30 * time.Second) {
			return fmt.Errorf("telegram client is not ready within timeout")
		}
		return nil
	}

	// Server not running, try to start it
	if err.Code != ipc.ErrCodeServerNotRunning {
		return fmt.Errorf("failed to connect to server: %s", err.Message)
	}
	return fmt.Errorf("server is not running; run: agent-telegram serve")
}

func serverReady(status any) bool {
	m, ok := status.(map[string]any)
	if !ok {
		return false
	}
	initialized, _ := m["initialized"].(bool)
	return initialized
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
func (r *Runner) Call(method string, params any) any {
	return r.call(method, params, true)
}

// CallInternal executes an RPC call without changing user-facing receipt metadata
// and without writing CLI audit/log events. It is intended for polling steps that
// belong to a higher-level command such as --wait-reply.
func (r *Runner) CallInternal(method string, params any) any {
	return r.call(method, params, false)
}

func (r *Runner) call(method string, params any, userVisible bool) any {
	// Ensure server is running.
	if err := r.ensureServer(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		Exit(1)
	}

	log := getCLILogger(r.socketFlag)
	start := time.Now()
	if userVisible {
		r.lastMethod = method
		r.lastSafety = operationSafety(method)
	}

	client := r.Client()
	var result any
	var err *ipc.ErrorObject
	if traced, ok := client.(traceRPCClient); ok {
		result, err = traced.CallWithTrace(method, params, r.traceID)
	} else {
		result, err = client.Call(method, params)
	}

	duration := time.Since(start)
	if userVisible {
		r.lastDuration = duration
	}
	if err != nil {
		if userVisible {
			log.Info("cli: call",
				"trace_id", r.traceID,
				"method", method,
				"params", truncateAny(params),
				"duration_ms", duration.Milliseconds(),
				"error_code", err.Code,
				"error_type", errorType(err),
				"error", err.Message,
			)
			r.writeAudit(method, params, nil, err, duration)
		}
		r.handleError(err)
	}

	if userVisible {
		log.Info("cli: call",
			"trace_id", r.traceID,
			"method", method,
			"params", truncateAny(params),
			"duration_ms", duration.Milliseconds(),
			"status", "ok",
		)
		r.writeAudit(method, params, result, nil, duration)
	}

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
	Exit(1)
}

func (r *Runner) writeAudit(method string, params, result any, rpcErr *ipc.ErrorObject, duration time.Duration) {
	event := observability.AuditEvent{
		Time:       time.Now().UTC(),
		TraceID:    r.traceID,
		Surface:    "cli",
		Method:     method,
		Safety:     operationSafety(method),
		DryRun:     r.dryRun,
		Status:     "ok",
		DurationMs: duration.Milliseconds(),
		Params:     params,
	}
	if rpcErr != nil {
		event.Status = "error"
		event.ErrorCode = rpcErr.Code
		event.ErrorType = errorType(rpcErr)
		event.Error = rpcErr.Message
	} else {
		event.ResultSummary = observability.SummarizeResult(result)
	}
	if err := observability.WriteAudit(r.socketFlag, event); err != nil {
		getCLILogger(r.socketFlag).Warn("cli: audit write failed", "error", err)
	}
}

func operationSafety(method string) string {
	if op, ok := operations.Get(method); ok {
		return op.Safety
	}
	return ""
}

func errorType(err *ipc.ErrorObject) string {
	if err == nil || err.Data == nil {
		return ""
	}
	if m, ok := err.Data.(map[string]any); ok {
		if value, ok := m["type"].(string); ok {
			return value
		}
	}
	return ""
}

// CallWithParams executes an RPC call with parameters.
// If --dry-run is set, prints a preview and exits without making the call.
func (r *Runner) CallWithParams(method string, params map[string]any) any {
	if r.dryRun {
		r.printDryRun(method, params)
		Exit(0)
	}
	return r.Call(method, params)
}

// printDryRun prints a dry-run summary of the action that would be performed.
func (r *Runner) printDryRun(method string, params map[string]any) {
	r.lastMethod = method
	r.lastSafety = operationSafety(method)
	r.writeAudit(method, params, map[string]any{"dryRun": true}, nil, 0)
	if r.outputFormat == OutputJSON {
		summary := map[string]any{
			"dry_run": true,
			"method":  method,
			"params":  observability.RedactAny(params),
			"traceId": r.traceID,
		}
		r.PrintJSON(summary)
		return
	}

	fmt.Fprintln(os.Stderr, "DRY RUN — would execute:")
	fmt.Fprintf(os.Stderr, "  Method: %s\n", method)
	if len(params) > 0 {
		fmt.Fprintln(os.Stderr, "  Params:")
		redactedParams, _ := observability.RedactAny(params).(map[string]any)
		for k, v := range redactedParams {
			fmt.Fprintf(os.Stderr, "    %s: %v\n", k, v)
		}
	}
	fmt.Fprintln(os.Stderr, "\nNo changes made.")
}

// PrintResult prints the result in the configured output format.
// JSON and IDs output goes to stdout. Human-readable output uses the formatter.
// The legacy field selector is applied before JSON/IDs formatting.
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
	result = ApplyOutputBudget(result, r.outputBudget)
	if r.receipt {
		result = map[string]any{
			"ok":       true,
			"traceId":  r.traceID,
			"method":   r.lastMethod,
			"safety":   r.lastSafety,
			"duration": r.lastDuration.String(),
			"result":   result,
		}
	}
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		Exit(1)
	}
	fmt.Println(string(data))
}

// MustParseInt64 parses an int64 from a string or exits on error.
func (r *Runner) MustParseInt64(s string) int64 {
	var value int64
	_, err := fmt.Sscanf(s, "%d", &value)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: invalid number: %v\n", err)
		Exit(1)
	}
	return value
}

// MustParseInt parses an int from a string or exits on error.
func (r *Runner) MustParseInt(s string) int {
	var value int
	_, err := fmt.Sscanf(s, "%d", &value)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: invalid number: %v\n", err)
		Exit(1)
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
	Exit(1)
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
	data, err := json.Marshal(observability.RedactAny(v))
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
