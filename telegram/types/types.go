// Package types provides common types for Telegram client.
package types // revive:disable:var-naming

import (
	"fmt"
	"time"
)

// PeerInfo is a base type for parameters that need peer or username.
type PeerInfo struct {
	Peer     string `json:"peer,omitempty"`
	Username string `json:"username,omitempty"`
}

// ValidatePeer validates that either peer or username is set.
func (p PeerInfo) ValidatePeer() error {
	if p.Peer == "" && p.Username == "" {
		return fmt.Errorf("peer or username is required")
	}
	return nil
}

// MsgID is a base type for parameters that need a message ID.
type MsgID struct {
	MessageID int64 `json:"messageId"`
}

// ValidateMessageID validates that messageId is set.
func (m MsgID) ValidateMessageID() error {
	if m.MessageID == 0 {
		return fmt.Errorf("messageId is required")
	}
	return nil
}

// RequiredText is a base type for parameters with a required text field.
type RequiredText struct {
	Text string `json:"text"`
}

// ValidateText validates that text is set.
func (r RequiredText) ValidateText() error {
	if r.Text == "" {
		return fmt.Errorf("text is required")
	}
	return nil
}

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
	ID             int64           `json:"id"`
	Date           int64           `json:"date"`
	Text           string          `json:"text,omitempty"`
	FromID         string          `json:"fromId,omitempty"`
	FromName       string          `json:"fromName,omitempty"`
	Out            bool            `json:"out"`
	Buttons        []InlineButton  `json:"buttons,omitempty"`

	// Additional message fields
	PeerID         string          `json:"peerId,omitempty"`        // Chat where message was sent
	EditDate       int64           `json:"editDate,omitempty"`      // When message was edited
	Media          map[string]any  `json:"media,omitempty"`         // Media attachment (photo, document, etc.)
	Views          int             `json:"views,omitempty"`         // View count for channel posts
	Forwards       int             `json:"forwards,omitempty"`      // Forward counter
	ReplyTo        map[string]any  `json:"replyTo,omitempty"`       // Reply information
	FwdFrom        map[string]any  `json:"fwdFrom,omitempty"`       // Forwarded from
	Reactions      []map[string]any `json:"reactions,omitempty"`    // Reactions to message
	Entities       []map[string]any `json:"entities,omitempty"`     // Message entities (formatting)
	Pinned         bool            `json:"pinned,omitempty"`        // Whether message is pinned
	ViaBotID       int64           `json:"viaBotId,omitempty"`      // ID of inline bot
	PostAuthor     string          `json:"postAuthor,omitempty"`    // Author of channel post
	GroupedID      int64           `json:"groupedId,omitempty"`     // Album/media group ID
	TTLPeriod      int             `json:"ttlPeriod,omitempty"`     // Time to live
	Mentioned      bool            `json:"mentioned,omitempty"`     // Whether we were mentioned
	Silent         bool            `json:"silent,omitempty"`        // Silent message (no notification)
	Post           bool            `json:"post,omitempty"`          // Channel post
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
	PeerInfo
	Limit int `json:"limit"`
}

// GetUpdatesResult is the result of GetUpdates.
type GetUpdatesResult struct {
	Updates []StoredUpdate `json:"updates"`
	Count   int            `json:"count"`
}

// ClearMessagesParams holds parameters for ClearMessages.
type ClearMessagesParams struct {
	PeerInfo
	MessageIDs []int64 `json:"messageIds"`
}

// Validate validates ClearMessagesParams.
func (p ClearMessagesParams) Validate() error {
	if err := p.ValidatePeer(); err != nil {
		return err
	}
	if len(p.MessageIDs) == 0 {
		return fmt.Errorf("messageIds is required")
	}
	return nil
}

// ClearMessagesResult is the result of ClearMessages.
type ClearMessagesResult struct {
	Success bool   `json:"success"`
	Cleared int    `json:"cleared"`
	Peer    string `json:"peer"`
}

// ClearHistoryParams holds parameters for ClearHistory.
type ClearHistoryParams struct {
	PeerInfo
	Revoke bool `json:"revoke,omitempty"`
}

// Validate validates ClearHistoryParams.
func (p ClearHistoryParams) Validate() error {
	return p.ValidatePeer()
}

// ClearHistoryResult is the result of ClearHistory.
type ClearHistoryResult struct {
	Success bool   `json:"success"`
	Peer    string `json:"peer"`
	Revoke  bool   `json:"revoke"`
}

// ForwardMessageParams holds parameters for ForwardMessage.
type ForwardMessageParams struct {
	FromPeer string `json:"fromPeer"`
	MessageID int64 `json:"messageId"`
	ToPeer   string `json:"toPeer"`
}

// Validate validates ForwardMessageParams.
func (p ForwardMessageParams) Validate() error {
	if p.FromPeer == "" {
		return fmt.Errorf("fromPeer is required")
	}
	if p.MessageID == 0 {
		return fmt.Errorf("messageId is required")
	}
	if p.ToPeer == "" {
		return fmt.Errorf("toPeer is required")
	}
	return nil
}

// ForwardMessageResult is the result of ForwardMessage.
type ForwardMessageResult struct {
	Success   bool  `json:"success"`
	MessageID int64 `json:"id"`
}
