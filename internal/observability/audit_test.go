package observability

import (
	"testing"
)

func TestWriteAuditRedactsParams(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	err := WriteAudit("", AuditEvent{
		TraceID: "trace-redact",
		Surface: "test",
		Method:  "send_message",
		Status:  "ok",
		Params: map[string]any{
			"phone":    "+88806283792",
			"token":    "secret",
			"password": "secret",
			"message":  "private text",
			"city":     "Tbilisi",
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	events, err := ReadAudit("", 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 1 {
		t.Fatalf("events len = %d, want 1", len(events))
	}
	params := events[0].Params.(map[string]any)
	if params["phone"] != "***3792" {
		t.Fatalf("phone = %v, want masked", params["phone"])
	}
	if params["token"] != redacted || params["password"] != redacted {
		t.Fatalf("params not redacted: %+v", params)
	}
	if params["message"] != "[TEXT REDACTED]" {
		t.Fatalf("message = %v, want safe redaction", params["message"])
	}
	if params["city"] != "[LOCATION REDACTED]" {
		t.Fatalf("city = %v, want safe redaction", params["city"])
	}
}

func TestWriteAuditPreservesPartialTimeoutCorrelation(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	event := AuditEvent{
		RunID:     "run-test",
		TraceID:   "trace-test",
		Surface:   "cli",
		Method:    "send_message",
		Status:    "partial",
		ErrorType: "TIMEOUT",
		ResultSummary: map[string]any{
			"wait": map[string]any{"afterMessageId": 123, "completed": false},
		},
	}
	if err := WriteAudit("", event); err != nil {
		t.Fatal(err)
	}
	events, err := ReadAudit("", 1)
	if err != nil {
		t.Fatal(err)
	}
	got := events[0]
	if got.Status != "partial" || got.ErrorType != "TIMEOUT" || got.RunID != "run-test" || got.TraceID != "trace-test" {
		t.Fatalf("event = %+v", got)
	}
}
