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
	logsRedact  string
	logsFollow  bool
	logsPoll    time.Duration

	auditLast    int
	auditTraceID string
	auditMethod  string
	auditSince   time.Duration
	auditRedact  string
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

// AddObservabilityCommands adds logs and audit commands.
func AddObservabilityCommands(rootCmd *cobra.Command) {
	rootCmd.AddCommand(LogsCmd, AuditCmd)

	LogsCmd.Flags().StringVar(&logsKind, "kind", "cli", "Log kind: cli or server")
	LogsCmd.Flags().IntVar(&logsLast, "last", 50, "Maximum log lines to return")
	LogsCmd.Flags().StringVar(&logsTraceID, "trace-id", "", "Filter log lines by trace ID")
	LogsCmd.Flags().StringVar(&logsRedact, "redaction", "safe", "Redaction mode: safe or redacted")
	LogsCmd.Flags().BoolVarP(&logsFollow, "follow", "f", false, "Follow log output as JSON Lines")
	LogsCmd.Flags().DurationVar(&logsPoll, "interval", time.Second, "Polling interval for --follow")

	AuditCmd.Flags().IntVar(&auditLast, "last", 50, "Maximum audit events to return")
	AuditCmd.Flags().StringVar(&auditTraceID, "trace-id", "", "Filter audit events by trace ID")
	AuditCmd.Flags().StringVar(&auditMethod, "method", "", "Filter audit events by method")
	AuditCmd.Flags().DurationVar(&auditSince, "since", 0, "Filter audit events since duration, e.g. 1h")
	AuditCmd.Flags().StringVar(&auditRedact, "redaction", "safe", "Redaction mode: safe or redacted")
}

func runLogs(cmd *cobra.Command, _ []string) {
	socketPath, _ := cmd.Flags().GetString("socket")
	path, err := logPathForKind(logsKind, socketPath)
	if err != nil {
		cliutil.NewRunnerFromCmd(cmd, true).PrintJSON(map[string]any{"ok": false, "error": err.Error()})
		return
	}
	lines, err := readLogLines(path, logsLast, logsTraceID)
	if err != nil {
		cliutil.NewRunnerFromCmd(cmd, true).PrintJSON(map[string]any{"ok": false, "error": err.Error(), "path": path})
		return
	}
	mode := observability.ParseRedactionMode(logsRedact)
	for i, line := range lines {
		lines[i] = observability.RedactLogLineForDisplay(line, mode)
	}
	if logsFollow {
		followLogLines(path, lines, logsTraceID, mode)
		return
	}
	cliutil.NewRunnerFromCmd(cmd, true).PrintJSON(map[string]any{
		"ok":        true,
		"kind":      logsKind,
		"path":      path,
		"redaction": string(mode),
		"count":     len(lines),
		"lines":     lines,
	})
}

func followLogLines(path string, initial []string, traceID string, mode observability.RedactionMode) {
	for _, line := range initial {
		fmt.Println(line)
	}
	all, err := readLogLines(path, 0, traceID)
	seen := len(all)
	if err != nil {
		seen = len(initial)
	}
	if logsPoll <= 0 {
		logsPoll = time.Second
	}
	for {
		time.Sleep(logsPoll)
		lines, err := readLogLines(path, 0, traceID)
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
	if auditTraceID != "" || auditMethod != "" || auditSince > 0 {
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
		"count":     len(events),
		"events":    events,
	})
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

func readLogLines(path string, last int, traceID string) ([]string, error) {
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
	if auditTraceID == "" && auditMethod == "" && auditSince <= 0 {
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

func tailAudit(events []observability.AuditEvent, last int) []observability.AuditEvent {
	if last <= 0 || len(events) <= last {
		return events
	}
	return events[len(events)-last:]
}
