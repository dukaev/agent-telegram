// Package message provides Telegram message operations.
package message

import (
	"agent-telegram/telegram/client"
)

// Client provides message operations.
type Client struct {
	*client.BaseClient
}

// NewClient creates a new message client.
func NewClient(tc client.ParentClient) *Client {
	return &Client{
		BaseClient: &client.BaseClient{Parent: tc},
	}
}

// MessageResult represents a single message result.
type MessageResult struct { // revive:disable:exported // Used internally
	ID       int64  `json:"id"`
	Date     int64  `json:"date"`
	Text     string `json:"text,omitempty"`
	FromID   string `json:"fromId,omitempty"`
	FromName string `json:"fromName,omitempty"`
	Out      bool   `json:"out"`
}
