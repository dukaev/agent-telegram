// Package telegram provides a Telegram client wrapper.
package telegram

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/gotd/td/session"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/tg"
	"golang.org/x/sync/singleflight"

	"agent-telegram/telegram/chat"
	domainclient "agent-telegram/telegram/client"
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
	peerCache      sync.Map           // username → InputPeerClass cache
	peerFlight     singleflight.Group // deduplicates concurrent peer resolutions
	ready          chan struct{}      // closed when client is fully initialized
	reloadCh       chan struct{}      // signals session reload request
	cancelFn       context.CancelFunc // cancels current client context
	mu             sync.Mutex         // protects cancelFn and ready
	runtimeMu      sync.RWMutex       // protects the current Telegram transport
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

// NewClient creates a Telegram facade with stable domain service instances.
func NewClient(appID int, appHash string) *Client {
	c := &Client{
		appID:    appID,
		appHash:  appHash,
		ready:    make(chan struct{}),
		reloadCh: make(chan struct{}, 1),
	}
	c.initDomainClients()
	return c
}

// WithSessionPath is retained for compatibility with older callers.
// Sessions are kept in memory; this value is only reported back by GetSessionPath.
func (c *Client) WithSessionPath(path string) *Client {
	c.sessionPath = path
	return c
}

// WithSessionStorage sets a custom session storage (e.g. EnvStorage).
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
	c.mu.Lock()
	readyGeneration := c.ready
	c.mu.Unlock()
	var storage session.Storage
	if c.sessionStorage != nil {
		storage = c.sessionStorage
	} else {
		storage = NewMemoryStorage(nil)
		c.sessionStorage = storage
	}

	// Create dispatcher
	dispatcher := tg.NewUpdateDispatcher()
	c.RegisterUpdateHandlers(dispatcher)

	// Create a new transport while keeping domain service instances stable.
	tgClient := telegram.NewClient(c.appID, c.appHash, telegram.Options{
		SessionStorage: storage,
		UpdateHandler:  dispatcher,
	})
	c.runtimeMu.Lock()
	c.client = tgClient
	c.runtimeMu.Unlock()

	// Run client
	return tgClient.Run(ctx, func(runCtx context.Context) error {
		return c.runClient(runCtx, tgClient, readyGeneration)
	})
}

// GetSessionPath returns the session path to use.
func (c *Client) GetSessionPath() (string, error) {
	return c.sessionPath, nil
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
func (c *Client) runClient(ctx context.Context, tgClient *telegram.Client, readyGeneration chan struct{}) error {
	// Check auth status
	status, err := tgClient.Auth().Status(ctx)
	if err != nil {
		return err
	}

	if !status.Authorized {
		// Server mode: don't try to authenticate, just fail
		// User should authenticate first.
		return fmt.Errorf("not authenticated - please run 'agent-telegram auth web' first")
	}

	// Get current user and log
	userInfo, err := tgClient.Self(ctx)
	if err != nil {
		return err
	}
	slog.Info("Logged in", "first_name", userInfo.FirstName, "username", userInfo.Username)

	// Set API for domain clients
	c.setDomainAPIs(tgClient.API())

	// Signal readiness only for the generation that started this transport. A
	// concurrent Reload may already have installed a fresh readiness channel.
	c.mu.Lock()
	if c.ready == readyGeneration {
		close(readyGeneration)
	}
	c.mu.Unlock()

	// Keep running
	<-ctx.Done()
	return nil
}

// setDomainAPIs sets the API client for all domain clients.
func (c *Client) setDomainAPIs(api *tg.Client) {
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
	State       string `json:"state"`
	Username    string `json:"username,omitempty"`
	FirstName   string `json:"firstName,omitempty"`
	UserID      int64  `json:"userId,omitempty"`
}

// Ready returns a channel that is closed when the client is fully initialized.
func (c *Client) Ready() <-chan struct{} {
	c.mu.Lock()
	ch := c.ready
	c.mu.Unlock()
	return ch
}

// IsInitialized returns true if the client API is ready.
func (c *Client) IsInitialized() bool {
	c.mu.Lock()
	ch := c.ready
	c.mu.Unlock()
	select {
	case <-ch:
		return true
	default:
		return false
	}
}

// GetStatus returns the current client status.
func (c *Client) GetStatus(ctx context.Context) ClientStatus {
	status := ClientStatus{
		Initialized: c.IsInitialized(),
		State:       "connecting",
	}

	c.runtimeMu.RLock()
	tgClient := c.client
	c.runtimeMu.RUnlock()
	if !status.Initialized || tgClient == nil {
		return status
	}

	// Try to get user info
	userInfo, err := tgClient.Self(ctx)
	if err == nil {
		status.Authorized = true
		status.State = "ready"
		status.Username = userInfo.Username
		status.FirstName = userInfo.FirstName
		status.UserID = userInfo.ID
	} else {
		status.State = "unauthorized"
	}

	return status
}

// Logout invalidates the current Telegram authorization and clears volatile storage.
func (c *Client) Logout(ctx context.Context) error {
	c.runtimeMu.RLock()
	tgClient := c.client
	c.runtimeMu.RUnlock()
	if tgClient == nil {
		return nil
	}
	_, err := tgClient.API().AuthLogOut(ctx)
	if clearer, ok := c.sessionStorage.(interface{ Clear() }); ok {
		clearer.Clear()
	}
	return err
}

// ImportSession imports raw session bytes into in-memory storage.
func (c *Client) ImportSession(ctx context.Context, data []byte) (bool, error) {
	memoryStorage, ok := c.sessionStorage.(*EnvStorage)
	if !ok {
		return false, nil
	}
	if len(data) == 0 {
		return false, nil
	}
	return true, memoryStorage.StoreSession(ctx, data)
}

// ExportSession returns the current in-memory session bytes when available.
func (c *Client) ExportSession() []byte {
	if exporter, ok := c.sessionStorage.(interface{ ExportSession() []byte }); ok {
		return exporter.ExportSession()
	}
	return nil
}

// ReloadCh returns the channel that signals reload requests.
func (c *Client) ReloadCh() <-chan struct{} {
	return c.reloadCh
}

// Reload signals the client to reload the in-memory session.
func (c *Client) Reload() {
	c.mu.Lock()
	// Reset ready channel for new connection
	c.ready = make(chan struct{})
	cancelFn := c.cancelFn
	c.mu.Unlock()

	// Clear resolved entities without replacing synchronization primitives that
	// may still be used by in-flight requests.
	c.peerCache.Clear()

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

// EnsureReady rejects requests while the Telegram transport is connecting or
// reloading. It is used by transport adapters before invoking domain handlers.
func (c *Client) EnsureReady(ctx context.Context) error {
	c.mu.Lock()
	ready := c.ready
	c.mu.Unlock()
	select {
	case <-ready:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	default:
		return domainclient.ErrNotInitialized
	}
}

func (c *Client) currentTelegramClient() *telegram.Client {
	c.runtimeMu.RLock()
	client := c.client
	c.runtimeMu.RUnlock()
	return client
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
