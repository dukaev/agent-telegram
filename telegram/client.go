// Package telegram provides a Telegram client wrapper.
package telegram

import (
	"bufio"
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync"

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
	ready       chan struct{} // closed when client is fully initialized
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

// NewClient creates a new Telegram client.
// Domain clients are created lazily in Start() when the Telegram client is ready.
func NewClient(appID int, appHash, phone string) *Client {
	return &Client{
		appID:   appID,
		appHash: appHash,
		phone:   phone,
		ready:   make(chan struct{}),
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
		UpdateHandler:  dispatcher,
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

	// Signal that client is ready
	close(c.ready)

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
