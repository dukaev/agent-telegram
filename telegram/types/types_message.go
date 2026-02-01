// Package types provides common types for Telegram client message operations.
package types // revive:disable:var-naming

import "fmt"

// SendReplyParams holds parameters for SendReply.
type SendReplyParams struct {
	PeerInfo
	MsgID
	Text string `json:"text" validate:"required"`
}

// Validate validates SendReplyParams.
func (p SendReplyParams) Validate() error {
	return ValidateStruct(p)
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
	PeerInfo
	MsgID
	Text string `json:"text" validate:"required"`
}

// Validate validates UpdateMessageParams.
func (p UpdateMessageParams) Validate() error {
	return ValidateStruct(p)
}

// UpdateMessageResult is the result of UpdateMessage.
type UpdateMessageResult struct {
	Success   bool  `json:"success"`
	MessageID int64 `json:"messageId"`
}

// DeleteMessageParams holds parameters for DeleteMessage.
type DeleteMessageParams struct {
	PeerInfo
	MsgID
}

// Validate validates DeleteMessageParams.
func (p DeleteMessageParams) Validate() error {
	return ValidateStruct(p)
}

// DeleteMessageResult is the result of DeleteMessage.
type DeleteMessageResult struct {
	Success   bool  `json:"success"`
	MessageID int64 `json:"messageId"`
}

// PinMessageParams holds parameters for PinMessage.
type PinMessageParams struct {
	PeerInfo
	MsgID
}

// Validate validates PinMessageParams.
func (p PinMessageParams) Validate() error {
	return ValidateStruct(p)
}

// PinMessageResult is the result of PinMessage.
type PinMessageResult struct {
	Success   bool  `json:"success"`
	MessageID int64 `json:"messageId"`
}

// UnpinMessageParams holds parameters for UnpinMessage.
type UnpinMessageParams struct {
	PeerInfo
	MsgID
}

// Validate validates UnpinMessageParams.
func (p UnpinMessageParams) Validate() error {
	return ValidateStruct(p)
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
	PeerInfo
	MsgID
	Limit int `json:"limit,omitempty"`
}

// Validate validates InspectInlineButtonsParams.
func (p InspectInlineButtonsParams) Validate() error {
	return ValidateStruct(p)
}

// InspectInlineButtonsResult is the result of InspectInlineButtons.
type InspectInlineButtonsResult struct {
	MessageID int64          `json:"messageId"`
	Buttons   []InlineButton `json:"buttons"`
}

// PressInlineButtonParams holds parameters for PressInlineButton.
type PressInlineButtonParams struct {
	PeerInfo
	MsgID
	ButtonText  string `json:"buttonText,omitempty"`
	ButtonIndex int    `json:"buttonIndex"`
}

// Validate validates PressInlineButtonParams.
func (p PressInlineButtonParams) Validate() error {
	if err := ValidateStruct(p); err != nil {
		return err
	}
	if p.ButtonIndex < 0 {
		return fmt.Errorf("buttonIndex must be >= 0")
	}
	return nil
}

// PressInlineButtonResult is the result of PressInlineButton.
type PressInlineButtonResult struct {
	Success   bool  `json:"success"`
	MessageID int64 `json:"messageId"`
}

// ReadMessagesParams holds parameters for ReadMessages.
type ReadMessagesParams struct {
	Peer  string `json:"peer" validate:"required"`
	MaxID int64  `json:"maxId,omitempty"` // Mark all messages up to this ID as read
}

// Validate validates ReadMessagesParams.
func (p ReadMessagesParams) Validate() error {
	return ValidateStruct(p)
}

// ReadMessagesResult is the result of ReadMessages.
type ReadMessagesResult struct {
	Success bool  `json:"success"`
	MaxID   int64 `json:"maxId"`
}

// SetTypingParams holds parameters for SetTyping.
type SetTypingParams struct {
	Peer   string `json:"peer" validate:"required"`
	Action string `json:"action,omitempty"` // typing, upload_photo, record_video, record_audio, etc.
}

// Validate validates SetTypingParams.
func (p SetTypingParams) Validate() error {
	return ValidateStruct(p)
}

// SetTypingResult is the result of SetTyping.
type SetTypingResult struct {
	Success bool `json:"success"`
}

// GetScheduledMessagesParams holds parameters for GetScheduledMessages.
type GetScheduledMessagesParams struct {
	Peer string `json:"peer" validate:"required"`
}

// Validate validates GetScheduledMessagesParams.
func (p GetScheduledMessagesParams) Validate() error {
	return ValidateStruct(p)
}

// ScheduledMessage represents a scheduled message.
type ScheduledMessage struct {
	ID      int64  `json:"id"`
	Date    int64  `json:"date"` // Scheduled send time
	Message string `json:"message,omitempty"`
	Peer    string `json:"peer"`
}

// GetScheduledMessagesResult is the result of GetScheduledMessages.
type GetScheduledMessagesResult struct {
	Messages []ScheduledMessage `json:"messages"`
	Count    int                `json:"count"`
}
