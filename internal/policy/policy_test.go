package policy

import (
	"context"
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"

	"agent-telegram/internal/ipc"
)

func TestParseValidatesAndNormalizesPolicy(t *testing.T) {
	tests := []struct {
		name    string
		body    string
		wantErr string
	}{
		{name: "valid", body: `{"version":1,"allowPeers":["Ada"]}`},
		{name: "unsupported version", body: `{"version":2}`, wantErr: "unsupported policy version 2"},
		{name: "malformed", body: `{`, wantErr: "parse policy"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse([]byte(tt.body))
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("error = %v, want %q", err, tt.wantErr)
				}
				return
			}
			if err != nil || len(got.AllowPeers) != 1 || got.AllowPeers[0] != "@ada" {
				t.Fatalf("policy = %+v, error = %v", got, err)
			}
		})
	}
}

func TestLoadMissingReturnsDefault(t *testing.T) {
	got, err := Load(filepath.Join(t.TempDir(), "missing.json"))
	if err != nil {
		t.Fatal(err)
	}
	want := Default()
	if got.Version != want.Version || got.Safeties != want.Safeties || got.PeerTypes != want.PeerTypes {
		t.Fatalf("policy = %+v, want default %+v", got, want)
	}
}

func TestEnforcerBlocksDisabledSafety(t *testing.T) {
	p := Default()
	p.Safeties.Write = false
	enforcer := NewEnforcer(p, nil)

	err := enforcer.Check(context.Background(), "send_message", json.RawMessage(`{"peer":"@ada","message":"hi"}`))
	if err == nil {
		t.Fatal("expected policy denial")
	}
	rpcErr := ipc.ErrorObjectFromError(err)
	if rpcErr == nil || rpcErr.Code != ipc.ErrCodePolicyDenied {
		t.Fatalf("rpc error = %+v, want policy denied", rpcErr)
	}
}

func TestEnforcerPeerAllowDenyAndTypes(t *testing.T) {
	p := Default()
	p.AllowPeers = []string{"ada"}
	p.DenyPeers = []string{"@blocked"}
	p.PeerTypes.Bots = false
	enforcer := NewEnforcer(p, nil)

	if err := enforcer.Check(context.Background(), "send_message", json.RawMessage(`{"peer":"@ada","message":"hi"}`)); err != nil {
		t.Fatalf("allowed peer denied: %v", err)
	}
	if err := enforcer.Check(context.Background(), "send_message", json.RawMessage(`{"peer":"@blocked","message":"hi"}`)); err == nil {
		t.Fatal("denied peer should be blocked")
	}
	if err := enforcer.Check(context.Background(), "send_message", json.RawMessage(`{"peer":"@helperbot","message":"hi"}`)); err == nil {
		t.Fatal("bot peer should be blocked")
	}
}

func TestSplitPeerListNormalizesCommonFormats(t *testing.T) {
	got := SplitPeerList("ada, @Grace\nhttps://t.me/TestChannel -10042")
	want := []string{"@ada", "@grace", "@testchannel", "-10042"}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got %v, want %v", got, want)
		}
	}
}

func TestExtractPeersPreservesLargeNumericIDs(t *testing.T) {
	peers := ExtractPeers(json.RawMessage(`{"peer":9007199254740993}`))
	if len(peers) != 1 || peers[0] != "9007199254740993" {
		t.Fatalf("peers = %v", peers)
	}
}

func TestEnforcerRejectsUnknownAndRequiresConfirmation(t *testing.T) {
	enforcer := NewEnforcer(Default(), nil)
	if err := enforcer.Check(context.Background(), "unregistered", nil); err == nil {
		t.Fatal("unregistered operation should be denied")
	}

	if err := enforcer.Check(context.Background(), "logout", nil); err == nil {
		t.Fatal("logout without confirmation should be denied")
	}
	confirmed := ipc.WithConfirmation(context.Background(), true)
	if err := enforcer.Check(confirmed, "logout", nil); err != nil {
		t.Fatalf("confirmed logout denied: %v", err)
	}
}
