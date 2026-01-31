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
func (c *Client) GetChats(ctx context.Context, limit, offset int) ([]map[string]interface{}, error) {
	if c.client == nil {
		return nil, fmt.Errorf("client not initialized")
	}

	api := c.client.API()

	// Get dialogs
	dialogsClass, err := api.MessagesGetDialogs(ctx, &tg.MessagesGetDialogsRequest{
		Limit: int(limit),
		OffsetDate: 0,
		OffsetID: 0,
		OffsetPeer: &tg.InputPeerEmpty{},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get dialogs: %w", err)
	}

	// Handle different response types
	var dialogs []tg.DialogClass
	var chats []tg.ChatClass
	var users []tg.UserClass

	switch d := dialogsClass.(type) {
	case *tg.MessagesDialogs:
		dialogs = d.Dialogs
		chats = d.Chats
		users = d.Users
	case *tg.MessagesDialogsSlice:
		dialogs = d.Dialogs
		chats = d.Chats
		users = d.Users
	case *tg.MessagesDialogsNotModified:
		return nil, fmt.Errorf("dialogs not modified")
	default:
		return nil, fmt.Errorf("unexpected dialogs type: %T", d)
	}

	// Build chat map
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

	// Build user map
	userMap := make(map[int64]tg.UserClass)
	for _, u := range users {
		switch user := u.(type) {
		case *tg.User:
			userMap[user.ID] = user
		}
	}

	// Convert to response format
	result := make([]map[string]interface{}, 0, len(dialogs))
	for _, dialogClass := range dialogs {
		dialog, ok := dialogClass.(*tg.Dialog)
		if !ok {
			continue
		}

		chatInfo := map[string]interface{}{
			"peer": dialog.Peer,
		}

		// Get chat details from peer
		switch p := dialog.Peer.(type) {
		case *tg.PeerUser:
			if userClass, ok := userMap[p.UserID]; ok {
				if user, ok := userClass.(*tg.User); ok {
					chatInfo["type"] = "user"
					chatInfo["user_id"] = user.ID
					chatInfo["first_name"] = user.FirstName
					chatInfo["last_name"] = user.LastName
					chatInfo["username"] = user.Username
					if user.Bot {
						chatInfo["bot"] = true
					}
				}
			}
		case *tg.PeerChat:
			if chatClass, ok := chatMap[int64(p.ChatID)]; ok {
				if chat, ok := chatClass.(*tg.Chat); ok {
					chatInfo["type"] = "chat"
					chatInfo["chat_id"] = chat.ID
					chatInfo["title"] = chat.Title
					chatInfo["participants_count"] = chat.ParticipantsCount
				}
			}
		case *tg.PeerChannel:
			if chatClass, ok := chatMap[int64(p.ChannelID)]; ok {
				if channel, ok := chatClass.(*tg.Channel); ok {
					chatInfo["type"] = "channel"
					chatInfo["channel_id"] = channel.ID
					chatInfo["title"] = channel.Title
					chatInfo["username"] = channel.Username
					chatInfo["megagroup"] = channel.Megagroup
				}
			}
		}

		// Add top message info
		if dialog.TopMessage > 0 {
			chatInfo["top_message_id"] = dialog.TopMessage
		}

		chatInfo["unread_count"] = dialog.UnreadCount
		chatInfo["read_inbox_max_id"] = dialog.ReadInboxMaxID
		chatInfo["read_outbox_max_id"] = dialog.ReadOutboxMaxID

		result = append(result, chatInfo)
	}

	return result, nil
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
	dispatcher.OnNewMessage(func(ctx context.Context, e tg.Entities, update *tg.UpdateNewMessage) error {
		peer := "unknown"
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
	dispatcher.OnEditMessage(func(ctx context.Context, e tg.Entities, update *tg.UpdateEditMessage) error {
		peer := "unknown"
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
		return "unknown"
	}
}
