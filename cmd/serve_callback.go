// Package cmd provides the serve-callback command.
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"agent-telegram/internal/callback"
	"agent-telegram/internal/config"
	"agent-telegram/internal/paths"
	"agent-telegram/telegram"
)

var (
	serveCallbackPort    int
	serveCallbackSession string
)

var serveCallbackCmd = &cobra.Command{
	Use:   "serve-callback",
	Short: "Start Telegram client with HTTP callback API",
	Long: `Start a Telegram client and an HTTP API server for managing callbacks.

Use the API to set your callback URL and subscribe to channel events:

  POST /callback/set-callback-url    {"callback_url": "https://example.com/hook"}
  GET  /callback/get-callback-info
  POST /callback/subscribe-channel   {"channel_id": "@mychannel", "event_types": "new_post,edit_post"}
  GET  /callback/subscriptions-list
  POST /callback/unsubscribe         {"subscription_id": 1}

The callback URL is verified before activation: a POST with {"verify_code": "..."}
is sent and the response must contain that code.`,
	GroupID: GroupIDServer,
	Run:     runServeCallback,
}

func init() {
	RootCmd.AddCommand(serveCallbackCmd)

	serveCallbackCmd.Flags().IntVar(&serveCallbackPort, "port", 3000, "HTTP API server port")
	serveCallbackCmd.Flags().StringVar(&serveCallbackSession, "session", "", "Path to Telegram session file")
}

func runServeCallback(_ *cobra.Command, _ []string) {
	storedCfg, err := config.LoadStoredConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	ctx, _ := setupContext()
	setupLogger()

	// Prepare state directory
	dir, err := paths.EnsureConfigDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Load persistent callback state
	store, err := callback.NewStore(dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading callback state: %v\n", err)
		os.Exit(1)
	}

	// Build manager (routes updates to subscriptions → callback URL)
	manager := callback.NewManager(store)

	// Attach manager as update callback
	updateStore := telegram.NewUpdateStore(1000)
	updateStore.SetOnUpdate(manager.HandleUpdate)

	// Create and start Telegram client
	tgClient := createTelegramClient(storedCfg.AppID, storedCfg.AppHash, serveCallbackSession)
	tgClient = tgClient.WithUpdateStore(updateStore)

	startTelegramClient(ctx, tgClient)
	waitForTelegramReady(ctx, tgClient)

	// Start the webhook sender goroutine
	go manager.Run(ctx)

	// Start HTTP API server
	apiServer := callback.NewServer(manager, serveCallbackPort)
	fmt.Fprintf(os.Stderr, "Callback API on http://localhost:%d\n", serveCallbackPort)

	if err := apiServer.Start(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "HTTP server error: %v\n", err)
		os.Exit(1)
	}

	fmt.Fprintln(os.Stderr, "serve-callback stopped.")
}
