package chat

import (
	"context"
	"fmt"

	"github.com/gotd/td/tg"
	"agent-telegram/telegram/types"
)

// EditTitle edits the title of a chat or channel.
func (c *Client) EditTitle(ctx context.Context, params types.EditTitleParams) (*types.EditTitleResult, error) {
	if c.API == nil {
		return nil, fmt.Errorf("api client not set")
	}

	peer, err := c.ResolvePeer(ctx, params.Peer)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve peer: %w", err)
	}

	switch p := peer.(type) {
	case *tg.InputPeerChannel:
		_, err := c.API.ChannelsEditTitle(ctx, &tg.ChannelsEditTitleRequest{
			Channel: &tg.InputChannel{
				ChannelID:  p.ChannelID,
				AccessHash: p.AccessHash,
			},
			Title: params.Title,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to edit channel title: %w", err)
		}
	case *tg.InputPeerChat:
		_, err := c.API.MessagesEditChatTitle(ctx, &tg.MessagesEditChatTitleRequest{
			ChatID: p.ChatID,
			Title:  params.Title,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to edit chat title: %w", err)
		}
	default:
		return nil, fmt.Errorf("peer must be a chat or channel")
	}

	return &types.EditTitleResult{
		Success: true,
		Title:   params.Title,
	}, nil
}

// SetPhoto sets the photo for a chat or channel.
func (c *Client) SetPhoto(_ context.Context, _ types.SetPhotoParams) (*types.SetPhotoResult, error) {
	if c.API == nil {
		return nil, fmt.Errorf("api client not set")
	}

	// This requires file upload functionality
	// For now, return an error indicating this needs to be implemented
	return nil, fmt.Errorf("set_photo requires file upload - use send_photo command or implement file upload")
}

// DeletePhoto deletes the photo from a chat or channel.
func (c *Client) DeletePhoto(ctx context.Context, params types.DeletePhotoParams) (*types.DeletePhotoResult, error) {
	if c.API == nil {
		return nil, fmt.Errorf("api client not set")
	}

	peer, err := c.ResolvePeer(ctx, params.Peer)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve peer: %w", err)
	}

	switch p := peer.(type) {
	case *tg.InputPeerChannel:
		_, err := c.API.ChannelsEditPhoto(ctx, &tg.ChannelsEditPhotoRequest{
			Channel: &tg.InputChannel{
				ChannelID:  p.ChannelID,
				AccessHash: p.AccessHash,
			},
			Photo: &tg.InputChatPhotoEmpty{},
		})
		if err != nil {
			return nil, fmt.Errorf("failed to delete channel photo: %w", err)
		}
	case *tg.InputPeerChat:
		_, err := c.API.MessagesEditChatPhoto(ctx, &tg.MessagesEditChatPhotoRequest{
			ChatID: p.ChatID,
			Photo:  &tg.InputChatPhotoEmpty{},
		})
		if err != nil {
			return nil, fmt.Errorf("failed to delete chat photo: %w", err)
		}
	default:
		return nil, fmt.Errorf("peer must be a chat or channel")
	}

	return &types.DeletePhotoResult{Success: true}, nil
}
