// Package telegram provides a Telegram client wrapper.
package telegram

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/gotd/td/session"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/tg"
)

const (
	// unknownPeer is the default peer string.
	unknownPeer = "unknown"
)

// Client wraps the Telegram client
type Client struct {
	client      *telegram.Client
	appID       int
	appHash     string
	phone       string
	sessionPath string
	updateStore *UpdateStore
}

// codeAuth reads verification code from stdin
type codeAuth struct{}

func (c codeAuth) Code(_ context.Context, _ *tg.AuthSentCode) (string, error) {
	fmt.Print("Enter verification code: ")
	reader := bufio.NewReader(os.Stdin)
	code, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return code[:len(code)-1], nil
}

func (c codeAuth) AcceptTOS(_ context.Context, _ tg.HelpTermsOfService) error {
	fmt.Println("Accepted TOS")
	return nil
}

// NewClient creates a new Telegram client
func NewClient(appID int, appHash, phone string) *Client {
	return &Client{
		appID:   appID,
		appHash: appHash,
		phone:   phone,
	}
}

// WithSessionPath sets a custom session path.
func (c *Client) WithSessionPath(path string) *Client {
	c.sessionPath = path
	return c
}

// WithUpdateStore sets a custom update store.
func (c *Client) WithUpdateStore(store *UpdateStore) *Client {
	c.updateStore = store
	return c
}

// Start starts the Telegram client
func (c *Client) Start(ctx context.Context) error {
	var sessionPath string

	// Use custom session path if provided, otherwise use default
	if c.sessionPath != "" {
		sessionPath = c.sessionPath
	} else {
		// Create session directory
		sessionDir := filepath.Join(os.Getenv("HOME"), ".agent-telegram")
		if err := os.MkdirAll(sessionDir, 0700); err != nil {
			return fmt.Errorf("failed to create session directory: %w", err)
		}
		sessionPath = filepath.Join(sessionDir, "session.json")
	}

	// Create dispatcher
	dispatcher := tg.NewUpdateDispatcher()

	// Register update handlers if store is configured
	c.RegisterUpdateHandlers(dispatcher)

	// Create client
	c.client = telegram.NewClient(c.appID, c.appHash, telegram.Options{
		SessionStorage: &session.FileStorage{
			Path: sessionPath,
		},
		UpdateHandler: dispatcher,
	})

	// Run client
	return c.client.Run(ctx, func(ctx context.Context) error {
		// Check auth status
		status, err := c.client.Auth().Status(ctx)
		if err != nil {
			return err
		}

		if !status.Authorized {
			// Phone auth flow
			flow := auth.NewFlow(
				auth.CodeOnly(c.phone, codeAuth{}),
				auth.SendCodeOptions{},
			)

			if err := c.client.Auth().IfNecessary(ctx, flow); err != nil {
				return fmt.Errorf("auth failed: %w", err)
			}
		}

		// Get current user
		user, err := c.client.Self(ctx)
		if err != nil {
			return err
		}

		fmt.Printf("Logged in as: %s (@%s)\n", user.FirstName, user.Username)

		// Keep running
		<-ctx.Done()
		return nil
	})
}

// Client returns the underlying telegram.Client
func (c *Client) Client() *telegram.Client {
	return c.client
}

// GetMe returns the current user information.
func (c *Client) GetMe(ctx context.Context) (*tg.User, error) {
	if c.client == nil {
		return nil, fmt.Errorf("client not initialized")
	}
	return c.client.Self(ctx)
}

// GetChats returns the list of dialogs/chats with pagination.
func (c *Client) GetChats(ctx context.Context, limit, _ int) ([]map[string]interface{}, error) {
	if c.client == nil {
		return nil, fmt.Errorf("client not initialized")
	}

	api := c.client.API()

	dialogsClass, err := api.MessagesGetDialogs(ctx, &tg.MessagesGetDialogsRequest{
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
func convertDialogsToResult(dialogs []tg.DialogClass, chatMap map[int64]tg.ChatClass, userMap map[int64]tg.UserClass) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(dialogs))
	for _, dialogClass := range dialogs {
		dialog, ok := dialogClass.(*tg.Dialog)
		if !ok {
			continue
		}

		chatInfo := map[string]interface{}{
			"peer":                dialog.Peer,
			"unread_count":        dialog.UnreadCount,
			"read_inbox_max_id":   dialog.ReadInboxMaxID,
			"read_outbox_max_id":  dialog.ReadOutboxMaxID,
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
func populateChatInfo(peer tg.PeerClass, chatInfo map[string]interface{}, chatMap map[int64]tg.ChatClass, userMap map[int64]tg.UserClass) {
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
func populateUserInfo(p *tg.PeerUser, chatInfo map[string]interface{}, userMap map[int64]tg.UserClass) {
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
}

// populateGroupInfo populates group chat information.
func populateGroupInfo(p *tg.PeerChat, chatInfo map[string]interface{}, chatMap map[int64]tg.ChatClass) {
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
}

// populateChannelInfo populates channel chat information.
func populateChannelInfo(p *tg.PeerChannel, chatInfo map[string]interface{}, chatMap map[int64]tg.ChatClass) {
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
}

// GetUpdates pops and returns stored updates.
func (c *Client) GetUpdates(limit int) []StoredUpdate {
	if c.updateStore == nil {
		return []StoredUpdate{}
	}
	return c.updateStore.Get(limit)
}

// RegisterUpdateHandlers registers update handlers on the dispatcher.
func (c *Client) RegisterUpdateHandlers(dispatcher tg.UpdateDispatcher) {
	if c.updateStore == nil {
		return
	}

	// New messages
	dispatcher.OnNewMessage(func(_ context.Context, _ tg.Entities, update *tg.UpdateNewMessage) error {
		peer := unknownPeer
		if msg, ok := update.Message.(*tg.Message); ok && msg.PeerID != nil {
			peer = peerToString(msg.PeerID)
		}
		c.updateStore.Add(NewStoredUpdate(UpdateTypeNewMessage, map[string]interface{}{
			"message": MessageData(update.Message),
			"peer":    peer,
		}))
		return nil
	})

	// Edited messages
	dispatcher.OnEditMessage(func(_ context.Context, _ tg.Entities, update *tg.UpdateEditMessage) error {
		peer := unknownPeer
		if msg, ok := update.Message.(*tg.Message); ok && msg.PeerID != nil {
			peer = peerToString(msg.PeerID)
		}
		c.updateStore.Add(NewStoredUpdate(UpdateTypeEditMessage, map[string]interface{}{
			"message": MessageData(update.Message),
			"peer":    peer,
		}))
		return nil
	})
}

// peerToString converts a PeerClass to a string representation.
func peerToString(peer tg.PeerClass) string {
	switch p := peer.(type) {
	case *tg.PeerUser:
		return fmt.Sprintf("user:%d", p.UserID)
	case *tg.PeerChat:
		return fmt.Sprintf("chat:%d", p.ChatID)
	case *tg.PeerChannel:
		return fmt.Sprintf("channel:%d", p.ChannelID)
	default:
		return unknownPeer
	}
}
