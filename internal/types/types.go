// Package types provides shared types and interfaces for the agent-telegram application.
package types

// PeerType represents the type of a Telegram peer (user, chat, or channel).
type PeerType int

const (
	// PeerTypeUser represents a user peer.
	PeerTypeUser PeerType = iota
	// PeerTypeChat represents a basic chat peer.
	PeerTypeChat
	// PeerTypeChannel represents a channel or supergroup peer.
	PeerTypeChannel
)

// String returns a string representation of the peer type.
func (p PeerType) String() string {
	switch p {
	case PeerTypeUser:
		return "user"
	case PeerTypeChat:
		return "chat"
	case PeerTypeChannel:
		return "channel"
	default:
		return "unknown"
	}
}

// AuthStatus represents the authentication status of a user.
type AuthStatus string

const (
	// AuthStatusUnauthorized indicates the user is not authenticated.
	AuthStatusUnauthorized AuthStatus = "unauthorized"
	// AuthStatusPhoneUnknown indicates the phone number is not yet known.
	AuthStatusPhoneUnknown AuthStatus = "phone_unknown"
	// AuthStatusPhoneKnown indicates the phone number is known but not verified.
	AuthStatusPhoneKnown AuthStatus = "phone_known"
	// AuthStatusCodeSent indicates the verification code has been sent.
	AuthStatusCodeSent AuthStatus = "code_sent"
	// AuthStatusCodeVerified indicates the verification code has been verified.
	AuthStatusCodeVerified AuthStatus = "code_verified"
	// AuthStatusTwoFARequired indicates 2FA password is required.
	AuthStatusTwoFARequired AuthStatus = "2fa_required"
	// AuthStatusAuthorized indicates the user is fully authenticated.
	AuthStatusAuthorized AuthStatus = "authorized"
)

// SendCodeResult represents the result of sending a verification code.
type SendCodeResult struct {
	PhoneCodeHash string `json:"phone_code_hash"`
	Timeout       int    `json:"timeout"`
}

// SignInResult represents the result of a sign-in attempt.
type SignInResult struct {
	Success       bool   `json:"success"`
	Requires2FA   bool   `json:"requires_2fa"`
	TwoFactorHint string `json:"two_factor_hint,omitempty"`
	AuthError     string `json:"auth_error,omitempty"`
}

// User represents a Telegram user.
type User struct {
	ID         int64  `json:"id"`
	FirstName  string `json:"first_name,omitempty"`
	LastName   string `json:"last_name,omitempty"`
	Username   string `json:"username,omitempty"`
	Phone      string `json:"phone,omitempty"`
	AccessHash int64  `json:"access_hash,omitempty"`
}
