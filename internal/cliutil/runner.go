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
	opts := runnerFlagOptionsFromCmd(cmd)

	return &Runner{
		socketFlag:    opts.socketPath,
		jsonOutput:    opts.format == OutputJSON,
		quiet:         opts.quiet,
		outputFormat:  opts.format,
		fieldSelector: NewFieldSelector(opts.fields),
		filterExprs:   opts.filterExpressions(),
		dryRun:        opts.dryRun,
		agentMode:     opts.agentMode,
		runID:         resolveRunID(opts.runID),
		traceID:       observability.NewTraceID(),
		outputBudget:  opts.outputBudget(),
		receipt:       opts.receipt,
		commandPath:   cmd.CommandPath(),
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
		r.recordCall(method)
	}

	if !r.ensureServerReady(method) {
		return nil
	}

	log := getCLILogger(r.socketFlag)
	result, err, duration := r.callRPC(method, params)
	if userVisible {
		r.lastDuration = duration
	}
	if err != nil {
		if userVisible {
			r.logCallError(log, method, params, err, duration)
		}
		r.handleError(err)
	}

	if userVisible {
		r.logCallSuccess(log, method, params, result, duration)
	}
	return r.applyResultFilters(result)
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
