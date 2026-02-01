// Package chat provides Telegram dialog operations.
package chat

import (
	"context"
	"fmt"

	"github.com/gotd/td/tg"
)

// GetChats returns the list of dialogs/chats with pagination.
func (c *Client) GetChats(ctx context.Context, limit, _ int) ([]map[string]any, error) {
	if c.API == nil {
		return nil, fmt.Errorf("client not initialized")
	}

	dialogsClass, err := c.API.MessagesGetDialogs(ctx, &tg.MessagesGetDialogsRequest{
		Limit:      limit,
		OffsetDate: 0,
		OffsetID:   0,
		OffsetPeer: &tg.InputPeerEmpty{},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get dialogs: %w", err)
	}

	dialogs, chats, users, err := extractDialogData(dialogsClass)
	if err != nil {
		return nil, err
	}

	chatMap := buildChatMap(chats)
	userMap := buildUserMap(users)

	return convertDialogsToResult(dialogs, chatMap, userMap), nil
}

// extractDialogData extracts dialogs, chats, and users from the response.
func extractDialogData(dialogsClass tg.MessagesDialogsClass) ([]tg.DialogClass, []tg.ChatClass, []tg.UserClass, error) {
	switch d := dialogsClass.(type) {
	case *tg.MessagesDialogs:
		return d.Dialogs, d.Chats, d.Users, nil
	case *tg.MessagesDialogsSlice:
		return d.Dialogs, d.Chats, d.Users, nil
	case *tg.MessagesDialogsNotModified:
		return nil, nil, nil, fmt.Errorf("dialogs not modified")
	default:
		return nil, nil, nil, fmt.Errorf("unexpected dialogs type: %T", d)
	}
}

// buildChatMap builds a map of chat ID to chat class.
func buildChatMap(chats []tg.ChatClass) map[int64]tg.ChatClass {
	chatMap := make(map[int64]tg.ChatClass)
	for _, ch := range chats {
		var id int64
		switch c := ch.(type) {
		case *tg.Chat:
			id = c.ID
		case *tg.Channel:
			id = c.ID
		case *tg.ChatEmpty:
			id = c.ID
		}
		chatMap[id] = ch
	}
	return chatMap
}

// buildUserMap builds a map of user ID to user class.
func buildUserMap(users []tg.UserClass) map[int64]tg.UserClass {
	userMap := make(map[int64]tg.UserClass)
	for _, u := range users {
		if user, ok := u.(*tg.User); ok {
			userMap[user.ID] = user
		}
	}
	return userMap
}

// convertDialogsToResult converts dialogs to the result format.
func convertDialogsToResult(
	dialogs []tg.DialogClass,
	chatMap map[int64]tg.ChatClass,
	userMap map[int64]tg.UserClass,
) []map[string]any {
	result := make([]map[string]any, 0, len(dialogs))
	for _, dialogClass := range dialogs {
		dialog, ok := dialogClass.(*tg.Dialog)
		if !ok {
			continue
		}

		chatInfo := map[string]any{
			"unread_count":      dialog.UnreadCount,
			"read_inbox_max_id": dialog.ReadInboxMaxID,
			"read_outbox_max_id": dialog.ReadOutboxMaxID,
		}

		if dialog.TopMessage > 0 {
			chatInfo["top_message_id"] = dialog.TopMessage
		}

		populateChatInfo(dialog.Peer, chatInfo, chatMap, userMap)
		result = append(result, chatInfo)
	}
	return result
}

// populateChatInfo populates chat info based on peer type.
func populateChatInfo(
	peer tg.PeerClass,
	chatInfo map[string]any,
	chatMap map[int64]tg.ChatClass,
	userMap map[int64]tg.UserClass,
) {
	switch p := peer.(type) {
	case *tg.PeerUser:
		populateUserInfo(p, chatInfo, userMap)
	case *tg.PeerChat:
		populateGroupInfo(p, chatInfo, chatMap)
	case *tg.PeerChannel:
		populateChannelInfo(p, chatInfo, chatMap)
	}
}

// populateUserInfo populates user chat information.
func populateUserInfo(p *tg.PeerUser, chatInfo map[string]any, userMap map[int64]tg.UserClass) {
	userClass, ok := userMap[p.UserID]
	if !ok {
		return
	}

	user, ok := userClass.(*tg.User)
	if !ok {
		return
	}

	chatInfo["type"] = "user"
	chatInfo["user_id"] = user.ID
	chatInfo["first_name"] = user.FirstName
	chatInfo["last_name"] = user.LastName
	chatInfo["username"] = user.Username
	if user.Bot {
		chatInfo["bot"] = true
	}

	// Add string peer for API usage
	if user.Username != "" {
		chatInfo["peer"] = "@" + user.Username
	} else {
		chatInfo["peer"] = fmt.Sprintf("user%d", user.ID)
	}
}

// populateGroupInfo populates group chat information.
func populateGroupInfo(p *tg.PeerChat, chatInfo map[string]any, chatMap map[int64]tg.ChatClass) {
	chatClass, ok := chatMap[p.ChatID]
	if !ok {
		return
	}

	chat, ok := chatClass.(*tg.Chat)
	if !ok {
		return
	}

	chatInfo["type"] = "chat"
	chatInfo["chat_id"] = chat.ID
	chatInfo["title"] = chat.Title
	chatInfo["participants_count"] = chat.ParticipantsCount

	// Add string peer for API usage (use negative chat ID as peer)
	chatInfo["peer"] = fmt.Sprintf("-%d", chat.ID)
}

// populateChannelInfo populates channel chat information.
func populateChannelInfo(p *tg.PeerChannel, chatInfo map[string]any, chatMap map[int64]tg.ChatClass) {
	chatClass, ok := chatMap[p.ChannelID]
	if !ok {
		return
	}

	channel, ok := chatClass.(*tg.Channel)
	if !ok {
		return
	}

	chatInfo["type"] = "channel"
	chatInfo["channel_id"] = channel.ID
	chatInfo["title"] = channel.Title
	chatInfo["username"] = channel.Username
	chatInfo["megagroup"] = channel.Megagroup

	// Add string peer for API usage
	if channel.Username != "" {
		chatInfo["peer"] = "@" + channel.Username
	} else {
		// Use negative channel ID as peer (channel IDs are marked with -100 prefix)
		chatInfo["peer"] = fmt.Sprintf("-100%d", channel.ID)
	}
}
