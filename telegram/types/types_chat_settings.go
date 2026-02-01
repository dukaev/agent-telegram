// Package types provides common types for Telegram chat settings.
package types

// SetSlowModeParams holds parameters for SetSlowMode.
type SetSlowModeParams struct {
	Peer    string `json:"peer" validate:"required"`
	Seconds int    `json:"seconds"` // 0 to disable, or 10, 30, 60, 300, 900, 3600
}

// Validate validates SetSlowModeParams.
func (p SetSlowModeParams) Validate() error {
	return ValidateStruct(p)
}

// SetSlowModeResult is the result of SetSlowMode.
type SetSlowModeResult struct {
	Success bool `json:"success"`
	Seconds int  `json:"seconds"`
}

// SetChatPermissionsParams holds parameters for SetChatPermissions.
type SetChatPermissionsParams struct {
	Peer            string `json:"peer" validate:"required"`
	SendMessages    bool   `json:"sendMessages"`
	SendMedia       bool   `json:"sendMedia"`
	SendStickers    bool   `json:"sendStickers"`
	SendGifs        bool   `json:"sendGifs"`
	SendGames       bool   `json:"sendGames"`
	SendInline      bool   `json:"sendInline"`
	EmbedLinks      bool   `json:"embedLinks"`
	SendPolls       bool   `json:"sendPolls"`
	ChangeInfo      bool   `json:"changeInfo"`
	InviteUsers     bool   `json:"inviteUsers"`
	PinMessages     bool   `json:"pinMessages"`
	ManageTopics    bool   `json:"manageTopics"`
	SendPhotos      bool   `json:"sendPhotos"`
	SendVideos      bool   `json:"sendVideos"`
	SendRoundvideos bool   `json:"sendRoundvideos"`
	SendAudios      bool   `json:"sendAudios"`
	SendVoices      bool   `json:"sendVoices"`
	SendDocs        bool   `json:"sendDocs"`
	SendPlain       bool   `json:"sendPlain"`
}

// Validate validates SetChatPermissionsParams.
func (p SetChatPermissionsParams) Validate() error {
	return ValidateStruct(p)
}

// SetChatPermissionsResult is the result of SetChatPermissions.
type SetChatPermissionsResult struct {
	Success bool `json:"success"`
}
