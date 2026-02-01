package chat

import (
	"context"
	"fmt"
	"math"
	"os"

	"github.com/gotd/td/telegram/uploader"
	"github.com/gotd/td/tg"
	"agent-telegram/telegram/types"
)

// Archive moves a chat to the archive folder.
func (c *Client) Archive(ctx context.Context, params types.ArchiveParams) (*types.ArchiveResult, error) {
	if err := c.CheckInitialized(); err != nil {
		return nil, err
	}

	peer, err := c.ResolvePeer(ctx, params.Peer)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve peer: %w", err)
	}

	// Archive folder has ID 1
	_, err = c.API.FoldersEditPeerFolders(ctx, []tg.InputFolderPeer{
		{Peer: peer, FolderID: 1},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to archive chat: %w", err)
	}

	return &types.ArchiveResult{Success: true, Peer: params.Peer}, nil
}

// Unarchive moves a chat from the archive folder back to main list.
func (c *Client) Unarchive(ctx context.Context, params types.UnarchiveParams) (*types.UnarchiveResult, error) {
	if err := c.CheckInitialized(); err != nil {
		return nil, err
	}

	peer, err := c.ResolvePeer(ctx, params.Peer)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve peer: %w", err)
	}

	// Main folder has ID 0
	_, err = c.API.FoldersEditPeerFolders(ctx, []tg.InputFolderPeer{
		{Peer: peer, FolderID: 0},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to unarchive chat: %w", err)
	}

	return &types.UnarchiveResult{Success: true, Peer: params.Peer}, nil
}

