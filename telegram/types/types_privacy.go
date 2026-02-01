// Package types provides common types for Telegram privacy settings.
package types

// GetPrivacyParams holds parameters for GetPrivacy.
type GetPrivacyParams struct {
	Key string `json:"key" validate:"required"` // status_timestamp, phone_number, profile_photo, etc.
}

// Validate validates GetPrivacyParams.
func (p GetPrivacyParams) Validate() error {
	return ValidateStruct(p)
}

// PrivacyRule represents a privacy rule.
type PrivacyRule struct {
	Type  string  `json:"type"`            // allow_all, allow_contacts, disallow_all, etc.
	Users []int64 `json:"users,omitempty"` // Specific user IDs
	Chats []int64 `json:"chats,omitempty"` // Specific chat IDs
}

// GetPrivacyResult is the result of GetPrivacy.
type GetPrivacyResult struct {
	Key   string        `json:"key"`
	Rules []PrivacyRule `json:"rules"`
}

// SetPrivacyParams holds parameters for SetPrivacy.
type SetPrivacyParams struct {
	Key   string        `json:"key" validate:"required"`
	Rules []PrivacyRule `json:"rules" validate:"required"`
}

// Validate validates SetPrivacyParams.
func (p SetPrivacyParams) Validate() error {
	return ValidateStruct(p)
}

// SetPrivacyResult is the result of SetPrivacy.
type SetPrivacyResult struct {
	Success bool `json:"success"`
}
