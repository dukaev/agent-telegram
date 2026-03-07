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

// GetMessage returns a single message by ID from a chat.
func (c *Client) GetMessage(ctx context.Context, params types.GetMessageParams) (*types.GetMessageResult, error) {
	if err := c.CheckInitialized(); err != nil {
		return nil, err
	}

	peer := params.Peer
	if peer == "" {
		peer = params.Username
	}

	inputPeer, err := c.ResolvePeer(ctx, normalizePeer(peer))
	if err != nil {
		return nil, fmt.Errorf("failed to resolve peer %s: %w", peer, err)
	}

	msgIDs := []tg.InputMessageClass{&tg.InputMessageID{ID: int(params.MessageID)}}

	var messagesClass tg.MessagesMessagesClass
	if ch, ok := inputPeer.(*tg.InputPeerChannel); ok {
		messagesClass, err = c.API.ChannelsGetMessages(ctx, &tg.ChannelsGetMessagesRequest{
			Channel: &tg.InputChannel{ChannelID: ch.ChannelID, AccessHash: ch.AccessHash},
			ID:      msgIDs,
		})
	} else {
		messagesClass, err = c.API.MessagesGetMessages(ctx, msgIDs)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get message: %w", err)
	}

	messages, users := extractMessagesData(messagesClass)
	if len(messages) == 0 {
		return nil, fmt.Errorf("message %d not found", params.MessageID)
	}

	userMap := make(map[int64]tg.UserClass)
	for _, u := range users {
		if user, ok := u.(*tg.User); ok {
			userMap[user.ID] = user
		}
	}

	results := convertMessagesToResult(messages, userMap)
	if len(results) == 0 {
		return nil, fmt.Errorf("message %d not found", params.MessageID)
	}

	return &types.GetMessageResult{
		Message: results[0],
	}, nil
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
