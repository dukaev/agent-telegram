// Package pin provides Telegram message pin operations.
package pin

import (
	"context"
	"fmt"

	"github.com/gotd/td/tg"
	"agent-telegram/telegram/types"
)

// Client provides pin operations.
type Client struct {
	api    *tg.Client
	parent ParentClient
}

// ParentClient is an interface for accessing parent client methods.
type ParentClient interface {
	ResolvePeer(ctx context.Context, peer string) (tg.InputPeerClass, error)
}

// NewClient creates a new pin client.
func NewClient(tc ParentClient) *Client {
	return &Client{
		parent: tc,
	}
}

// SetAPI sets the API client (called when the telegram client is initialized).
func (c *Client) SetAPI(api *tg.Client) {
	c.api = api
}

// resolvePeer resolves a peer string to InputPeerClass using the parent client's cache.
func (c *Client) resolvePeer(ctx context.Context, peer string) (tg.InputPeerClass, error) {
	if c.parent == nil {
		return nil, fmt.Errorf("parent client not set")
	}
	return c.parent.ResolvePeer(ctx, peer)
}

// PinMessage pins a message.
func (c *Client) PinMessage(ctx context.Context, params types.PinMessageParams) (*types.PinMessageResult, error) {
	if c.api == nil {
		return nil, fmt.Errorf("client not initialized")
	}

	inputPeer, err := c.resolvePeer(ctx, params.Peer)
	if err != nil {
		return nil, err
	}

	_, err = c.api.MessagesUpdatePinnedMessage(ctx, &tg.MessagesUpdatePinnedMessageRequest{
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
	if c.api == nil {
		return nil, fmt.Errorf("client not initialized")
	}

	inputPeer, err := c.resolvePeer(ctx, params.Peer)
	if err != nil {
		return nil, err
	}

	_, err = c.api.MessagesUpdatePinnedMessage(ctx, &tg.MessagesUpdatePinnedMessageRequest{
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
