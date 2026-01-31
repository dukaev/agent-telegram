// Package telegram provides Telegram client contact functionality.
package telegram

import (
	"context"
	"fmt"
	"time"

	"github.com/gotd/td/tg"
)

// SendContact sends a contact to a peer.
func (c *Client) SendContact(ctx context.Context, params SendContactParams) (*SendContactResult, error) {
	if c.client == nil {
		return nil, fmt.Errorf("client not initialized")
	}

	api := c.client.API()

	inputPeer, err := resolvePeer(ctx, api, params.Peer)
	if err != nil {
		return nil, err
	}

	contact := &tg.InputMediaContact{
		PhoneNumber: params.Phone,
		FirstName:   params.FirstName,
		LastName:    params.LastName,
	}

	result, err := api.MessagesSendMedia(ctx, &tg.MessagesSendMediaRequest{
		Peer:     inputPeer,
		Media:    contact,
		RandomID: time.Now().UnixNano(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to send contact: %w", err)
	}

	msgID := extractMessageID(result)
	return &SendContactResult{
		ID:    msgID,
		Date:  time.Now().Unix(),
		Peer:  params.Peer,
		Phone: params.Phone,
	}, nil
}
