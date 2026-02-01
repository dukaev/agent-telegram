package chat

import (
	"context"
	"fmt"

	"github.com/gotd/td/tg"
	"agent-telegram/telegram/types"
)

// GetInviteLink gets or creates an invite link for a chat or channel.
//
//nolint:funlen // Function requires complex invite link retrieval logic
func (c *Client) GetInviteLink(
	ctx context.Context,
	params types.GetInviteLinkParams,
) (*types.GetInviteLinkResult, error) {
	if err := c.CheckInitialized(); err != nil {
		return nil, err
	}

	peer, err := c.ResolvePeer(ctx, params.Peer)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve peer: %w", err)
	}

	switch p := peer.(type) {
	case *tg.InputPeerChannel:
		var result tg.ExportedChatInviteClass
		if params.CreateNew {
			// Create new invite link
			result, err = c.API.MessagesExportChatInvite(ctx, &tg.MessagesExportChatInviteRequest{
				Peer: &tg.InputPeerChannel{
					ChannelID:  p.ChannelID,
					AccessHash: p.AccessHash,
				},
			})
		} else {
			// Try to get existing invite link
			if existing, found := c.tryGetExistingChannelInvite(ctx, p.ChannelID, p.AccessHash); found {
				return existing, nil
			}
			// No existing link, create one
			result, err = c.API.MessagesExportChatInvite(ctx, &tg.MessagesExportChatInviteRequest{
				Peer: &tg.InputPeerChannel{
					ChannelID:  p.ChannelID,
					AccessHash: p.AccessHash,
				},
			})
		}

		if err != nil {
			return nil, fmt.Errorf("failed to get invite link: %w", err)
		}

		if r, ok := result.(*tg.ChatInviteExported); ok {
			return &types.GetInviteLinkResult{
				Link:          r.Link,
				Usage:         r.Usage,
				UsageLimit:    r.UsageLimit,
				RequestNeeded: r.RequestNeeded,
				Expired:       r.Revoked,
			}, nil
		}
	case *tg.InputPeerChat:
		// For chats, use MessagesExportChatInvite
		result, err := c.API.MessagesExportChatInvite(ctx, &tg.MessagesExportChatInviteRequest{
			Peer: p,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to get invite link: %w", err)
		}

		if r, ok := result.(*tg.ChatInviteExported); ok {
			return &types.GetInviteLinkResult{
				Link:       r.Link,
				Usage:      r.Usage,
				UsageLimit: r.UsageLimit,
				Expired:    false,
			}, nil
		}
	}

	return nil, fmt.Errorf("failed to get invite link")
}

// tryGetExistingChannelInvite tries to get existing invite link from channel info.
func (c *Client) tryGetExistingChannelInvite(
	ctx context.Context,
	channelID int64,
	accessHash int64,
) (*types.GetInviteLinkResult, bool) {
	fullInfo, err := c.API.ChannelsGetFullChannel(ctx, &tg.InputChannel{
		ChannelID:  channelID,
		AccessHash: accessHash,
	})
	if err != nil {
		return nil, false
	}

	fullCh, ok := fullInfo.FullChat.(*tg.ChannelFull)
	if !ok || fullCh.ExportedInvite == nil {
		return nil, false
	}

	invite, ok := fullCh.ExportedInvite.(*tg.ChatInviteExported)
	if !ok {
		return nil, false
	}

	return &types.GetInviteLinkResult{
		Link:          invite.Link,
		Usage:         invite.Usage,
		UsageLimit:    invite.UsageLimit,
		RequestNeeded: invite.RequestNeeded,
		Expired:       invite.Revoked,
	}, true
}
