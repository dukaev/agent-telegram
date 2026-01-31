// Package message provides Telegram message retrieval operations.
package message

import (
	"context"
	"fmt"
	"strings"

	"github.com/gotd/td/tg"
	"agent-telegram/telegram/types"
)

// GetMessages returns messages from a dialog with the given username.
func (c *Client) GetMessages(ctx context.Context, params types.GetMessagesParams) (*types.GetMessagesResult, error) {
	if c.api == nil {
		return nil, fmt.Errorf("client not initialized")
	}

	// Ensure username has @ prefix for resolvePeer
	username := params.Username
	if !strings.HasPrefix(username, "@") {
		username = "@" + username
	}

	// Resolve username to get input peer
	inputPeer, err := c.resolvePeer(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve username @%s: %w", username, err)
	}

	// Get messages from the peer
	messagesClass, err := c.api.MessagesGetHistory(ctx, &tg.MessagesGetHistoryRequest{
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

	// Build user map
	userMap := make(map[int64]tg.UserClass)
	for _, u := range users {
		if user, ok := u.(*tg.User); ok {
			userMap[user.ID] = user
		}
	}

	// Convert to result format
	messageResults := convertMessagesToResult(messages, userMap)

	return &types.GetMessagesResult{
		Messages: messageResults,
		Limit:    params.Limit,
		Offset:   params.Offset,
		Count:    len(messageResults),
		Username: strings.TrimPrefix(params.Username, "@"),
	}, nil
}
