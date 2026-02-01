// Package types provides common types for Telegram client reaction operations.
package types // revive:disable:var-naming

// AddReactionParams holds parameters for AddReaction.
type AddReactionParams struct {
	PeerInfo
	MsgID
	Emoji string `json:"emoji" validate:"required"`
	Big   bool   `json:"big,omitempty"`
}

// Validate validates AddReactionParams.
func (p AddReactionParams) Validate() error {
	return ValidateStruct(p)
}

// AddReactionResult is the result of AddReaction.
type AddReactionResult struct {
	Success   bool   `json:"success"`
	MessageID int64  `json:"messageId"`
	Emoji     string `json:"emoji"`
}

// RemoveReactionParams holds parameters for RemoveReaction.
type RemoveReactionParams struct {
	PeerInfo
	MsgID
}

// Validate validates RemoveReactionParams.
func (p RemoveReactionParams) Validate() error {
	return ValidateStruct(p)
}

// RemoveReactionResult is the result of RemoveReaction.
type RemoveReactionResult struct {
	Success   bool  `json:"success"`
	MessageID int64 `json:"messageId"`
}

// Reaction represents a reaction on a message.
type Reaction struct {
	Emoji   string  `json:"emoji"`
	Count   int     `json:"count"`
	FromMe  bool    `json:"fromMe"`
	UserIDs []int64 `json:"userIds,omitempty"`
}

// ListReactionsParams holds parameters for ListReactions.
type ListReactionsParams struct {
	PeerInfo
	MsgID
	Limit int `json:"limit,omitempty"`
}

// Validate validates ListReactionsParams.
func (p ListReactionsParams) Validate() error {
	return ValidateStruct(p)
}

// ListReactionsResult is the result of ListReactions.
type ListReactionsResult struct {
	MessageID int64      `json:"messageId"`
	Reactions []Reaction `json:"reactions"`
}
