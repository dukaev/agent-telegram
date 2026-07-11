// Package cmd provides CLI commands.
package cmd

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/gotd/td/session"
	"github.com/spf13/cobra"

	"agent-telegram/internal/config"
	"agent-telegram/internal/ipc"
	"agent-telegram/internal/observability"
	"agent-telegram/internal/paths"
	"agent-telegram/internal/policy"
	"agent-telegram/internal/sessionstore"
	telegramipc "agent-telegram/internal/telegram/ipc"
	"agent-telegram/telegram"
)

const envTelegramSession = "TELEGRAM_SESSION"
const envLogoutOnStop = "AGENT_TELEGRAM_LOGOUT_ON_STOP"

var (
	serveSocket       string
	serveForeground   bool
	serveLogoutOnStop bool
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
	serveCmd.Flags().BoolVarP(&serveForeground, "foreground", "f", false,
		"Run in foreground (default: background)")
	serveCmd.Flags().BoolVar(&serveLogoutOnStop, "logout-on-stop", false,
		"Logout from Telegram when the server stops")
}

//nolint:funlen // Server startup logic requires sequential steps
func runServe(cmd *cobra.Command, _ []string) {
	// Load credentials from config.json (saved by auth commands)
	storedCfg, err := config.LoadStoredConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	appID, appHash := storedCfg.AppID, storedCfg.AppHash

	socketPath := getSocketPath()

	if !serveForeground {
		if err := daemonize(socketPath); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to daemonize: %v\n", err)
			os.Exit(1)
		}
	}

	// Acquire lock to prevent multiple instances
	lockPath, err := paths.LockFilePathForSocket(socketPath)
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
	pidPath, err := paths.PIDFilePathForSocket(socketPath)
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
	setupLoggerForSocket(socketPath)

	tgClient := createTelegramClient(appID, appHash, telegramClientOptions{
		Provider: firstConfigured(sessionFlagValue(cmd, "session-provider"), storedCfg.SessionProvider),
		Profile:  firstConfigured(sessionFlagValue(cmd, "profile"), storedCfg.SessionProfile),
	})
	logoutOnStop := boolFromEnv(envLogoutOnStop, serveLogoutOnStop)
	var logoutOnce sync.Once
	logout := func() {
		logoutOnce.Do(func() {
			logoutTelegramClient(tgClient, logoutOnStop)
		})
	}

	ctx, cancel := setupContextWithShutdown(logout)

	startTelegramClient(ctx, tgClient)
	go waitForTelegramReady(ctx, tgClient)

	srv := createIPCServer(socketPath, tgClient, cancel)
	if err := srv.Start(ctx); err != nil {
		slog.Error("server error", "error", err)
		os.Exit(1)
	}
	logout()
	slog.Info("Server stopped")
}

// setupLogger configures structured logging to file in ~/.agent-telegram/.
func setupLogger() {
	setupLoggerForSocket("")
}

// setupLoggerForSocket configures structured logging for one socket instance.
func setupLoggerForSocket(socketPath string) {
	logPath, err := paths.LogFilePathForSocket(socketPath)
	if err != nil {
		// Fallback to stderr if path cannot be determined
		slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, nil)))
		slog.Error("Failed to get log path, using stderr", "error", err)
		return
	}

	logWriter, err := observability.NewRotatingWriter(logPath)
	if err != nil {
		// Fallback to stderr if file cannot be opened
		slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, nil)))
		slog.Error("Failed to open log file, using stderr", "error", err)
		return
	}

	logger := slog.New(slog.NewJSONHandler(logWriter, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)
}

func setupContextWithShutdown(beforeCancel func()) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(context.Background())

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigCh
		fmt.Println("\nShutting down server...")
		if beforeCancel != nil {
			beforeCancel()
		}
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
		socketPath = paths.DefaultSocketPath
	}
	return socketPath
}

type telegramClientOptions struct {
	SessionData []byte
	Provider    string
	Profile     string
}

