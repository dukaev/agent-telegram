package chat

import (
	"context"
	"fmt"

	"github.com/gotd/td/tg"
	"agent-telegram/telegram/types"
)

// CreateGroup creates a new group chat.
func (c *Client) CreateGroup(ctx context.Context, params types.CreateGroupParams) (*types.CreateGroupResult, error) {
	if c.API == nil {
		return nil, fmt.Errorf("api client not set")
	}

	// Resolve members to InputPeerClass
	var users []tg.InputUserClass
	for _, member := range params.Members {
		peer, err := c.ResolvePeer(ctx, member)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve member %s: %w", member, err)
		}

		switch p := peer.(type) {
		case *tg.InputPeerUser:
			users = append(users, &tg.InputUser{
				UserID:     p.UserID,
				AccessHash: p.AccessHash,
			})
		default:
			return nil, fmt.Errorf("member %s is not a user", member)
		}
	}

	// Create group
	result, err := c.API.MessagesCreateChat(ctx, &tg.MessagesCreateChatRequest{
		Users: users,
		Title: params.Title,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create group: %w", err)
	}

	groupResult := &types.CreateGroupResult{
		Success: true,
		Title:   params.Title,
	}

	// Extract chat ID from result.Updates
	switch r := result.Updates.(type) {
	case *tg.Updates:
		for _, chat := range r.Chats {
			if ch, ok := chat.(*tg.Chat); ok {
				groupResult.ChatID = ch.ID
			}
		}
	case *tg.UpdatesCombined:
		for _, chat := range r.Chats {
			if ch, ok := chat.(*tg.Chat); ok {
				groupResult.ChatID = ch.ID
			}
		}
	}

	return groupResult, nil
}

// CreateChannel creates a new channel or supergroup.
func (c *Client) CreateChannel(
	ctx context.Context,
	params types.CreateChannelParams,
) (*types.CreateChannelResult, error) {
	if c.API == nil {
		return nil, fmt.Errorf("api client not set")
	}

	// Create channel
	result, err := c.API.ChannelsCreateChannel(ctx, &tg.ChannelsCreateChannelRequest{
		Title:     params.Title,
		About:     params.Description,
		ForImport: false,
		Megagroup: params.Megagroup,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create channel: %w", err)
	}

	channelResult := &types.CreateChannelResult{
		Success: true,
		Title:   params.Title,
	}

	// Extract chat ID from result
	var inputChannel *tg.InputChannel
	switch r := result.(type) {
	case *tg.Updates:
		for _, chat := range r.Chats {
			if ch, ok := chat.(*tg.Channel); ok {
				channelResult.ChatID = ch.ID
				inputChannel = &tg.InputChannel{
					ChannelID:  ch.ID,
					AccessHash: ch.AccessHash,
				}
			}
		}
	case *tg.UpdatesCombined:
		for _, chat := range r.Chats {
			if ch, ok := chat.(*tg.Channel); ok {
				channelResult.ChatID = ch.ID
				inputChannel = &tg.InputChannel{
					ChannelID:  ch.ID,
					AccessHash: ch.AccessHash,
				}
			}
		}
	}

	// Set username if provided
	if params.Username != "" && inputChannel != nil {
		_, err = c.API.ChannelsUpdateUsername(ctx, &tg.ChannelsUpdateUsernameRequest{
			Channel:  inputChannel,
			Username: params.Username,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to set username: %w", err)
		}
	}

	return channelResult, nil
}
