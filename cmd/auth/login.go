// Package auth provides authentication commands.
package auth

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"agent-telegram/internal/authflow"
	"agent-telegram/internal/cliutil"
	"agent-telegram/internal/config"
	"agent-telegram/internal/ipc"
	"agent-telegram/internal/paths"
	"agent-telegram/internal/sessionstore"
)

// Default Telegram API credentials.
const (
	defaultAppID   = "35699202"
	defaultAppHash = "7e97f16795114cf3046d1aebf9de886d"
)

var (
	authAppID    string
	authAppHash  string
	authStateDir string
	authStateTTL time.Duration
	authReload   bool
	authWebMock  bool
)

var newAuthBackend = func(cfg *config.Config) authflow.Backend {
	return authflow.NewTelegramBackend(cfg, silentLogger())
}

type authRuntimeConfig struct {
	AppID    string
	AppHash  string
	StateDir string
	StateTTL time.Duration
	Reload   bool
	WebPort  int
	WebMock  bool
}

func authRuntimeFromGlobals() authRuntimeConfig {
	return authRuntimeConfig{
		AppID:    authAppID,
		AppHash:  authAppHash,
		StateDir: authStateDir,
		StateTTL: authStateTTL,
		Reload:   authReload,
		WebPort:  authWebPort,
		WebMock:  authWebMock,
	}
}

func (r authRuntimeConfig) stateStore() *authflow.StateStore {
	return authflow.NewStateStore(r.StateDir)
}

func (r authRuntimeConfig) authConfig() (*config.Config, error) {
	appIDStr := firstNonEmpty(r.AppID, os.Getenv("TELEGRAM_APP_ID"), os.Getenv("AGENT_TELEGRAM_APP_ID"), defaultAppID)
	appHash := firstNonEmpty(
		r.AppHash,
		os.Getenv("TELEGRAM_APP_HASH"),
		os.Getenv("AGENT_TELEGRAM_APP_HASH"),
		defaultAppHash,
	)
	appID, err := strconv.Atoi(appIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid app-id: %w", err)
	}
	return config.LoadFromArgs(appID, appHash, "", filepath.Join(defaultConfigDir(), ".agent-telegram")), nil
}

// AuthCmd starts the browser-based authentication flow.
var AuthCmd = &cobra.Command{
	GroupID: "auth",
	Use:     "auth",
	Short:   "Login through a local browser page",
	Long: `Start a local browser-based Telegram login flow.

QR login is the only supported sign-in method. The page is printed to stderr
as a one-time localhost URL, then the command waits for completion and emits
JSON on stdout.`,
	Args: cobra.NoArgs,
	Run:  runAuthWeb,
}

// AddAuthCommand adds the auth command to the root command.
func AddAuthCommand(rootCmd *cobra.Command) {
	rootCmd.AddCommand(AuthCmd)

	addAuthBaseFlags(AuthCmd)
	addWebAuthFlags(AuthCmd)
}

func addAuthBaseFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&authAppID, "app-id", os.Getenv("TELEGRAM_APP_ID"), "Telegram API App ID")
	cmd.Flags().StringVar(&authAppHash, "app-hash", os.Getenv("TELEGRAM_APP_HASH"), "Telegram API App Hash")
	cmd.Flags().StringVar(&authStateDir, "state-dir", authflow.DefaultStateDir(), "Directory for temporary auth state")
	cmd.Flags().DurationVar(&authStateTTL, "state-ttl", authflow.DefaultStateTTL, "Temporary auth state lifetime")
}

func addWebAuthFlags(cmd *cobra.Command) {
	cmd.Flags().BoolVar(&authReload, "reload-server", true, "Reload running IPC server after successful login")
	cmd.Flags().IntVar(&authWebPort, "port", 0, "Local web auth port (0 chooses a free port)")
	cmd.Flags().BoolVar(&authWebMock, "mock", false, "Use mock web auth data without Telegram")
}

