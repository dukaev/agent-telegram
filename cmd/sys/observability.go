package sys

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"agent-telegram/internal/cliutil"
	"agent-telegram/internal/observability"
	"agent-telegram/internal/paths"
)

var (
	logsKind    string
	logsLast    int
	logsTraceID string
	logsRunID   string
	logsRedact  string
	logsFollow  bool
	logsPoll    time.Duration

	auditLast    int
	auditTraceID string
	auditRunID   string
	auditMethod  string
	auditSince   time.Duration
	auditRedact  string

	inspectLast   int
	inspectRedact string
)

// LogsCmd tails instance-scoped CLI/server logs.
var LogsCmd = &cobra.Command{
	GroupID: "server",
	Use:     "logs",
	Short:   "Read agent-telegram logs",
	Run:     runLogs,
}

// AuditCmd reads the redacted action audit journal.
var AuditCmd = &cobra.Command{
	GroupID: "server",
	Use:     "audit",
	Short:   "Read redacted action audit events",
	Run:     runAudit,
}

// TraceCmd groups trace inspection helpers.
var TraceCmd = &cobra.Command{
	GroupID: "server",
	Use:     "trace",
	Short:   "Inspect one traced operation",
}

// TraceInspectCmd returns an audit/log bundle for one trace ID.
var TraceInspectCmd = &cobra.Command{
	Use:   "inspect <trace_id>",
	Short: "Inspect audit and logs for a trace ID",
	Args:  cobra.ExactArgs(1),
	Run:   runTraceInspect,
}

// RunCmd groups agent run inspection helpers.
var RunCmd = &cobra.Command{
	GroupID: "server",
	Use:     "run",
	Short:   "Inspect one agent run",
}

// RunInspectCmd returns an audit/log bundle for one run ID.
var RunInspectCmd = &cobra.Command{
	Use:   "inspect <run_id>",
	Short: "Inspect audit and logs for a run ID",
	Args:  cobra.ExactArgs(1),
	Run:   runRunInspect,
}

// AddObservabilityCommands adds logs and audit commands.
func AddObservabilityCommands(rootCmd *cobra.Command) {
	rootCmd.AddCommand(LogsCmd, AuditCmd, TraceCmd, RunCmd)
	TraceCmd.AddCommand(TraceInspectCmd)
	RunCmd.AddCommand(RunInspectCmd)

	LogsCmd.Flags().StringVar(&logsKind, "kind", "cli", "Log kind: cli or server")
	LogsCmd.Flags().IntVar(&logsLast, "last", 50, "Maximum log lines to return")
	LogsCmd.Flags().StringVar(&logsTraceID, "trace-id", "", "Filter log lines by trace ID")
	LogsCmd.Flags().StringVar(&logsRunID, "run-id", "", "Filter log lines by run ID")
	LogsCmd.Flags().StringVar(&logsRedact, "redaction", "safe", "Redaction mode: safe or redacted")
	LogsCmd.Flags().BoolVarP(&logsFollow, "follow", "f", false, "Follow log output as JSON Lines")
	LogsCmd.Flags().DurationVar(&logsPoll, "interval", time.Second, "Polling interval for --follow")

	AuditCmd.Flags().IntVar(&auditLast, "last", 50, "Maximum audit events to return")
	AuditCmd.Flags().StringVar(&auditTraceID, "trace-id", "", "Filter audit events by trace ID")
	AuditCmd.Flags().StringVar(&auditRunID, "run-id", "", "Filter audit events by run ID")
	AuditCmd.Flags().StringVar(&auditMethod, "method", "", "Filter audit events by method")
	AuditCmd.Flags().DurationVar(&auditSince, "since", 0, "Filter audit events since duration, e.g. 1h")
	AuditCmd.Flags().StringVar(&auditRedact, "redaction", "safe", "Redaction mode: safe or redacted")

	for _, cmd := range []*cobra.Command{TraceInspectCmd, RunInspectCmd} {
		cmd.Flags().IntVar(&inspectLast, "last", 50, "Maximum log lines/events per source")
		cmd.Flags().StringVar(&inspectRedact, "redaction", "safe", "Redaction mode: safe or redacted")
	}
}

