package policy

import (
	"context"
	"encoding/json"
	"testing"

	"agent-telegram/internal/ipc"
)

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
