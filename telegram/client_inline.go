// Package telegram provides Telegram client inline button functionality.
package telegram

import (
	"context"
	"fmt"

	"github.com/gotd/td/tg"
)

// InspectInlineButtons inspects inline buttons in a message.
func (c *Client) InspectInlineButtons(
	ctx context.Context, params InspectInlineButtonsParams,
) (*InspectInlineButtonsResult, error) {
	if c.client == nil {
		return nil, fmt.Errorf("client not initialized")
	}

	api := c.client.API()

	// Get messages to find the one with inline buttons
	messages, err := api.MessagesGetMessages(ctx, []tg.InputMessageClass{
		&tg.InputMessageID{ID: int(params.MessageID)},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get message: %w", err)
	}

	var buttons []InlineButton

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

	return &InspectInlineButtonsResult{
		MessageID: params.MessageID,
		Buttons:   buttons,
	}, nil
}

// extractButtons extracts inline buttons from ReplyMarkup.
func extractButtons(markup tg.ReplyMarkupClass) []InlineButton {
	rm, ok := markup.(*tg.ReplyInlineMarkup)
	if !ok {
		return nil
	}

	var result []InlineButton
	for _, row := range rm.Rows {
		for _, button := range row.Buttons {
			switch b := button.(type) {
			case *tg.KeyboardButtonURL:
				result = append(result, InlineButton{
					Text:  b.Text,
					Data:  b.URL,
					Index: len(result),
				})
			case *tg.KeyboardButtonCallback:
				result = append(result, InlineButton{
					Text:  b.Text,
					Data:  string(b.Data),
					Index: len(result),
				})
			case *tg.KeyboardButtonSwitchInline:
				result = append(result, InlineButton{
					Text:  b.Text,
					Data:  b.Query,
					Index: len(result),
				})
			case *tg.KeyboardButtonGame:
				result = append(result, InlineButton{
					Text:  b.Text,
					Index: len(result),
				})
			case *tg.KeyboardButtonBuy:
				result = append(result, InlineButton{
					Text:  b.Text,
					Index: len(result),
				})
			case *tg.KeyboardButtonURLAuth:
				result = append(result, InlineButton{
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
	ctx context.Context, params PressInlineButtonParams,
) (*PressInlineButtonResult, error) {
	if c.client == nil {
		return nil, fmt.Errorf("client not initialized")
	}

	// First, inspect the buttons to find the callback data
	inspectResult, err := c.InspectInlineButtons(ctx, InspectInlineButtonsParams{
		Peer:      params.Peer,
		MessageID: params.MessageID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to inspect buttons: %w", err)
	}

	if params.ButtonIndex < 0 || params.ButtonIndex >= len(inspectResult.Buttons) {
		return nil, fmt.Errorf("button index out of range")
	}

	button := inspectResult.Buttons[params.ButtonIndex]

	// Press the button using the callback data
	api := c.client.API()
	_, err = api.MessagesGetBotCallbackAnswer(ctx, &tg.MessagesGetBotCallbackAnswerRequest{
		Peer:  &tg.InputPeerUser{},
		MsgID: int(params.MessageID),
		Data:  []byte(button.Data),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to press button: %w", err)
	}

	return &PressInlineButtonResult{
		Success:   true,
		MessageID: params.MessageID,
	}, nil
}
