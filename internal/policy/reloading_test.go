package policy

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

func TestReloadingEnforcerAppliesOnlyValidSnapshots(t *testing.T) {
	path := filepath.Join(t.TempDir(), "policy.json")
	writePolicyFile(t, path, `{"version":1,"allowPeers":["@grace"]}`)
	params := json.RawMessage(`{"peer":"@ada","message":"hello"}`)
	enforcer := NewReloadingEnforcer(path, nil)
	if err := enforcer.Check(context.Background(), "send_message", params); err == nil {
		t.Fatal("@ada should initially be denied")
	}

	writePolicyFile(t, path, `{"version":1,"allowPeers":["@ada"]}`)
	if err := enforcer.Check(context.Background(), "send_message", params); err != nil {
		t.Fatalf("valid edit was not applied: %v", err)
	}

	writePolicyFile(t, path, `{"version":1,"allowPeers":[`)
	if err := enforcer.Check(context.Background(), "send_message", params); err != nil {
		t.Fatalf("malformed edit replaced last valid snapshot: %v", err)
	}

	writePolicyFile(t, path, `{"version":1,"allowPeers":["@grace"]}`)
	if err := enforcer.Check(context.Background(), "send_message", params); err == nil {
		t.Fatal("corrected edit was not applied")
	}

	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	if err := enforcer.Check(context.Background(), "send_message", params); err != nil {
		t.Fatalf("deletion should activate defaults: %v", err)
	}
}

func TestReloadingEnforcerDetectsSameSizeRapidReplacement(t *testing.T) {
	path := filepath.Join(t.TempDir(), "policy.json")
	params := json.RawMessage(`{"peer":"@ada","message":"hello"}`)
	writePolicyFile(t, path, `{"version":1,"allowPeers":["@xxxx"]}`)
	enforcer := NewReloadingEnforcer(path, nil)
	if err := enforcer.Check(context.Background(), "send_message", params); err == nil {
		t.Fatal("@ada should initially be denied")
	}

	// The replacement has the same byte length and may share the filesystem's
	// timestamp granularity with the original.
	writePolicyFile(t, path, `{"version":1,"allowPeers":["@adaa"]}`)
	if err := enforcer.Check(context.Background(), "send_message", json.RawMessage(`{"peer":"@adaa"}`)); err != nil {
		t.Fatalf("same-size replacement was not applied: %v", err)
	}
}

func TestReloadingEnforcerRecoversFromUnsupportedVersion(t *testing.T) {
	path := filepath.Join(t.TempDir(), "policy.json")
	params := json.RawMessage(`{"peer":"@ada"}`)
	writePolicyFile(t, path, `{"version":1,"allowPeers":["@ada"]}`)
	enforcer := NewReloadingEnforcer(path, nil)

	writePolicyFile(t, path, `{"version":2,"allowPeers":["@grace"]}`)
	if err := enforcer.Check(context.Background(), "send_message", params); err != nil {
		t.Fatalf("unsupported version replaced last valid snapshot: %v", err)
	}

	writePolicyFile(t, path, `{"version":1,"allowPeers":["@grace"]}`)
	if err := enforcer.Check(context.Background(), "send_message", params); err == nil {
		t.Fatal("valid replacement after unsupported version was not applied")
	}
}

func TestReloadingEnforcerConcurrentChecksDuringRewrites(t *testing.T) {
	path := filepath.Join(t.TempDir(), "policy.json")
	writePolicyFile(t, path, `{"version":1,"allowPeers":["@ada"]}`)
	enforcer := NewReloadingEnforcer(path, nil)
	params := json.RawMessage(`{"peer":"@ada"}`)

	start := make(chan struct{})
	var wg sync.WaitGroup
	for range 32 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			for range 50 {
				_ = enforcer.Check(context.Background(), "send_message", params)
			}
		}()
	}
	close(start)
	for i := range 50 {
		peer := "@ada"
		if i%2 == 1 {
			peer = "@grace"
		}
		writePolicyFile(t, path, `{"version":1,"allowPeers":["`+peer+`"]}`)
	}
	wg.Wait()

	writePolicyFile(t, path, `{"version":1,"allowPeers":["@ada"]}`)
	if err := enforcer.Check(context.Background(), "send_message", params); err != nil {
		t.Fatalf("final valid snapshot was not applied: %v", err)
	}
}

func TestReloadingEnforcerWarnsOncePerFailedDigest(t *testing.T) {
	var logs bytes.Buffer
	previous := slog.Default()
	slog.SetDefault(slog.New(slog.NewJSONHandler(&logs, nil)))
	t.Cleanup(func() { slog.SetDefault(previous) })

	path := filepath.Join(t.TempDir(), "policy.json")
	writePolicyFile(t, path, `{"version":1}`)
	enforcer := NewReloadingEnforcer(path, nil)

	writePolicyFile(t, path, `{`)
	_ = enforcer.Check(context.Background(), "send_message", nil)
	_ = enforcer.Check(context.Background(), "send_message", nil)
	if got := strings.Count(logs.String(), `"msg":"policy reload failed"`); got != 1 {
		t.Fatalf("failure warnings = %d, want 1; logs: %s", got, logs.String())
	}

	writePolicyFile(t, path, `{"version":2}`)
	_ = enforcer.Check(context.Background(), "send_message", nil)
	if got := strings.Count(logs.String(), `"msg":"policy reload failed"`); got != 2 {
		t.Fatalf("failure warnings = %d, want 2; logs: %s", got, logs.String())
	}
}

func writePolicyFile(t *testing.T, path, body string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
}
