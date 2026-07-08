package cliutil

import (
	"encoding/json"
	"testing"

	"agent-telegram/internal/ipc"
)

func TestPrintDryRunReceiptIncludesActionMetadata(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	runner := NewRunner("", true)
	runner.outputFormat = OutputJSON
	runner.receipt = true
	runner.dryRun = true
	runner.traceID = "trace-test"

	output := captureStdout(t, func() {
		runner.printDryRun("send_message", map[string]any{
			"peer":    "@example",
			"message": "hello",
		})
	})

	var body struct {
		OK      bool   `json:"ok"`
		TraceID string `json:"traceId"`
		Method  string `json:"method"`
		Safety  string `json:"safety"`
		Result  struct {
			DryRun bool   `json:"dry_run"`
			Method string `json:"method"`
		} `json:"result"`
	}
	if err := json.Unmarshal([]byte(output), &body); err != nil {
		t.Fatal(err)
	}
	if !body.OK || body.TraceID != "trace-test" {
		t.Fatalf("receipt metadata = %+v, want ok trace-test", body)
	}
	if body.Method != "send_message" || body.Safety == "" {
		t.Fatalf("receipt action metadata = %+v, want method and safety", body)
	}
	if !body.Result.DryRun || body.Result.Method != "send_message" {
		t.Fatalf("dry-run result = %+v, want send_message dry-run", body.Result)
	}
}

func TestAgentErrorEnvelopeIncludesDiagnosisAndNextActions(t *testing.T) {
	runner := NewRunner("", true)
	runner.agentMode = true
	runner.runID = "run-test"
	runner.traceID = "trace-test"
	runner.lastMethod = "send_message"
	runner.lastSafety = "write"

	output := captureStdout(t, func() {
		runner.printErrorEnvelope(ipc.ErrServerNotRunning)
	})

	var body struct {
		OK      bool   `json:"ok"`
		RunID   string `json:"runId"`
		TraceID string `json:"traceId"`
		Method  string `json:"method"`
		Error   struct {
			Type      string `json:"type"`
			Retryable bool   `json:"retryable"`
		} `json:"error"`
		Diagnosis struct {
			Category string `json:"category"`
		} `json:"diagnosis"`
		NextActions []struct {
			Kind    string `json:"kind"`
			Command string `json:"command"`
		} `json:"nextActions"`
	}
	if err := json.Unmarshal([]byte(output), &body); err != nil {
		t.Fatal(err)
	}
	if body.OK || body.RunID != "run-test" || body.TraceID != "trace-test" || body.Method != "send_message" {
		t.Fatalf("unexpected envelope metadata: %+v", body)
	}
	if body.Error.Type != ipc.ErrorTypeServerNotRunning || !body.Error.Retryable {
		t.Fatalf("unexpected error: %+v", body.Error)
	}
	if body.Diagnosis.Category != "server_not_running" {
		t.Fatalf("diagnosis = %+v, want server_not_running", body.Diagnosis)
	}
	if len(body.NextActions) == 0 || body.NextActions[0].Kind != "start_server" || body.NextActions[0].Command == "" {
		t.Fatalf("nextActions = %+v", body.NextActions)
	}
}
