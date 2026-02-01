// Package types provides common types for Telegram client.
package types // revive:disable:var-naming

import (
	"fmt"
	"time"
)

// PeerInfo is a base type for parameters that need peer or username.
type PeerInfo struct {
	Peer     string `json:"peer,omitempty"`
	Username string `json:"username,omitempty"`
}

// ValidatePeer validates that either peer or username is set.
func (p PeerInfo) ValidatePeer() error {
	if p.Peer == "" && p.Username == "" {
		return fmt.Errorf("peer or username is required")
	}
	return nil
}

// MsgID is a base type for parameters that need a message ID.
type MsgID struct {
	MessageID int64 `json:"messageId"`
}

// ValidateMessageID validates that messageId is set.
func (m MsgID) ValidateMessageID() error {
	if m.MessageID == 0 {
		return fmt.Errorf("messageId is required")
	}
	return nil
}

// RequiredText is a base type for parameters with a required text field.
type RequiredText struct {
	Text string `json:"text"`
}

// ValidateText validates that text is set.
func (r RequiredText) ValidateText() error {
	if r.Text == "" {
		return fmt.Errorf("text is required")
	}
	return nil
}

// UpdateType represents the type of Telegram update.
type UpdateType string

const (
	// UpdateTypeNewMessage is a new message update.
	UpdateTypeNewMessage UpdateType = "new_message"
	// UpdateTypeEditMessage is an edited message update.
	UpdateTypeEditMessage UpdateType = "edit_message"
	// UpdateTypeNewChat is a new chat update.
	UpdateTypeNewChat UpdateType = "new_chat"
	// UpdateTypeDelete is a delete update.
	UpdateTypeDelete UpdateType = "delete"
	// UpdateTypeOther is an other type update.
	UpdateTypeOther UpdateType = "other"
)

// StoredUpdate represents a stored Telegram update.
type StoredUpdate struct {
	ID        int64     `json:"id"`
	Type      UpdateType `json:"type"`
	Timestamp time.Time `json:"timestamp"`
	Data      map[string]any `json:"data"`
}

// MessageResult represents a single message result.
type MessageResult struct {
	ID             int64           `json:"id"`
	Date           int64           `json:"date"`
	Text           string          `json:"text,omitempty"`
	FromID         string          `json:"fromId,omitempty"`
	FromName       string          `json:"fromName,omitempty"`
	Out            bool            `json:"out"`
	Buttons        []InlineButton  `json:"buttons,omitempty"`

	// Additional message fields
	PeerID         string          `json:"peerId,omitempty"`        // Chat where message was sent
	EditDate       int64           `json:"editDate,omitempty"`      // When message was edited
	Media          map[string]any  `json:"media,omitempty"`         // Media attachment (photo, document, etc.)
	Views          int             `json:"views,omitempty"`         // View count for channel posts
	Forwards       int             `json:"forwards,omitempty"`      // Forward counter
	ReplyTo        map[string]any  `json:"replyTo,omitempty"`       // Reply information
	FwdFrom        map[string]any  `json:"fwdFrom,omitempty"`       // Forwarded from
	Reactions      []map[string]any `json:"reactions,omitempty"`    // Reactions to message
	Entities       []map[string]any `json:"entities,omitempty"`     // Message entities (formatting)
	Pinned         bool            `json:"pinned,omitempty"`        // Whether message is pinned
	ViaBotID       int64           `json:"viaBotId,omitempty"`      // ID of inline bot
	PostAuthor     string          `json:"postAuthor,omitempty"`    // Author of channel post
	GroupedID      int64           `json:"groupedId,omitempty"`     // Album/media group ID
	TTLPeriod      int             `json:"ttlPeriod,omitempty"`     // Time to live
	Mentioned      bool            `json:"mentioned,omitempty"`     // Whether we were mentioned
	Silent         bool            `json:"silent,omitempty"`        // Silent message (no notification)
	Post           bool            `json:"post,omitempty"`          // Channel post
}

