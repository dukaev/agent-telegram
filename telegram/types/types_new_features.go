// Package types provides types for new Telegram features.
package types // revive:disable:var-naming

import "errors"

// Errors for new features.
var (
	ErrRequiresStickerOrFile = errors.New("stickerId or file is required")
)

// ReadMessagesParams holds parameters for ReadMessages.
type ReadMessagesParams struct {
	Peer  string `json:"peer" validate:"required"`
	MaxID int64  `json:"maxId,omitempty"` // Mark all messages up to this ID as read
}

// Validate validates ReadMessagesParams.
func (p ReadMessagesParams) Validate() error {
	return ValidateStruct(p)
}

// ReadMessagesResult is the result of ReadMessages.
type ReadMessagesResult struct {
	Success bool `json:"success"`
	MaxID   int64 `json:"maxId"`
}

// SetTypingParams holds parameters for SetTyping.
type SetTypingParams struct {
	Peer   string `json:"peer" validate:"required"`
	Action string `json:"action,omitempty"` // typing, upload_photo, record_video, record_audio, etc.
}

// Validate validates SetTypingParams.
func (p SetTypingParams) Validate() error {
	return ValidateStruct(p)
}

// SetTypingResult is the result of SetTyping.
type SetTypingResult struct {
	Success bool `json:"success"`
}

// SendGIFParams holds parameters for SendGIF.
type SendGIFParams struct {
	Peer    string `json:"peer" validate:"required"`
	File    string `json:"file" validate:"required"` // Path to GIF file
	Caption string `json:"caption,omitempty"`
}

// Validate validates SendGIFParams.
func (p SendGIFParams) Validate() error {
	return ValidateStruct(p)
}

// SendGIFResult is the result of SendGIF.
type SendGIFResult struct {
	ID      int64  `json:"id"`
	Date    int64  `json:"date"`
	Peer    string `json:"peer"`
	Caption string `json:"caption,omitempty"`
}

// GetScheduledMessagesParams holds parameters for GetScheduledMessages.
type GetScheduledMessagesParams struct {
	Peer string `json:"peer" validate:"required"`
}

// Validate validates GetScheduledMessagesParams.
func (p GetScheduledMessagesParams) Validate() error {
	return ValidateStruct(p)
}

// ScheduledMessage represents a scheduled message.
type ScheduledMessage struct {
	ID      int64  `json:"id"`
	Date    int64  `json:"date"`    // Scheduled send time
	Message string `json:"message,omitempty"`
	Peer    string `json:"peer"`
}

// GetScheduledMessagesResult is the result of GetScheduledMessages.
type GetScheduledMessagesResult struct {
	Messages []ScheduledMessage `json:"messages"`
	Count    int                `json:"count"`
}

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

// SendVoiceParams holds parameters for SendVoice.
type SendVoiceParams struct {
	Peer     string `json:"peer" validate:"required"`
	File     string `json:"file" validate:"required"` // Path to voice file (OGG/OPUS)
	Duration int    `json:"duration,omitempty"`       // Duration in seconds
	Caption  string `json:"caption,omitempty"`
}

// Validate validates SendVoiceParams.
func (p SendVoiceParams) Validate() error {
	return ValidateStruct(p)
}

// SendVoiceResult is the result of SendVoice.
type SendVoiceResult struct {
	ID       int64  `json:"id"`
	Date     int64  `json:"date"`
	Peer     string `json:"peer"`
	Duration int    `json:"duration,omitempty"`
}

// SendVideoNoteParams holds parameters for SendVideoNote.
type SendVideoNoteParams struct {
	Peer     string `json:"peer" validate:"required"`
	File     string `json:"file" validate:"required"` // Path to video file
	Duration int    `json:"duration,omitempty"`       // Duration in seconds
	Length   int    `json:"length,omitempty"`         // Video width/height (square)
}

// Validate validates SendVideoNoteParams.
func (p SendVideoNoteParams) Validate() error {
	return ValidateStruct(p)
}

// SendVideoNoteResult is the result of SendVideoNote.
type SendVideoNoteResult struct {
	ID       int64  `json:"id"`
	Date     int64  `json:"date"`
	Peer     string `json:"peer"`
	Duration int    `json:"duration,omitempty"`
}

// SendStickerParams holds parameters for SendSticker.
type SendStickerParams struct {
	Peer      string `json:"peer" validate:"required"`
	StickerID string `json:"stickerId,omitempty"` // Sticker file_id or short_name
	File      string `json:"file,omitempty"`      // Path to sticker file (WEBP)
}

// Validate validates SendStickerParams.
func (p SendStickerParams) Validate() error {
	if err := ValidateStruct(p); err != nil {
		return err
	}
	if p.StickerID == "" && p.File == "" {
		return ErrRequiresStickerOrFile
	}
	return nil
}

// SendStickerResult is the result of SendSticker.
type SendStickerResult struct {
	ID   int64  `json:"id"`
	Date int64  `json:"date"`
	Peer string `json:"peer"`
}

