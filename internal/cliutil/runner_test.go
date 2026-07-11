package cliutil

import (
	"encoding/json"
	"strings"
	"testing"

	"agent-telegram/internal/ipc"
	"agent-telegram/internal/observability"
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

func TestFailTypedTimeoutProducesPartialEnvelopeAndAudit(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	SetExitFunc(func(code int) {
		if code != 1 {
			t.Errorf("exit code = %d, want 1", code)
		}
	})
	t.Cleanup(func() { SetExitFunc(nil) })

	runner := NewRunner("", true)
	runner.agentMode = true
	runner.runID = "run-test"
	runner.traceID = "trace-test"
	runner.SetAction("send_message")
	runner.writeAudit("send_message", map[string]any{"peer": "@bot", "message": "must not appear"}, map[string]any{"id": 123}, nil, 0)
	err := ipc.NewTypedError(ipc.ErrCodeTimeout, ipc.ErrorTypeTimeout, "no reply within 20s", map[string]any{
		"actionSucceeded": true,
	})
	details := FailureDetails{
		PartialResult: map[string]any{
			"action": map[string]any{"id": 123},
			"wait": map[string]any{
				"afterMessageId": 123,
				"completed":      false,
			},
		},
		NextActions: []map[string]any{{"kind": "wait_for_reply", "command": "agent-telegram msg wait @bot"}},
		AuditStatus: "partial",
		AuditSummary: map[string]any{
			"wait": map[string]any{"afterMessageId": 123, "completed": false},
		},
	}

	output := captureStdout(t, func() { runner.FailTyped(err, details) })
	var body struct {
		RunID   string `json:"runId"`
		TraceID string `json:"traceId"`
		Error   struct {
			Type      string `json:"type"`
			Retryable bool   `json:"retryable"`
			Data      struct {
				ActionSucceeded bool `json:"actionSucceeded"`
			} `json:"data"`
		} `json:"error"`
		PartialResult struct {
			Wait struct {
				Completed bool `json:"completed"`
			} `json:"wait"`
		} `json:"partialResult"`
	}
	if err := json.Unmarshal([]byte(output), &body); err != nil {
		t.Fatal(err)
	}
	if body.Error.Type != ipc.ErrorTypeTimeout || !body.Error.Retryable || !body.Error.Data.ActionSucceeded {
		t.Fatalf("error = %+v, want retryable TIMEOUT after successful action", body.Error)
	}
	if body.PartialResult.Wait.Completed {
		t.Fatalf("partial timeout = %+v", body.PartialResult)
	}
	if body.RunID != "run-test" || body.TraceID != "trace-test" {
		t.Fatalf("correlation IDs = %+v", body)
	}

	events, readErr := observability.ReadAudit("", 10)
	if readErr != nil {
		t.Fatal(readErr)
	}
	if len(events) != 2 || events[0].Status != "ok" || events[0].RunID != "run-test" || events[0].TraceID != "trace-test" {
		t.Fatalf("successful action audit was not preserved: %+v", events)
	}
	params, _ := events[0].Params.(map[string]any)
	if params["message"] != "[TEXT REDACTED]" {
		t.Fatalf("message text leaked into audit: %+v", params)
	}
	last := events[len(events)-1]
	if last.Status != "partial" || last.ErrorType != ipc.ErrorTypeTimeout || last.RunID != "run-test" || last.TraceID != "trace-test" {
		t.Fatalf("audit = %+v", last)
	}
}

func TestFailTypedStandaloneWaitOmitsPartialResultAndActionSucceeded(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	SetExitFunc(func(int) {})
	t.Cleanup(func() { SetExitFunc(nil) })
	runner := NewRunner("", true)
	runner.agentMode = true
	runner.SetAction("wait_for_reply")
	err := ipc.NewTypedError(ipc.ErrCodeTimeout, ipc.ErrorTypeTimeout, "no reply", map[string]any{})

	output := captureStdout(t, func() {
		runner.FailTyped(err, FailureDetails{AuditStatus: "error"})
	})
	if strings.Contains(output, "partialResult") || strings.Contains(output, "actionSucceeded") {
		t.Fatalf("standalone wait contains action-only fields: %s", output)
	}
	events, readErr := observability.ReadAudit("", 10)
	if readErr != nil {
		t.Fatal(readErr)
	}
	if last := events[len(events)-1]; last.Status != "error" || last.ErrorType != ipc.ErrorTypeTimeout {
		t.Fatalf("audit = %+v", last)
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
