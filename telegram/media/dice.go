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

	return &types.SendDiceResult{
		ID:       msgID,
		Date:     time.Now().Unix(),
		Peer:     params.Peer,
		Value:    value,
		Emoticon: emoticon,
	}, nil
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
			if newMsg, ok := update.(*tg.UpdateNewMessage); ok {
				if msg, ok := newMsg.Message.(*tg.Message); ok {
					if dice, ok := msg.Media.(*tg.MessageMediaDice); ok {
						return dice.Value
					}
				}
			}
		}
	}
	return 0
}
