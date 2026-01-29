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
	client  *telegram.Client
	appID   int
	appHash string
	phone   string
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

// Start starts the Telegram client
func (c *Client) Start(ctx context.Context) error {
	// Create session directory
	sessionDir := filepath.Join(os.Getenv("HOME"), ".agent-telegram")
	if err := os.MkdirAll(sessionDir, 0700); err != nil {
		return fmt.Errorf("failed to create session directory: %w", err)
	}

	sessionPath := filepath.Join(sessionDir, "session.json")

	// Create dispatcher
	dispatcher := tg.NewUpdateDispatcher()

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
