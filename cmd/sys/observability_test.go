package sys

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"agent-telegram/internal/observability"
)

func TestReadLogLinesFiltersTraceAndLast(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cli.log")
	if err := os.WriteFile(path, []byte("one\ntrace-a run-a two\ntrace-a run-b three\nfour\n"), 0600); err != nil {
		t.Fatal(err)
	}

	lines, err := readLogLines(path, 1, "trace-a", "")
	if err != nil {
		t.Fatal(err)
	}
	if len(lines) != 1 || lines[0] != "trace-a run-b three" {
		t.Fatalf("lines = %#v", lines)
	}

	lines, err = readLogLines(path, 10, "", "run-a")
	if err != nil {
		t.Fatal(err)
	}
	if len(lines) != 1 || lines[0] != "trace-a run-a two" {
		t.Fatalf("lines = %#v", lines)
	}
}

func TestFilterAudit(t *testing.T) {
	oldTrace, oldRun, oldMethod, oldSince := auditTraceID, auditRunID, auditMethod, auditSince
	defer func() {
		auditTraceID, auditRunID, auditMethod, auditSince = oldTrace, oldRun, oldMethod, oldSince
	}()

	auditTraceID = "trace-1"
	auditRunID = "run-1"
	auditMethod = "send_message"
	auditSince = time.Hour
	events := []observability.AuditEvent{
		{RunID: "run-1", TraceID: "trace-1", Method: "send_message", Time: time.Now().UTC()},
		{RunID: "run-2", TraceID: "trace-1", Method: "send_message", Time: time.Now().UTC()},
		{RunID: "run-1", TraceID: "trace-2", Method: "send_message", Time: time.Now().UTC()},
		{RunID: "run-1", TraceID: "trace-1", Method: "get_me", Time: time.Now().UTC()},
		{RunID: "run-1", TraceID: "trace-1", Method: "send_message", Time: time.Now().UTC().Add(-2 * time.Hour)},
	}

	filtered := filterAudit(events)
	if len(filtered) != 1 {
		t.Fatalf("filtered len = %d, want 1: %+v", len(filtered), filtered)
	}
}

func TestTailAuditKeepsNewestAfterFiltering(t *testing.T) {
	events := []observability.AuditEvent{
		{TraceID: "trace-1", Method: "send_message"},
		{TraceID: "trace-1", Method: "get_messages"},
		{TraceID: "trace-1", Method: "get_messages"},
	}

	tailed := tailAudit(events, 2)
	if len(tailed) != 2 {
		t.Fatalf("tailed len = %d, want 2", len(tailed))
	}
	if tailed[0].Method != "get_messages" || tailed[1].Method != "get_messages" {
		t.Fatalf("tailed = %+v, want newest events", tailed)
	}
}