func runLogs(cmd *cobra.Command, _ []string) {
	socketPath, _ := cmd.Flags().GetString("socket")
	path, err := logPathForKind(logsKind, socketPath)
	if err != nil {
		cliutil.NewRunnerFromCmd(cmd, true).PrintJSON(map[string]any{"ok": false, "error": err.Error()})
		return
	}
	lines, err := readLogLines(path, logsLast, logsTraceID, logsRunID)
	if err != nil {
		cliutil.NewRunnerFromCmd(cmd, true).PrintJSON(map[string]any{"ok": false, "error": err.Error(), "path": path})
		return
	}
	mode := observability.ParseRedactionMode(logsRedact)
	for i, line := range lines {
		lines[i] = observability.RedactLogLineForDisplay(line, mode)
	}
	if logsFollow {
		followLogLines(path, lines, logsTraceID, logsRunID, mode)
		return
	}
	cliutil.NewRunnerFromCmd(cmd, true).PrintJSON(map[string]any{
		"ok":        true,
		"kind":      logsKind,
		"path":      path,
		"redaction": string(mode),
		"runId":     logsRunID,
		"traceId":   logsTraceID,
		"count":     len(lines),
		"lines":     lines,
	})
}

func followLogLines(path string, initial []string, traceID, runID string, mode observability.RedactionMode) {
	for _, line := range initial {
		fmt.Println(line)
	}
	all, err := readLogLines(path, 0, traceID, runID)
	seen := len(all)
	if err != nil {
		seen = len(initial)
	}
	if logsPoll <= 0 {
		logsPoll = time.Second
	}
	for {
		time.Sleep(logsPoll)
		lines, err := readLogLines(path, 0, traceID, runID)
		if err != nil || len(lines) <= seen {
			continue
		}
		for _, line := range lines[seen:] {
			fmt.Println(observability.RedactLogLineForDisplay(line, mode))
		}
		seen = len(lines)
	}
}

func runAudit(cmd *cobra.Command, _ []string) {
	socketPath, _ := cmd.Flags().GetString("socket")
	readLast := auditLast
	if auditTraceID != "" || auditRunID != "" || auditMethod != "" || auditSince > 0 {
		readLast = 0
	}
	events, err := observability.ReadAudit(socketPath, readLast)
	if err != nil {
		cliutil.NewRunnerFromCmd(cmd, true).PrintJSON(map[string]any{"ok": false, "error": err.Error()})
		return
	}
	events = filterAudit(events)
	events = tailAudit(events, auditLast)
	mode := observability.ParseRedactionMode(auditRedact)
	events = observability.RedactAuditEventsForDisplay(events, mode)
	cliutil.NewRunnerFromCmd(cmd, true).PrintJSON(map[string]any{
		"ok":        true,
		"redaction": string(mode),
		"runId":     auditRunID,
		"traceId":   auditTraceID,
		"count":     len(events),
		"events":    events,
	})
}

func runTraceInspect(cmd *cobra.Command, args []string) {
	inspectBy(cmd, "traceId", args[0])
}

func runRunInspect(cmd *cobra.Command, args []string) {
	inspectBy(cmd, "runId", args[0])
}

func inspectBy(cmd *cobra.Command, key, value string) {
	socketPath, _ := cmd.Flags().GetString("socket")
	mode := observability.ParseRedactionMode(inspectRedact)

	events, err := observability.ReadAudit(socketPath, 0)
	if err != nil {
		cliutil.NewRunnerFromCmd(cmd, true).PrintJSON(map[string]any{"ok": false, "error": err.Error()})
		return
	}
	switch key {
	case "traceId":
		events = auditByTrace(events, value)
	case "runId":
		events = auditByRun(events, value)
	}
	events = tailAudit(events, inspectLast)
	events = observability.RedactAuditEventsForDisplay(events, mode)

	cliLines, cliPath, cliErr := readLinesForInspect("cli", socketPath, key, value)
	serverLines, serverPath, serverErr := readLinesForInspect("server", socketPath, key, value)
	cliutil.NewRunnerFromCmd(cmd, true).PrintJSON(map[string]any{
		"ok":        true,
		key:         value,
		"redaction": string(mode),
		"audit": map[string]any{
			"count":  len(events),
			"events": events,
		},
		"logs": map[string]any{
			"cli": map[string]any{
				"path":  cliPath,
				"error": errorString(cliErr),
				"count": len(cliLines),
				"lines": redactLines(cliLines, mode),
			},
			"server": map[string]any{
				"path":  serverPath,
				"error": errorString(serverErr),
				"count": len(serverLines),
				"lines": redactLines(serverLines, mode),
			},
		},
		"diagnosis":   diagnosisFromEvents(events),
		"nextActions": inspectNextActions(key, value),
	})
}

