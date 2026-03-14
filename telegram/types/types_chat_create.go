// Package types provides common types for Telegram client chat creation operations.
package types // revive:disable:var-naming

import "fmt"

// CreateGroupParams holds parameters for CreateGroup.
type CreateGroupParams struct {
	Title   string   `json:"title" validate:"required"`   // Group title
	Members []string `json:"members" validate:"required"` // List of usernames to add
}

// Validate validates CreateGroupParams.
func (p CreateGroupParams) Validate() error {
	if len(p.Members) == 0 {
		return fmt.Errorf("at least one member is required")
	}
	return nil
}

// CreateGroupResult is the result of CreateGroup.
type CreateGroupResult struct {
	Success bool   `json:"success"`
	ChatID  int64  `json:"chatId"`
	Title   string `json:"title"`
}

// CreateChannelParams holds parameters for CreateChannel.
type CreateChannelParams struct {
	NoValidation
	Title       string `json:"title" validate:"required"` // Channel title
	Description string `json:"description,omitempty"`     // Channel description
	Username    string `json:"username,omitempty"`        // Channel username (optional)
	Megagroup   bool   `json:"megagroup,omitempty"`       // Create as supergroup instead of channel
}

// CreateChannelResult is the result of CreateChannel.
type CreateChannelResult struct {
	Success bool   `json:"success"`
	ChatID  int64  `json:"chatId"`
	Title   string `json:"title"`
}

// EditTitleParams holds parameters for EditTitle.
type EditTitleParams struct {
	NoValidation
	Peer  string `json:"peer" validate:"required"`  // Chat/channel username or ID
	Title string `json:"title" validate:"required"` // New title
}

// EditTitleResult is the result of EditTitle.
type EditTitleResult struct {
	Success bool   `json:"success"`
	Title   string `json:"title"`
}

// SetPhotoParams holds parameters for SetPhoto.
type SetPhotoParams struct {
	NoValidation
	Peer string `json:"peer" validate:"required"` // Chat/channel username or ID
	File string `json:"file" validate:"required"` // Path to photo file
}

// SetPhotoResult is the result of SetPhoto.
type SetPhotoResult struct {
	Success bool `json:"success"`
}

// DeletePhotoParams holds parameters for DeletePhoto.
type DeletePhotoParams struct {
	NoValidation
	Peer string `json:"peer" validate:"required"` // Chat/channel username or ID
}

// DeletePhotoResult is the result of DeletePhoto.
type DeletePhotoResult struct {
	Success bool `json:"success"`
}
