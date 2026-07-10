// Package types provides common types for Telegram client chat operations.
package types // revive:disable:var-naming

// GetChatsParams holds parameters for GetChats.
type GetChatsParams struct {
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

// Validate validates GetChatsParams and sets defaults.
func (p *GetChatsParams) Validate() error {
	if p.Limit <= 0 {
		p.Limit = 10
	}
	if p.Limit > 100 {
		p.Limit = 100
	}
	return nil
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
	MessageIDs []int64 `json:"messageIds" validate:"required"`
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

// ClearHistoryResult is the result of ClearHistory.
type ClearHistoryResult struct {
	Success bool   `json:"success"`
	Peer    string `json:"peer"`
	Revoke  bool   `json:"revoke"`
}

// ForwardMessageParams holds parameters for ForwardMessage.
type ForwardMessageParams struct {
	FromPeer  string `json:"fromPeer" validate:"required"`
	MessageID int64  `json:"messageId" validate:"required"`
	ToPeer    string `json:"toPeer" validate:"required"`
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

// PinChatResult is the result of PinChat.
type PinChatResult struct {
	Success bool   `json:"success"`
	Peer    string `json:"peer"`
	Pinned  bool   `json:"pinned"`
}

// ArchiveParams holds parameters for Archive.
type ArchiveParams struct {
	PeerInfo
}

// ArchiveResult is the result of Archive.
type ArchiveResult struct {
	Success bool   `json:"success"`
	Peer    string `json:"peer"`
}

// UnarchiveParams holds parameters for Unarchive.
type UnarchiveParams struct {
	PeerInfo
}

// UnarchiveResult is the result of Unarchive.
type UnarchiveResult struct {
	Success bool   `json:"success"`
	Peer    string `json:"peer"`
}

// MuteParams holds parameters for Mute.
type MuteParams struct {
	PeerInfo
}

// MuteResult is the result of Mute.
type MuteResult struct {
	Success bool   `json:"success"`
	Peer    string `json:"peer"`
}

// UnmuteParams holds parameters for Unmute.
type UnmuteParams struct {
	PeerInfo
}

// UnmuteResult is the result of Unmute.
type UnmuteResult struct {
	Success bool   `json:"success"`
	Peer    string `json:"peer"`
}
