// Package telegram provides Telegram client video functionality.
package telegram

import (
	"context"
	"fmt"
	"time"

	"github.com/gotd/td/tg"
)

// SendVideo sends a video to a peer.
func (c *Client) SendVideo(ctx context.Context, params SendVideoParams) (*SendVideoResult, error) {
	if c.client == nil {
		return nil, fmt.Errorf("client not initialized")
	}

	api := c.client.API()

	inputPeer, err := resolvePeer(ctx, api, params.Peer)
	if err != nil {
		return nil, err
	}

	uploadedFile, err := uploadFile(ctx, api, params.File)
	if err != nil {
		return nil, err
	}

	media := &tg.InputMediaUploadedDocument{
		File:     uploadedFile,
		MimeType: "video/mp4",
	}

	result, err := api.MessagesSendMedia(ctx, &tg.MessagesSendMediaRequest{
		Peer:     inputPeer,
		Media:    media,
		RandomID: time.Now().UnixNano(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to send video: %w", err)
	}

	msgID := extractMessageID(result)
	return &SendVideoResult{
		ID:      msgID,
		Date:    time.Now().Unix(),
		Peer:    params.Peer,
		Caption: params.Caption,
	}, nil
}
