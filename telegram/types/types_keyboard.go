// Package types provides shared parameter and result types for Telegram operations.
package types

// ReplyKeyboardResult is the result of InspectReplyKeyboard.
type ReplyKeyboardResult struct {
	Peer      string       `json:"peer"`
	MessageID int64        `json:"messageId,omitempty"`
	Found     bool         `json:"found"`
	Keyboard  ReplyKeyboard `json:"keyboard,omitempty"`
	ForceReply bool        `json:"forceReply,omitempty"`
	Hidden    bool         `json:"hidden,omitempty"`
}

// ReplyKeyboard represents a reply keyboard.
type ReplyKeyboard struct {
	Resize     bool                `json:"resize,omitempty"`
	SingleUse  bool                `json:"singleUse,omitempty"`
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
	PollType     string `json:"pollType,omitempty"`
	ButtonID     int    `json:"buttonId,omitempty"`
	MaxQuantity  int    `json:"maxQuantity,omitempty"`
	UserID       int64  `json:"userId,omitempty"`
}
