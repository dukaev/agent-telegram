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
	ID        int64                  `json:"id"`
	Type      UpdateType             `json:"type"`
	Timestamp time.Time              `json:"timestamp"`
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
	Limit  int                      `json:"limit"`
	Offset int                      `json:"offset"`
	Count  int                      `json:"count"`
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

// GetMeResult represents the result of GetMe.
type GetMeResult struct {
	ID        int64  `json:"id"`
	Username  string `json:"username"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Phone     string `json:"phone"`
	Verified  bool   `json:"verified"`
	Bot       bool   `json:"bot"`
}

// SendMessageParams holds parameters for SendMessage.
type SendMessageParams struct {
	Peer    string `json:"peer"`
	Message string `json:"message"`
}

// SendMessageResult is the result of SendMessage.
type SendMessageResult struct {
	ID      int64  `json:"id"`
	Date    int64  `json:"date"`
	Message string `json:"message"`
	Peer    string `json:"peer"`
}

// SendLocationParams holds parameters for SendLocation.
type SendLocationParams struct {
	Peer    string  `json:"peer"`
	Latitude float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

// SendLocationResult is the result of SendLocation.
type SendLocationResult struct {
	ID       int64   `json:"id"`
	Date     int64   `json:"date"`
	Peer     string  `json:"peer"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

// SendPhotoParams holds parameters for SendPhoto.
type SendPhotoParams struct {
	Peer    string `json:"peer"`
	File    string `json:"file"`
	Caption string `json:"caption,omitempty"`
}

// SendPhotoResult is the result of SendPhoto.
type SendPhotoResult struct {
	ID      int64  `json:"id"`
	Date    int64  `json:"date"`
	Peer    string `json:"peer"`
	Caption string `json:"caption,omitempty"`
}

// SendContactParams holds parameters for SendContact.
type SendContactParams struct {
	Peer      string `json:"peer"`
	Phone     string `json:"phone"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName,omitempty"`
}

// SendContactResult is the result of SendContact.
type SendContactResult struct {
	ID      int64  `json:"id"`
	Date    int64  `json:"date"`
	Peer    string `json:"peer"`
	Phone   string `json:"phone"`
}

// ClearMessagesParams holds parameters for ClearMessages.
type ClearMessagesParams struct {
	Peer      string `json:"peer"`
	MessageIDs []int64 `json:"messageIds"`
}

// ClearMessagesResult is the result of ClearMessages.
type ClearMessagesResult struct {
	Success  bool   `json:"success"`
	Cleared  int    `json:"cleared"`
	Peer     string `json:"peer"`
}

// ClearHistoryParams holds parameters for ClearHistory.
type ClearHistoryParams struct {
	Peer     string `json:"peer"`
	Revoke   bool   `json:"revoke,omitempty"`
}

// ClearHistoryResult is the result of ClearHistory.
type ClearHistoryResult struct {
	Success   bool   `json:"success"`
	Peer      string `json:"peer"`
	Revoke    bool   `json:"revoke"`
}

// BlockPeerParams holds parameters for BlockPeer.
type BlockPeerParams struct {
	Peer string `json:"peer"`
}

// BlockPeerResult is the result of BlockPeer.
type BlockPeerResult struct {
	Success bool   `json:"success"`
	Peer    string `json:"peer"`
}

// SendFileParams holds parameters for SendFile.
type SendFileParams struct {
	Peer    string `json:"peer"`
	File    string `json:"file"`
	Caption string `json:"caption,omitempty"`
}

// SendFileResult is the result of SendFile.
type SendFileResult struct {
	ID      int64  `json:"id"`
	Date    int64  `json:"date"`
	Peer    string `json:"peer"`
	Caption string `json:"caption,omitempty"`
}

// PollOption represents a poll option.
type PollOption struct {
	Text string `json:"text"`
}

// SendPollParams holds parameters for SendPoll.
type SendPollParams struct {
	Peer       string       `json:"peer"`
	Question   string       `json:"question"`
	Options    []PollOption `json:"options"`
	Anonymous  bool         `json:"anonymous,omitempty"`
	Quiz       bool         `json:"quiz,omitempty"`
	CorrectIdx int          `json:"correctIdx,omitempty"`
}

// SendPollResult is the result of SendPoll.
type SendPollResult struct {
	ID       int64  `json:"id"`
	Date     int64  `json:"date"`
	Peer     string `json:"peer"`
	Question string `json:"question"`
}

// SendVideoParams holds parameters for SendVideo.
type SendVideoParams struct {
	Peer    string `json:"peer"`
	File    string `json:"file"`
	Caption string `json:"caption,omitempty"`
}

// SendVideoResult is the result of SendVideo.
type SendVideoResult struct {
	ID      int64  `json:"id"`
	Date    int64  `json:"date"`
	Peer    string `json:"peer"`
	Caption string `json:"caption,omitempty"`
}
