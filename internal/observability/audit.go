package observability

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"agent-telegram/internal/paths"
)

// AuditEvent is a durable, redacted record of one user-facing operation.
type AuditEvent struct {
	Time          time.Time      `json:"time"`
	RunID         string         `json:"runId,omitempty"`
	TraceID       string         `json:"traceId,omitempty"`
	Surface       string         `json:"surface"`
	Method        string         `json:"method,omitempty"`
	Safety        string         `json:"safety,omitempty"`
	DryRun        bool           `json:"dryRun,omitempty"`
	Status        string         `json:"status"`
	DurationMs    int64          `json:"durationMs,omitempty"`
	Params        any            `json:"params,omitempty"`
	ResultSummary map[string]any `json:"resultSummary,omitempty"`
	ErrorCode     int            `json:"errorCode,omitempty"`
	ErrorType     string         `json:"errorType,omitempty"`
	Error         string         `json:"error,omitempty"`
}

// WriteAudit appends an event to the instance-scoped audit journal.
func WriteAudit(socketPath string, event AuditEvent) error {
	if event.Time.IsZero() {
		event.Time = time.Now().UTC()
	}
	event.Params = redactSafeParams(event.Params)
	event.ResultSummary = redactedSummary(event.ResultSummary)
	event.Error = truncate(event.Error, 300)

	path, err := paths.AuditFilePathForSocket(socketPath)
	if err != nil {
		return err
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		return fmt.Errorf("open audit journal: %w", err)
	}
	defer func() { _ = f.Close() }()

	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal audit event: %w", err)
	}
	if _, err := f.Write(append(data, '\n')); err != nil {
		return fmt.Errorf("write audit event: %w", err)
	}
	return nil
}

// ReadAudit reads up to last events from newest to oldest.
func ReadAudit(socketPath string, last int) ([]AuditEvent, error) {
	path, err := paths.AuditFilePathForSocket(socketPath)
	if err != nil {
		return nil, err
	}
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("open audit journal: %w", err)
	}
	defer func() { _ = f.Close() }()

	var events []AuditEvent
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		var event AuditEvent
		if err := json.Unmarshal(scanner.Bytes(), &event); err == nil {
			events = append(events, event)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read audit journal: %w", err)
	}
	if last > 0 && len(events) > last {
		events = events[len(events)-last:]
	}
	return events, nil
}

func redactedSummary(summary map[string]any) map[string]any {
	if summary == nil {
		return nil
	}
	value, ok := RedactAny(summary).(map[string]any)
	if !ok {
		return summary
	}
	return value
}

func truncate(value string, max int) string {
	if max <= 0 || len(value) <= max {
		return value
	}
	return value[:max] + "..."
}
