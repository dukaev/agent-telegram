// Package message provides reply keyboard operations.
package message

import (
	"context"
	"fmt"

	"github.com/gotd/td/tg"
	"agent-telegram/telegram/types"
)

// InspectReplyKeyboard inspects the reply keyboard from a chat.
//nolint:funlen // Function requires iterating through messages to find keyboard markup
func (c *Client) InspectReplyKeyboard(ctx context.Context, params types.PeerInfo) (*types.ReplyKeyboardResult, error) {
	if c.api == nil {
		return nil, fmt.Errorf("client not initialized")
	}

	// Normalize peer
	peer := params.Peer
	if peer == "" {
		peer = "@" + params.Username
	}

	// Resolve peer
	inputPeer, err := c.resolvePeer(ctx, peer)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve peer: %w", err)
	}

	// Get recent messages to find reply keyboard
	messagesClass, err := c.api.MessagesGetHistory(ctx, &tg.MessagesGetHistoryRequest{
		Peer:      inputPeer,
		Limit:     50,
		OffsetID:  0,
		AddOffset: 0,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get messages: %w", err)
	}

	messages, _ := extractMessagesData(messagesClass)

	// Search for reply keyboard in messages (newest first)
	for i := len(messages) - 1; i >= 0; i-- {
		msg, ok := messages[i].(*tg.Message)
		if !ok {
			continue
		}

		if msg.ReplyMarkup != nil {
			if rm, ok := msg.ReplyMarkup.(*tg.ReplyKeyboardMarkup); ok {
				// Found reply keyboard markup
				return &types.ReplyKeyboardResult{
					Peer:         peer,
					MessageID:    int64(msg.ID),
					Keyboard:     convertReplyKeyboardMarkup(rm),
					Found:        true,
				}, nil
			}
			if _, ok := msg.ReplyMarkup.(*tg.ReplyKeyboardForceReply); ok {
				// Force reply keyboard
				return &types.ReplyKeyboardResult{
					Peer:      peer,
					MessageID: int64(msg.ID),
					ForceReply: true,
					Found:     true,
				}, nil
			}
			if _, ok := msg.ReplyMarkup.(*tg.ReplyKeyboardHide); ok {
				// Keyboard is hidden
				return &types.ReplyKeyboardResult{
					Peer:      peer,
					MessageID: int64(msg.ID),
					Hidden:    true,
					Found:     true,
				}, nil
			}
		}
	}

	// No keyboard found
	return &types.ReplyKeyboardResult{
		Peer:  peer,
		Found: false,
	}, nil
}

// convertReplyKeyboardMarkup converts tg.ReplyKeyboardMarkup to our format.
func convertReplyKeyboardMarkup(rm *tg.ReplyKeyboardMarkup) types.ReplyKeyboard {
	keyboard := types.ReplyKeyboard{
		Resize:     rm.Resize,
		SingleUse:  rm.SingleUse,
		Selective:  rm.Selective,
		Persistent: rm.Persistent,
		Placeholder: rm.Placeholder,
		Rows:       make([][]types.KeyboardButton, 0, len(rm.Rows)),
	}

	for _, row := range rm.Rows {
		buttonRow := make([]types.KeyboardButton, 0, len(row.Buttons))
		for _, btn := range row.Buttons {
			buttonRow = append(buttonRow, convertKeyboardButton(btn))
		}
		keyboard.Rows = append(keyboard.Rows, buttonRow)
	}

	return keyboard
}

// convertKeyboardButton converts a KeyboardButtonClass to our format.
func convertKeyboardButton(btn tg.KeyboardButtonClass) types.KeyboardButton {
	result := types.KeyboardButton{}

	switch b := btn.(type) {
	case *tg.KeyboardButton:
		result.Text = b.Text
		result.Type = "text"

	case *tg.KeyboardButtonURL:
		result.Text = b.Text
		result.URL = b.URL
		result.Type = "url"

	case *tg.KeyboardButtonRequestPhone:
		result.Text = b.Text
		result.Type = "request_phone"

	case *tg.KeyboardButtonRequestGeoLocation:
		result.Text = b.Text
		result.Type = "request_location"

	case *tg.KeyboardButtonRequestPoll:
		result.Text = b.Text
		result.Type = "request_poll"
		if b.Quiz {
			result.PollType = "quiz"
		} else {
			result.PollType = "regular"
		}

	case *tg.KeyboardButtonWebView:
		result.Text = b.Text
		result.URL = b.URL
		result.Type = "web_view"

	case *tg.KeyboardButtonRequestPeer:
		result.Text = b.Text
		result.Type = "request_peer"
		result.ButtonID = b.ButtonID
		result.MaxQuantity = b.MaxQuantity

	case *tg.KeyboardButtonUserProfile:
		result.Text = b.Text
		result.Type = "user_profile"
		result.UserID = b.UserID

	default:
		result.Type = "unknown"
	}

	return result
}
