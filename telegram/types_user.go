// Package telegram provides common types for Telegram client user operations.
package telegram

import "fmt"

// GetMeResult represents the result of GetMe.
type GetMeResult struct {
	ID        int64  `json:"id"`
	Username  string `json:"username"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Phone     string `json:"phone"`
	Verified  bool   `json:"verified"`
	Bot       bool   `json:"bot"`
}

// GetUserInfoParams holds parameters for GetUserInfo.
type GetUserInfoParams struct {
	Username string `json:"username"`
}

// Validate validates GetUserInfoParams.
func (p GetUserInfoParams) Validate() error {
	if p.Username == "" {
		return fmt.Errorf("username is required")
	}
	return nil
}

// GetUserInfoResult is the result of GetUserInfo.
type GetUserInfoResult struct {
	ID        int64  `json:"id"`
	Username  string `json:"username"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Phone     string `json:"phone,omitempty"`
	Bio       string `json:"bio,omitempty"`
	Verified  bool   `json:"verified"`
	Bot       bool   `json:"bot"`
}

// UpdateProfileParams holds parameters for UpdateProfile.
type UpdateProfileParams struct {
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName,omitempty"`
	Bio       string `json:"bio,omitempty"`
}

// Validate validates UpdateProfileParams.
func (p UpdateProfileParams) Validate() error {
	if p.FirstName == "" {
		return fmt.Errorf("firstName is required")
	}
	return nil
}

// UpdateProfileResult is the result of UpdateProfile.
type UpdateProfileResult struct {
	Success bool `json:"success"`
}

// UpdateAvatarParams holds parameters for UpdateAvatar.
type UpdateAvatarParams struct {
	File string `json:"file"`
}

// Validate validates UpdateAvatarParams.
func (p UpdateAvatarParams) Validate() error {
	if p.File == "" {
		return fmt.Errorf("file is required")
	}
	return nil
}

// UpdateAvatarResult is the result of UpdateAvatar.
type UpdateAvatarResult struct {
	Success bool `json:"success"`
}

// BlockPeerParams holds parameters for BlockPeer.
type BlockPeerParams struct {
	PeerInfo
}

// Validate validates BlockPeerParams.
func (p BlockPeerParams) Validate() error {
	return p.ValidatePeer()
}

// BlockPeerResult is the result of BlockPeer.
type BlockPeerResult struct {
	Success bool   `json:"success"`
	Peer    string `json:"peer"`
}

// UnblockPeerParams holds parameters for UnblockPeer.
type UnblockPeerParams struct {
	PeerInfo
}

// Validate validates UnblockPeerParams.
func (p UnblockPeerParams) Validate() error {
	return p.ValidatePeer()
}

// UnblockPeerResult is the result of UnblockPeer.
type UnblockPeerResult struct {
	Success bool   `json:"success"`
	Peer    string `json:"peer"`
}
