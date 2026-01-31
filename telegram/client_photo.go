// Package telegram provides Telegram client photo functionality.
package telegram

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/gotd/td/tg"
	"github.com/gotd/td/telegram/uploader"
)

// SendPhoto sends a photo to a peer.
func (c *Client) SendPhoto(ctx context.Context, params SendPhotoParams) (*SendPhotoResult, error) {
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

	media := &tg.InputMediaUploadedPhoto{File: uploadedFile}

	return sendPhotoMedia(ctx, api, inputPeer, media, params.Peer)
}

// resolvePeer resolves peer from username.
func resolvePeer(ctx context.Context, api *tg.Client, peer string) (tg.InputPeerClass, error) {
	peer = strings.TrimPrefix(peer, "@")

	inputPeer, err := (&Client{}).resolveUsername(ctx, api, peer)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve peer @%s: %w", peer, err)
	}
	return inputPeer, nil
}

// uploadFile uploads a file to Telegram.
func uploadFile(ctx context.Context, api *tg.Client, filePath string) (tg.InputFileClass, error) {
	// #nosec G304 -- filePath is validated in handler
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer func() { _ = file.Close() }()

	fileInfo, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}

	u := uploader.NewUploader(api)
	upload := uploader.NewUpload(fileInfo.Name(), file, fileInfo.Size())

	uploadedFile, err := u.Upload(ctx, upload)
	if err != nil {
		return nil, fmt.Errorf("failed to upload file: %w", err)
	}

	return uploadedFile, nil
}

// sendPhotoMedia sends the uploaded photo media.
func sendPhotoMedia(
	ctx context.Context,
	api *tg.Client,
	inputPeer tg.InputPeerClass,
	media *tg.InputMediaUploadedPhoto,
	peer string,
) (*SendPhotoResult, error) {
	result, err := api.MessagesSendMedia(ctx, &tg.MessagesSendMediaRequest{
		Peer:     inputPeer,
		Media:    media,
		RandomID: time.Now().UnixNano(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to send photo: %w", err)
	}

	msgID := extractMessageID(result)
	return &SendPhotoResult{
		ID:      msgID,
		Date:    time.Now().Unix(),
		Peer:    peer,
		Caption: "", // Caption not supported yet
	}, nil
}

// extractMessageID extracts message ID from Updates response.
func extractMessageID(result tg.UpdatesClass) int64 {
	switch r := result.(type) {
	case *tg.Updates:
		if len(r.Updates) > 0 {
			if msg, ok := r.Updates[0].(*tg.UpdateMessageID); ok {
				return int64(msg.ID)
			}
		}
	case *tg.UpdateShortSentMessage:
		return int64(r.ID)
	}
	return 0
}
