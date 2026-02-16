// Package telegram provides Telegram client update handlers.
package telegram

import (
	"context"

	"github.com/gotd/td/tg"

	"agent-telegram/telegram/helpers"
	"agent-telegram/telegram/types"
)

// RegisterUpdateHandlers registers update handlers on the dispatcher.
func (c *Client) RegisterUpdateHandlers(dispatcher tg.UpdateDispatcher) {
	if c.updateStore == nil {
		return
	}

	// New messages (including service messages with gift actions)
	dispatcher.OnNewMessage(func(_ context.Context, entities tg.Entities, update *tg.UpdateNewMessage) error {
		if data := giftActionData(update.Message, entities); data != nil {
			c.updateStore.Add(NewStoredUpdate(types.UpdateTypeStarGift, data))
			return nil
		}
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

// giftActionData extracts gift data from a service message, or returns nil.
func giftActionData(msg tg.MessageClass, entities tg.Entities) map[string]any {
	svc, ok := msg.(*tg.MessageService)
	if !ok {
		return nil
	}

	data := map[string]any{
		"msgId": svc.ID,
		"date":  svc.Date,
		"out":   svc.Out,
	}
	if svc.FromID != nil {
		data["from"] = helpers.FormatPeer(svc.FromID, helpers.PeerFormatTyped)
		if name := getSenderName(entities, svc.FromID); name != "" {
			data["fromName"] = name
		}
	}
	if svc.PeerID != nil {
		data["peer"] = helpers.FormatPeer(svc.PeerID, helpers.PeerFormatTyped)
	}

	switch action := svc.Action.(type) {
	case *tg.MessageActionStarGift:
		data["action"] = "star_gift"
		data["convertStars"] = action.ConvertStars
		data["saved"] = action.Saved
		data["nameHidden"] = action.NameHidden
		if action.Message.Text != "" {
			data["message"] = action.Message.Text
		}
		if action.FromID != nil {
			data["giftFrom"] = helpers.FormatPeer(action.FromID, helpers.PeerFormatTyped)
		}
		fillGiftData(data, action.Gift)

	case *tg.MessageActionStarGiftUnique:
		data["action"] = "star_gift_unique"
		data["upgrade"] = action.Upgrade
		data["transferred"] = action.Transferred
		data["saved"] = action.Saved
		if action.TransferStars > 0 {
			data["transferStars"] = action.TransferStars
		}
		if action.FromID != nil {
			data["giftFrom"] = helpers.FormatPeer(action.FromID, helpers.PeerFormatTyped)
		}
		fillGiftData(data, action.Gift)

	default:
		return nil
	}

	return data
}

// fillGiftData adds gift-specific fields to the data map.
func fillGiftData(data map[string]any, gift tg.StarGiftClass) {
	switch g := gift.(type) {
	case *tg.StarGift:
		data["giftId"] = g.ID
		data["stars"] = g.Stars
		if g.Title != "" {
			data["title"] = g.Title
		}
	case *tg.StarGiftUnique:
		data["giftId"] = g.ID
		data["title"] = g.Title
		data["slug"] = g.Slug
		data["num"] = g.Num
		if len(g.ResellAmount) > 0 {
			data["resellStars"] = g.ResellAmount[0].GetAmount()
		}
		data["availabilityTotal"] = g.AvailabilityTotal
		data["ownerName"] = g.OwnerName
	}
	// Fallback name for untitled regular gifts
	if data["title"] == nil || data["title"] == "" {
		if id, ok := data["giftId"].(int64); ok {
			if name, exists := giftNamesFallback[id]; exists {
				data["title"] = name
			}
		}
	}
}

// giftNamesFallback provides names for common untitled gifts.
var giftNamesFallback = map[int64]string{
	5170145012310081615: "Heart",
	5170233102089322756: "Bear",
}

