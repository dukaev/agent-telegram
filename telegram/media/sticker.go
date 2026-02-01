// Package media provides sticker operations.
package media

import (
	"context"
	"fmt"
	"time"

	"github.com/gotd/td/tg"
	"agent-telegram/telegram/types"
)

// SendSticker sends a sticker to a peer.
func (c *Client) SendSticker(ctx context.Context, params types.SendStickerParams) (*types.SendStickerResult, error) {
	if err := c.CheckInitialized(); err != nil {
		return nil, err
	}

	inputPeer, err := c.ResolvePeer(ctx, params.Peer)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve peer: %w", err)
	}

	var media tg.InputMediaClass

	if params.File != "" {
		// Upload sticker file
		uploadedFile, err := uploadFile(ctx, c.API, params.File)
		if err != nil {
			return nil, fmt.Errorf("failed to upload sticker: %w", err)
		}

		attributes := []tg.DocumentAttributeClass{
			&tg.DocumentAttributeSticker{
				Stickerset: &tg.InputStickerSetEmpty{},
			},
		}

		media = &tg.InputMediaUploadedDocument{
			File:       uploadedFile,
			MimeType:   "image/webp",
			Attributes: attributes,
		}
	} else {
		return nil, fmt.Errorf("sticker file is required")
	}

	result, err := c.API.MessagesSendMedia(ctx, &tg.MessagesSendMediaRequest{
		Peer:     inputPeer,
		Media:    media,
		RandomID: time.Now().UnixNano(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to send sticker: %w", err)
	}

	msgID := extractMessageID(result)
	return &types.SendStickerResult{
		ID:   msgID,
		Date: time.Now().Unix(),
		Peer: params.Peer,
	}, nil
}

// GetStickerPacks retrieves all sticker packs.
func (c *Client) GetStickerPacks(
	ctx context.Context,
	_ types.GetStickerPacksParams,
) (*types.GetStickerPacksResult, error) {
	if err := c.CheckInitialized(); err != nil {
		return nil, err
	}

	result, err := c.API.MessagesGetAllStickers(ctx, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get sticker packs: %w", err)
	}

	packs := make([]types.StickerPack, 0)

	if stickers, ok := result.(*tg.MessagesAllStickers); ok {
		for _, set := range stickers.Sets {
			packs = append(packs, types.StickerPack{
				ID:        set.ID,
				Title:     set.Title,
				ShortName: set.ShortName,
				Count:     set.Count,
			})
		}
	}

	return &types.GetStickerPacksResult{
		Packs: packs,
		Count: len(packs),
	}, nil
}

// SendGIF sends a GIF/animation to a peer.
func (c *Client) SendGIF(ctx context.Context, params types.SendGIFParams) (*types.SendGIFResult, error) {
	if err := c.CheckInitialized(); err != nil {
		return nil, err
	}

	inputPeer, err := c.ResolvePeer(ctx, params.Peer)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve peer: %w", err)
	}

	uploadedFile, err := uploadFile(ctx, c.API, params.File)
	if err != nil {
		return nil, fmt.Errorf("failed to upload GIF: %w", err)
	}

	// GIF/animation attributes
	attributes := []tg.DocumentAttributeClass{
		&tg.DocumentAttributeAnimated{},
	}

	media := &tg.InputMediaUploadedDocument{
		File:       uploadedFile,
		MimeType:   "video/mp4", // Modern GIFs are MP4
		Attributes: attributes,
	}

	result, err := c.API.MessagesSendMedia(ctx, &tg.MessagesSendMediaRequest{
		Peer:     inputPeer,
		Media:    media,
		Message:  params.Caption,
		RandomID: time.Now().UnixNano(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to send GIF: %w", err)
	}

	msgID := extractMessageID(result)
	return &types.SendGIFResult{
		ID:      msgID,
		Date:    time.Now().Unix(),
		Peer:    params.Peer,
		Caption: params.Caption,
	}, nil
}
