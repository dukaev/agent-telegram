package telegram

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/gotd/td/tg"
	"agent-telegram/telegram/helpers"
)

// ResolvePeer resolves a peer string to InputPeerClass with caching.
// This method is shared across all domain clients to avoid duplicate API calls.
// Supported formats:
//   - @username - resolves via API
//   - 123456789 - positive number = user ID
//   - -123456789 - negative number = chat ID
//   - -100XXXXXXXXX - channel ID (common Telegram format)
func (c *Client) ResolvePeer(ctx context.Context, peer string) (tg.InputPeerClass, error) {
	// Check cache first
	if cached, ok := c.peerCache.Load(peer); ok {
		if inputPeer, ok := cached.(tg.InputPeerClass); ok {
			return inputPeer, nil
		}
	}

	var inputPeer tg.InputPeerClass
	var err error

	// "me", "self", "current_user" resolves to InputPeerSelf (Saved Messages)
	if peer == "me" || peer == "self" || peer == "current_user" {
		inputPeer = &tg.InputPeerSelf{}
		c.peerCache.Store(peer, inputPeer)
		return inputPeer, nil
	}

	// If peer starts with @, it's a username - resolve it with cache
	if len(peer) > 0 && peer[0] == '@' {
		inputPeer, err = c.resolveUsername(ctx, peer[1:])
	} else {
		// Try to parse as numeric ID
		inputPeer, err = c.resolveNumericPeer(ctx, peer)
	}

	if err != nil {
		return nil, err
	}

	// Store in cache
	c.peerCache.Store(peer, inputPeer)
	return inputPeer, nil
}

// resolveUsername resolves a username via Telegram API.
func (c *Client) resolveUsername(ctx context.Context, username string) (tg.InputPeerClass, error) {
	peerClass, err := c.client.API().ContactsResolveUsername(ctx, &tg.ContactsResolveUsernameRequest{Username: username})
	if err != nil {
		return nil, err
	}

	switch p := peerClass.Peer.(type) {
	case *tg.PeerUser:
		return &tg.InputPeerUser{
			UserID:     p.UserID,
			AccessHash: helpers.GetAccessHash(peerClass, p.UserID),
		}, nil
	case *tg.PeerChat:
		return &tg.InputPeerChat{
			ChatID: p.ChatID,
		}, nil
	case *tg.PeerChannel:
		return &tg.InputPeerChannel{
			ChannelID:  p.ChannelID,
			AccessHash: helpers.GetAccessHash(peerClass, p.ChannelID),
		}, nil
	default:
		return nil, fmt.Errorf("unknown peer type")
	}
}

// resolveNumericPeer resolves a numeric peer ID by looking up in dialogs.
func (c *Client) resolveNumericPeer(ctx context.Context, peer string) (tg.InputPeerClass, error) {
	id, err := strconv.ParseInt(peer, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid peer format: %s", peer)
	}

	// Determine peer type based on ID format
	if id > 0 {
		// Positive = user ID, need to find access hash from dialogs
		return c.resolveUserByID(ctx, id)
	}

	// Negative ID
	absID := -id

	// Check if it's a channel (starts with 100)
	if strings.HasPrefix(peer, "-100") && len(peer) > 4 {
		// Channel ID format: -100XXXXXXXXX
		channelID := absID - 1000000000000
		if channelID > 0 {
			return c.resolveChannelByID(ctx, channelID)
		}
	}

	// Regular chat (basic group)
	return &tg.InputPeerChat{ChatID: absID}, nil
}

// resolveUserByID finds a user by ID from dialogs.
func (c *Client) resolveUserByID(ctx context.Context, userID int64) (tg.InputPeerClass, error) {
	// Get dialogs to find the user with access hash
	result, err := c.client.API().MessagesGetDialogs(ctx, &tg.MessagesGetDialogsRequest{
		OffsetPeer: &tg.InputPeerEmpty{},
		Limit:      100,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get dialogs: %w", err)
	}

	// Extract users from dialogs
	var users []tg.UserClass
	switch r := result.(type) {
	case *tg.MessagesDialogs:
		users = r.Users
	case *tg.MessagesDialogsSlice:
		users = r.Users
	}

	// Find the user
	for _, u := range users {
		if user, ok := u.(*tg.User); ok && user.ID == userID {
			return &tg.InputPeerUser{
				UserID:     user.ID,
				AccessHash: user.AccessHash,
			}, nil
		}
	}

	return nil, fmt.Errorf("user %d not found in dialogs", userID)
}

// resolveChannelByID finds a channel by ID from dialogs.
func (c *Client) resolveChannelByID(ctx context.Context, channelID int64) (tg.InputPeerClass, error) {
	// Get dialogs to find the channel with access hash
	result, err := c.client.API().MessagesGetDialogs(ctx, &tg.MessagesGetDialogsRequest{
		OffsetPeer: &tg.InputPeerEmpty{},
		Limit:      100,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get dialogs: %w", err)
	}

	// Extract chats from dialogs
	var chats []tg.ChatClass
	switch r := result.(type) {
	case *tg.MessagesDialogs:
		chats = r.Chats
	case *tg.MessagesDialogsSlice:
		chats = r.Chats
	}

	// Find the channel
	for _, ch := range chats {
		if channel, ok := ch.(*tg.Channel); ok && channel.ID == channelID {
			return &tg.InputPeerChannel{
				ChannelID:  channel.ID,
				AccessHash: channel.AccessHash,
			}, nil
		}
	}

	return nil, fmt.Errorf("channel %d not found in dialogs", channelID)
}

