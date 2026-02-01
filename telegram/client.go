// Package telegram provides a Telegram client wrapper.
package telegram

import (
	"bufio"
	"context"
	"fmt"
	"log/slog"
	"os"
	"sync"
	"path/filepath"

	"github.com/gotd/td/session"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/tg"
	"agent-telegram/telegram/chat"
	"agent-telegram/telegram/media"
	"agent-telegram/telegram/message"
	"agent-telegram/telegram/pin"
	"agent-telegram/telegram/reaction"
	"agent-telegram/telegram/search"
	"agent-telegram/telegram/types"
	"agent-telegram/telegram/user"
)

// Client wraps the Telegram client
type Client struct {
	client      *telegram.Client
	appID       int
	appHash     string
	phone       string
	sessionPath string
	updateStore *UpdateStore
	peerCache   sync.Map // username → InputPeerClass cache
	// Domain clients
	message  *message.Client
	media    *media.Client
	chat     *chat.Client
	user     *user.Client
	pin      *pin.Client
	reaction *reaction.Client
	search   *search.Client
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
	slog.Info("Accepted TOS")
	return nil
}

// NewClient creates a new Telegram client
func NewClient(appID int, appHash, phone string) *Client {
	tc := &Client{
		appID:   appID,
		appHash: appHash,
		phone:   phone,
	}
	// Initialize domain clients (will be finalized when telegram client is set)
	tc.message = message.NewClient(tc)
	tc.media = media.NewClient(tc)
	tc.chat = chat.NewClient(tc)
	tc.user = user.NewClient(tc)
	tc.pin = pin.NewClient(tc)
	tc.reaction = reaction.NewClient(tc)
	tc.search = search.NewClient(tc)
	return tc
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
	sessionPath, err := c.getSessionPath()
	if err != nil {
		return err
	}

	// Create dispatcher
	dispatcher := tg.NewUpdateDispatcher()
	c.RegisterUpdateHandlers(dispatcher)

	// Create client
	c.client = telegram.NewClient(c.appID, c.appHash, telegram.Options{
		SessionStorage: &session.FileStorage{Path: sessionPath},
		UpdateHandler: dispatcher,
	})

	c.initDomainClients()

	// Run client
	return c.client.Run(ctx, c.runClient)
}

// getSessionPath returns the session path to use.
func (c *Client) getSessionPath() (string, error) {
	if c.sessionPath != "" {
		return c.sessionPath, nil
	}

	sessionDir := filepath.Join(os.Getenv("HOME"), ".agent-telegram")
	if err := os.MkdirAll(sessionDir, 0700); err != nil {
		return "", fmt.Errorf("failed to create session directory: %w", err)
	}
	return filepath.Join(sessionDir, "session.json"), nil
}

// initDomainClients initializes all domain clients.
func (c *Client) initDomainClients() {
	c.message = message.NewClient(c)
	c.media = media.NewClient(c)
	c.chat = chat.NewClient(c)
	c.user = user.NewClient(c)
	c.pin = pin.NewClient(c)
	c.reaction = reaction.NewClient(c)
	c.search = search.NewClient(c)
}

// runClient is the main client run loop.
func (c *Client) runClient(ctx context.Context) error {
	// Check auth status
	status, err := c.client.Auth().Status(ctx)
	if err != nil {
		return err
	}

	if !status.Authorized {
		if err := c.authenticate(ctx); err != nil {
			return err
		}
	}

	// Get current user and log
	userInfo, err := c.client.Self(ctx)
	if err != nil {
		return err
	}
	slog.Info("Logged in", "first_name", userInfo.FirstName, "username", userInfo.Username)

	// Set API for domain clients
	c.setDomainAPIs()

	// Keep running
	<-ctx.Done()
	return nil
}

// authenticate performs the phone authentication flow.
func (c *Client) authenticate(ctx context.Context) error {
	flow := auth.NewFlow(
		auth.CodeOnly(c.phone, codeAuth{}),
		auth.SendCodeOptions{},
	)

	if err := c.client.Auth().IfNecessary(ctx, flow); err != nil {
		return fmt.Errorf("auth failed: %w", err)
	}
	return nil
}

