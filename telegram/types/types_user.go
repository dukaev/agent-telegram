// Package types provides common types for Telegram client user operations.
package types // revive:disable:var-naming

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
	Username string `json:"username,omitempty"`
	UserID   int64  `json:"userId,omitempty"`
}

// Validate validates GetUserInfoParams.
func (p GetUserInfoParams) Validate() error {
	if p.Username == "" && p.UserID == 0 {
		return ErrUsernameOrIDRequired
	}
	return nil
}

// ErrUsernameOrIDRequired is returned when neither username nor userId is provided.
var ErrUsernameOrIDRequired = fmt.Errorf("username or userId is required")

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
	NoValidation
	FirstName string `json:"firstName" validate:"required"`
	LastName  string `json:"lastName,omitempty"`
	Bio       string `json:"bio,omitempty"`
}

// UpdateProfileResult is the result of UpdateProfile.
type UpdateProfileResult struct {
	Success bool `json:"success"`
}

// UpdateAvatarParams holds parameters for UpdateAvatar.
type UpdateAvatarParams struct {
	NoValidation
	File string `json:"file" validate:"required"`
}

// UpdateAvatarResult is the result of UpdateAvatar.
type UpdateAvatarResult struct {
	Success bool `json:"success"`
}

// BlockPeerParams holds parameters for BlockPeer.
type BlockPeerParams struct {
	PeerInfo
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

// UnblockPeerResult is the result of UnblockPeer.
type UnblockPeerResult struct {
	Success bool   `json:"success"`
	Peer    string `json:"peer"`
}
