// Package media provides Telegram dice operations.
package media

import (
	"context"
	"fmt"
	"time"

	"github.com/gotd/td/tg"

	"agent-telegram/telegram/types"
)

// defaultDiceEmoticon is the default dice emoji.
const defaultDiceEmoticon = "🎲"

// SendDice sends a dice (random value) to a peer.
func (c *Client) SendDice(ctx context.Context, params types.SendDiceParams) (*types.SendDiceResult, error) {
	if err := c.CheckInitialized(); err != nil {
		return nil, err
	}

	inputPeer, err := c.ResolvePeer(ctx, params.Peer)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve peer %s: %w", params.Peer, err)
	}

	emoticon := params.Emoticon
	if emoticon == "" {
		emoticon = defaultDiceEmoticon
	}

	dice := &tg.InputMediaDice{
		Emoticon: emoticon,
	}

	req := &tg.MessagesSendMediaRequest{
		Peer:     inputPeer,
		Media:    dice,
		RandomID: time.Now().UnixNano(),
	}
	if params.ReplyTo != 0 {
		req.ReplyTo = &tg.InputReplyToMessage{ReplyToMsgID: int(params.ReplyTo)}
	}

	result, err := c.API.MessagesSendMedia(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to send dice: %w", err)
	}

	msgID := extractMessageID(result)
	value := extractDiceValue(result)

	// For channels, dice value arrives later via edit update.
	// Poll the message to get the actual value.
	if value == 0 && msgID != 0 {
		value = c.waitForDiceValue(ctx, inputPeer, int(msgID))
	}

	return &types.SendDiceResult{
		ID:       msgID,
		Date:     time.Now().Unix(),
		Peer:     params.Peer,
		Value:    value,
		Emoticon: emoticon,
	}, nil
}

// waitForDiceValue polls a sent message to get the dice value (for channels).
func (c *Client) waitForDiceValue(ctx context.Context, peer tg.InputPeerClass, msgID int) int {
	// Dice animations take ~3s in Telegram. Wait before first poll.
	time.Sleep(3 * time.Second)

	for range 5 {
		var msgs tg.MessagesMessagesClass
		var err error

		if ch, ok := peer.(*tg.InputPeerChannel); ok {
			msgs, err = c.API.ChannelsGetMessages(ctx, &tg.ChannelsGetMessagesRequest{
				Channel: &tg.InputChannel{ChannelID: ch.ChannelID, AccessHash: ch.AccessHash},
				ID:      []tg.InputMessageClass{&tg.InputMessageID{ID: msgID}},
			})
		} else {
			msgs, err = c.API.MessagesGetMessages(ctx, []tg.InputMessageClass{
				&tg.InputMessageID{ID: msgID},
			})
		}
		if err != nil {
			time.Sleep(time.Second)
			continue
		}
		if v := extractDiceFromMessages(msgs); v != 0 {
			return v
		}
		time.Sleep(time.Second)
	}
	return 0
}

// extractDiceFromMessages extracts dice value from a MessagesMessagesClass response.
func extractDiceFromMessages(msgs tg.MessagesMessagesClass) int {
	var messages []tg.MessageClass
	switch m := msgs.(type) {
	case *tg.MessagesMessages:
		messages = m.Messages
	case *tg.MessagesMessagesSlice:
		messages = m.Messages
	case *tg.MessagesChannelMessages:
		messages = m.Messages
	}
	for _, msg := range messages {
		if m, ok := msg.(*tg.Message); ok {
			if dice, ok := m.Media.(*tg.MessageMediaDice); ok {
				return dice.Value
			}
		}
	}
	return 0
}

// extractDiceValue extracts the dice value from the Updates response.
func extractDiceValue(result tg.UpdatesClass) int {
	switch r := result.(type) {
	case *tg.UpdateShortSentMessage:
		if media, ok := r.GetMedia(); ok {
			if dice, ok := media.(*tg.MessageMediaDice); ok {
				return dice.Value
			}
		}
	case *tg.Updates:
		for _, update := range r.Updates {
			var msg *tg.Message
			switch u := update.(type) {
			case *tg.UpdateNewMessage:
				msg, _ = u.Message.(*tg.Message)
			case *tg.UpdateNewChannelMessage:
				msg, _ = u.Message.(*tg.Message)
			}
			if msg != nil {
				if dice, ok := msg.Media.(*tg.MessageMediaDice); ok {
					return dice.Value
				}
			}
		}
	}
	return 0
}