// setDomainAPIs sets the API client for all domain clients.
func (c *Client) setDomainAPIs() {
	api := c.client.API()
	c.message.SetAPI(api)
	c.media.SetAPI(api)
	c.chat.SetAPI(api)
	c.user.SetAPI(api)
	c.pin.SetAPI(api)
	c.reaction.SetAPI(api)
	c.search.SetAPI(api)
}

// Client returns the underlying telegram.Client
func (c *Client) Client() *telegram.Client {
	return c.client
}

// Message returns the message client.
func (c *Client) Message() *message.Client {
	return c.message
}

// Media returns the media client.
func (c *Client) Media() *media.Client {
	return c.media
}

// Chat returns the chat client.
func (c *Client) Chat() *chat.Client {
	return c.chat
}

// User returns the user client.
func (c *Client) User() *user.Client {
	return c.user
}

// Pin returns the pin client.
func (c *Client) Pin() *pin.Client {
	return c.pin
}

// Reaction returns the reaction client.
func (c *Client) Reaction() *reaction.Client {
	return c.reaction
}

// Search returns the search client.
func (c *Client) Search() *search.Client {
	return c.search
}

// GetMe returns the current user information.
func (c *Client) GetMe(ctx context.Context) (*tg.User, error) {
	if c.client == nil {
		return nil, fmt.Errorf("client not initialized")
	}
	return c.client.Self(ctx)
}

// GetUpdates pops and returns stored updates.
func (c *Client) GetUpdates(limit int) []types.StoredUpdate {
	if c.updateStore == nil {
		return []types.StoredUpdate{}
	}
	return c.updateStore.Get(limit)
}

// InspectReplyKeyboard inspects the reply keyboard from a chat.
func (c *Client) InspectReplyKeyboard(ctx context.Context, params types.PeerInfo) (*types.ReplyKeyboardResult, error) {
	return c.message.InspectReplyKeyboard(ctx, params)
}

// ResolvePeer resolves a peer string to InputPeerClass with caching.
// This method is shared across all domain clients to avoid duplicate API calls.
func (c *Client) ResolvePeer(ctx context.Context, peer string) (tg.InputPeerClass, error) {
	// If peer starts with @, it's a username - resolve it with cache
	if len(peer) > 0 && peer[0] == '@' {
		// Check cache first
		if cached, ok := c.peerCache.Load(peer); ok {
			if inputPeer, ok := cached.(tg.InputPeerClass); ok {
				return inputPeer, nil
			}
		}

		// Not in cache, resolve from API
		peerClass, err := c.client.API().ContactsResolveUsername(ctx, &tg.ContactsResolveUsernameRequest{Username: peer[1:]})
		if err != nil {
			return nil, err
		}

		var inputPeer tg.InputPeerClass
		switch p := peerClass.Peer.(type) {
		case *tg.PeerUser:
			inputPeer = &tg.InputPeerUser{
				UserID:     p.UserID,
				AccessHash: getAccessHashFromPeerClass(peerClass, p.UserID),
			}
		case *tg.PeerChat:
			inputPeer = &tg.InputPeerChat{
				ChatID: p.ChatID,
			}
		case *tg.PeerChannel:
			inputPeer = &tg.InputPeerChannel{
				ChannelID:  p.ChannelID,
				AccessHash: getAccessHashFromPeerClass(peerClass, p.ChannelID),
			}
		default:
			return nil, fmt.Errorf("unknown peer type")
		}

		// Store in cache
		c.peerCache.Store(peer, inputPeer)
		return inputPeer, nil
	}

	// Try to parse as user ID
	// For now, just return empty peer (will be expanded later)
	return &tg.InputPeerEmpty{}, nil
}

// getAccessHashFromPeerClass extracts access hash from the resolved peer.
func getAccessHashFromPeerClass(peerClass *tg.ContactsResolvedPeer, id int64) int64 {
	for _, chat := range peerClass.Chats {
		switch c := chat.(type) {
		case *tg.Channel:
			if c.ID == id {
				return c.AccessHash
			}
		case *tg.Chat:
			if c.ID == id {
				return 0
			}
		}
	}
	for _, user := range peerClass.Users {
		if u, ok := user.(*tg.User); ok && u.ID == id {
			return u.AccessHash
		}
	}
	return 0
}
