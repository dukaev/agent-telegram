package telegram

import (
	"context"
	"fmt"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/tg"
	"agent-telegram/telegram/chat"
	"agent-telegram/telegram/media"
	"agent-telegram/telegram/message"
	"agent-telegram/telegram/pin"
	"agent-telegram/telegram/reaction"
	"agent-telegram/telegram/search"
	"agent-telegram/telegram/types"
	"agent-telegram/telegram/user"
)

// Client returns the underlying telegram.Client
func (c *Client) Client() *telegram.Client {
	return c.client
}

// Message returns the message client.
func (c *Client) Message() *message.Client {
	return c.message
}

// Media returns the media client.
func (c *Client) Media() *media.Client {
	return c.media
}

// Chat returns the chat client.
func (c *Client) Chat() *chat.Client {
	return c.chat
}

// User returns the user client.
func (c *Client) User() *user.Client {
	return c.user
}

// Pin returns the pin client.
func (c *Client) Pin() *pin.Client {
	return c.pin
}

// Reaction returns the reaction client.
func (c *Client) Reaction() *reaction.Client {
	return c.reaction
}

// Search returns the search client.
func (c *Client) Search() *search.Client {
	return c.search
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

// InspectReplyKeyboard inspects the reply keyboard from a chat.
func (c *Client) InspectReplyKeyboard(ctx context.Context, params types.PeerInfo) (*types.ReplyKeyboardResult, error) {
	return c.message.InspectReplyKeyboard(ctx, params)
}
