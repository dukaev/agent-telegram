// Package telegram provides Telegram client pin functionality.
package telegram

import (
	"context"
	"fmt"

	"github.com/gotd/td/tg"
)

// PinMessage pins a message.
func (c *Client) PinMessage(ctx context.Context, params PinMessageParams) (*PinMessageResult, error) {
	if c.client == nil {
		return nil, fmt.Errorf("client not initialized")
	}

	api := c.client.API()

	inputPeer, err := resolvePeer(ctx, api, params.Peer)
	if err != nil {
		return nil, err
	}

	_, err = api.MessagesUpdatePinnedMessage(ctx, &tg.MessagesUpdatePinnedMessageRequest{
		Peer:      inputPeer,
		ID:        int(params.MessageID),
		Unpin:     false,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to pin message: %w", err)
	}

	return &PinMessageResult{
		Success:   true,
		MessageID: params.MessageID,
	}, nil
}

// UnpinMessage unpins a message.
func (c *Client) UnpinMessage(ctx context.Context, params UnpinMessageParams) (*UnpinMessageResult, error) {
	if c.client == nil {
		return nil, fmt.Errorf("client not initialized")
	}

	api := c.client.API()

	inputPeer, err := resolvePeer(ctx, api, params.Peer)
	if err != nil {
		return nil, err
	}

	_, err = api.MessagesUpdatePinnedMessage(ctx, &tg.MessagesUpdatePinnedMessageRequest{
		Peer:      inputPeer,
		ID:        int(params.MessageID),
		Unpin:     true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to unpin message: %w", err)
	}

	return &UnpinMessageResult{
		Success:   true,
		MessageID: params.MessageID,
	}, nil
}
