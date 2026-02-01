// Package media provides Telegram photo operations.
package media

import (
	"context"
	"fmt"
	"time"

	"github.com/gotd/td/tg"
	"agent-telegram/telegram/types"
)

// SendPhoto sends a photo to a peer.
func (c *Client) SendPhoto(ctx context.Context, params types.SendPhotoParams) (*types.SendPhotoResult, error) {
	if err := c.CheckInitialized(); err != nil {
		return nil, err
	}

	inputPeer, err := c.ResolvePeer(ctx, params.Peer)
	if err != nil {
		return nil, err
	}

	uploadedFile, err := uploadFile(ctx, c.API, params.File)
	if err != nil {
		return nil, err
	}

	media := &tg.InputMediaUploadedPhoto{File: uploadedFile}

	result, err := c.API.MessagesSendMedia(ctx, &tg.MessagesSendMediaRequest{
		Peer:     inputPeer,
		Media:    media,
		RandomID: time.Now().UnixNano(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to send photo: %w", err)
	}

	msgID := extractMessageID(result)
	return &types.SendPhotoResult{
		ID:      msgID,
		Date:    time.Now().Unix(),
		Peer:    params.Peer,
		Caption: "",
	}, nil
}
