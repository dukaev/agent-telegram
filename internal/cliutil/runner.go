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

type runRPCClient interface {
	CallWithTraceAndRun(method string, params any, traceID, runID string) (any, *ipc.ErrorObject)
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
	agentMode     bool
	runID         string
	traceID       string
	outputBudget  OutputBudgetOptions
	receipt       bool
	lastMethod    string
	lastSafety    string
	commandPath   string
}

// NewRunner creates a new command runner with the given socket flag and JSON output setting.
func NewRunner(socketFlag string, jsonOutput bool) *Runner {
	return &Runner{
		socketFlag: socketFlag,
		jsonOutput: jsonOutput,
		runID:      observability.NewRunID(),
		traceID:    observability.NewTraceID(),
		outputBudget: OutputBudgetOptions{
			Verbosity: VerbosityFull,
		},
	}
}

// NewRunnerFromCmd creates a runner from a cobra command.
// It extracts the socket, quiet, output, filter, and dry-run flags.
// The second parameter is ignored; JSON is the default output mode.
func NewRunnerFromCmd(cmd *cobra.Command, _ bool) *Runner {
	socketPath, _ := cmd.Flags().GetString("socket")
	quiet, _ := cmd.Flags().GetBool("quiet")
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
	agentMode, _ := cmd.Flags().GetBool("agent")
	runIDFlag, _ := cmd.Flags().GetString("run-id")

	format := ParseOutputFormat(outputFlag)
	if agentMode {
		format = OutputJSON
		receipt = true
		quiet = true
		if !flagChanged(cmd, "verbosity") {
			verbosityFlag = string(VerbosityCompact)
		}
		if maxItems <= 0 {
			maxItems = 8
		}
		if maxTextChars <= 0 {
			maxTextChars = 180
		}
	}

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
		agentMode:     agentMode,
		runID:         resolveRunID(runIDFlag),
		traceID:       observability.NewTraceID(),
		outputBudget: OutputBudgetOptions{
			Verbosity:    ParseVerbosity(verbosityFlag),
			MaxItems:     maxItems,
			MaxTextChars: maxTextChars,
			Include:      includeFields,
			Omit:         omitFields,
			Summary:      summary,
		},
		receipt:     receipt,
		commandPath: cmd.CommandPath(),
	}
}

func flagChanged(cmd *cobra.Command, name string) bool {
	if cmd == nil {
		return false
	}
	if flag := cmd.Flags().Lookup(name); flag != nil && flag.Changed {
		return true
	}
	if flag := cmd.InheritedFlags().Lookup(name); flag != nil && flag.Changed {
		return true
	}
	return false
}

func resolveRunID(value string) string {
	if value != "" {
		return observability.SanitizeRunID(value)
	}
	return observability.NewRunID()
}

// SetIDKey sets the ID key used for --output ids mode.
func (r *Runner) SetIDKey(key string) {
	r.idKey = key
}

// SetAction sets receipt/error metadata for commands that do not call Runner.Call.
func (r *Runner) SetAction(method string) {
	r.lastMethod = method
	r.lastSafety = operationSafety(method)
}

// AgentMode reports whether the compact agent contract is enabled.
func (r *Runner) AgentMode() bool {
	return r.agentMode
}

