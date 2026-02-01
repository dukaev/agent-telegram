// Package types provides common types for Telegram client chat operations.
package types // revive:disable:var-naming

import "fmt"

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
	FromPeer  string `json:"fromPeer"`
	MessageID int64  `json:"messageId"`
	ToPeer    string `json:"toPeer"`
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

// PinChatParams holds parameters for PinChat (pin chat in dialog list).
type PinChatParams struct {
	PeerInfo
	Disable bool `json:"disable"` // true to unpin, false to pin
}

// Validate validates PinChatParams.
func (p PinChatParams) Validate() error {
	return p.ValidatePeer()
}

// PinChatResult is the result of PinChat.
type PinChatResult struct {
	Success bool   `json:"success"`
	Peer    string `json:"peer"`
	Pinned  bool   `json:"pinned"`
}
