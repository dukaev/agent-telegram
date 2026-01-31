// Package telegram provides Telegram client update handlers.
package telegram

import (
	"context"
	"fmt"

	"github.com/gotd/td/tg"
	"agent-telegram/telegram/types"
)

// RegisterUpdateHandlers registers update handlers on the dispatcher.
func (c *Client) RegisterUpdateHandlers(dispatcher tg.UpdateDispatcher) {
	if c.updateStore == nil {
		return
	}

	// New messages
	dispatcher.OnNewMessage(func(_ context.Context, _ tg.Entities, update *tg.UpdateNewMessage) error {
		peer := unknownPeer
		if msg, ok := update.Message.(*tg.Message); ok && msg.PeerID != nil {
			peer = peerToString(msg.PeerID)
		}
		c.updateStore.Add(NewStoredUpdate(types.UpdateTypeNewMessage, map[string]any{
			"message": MessageData(update.Message),
			"peer":    peer,
		}))
		return nil
	})

	// Edited messages
	dispatcher.OnEditMessage(func(_ context.Context, _ tg.Entities, update *tg.UpdateEditMessage) error {
		peer := unknownPeer
		if msg, ok := update.Message.(*tg.Message); ok && msg.PeerID != nil {
			peer = peerToString(msg.PeerID)
		}
		c.updateStore.Add(NewStoredUpdate(types.UpdateTypeEditMessage, map[string]any{
			"message": MessageData(update.Message),
			"peer":    peer,
		}))
		return nil
	})
}

// peerToString converts a PeerClass to a string representation.
func peerToString(peer tg.PeerClass) string {
	switch p := peer.(type) {
	case *tg.PeerUser:
		return fmt.Sprintf("user:%d", p.UserID)
	case *tg.PeerChat:
		return fmt.Sprintf("chat:%d", p.ChatID)
	case *tg.PeerChannel:
		return fmt.Sprintf("channel:%d", p.ChannelID)
	default:
		return unknownPeer
	}
}
