package types

import "fmt"

// PromoteAdminParams holds parameters for PromoteAdmin.
type PromoteAdminParams struct {
	Peer              string `json:"peer"`                        // Chat/channel username or ID
	User              string `json:"user"`                        // Username to promote
	CanChangeInfo     bool   `json:"canChangeInfo,omitempty"`     // Can change chat info
	CanPostMessages   bool   `json:"canPostMessages,omitempty"`   // Can post messages
	CanEditMessages   bool   `json:"canEditMessages,omitempty"`   // Can edit messages
	CanDeleteMessages bool   `json:"canDeleteMessages,omitempty"` // Can delete messages
	CanBanUsers       bool   `json:"canBanUsers,omitempty"`       // Can ban users
	CanInviteUsers    bool   `json:"canInviteUsers,omitempty"`    // Can invite users
	CanPinMessages    bool   `json:"canPinMessages,omitempty"`    // Can pin messages
	CanAddAdmins      bool   `json:"canAddAdmins,omitempty"`      // Can add new admins
	Anonymous         bool   `json:"anonymous,omitempty"`         // Is anonymous admin
}

// Validate validates PromoteAdminParams.
func (p PromoteAdminParams) Validate() error {
	if p.Peer == "" {
		return fmt.Errorf("peer is required")
	}
	if p.User == "" {
		return fmt.Errorf("user is required")
	}
	return nil
}

// PromoteAdminResult is the result of PromoteAdmin.
type PromoteAdminResult struct {
	Success bool `json:"success"`
}

// DemoteAdminParams holds parameters for DemoteAdmin.
type DemoteAdminParams struct {
	Peer string `json:"peer"` // Chat/channel username or ID
	User string `json:"user"` // Username to demote
}

// Validate validates DemoteAdminParams.
func (p DemoteAdminParams) Validate() error {
	if p.Peer == "" {
		return fmt.Errorf("peer is required")
	}
	if p.User == "" {
		return fmt.Errorf("user is required")
	}
	return nil
}

// DemoteAdminResult is the result of DemoteAdmin.
type DemoteAdminResult struct {
	Success bool `json:"success"`
}

// GetInviteLinkParams holds parameters for GetInviteLink.
type GetInviteLinkParams struct {
	Peer      string `json:"peer"`                // Chat/channel username or ID
	CreateNew bool   `json:"createNew,omitempty"` // Create a new link
}

// Validate validates GetInviteLinkParams.
func (p GetInviteLinkParams) Validate() error {
	if p.Peer == "" {
		return fmt.Errorf("peer is required")
	}
	return nil
}

// GetInviteLinkResult is the result of GetInviteLink.
type GetInviteLinkResult struct {
	Link          string `json:"link"`
	Usage         int    `json:"usage,omitempty"`
	UsageLimit    int    `json:"usageLimit,omitempty"`
	RequestNeeded bool   `json:"requestNeeded,omitempty"`
	Expired       bool   `json:"expired,omitempty"`
}
