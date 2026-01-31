// Package telegram provides common types for Telegram client user operations.
package telegram

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

// UpdateProfileResult is the result of UpdateProfile.
type UpdateProfileResult struct {
	Success bool `json:"success"`
}

// UpdateAvatarParams holds parameters for UpdateAvatar.
type UpdateAvatarParams struct {
	File string `json:"file"`
}

// UpdateAvatarResult is the result of UpdateAvatar.
type UpdateAvatarResult struct {
	Success bool `json:"success"`
}

// BlockPeerParams holds parameters for BlockPeer.
type BlockPeerParams struct {
	Peer     string `json:"peer,omitempty"`
	Username string `json:"username,omitempty"`
}

// BlockPeerResult is the result of BlockPeer.
type BlockPeerResult struct {
	Success bool   `json:"success"`
	Peer    string `json:"peer"`
}

// UnblockPeerParams holds parameters for UnblockPeer.
type UnblockPeerParams struct {
	Peer     string `json:"peer,omitempty"`
	Username string `json:"username,omitempty"`
}

// UnblockPeerResult is the result of UnblockPeer.
type UnblockPeerResult struct {
	Success bool   `json:"success"`
	Peer    string `json:"peer"`
}
