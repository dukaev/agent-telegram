package chat

import (
	"context"
	"fmt"

	"github.com/gotd/td/tg"
	"agent-telegram/telegram/types"
)

// Leave leaves a chat or channel.
func (c *Client) Leave(ctx context.Context, params types.LeaveParams) (*types.LeaveResult, error) {
	if err := c.CheckInitialized(); err != nil {
		return nil, err
	}

	peer, err := c.ResolvePeer(ctx, params.Peer)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve peer: %w", err)
	}

	switch p := peer.(type) {
	case *tg.InputPeerChannel:
		_, err := c.API.ChannelsLeaveChannel(ctx, &tg.InputChannel{
			ChannelID:  p.ChannelID,
			AccessHash: p.AccessHash,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to leave channel: %w", err)
		}
	case *tg.InputPeerChat:
		_, err := c.API.MessagesDeleteChatUser(ctx, &tg.MessagesDeleteChatUserRequest{
			ChatID: p.ChatID,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to leave chat: %w", err)
		}
	default:
		return nil, fmt.Errorf("peer must be a chat or channel")
	}

	return &types.LeaveResult{Success: true}, nil
}

// Invite invites users to a chat or channel.
func (c *Client) Invite(ctx context.Context, params types.InviteParams) (*types.InviteResult, error) {
	if err := c.CheckInitialized(); err != nil {
		return nil, err
	}

	peer, err := c.ResolvePeer(ctx, params.Peer)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve peer: %w", err)
	}

	// Resolve all members to add
	for _, member := range params.Members {
		userPeer, err := c.ResolvePeer(ctx, member)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve member %s: %w", member, err)
		}

		var inputUser *tg.InputUser
		switch p := userPeer.(type) {
		case *tg.InputPeerUser:
			inputUser = &tg.InputUser{
				UserID:     p.UserID,
				AccessHash: p.AccessHash,
			}
		default:
			return nil, fmt.Errorf("member %s is not a user", member)
		}

		switch p := peer.(type) {
		case *tg.InputPeerChannel:
			_, err := c.API.ChannelsInviteToChannel(ctx, &tg.ChannelsInviteToChannelRequest{
				Channel: &tg.InputChannel{
					ChannelID:  p.ChannelID,
					AccessHash: p.AccessHash,
				},
				Users: []tg.InputUserClass{inputUser},
			})
			if err != nil {
				return nil, fmt.Errorf("failed to invite to channel: %w", err)
			}
		case *tg.InputPeerChat:
			_, err := c.API.MessagesAddChatUser(ctx, &tg.MessagesAddChatUserRequest{
				ChatID: p.ChatID,
				UserID: inputUser,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to invite to chat: %w", err)
			}
		}
	}

	return &types.InviteResult{Success: true}, nil
}
