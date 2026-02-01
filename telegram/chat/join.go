// Package chat provides Telegram chat operations.
package chat

import (
	"context"
	"fmt"

	"github.com/gotd/td/tg"
	"agent-telegram/telegram/types"
)

// Join joins a chat or channel using an invite link.
func (c *Client) Join(ctx context.Context, inviteLink string) (tg.UpdatesClass, error) {
	if c.API == nil {
		return nil, fmt.Errorf("api client not set")
	}

	// Extract hash from invite link
	hash, err := extractInviteHash(inviteLink)
	if err != nil {
		return nil, fmt.Errorf("invalid invite link: %w", err)
	}

	// Use messages.ImportChatInvite to join
	result, err := c.API.MessagesImportChatInvite(ctx, hash)
	if err != nil {
		return nil, fmt.Errorf("failed to join chat: %w", err)
	}

	return result, nil
}

// extractInviteHash extracts the hash from various invite link formats.
func extractInviteHash(link string) (string, error) {
	// Common patterns:
	// https://t.me/+hash
	// https://t.me/joinchat/hash
	// tg://join?invite=hash
	// +hash
	// hash

	// Trim whitespace
	link = trimSpace(link)
	if link == "" {
		return "", fmt.Errorf("empty invite link")
	}

	// Remove common prefixes
	prefixes := []string{
		"https://t.me/+",
		"https://t.me/joinchat/",
		"tg://join?invite=",
		"+",
	}

	for _, prefix := range prefixes {
		if len(link) > len(prefix) && link[:len(prefix)] == prefix {
			return link[len(prefix):], nil
		}
	}

	// Assume the link is already a hash
	return link, nil
}

func trimSpace(s string) string {
	// Simple trim implementation
	start := 0
	end := len(s)
	for start < end && (s[start] == ' ' || s[start] == '\t' || s[start] == '\n' || s[start] == '\r') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t' || s[end-1] == '\n' || s[end-1] == '\r') {
		end--
	}
	return s[start:end]
}

// JoinChat joins a chat or channel using an invite link.
func (c *Client) JoinChat(ctx context.Context, params types.JoinChatParams) (*types.JoinChatResult, error) {
	_, err := c.Join(ctx, params.InviteLink)
	if err != nil {
		return nil, err
	}

	return &types.JoinChatResult{
		Success: true,
	}, nil
}

// Subscribe subscribes to a channel by username.
func (c *Client) Subscribe(ctx context.Context, channel string) (tg.UpdatesClass, error) {
	if c.API == nil {
		return nil, fmt.Errorf("api client not set")
	}

	// Resolve peer to get InputChannel
	peer, err := c.ResolvePeer(ctx, channel)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve channel: %w", err)
	}

	var inputChannel *tg.InputChannel
	switch p := peer.(type) {
	case *tg.InputPeerChannel:
		inputChannel = &tg.InputChannel{
			ChannelID:  p.ChannelID,
			AccessHash: p.AccessHash,
		}
	default:
		return nil, fmt.Errorf("not a channel: %s", channel)
	}

	// Join the channel
	result, err := c.API.ChannelsJoinChannel(ctx, inputChannel)
	if err != nil {
		return nil, fmt.Errorf("failed to subscribe to channel: %w", err)
	}

	return result, nil
}

// SubscribeChannel subscribes to a channel by username.
func (c *Client) SubscribeChannel(
	ctx context.Context,
	params types.SubscribeChannelParams,
) (*types.SubscribeChannelResult, error) {
	updates, err := c.Subscribe(ctx, params.Channel)
	if err != nil {
		return nil, err
	}

	result := &types.SubscribeChannelResult{
		Success: true,
	}

	// Extract channel info from chats
	switch u := updates.(type) {
	case *tg.Updates:
		for _, chat := range u.Chats {
			if ch, ok := chat.(*tg.Channel); ok {
				result.ChatID = ch.ID
				result.Title = ch.Title
			}
		}
	case *tg.UpdatesCombined:
		for _, chat := range u.Chats {
			if ch, ok := chat.(*tg.Channel); ok {
				result.ChatID = ch.ID
				result.Title = ch.Title
			}
		}
	}

	return result, nil
}
