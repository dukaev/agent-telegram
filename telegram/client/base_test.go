package client

import (
	"context"
	"errors"
	"testing"

	"github.com/gotd/td/tg"
)

type fakeParent struct {
	peer   tg.InputPeerClass
	err    error
	cached map[string]tg.InputPeerClass
}

func (f *fakeParent) ResolvePeer(context.Context, string) (tg.InputPeerClass, error) {
	return f.peer, f.err
}

func (f *fakeParent) CachePeer(peer string, inputPeer tg.InputPeerClass) {
	if f.cached == nil {
		f.cached = map[string]tg.InputPeerClass{}
	}
	f.cached[peer] = inputPeer
}

func TestBaseClientInitializationAndParent(t *testing.T) {
	base := &BaseClient{}
	if base.IsInitialized() {
		t.Fatal("zero base should not be initialized")
	}
	if !errors.Is(base.CheckInitialized(), ErrNotInitialized) {
		t.Fatal("CheckInitialized should return ErrNotInitialized")
	}
	if _, err := base.ResolvePeer(context.Background(), "@p"); err == nil {
		t.Fatal("ResolvePeer without parent should fail")
	}

	parent := &fakeParent{peer: &tg.InputPeerSelf{}}
	base.Parent = parent
	peer, err := base.ResolvePeer(context.Background(), "@p")
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := peer.(*tg.InputPeerSelf); !ok {
		t.Fatalf("peer = %T", peer)
	}
	base.CachePeer("@p", peer)
	if parent.cached["@p"] != peer {
		t.Fatal("CachePeer did not delegate to parent")
	}
}
