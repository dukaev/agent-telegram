package cliutil

import (
	"encoding/json"
	"testing"
)

func TestPrintSchemaIncludesAgenticMetadata(t *testing.T) {
	output := captureStdout(t, func() {
		if ok := printSchema("send_message"); !ok {
			t.Fatal("printSchema returned false")
		}
	})

	var body struct {
		Method               string         `json:"method"`
		Safety               string         `json:"safety"`
		Idempotent           bool           `json:"idempotent"`
		Retryable            bool           `json:"retryable"`
		RequiresConfirmation bool           `json:"requiresConfirmation"`
		InputSchema          map[string]any `json:"inputSchema"`
		OutputSchema         map[string]any `json:"outputSchema"`
		Schema               map[string]any `json:"schema"`
	}
	if err := json.Unmarshal([]byte(output), &body); err != nil {
		t.Fatal(err)
	}

	if body.Method != "send_message" {
		t.Fatalf("method = %q, want send_message", body.Method)
	}
	if body.Safety != "write" {
		t.Fatalf("safety = %q, want write", body.Safety)
	}
	if body.Idempotent {
		t.Fatal("send_message should not be idempotent")
	}
	if body.Retryable {
		t.Fatal("send_message should not be retryable")
	}
	if body.RequiresConfirmation {
		t.Fatal("send_message should not require confirmation")
	}
	if body.InputSchema["type"] != "object" {
		t.Fatalf("inputSchema.type = %v, want object", body.InputSchema["type"])
	}
	if body.OutputSchema["type"] != "object" {
		t.Fatalf("outputSchema.type = %v, want object", body.OutputSchema["type"])
	}
	if body.Schema["type"] != body.OutputSchema["type"] {
		t.Fatal("legacy schema should mirror outputSchema")
	}
}

func TestPrintSchemaUnknownMethod(t *testing.T) {
	if ok := printSchema("does_not_exist"); ok {
		t.Fatal("printSchema should return false for an unknown method")
	}
}
