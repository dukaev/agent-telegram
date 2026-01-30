// Package main provides the agent-telegram CLI tool.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"

	"agent-telegram/cli"
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
	fmt.Println("  login       Interactive login (creates .env)")
	fmt.Println("  version     Print version information")
	fmt.Println("\nTelegram options:")
	fmt.Println("  -app-id     Telegram API App ID (from my.telegram.org)")
	fmt.Println("  -app-hash   Telegram API App Hash (from my.telegram.org)")
	fmt.Println("  -phone      Phone number (e.g. +1234567890)")
	fmt.Println("\nServe options:")
	fmt.Println("  -port       Port to listen on (default: 8080)")
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

func runLogin(_ []string) {
	// Get project directory
	projectDir, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting working directory: %v\n", err)
		os.Exit(1)
	}

	envPath := filepath.Join(projectDir, ".env")

	// Check if .env exists
	if _, err := os.Stat(envPath); os.IsNotExist(err) {
		// Run interactive login UI
		phone, code, password, err := cli.RunLoginUI()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		// Validate inputs
		if phone == "" {
			fmt.Fprintln(os.Stderr, "Error: Phone is required")
			os.Exit(1)
		}

		// Save to .env
		if err := cli.SaveEnvFile(projectDir, phone); err != nil {
			fmt.Fprintf(os.Stderr, "Error saving .env file: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("\n✓ Saved credentials to %s\n", envPath)
		if code != "" {
			fmt.Printf("✓ Verification code: %s\n", code)
		}
		if password != "" {
			fmt.Printf("✓ 2FA password: %s\n", password)
		}
	}

	fmt.Println("ok")
}