// NewRunnerFromRoot creates a runner from a root command with socket flag.
func NewRunnerFromRoot(rootCmd *cobra.Command, jsonOutput bool) *Runner {
	socketPath, _ := rootCmd.Flags().GetString("socket")
	return &Runner{
		socketFlag: socketPath,
		jsonOutput: jsonOutput,
		runID:      observability.NewRunID(),
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
	return fmt.Errorf("server is not running; run: agent-telegram server ensure")
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
	if userVisible {
		r.lastMethod = method
		r.lastSafety = operationSafety(method)
	}

	// Ensure server is running.
	if err := r.ensureServer(); err != nil {
		if r.agentMode {
			r.handleError(r.ensureErrorToRPC(err, method))
			return nil
		}
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		Exit(1)
	}

	log := getCLILogger(r.socketFlag)
	start := time.Now()

	client := r.Client()
	var result any
	var err *ipc.ErrorObject
	if runAware, ok := client.(runRPCClient); ok {
		result, err = runAware.CallWithTraceAndRun(method, params, r.traceID, r.runID)
	} else if traced, ok := client.(traceRPCClient); ok {
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
				"run_id", r.runID,
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
			"run_id", r.runID,
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
	if r.agentMode {
		r.printErrorEnvelope(err)
		Exit(1)
		return
	}
	switch err.Code {
	case ipc.ErrCodeServerNotRunning:
		fmt.Fprintln(os.Stderr, "Error: Server is not running")
		fmt.Fprintln(os.Stderr, "Run: agent-telegram server ensure")
	case ipc.ErrCodeNotAuthorized:
		fmt.Fprintln(os.Stderr, "Error: Not authorized")
		fmt.Fprintln(os.Stderr, "Please authenticate first: agent-telegram auth web")
	case ipc.ErrCodeNotInitialized:
		fmt.Fprintln(os.Stderr, "Error: Client not initialized")
		fmt.Fprintln(os.Stderr, "The server may still be starting up. Please try again in a few seconds.")
	default:
		fmt.Fprintf(os.Stderr, "Error: %s\n", err.Message)
	}
	Exit(1)
}

func (r *Runner) ensureErrorToRPC(err error, method string) *ipc.ErrorObject {
	msg := err.Error()
	switch msg {
	case "server is not running; run: agent-telegram server ensure":
		return ipc.NewTypedError(ipc.ErrCodeServerNotRunning, ipc.ErrorTypeServerNotRunning, msg, map[string]any{
			"nextCommand": agentCommand(r.runID, "server", "ensure"),
		})
	case "telegram client is not ready within timeout":
		return ipc.NewTypedError(ipc.ErrCodeNotInitialized, ipc.ErrorTypeNotInitialized, msg, map[string]any{
			"nextCommand": agentCommand(r.runID, "server", "wait-ready"),
		})
	default:
		return ipc.NewTypedError(-32000, ipc.ErrorTypeInternal, fmt.Sprintf("%s failed before %s", method, msg), nil)
	}
}

func (r *Runner) printErrorEnvelope(err *ipc.ErrorObject) {
	payload := map[string]any{
		"ok":          false,
		"runId":       r.runID,
		"traceId":     r.traceID,
		"command":     r.commandPath,
		"method":      r.lastMethod,
		"safety":      r.lastSafety,
		"error":       errorObjectForOutput(err),
		"diagnosis":   diagnosisForError(err),
		"nextActions": nextActionsForError(err, r),
	}
	data, jsonErr := json.MarshalIndent(payload, "", "  ")
	if jsonErr != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err.Message)
		return
	}
	fmt.Println(string(data))
}

func errorObjectForOutput(err *ipc.ErrorObject) map[string]any {
	out := map[string]any{
		"code":      err.Code,
		"type":      errorType(err),
		"message":   err.Message,
		"retryable": retryableError(err),
	}
	if err.Data != nil {
		out["data"] = observability.RedactAny(err.Data)
	}
	return out
}

func diagnosisForError(err *ipc.ErrorObject) map[string]any {
	errType := errorType(err)
	switch errType {
	case ipc.ErrorTypeServerNotRunning:
		return map[string]any{
			"category": "server_not_running",
			"summary":  "The local IPC server is not reachable.",
			"retry":    "Start the server, then retry the original command.",
		}
	case ipc.ErrorTypeNotAuthorized:
		return map[string]any{
			"category": "not_authorized",
			"summary":  "Telegram session is missing or expired.",
			"retry":    "Authenticate first, then retry the original command.",
		}
	case ipc.ErrorTypeNotInitialized:
		return map[string]any{
			"category": "not_initialized",
			"summary":  "The server is running but Telegram is not ready yet.",
			"retry":    "Wait for readiness, then retry.",
		}
	case ipc.ErrorTypeFloodWait:
		return map[string]any{
			"category": "rate_limited",
			"summary":  "Telegram returned a flood wait.",
			"retry":    "Wait for retryAfter seconds when present.",
		}
	case ipc.ErrorTypeValidation:
		return map[string]any{
			"category": "validation",
			"summary":  "Command parameters failed validation.",
			"retry":    "Inspect the schema and retry with valid parameters.",
		}
	case ipc.ErrorTypePeerNotFound:
		return map[string]any{
			"category": "peer_not_found",
			"summary":  "Telegram could not resolve the requested peer.",
			"retry":    "Verify the username, ID, invite link, or access rights.",
		}
	default:
		return map[string]any{
			"category": "unknown",
			"summary":  "The command failed.",
			"retry":    "Inspect audit/logs with the trace ID.",
		}
	}
}

