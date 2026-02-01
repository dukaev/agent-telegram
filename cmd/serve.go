// Package cmd provides CLI commands.
package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"

	"agent-telegram/internal/ipc"
	telegramipc "agent-telegram/internal/telegram/ipc"
	"agent-telegram/telegram"
)

var (
	serveSocket   string
	serveSession  string
	serveForeground bool
)

// serveCmd represents the serve command.
var serveCmd = &cobra.Command{
	Use:     "serve",
	Short:   "Start IPC server with Telegram (Unix socket)",
	Long: `Start the IPC server with a Telegram client in the background.

The server listens on a Unix socket and handles requests from other commands.
Telegram client runs in background and stays connected.`,
	Run:     runServe,
	GroupID: GroupIDServer,
}

func init() {
	RootCmd.AddCommand(serveCmd)

	serveCmd.Flags().StringVarP(&serveSocket, "socket", "s", "",
		"Path to Unix socket (default: /tmp/agent-telegram.sock)")
	serveCmd.Flags().StringVarP(&serveSession, "session", "", "",
		"Path to Telegram session file (default: ~/.agent-telegram/session.json)")
	serveCmd.Flags().BoolVarP(&serveForeground, "foreground", "f", false,
		"Run in foreground (default: background)")
}

func runServe(_ *cobra.Command, _ []string) {
	_ = godotenv.Load()

	if !serveForeground {
		if err := daemonize(); err != nil {
			log.Fatalf("Failed to daemonize: %v", err)
		}
	}

	ctx := setupContext()
	socketPath := getSocketPath()
	sessionPath := getSessionPath()
	appID, appHash, phone := loadTelegramCredentials()

	tgClient := createTelegramClient(appID, appHash, phone, sessionPath)
	startTelegramClient(ctx, tgClient)

	srv := createIPCServer(socketPath, tgClient)
	startServer(ctx, srv)

	fmt.Println("Server stopped.")
}

// setupContext creates a context with signal handling.
func setupContext() context.Context {
	ctx, cancel := context.WithCancel(context.Background())

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigCh
		fmt.Println("\nShutting down server...")
		cancel()
	}()

	return ctx
}

// getSocketPath returns the socket path from flags or default.
func getSocketPath() string {
	socketPath, _ := RootCmd.Flags().GetString("socket")
	if serveSocket != "" {
		socketPath = serveSocket
	}
	if socketPath == "" {
		socketPath = "/tmp/agent-telegram.sock"
	}
	return socketPath
}

// getSessionPath returns the session path from flags or environment.
func getSessionPath() string {
	sessionPath := serveSession
	if sessionPath == "" {
		sessionPath = os.Getenv("AGENT_TELEGRAM_SESSION_PATH")
	}
	return sessionPath
}

// loadTelegramCredentials loads and validates Telegram credentials.
func loadTelegramCredentials() (appID int, appHash, phone string) {
	appIDStr := getEnv("TELEGRAM_APP_ID", "AGENT_TELEGRAM_APP_ID")
	appHash = getEnv("TELEGRAM_APP_HASH", "AGENT_TELEGRAM_APP_HASH")
	phone = os.Getenv("TELEGRAM_PHONE")

	if appIDStr == "" || appHash == "" {
		fmt.Fprintf(os.Stderr,
			"Missing Telegram credentials. Set TELEGRAM_APP_ID and "+
				"TELEGRAM_APP_HASH (or AGENT_TELEGRAM_APP_ID and AGENT_TELEGRAM_APP_HASH) "+
				"in .env or environment.\n")
		os.Exit(1)
	}

	var err error
	appID, err = strconv.Atoi(appIDStr)
	if err != nil {
		log.Fatalf("Invalid APP_ID: %v", err)
	}

	return appID, appHash, phone
}

// createTelegramClient creates and configures the Telegram client.
func createTelegramClient(appID int, appHash, phone, sessionPath string) *telegram.Client {
	tgClient := telegram.NewClient(appID, appHash, phone)
	if sessionPath != "" {
		tgClient = tgClient.WithSessionPath(sessionPath)
	}
	return tgClient.WithUpdateStore(telegram.NewUpdateStore(1000))
}

// startTelegramClient starts the Telegram client in background.
func startTelegramClient(ctx context.Context, tgClient *telegram.Client) {
	go func() {
		if err := tgClient.Start(ctx); err != nil {
			log.Printf("Telegram client error: %v", err)
		}
	}()
}

// createIPCServer creates and configures the IPC server.
func createIPCServer(socketPath string, tgClient *telegram.Client) *ipc.SocketServer {
	srv := ipc.NewSocketServer(socketPath)
	ipc.RegisterPingPong(srv)
	telegramipc.RegisterHandlers(srv, tgClient)

	srv.Register("status", func(_ json.RawMessage) (any, *ipc.ErrorObject) {
		return map[string]any{
			"status": "running",
			"pid":    os.Getpid(),
		}, nil
	})

	srv.Register("shutdown", func(_ json.RawMessage) (any, *ipc.ErrorObject) {
		// Trigger graceful shutdown by canceling context
		return map[string]any{
			"success": true,
			"message": "Shutting down...",
		}, nil
	})

	return srv
}

// startServer starts the IPC server.
func startServer(ctx context.Context, srv *ipc.SocketServer) {
	if err := srv.Start(ctx); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

// getEnv returns the first non-empty environment variable from the given keys.
func getEnv(keys ...string) string {
	for _, key := range keys {
		if val := os.Getenv(key); val != "" {
			return val
		}
	}
	return ""
}

// daemonize forks the process to run in background.
func daemonize() error {
	// Get the path to the current executable
	exec, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	// Create log file
	logFile, err := os.OpenFile("agent-telegram.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		return fmt.Errorf("failed to create log file: %w", err)
	}

	// Fork the process
	attr := &os.ProcAttr{
		Files: []*os.File{nil, logFile, logFile}, // stdin: nil, stdout: logFile, stderr: logFile
	}

	proc, err := os.StartProcess(exec, os.Args, attr)
	if err != nil {
		_ = logFile.Close()
		return fmt.Errorf("failed to start daemon process: %w", err)
	}

	// Parent process exits
	fmt.Printf("Daemon started with PID %d\n", proc.Pid)
	fmt.Println("Logs are being written to agent-telegram.log")
	os.Exit(0)

	return nil
}
