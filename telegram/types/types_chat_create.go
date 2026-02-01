package types

import "fmt"

// CreateGroupParams holds parameters for CreateGroup.
type CreateGroupParams struct {
	Title   string   `json:"title"`   // Group title
	Members []string `json:"members"` // List of usernames to add
}

// Validate validates CreateGroupParams.
func (p CreateGroupParams) Validate() error {
	if p.Title == "" {
		return fmt.Errorf("title is required")
	}
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
	Title       string `json:"title"`                // Channel title
	Description string `json:"description,omitempty"` // Channel description
	Username    string `json:"username,omitempty"`    // Channel username (optional)
	Megagroup   bool   `json:"megagroup,omitempty"`   // Create as supergroup instead of channel
}

// Validate validates CreateChannelParams.
func (p CreateChannelParams) Validate() error {
	if p.Title == "" {
		return fmt.Errorf("title is required")
	}
	return nil
}

// CreateChannelResult is the result of CreateChannel.
type CreateChannelResult struct {
	Success bool   `json:"success"`
	ChatID  int64  `json:"chatId"`
	Title   string `json:"title"`
}

// EditTitleParams holds parameters for EditTitle.
type EditTitleParams struct {
	Peer  string `json:"peer"`  // Chat/channel username or ID
	Title string `json:"title"` // New title
}

// Validate validates EditTitleParams.
func (p EditTitleParams) Validate() error {
	if p.Peer == "" {
		return fmt.Errorf("peer is required")
	}
	if p.Title == "" {
		return fmt.Errorf("title is required")
	}
	return nil
}

// EditTitleResult is the result of EditTitle.
type EditTitleResult struct {
	Success bool   `json:"success"`
	Title   string `json:"title"`
}

// SetPhotoParams holds parameters for SetPhoto.
type SetPhotoParams struct {
	Peer string `json:"peer"` // Chat/channel username or ID
	File string `json:"file"` // Path to photo file
}

// Validate validates SetPhotoParams.
func (p SetPhotoParams) Validate() error {
	if p.Peer == "" {
		return fmt.Errorf("peer is required")
	}
	if p.File == "" {
		return fmt.Errorf("file is required")
	}
	return nil
}

// SetPhotoResult is the result of SetPhoto.
type SetPhotoResult struct {
	Success bool `json:"success"`
}

// DeletePhotoParams holds parameters for DeletePhoto.
type DeletePhotoParams struct {
	Peer string `json:"peer"` // Chat/channel username or ID
}

// Validate validates DeletePhotoParams.
func (p DeletePhotoParams) Validate() error {
	if p.Peer == "" {
		return fmt.Errorf("peer is required")
	}
	return nil
}

// DeletePhotoResult is the result of DeletePhoto.
type DeletePhotoResult struct {
	Success bool `json:"success"`
}
