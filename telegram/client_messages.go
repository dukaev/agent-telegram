// Package telegram provides Telegram client message functionality.
package telegram

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gotd/td/tg"
)

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

// SendMessage sends a message to a peer.
func (c *Client) SendMessage(ctx context.Context, params SendMessageParams) (*SendMessageResult, error) {
	if c.client == nil {
		return nil, fmt.Errorf("client not initialized")
	}

	api := c.client.API()

	// Clean peer (remove @ prefix)
	peer := strings.TrimPrefix(params.Peer, "@")

	// Resolve username to get input peer
	inputPeer, err := c.resolveUsername(ctx, api, peer)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve peer @%s: %w", peer, err)
	}

	// Send message
	result, err := api.MessagesSendMessage(ctx, &tg.MessagesSendMessageRequest{
		Peer:    inputPeer,
		Message: params.Message,
		RandomID: time.Now().UnixNano(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to send message: %w", err)
	}

	// Extract message ID from response
	var msgID int64
	switch r := result.(type) {
	case *tg.Updates:
		if len(r.Updates) > 0 {
			if msg, ok := r.Updates[0].(*tg.UpdateMessageID); ok {
				msgID = int64(msg.ID)
			}
		}
	case *tg.UpdateShortSentMessage:
		msgID = int64(r.ID)
	}

	return &SendMessageResult{
		ID:      msgID,
		Date:    time.Now().Unix(),
		Message: params.Message,
		Peer:    peer,
	}, nil
}
