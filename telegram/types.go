// Package telegram provides common types for Telegram client.
package telegram

import "time"

// UpdateType represents the type of Telegram update.
type UpdateType string

const (
	// UpdateTypeNewMessage is a new message update.
	UpdateTypeNewMessage UpdateType = "new_message"
	// UpdateTypeEditMessage is an edited message update.
	UpdateTypeEditMessage UpdateType = "edit_message"
	// UpdateTypeNewChat is a new chat update.
	UpdateTypeNewChat UpdateType = "new_chat"
	// UpdateTypeDelete is a delete update.
	UpdateTypeDelete UpdateType = "delete"
	// UpdateTypeOther is an other type update.
	UpdateTypeOther UpdateType = "other"
)

// StoredUpdate represents a stored Telegram update.
type StoredUpdate struct {
	ID        int64     `json:"id"`
	Type      UpdateType `json:"type"`
	Timestamp time.Time `json:"timestamp"`
	Data      map[string]any `json:"data"`
}

// MessageResult represents a single message result.
type MessageResult struct {
	ID       int64  `json:"id"`
	Date     int64  `json:"date"`
	Text     string `json:"text,omitempty"`
	FromID   string `json:"fromId,omitempty"`
	FromName string `json:"fromName,omitempty"`
	Out      bool   `json:"out"`
}

// GetMessagesParams holds parameters for GetMessages.
type GetMessagesParams struct {
	Username string `json:"username"`
	Limit    int    `json:"limit"`
	Offset   int    `json:"offset"`
}

// GetMessagesResult is the result of GetMessages.
type GetMessagesResult struct {
	Messages []MessageResult `json:"messages"`
	Limit    int             `json:"limit"`
	Offset   int             `json:"offset"`
	Count    int             `json:"count"`
	Username string          `json:"username"`
}

// GetChatsParams holds parameters for GetChats.
type GetChatsParams struct {
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

// GetChatsResult is the result of GetChats.
type GetChatsResult struct {
	Chats  []map[string]any `json:"chats"`
	Limit  int              `json:"limit"`
	Offset int              `json:"offset"`
	Count  int              `json:"count"`
}

// GetUpdatesParams holds parameters for GetUpdates.
type GetUpdatesParams struct {
	Limit int `json:"limit"`
}

// GetUpdatesResult is the result of GetUpdates.
type GetUpdatesResult struct {
	Updates []StoredUpdate `json:"updates"`
	Count   int            `json:"count"`
}

// ClearMessagesParams holds parameters for ClearMessages.
type ClearMessagesParams struct {
	Peer      string   `json:"peer,omitempty"`
	Username  string   `json:"username,omitempty"`
	MessageIDs []int64 `json:"messageIds"`
}

// ClearMessagesResult is the result of ClearMessages.
type ClearMessagesResult struct {
	Success bool   `json:"success"`
	Cleared int    `json:"cleared"`
	Peer    string `json:"peer"`
}

// ClearHistoryParams holds parameters for ClearHistory.
type ClearHistoryParams struct {
	Peer     string `json:"peer,omitempty"`
	Username string `json:"username,omitempty"`
	Revoke   bool   `json:"revoke,omitempty"`
}

// ClearHistoryResult is the result of ClearHistory.
type ClearHistoryResult struct {
	Success bool   `json:"success"`
	Peer    string `json:"peer"`
	Revoke  bool   `json:"revoke"`
}
