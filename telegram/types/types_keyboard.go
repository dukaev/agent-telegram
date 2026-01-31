// Package types provides common types for Telegram client.
package types

// ReplyKeyboardResult is the result of InspectReplyKeyboard.
type ReplyKeyboardResult struct {
	Peer      string       `json:"peer"`
	MessageID int64        `json:"message_id,omitempty"`
	Found     bool         `json:"found"`
	Keyboard  ReplyKeyboard `json:"keyboard,omitempty"`
	ForceReply bool        `json:"force_reply,omitempty"`
	Hidden    bool         `json:"hidden,omitempty"`
}

// ReplyKeyboard represents a reply keyboard.
type ReplyKeyboard struct {
	Resize     bool                `json:"resize,omitempty"`
	SingleUse  bool                `json:"single_use,omitempty"`
	Selective  bool                `json:"selective,omitempty"`
	Persistent bool                `json:"persistent,omitempty"`
	Placeholder string             `json:"placeholder,omitempty"`
	Rows       [][]KeyboardButton  `json:"rows"`
}

// KeyboardButton represents a keyboard button.
type KeyboardButton struct {
	Text         string `json:"text"`
	Type         string `json:"type"`
	URL          string `json:"url,omitempty"`
	PollType     string `json:"poll_type,omitempty"`
	ButtonID     int    `json:"button_id,omitempty"`
	MaxQuantity  int    `json:"max_quantity,omitempty"`
	UserID       int64  `json:"user_id,omitempty"`
}
