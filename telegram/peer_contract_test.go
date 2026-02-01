//go:build contracts

// Package telegram provides contract tests for peer resolution.
//
// Run with: go test -tags=contracts ./telegram/...
package telegram

import (
	"encoding/json"
	"testing"

	"agent-telegram/internal/testutil"
)

// TestResolveUsername_ParsesUserResponse verifies parsing of resolved user.
func TestResolveUsername_ParsesUserResponse(t *testing.T) {
	if !testutil.FixtureExists("contacts/resolve_username_user.json") {
		t.Skip("fixture not available")
	}

	fixture := testutil.LoadFixture(t, "contacts/resolve_username_user.json")

	var response struct {
		Type string `json:"_"`
		Peer struct {
			Type   string `json:"_"`
			UserID int64  `json:"user_id"`
		} `json:"peer"`
		Users []struct {
			Type       string `json:"_"`
			ID         int64  `json:"id"`
			AccessHash int64  `json:"access_hash"`
			FirstName  string `json:"first_name"`
			Username   string `json:"username"`
		} `json:"users"`
	}

	if err := json.Unmarshal(fixture.Response, &response); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	// Verify response type
	if response.Type != "contacts.resolvedPeer" {
		t.Errorf("expected type contacts.resolvedPeer, got %s", response.Type)
	}

	// Verify peer is user
	if response.Peer.Type != "peerUser" {
		t.Errorf("expected peer type peerUser, got %s", response.Peer.Type)
	}

	// Verify user exists in users array
	if len(response.Users) == 0 {
		t.Fatal("expected at least one user")
	}

	// Verify user ID matches peer
	found := false
	for _, u := range response.Users {
		if u.ID == response.Peer.UserID {
			found = true
			if u.AccessHash == 0 {
				t.Error("expected non-zero access_hash")
			}
			break
		}
	}
	if !found {
		t.Errorf("user with ID %d not found in users array", response.Peer.UserID)
	}
}

// TestResolveUsername_ParsesChannelResponse verifies parsing of resolved channel.
func TestResolveUsername_ParsesChannelResponse(t *testing.T) {
	if !testutil.FixtureExists("contacts/resolve_username_channel.json") {
		t.Skip("fixture not available")
	}

	fixture := testutil.LoadFixture(t, "contacts/resolve_username_channel.json")

	var response struct {
		Type string `json:"_"`
		Peer struct {
			Type      string `json:"_"`
			ChannelID int64  `json:"channel_id"`
		} `json:"peer"`
		Chats []struct {
			Type       string `json:"_"`
			ID         int64  `json:"id"`
			AccessHash int64  `json:"access_hash"`
			Title      string `json:"title"`
			Username   string `json:"username"`
			Broadcast  bool   `json:"broadcast"`
		} `json:"chats"`
	}

	if err := json.Unmarshal(fixture.Response, &response); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	// Verify peer is channel
	if response.Peer.Type != "peerChannel" {
		t.Errorf("expected peer type peerChannel, got %s", response.Peer.Type)
	}

	// Verify channel exists in chats array
	if len(response.Chats) == 0 {
		t.Fatal("expected at least one chat")
	}

	// Verify channel ID matches peer
	found := false
	for _, ch := range response.Chats {
		if ch.ID == response.Peer.ChannelID {
			found = true
			if ch.AccessHash == 0 {
				t.Error("expected non-zero access_hash")
			}
			if ch.Type != "channel" {
				t.Errorf("expected chat type channel, got %s", ch.Type)
			}
			break
		}
	}
	if !found {
		t.Errorf("channel with ID %d not found in chats array", response.Peer.ChannelID)
	}
}

// TestResolveUsername_Variants tests multiple username resolution scenarios.
func TestResolveUsername_Variants(t *testing.T) {
	tests := []struct {
		name         string
		fixture      string
		expectedPeer string
		wantErr      bool
	}{
		{
			name:         "resolve_user",
			fixture:      "contacts/resolve_username_user.json",
			expectedPeer: "peerUser",
			wantErr:      false,
		},
		{
			name:         "resolve_channel",
			fixture:      "contacts/resolve_username_channel.json",
			expectedPeer: "peerChannel",
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !testutil.FixtureExists(tt.fixture) {
				t.Skipf("fixture %s not available", tt.fixture)
			}

			fixture := testutil.LoadFixture(t, tt.fixture)

			var response struct {
				Peer struct {
					Type string `json:"_"`
				} `json:"peer"`
			}

			if err := json.Unmarshal(fixture.Response, &response); err != nil {
				if !tt.wantErr {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}

			if tt.wantErr {
				t.Fatal("expected error but got none")
			}

			if response.Peer.Type != tt.expectedPeer {
				t.Errorf("expected peer type %s, got %s", tt.expectedPeer, response.Peer.Type)
			}
		})
	}
}
