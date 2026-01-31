// Package user provides Telegram user operations.
package user

import (
	"context"
	"fmt"

	"github.com/gotd/td/tg"
)

// Client provides user operations.
type Client struct {
	api    *tg.Client
	parent ParentClient
}

// ParentClient is an interface for accessing parent client methods.
type ParentClient interface {
	ResolvePeer(ctx context.Context, peer string) (tg.InputPeerClass, error)
}

// NewClient creates a new user client.
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

// trimUsernamePrefix removes the @ prefix from username.
func trimUsernamePrefix(username string) string {
	if len(username) > 0 && username[0] == '@' {
		return username[1:]
	}
	return username
}
