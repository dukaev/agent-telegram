package ipc_test

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"agent-telegram/internal/ipc"
	"agent-telegram/internal/policy"
)

func TestSocketServerUsesReloadedPolicyWithoutRestart(t *testing.T) {
	dir := t.TempDir()
	policyPath := filepath.Join(dir, "policy.json")
	socketDir, err := os.MkdirTemp("/tmp", "at-ipc-")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(socketDir) })
	socketPath := filepath.Join(socketDir, "server.sock")
	if err := os.WriteFile(policyPath, []byte(`{"version":1,"allowPeers":["@grace"]}`), 0o600); err != nil {
		t.Fatal(err)
	}

	srv := ipc.NewSocketServer(socketPath)
	srv.SetPolicyChecker(policy.NewReloadingEnforcer(policyPath, nil))
	srv.Register("send_message", func(context.Context, json.RawMessage) (any, *ipc.ErrorObject) {
		return map[string]any{"id": 1}, nil
	})
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- srv.Start(ctx) }()
	t.Cleanup(func() {
		cancel()
		select {
		case err := <-done:
			if err != nil {
				t.Errorf("server shutdown: %v", err)
			}
		case <-time.After(2 * time.Second):
			t.Error("server did not stop")
		}
	})
	waitForSocket(t, socketPath)

	client := ipc.NewClient(socketPath)
	params := map[string]any{"peer": "@ada", "message": "hello"}
	if _, rpcErr := client.Call("send_message", params); rpcErr == nil || rpcErr.Code != ipc.ErrCodePolicyDenied {
		t.Fatalf("initial error = %+v, want policy denied", rpcErr)
	}
	if err := os.WriteFile(policyPath, []byte(`{"version":1,"allowPeers":["@ada"]}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, rpcErr := client.Call("send_message", params); rpcErr != nil {
		t.Fatalf("request after edit failed without restart: %+v", rpcErr)
	}
}

func waitForSocket(t *testing.T, socketPath string) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	var lastErr error
	for time.Now().Before(deadline) {
		if _, err := os.Stat(socketPath); err == nil {
			return
		} else {
			lastErr = err
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("socket %q was not created: %v", socketPath, lastErr)
}
