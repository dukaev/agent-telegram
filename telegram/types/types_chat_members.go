// Package types provides common types for Telegram client member operations.
package types

import "fmt"

// LeaveParams holds parameters for Leave.
type LeaveParams struct {
	NoValidation
	Peer string `json:"peer" validate:"required"` // Chat/channel username or ID
}

// LeaveResult is the result of Leave.
type LeaveResult struct {
	Success bool `json:"success"`
}

// InviteParams holds parameters for Invite.
type InviteParams struct {
	Peer    string   `json:"peer" validate:"required"`    // Chat/channel username or ID
	Members []string `json:"members" validate:"required"` // List of usernames to invite
}

// Validate validates InviteParams.
func (p InviteParams) Validate() error {
	if len(p.Members) == 0 {
		return fmt.Errorf("at least one member is required")
	}
	return nil
}

// InviteResult is the result of Invite.
type InviteResult struct {
	Success bool `json:"success"`
}

// Participant represents a chat participant.
type Participant struct {
	ID        int64  `json:"id"`
	FirstName string `json:"firstName,omitempty"`
	LastName  string `json:"lastName,omitempty"`
	Username  string `json:"username,omitempty"`
	Bot       bool   `json:"bot,omitempty"`
	Admin     bool   `json:"admin,omitempty"`
	Creator   bool   `json:"creator,omitempty"`
	Peer      string `json:"peer,omitempty"`
}

// GetParticipantsParams holds parameters for GetParticipants.
type GetParticipantsParams struct {
	NoValidation
	Peer  string `json:"peer" validate:"required"` // Chat/channel username or ID
	Limit int    `json:"limit,omitempty"`          // Maximum number of participants (default 100)
}

// GetParticipantsResult is the result of GetParticipants.
type GetParticipantsResult struct {
	Peer         string        `json:"peer"`
	Participants []Participant `json:"participants"`
	Count        int           `json:"count"`
}

// GetAdminsParams holds parameters for GetAdmins.
type GetAdminsParams struct {
	NoValidation
	Peer  string `json:"peer" validate:"required"` // Chat/channel username or ID
	Limit int    `json:"limit,omitempty"`          // Maximum number of admins (default 100)
}

// GetAdminsResult is the result of GetAdmins.
type GetAdminsResult struct {
	Peer   string        `json:"peer"`
	Admins []Participant `json:"admins"`
	Count  int           `json:"count"`
}

// GetBannedParams holds parameters for GetBanned.
type GetBannedParams struct {
	NoValidation
	Peer  string `json:"peer" validate:"required"` // Chat/channel username or ID
	Limit int    `json:"limit,omitempty"`          // Maximum number of banned users (default 100)
}

// GetBannedResult is the result of GetBanned.
type GetBannedResult struct {
	Peer   string        `json:"peer"`
	Banned []Participant `json:"banned"`
	Count  int           `json:"count"`
}
