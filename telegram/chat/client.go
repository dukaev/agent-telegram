// Package chat provides Telegram chat operations.
package chat

import (
	"agent-telegram/telegram/client"
)

// Client provides chat operations.
type Client struct {
	*client.BaseClient
}

// NewClient creates a new chat client.
func NewClient(tc client.ParentClient) *Client {
	return &Client{
		BaseClient: &client.BaseClient{Parent: tc},
	}
}