func nextActionsForError(err *ipc.ErrorObject, r *Runner) []map[string]any {
	errType := errorType(err)
	switch errType {
	case ipc.ErrorTypeServerNotRunning:
		return []map[string]any{{
			"kind":    "start_server",
			"command": agentCommand(r.runID, "server", "ensure"),
			"safety":  "local",
			"reason":  "server is required before RPC-backed commands",
		}}
	case ipc.ErrorTypeNotAuthorized:
		return []map[string]any{{
			"kind":    "authenticate",
			"command": "AGENT_TELEGRAM_PHONE=... agent-telegram auth web --agent --run-id " + r.runID,
			"safety":  "sensitive",
			"reason":  "Telegram session is required",
		}}
	case ipc.ErrorTypeNotInitialized:
		return []map[string]any{{
			"kind":    "wait_ready",
			"command": agentCommand(r.runID, "server", "wait-ready"),
			"safety":  "local",
			"reason":  "server is still initializing",
		}}
	case ipc.ErrorTypeValidation:
		command := r.commandPath + " --schema"
		if r.commandPath == "" {
			command = "agent-telegram manifest"
		}
		return []map[string]any{{
			"kind":    "inspect_schema",
			"command": command,
			"safety":  "read",
			"reason":  "schema explains required parameters",
		}}
	default:
		return []map[string]any{{
			"kind":    "inspect_trace",
			"command": "agent-telegram trace inspect " + r.traceID + " --agent --run-id " + r.runID,
			"safety":  "read",
			"reason":  "trace bundle can show related audit and log events",
		}}
	}
}

func retryableError(err *ipc.ErrorObject) bool {
	switch errorType(err) {
	case ipc.ErrorTypeServerNotRunning, ipc.ErrorTypeNotInitialized, ipc.ErrorTypeTimeout, ipc.ErrorTypeFloodWait:
		return true
	default:
		return false
	}
}

func agentCommand(runID string, args ...string) string {
	base := "agent-telegram"
	for _, arg := range args {
		base += " " + arg
	}
	base += " --agent"
	if runID != "" {
		base += " --run-id " + runID
	}
	return base
}

func (r *Runner) writeAudit(method string, params, result any, rpcErr *ipc.ErrorObject, duration time.Duration) {
	event := observability.AuditEvent{
		Time:       time.Now().UTC(),
		RunID:      r.runID,
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
			"runId":   r.runID,
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
// JSON and IDs output goes to stdout.
func (r *Runner) PrintResult(result any, _ func(any)) {
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
	default:
		r.PrintJSON(result)
	}
}

// PrintJSON prints the result as JSON.
func (r *Runner) PrintJSON(result any) {
	result = ApplyOutputBudget(result, r.outputBudget)
	if r.receipt {
		result = map[string]any{
			"ok":       true,
			"runId":    r.runID,
			"traceId":  r.traceID,
			"command":  r.commandPath,
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
		r.Fatal(fmt.Sprintf("invalid number: %v", err))
	}
	return value
}

// MustParseInt parses an int from a string or exits on error.
func (r *Runner) MustParseInt(s string) int {
	var value int
	_, err := fmt.Sscanf(s, "%d", &value)
	if err != nil {
		r.Fatal(fmt.Sprintf("invalid number: %v", err))
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
	if r.agentMode {
		r.printErrorEnvelope(ipc.NewTypedError(ipc.ErrCodeInvalidParams, ipc.ErrorTypeValidation, msg, nil))
		Exit(1)
		return
	}
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
