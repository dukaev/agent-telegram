// Package types provides common types for Telegram client join operations.
package types

// JoinChatParams holds parameters for JoinChat.
type JoinChatParams struct {
	InviteLink string `json:"inviteLink" validate:"required"`
}

// Validate validates JoinChatParams.
func (p JoinChatParams) Validate() error {
	return ValidateStruct(p)
}

// JoinChatResult is the result of JoinChat.
type JoinChatResult struct {
	Success bool   `json:"success"`
	ChatID  int64  `json:"chatId"`
	Title   string `json:"title"`
}

// SubscribeChannelParams holds parameters for SubscribeChannel.
type SubscribeChannelParams struct {
	Channel string `json:"channel" validate:"required"` // @username or username
}

// Validate validates SubscribeChannelParams.
func (p SubscribeChannelParams) Validate() error {
	return ValidateStruct(p)
}

// SubscribeChannelResult is the result of SubscribeChannel.
type SubscribeChannelResult struct {
	Success bool   `json:"success"`
	ChatID  int64  `json:"chatId"`
	Title   string `json:"title"`
}

// ForumTopic represents a forum topic.
type ForumTopic struct {
	ID        int64  `json:"id"`                   // Topic ID
	Title     string `json:"title"`                // Topic title
	IconColor int32  `json:"iconColor,omitempty"`  // Icon color
	IconEmoji string `json:"iconEmoji,omitempty"`  // Icon emoji
	Top       bool   `json:"top,omitempty"`        // Whether topic is pinned
	Closed    bool   `json:"closed,omitempty"`     // Whether topic is closed
}

// GetTopicsParams holds parameters for GetTopics.
type GetTopicsParams struct {
	Peer  string `json:"peer" validate:"required"` // Channel username or ID
	Limit int    `json:"limit,omitempty"`          // Maximum number of topics to return
}

// Validate validates GetTopicsParams.
func (p GetTopicsParams) Validate() error {
	return ValidateStruct(p)
}

// GetTopicsResult is the result of GetTopics.
type GetTopicsResult struct {
	Peer   string       `json:"peer"`
	Topics []ForumTopic `json:"topics"`
	Count  int          `json:"count"`
}
