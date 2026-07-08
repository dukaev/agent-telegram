package types

import "testing"

func TestPeerTypeString(t *testing.T) {
	tests := map[PeerType]string{
		PeerTypeUser:    "user",
		PeerTypeChat:    "chat",
		PeerTypeChannel: "channel",
		PeerType(99):    "unknown",
	}
	for peerType, want := range tests {
		if got := peerType.String(); got != want {
			t.Fatalf("%v.String() = %q, want %q", int(peerType), got, want)
		}
	}
}
