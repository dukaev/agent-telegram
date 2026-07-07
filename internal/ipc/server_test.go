package ipc

import (
	"encoding/json"
	"testing"
)

func TestServerPropagatesTraceID(t *testing.T) {
	srv := NewServer()
	srv.Register("ok", func(_ json.RawMessage) (interface{}, *ErrorObject) {
		return map[string]any{"ok": true}, nil
	})

	resp := srv.handleRequest(&Request{
		JSONRPC: "2.0",
		Method:  "ok",
		ID:      1,
		TraceID: "trace-ipc",
	})

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