// GetMessagesParams holds parameters for GetMessages.
type GetMessagesParams struct {
	Username string `json:"username"`
	Limit    int    `json:"limit"`
	Offset   int    `json:"offset"`
}

// GetMessagesResult is the result of GetMessages.
type GetMessagesResult struct {
	Messages []MessageResult `json:"messages"`
	Limit    int             `json:"limit"`
	Offset   int             `json:"offset"`
	Count    int             `json:"count"`
	Username string          `json:"username"`
}

// GetChatsParams holds parameters for GetChats.
type GetChatsParams struct {
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

// GetChatsResult is the result of GetChats.
type GetChatsResult struct {
	Chats  []map[string]any `json:"chats"`
	Limit  int              `json:"limit"`
	Offset int              `json:"offset"`
	Count  int              `json:"count"`
}

// GetUpdatesParams holds parameters for GetUpdates.
type GetUpdatesParams struct {
	PeerInfo
	Limit int `json:"limit"`
}

// GetUpdatesResult is the result of GetUpdates.
type GetUpdatesResult struct {
	Updates []StoredUpdate `json:"updates"`
	Count   int            `json:"count"`
}

// ClearMessagesParams holds parameters for ClearMessages.
type ClearMessagesParams struct {
	PeerInfo
	MessageIDs []int64 `json:"messageIds"`
}

// Validate validates ClearMessagesParams.
func (p ClearMessagesParams) Validate() error {
	if err := p.ValidatePeer(); err != nil {
		return err
	}
	if len(p.MessageIDs) == 0 {
		return fmt.Errorf("messageIds is required")
	}
	return nil
}

// ClearMessagesResult is the result of ClearMessages.
type ClearMessagesResult struct {
	Success bool   `json:"success"`
	Cleared int    `json:"cleared"`
	Peer    string `json:"peer"`
}

// ClearHistoryParams holds parameters for ClearHistory.
type ClearHistoryParams struct {
	PeerInfo
	Revoke bool `json:"revoke,omitempty"`
}

// Validate validates ClearHistoryParams.
func (p ClearHistoryParams) Validate() error {
	return p.ValidatePeer()
}

// ClearHistoryResult is the result of ClearHistory.
type ClearHistoryResult struct {
	Success bool   `json:"success"`
	Peer    string `json:"peer"`
	Revoke  bool   `json:"revoke"`
}

// ForwardMessageParams holds parameters for ForwardMessage.
type ForwardMessageParams struct {
	FromPeer string `json:"fromPeer"`
	MessageID int64 `json:"messageId"`
	ToPeer   string `json:"toPeer"`
}

// Validate validates ForwardMessageParams.
func (p ForwardMessageParams) Validate() error {
	if p.FromPeer == "" {
		return fmt.Errorf("fromPeer is required")
	}
	if p.MessageID == 0 {
		return fmt.Errorf("messageId is required")
	}
	if p.ToPeer == "" {
		return fmt.Errorf("toPeer is required")
	}
	return nil
}

// ForwardMessageResult is the result of ForwardMessage.
type ForwardMessageResult struct {
	Success   bool  `json:"success"`
	MessageID int64 `json:"id"`
}

// PinChatParams holds parameters for PinChat (pin chat in dialog list).
type PinChatParams struct {
	PeerInfo
	Disable bool `json:"disable"` // true to unpin, false to pin
}

// Validate validates PinChatParams.
func (p PinChatParams) Validate() error {
	return p.ValidatePeer()
}

// PinChatResult is the result of PinChat.
type PinChatResult struct {
	Success bool   `json:"success"`
	Peer    string `json:"peer"`
	Pinned  bool   `json:"pinned"`
}

// JoinChatParams holds parameters for JoinChat.
type JoinChatParams struct {
	InviteLink string `json:"inviteLink"`
}

// Validate validates JoinChatParams.
func (p JoinChatParams) Validate() error {
	if p.InviteLink == "" {
		return fmt.Errorf("inviteLink is required")
	}
	return nil
}

// JoinChatResult is the result of JoinChat.
type JoinChatResult struct {
	Success bool   `json:"success"`
	ChatID  int64  `json:"chatId"`
	Title   string `json:"title"`
}

// SubscribeChannelParams holds parameters for SubscribeChannel.
type SubscribeChannelParams struct {
	Channel string `json:"channel"` // @username or username
}

// Validate validates SubscribeChannelParams.
func (p SubscribeChannelParams) Validate() error {
	if p.Channel == "" {
		return fmt.Errorf("channel is required")
	}
	return nil
}

// SubscribeChannelResult is the result of SubscribeChannel.
type SubscribeChannelResult struct {
	Success bool   `json:"success"`
	ChatID  int64  `json:"chatId"`
	Title   string `json:"title"`
}

// SearchGlobalParams holds parameters for SearchGlobal.
type SearchGlobalParams struct {
	Query string `json:"query"`
	Type  string `json:"type,omitempty"` // bots, users, chats, channels, or empty for all
	Limit int    `json:"limit,omitempty"`
}

// SearchResult represents a single search result.
type SearchResult struct {
	ID       int64         `json:"id"`
	Date     int64         `json:"date"`
	Text     string        `json:"text,omitempty"`
	Peer     string        `json:"peer,omitempty"`
	FromID   string        `json:"fromId,omitempty"`
	FromName string        `json:"fromName,omitempty"`
	Media    map[string]any `json:"media,omitempty"`
}

// SearchGlobalResult is the result of SearchGlobal.
type SearchGlobalResult struct {
	Query   string        `json:"query"`
	Type    string        `json:"type,omitempty"`
	Results []SearchResult `json:"results"`
	Count   int           `json:"count"`
}

// SearchInChatParams holds parameters for SearchInChat.
type SearchInChatParams struct {
	Peer   string `json:"peer"`
	Query  string `json:"query"`
	Type   string `json:"type,omitempty"` // text, photos, videos, documents, links, audio, voice
	Limit  int    `json:"limit,omitempty"`
	Offset int    `json:"offset,omitempty"`
}

// SearchInChatResult is the result of SearchInChat.
type SearchInChatResult struct {
	Peer     string          `json:"peer"`
	Query    string          `json:"query"`
	Type     string          `json:"type,omitempty"`
	Messages []MessageResult `json:"messages"`
	Count    int             `json:"count"`
	Total    int             `json:"total"`
	Limit    int             `json:"limit"`
	Offset   int             `json:"offset"`
}

// Contact represents a contact in the user's contact list.
type Contact struct {
	ID          int64  `json:"id"`
	FirstName   string `json:"firstName,omitempty"`
	LastName    string `json:"lastName,omitempty"`
	Username    string `json:"username,omitempty"`
	Phone       string `json:"phone,omitempty"`
	Mutual      bool   `json:"mutual,omitempty"`
	Verified    bool   `json:"verified,omitempty"`
	Bot         bool   `json:"bot,omitempty"`
	Peer        string `json:"peer,omitempty"`
}

// GetContactsParams holds parameters for GetContacts.
type GetContactsParams struct {
	Query string `json:"query,omitempty"` // Search query to filter contacts
	Limit int    `json:"limit,omitempty"` // Maximum number of contacts to return
}

// Validate validates GetContactsParams (all fields are optional).
func (p GetContactsParams) Validate() error {
	return nil
}

// GetContactsResult is the result of GetContacts.
type GetContactsResult struct {
	Contacts []Contact `json:"contacts"`
	Count    int       `json:"count"`
	Query    string    `json:"query,omitempty"`
}

// AddContactParams holds parameters for AddContact.
type AddContactParams struct {
	Phone    string `json:"phone"`              // Phone number (with country code, e.g. +1234567890)
	FirstName string `json:"firstName"`          // First name
	LastName  string `json:"lastName,omitempty"` // Last name (optional)
}

// Validate validates AddContactParams.
func (p AddContactParams) Validate() error {
	if p.Phone == "" {
		return fmt.Errorf("phone is required")
	}
	if p.FirstName == "" {
		return fmt.Errorf("firstName is required")
	}
	return nil
}

// AddContactResult is the result of AddContact.
type AddContactResult struct {
	Success bool   `json:"success"`
	Contact Contact `json:"contact,omitempty"`
}

// DeleteContactParams holds parameters for DeleteContact.
type DeleteContactParams struct {
	Username string `json:"username,omitempty"` // Username to delete (e.g. "username" or "@username")
	UserID   int64  `json:"userId,omitempty"`   // User ID to delete
}

// Validate validates DeleteContactParams.
func (p DeleteContactParams) Validate() error {
	if p.Username == "" && p.UserID == 0 {
		return fmt.Errorf("username or userId is required")
	}
	return nil
}

// DeleteContactResult is the result of DeleteContact.
type DeleteContactResult struct {
	Success bool `json:"success"`
}

// GetTopicsParams holds parameters for GetTopics.
type GetTopicsParams struct {
	Peer  string `json:"peer"`           // Channel username or ID
	Limit int    `json:"limit,omitempty"` // Maximum number of topics to return
}

// Validate validates GetTopicsParams.
func (p GetTopicsParams) Validate() error {
	if p.Peer == "" {
		return fmt.Errorf("peer is required")
	}
	return nil
}

// ForumTopic represents a forum topic.
type ForumTopic struct {
	ID       int64  `json:"id"`                  // Topic ID
	Title    string `json:"title"`               // Topic title
	IconColor int32  `json:"iconColor,omitempty"` // Icon color
	IconEmoji string `json:"iconEmoji,omitempty"` // Icon emoji
	Top      bool   `json:"top,omitempty"`       // Whether topic is pinned
	Closed   bool   `json:"closed,omitempty"`    // Whether topic is closed
}

// GetTopicsResult is the result of GetTopics.
type GetTopicsResult struct {
	Peer   string        `json:"peer"`
	Topics []ForumTopic  `json:"topics"`
	Count  int           `json:"count"`
}

// CreateGroupParams holds parameters for CreateGroup.
type CreateGroupParams struct {
	Title   string   `json:"title"`             // Group title
	Members []string `json:"members"`           // List of usernames to add
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
	Title       string `json:"title"`                 // Channel title
	Description string `json:"description,omitempty"`  // Channel description
	Username    string `json:"username,omitempty"`     // Channel username (optional)
	Megagroup   bool   `json:"megagroup,omitempty"`    // Create as supergroup instead of channel
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
	Peer  string `json:"peer"`           // Chat/channel username or ID
	Title string `json:"title"`          // New title
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
	Peer  string `json:"peer"`           // Chat/channel username or ID
	File  string `json:"file"`           // Path to photo file
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
	Peer  string `json:"peer"`           // Chat/channel username or ID
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

// LeaveParams holds parameters for Leave.
type LeaveParams struct {
	Peer  string `json:"peer"`           // Chat/channel username or ID
}

// Validate validates LeaveParams.
func (p LeaveParams) Validate() error {
	if p.Peer == "" {
		return fmt.Errorf("peer is required")
	}
	return nil
}

// LeaveResult is the result of Leave.
type LeaveResult struct {
	Success bool `json:"success"`
}

// InviteParams holds parameters for Invite.
type InviteParams struct {
	Peer   string   `json:"peer"`           // Chat/channel username or ID
	Members []string `json:"members"`        // List of usernames to invite
}

// Validate validates InviteParams.
func (p InviteParams) Validate() error {
	if p.Peer == "" {
		return fmt.Errorf("peer is required")
	}
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
	ID          int64  `json:"id"`
	FirstName   string `json:"firstName,omitempty"`
	LastName    string `json:"lastName,omitempty"`
	Username    string `json:"username,omitempty"`
	Bot         bool   `json:"bot,omitempty"`
	Admin       bool   `json:"admin,omitempty"`
	Creator     bool   `json:"creator,omitempty"`
	Peer        string `json:"peer,omitempty"`
}

// GetParticipantsParams holds parameters for GetParticipants.
type GetParticipantsParams struct {
	Peer  string `json:"peer"`           // Chat/channel username or ID
	Limit int    `json:"limit,omitempty"` // Maximum number of participants (default 100)
}

// Validate validates GetParticipantsParams.
func (p GetParticipantsParams) Validate() error {
	if p.Peer == "" {
		return fmt.Errorf("peer is required")
	}
	return nil
}

// GetParticipantsResult is the result of GetParticipants.
type GetParticipantsResult struct {
	Peer         string        `json:"peer"`
	Participants []Participant `json:"participants"`
	Count        int           `json:"count"`
}

// GetAdminsParams holds parameters for GetAdmins.
type GetAdminsParams struct {
	Peer  string `json:"peer"`           // Chat/channel username or ID
	Limit int    `json:"limit,omitempty"` // Maximum number of admins (default 100)
}

// Validate validates GetAdminsParams.
func (p GetAdminsParams) Validate() error {
	if p.Peer == "" {
		return fmt.Errorf("peer is required")
	}
	return nil
}

// GetAdminsResult is the result of GetAdmins.
type GetAdminsResult struct {
	Peer  string        `json:"peer"`
	Admins []Participant `json:"admins"`
	Count int           `json:"count"`
}

// GetBannedParams holds parameters for GetBanned.
type GetBannedParams struct {
	Peer  string `json:"peer"`           // Chat/channel username or ID
	Limit int    `json:"limit,omitempty"` // Maximum number of banned users (default 100)
}

// Validate validates GetBannedParams.
func (p GetBannedParams) Validate() error {
	if p.Peer == "" {
		return fmt.Errorf("peer is required")
	}
	return nil
}

// GetBannedResult is the result of GetBanned.
type GetBannedResult struct {
	Peer   string        `json:"peer"`
	Banned []Participant `json:"banned"`
	Count  int           `json:"count"`
}

// PromoteAdminParams holds parameters for PromoteAdmin.
type PromoteAdminParams struct {
	Peer        string `json:"peer"`                   // Chat/channel username or ID
	User        string `json:"user"`                   // Username to promote
	CanChangeInfo bool  `json:"canChangeInfo,omitempty"`  // Can change chat info
	CanPostMessages bool `json:"canPostMessages,omitempty"` // Can post messages
	CanEditMessages bool  `json:"canEditMessages,omitempty"`  // Can edit messages
	CanDeleteMessages bool `json:"canDeleteMessages,omitempty"` // Can delete messages
	CanBanUsers bool  `json:"canBanUsers,omitempty"`   // Can ban users
	CanInviteUsers bool  `json:"canInviteUsers,omitempty"` // Can invite users
	CanPinMessages bool  `json:"canPinMessages,omitempty"` // Can pin messages
	CanAddAdmins bool  `json:"canAddAdmins,omitempty"` // Can add new admins
	Anonymous bool  `json:"anonymous,omitempty"`      // Is anonymous admin
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
	Peer        string `json:"peer"`                   // Chat/channel username or ID
	User        string `json:"user"`                   // Username to demote
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
	Peer     string `json:"peer"`               // Chat/channel username or ID
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
	Link       string `json:"link"`
	Usage      int    `json:"usage,omitempty"`
	UsageLimit int    `json:"usageLimit,omitempty"`
	RequestNeeded bool `json:"requestNeeded,omitempty"`
	Expired    bool   `json:"expired,omitempty"`
}
