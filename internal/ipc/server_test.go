package ipc

import (
	"context"
	"encoding/json"
	"testing"
)

type testPolicyChecker struct {
	calls int
	err   error
}

func (p *testPolicyChecker) Check(_ context.Context, _ string, _ json.RawMessage) error {
	p.calls++
	return p.err
}

func TestServerPropagatesTraceID(t *testing.T) {
	srv := NewServer()
	srv.Register("ok", func(_ json.RawMessage) (interface{}, *ErrorObject) {
		return map[string]any{"ok": true}, nil
	})

	resp := srv.handleRequest(&Request{
		JSONRPC: "2.0",
		Method:  "ok",
		ID:      1,
		RunID:   "run-ipc",
		TraceID: "trace-ipc",
	})

	if resp.RunID != "run-ipc" {
		t.Fatalf("run id = %q, want run-ipc", resp.RunID)
	}
	if resp.TraceID != "trace-ipc" {
		t.Fatalf("trace id = %q, want trace-ipc", resp.TraceID)
	}
	if resp.Error != nil {
		t.Fatalf("unexpected error: %+v", resp.Error)
	}
}

func TestServerMethodNotFoundIsTyped(t *testing.T) {
	resp := NewServer().handleRequest(&Request{
		JSONRPC: "2.0",
		Method:  "missing",
		ID:      1,
		TraceID: "trace-ipc",
	})

	if resp.TraceID != "trace-ipc" {
		t.Fatalf("trace id = %q, want trace-ipc", resp.TraceID)
	}
	if resp.Error == nil {
		t.Fatal("expected error")
	}
	data, ok := resp.Error.Data.(map[string]any)
	if !ok {
		t.Fatalf("error data type = %T, want map", resp.Error.Data)
	}
	if data["type"] != ErrorTypeMethodNotFound {
		t.Fatalf("error type = %v, want %s", data["type"], ErrorTypeMethodNotFound)
	}
}

func TestServerPolicyCheckerBlocksBeforeHandler(t *testing.T) {
	checker := &testPolicyChecker{
		err: NewPolicyDeniedError("send_message", "write operations are disabled"),
	}
	srv := NewServer()
	srv.SetPolicyChecker(checker)
	called := false
	srv.Register("send_message", func(_ json.RawMessage) (interface{}, *ErrorObject) {
		called = true
		return map[string]any{"ok": true}, nil
	})

	resp := srv.handleRequest(&Request{
		JSONRPC: "2.0",
		Method:  "send_message",
		Params:  json.RawMessage(`{"peer":"@ada","message":"hi"}`),
		ID:      1,
	})

	if called {
		t.Fatal("handler should not execute when policy denies")
	}
	if checker.calls != 1 {
		t.Fatalf("policy calls = %d, want 1", checker.calls)
	}
	if resp.Error == nil || resp.Error.Code != ErrCodePolicyDenied {
		t.Fatalf("error = %+v, want policy denied", resp.Error)
	}
}
