// Package pin provides Telegram message pin operations.
package pin

import (
	"context"
	"fmt"

	"github.com/gotd/td/tg"
	"agent-telegram/telegram/client"
	"agent-telegram/telegram/types"
)

// Client provides pin operations.
type Client struct {
	*client.BaseClient
}

// NewClient creates a new pin client.
func NewClient(tc client.ParentClient) *Client {
	return &Client{
		BaseClient: &client.BaseClient{Parent: tc},
	}
}

// PinMessage pins a message.
func (c *Client) PinMessage(ctx context.Context, params types.PinMessageParams) (*types.PinMessageResult, error) {
	if c.API == nil {
		return nil, fmt.Errorf("client not initialized")
	}

	inputPeer, err := c.ResolvePeer(ctx, params.Peer)
	if err != nil {
		return nil, err
	}

	_, err = c.API.MessagesUpdatePinnedMessage(ctx, &tg.MessagesUpdatePinnedMessageRequest{
		Peer:  inputPeer,
		ID:    int(params.MessageID),
		Unpin: false,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to pin message: %w", err)
	}

	return &types.PinMessageResult{
		Success:   true,
		MessageID: params.MessageID,
	}, nil
}

// UnpinMessage unpins a message.
func (c *Client) UnpinMessage(ctx context.Context, params types.UnpinMessageParams) (*types.UnpinMessageResult, error) {
	if c.API == nil {
		return nil, fmt.Errorf("client not initialized")
	}

	inputPeer, err := c.ResolvePeer(ctx, params.Peer)
	if err != nil {
		return nil, err
	}

	_, err = c.API.MessagesUpdatePinnedMessage(ctx, &tg.MessagesUpdatePinnedMessageRequest{
		Peer:  inputPeer,
		ID:    int(params.MessageID),
		Unpin: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to unpin message: %w", err)
	}

	return &types.UnpinMessageResult{
		Success:   true,
		MessageID: params.MessageID,
	}, nil
}
