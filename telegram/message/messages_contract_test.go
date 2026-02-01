//go:build contracts

// Package message provides contract tests for message operations.
//
// Run with: go test -tags=contracts ./telegram/message/...
package message

import (
	"encoding/json"
	"testing"

	"agent-telegram/internal/testutil"
)

// TestGetHistory_ParsesPrivateChat verifies that we correctly parse
// a messages.getHistory response for a private chat.
func TestGetHistory_ParsesPrivateChat(t *testing.T) {
	if !testutil.FixtureExists("messages/get_history_private_chat.json") {
		t.Skip("fixture not available")
	}

	fixture := testutil.LoadFixture(t, "messages/get_history_private_chat.json")

	// Verify meta
	if fixture.Meta.Method != "messages.getHistory" {
		t.Errorf("expected method messages.getHistory, got %s", fixture.Meta.Method)
	}

	// Parse response structure
	var response struct {
		Type     string `json:"_"`
		Messages []struct {
			Type    string `json:"_"`
			ID      int    `json:"id"`
			Message string `json:"message"`
			Date    int64  `json:"date"`
			Out     bool   `json:"out"`
			PeerID  struct {
				Type   string `json:"_"`
				UserID int64  `json:"user_id"`
			} `json:"peer_id"`
			FromID struct {
				Type   string `json:"_"`
				UserID int64  `json:"user_id"`
			} `json:"from_id"`
		} `json:"messages"`
		Users []struct {
			Type      string `json:"_"`
			ID        int64  `json:"id"`
			FirstName string `json:"first_name"`
			Username  string `json:"username"`
		} `json:"users"`
	}

	if err := json.Unmarshal(fixture.Response, &response); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	// Verify structure
	if response.Type != "messages.messages" {
		t.Errorf("expected response type messages.messages, got %s", response.Type)
	}

	if len(response.Messages) == 0 {
		t.Error("expected at least one message")
	}

	// Verify message fields
	for i, msg := range response.Messages {
		if msg.ID == 0 {
			t.Errorf("message %d: expected non-zero ID", i)
		}
		if msg.Date == 0 {
			t.Errorf("message %d: expected non-zero date", i)
		}
		if msg.PeerID.Type == "" {
			t.Errorf("message %d: expected peer_id type", i)
		}
	}

	// Verify users
	if len(response.Users) == 0 {
		t.Error("expected at least one user")
	}

	for i, user := range response.Users {
		if user.ID == 0 {
			t.Errorf("user %d: expected non-zero ID", i)
		}
	}
}

// TestGetHistory_MessageOrder verifies messages are in reverse chronological order.
func TestGetHistory_MessageOrder(t *testing.T) {
	if !testutil.FixtureExists("messages/get_history_private_chat.json") {
		t.Skip("fixture not available")
	}

	fixture := testutil.LoadFixture(t, "messages/get_history_private_chat.json")

	var response struct {
		Messages []struct {
			ID   int   `json:"id"`
			Date int64 `json:"date"`
		} `json:"messages"`
	}

	if err := json.Unmarshal(fixture.Response, &response); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	// Messages should be in descending order by ID (newest first)
	for i := 1; i < len(response.Messages); i++ {
		if response.Messages[i].ID >= response.Messages[i-1].ID {
			t.Errorf("messages not in descending order: msg[%d].id=%d >= msg[%d].id=%d",
				i, response.Messages[i].ID, i-1, response.Messages[i-1].ID)
		}
	}
}
