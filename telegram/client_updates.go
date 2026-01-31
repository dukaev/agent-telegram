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
	dispatcher.OnNewMessage(func(_ context.Context, entities tg.Entities, update *tg.UpdateNewMessage) error {
		peer := unknownPeer
		peerName := ""
		if msg, ok := update.Message.(*tg.Message); ok {
			if msg.PeerID != nil {
				peer = peerToString(msg.PeerID)
				peerName = getPeerName(entities, msg.PeerID)
			}
		}
		c.updateStore.Add(NewStoredUpdate(types.UpdateTypeNewMessage, map[string]any{
			"message": MessageData(update.Message, entities),
			"peer":    peer,
			"peer_name": peerName,
		}))
		return nil
	})

	// Edited messages
	dispatcher.OnEditMessage(func(_ context.Context, entities tg.Entities, update *tg.UpdateEditMessage) error {
		peer := unknownPeer
		peerName := ""
		if msg, ok := update.Message.(*tg.Message); ok {
			if msg.PeerID != nil {
				peer = peerToString(msg.PeerID)
				peerName = getPeerName(entities, msg.PeerID)
			}
		}
		c.updateStore.Add(NewStoredUpdate(types.UpdateTypeEditMessage, map[string]any{
			"message": MessageData(update.Message, entities),
			"peer":    peer,
			"peer_name": peerName,
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

// getPeerName gets the name of a peer from entities.
func getPeerName(entities tg.Entities, peerID tg.PeerClass) string {
	switch p := peerID.(type) {
	case *tg.PeerUser:
		if user, ok := entities.Users[p.UserID]; ok {
			name := user.FirstName
			if user.LastName != "" {
				name += " " + user.LastName
			}
			if name == "" && user.Username != "" {
				name = user.Username
			}
			if name == "" {
				name = fmt.Sprintf("user:%d", user.ID)
			}
			if user.Bot {
				name += " (bot)"
			}
			return name
		}
	case *tg.PeerChat:
		return fmt.Sprintf("chat:%d", p.ChatID)
	case *tg.PeerChannel:
		if channel, ok := entities.Channels[p.ChannelID]; ok {
			return channel.Title
		}
	}
	return ""
}
