// Package telegram provides common types for Telegram client message operations.
package telegram

// SendReplyParams holds parameters for SendReply.
type SendReplyParams struct {
	Peer      string `json:"peer,omitempty"`
	Username  string `json:"username,omitempty"`
	MessageID int64  `json:"messageId"`
	Text      string `json:"text"`
}

// SendReplyResult is the result of SendReply.
type SendReplyResult struct {
	ID      int64  `json:"id"`
	Date    int64  `json:"date"`
	Peer    string `json:"peer"`
	Text    string `json:"text"`
	ReplyTo int64  `json:"replyTo"`
}

// UpdateMessageParams holds parameters for UpdateMessage.
type UpdateMessageParams struct {
	Peer      string `json:"peer,omitempty"`
	Username  string `json:"username,omitempty"`
	MessageID int64  `json:"messageId"`
	Text      string `json:"text"`
}

// UpdateMessageResult is the result of UpdateMessage.
type UpdateMessageResult struct {
	Success   bool  `json:"success"`
	MessageID int64 `json:"messageId"`
}

// DeleteMessageParams holds parameters for DeleteMessage.
type DeleteMessageParams struct {
	Peer      string `json:"peer,omitempty"`
	Username  string `json:"username,omitempty"`
	MessageID int64  `json:"messageId"`
}

// DeleteMessageResult is the result of DeleteMessage.
type DeleteMessageResult struct {
	Success   bool  `json:"success"`
	MessageID int64 `json:"messageId"`
}

// PinMessageParams holds parameters for PinMessage.
type PinMessageParams struct {
	Peer      string `json:"peer,omitempty"`
	Username  string `json:"username,omitempty"`
	MessageID int64  `json:"messageId"`
}

// PinMessageResult is the result of PinMessage.
type PinMessageResult struct {
	Success   bool  `json:"success"`
	MessageID int64 `json:"messageId"`
}

// UnpinMessageParams holds parameters for UnpinMessage.
type UnpinMessageParams struct {
	Peer      string `json:"peer,omitempty"`
	Username  string `json:"username,omitempty"`
	MessageID int64  `json:"messageId"`
}

// UnpinMessageResult is the result of UnpinMessage.
type UnpinMessageResult struct {
	Success   bool  `json:"success"`
	MessageID int64 `json:"messageId"`
}

// InlineButton represents an inline button.
type InlineButton struct {
	Text  string `json:"text"`
	Data  string `json:"data,omitempty"`
	Index int    `json:"index"`
}

// InspectInlineButtonsParams holds parameters for InspectInlineButtons.
type InspectInlineButtonsParams struct {
	Peer      string `json:"peer,omitempty"`
	Username  string `json:"username,omitempty"`
	MessageID int64  `json:"messageId"`
	Limit     int    `json:"limit,omitempty"`
}

// InspectInlineButtonsResult is the result of InspectInlineButtons.
type InspectInlineButtonsResult struct {
	MessageID int64          `json:"messageId"`
	Buttons   []InlineButton `json:"buttons"`
}

// PressInlineButtonParams holds parameters for PressInlineButton.
type PressInlineButtonParams struct {
	Peer        string `json:"peer,omitempty"`
	Username    string `json:"username,omitempty"`
	MessageID   int64  `json:"messageId"`
	ButtonText  string `json:"buttonText,omitempty"`
	ButtonIndex int    `json:"buttonIndex"`
}

// PressInlineButtonResult is the result of PressInlineButton.
type PressInlineButtonResult struct {
	Success   bool  `json:"success"`
	MessageID int64 `json:"messageId"`
}

// AddReactionParams holds parameters for AddReaction.
type AddReactionParams struct {
	Peer      string `json:"peer,omitempty"`
	Username  string `json:"username,omitempty"`
	MessageID int64  `json:"messageId"`
	Emoji     string `json:"emoji"`
	Big       bool   `json:"big,omitempty"`
}

// AddReactionResult is the result of AddReaction.
type AddReactionResult struct {
	Success   bool  `json:"success"`
	MessageID int64 `json:"messageId"`
	Emoji     string `json:"emoji"`
}

// RemoveReactionParams holds parameters for RemoveReaction.
type RemoveReactionParams struct {
	Peer      string `json:"peer,omitempty"`
	Username  string `json:"username,omitempty"`
	MessageID int64  `json:"messageId"`
}

// RemoveReactionResult is the result of RemoveReaction.
type RemoveReactionResult struct {
	Success   bool  `json:"success"`
	MessageID int64 `json:"messageId"`
}

// Reaction represents a reaction on a message.
type Reaction struct {
	Emoji      string `json:"emoji"`
	Count      int    `json:"count"`
	FromMe     bool   `json:"fromMe"`
	UserIDs    []int64 `json:"userIds,omitempty"`
}

// ListReactionsParams holds parameters for ListReactions.
type ListReactionsParams struct {
	Peer      string `json:"peer,omitempty"`
	Username  string `json:"username,omitempty"`
	MessageID int64  `json:"messageId"`
	Limit     int    `json:"limit,omitempty"`
}

// ListReactionsResult is the result of ListReactions.
type ListReactionsResult struct {
	MessageID int64      `json:"messageId"`
	Reactions []Reaction `json:"reactions"`
}
