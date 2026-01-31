// Package telegram provides Telegram client update handlers.
package telegram

import (
	"context"

	"github.com/gotd/td/tg"
	"agent-telegram/telegram/types"
)

// RegisterUpdateHandlers registers update handlers on the dispatcher.
func (c *Client) RegisterUpdateHandlers(dispatcher tg.UpdateDispatcher) {
	if c.updateStore == nil {
		return
	}

	// New messages
	dispatcher.OnNewMessage(func(_ context.Context, entities tg.Entities, update *tg.UpdateNewMessage) error {
		c.updateStore.Add(NewStoredUpdate(types.UpdateTypeNewMessage, map[string]any{
			"message": MessageData(update.Message, entities),
		}))
		return nil
	})

	// Edited messages
	dispatcher.OnEditMessage(func(_ context.Context, entities tg.Entities, update *tg.UpdateEditMessage) error {
		c.updateStore.Add(NewStoredUpdate(types.UpdateTypeEditMessage, map[string]any{
			"message": MessageData(update.Message, entities),
		}))
		return nil
	})
}
