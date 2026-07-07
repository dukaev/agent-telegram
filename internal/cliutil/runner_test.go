package cliutil

import (
	"encoding/json"
	"io"
	"os"
	"testing"
)

func TestPrintDryRunReceiptIncludesActionMetadata(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	runner := NewRunner("", true)
	runner.outputFormat = OutputJSON
	runner.receipt = true
	runner.dryRun = true
	runner.traceID = "trace-test"

	output := captureRunnerStdout(t, func() {
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

func captureRunnerStdout(t *testing.T, fn func()) string {
	t.Helper()

	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w
	defer func() { os.Stdout = old }()

	type readResult struct {
		data []byte
		err  error
	}
	readDone := make(chan readResult, 1)
	go func() {
		data, err := io.ReadAll(r)
		readDone <- readResult{data: data, err: err}
	}()

	fn()

	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	result := <-readDone
	if result.err != nil {
		t.Fatal(result.err)
	}
	return string(result.data)
}
