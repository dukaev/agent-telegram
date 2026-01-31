// Package message provides Telegram message operations.
package message

import (
	"context"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/tg"
	"agent-telegram/telegram/types"
)

// Client provides message operations.
type Client struct {
	api      *tg.Client
	telegram *telegram.Client
}

// NewClient creates a new message client.
func NewClient(tc *telegram.Client) *Client {
	return &Client{
		telegram: tc,
	}
}

// SetAPI sets the API client (called when the telegram client is initialized).
func (c *Client) SetAPI(api *tg.Client) {
	c.api = api
}

// MessageResult represents a single message result.
type MessageResult struct { // revive:disable:exported // Used internally
	ID       int64  `json:"id"`
	Date     int64  `json:"date"`
	Text     string `json:"text,omitempty"`
	FromID   string `json:"fromId,omitempty"`
	FromName string `json:"fromName,omitempty"`
	Out      bool   `json:"out"`
}

// buildUserDisplayName builds a display name from a user.
func buildUserDisplayName(user *tg.User) string {
	name := user.FirstName
	if user.LastName != "" {
		name += " " + user.LastName
	}
	if name == "" {
		name = user.Username
	}
	if name == "" {
		name = ""
	}
	return name
}

// extractMessageID extracts message ID from response.
func extractMessageID(result tg.UpdatesClass) int64 {
	switch r := result.(type) {
	case *tg.Updates:
		if len(r.Updates) > 0 {
			if msg, ok := r.Updates[0].(*tg.UpdateMessageID); ok {
				return int64(msg.ID)
			}
		}
	case *tg.UpdateShortSentMessage:
		return int64(r.ID)
	}
	return 0
}

// resolvePeer resolves a peer string to InputPeerClass.
func resolvePeer(ctx context.Context, api *tg.Client, peer string) (tg.InputPeerClass, error) {
	// If peer starts with @, it's a username - resolve it
	if len(peer) > 0 && peer[0] == '@' {
		peerClass, err := api.ContactsResolveUsername(ctx, &tg.ContactsResolveUsernameRequest{Username: peer[1:]})
		if err != nil {
			return nil, err
		}

		switch p := peerClass.Peer.(type) {
		case *tg.PeerUser:
			return &tg.InputPeerUser{
				UserID:     p.UserID,
				AccessHash: getAccessHash(peerClass, p.UserID),
			}, nil
		case *tg.PeerChat:
			return &tg.InputPeerChat{
				ChatID: p.ChatID,
			}, nil
		case *tg.PeerChannel:
			return &tg.InputPeerChannel{
				ChannelID:  p.ChannelID,
				AccessHash: getAccessHash(peerClass, p.ChannelID),
			}, nil
		default:
			return nil, nil
		}
	}

	// Try to parse as user ID
	// For now, just return empty peer (will be expanded later)
	return &tg.InputPeerEmpty{}, nil
}

// getAccessHash extracts access hash from the resolved peer.
func getAccessHash(peerClass *tg.ContactsResolvedPeer, id int64) int64 {
	for _, chat := range peerClass.Chats {
		switch c := chat.(type) {
		case *tg.Channel:
			if c.ID == id {
				return c.AccessHash
			}
		case *tg.Chat:
			if c.ID == id {
				return 0
			}
		}
	}
	for _, user := range peerClass.Users {
		if u, ok := user.(*tg.User); ok && u.ID == id {
			return u.AccessHash
		}
	}
	return 0
}

// extractMessagesData extracts messages and users from the response.
func extractMessagesData(messagesClass tg.MessagesMessagesClass) ([]tg.MessageClass, []tg.UserClass) {
	switch m := messagesClass.(type) {
	case *tg.MessagesMessages:
		return m.Messages, m.Users
	case *tg.MessagesMessagesSlice:
		return m.Messages, m.Users
	case *tg.MessagesChannelMessages:
		return m.Messages, m.Users
	default:
		return nil, nil
	}
}

// convertMessagesToResult converts messages to the result format.
func convertMessagesToResult(messages []tg.MessageClass, userMap map[int64]tg.UserClass) []types.MessageResult {
	result := make([]types.MessageResult, 0, len(messages))
	for _, msgClass := range messages {
		msg, ok := msgClass.(*tg.Message)
		if !ok {
			continue
		}

		msgResult := types.MessageResult{
			ID:   int64(msg.ID),
			Date: int64(msg.Date),
			Out:  msg.Out,
		}

		// Extract text
		if msg.Message != "" {
			msgResult.Text = msg.Message
		}

		// Extract sender info
		if msg.FromID != nil {
			if fromUser, ok := msg.FromID.(*tg.PeerUser); ok {
				msgResult.FromID = ""
				if user, ok := userMap[fromUser.UserID].(*tg.User); ok {
					msgResult.FromName = buildUserDisplayName(user)
				}
			}
		}

		result = append(result, msgResult)
	}
	return result
}
