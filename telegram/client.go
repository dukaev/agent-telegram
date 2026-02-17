// Package telegram provides a Telegram client wrapper.
package telegram

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync"

	"github.com/gotd/td/session"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/tg"

	"agent-telegram/telegram/chat"
	"agent-telegram/telegram/gift"
	"agent-telegram/telegram/media"
	"agent-telegram/telegram/message"
	"agent-telegram/telegram/pin"
	"agent-telegram/telegram/reaction"
	"agent-telegram/telegram/search"
	"agent-telegram/telegram/user"
)

// Client wraps the Telegram client
type Client struct {
	client         *telegram.Client
	appID          int
	appHash        string
	sessionPath    string
	sessionStorage session.Storage // optional: in-memory session (e.g. from env)
	updateStore    *UpdateStore
	peerCache      sync.Map // username → InputPeerClass cache
	ready          chan struct{} // closed when client is fully initialized
	reloadCh       chan struct{} // signals session reload request
	cancelFn       context.CancelFunc // cancels current client context
	mu             sync.Mutex // protects cancelFn
	// Domain clients
	message  *message.Client
	media    *media.Client
	chat     *chat.Client
	user     *user.Client
	pin      *pin.Client
	reaction *reaction.Client
	search   *search.Client
	gift     *gift.Client
}

// NewClient creates a new Telegram client.
// Domain clients are created lazily in Start() when the Telegram client is ready.
func NewClient(appID int, appHash string) *Client {
	return &Client{
		appID:    appID,
		appHash:  appHash,
		ready:    make(chan struct{}),
		reloadCh: make(chan struct{}, 1),
	}
}

// WithSessionPath sets a custom session path.
func (c *Client) WithSessionPath(path string) *Client {
	c.sessionPath = path
	return c
}

// WithSessionStorage sets a custom session storage (e.g. EnvStorage).
// When set, this takes priority over file-based storage.
func (c *Client) WithSessionStorage(s session.Storage) *Client {
	c.sessionStorage = s
	return c
}

// WithUpdateStore sets a custom update store.
func (c *Client) WithUpdateStore(store *UpdateStore) *Client {
	c.updateStore = store
	return c
}

// Start starts the Telegram client
func (c *Client) Start(ctx context.Context) error {
	// Determine session storage: env-based takes priority over file-based
	var storage session.Storage
	if c.sessionStorage != nil {
		storage = c.sessionStorage
	} else {
		sessionPath, err := c.GetSessionPath()
		if err != nil {
			return err
		}
		storage = &session.FileStorage{Path: sessionPath}
	}

	// Create dispatcher
	dispatcher := tg.NewUpdateDispatcher()
	c.RegisterUpdateHandlers(dispatcher)

	// Create client
	c.client = telegram.NewClient(c.appID, c.appHash, telegram.Options{
		SessionStorage: storage,
		UpdateHandler:  dispatcher,
	})

	c.initDomainClients()

	// Run client
	return c.client.Run(ctx, c.runClient)
}

// GetSessionPath returns the session path to use.
func (c *Client) GetSessionPath() (string, error) {
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
	c.gift = gift.NewClient(c)
}

// runClient is the main client run loop.
func (c *Client) runClient(ctx context.Context) error {
	// Check auth status
	status, err := c.client.Auth().Status(ctx)
	if err != nil {
		return err
	}

	if !status.Authorized {
		// Server mode: don't try to authenticate, just fail
		// User should run 'login' command first
		return fmt.Errorf("not authenticated - please run 'agent-telegram login' first")
	}

	// Get current user and log
	userInfo, err := c.client.Self(ctx)
	if err != nil {
		return err
	}
	slog.Info("Logged in", "first_name", userInfo.FirstName, "username", userInfo.Username)

	// Set API for domain clients
	c.setDomainAPIs()

	// Signal that client is ready
	close(c.ready)

	// Keep running
	<-ctx.Done()
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
	c.gift.SetAPI(api)
}

// ClientStatus represents the current status of the Telegram client.
type ClientStatus struct {
	Initialized bool   `json:"initialized"`
	Authorized  bool   `json:"authorized"`
	Username    string `json:"username,omitempty"`
	FirstName   string `json:"firstName,omitempty"`
	UserID      int64  `json:"userId,omitempty"`
}

// Ready returns a channel that is closed when the client is fully initialized.
func (c *Client) Ready() <-chan struct{} {
	return c.ready
}

// IsInitialized returns true if the client API is ready.
func (c *Client) IsInitialized() bool {
	select {
	case <-c.ready:
		return true
	default:
		return false
	}
}

// GetStatus returns the current client status.
func (c *Client) GetStatus(ctx context.Context) ClientStatus {
	status := ClientStatus{
		Initialized: c.IsInitialized(),
	}

	if !status.Initialized || c.client == nil {
		return status
	}

	// Try to get user info
	userInfo, err := c.client.Self(ctx)
	if err == nil {
		status.Authorized = true
		status.Username = userInfo.Username
		status.FirstName = userInfo.FirstName
		status.UserID = userInfo.ID
	}

	return status
}

// ReloadCh returns the channel that signals reload requests.
func (c *Client) ReloadCh() <-chan struct{} {
	return c.reloadCh
}

// Reload signals the client to reload the session.
// This will disconnect and reconnect with the new session file.
func (c *Client) Reload() {
	// Reset ready channel for new connection
	c.mu.Lock()
	c.ready = make(chan struct{})
	// Clear peer cache since we might be a different user
	c.peerCache = sync.Map{}
	cancelFn := c.cancelFn
	c.mu.Unlock()

	// Signal reload request first
	select {
	case c.reloadCh <- struct{}{}:
	default:
		// Already pending reload
	}

	// Cancel current client to trigger disconnect
	if cancelFn != nil {
		cancelFn()
	}
}

// SetCancelFn stores the cancel function for the current client context.
func (c *Client) SetCancelFn(cancel context.CancelFunc) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cancelFn = cancel
}

// cancelClient cancels the current client connection.
//
//nolint:unused // Reserved for future graceful shutdown support
func (c *Client) cancelClient() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.cancelFn != nil {
		c.cancelFn()
		c.cancelFn = nil
	}
}
