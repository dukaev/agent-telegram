// Package telegram defines interfaces for domain clients.
package telegram

import (
	"context"

	"agent-telegram/telegram/types"
)

// ChatClient defines the interface for chat operations.
type ChatClient interface {
	GetChats(ctx context.Context, params *types.GetChatsParams) (*types.GetChatsResult, error)
	GetTopics(ctx context.Context, params types.GetTopicsParams) (*types.GetTopicsResult, error)
	CreateGroup(ctx context.Context, params types.CreateGroupParams) (*types.CreateGroupResult, error)
	CreateChannel(ctx context.Context, params types.CreateChannelParams) (*types.CreateChannelResult, error)
	EditTitle(ctx context.Context, params types.EditTitleParams) (*types.EditTitleResult, error)
	SetPhoto(ctx context.Context, params types.SetPhotoParams) (*types.SetPhotoResult, error)
	DeletePhoto(ctx context.Context, params types.DeletePhotoParams) (*types.DeletePhotoResult, error)
	Leave(ctx context.Context, params types.LeaveParams) (*types.LeaveResult, error)
	Invite(ctx context.Context, params types.InviteParams) (*types.InviteResult, error)
	GetParticipants(ctx context.Context, params types.GetParticipantsParams) (*types.GetParticipantsResult, error)
	GetAdmins(ctx context.Context, params types.GetAdminsParams) (*types.GetAdminsResult, error)
	GetBanned(ctx context.Context, params types.GetBannedParams) (*types.GetBannedResult, error)
	PromoteAdmin(ctx context.Context, params types.PromoteAdminParams) (*types.PromoteAdminResult, error)
	DemoteAdmin(ctx context.Context, params types.DemoteAdminParams) (*types.DemoteAdminResult, error)
	GetInviteLink(ctx context.Context, params types.GetInviteLinkParams) (*types.GetInviteLinkResult, error)
	ClearMessages(ctx context.Context, params types.ClearMessagesParams) (*types.ClearMessagesResult, error)
	ClearHistory(ctx context.Context, params types.ClearHistoryParams) (*types.ClearHistoryResult, error)
	PinChat(ctx context.Context, params types.PinChatParams) (*types.PinChatResult, error)
	JoinChat(ctx context.Context, params types.JoinChatParams) (*types.JoinChatResult, error)
	SubscribeChannel(ctx context.Context, params types.SubscribeChannelParams) (*types.SubscribeChannelResult, error)
	// New features
	SetSlowMode(ctx context.Context, params types.SetSlowModeParams) (*types.SetSlowModeResult, error)
	SetChatPermissions(ctx context.Context, params types.SetChatPermissionsParams) (*types.SetChatPermissionsResult, error)
	GetFolders(ctx context.Context, params types.GetFoldersParams) (*types.GetFoldersResult, error)
	CreateFolder(ctx context.Context, params types.CreateFolderParams) (*types.CreateFolderResult, error)
	DeleteFolder(ctx context.Context, params types.DeleteFolderParams) (*types.DeleteFolderResult, error)
	// Archive and mute
	Archive(ctx context.Context, params types.ArchiveParams) (*types.ArchiveResult, error)
	Unarchive(ctx context.Context, params types.UnarchiveParams) (*types.UnarchiveResult, error)
	Mute(ctx context.Context, params types.MuteParams) (*types.MuteResult, error)
	Unmute(ctx context.Context, params types.UnmuteParams) (*types.UnmuteResult, error)
}

// MessageClient defines the interface for message operations.
type MessageClient interface {
	GetMessages(ctx context.Context, params types.GetMessagesParams) (*types.GetMessagesResult, error)
	SendMessage(ctx context.Context, params types.SendMessageParams) (*types.SendMessageResult, error)
	SendReply(ctx context.Context, params types.SendReplyParams) (*types.SendReplyResult, error)
	UpdateMessage(ctx context.Context, params types.UpdateMessageParams) (*types.UpdateMessageResult, error)
	DeleteMessage(ctx context.Context, params types.DeleteMessageParams) (*types.DeleteMessageResult, error)
	ForwardMessage(ctx context.Context, params types.ForwardMessageParams) (*types.ForwardMessageResult, error)
	InspectInlineButtons(
		ctx context.Context, params types.InspectInlineButtonsParams,
	) (*types.InspectInlineButtonsResult, error)
	PressInlineButton(ctx context.Context, params types.PressInlineButtonParams) (*types.PressInlineButtonResult, error)
	InspectReplyKeyboard(ctx context.Context, params types.PeerInfo) (*types.ReplyKeyboardResult, error)
	// New features
	ReadMessages(ctx context.Context, params types.ReadMessagesParams) (*types.ReadMessagesResult, error)
	SetTyping(ctx context.Context, params types.SetTypingParams) (*types.SetTypingResult, error)
	GetScheduledMessages(
		ctx context.Context, params types.GetScheduledMessagesParams,
	) (*types.GetScheduledMessagesResult, error)
}