// createTelegramClient creates and configures the Telegram client.
// Explicit bytes and TELEGRAM_SESSION stay in memory. Otherwise the selected
// pluggable provider is used (native Keychain on macOS builds).
func createTelegramClient(appID int, appHash string, opts telegramClientOptions) *telegram.Client {
	tgClient := telegram.NewClient(appID, appHash)
	var storage session.Storage
	if len(opts.SessionData) > 0 {
		storage = telegram.NewMemoryStorage(opts.SessionData)
		slog.Info("Using caller-provided in-memory Telegram session")
	}
	if envSession := os.Getenv(envTelegramSession); envSession != "" {
		envStorage, err := telegram.NewEnvStorage(envSession)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: invalid %s: %v\n", envTelegramSession, err)
			os.Exit(1)
		}
		slog.Info("Using session from environment variable")
		storage = envStorage
	}
	if storage == nil {
		managed, err := sessionstore.Open(opts.Provider, opts.Profile)
		if err != nil {
			slog.Warn("Failed to open configured session provider; using memory", "error", err)
			storage = telegram.NewMemoryStorage(nil)
		} else if managed.Persistent() {
			if pending, pendingErr := config.ConsumePendingSession(); pendingErr != nil {
				slog.Warn("Failed to consume pending Telegram session", "error", pendingErr)
			} else if len(pending) > 0 {
				if storeErr := managed.StoreSession(context.Background(), pending); storeErr != nil {
					slog.Warn("Failed to migrate pending session", "error", storeErr)
				} else {
					slog.Info("Migrated one-time session handoff", "provider", managed.Provider(), "profile", managed.Profile())
				}
			}
			storage = managed
			slog.Info("Using persistent Telegram session", "provider", managed.Provider(), "profile", managed.Profile())
		} else {
			pending, pendingErr := config.ConsumePendingSession()
			if pendingErr != nil {
				slog.Warn("Failed to consume pending Telegram session", "error", pendingErr)
			} else if len(pending) > 0 {
				if storeErr := managed.StoreSession(context.Background(), pending); storeErr != nil {
					slog.Warn("Failed to import pending Telegram session", "error", storeErr)
				}
			}
			storage = managed
			slog.Info("Using volatile Telegram session", "provider", managed.Provider(), "profile", managed.Profile())
		}
	}
	tgClient = tgClient.WithSessionStorage(storage)

	return tgClient.WithUpdateStore(telegram.NewUpdateStore(1000))
}

func sessionFlagValue(cmd *cobra.Command, name string) string {
	if cmd == nil {
		return ""
	}
	if value, err := cmd.Flags().GetString(name); err == nil {
		return value
	}
	if root := cmd.Root(); root != nil {
		value, _ := root.PersistentFlags().GetString(name)
		return value
	}
	return ""
}