func readLinesForInspect(kind, socketPath, key, value string) ([]string, string, error) {
	path, err := logPathForKind(kind, socketPath)
	if err != nil {
		return nil, "", err
	}
	traceID, runID := "", ""
	switch key {
	case "traceId":
		traceID = value
	case "runId":
		runID = value
	}
	lines, err := readLogLines(path, inspectLast, traceID, runID)
	return lines, path, err
}

func redactLines(lines []string, mode observability.RedactionMode) []string {
	out := make([]string, len(lines))
	for i, line := range lines {
		out[i] = observability.RedactLogLineForDisplay(line, mode)
	}
	return out
}

func diagnosisFromEvents(events []observability.AuditEvent) map[string]any {
	if len(events) == 0 {
		return map[string]any{"status": "unknown", "summary": "No audit events found."}
	}
	last := events[len(events)-1]
	if last.Status == "error" {
		return map[string]any{
			"status":    "error",
			"method":    last.Method,
			"errorType": last.ErrorType,
			"errorCode": last.ErrorCode,
			"summary":   last.Error,
		}
	}
	return map[string]any{
		"status":  "ok",
		"method":  last.Method,
		"summary": "Latest audited operation completed successfully.",
	}
}

func inspectNextActions(key, value string) []map[string]any {
	if key == "traceId" {
		return []map[string]any{{
			"kind":    "inspect_audit",
			"command": "agent-telegram audit --trace-id " + value + " --agent",
			"safety":  "read",
			"reason":  "view only the redacted audit event for this trace",
		}}
	}
	return []map[string]any{{
		"kind":    "inspect_audit",
		"command": "agent-telegram audit --run-id " + value + " --agent",
		"safety":  "read",
		"reason":  "view the redacted audit timeline for this run",
	}}
}

func errorString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

func logPathForKind(kind, socketPath string) (string, error) {
	switch strings.ToLower(kind) {
	case "cli":
		return paths.CLILogFilePathForSocket(socketPath)
	case "server":
		return paths.LogFilePathForSocket(socketPath)
	default:
		return "", fmt.Errorf("unknown log kind %q", kind)
	}
}

func readLogLines(path string, last int, traceID, runID string) ([]string, error) {
	//nolint:gosec // path is resolved from the local instance socket path.
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer func() { _ = f.Close() }()

	var lines []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if traceID != "" && !strings.Contains(line, traceID) {
			continue
		}
		if runID != "" && !strings.Contains(line, runID) {
			continue
		}
		lines = append(lines, line)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	if last > 0 && len(lines) > last {
		lines = lines[len(lines)-last:]
	}
	return lines, nil
}

func filterAudit(events []observability.AuditEvent) []observability.AuditEvent {
	if auditTraceID == "" && auditRunID == "" && auditMethod == "" && auditSince <= 0 {
		return events
	}
	cutoff := time.Time{}
	if auditSince > 0 {
		cutoff = time.Now().UTC().Add(-auditSince)
	}
	out := make([]observability.AuditEvent, 0, len(events))
	for _, event := range events {
		if auditTraceID != "" && event.TraceID != auditTraceID {
			continue
		}
		if auditRunID != "" && event.RunID != auditRunID {
			continue
		}
		if auditMethod != "" && event.Method != auditMethod {
			continue
		}
		if !cutoff.IsZero() && event.Time.Before(cutoff) {
			continue
		}
		out = append(out, event)
	}
	return out
}

func auditByTrace(events []observability.AuditEvent, traceID string) []observability.AuditEvent {
	out := make([]observability.AuditEvent, 0, len(events))
	for _, event := range events {
		if event.TraceID == traceID {
			out = append(out, event)
		}
	}
	return out
}

func auditByRun(events []observability.AuditEvent, runID string) []observability.AuditEvent {
	out := make([]observability.AuditEvent, 0, len(events))
	for _, event := range events {
		if event.RunID == runID {
			out = append(out, event)
		}
	}
	return out
}

func tailAudit(events []observability.AuditEvent, last int) []observability.AuditEvent {
	if last <= 0 || len(events) <= last {
		return events
	}
	return events[len(events)-last:]
}
