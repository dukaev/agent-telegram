// Package cmd provides the webhook command.
package cmd

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"agent-telegram/internal/config"
	"agent-telegram/internal/webhook"
	"agent-telegram/telegram"
)

var (
	webhookURL        string
	webhookTo         string
	webhookType       string
	webhookRetries    int
	webhookRetryDelay int
	webhookSession    string
)

// webhookCmd starts a standalone Telegram client that forwards updates to a URL.
var webhookCmd = &cobra.Command{
	Use:   "webhook",
	Short: "Forward Telegram updates to an HTTP endpoint",
	Long: `Start a Telegram client that forwards matching updates to the specified URL.

Each update is sent as a POST request with a JSON body.

Example:
  agent-telegram webhook --url https://example.com/hook
  agent-telegram webhook --url https://example.com/hook --to @mychannel --type new_message`,
	GroupID: GroupIDServer,
	Run:     runWebhook,
}

func init() {
	RootCmd.AddCommand(webhookCmd)

	webhookCmd.Flags().StringVar(&webhookURL, "url", "", "HTTP endpoint to POST updates to (required)")
	webhookCmd.Flags().StringVarP(&webhookTo, "to", "t", "", "Filter updates by peer (@username, channel ID, etc.)")
	webhookCmd.Flags().StringVar(&webhookType, "type", "", "Filter by update type (new_message, edit_message, star_gift)")
	webhookCmd.Flags().IntVar(&webhookRetries, "retries", 3, "Number of delivery retries on failure")
	webhookCmd.Flags().IntVar(&webhookRetryDelay, "retry-delay", 2, "Base retry delay in seconds (exponential backoff)")
	webhookCmd.Flags().StringVar(&webhookSession, "session", "", "Path to Telegram session file")

	_ = webhookCmd.MarkFlagRequired("url")
}

func runWebhook(_ *cobra.Command, _ []string) {
	storedCfg, err := config.LoadStoredConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	ctx, _ := setupContext()
	setupLogger()

	// Build filter
	filter := webhook.Filter{}
	if webhookTo != "" {
		filter.Peer = webhookTo
	}
	if webhookType != "" {
		filter.Types = strings.Split(webhookType, ",")
	}

	// Create webhook sender
	sender := webhook.New(
		webhookURL,
		webhook.WithRetries(webhookRetries),
		webhook.WithRetryDelay(time.Duration(webhookRetryDelay)*time.Second),
		webhook.WithFilter(filter),
	)

	// Create update store and attach sender callback
	store := telegram.NewUpdateStore(1000)
	store.SetOnUpdate(sender.Send)

	// Create Telegram client
	sessionPath := webhookSession
	tgClient := createTelegramClient(storedCfg.AppID, storedCfg.AppHash, sessionPath)
	tgClient = tgClient.WithUpdateStore(store)

	// Start webhook sender goroutine
	go sender.Run(ctx)

	// Start Telegram client with retry logic
	startTelegramClient(ctx, tgClient)

	// Wait for ready
	waitForTelegramReady(ctx, tgClient)

	fmt.Fprintf(os.Stderr, "Webhook active — forwarding updates to %s\n", webhookURL)

	// Block until shutdown
	<-ctx.Done()
	fmt.Fprintln(os.Stderr, "Webhook stopped.")
}