func firstConfigured(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

func boolFromEnv(envKey string, fallback bool) bool {
	value := os.Getenv(envKey)
	if value == "" {
		return fallback
	}
	switch value {
	case "1", "true", "TRUE", "yes", "YES", "on", "ON":
		return true
	case "0", "false", "FALSE", "no", "NO", "off", "OFF":
		return false
	default:
		return fallback
	}
}

// startTelegramClient starts the Telegram client in background with retry logic.
// It also handles session reload requests.
func startTelegramClient(ctx context.Context, tgClient *telegram.Client) {
	const maxRetries = 5

	go func() {
		for {
			// Create a cancellable context for this client session
			clientCtx, clientCancel := context.WithCancel(ctx)
			tgClient.SetCancelFn(clientCancel)

			// Start client with retry logic
			startWithRetry(clientCtx, tgClient, maxRetries)

			// Client exited - check why
			// First check if main context is done (shutdown)
			select {
			case <-ctx.Done():
				clientCancel()
				return
			default:
			}

			// Check if reload was requested (non-blocking)
			select {
			case <-tgClient.ReloadCh():
				slog.Info("Reloading session...")
				clientCancel()
				// Small delay to ensure clean disconnect
				time.Sleep(500 * time.Millisecond)
				// Loop continues, will start new client
				continue
			default:
				// No reload requested, client crashed - exit
				clientCancel()
				return
			}
		}
	}()
}

// startWithRetry attempts to start the Telegram client with retries.
func startWithRetry(ctx context.Context, tgClient *telegram.Client, maxRetries int) {
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
}

// waitForTelegramReady waits for the Telegram client to be fully initialized.
func waitForTelegramReady(ctx context.Context, tgClient *telegram.Client) {
	select {
	case <-tgClient.Ready():
		slog.Info("Telegram client ready")
	case <-ctx.Done():
		// Context cancelled, server is shutting down
	case <-time.After(60 * time.Second):
		slog.Warn("Telegram client not ready after timeout, starting IPC server anyway")
	}
}

// createIPCServer creates and configures the IPC server.
func createIPCServer(
	socketPath string,
	tgClient *telegram.Client,
	cancel context.CancelFunc,
) *ipc.SocketServer {
	srv := ipc.NewSocketServer(socketPath)
	ipc.RegisterPingPong(srv)
	policyChecker := loadPolicyChecker(tgClient)
	srv.SetPolicyChecker(policyChecker)
	telegramipc.RegisterHandlers(srv, tgClient)

	registerControlHandlers(srv, tgClient, cancel)
	return srv
}

func registerControlHandlers(srv ipc.MethodRegistrar, tgClient *telegram.Client, cancel context.CancelFunc) {
	srv.Register("status", func(parent context.Context, _ json.RawMessage) (any, *ipc.ErrorObject) {
		// Create a short-lived context for the status check
		ctx, statusCancel := context.WithTimeout(parent, 2*time.Second)
		defer statusCancel()

		tgStatus := tgClient.GetStatus(ctx)
		storageStatus := tgClient.SessionStorageStatus()

		return map[string]any{
			"status":             "running",
			"pid":                os.Getpid(),
			"session_storage":    storageStatus.Provider,
			"session_profile":    storageStatus.Profile,
			"session_persistent": storageStatus.Persistent,
			"initialized":        tgStatus.Initialized,
			"authorized":         tgStatus.Authorized,
			"telegram_state":     tgStatus.State,
			"username":           tgStatus.Username,
			"first_name":         tgStatus.FirstName,
			"user_id":            tgStatus.UserID,
		}, nil
	})

	srv.Register("shutdown", func(_ context.Context, _ json.RawMessage) (any, *ipc.ErrorObject) {
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

	srv.Register("logout", func(_ context.Context, _ json.RawMessage) (any, *ipc.ErrorObject) {
		go func() {
			time.Sleep(100 * time.Millisecond)
			logoutTelegramClient(tgClient, true)
			cancel()
		}()
		return map[string]any{
			"success": true,
			"message": "Logging out and shutting down...",
		}, nil
	})

	srv.Register("reload_session", func(_ context.Context, params json.RawMessage) (any, *ipc.ErrorObject) {
		slog.Info("Reload session requested via IPC")
		sessionData, err := parseReloadSessionData(params)
		if err != nil {
			return nil, ipc.NewTypedError(ipc.ErrCodeInvalidParams, ipc.ErrorTypeValidation, err.Error(), nil)
		}
		if err := importSessionForMemoryStorage(tgClient, sessionData); err != nil {
			return nil, ipc.NewTypedError(ipc.ErrCodeInternalError, ipc.ErrorTypeInternal, err.Error(), nil)
		}
		tgClient.Reload()
		return map[string]any{
			"success": true,
			"message": "Session reload initiated",
		}, nil
	})

}

func parseReloadSessionData(params json.RawMessage) ([]byte, error) {
	var body struct {
		Session  string `json:"session"`
		Provider string `json:"provider"`
		Profile  string `json:"profile"`
	}
	if err := json.Unmarshal(params, &body); err != nil {
		return nil, fmt.Errorf("invalid reload_session payload: %w", err)
	}
	if body.Session != "" {
		data, err := base64.StdEncoding.DecodeString(body.Session)
		if err != nil {
			return nil, fmt.Errorf("session must be base64: %w", err)
		}
		if len(data) == 0 {
			return nil, fmt.Errorf("session is empty")
		}
		return data, nil
	}
	storage, err := sessionstore.Open(body.Provider, body.Profile)
	if err != nil {
		return nil, err
	}
	data, err := storage.LoadSession(context.Background())
	if err != nil {
		return nil, fmt.Errorf("load session from %s/%s: %w", storage.Provider(), storage.Profile(), err)
	}
	return data, nil
}

func importSessionForMemoryStorage(tgClient *telegram.Client, sessionData []byte) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	imported, err := tgClient.ImportSession(ctx, sessionData)
	if err != nil {
		return err
	}
	if !imported {
		return fmt.Errorf("server session storage is not writable")
	}
	return nil
}

func loadPolicyChecker(resolver policy.PeerResolver) ipc.PolicyChecker {
	path, err := policy.DefaultPath()
	if err != nil {
		slog.Warn("failed to resolve local policy path, using defaults", "error", err)
		return policy.NewEnforcer(policy.Default(), resolver)
	}
	return policy.NewReloadingEnforcer(path, resolver)
}

func logoutTelegramClient(tgClient *telegram.Client, enabled bool) {
	if !enabled || tgClient == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := tgClient.Logout(ctx); err != nil {
		slog.Warn("Telegram logout failed", "error", err)
		return
	}
	slog.Info("Telegram session logged out")
}

// isDaemonChild checks if this process is a daemon child.
func isDaemonChild() bool {
	return os.Getenv("AGENT_TELEGRAM_DAEMON") == "1"
}

// daemonize forks the process to run in background.
func daemonize(socketPath string) error {
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
	logPath, err := paths.LogFilePathForSocket(socketPath)
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
