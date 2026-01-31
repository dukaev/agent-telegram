// Package telegram provides Telegram client file functionality.
package telegram

import (
	"context"
	"fmt"
	"time"

	"github.com/gotd/td/tg"
)

// SendFile sends a file to a peer.
func (c *Client) SendFile(ctx context.Context, params SendFileParams) (*SendFileResult, error) {
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
		MimeType: "application/octet-stream",
	}

	return sendDocumentMedia(ctx, api, inputPeer, media, params.Peer, params.Caption)
}

// sendDocumentMedia sends the uploaded document media.
func sendDocumentMedia(
	ctx context.Context,
	api *tg.Client,
	inputPeer tg.InputPeerClass,
	media *tg.InputMediaUploadedDocument,
	peer string,
	caption string,
) (*SendFileResult, error) {
	result, err := api.MessagesSendMedia(ctx, &tg.MessagesSendMediaRequest{
		Peer:     inputPeer,
		Media:    media,
		RandomID: time.Now().UnixNano(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to send file: %w", err)
	}

	msgID := extractMessageID(result)
	return &SendFileResult{
		ID:      msgID,
		Date:    time.Now().Unix(),
		Peer:    peer,
		Caption: caption,
	}, nil
}
