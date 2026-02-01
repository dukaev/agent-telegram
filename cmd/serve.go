// Package cmd provides CLI commands.
package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"

	"agent-telegram/internal/ipc"
	"agent-telegram/internal/paths"
	telegramipc "agent-telegram/internal/telegram/ipc"
	"agent-telegram/telegram"
)

var (
	serveSocket     string
	serveSession    string
	serveForeground bool
)

// serveCmd represents the serve command.
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start IPC server with Telegram (Unix socket)",
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

//nolint:funlen // Server startup logic requires sequential steps
func runServe(_ *cobra.Command, _ []string) {
	_ = godotenv.Load()

	// Validate credentials BEFORE daemonize so user sees errors
	appID, appHash, phone := loadTelegramCredentials()

	if !serveForeground {
		if err := daemonize(); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to daemonize: %v\n", err)
			os.Exit(1)
		}
	}

	// Acquire lock to prevent multiple instances
	lockPath, err := paths.LockFilePath()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to get lock path: %v\n", err)
		os.Exit(1)
	}
	lock := paths.NewLockFile(lockPath)
	acquired, err := lock.TryLock()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to acquire lock: %v\n", err)
		os.Exit(1)
	}
	if !acquired {
		fmt.Fprintln(os.Stderr, "Another server instance is already running")
		os.Exit(1)
	}
	defer func() { _ = lock.Unlock() }()

	// Write PID file (defer cleanup is set up after all early exits)
	pidPath, err := paths.PIDFilePath()
	if err != nil {
		_ = lock.Unlock()
		fmt.Fprintf(os.Stderr, "Failed to get PID path: %v\n", err)
		os.Exit(1) //nolint:gocritic // lock.Unlock() called explicitly
	}
	if err := paths.WritePID(pidPath); err != nil {
		_ = lock.Unlock()
		fmt.Fprintf(os.Stderr, "Failed to write PID file: %v\n", err)
		os.Exit(1)
	}
	defer func() { _ = paths.RemovePID(pidPath) }()

	// Setup slog after daemonize (when in foreground mode or in child process)
	setupLogger()

	ctx, cancel := setupContext()
	socketPath := getSocketPath()
	sessionPath := getSessionPath()

	tgClient := createTelegramClient(appID, appHash, phone, sessionPath)
	startTelegramClient(ctx, tgClient)

	srv := createIPCServer(socketPath, tgClient, cancel)
	if err := srv.Start(ctx); err != nil {
		slog.Error("server error", "error", err)
		os.Exit(1)
	}
	slog.Info("Server stopped")
}

// setupLogger configures structured logging to file in ~/.agent-telegram/.
func setupLogger() {
	logPath, err := paths.LogFilePath()
	if err != nil {
		// Fallback to stderr if path cannot be determined
		slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, nil)))
		slog.Error("Failed to get log path, using stderr", "error", err)
		return
	}

	//nolint:gosec // logPath is from trusted paths.LogFilePath()
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		// Fallback to stderr if file cannot be opened
		slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, nil)))
		slog.Error("Failed to open log file, using stderr", "error", err)
		return
	}

	logger := slog.New(slog.NewJSONHandler(logFile, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)
}

// setupContext creates a context with signal handling.
// Returns both the context and cancel function for use in shutdown handler.
func setupContext() (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(context.Background())

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigCh
		fmt.Println("\nShutting down server...")
		cancel()
	}()

	return ctx, cancel
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
		slog.Error("Invalid APP_ID", "error", err)
		os.Exit(1)
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

// startTelegramClient starts the Telegram client in background with retry logic.
func startTelegramClient(ctx context.Context, tgClient *telegram.Client) {
	const maxRetries = 5

	go func() {
		for retry := 1; retry <= maxRetries; retry++ {
			err := tgClient.Start(ctx)
			if err == nil || ctx.Err() != nil {
				return
			}

			wait := time.Duration(retry) * 5 * time.Second
			slog.Error("telegram client error", "error", err,
				"attempt", retry, "max_retries", maxRetries,
				"retry_after", wait.String())

			select {
			case <-time.After(wait):
				// Continue retry
			case <-ctx.Done():
				return
			}
		}
		slog.Error("telegram client failed after retries", "retries", maxRetries)
	}()
}

// createIPCServer creates and configures the IPC server.
func createIPCServer(socketPath string, tgClient *telegram.Client, cancel context.CancelFunc) *ipc.SocketServer {
	srv := ipc.NewSocketServer(socketPath)
	ipc.RegisterPingPong(srv)
	telegramipc.RegisterHandlers(srv, tgClient)

	srv.Register("status", func(_ json.RawMessage) (any, *ipc.ErrorObject) {
		// Create a short-lived context for the status check
		ctx, statusCancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer statusCancel()

		tgStatus := tgClient.GetStatus(ctx)

		return map[string]any{
			"status":      "running",
			"pid":         os.Getpid(),
			"initialized": tgStatus.Initialized,
			"authorized":  tgStatus.Authorized,
			"username":    tgStatus.Username,
			"first_name":  tgStatus.FirstName,
			"user_id":     tgStatus.UserID,
		}, nil
	})

	srv.Register("shutdown", func(_ json.RawMessage) (any, *ipc.ErrorObject) {
		// Trigger graceful shutdown by canceling context
		go func() {
			time.Sleep(100 * time.Millisecond) // Small delay to send response first
			cancel()
		}()
		return map[string]any{
			"success": true,
			"message": "Shutting down...",
		}, nil
	})

	return srv
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

// isDaemonChild checks if this process is a daemon child.
func isDaemonChild() bool {
	return os.Getenv("AGENT_TELEGRAM_DAEMON") == "1"
}

// daemonize forks the process to run in background.
func daemonize() error {
	// If we're already a daemon child, don't fork again
	if isDaemonChild() {
		return nil
	}

	// Get the path to the current executable
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	// Get log file path
	logPath, err := paths.LogFilePath()
	if err != nil {
		return fmt.Errorf("failed to get log path: %w", err)
	}

	// Create log file
	//nolint:gosec // logPath is from trusted paths.LogFilePath()
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		return fmt.Errorf("failed to create log file: %w", err)
	}

	// Prepare environment with daemon marker
	env := append(os.Environ(), "AGENT_TELEGRAM_DAEMON=1")

	// Fork the process
	attr := &os.ProcAttr{
		Env:   env,
		Files: []*os.File{nil, logFile, logFile}, // stdin: nil, stdout: logFile, stderr: logFile
	}

	proc, err := os.StartProcess(execPath, os.Args, attr)
	if err != nil {
		_ = logFile.Close()
		return fmt.Errorf("failed to start daemon process: %w", err)
	}

	// Parent process exits
	fmt.Printf("Daemon started with PID %d\n", proc.Pid)
	fmt.Printf("Logs: %s\n", logPath)
	os.Exit(0)

	return nil
}