// MediaClient defines the interface for media operations.
type MediaClient interface {
	SendPhoto(ctx context.Context, params types.SendPhotoParams) (*types.SendPhotoResult, error)
	SendVideo(ctx context.Context, params types.SendVideoParams) (*types.SendVideoResult, error)
	SendFile(ctx context.Context, params types.SendFileParams) (*types.SendFileResult, error)
	SendContact(ctx context.Context, params types.SendContactParams) (*types.SendContactResult, error)
	SendLocation(ctx context.Context, params types.SendLocationParams) (*types.SendLocationResult, error)
	SendPoll(ctx context.Context, params types.SendPollParams) (*types.SendPollResult, error)
	// New features
	SendVoice(ctx context.Context, params types.SendVoiceParams) (*types.SendVoiceResult, error)
	SendVideoNote(ctx context.Context, params types.SendVideoNoteParams) (*types.SendVideoNoteResult, error)
	SendSticker(ctx context.Context, params types.SendStickerParams) (*types.SendStickerResult, error)
	GetStickerPacks(ctx context.Context, params types.GetStickerPacksParams) (*types.GetStickerPacksResult, error)
	SendGIF(ctx context.Context, params types.SendGIFParams) (*types.SendGIFResult, error)
}

// UserClient defines the interface for user operations.
type UserClient interface {
	GetUserInfo(ctx context.Context, params types.GetUserInfoParams) (*types.GetUserInfoResult, error)
	GetContacts(ctx context.Context, params types.GetContactsParams) (*types.GetContactsResult, error)
	AddContact(ctx context.Context, params types.AddContactParams) (*types.AddContactResult, error)
	DeleteContact(ctx context.Context, params types.DeleteContactParams) (*types.DeleteContactResult, error)
	UpdateProfile(ctx context.Context, params types.UpdateProfileParams) (*types.UpdateProfileResult, error)
	UpdateAvatar(ctx context.Context, params types.UpdateAvatarParams) (*types.UpdateAvatarResult, error)
	BlockPeer(ctx context.Context, params types.BlockPeerParams) (*types.BlockPeerResult, error)
	UnblockPeer(ctx context.Context, params types.UnblockPeerParams) (*types.UnblockPeerResult, error)
	// New features
	GetPrivacy(ctx context.Context, params types.GetPrivacyParams) (*types.GetPrivacyResult, error)
	SetPrivacy(ctx context.Context, params types.SetPrivacyParams) (*types.SetPrivacyResult, error)
}

// PinClient defines the interface for pin operations.
type PinClient interface {
	PinMessage(ctx context.Context, params types.PinMessageParams) (*types.PinMessageResult, error)
	UnpinMessage(ctx context.Context, params types.UnpinMessageParams) (*types.UnpinMessageResult, error)
}

// ReactionClient defines the interface for reaction operations.
type ReactionClient interface {
	AddReaction(ctx context.Context, params types.AddReactionParams) (*types.AddReactionResult, error)
	RemoveReaction(ctx context.Context, params types.RemoveReactionParams) (*types.RemoveReactionResult, error)
	ListReactions(ctx context.Context, params types.ListReactionsParams) (*types.ListReactionsResult, error)
}

// SearchClient defines the interface for search operations.
type SearchClient interface {
	SearchGlobal(ctx context.Context, params types.SearchGlobalParams) (*types.SearchGlobalResult, error)
	SearchInChat(ctx context.Context, params types.SearchInChatParams) (*types.SearchInChatResult, error)
}

// GiftClient defines the interface for gift operations.
type GiftClient interface {
	GetStarGifts(ctx context.Context, params types.GetStarGiftsParams) (*types.GetStarGiftsResult, error)
	SendStarGift(ctx context.Context, params types.SendStarGiftParams) (*types.SendStarGiftResult, error)
	GetSavedGifts(ctx context.Context, params types.GetSavedGiftsParams) (*types.GetSavedGiftsResult, error)
	TransferStarGift(ctx context.Context, params types.TransferStarGiftParams) (*types.TransferStarGiftResult, error)
	ConvertStarGift(ctx context.Context, params types.ConvertStarGiftParams) (*types.ConvertStarGiftResult, error)
}
