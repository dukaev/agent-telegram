// Package media provides voice message operations.
package media

import (
	"context"
	"fmt"
	"time"

	"github.com/gotd/td/tg"
	"agent-telegram/telegram/types"
)

// SendVoice sends a voice message to a peer.
func (c *Client) SendVoice(ctx context.Context, params types.SendVoiceParams) (*types.SendVoiceResult, error) {
	if err := c.CheckInitialized(); err != nil {
		return nil, err
	}

	inputPeer, err := c.ResolvePeer(ctx, params.Peer)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve peer: %w", err)
	}

	uploadedFile, err := uploadFile(ctx, c.API, params.File)
	if err != nil {
		return nil, fmt.Errorf("failed to upload voice: %w", err)
	}

	// Voice message attributes
	attributes := []tg.DocumentAttributeClass{
		&tg.DocumentAttributeAudio{
			Voice:    true,
			Duration: params.Duration,
		},
	}

	media := &tg.InputMediaUploadedDocument{
		File:       uploadedFile,
		MimeType:   "audio/ogg",
		Attributes: attributes,
	}

	result, err := c.API.MessagesSendMedia(ctx, &tg.MessagesSendMediaRequest{
		Peer:     inputPeer,
		Media:    media,
		Message:  params.Caption,
		RandomID: time.Now().UnixNano(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to send voice: %w", err)
	}

	msgID := extractMessageID(result)
	return &types.SendVoiceResult{
		ID:       msgID,
		Date:     time.Now().Unix(),
		Peer:     params.Peer,
		Duration: params.Duration,
	}, nil
}

// SendVideoNote sends a video note (round video) to a peer.
func (c *Client) SendVideoNote(
	ctx context.Context, params types.SendVideoNoteParams,
) (*types.SendVideoNoteResult, error) {
	if err := c.CheckInitialized(); err != nil {
		return nil, err
	}

	inputPeer, err := c.ResolvePeer(ctx, params.Peer)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve peer: %w", err)
	}

	uploadedFile, err := uploadFile(ctx, c.API, params.File)
	if err != nil {
		return nil, fmt.Errorf("failed to upload video note: %w", err)
	}

	// Default length for video note (square video)
	length := params.Length
	if length == 0 {
		length = 240
	}

	// Video note attributes
	attributes := []tg.DocumentAttributeClass{
		&tg.DocumentAttributeVideo{
			RoundMessage: true,
			Duration:     float64(params.Duration),
			W:            length,
			H:            length,
		},
	}

	media := &tg.InputMediaUploadedDocument{
		File:       uploadedFile,
		MimeType:   "video/mp4",
		Attributes: attributes,
	}

	result, err := c.API.MessagesSendMedia(ctx, &tg.MessagesSendMediaRequest{
		Peer:     inputPeer,
		Media:    media,
		RandomID: time.Now().UnixNano(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to send video note: %w", err)
	}

	msgID := extractMessageID(result)
	return &types.SendVideoNoteResult{
		ID:       msgID,
		Date:     time.Now().Unix(),
		Peer:     params.Peer,
		Duration: params.Duration,
	}, nil
}
