// Package media provides Telegram contact operations.
package media

import (
	"context"
	"fmt"
	"time"

	"github.com/gotd/td/tg"
	"agent-telegram/telegram/types"
)

// SendContact sends a contact to a peer.
func (c *Client) SendContact(ctx context.Context, params types.SendContactParams) (*types.SendContactResult, error) {
	if err := c.CheckInitialized(); err != nil {
		return nil, err
	}

	inputPeer, err := c.ResolvePeer(ctx, params.Peer)
	if err != nil {
		return nil, err
	}

	contact := &tg.InputMediaContact{
		PhoneNumber: params.Phone,
		FirstName:   params.FirstName,
		LastName:    params.LastName,
	}

	result, err := c.API.MessagesSendMedia(ctx, &tg.MessagesSendMediaRequest{
		Peer:     inputPeer,
		Media:    contact,
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
