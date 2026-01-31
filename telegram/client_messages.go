// Package telegram provides Telegram client message functionality.
package telegram

import (
	"context"
	"fmt"
	"strings"

	"github.com/gotd/td/tg"
)

// GetMessagesParams holds parameters for GetMessages.
type GetMessagesParams struct {
	Username string
	Limit    int
	Offset   int
}

// MessageResult represents a single message result.
type MessageResult struct {
	ID       int64  `json:"id"`
	Date     int64  `json:"date"`
	Text     string `json:"text,omitempty"`
	FromID   string `json:"fromId,omitempty"`
	FromName string `json:"fromName,omitempty"`
	Out      bool   `json:"out"`
}

// GetMessagesResult is the result of GetMessages.
type GetMessagesResult struct {
	Messages []MessageResult `json:"messages"`
	Limit    int             `json:"limit"`
	Offset   int             `json:"offset"`
	Count    int             `json:"count"`
	Username string          `json:"username"`
}

// GetMessages returns messages from a dialog with the given username.
func (c *Client) GetMessages(ctx context.Context, params GetMessagesParams) (*GetMessagesResult, error) {
	if c.client == nil {
		return nil, fmt.Errorf("client not initialized")
	}

	api := c.client.API()

	// Clean username (remove @ prefix)
	username := strings.TrimPrefix(params.Username, "@")

	// Resolve username to get input peer
	inputPeer, err := c.resolveUsername(ctx, api, username)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve username @%s: %w", username, err)
	}

	// Get messages from the peer
	messagesClass, err := api.MessagesGetHistory(ctx, &tg.MessagesGetHistoryRequest{
		Peer:      inputPeer,
		Limit:     params.Limit,
		OffsetID:  params.Offset,
		AddOffset: 0,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get messages: %w", err)
	}

	// Extract messages from response
	messages, users := extractMessagesData(messagesClass)
	userMap := buildUserMap(users)

	// Convert to result format
	messageResults := convertMessagesToResult(messages, userMap)

	return &GetMessagesResult{
		Messages: messageResults,
		Limit:    params.Limit,
		Offset:   params.Offset,
		Count:    len(messageResults),
		Username: username,
	}, nil
}

// resolveUsername resolves a username to an InputPeerClass.
func (c *Client) resolveUsername(ctx context.Context, api *tg.Client, username string) (tg.InputPeerClass, error) {
	// Search for the user/channel
	peerClass, err := api.ContactsResolveUsername(ctx, &tg.ContactsResolveUsernameRequest{Username: username})
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
		return nil, fmt.Errorf("unsupported peer type: %T", p)
	}
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
func convertMessagesToResult(messages []tg.MessageClass, userMap map[int64]tg.UserClass) []MessageResult {
	result := make([]MessageResult, 0, len(messages))
	for _, msgClass := range messages {
		msg, ok := msgClass.(*tg.Message)
		if !ok {
			continue
		}

		msgResult := MessageResult{
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
				msgResult.FromID = fmt.Sprintf("user:%d", fromUser.UserID)
				if user, ok := userMap[fromUser.UserID].(*tg.User); ok {
					msgResult.FromName = buildUserDisplayName(user)
				}
			}
		}

		result = append(result, msgResult)
	}
	return result
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
		name = fmt.Sprintf("user_%d", user.ID)
	}
	return name
}
