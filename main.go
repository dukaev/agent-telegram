// Package main provides the agent-telegram CLI tool.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"agent-telegram/cli"
	"agent-telegram/internal/auth"
	"agent-telegram/telegram"
)

var (
	version = "dev"
)

func main() {
	// Check if subcommand is provided
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "serve":
		runServe(os.Args[2:])
	case "telegram":
		runTelegram(os.Args[2:])
	case "login":
		runLogin(os.Args[2:])
	case "version":
		fmt.Printf("agent-telegram %s\n", version)
	default:
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Usage: agent-telegram <command>")
	fmt.Println("\nCommands:")
	fmt.Println("  serve       Start the HTTP server")
	fmt.Println("  telegram    Connect to Telegram")
	fmt.Println("  login       Interactive login with Telegram authentication")
	fmt.Println("  version     Print version information")
	fmt.Println("\nLogin options:")
	fmt.Println("  -app-id     Telegram API App ID (from my.telegram.org)")
	fmt.Println("  -app-hash   Telegram API App Hash (from my.telegram.org)")
	fmt.Println("  -phone      Phone number (optional, can enter in UI)")
	fmt.Println("\nEnvironment variables:")
	fmt.Println("  AGENT_TELEGRAM_APP_ID     Telegram API App ID")
	fmt.Println("  AGENT_TELEGRAM_APP_HASH   Telegram API App Hash")
	fmt.Println("  AGENT_TELEGRAM_PHONE      Phone number (optional)")
	fmt.Println("\nExamples:")
	fmt.Println("  agent-telegram login -app-id 12345 -app-hash abcdef")
	fmt.Println("  AGENT_TELEGRAM_APP_ID=12345 AGENT_TELEGRAM_APP_HASH=abcdef agent-telegram login")
}

func runServe(args []string) {
	serveCmd := flag.NewFlagSet("serve", flag.ExitOnError)
	servePort := serveCmd.String("port", "8080", "Port to listen on")

	if err := serveCmd.Parse(args); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing flags: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Starting server on port %s...\n", *servePort)
	// Server implementation goes here
}

func runTelegram(args []string) {
	tgCmd := flag.NewFlagSet("telegram", flag.ExitOnError)

	appIDStr := tgCmd.String("app-id", os.Getenv("TELEGRAM_APP_ID"), "Telegram API App ID")
	appHash := tgCmd.String("app-hash", os.Getenv("TELEGRAM_APP_HASH"), "Telegram API App Hash")
	phone := tgCmd.String("phone", os.Getenv("TELEGRAM_PHONE"), "Phone number")

	if err := tgCmd.Parse(args); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing flags: %v\n", err)
		os.Exit(1)
	}

	// Validate required parameters
	if *appIDStr == "" || *appHash == "" || *phone == "" {
		fmt.Fprintln(os.Stderr, "Error: Missing required parameters")
		fmt.Fprintln(os.Stderr, "\nSet environment variables:")
		fmt.Fprintln(os.Stderr, "  export TELEGRAM_APP_ID=your_app_id")
		fmt.Fprintln(os.Stderr, "  export TELEGRAM_APP_HASH=your_app_hash")
		fmt.Fprintln(os.Stderr, "  export TELEGRAM_PHONE=+1234567890")
		fmt.Fprintln(os.Stderr, "\nOr use flags:")
		fmt.Fprintln(os.Stderr, "  agent-telegram telegram -app-id <id> -app-hash <hash> -phone <number>")
		os.Exit(1)
	}

	// Parse app ID
	appID, err := strconv.Atoi(*appIDStr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Invalid app-id: %v\n", err)
		os.Exit(1)
	}

	// Create context with cancellation
	ctx, cancel := context.WithCancel(context.Background())

	// Handle interrupt signals
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigCh
		fmt.Println("\nShutting down...")
		cancel()
	}()

	// Create and start Telegram client
	client := telegram.NewClient(appID, *appHash, *phone)
	if err := client.Start(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		cancel()
		os.Exit(1)
	}
	// Cancel context on normal exit
	cancel()
}

func runLogin(args []string) {
	loginCmd := flag.NewFlagSet("login", flag.ExitOnError)
	appIDStr := loginCmd.String("app-id", os.Getenv("AGENT_TELEGRAM_APP_ID"), "Telegram API App ID")
	appHash := loginCmd.String("app-hash", os.Getenv("AGENT_TELEGRAM_APP_HASH"), "Telegram API App Hash")
	phone := loginCmd.String("phone", os.Getenv("AGENT_TELEGRAM_PHONE"), "Phone number")
	mockMode := loginCmd.Bool("mock", false, "Mock mode for UI testing (no real API calls)")

	if err := loginCmd.Parse(args); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing flags: %v\n", err)
		os.Exit(1)
	}

	// Parse app ID (optional, uses hardcoded default if not provided)
	var appID int
	var err error
	if *appIDStr != "" {
		appID, err = strconv.Atoi(*appIDStr)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Invalid app-id: %v\n", err)
			os.Exit(1)
		}
	}

	// Create context
	ctx := context.Background()

	// Create auth service (skip in mock mode)
	var authService *auth.Service
	if !*mockMode {
		if *phone != "" {
			if appID != 0 && *appHash != "" {
				authService, err = auth.NewServiceWithConfig(ctx, appID, *appHash, *phone)
			} else {
				authService, err = auth.NewService(ctx)
			}
		} else {
			authService, err = auth.NewService(ctx)
		}
		if err != nil {
			log.Fatalf("Failed to create auth service: %v", err)
		}
	}

	// Run interactive login UI with Telegram authentication
	sessionPath, err := cli.RunLoginUIWithAuth(ctx, authService)
	if err != nil {
		log.Fatalf("Login failed: %v", err)
	}

	fmt.Printf("\n✓ Session saved to: %s\n", sessionPath)
	fmt.Println("✓ You can now use the telegram command to connect")
}