func finishAuth(cmd *cobra.Command, runtime authRuntimeConfig, state *authflow.State) (map[string]any, error) {
	sessionData, err := state.SessionData()
	if err != nil {
		return nil, err
	}
	provider, profile := authSessionSelection(cmd)
	storage, err := sessionstore.Open(provider, profile)
	if err != nil {
		return nil, fmt.Errorf("open session storage: %w", err)
	}
	selection := storage.Selection()
	if selection.Persistent {
		if err := storage.StoreSession(context.Background(), sessionData); err != nil {
			return nil, fmt.Errorf("save Telegram session to %s: %w", selection.Provider, err)
		}
	}
	if err := config.SaveConfigForSession(state.AppID, state.AppHash, selection.Provider, selection.Profile); err != nil {
		return nil, fmt.Errorf("failed to save config: %w", err)
	}
	serverReloaded := false
	sessionPending := false
	if runtime.Reload {
		socketPath, _ := cmd.Flags().GetString("socket")
		if socketPath == "" {
			socketPath, _ = cmd.Root().PersistentFlags().GetString("socket")
		}
		serverReloaded = reloadServerSession(socketPath, sessionData, selection)
		if !serverReloaded && !selection.Persistent {
			if err := config.SavePendingSession(sessionData); err != nil {
				return nil, fmt.Errorf("save session for daemon startup: %w", err)
			}
			sessionPending = true
		}
	}
	if err := runtime.stateStore().Delete(state.ID); err != nil {
		return nil, err
	}
	return map[string]any{
		"ok":                true,
		"next":              "done",
		"phone":             maskPhone(state.Phone),
		"sessionStorage":    selection.Provider,
		"sessionProfile":    selection.Profile,
		"sessionPersistent": selection.Persistent,
		"serverReloaded":    serverReloaded,
		"sessionPending":    sessionPending,
	}, nil
}

func authSessionSelection(cmd *cobra.Command) (string, string) {
	provider := commandStringFlag(cmd, "session-provider")
	profile := commandStringFlag(cmd, "profile")
	if provider != "" && profile != "" {
		return provider, profile
	}
	if stored, err := config.LoadStoredConfig(); err == nil {
		if provider == "" {
			provider = stored.SessionProvider
		}
		if profile == "" {
			profile = stored.SessionProfile
		}
	}
	return provider, profile
}

func commandStringFlag(cmd *cobra.Command, name string) string {
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

func trimAllSpace(value string) string {
	return strings.TrimSpace(value)
}

func writeJSON(value any) {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(value); err != nil {
		fmt.Fprintf(os.Stderr, "failed to write JSON: %v\n", err)
		cliutil.Exit(1)
	}
}

func failJSON(message string) {
	writeJSON(map[string]any{
		"ok":    false,
		"error": message,
	})
	cliutil.Exit(1)
}

func reloadServerSession(socketPath string, sessionData []byte, selection sessionstore.Selection) bool {
	if len(sessionData) == 0 {
		fmt.Fprintln(os.Stderr, "warning: no in-memory session data to reload")
		return false
	}
	if socketPath == "" {
		socketPath = paths.DefaultSocketPath
	}
	client := ipc.NewClient(socketPath)
	if _, err := client.Call("status", nil); err != nil {
		return false
	}
	payload := map[string]any{}
	if selection.Persistent {
		payload["provider"] = selection.Provider
		payload["profile"] = selection.Profile
	} else {
		payload["session"] = base64.StdEncoding.EncodeToString(sessionData)
	}
	if _, err := client.Call("reload_session", payload); err != nil {
		fmt.Fprintf(os.Stderr, "warning: failed to reload server session: %v\n", err)
		return false
	}
	time.Sleep(2 * time.Second)
	if _, err := client.Call("status", nil); err != nil {
		fmt.Fprintln(os.Stderr, "warning: server may still be reloading")
	}
	return true
}

func silentLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func defaultConfigDir() string {
	if home := os.Getenv("HOME"); home != "" {
		return home
	}
	return "."
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

func maskPhone(phone string) string {
	if phone == "" {
		return ""
	}
	digits := make([]rune, 0, len(phone))
	for _, r := range phone {
		if r >= '0' && r <= '9' {
			digits = append(digits, r)
		}
	}
	if len(digits) <= 4 {
		return "***"
	}
	return "***" + string(digits[len(digits)-4:])
}
