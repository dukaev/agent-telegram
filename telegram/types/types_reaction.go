// Package types provides common types for Telegram client reaction operations.
package types // revive:disable:var-naming

import "fmt"

// AddReactionParams holds parameters for AddReaction.
type AddReactionParams struct {
	PeerInfo
	MsgID
	Emoji string `json:"emoji"`
	Big   bool   `json:"big,omitempty"`
}

// Validate validates AddReactionParams.
func (p AddReactionParams) Validate() error {
	if err := p.ValidatePeer(); err != nil {
		return err
	}
	if err := p.ValidateMessageID(); err != nil {
		return err
	}
	if p.Emoji == "" {
		return fmt.Errorf("emoji is required")
	}
	return nil
}

// AddReactionResult is the result of AddReaction.
type AddReactionResult struct {
	Success   bool  `json:"success"`
	MessageID int64 `json:"messageId"`
	Emoji     string `json:"emoji"`
}

// RemoveReactionParams holds parameters for RemoveReaction.
type RemoveReactionParams struct {
	PeerInfo
	MsgID
}

// Validate validates RemoveReactionParams.
func (p RemoveReactionParams) Validate() error {
	if err := p.ValidatePeer(); err != nil {
		return err
	}
	if err := p.ValidateMessageID(); err != nil {
		return err
	}
	return nil
}

// RemoveReactionResult is the result of RemoveReaction.
type RemoveReactionResult struct {
	Success   bool  `json:"success"`
	MessageID int64 `json:"messageId"`
}

// Reaction represents a reaction on a message.
type Reaction struct {
	Emoji   string `json:"emoji"`
	Count   int    `json:"count"`
	FromMe  bool   `json:"fromMe"`
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
	if err := p.ValidatePeer(); err != nil {
		return err
	}
	if err := p.ValidateMessageID(); err != nil {
		return err
	}
	return nil
}

// ListReactionsResult is the result of ListReactions.
type ListReactionsResult struct {
	MessageID int64      `json:"messageId"`
	Reactions []Reaction `json:"reactions"`
}