// GetStickerPacksParams holds parameters for GetStickerPacks.
type GetStickerPacksParams struct {
	// No required params - returns all sticker packs
}

// Validate validates GetStickerPacksParams.
func (p GetStickerPacksParams) Validate() error {
	return nil
}

// StickerPack represents a sticker pack.
type StickerPack struct {
	ID        int64  `json:"id"`
	Title     string `json:"title"`
	ShortName string `json:"shortName"`
	Count     int    `json:"count"`
	Animated  bool   `json:"animated,omitempty"`
	Video     bool   `json:"video,omitempty"`
}

// GetStickerPacksResult is the result of GetStickerPacks.
type GetStickerPacksResult struct {
	Packs []StickerPack `json:"packs"`
	Count int           `json:"count"`
}

// GetFoldersParams holds parameters for GetFolders.
type GetFoldersParams struct {
	// No required params
}

// Validate validates GetFoldersParams.
func (p GetFoldersParams) Validate() error {
	return nil
}

// ChatFolder represents a chat folder.
type ChatFolder struct {
	ID             int      `json:"id"`
	Title          string   `json:"title"`
	IncludedChats  []string `json:"includedChats,omitempty"`
	ExcludedChats  []string `json:"excludedChats,omitempty"`
	IncludeContacts bool    `json:"includeContacts,omitempty"`
	IncludeNonContacts bool `json:"includeNonContacts,omitempty"`
	IncludeGroups  bool     `json:"includeGroups,omitempty"`
	IncludeChannels bool    `json:"includeChannels,omitempty"`
	IncludeBots    bool     `json:"includeBots,omitempty"`
}

// GetFoldersResult is the result of GetFolders.
type GetFoldersResult struct {
	Folders []ChatFolder `json:"folders"`
	Count   int          `json:"count"`
}

// CreateFolderParams holds parameters for CreateFolder.
type CreateFolderParams struct {
	Title           string   `json:"title" validate:"required"`
	IncludedChats   []string `json:"includedChats,omitempty"`
	ExcludedChats   []string `json:"excludedChats,omitempty"`
	IncludeContacts bool     `json:"includeContacts,omitempty"`
	IncludeNonContacts bool  `json:"includeNonContacts,omitempty"`
	IncludeGroups   bool     `json:"includeGroups,omitempty"`
	IncludeChannels bool     `json:"includeChannels,omitempty"`
	IncludeBots     bool     `json:"includeBots,omitempty"`
}

// Validate validates CreateFolderParams.
func (p CreateFolderParams) Validate() error {
	return ValidateStruct(p)
}

// CreateFolderResult is the result of CreateFolder.
type CreateFolderResult struct {
	Success bool `json:"success"`
	ID      int  `json:"id"`
}

// DeleteFolderParams holds parameters for DeleteFolder.
type DeleteFolderParams struct {
	ID int `json:"id" validate:"required"`
}

// Validate validates DeleteFolderParams.
func (p DeleteFolderParams) Validate() error {
	return ValidateStruct(p)
}

// DeleteFolderResult is the result of DeleteFolder.
type DeleteFolderResult struct {
	Success bool `json:"success"`
}

// SetChatPermissionsParams holds parameters for SetChatPermissions.
type SetChatPermissionsParams struct {
	Peer              string `json:"peer" validate:"required"`
	SendMessages      bool   `json:"sendMessages"`
	SendMedia         bool   `json:"sendMedia"`
	SendStickers      bool   `json:"sendStickers"`
	SendGifs          bool   `json:"sendGifs"`
	SendGames         bool   `json:"sendGames"`
	SendInline        bool   `json:"sendInline"`
	EmbedLinks        bool   `json:"embedLinks"`
	SendPolls         bool   `json:"sendPolls"`
	ChangeInfo        bool   `json:"changeInfo"`
	InviteUsers       bool   `json:"inviteUsers"`
	PinMessages       bool   `json:"pinMessages"`
	ManageTopics      bool   `json:"manageTopics"`
	SendPhotos        bool   `json:"sendPhotos"`
	SendVideos        bool   `json:"sendVideos"`
	SendRoundvideos   bool   `json:"sendRoundvideos"`
	SendAudios        bool   `json:"sendAudios"`
	SendVoices        bool   `json:"sendVoices"`
	SendDocs          bool   `json:"sendDocs"`
	SendPlain         bool   `json:"sendPlain"`
}

// Validate validates SetChatPermissionsParams.
func (p SetChatPermissionsParams) Validate() error {
	return ValidateStruct(p)
}

// SetChatPermissionsResult is the result of SetChatPermissions.
type SetChatPermissionsResult struct {
	Success bool `json:"success"`
}

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
	Type  string   `json:"type"`            // allow_all, allow_contacts, disallow_all, etc.
	Users []int64  `json:"users,omitempty"` // Specific user IDs
	Chats []int64  `json:"chats,omitempty"` // Specific chat IDs
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
