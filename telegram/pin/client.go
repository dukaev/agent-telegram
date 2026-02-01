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
	if err := c.updatePin(ctx, params.Peer, params.MessageID, false); err != nil {
		return nil, err
	}
	return &types.PinMessageResult{Success: true, MessageID: params.MessageID}, nil
}

// UnpinMessage unpins a message.
func (c *Client) UnpinMessage(ctx context.Context, params types.UnpinMessageParams) (*types.UnpinMessageResult, error) {
	if err := c.updatePin(ctx, params.Peer, params.MessageID, true); err != nil {
		return nil, err
	}
	return &types.UnpinMessageResult{Success: true, MessageID: params.MessageID}, nil
}

// updatePin is the shared implementation for pin/unpin operations.
func (c *Client) updatePin(ctx context.Context, peer string, messageID int64, unpin bool) error {
	if err := c.CheckInitialized(); err != nil {
		return err
	}

	inputPeer, err := c.ResolvePeer(ctx, peer)
	if err != nil {
		return err
	}

	_, err = c.API.MessagesUpdatePinnedMessage(ctx, &tg.MessagesUpdatePinnedMessageRequest{
		Peer:  inputPeer,
		ID:    int(messageID),
		Unpin: unpin,
	})
	if err != nil {
		action := "pin"
		if unpin {
			action = "unpin"
		}
		return fmt.Errorf("failed to %s message: %w", action, err)
	}
	return nil
}