// Mute mutes notifications for a chat.
func (c *Client) Mute(ctx context.Context, params types.MuteParams) (*types.MuteResult, error) {
	if err := c.CheckInitialized(); err != nil {
		return nil, err
	}

	peer, err := c.ResolvePeer(ctx, params.Peer)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve peer: %w", err)
	}

	// Mute until max int32 (forever)
	_, err = c.API.AccountUpdateNotifySettings(ctx, &tg.AccountUpdateNotifySettingsRequest{
		Peer: &tg.InputNotifyPeer{Peer: peer},
		Settings: tg.InputPeerNotifySettings{
			MuteUntil: math.MaxInt32,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to mute chat: %w", err)
	}

	return &types.MuteResult{Success: true, Peer: params.Peer}, nil
}

// Unmute unmutes notifications for a chat.
func (c *Client) Unmute(ctx context.Context, params types.UnmuteParams) (*types.UnmuteResult, error) {
	if err := c.CheckInitialized(); err != nil {
		return nil, err
	}

	peer, err := c.ResolvePeer(ctx, params.Peer)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve peer: %w", err)
	}

	// Unmute by setting mute_until to 0
	_, err = c.API.AccountUpdateNotifySettings(ctx, &tg.AccountUpdateNotifySettingsRequest{
		Peer: &tg.InputNotifyPeer{Peer: peer},
		Settings: tg.InputPeerNotifySettings{
			MuteUntil: 0,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to unmute chat: %w", err)
	}

	return &types.UnmuteResult{Success: true, Peer: params.Peer}, nil
}

// EditTitle edits the title of a chat or channel.
func (c *Client) EditTitle(ctx context.Context, params types.EditTitleParams) (*types.EditTitleResult, error) {
	if err := c.CheckInitialized(); err != nil {
		return nil, err
	}

	peer, err := c.ResolvePeer(ctx, params.Peer)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve peer: %w", err)
	}

	switch p := peer.(type) {
	case *tg.InputPeerChannel:
		_, err := c.API.ChannelsEditTitle(ctx, &tg.ChannelsEditTitleRequest{
			Channel: &tg.InputChannel{
				ChannelID:  p.ChannelID,
				AccessHash: p.AccessHash,
			},
			Title: params.Title,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to edit channel title: %w", err)
		}
	case *tg.InputPeerChat:
		_, err := c.API.MessagesEditChatTitle(ctx, &tg.MessagesEditChatTitleRequest{
			ChatID: p.ChatID,
			Title:  params.Title,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to edit chat title: %w", err)
		}
	default:
		return nil, fmt.Errorf("peer must be a chat or channel")
	}

	return &types.EditTitleResult{
		Success: true,
		Title:   params.Title,
	}, nil
}

// SetPhoto sets the photo for a chat or channel.
func (c *Client) SetPhoto(ctx context.Context, params types.SetPhotoParams) (*types.SetPhotoResult, error) {
	if err := c.CheckInitialized(); err != nil {
		return nil, err
	}

	peer, err := c.ResolvePeer(ctx, params.Peer)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve peer: %w", err)
	}

	// Upload the photo file
	// #nosec G304 -- filePath is validated in handler
	file, err := os.Open(params.File)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer func() { _ = file.Close() }()

	fileInfo, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}

	u := uploader.NewUploader(c.API)
	upload := uploader.NewUpload(fileInfo.Name(), file, fileInfo.Size())
	uploadedFile, err := u.Upload(ctx, upload)
	if err != nil {
		return nil, fmt.Errorf("failed to upload file: %w", err)
	}

	photo := &tg.InputChatUploadedPhoto{
		File: uploadedFile,
	}

	switch p := peer.(type) {
	case *tg.InputPeerChannel:
		_, err := c.API.ChannelsEditPhoto(ctx, &tg.ChannelsEditPhotoRequest{
			Channel: &tg.InputChannel{
				ChannelID:  p.ChannelID,
				AccessHash: p.AccessHash,
			},
			Photo: photo,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to set channel photo: %w", err)
		}
	case *tg.InputPeerChat:
		_, err := c.API.MessagesEditChatPhoto(ctx, &tg.MessagesEditChatPhotoRequest{
			ChatID: p.ChatID,
			Photo:  photo,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to set chat photo: %w", err)
		}
	default:
		return nil, fmt.Errorf("peer must be a chat or channel")
	}

	return &types.SetPhotoResult{Success: true}, nil
}

// DeletePhoto deletes the photo from a chat or channel.
func (c *Client) DeletePhoto(ctx context.Context, params types.DeletePhotoParams) (*types.DeletePhotoResult, error) {
	if err := c.CheckInitialized(); err != nil {
		return nil, err
	}

	peer, err := c.ResolvePeer(ctx, params.Peer)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve peer: %w", err)
	}

	switch p := peer.(type) {
	case *tg.InputPeerChannel:
		_, err := c.API.ChannelsEditPhoto(ctx, &tg.ChannelsEditPhotoRequest{
			Channel: &tg.InputChannel{
				ChannelID:  p.ChannelID,
				AccessHash: p.AccessHash,
			},
			Photo: &tg.InputChatPhotoEmpty{},
		})
		if err != nil {
			return nil, fmt.Errorf("failed to delete channel photo: %w", err)
		}
	case *tg.InputPeerChat:
		_, err := c.API.MessagesEditChatPhoto(ctx, &tg.MessagesEditChatPhotoRequest{
			ChatID: p.ChatID,
			Photo:  &tg.InputChatPhotoEmpty{},
		})
		if err != nil {
			return nil, fmt.Errorf("failed to delete chat photo: %w", err)
		}
	default:
		return nil, fmt.Errorf("peer must be a chat or channel")
	}

	return &types.DeletePhotoResult{Success: true}, nil
}

// SetSlowMode sets slow mode for a channel/supergroup.
func (c *Client) SetSlowMode(ctx context.Context, params types.SetSlowModeParams) (*types.SetSlowModeResult, error) {
	if err := c.CheckInitialized(); err != nil {
		return nil, err
	}

	peer, err := c.ResolvePeer(ctx, params.Peer)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve peer: %w", err)
	}

	inputChannel, ok := peer.(*tg.InputPeerChannel)
	if !ok {
		return nil, fmt.Errorf("peer must be a channel or supergroup")
	}

	_, err = c.API.ChannelsToggleSlowMode(ctx, &tg.ChannelsToggleSlowModeRequest{
		Channel: &tg.InputChannel{
			ChannelID:  inputChannel.ChannelID,
			AccessHash: inputChannel.AccessHash,
		},
		Seconds: params.Seconds,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to set slow mode: %w", err)
	}

	return &types.SetSlowModeResult{
		Success: true,
		Seconds: params.Seconds,
	}, nil
}

// SetChatPermissions sets default permissions for a chat/channel.
//
func (c *Client) SetChatPermissions(
	ctx context.Context,
	params types.SetChatPermissionsParams,
) (*types.SetChatPermissionsResult, error) {
	if err := c.CheckInitialized(); err != nil {
		return nil, err
	}

	peer, err := c.ResolvePeer(ctx, params.Peer)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve peer: %w", err)
	}

	// Build banned rights (inverted - true means banned)
	rights := &tg.ChatBannedRights{
		UntilDate:    0, // Permanent
		ViewMessages: false,
		SendMessages: !params.SendMessages,
		SendMedia:    !params.SendMedia,
		SendStickers: !params.SendStickers,
		SendGifs:     !params.SendGifs,
		SendGames:    !params.SendGames,
		SendInline:   !params.SendInline,
		EmbedLinks:   !params.EmbedLinks,
		SendPolls:    !params.SendPolls,
		ChangeInfo:   !params.ChangeInfo,
		InviteUsers:  !params.InviteUsers,
		PinMessages:  !params.PinMessages,
		ManageTopics: !params.ManageTopics,
		SendPhotos:   !params.SendPhotos,
		SendVideos:   !params.SendVideos,
		SendRoundvideos: !params.SendRoundvideos,
		SendAudios:   !params.SendAudios,
		SendVoices:   !params.SendVoices,
		SendDocs:     !params.SendDocs,
		SendPlain:    !params.SendPlain,
	}

	switch p := peer.(type) {
	case *tg.InputPeerChannel:
		_, err = c.API.MessagesEditChatDefaultBannedRights(ctx, &tg.MessagesEditChatDefaultBannedRightsRequest{
			Peer:         p,
			BannedRights: *rights,
		})
	case *tg.InputPeerChat:
		_, err = c.API.MessagesEditChatDefaultBannedRights(ctx, &tg.MessagesEditChatDefaultBannedRightsRequest{
			Peer:         p,
			BannedRights: *rights,
		})
	default:
		return nil, fmt.Errorf("peer must be a chat or channel")
	}

	if err != nil {
		return nil, fmt.Errorf("failed to set chat permissions: %w", err)
	}

	return &types.SetChatPermissionsResult{Success: true}, nil
}
