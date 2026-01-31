// Package serve provides the serve command implementation.
package serve

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/joho/godotenv"

	"agent-telegram/internal/ipc"
	telegramipc "agent-telegram/internal/telegram/ipc"
	"agent-telegram/telegram"
)

// getEnv returns the first non-empty environment variable from the given keys.
func getEnv(keys ...string) string {
	for _, key := range keys {
		if val := os.Getenv(key); val != "" {
			return val
		}
	}
	return ""
}

// Run executes the serve command with the given arguments.
func Run(args []string) {
	// Load .env file (ignore error if file doesn't exist)
	_ = godotenv.Load()

	serveCmd := flag.NewFlagSet("serve", flag.ExitOnError)
	socketPath := serveCmd.String("socket", ipc.DefaultSocketPath(), "Path to Unix socket")
	sessionPath := serveCmd.String("session", "", "Path to Telegram session file (default: ~/.agent-telegram/session.json, or AGENT_TELEGRAM_SESSION_PATH env)")

	if err := serveCmd.Parse(args); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing flags: %v\n", err)
		os.Exit(1)
	}

	// Use AGENT_TELEGRAM_SESSION_PATH env var if session flag is not set
	if *sessionPath == "" {
		*sessionPath = os.Getenv("AGENT_TELEGRAM_SESSION_PATH")
	}

	// Create context with cancellation
	ctx, cancel := context.WithCancel(context.Background())

	// Handle interrupt signals
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigCh
		fmt.Println("\nShutting down server...")
		cancel()
	}()

	// Setup Telegram client from environment
	// Check both TELEGRAM_* and AGENT_TELEGRAM_* prefixes
	appIDStr := getEnv("TELEGRAM_APP_ID", "AGENT_TELEGRAM_APP_ID")
	appHash := getEnv("TELEGRAM_APP_HASH", "AGENT_TELEGRAM_APP_HASH")
	phone := os.Getenv("TELEGRAM_PHONE") // Phone is optional for serve (session already exists)

	if appIDStr == "" || appHash == "" {
		cancel()
		fmt.Fprintf(os.Stderr, "Missing Telegram credentials. Set TELEGRAM_APP_ID and TELEGRAM_APP_HASH (or AGENT_TELEGRAM_APP_ID and AGENT_TELEGRAM_APP_HASH) in .env or environment.\n")
		os.Exit(1)
	}

	appID, err := strconv.Atoi(appIDStr)
	if err != nil {
		cancel()
		log.Fatalf("Invalid APP_ID: %v", err)
	}

	tgClient := telegram.NewClient(appID, appHash, phone)
	if *sessionPath != "" {
		tgClient = tgClient.WithSessionPath(*sessionPath)
	}

	// Start Telegram client in background
	go func() {
		if err := tgClient.Start(ctx); err != nil {
			log.Printf("Telegram client error: %v", err)
		}
	}()

	// Create and configure IPC server
	srv := ipc.NewSocketServer(*socketPath)
	ipc.RegisterPingPong(srv)
	telegramipc.RegisterHandlers(srv, tgClient)

	// Register additional methods
	srv.Register("status", func(_ json.RawMessage) (interface{}, *ipc.ErrorObject) {
		return map[string]interface{}{
			"status": "running",
			"pid":    os.Getpid(),
		}, nil
	})

	// Start server
	if err := srv.Start(ctx); err != nil {
		log.Fatalf("Server error: %v", err)
	}

	fmt.Println("Server stopped.")
	cancel()
}
