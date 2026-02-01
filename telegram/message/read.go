// Package message provides message read operations.
package message

import (
	"context"
	"fmt"

	"github.com/gotd/td/tg"
	"agent-telegram/telegram/types"
)

// ReadMessages marks messages as read in a chat.
func (c *Client) ReadMessages(ctx context.Context, params types.ReadMessagesParams) (*types.ReadMessagesResult, error) {
	if err := c.CheckInitialized(); err != nil {
		return nil, err
	}

	inputPeer, err := c.ResolvePeer(ctx, params.Peer)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve peer: %w", err)
	}

	maxID := int(params.MaxID)

	// Handle channels differently
	switch p := inputPeer.(type) {
	case *tg.InputPeerChannel:
		_, err = c.API.ChannelsReadHistory(ctx, &tg.ChannelsReadHistoryRequest{
			Channel: &tg.InputChannel{
				ChannelID:  p.ChannelID,
				AccessHash: p.AccessHash,
			},
			MaxID: maxID,
		})
	default:
		_, err = c.API.MessagesReadHistory(ctx, &tg.MessagesReadHistoryRequest{
			Peer:  inputPeer,
			MaxID: maxID,
		})
	}

	if err != nil {
		return nil, fmt.Errorf("failed to mark messages as read: %w", err)
	}

	return &types.ReadMessagesResult{
		Success: true,
		MaxID:   params.MaxID,
	}, nil
}

// SetTyping sends a typing indicator to a chat.
func (c *Client) SetTyping(ctx context.Context, params types.SetTypingParams) (*types.SetTypingResult, error) {
	if err := c.CheckInitialized(); err != nil {
		return nil, err
	}

	inputPeer, err := c.ResolvePeer(ctx, params.Peer)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve peer: %w", err)
	}

	// Determine action type
	var action tg.SendMessageActionClass
	switch params.Action {
	case "upload_photo":
		action = &tg.SendMessageUploadPhotoAction{}
	case "record_video":
		action = &tg.SendMessageRecordVideoAction{}
	case "upload_video":
		action = &tg.SendMessageUploadVideoAction{}
	case "record_audio", "record_voice":
		action = &tg.SendMessageRecordAudioAction{}
	case "upload_audio", "upload_voice":
		action = &tg.SendMessageUploadAudioAction{}
	case "upload_document", "upload_file":
		action = &tg.SendMessageUploadDocumentAction{}
	case "geo", "location":
		action = &tg.SendMessageGeoLocationAction{}
	case "choose_contact":
		action = &tg.SendMessageChooseContactAction{}
	case "game":
		action = &tg.SendMessageGamePlayAction{}
	case "record_round", "record_video_note":
		action = &tg.SendMessageRecordRoundAction{}
	case "upload_round", "upload_video_note":
		action = &tg.SendMessageUploadRoundAction{}
	case "choose_sticker":
		action = &tg.SendMessageChooseStickerAction{}
	case "cancel":
		action = &tg.SendMessageCancelAction{}
	default: // "typing" or empty
		action = &tg.SendMessageTypingAction{}
	}

	_, err = c.API.MessagesSetTyping(ctx, &tg.MessagesSetTypingRequest{
		Peer:   inputPeer,
		Action: action,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to set typing: %w", err)
	}

	return &types.SetTypingResult{Success: true}, nil
}
