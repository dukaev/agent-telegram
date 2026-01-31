// Package message provides Telegram inline button operations.
package message

import (
	"context"
	"fmt"

	"github.com/gotd/td/tg"
	"agent-telegram/telegram/types"
)

// InspectInlineButtons inspects inline buttons in a message.
func (c *Client) InspectInlineButtons(
	ctx context.Context, params types.InspectInlineButtonsParams,
) (*types.InspectInlineButtonsResult, error) {
	if c.api == nil {
		return nil, fmt.Errorf("client not initialized")
	}

	// Get messages to find the one with inline buttons
	messages, err := c.api.MessagesGetMessages(ctx, []tg.InputMessageClass{
		&tg.InputMessageID{ID: int(params.MessageID)},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get message: %w", err)
	}

	var buttons []types.InlineButton

	// Extract buttons from the message
	switch m := messages.(type) {
	case *tg.MessagesMessages:
		for _, msg := range m.Messages {
			if userMsg, ok := msg.(*tg.Message); ok {
				if userMsg.ReplyMarkup != nil {
					buttons = extractButtons(userMsg.ReplyMarkup)
				}
			}
		}
	case *tg.MessagesMessagesSlice:
		for _, msg := range m.Messages {
			if userMsg, ok := msg.(*tg.Message); ok {
				if userMsg.ReplyMarkup != nil {
					buttons = extractButtons(userMsg.ReplyMarkup)
				}
			}
		}
	}

	return &types.InspectInlineButtonsResult{
		MessageID: params.MessageID,
		Buttons:   buttons,
	}, nil
}

// extractButtons extracts inline buttons from ReplyMarkup.
func extractButtons(markup tg.ReplyMarkupClass) []types.InlineButton {
	rm, ok := markup.(*tg.ReplyInlineMarkup)
	if !ok {
		return nil
	}

	var result []types.InlineButton
	for _, row := range rm.Rows {
		for _, button := range row.Buttons {
			switch b := button.(type) {
			case *tg.KeyboardButtonURL:
				result = append(result, types.InlineButton{
					Text:  b.Text,
					Data:  b.URL,
					Index: len(result),
				})
			case *tg.KeyboardButtonCallback:
				result = append(result, types.InlineButton{
					Text:  b.Text,
					Data:  string(b.Data),
					Index: len(result),
				})
			case *tg.KeyboardButtonSwitchInline:
				result = append(result, types.InlineButton{
					Text:  b.Text,
					Data:  b.Query,
					Index: len(result),
				})
			case *tg.KeyboardButtonGame:
				result = append(result, types.InlineButton{
					Text:  b.Text,
					Index: len(result),
				})
			case *tg.KeyboardButtonBuy:
				result = append(result, types.InlineButton{
					Text:  b.Text,
					Index: len(result),
				})
			case *tg.KeyboardButtonURLAuth:
				result = append(result, types.InlineButton{
					Text:  b.Text,
					Data:  b.URL,
					Index: len(result),
				})
			}
		}
	}

	return result
}

// PressInlineButton presses an inline button.
func (c *Client) PressInlineButton(
	ctx context.Context, params types.PressInlineButtonParams,
) (*types.PressInlineButtonResult, error) {
	if c.api == nil {
		return nil, fmt.Errorf("client not initialized")
	}

	// Resolve peer for the callback request
	peer, err := resolvePeer(ctx, c.api, params.Peer)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve peer: %w", err)
	}

	// First, inspect the buttons to find the callback data
	inspectResult, err := c.InspectInlineButtons(ctx, types.InspectInlineButtonsParams{
		PeerInfo: params.PeerInfo,
		MsgID:    params.MsgID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to inspect buttons: %w", err)
	}

	if params.ButtonIndex < 0 || params.ButtonIndex >= len(inspectResult.Buttons) {
		return nil, fmt.Errorf("button index out of range")
	}

	button := inspectResult.Buttons[params.ButtonIndex]

	// Press the button using the callback data
	_, err = c.api.MessagesGetBotCallbackAnswer(ctx, &tg.MessagesGetBotCallbackAnswerRequest{
		Peer:  peer,
		MsgID: int(params.MsgID.MessageID),
		Data:  []byte(button.Data),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to press button: %w", err)
	}

	return &types.PressInlineButtonResult{
		Success:   true,
		MessageID: params.MsgID.MessageID,
	}, nil
}
