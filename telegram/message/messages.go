// Package message provides Telegram message retrieval operations.
package message

import (
	"context"
	"fmt"
	"strings"
	"unicode"

	"github.com/gotd/td/tg"
	"agent-telegram/telegram/types"
)

// normalizePeer normalizes a peer string for ResolvePeer.
// Passes through numeric IDs, "me"/"self"/"current_user", and @-prefixed usernames.
// Adds @ prefix to bare usernames.
func normalizePeer(peer string) string {
	if peer == "" {
		return peer
	}
	// Already has @ prefix
	if strings.HasPrefix(peer, "@") {
		return peer
	}
	// Special self-referencing peers
	if peer == "me" || peer == "self" || peer == "current_user" {
		return peer
	}
	// Numeric ID (positive or negative)
	if peer[0] == '-' || unicode.IsDigit(rune(peer[0])) {
		return peer
	}
	// Bare username — add @
	return "@" + peer
}

// GetMessages returns messages from a dialog with the given username.
func (c *Client) GetMessages(ctx context.Context, params types.GetMessagesParams) (*types.GetMessagesResult, error) {
	if err := c.CheckInitialized(); err != nil {
		return nil, err
	}

	// Set defaults
	if params.Limit <= 0 {
		params.Limit = 10
	}
	if params.Limit > 100 {
		params.Limit = 100
	}
	if params.Offset < 0 {
		params.Offset = 0
	}

	// Resolve peer (supports @username, numeric ID, me/self/current_user)
	inputPeer, err := c.ResolvePeer(ctx, normalizePeer(params.Username))
	if err != nil {
		return nil, fmt.Errorf("failed to resolve peer %s: %w", params.Username, err)
	}

	// Get messages from the peer
	messagesClass, err := c.API.MessagesGetHistory(ctx, &tg.MessagesGetHistoryRequest{
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
