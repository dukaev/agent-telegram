// Package media provides Telegram document operations (file, video).
package media

import (
	"context"
	"fmt"
	"time"

	"github.com/gotd/td/tg"
	"agent-telegram/telegram/types"
)

// SendDocument sends a document to a peer with custom mime type.
func (c *Client) SendDocument(
	ctx context.Context, peer, file, mimeType, caption string,
) (*types.SendFileResult, error) {
	if c.api == nil {
		return nil, fmt.Errorf("client not initialized")
	}

	inputPeer, err := c.resolvePeer(ctx, peer)
	if err != nil {
		return nil, err
	}

	uploadedFile, err := uploadFile(ctx, c.api, file)
	if err != nil {
		return nil, err
	}

	media := &tg.InputMediaUploadedDocument{
		File:     uploadedFile,
		MimeType: mimeType,
	}

	result, err := c.api.MessagesSendMedia(ctx, &tg.MessagesSendMediaRequest{
		Peer:     inputPeer,
		Media:    media,
		RandomID: time.Now().UnixNano(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to send document: %w", err)
	}

	msgID := extractMessageID(result)
	return &types.SendFileResult{
		ID:      msgID,
		Date:    time.Now().Unix(),
		Peer:    peer,
		Caption: caption,
	}, nil
}

// SendFile sends a file to a peer.
func (c *Client) SendFile(ctx context.Context, params types.SendFileParams) (*types.SendFileResult, error) {
	return c.SendDocument(ctx, params.Peer, params.File, "application/octet-stream", params.Caption)
}

// SendVideo sends a video to a peer.
func (c *Client) SendVideo(ctx context.Context, params types.SendVideoParams) (*types.SendVideoResult, error) {
	fileResult, err := c.SendDocument(ctx, params.Peer, params.File, "video/mp4", params.Caption)
	if err != nil {
		return nil, err
	}

	return &types.SendVideoResult{
		ID:      fileResult.ID,
		Date:    fileResult.Date,
		Peer:    params.Peer,
		Caption: params.Caption,
	}, nil
}
