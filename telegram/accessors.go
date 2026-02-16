package telegram

import (
	"context"
	"fmt"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/tg"
	"agent-telegram/telegram/types"
)

// Client returns the underlying telegram.Client
func (c *Client) Client() *telegram.Client {
	return c.client
}

// Message returns the message client.
func (c *Client) Message() MessageClient {
	return c.message
}

// Media returns the media client.
func (c *Client) Media() MediaClient {
	return c.media
}

// Chat returns the chat client.
func (c *Client) Chat() ChatClient {
	return c.chat
}

// User returns the user client.
func (c *Client) User() UserClient {
	return c.user
}

// Pin returns the pin client.
func (c *Client) Pin() PinClient {
	return c.pin
}

// Reaction returns the reaction client.
func (c *Client) Reaction() ReactionClient {
	return c.reaction
}

// Search returns the search client.
func (c *Client) Search() SearchClient {
	return c.search
}

// Gift returns the gift client.
func (c *Client) Gift() GiftClient {
	return c.gift
}

// GetMe returns the current user information.
func (c *Client) GetMe(ctx context.Context) (*tg.User, error) {
	if c.client == nil {
		return nil, fmt.Errorf("client not initialized")
	}
	return c.client.Self(ctx)
}

// GetUpdates pops and returns stored updates.
func (c *Client) GetUpdates(limit int) []types.StoredUpdate {
	if c.updateStore == nil {
		return []types.StoredUpdate{}
	}
	return c.updateStore.Get(limit)
}
