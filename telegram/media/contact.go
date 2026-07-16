// Package media provides Telegram contact operations.
package media

import (
	"context"
	"fmt"
	"time"

	"agent-telegram/telegram/internal/replytarget"
	"agent-telegram/telegram/types"
	"github.com/gotd/td/tg"
)

// SendContact sends a contact to a peer.
func (c *Client) SendContact(ctx context.Context, params types.SendContactParams) (*types.SendContactResult, error) {
	inputPeer, err := c.InitAndResolve(ctx, params.Peer)
	if err != nil {
		return nil, err
	}

	contact := &tg.InputMediaContact{
		PhoneNumber: params.Phone,
		FirstName:   params.FirstName,
		LastName:    params.LastName,
	}

	result, err := c.API().MessagesSendMedia(ctx, &tg.MessagesSendMediaRequest{
		Peer:     inputPeer,
		Media:    contact,
		ReplyTo:  replytarget.Build(params.ThreadTarget),
		RandomID: time.Now().UnixNano(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to send contact: %w", err)
	}

	msgID := extractMessageID(result)
	return &types.SendContactResult{
		ID:    msgID,
		Date:  time.Now().Unix(),
		Peer:  params.Peer,
		Phone: params.Phone,
	}, nil
}
