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

	"github.com/joho/godotenv"
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
	authAppID        string
	authAppHash      string
	authPhone        string
	authStateDir     string
	authStateTTL     time.Duration
	authStateID      string
	authCodeStdin    bool
	authPassStdin    bool
	authReload       bool
	authStatusPhone  string
	authWebMock      bool
	authWebMockSaved bool
)

var newAuthBackend = func(cfg *config.Config) authflow.Backend {
	return authflow.NewTelegramBackend(cfg, silentLogger())
}

type authRuntimeConfig struct {
	AppID        string
	AppHash      string
	Phone        string
	StateDir     string
	StateTTL     time.Duration
	StateID      string
	CodeStdin    bool
	PassStdin    bool
	Reload       bool
	StatusPhone  string
	WebQR        bool
	WebPort      int
	WebMock      bool
	WebMockSaved bool
}

func authRuntimeFromGlobals() authRuntimeConfig {
	return authRuntimeConfig{
		AppID:        authAppID,
		AppHash:      authAppHash,
		Phone:        authPhone,
		StateDir:     authStateDir,
		StateTTL:     authStateTTL,
		StateID:      authStateID,
		CodeStdin:    authCodeStdin,
		PassStdin:    authPassStdin,
		Reload:       authReload,
		StatusPhone:  authStatusPhone,
		WebQR:        authWebQR,
		WebPort:      authWebPort,
		WebMock:      authWebMock,
		WebMockSaved: authWebMockSaved,
	}
}

func (r authRuntimeConfig) stateStore() *authflow.StateStore {
	return authflow.NewStateStore(r.StateDir)
}

func (r authRuntimeConfig) authConfig(phone string) (*config.Config, error) {
	appIDStr := firstNonEmpty(r.AppID, os.Getenv("TELEGRAM_APP_ID"), os.Getenv("AGENT_TELEGRAM_APP_ID"), defaultAppID)
	appHash := firstNonEmpty(
		r.AppHash,
		os.Getenv("TELEGRAM_APP_HASH"),
		os.Getenv("AGENT_TELEGRAM_APP_HASH"),
		defaultAppHash,
	)
	phone = firstNonEmpty(phone, os.Getenv("AGENT_TELEGRAM_PHONE"))

	appID, err := strconv.Atoi(appIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid app-id: %w", err)
	}
	return config.LoadFromArgs(appID, appHash, phone, filepath.Join(defaultConfigDir(), ".agent-telegram")), nil
}

// AuthCmd groups headless authentication commands.
var AuthCmd = &cobra.Command{
	GroupID: "auth",
	Use:     "auth",
	Short:   "Headless Telegram authentication",
	Long: `Headless Telegram authentication commands for agentic workflows.

Use web for browser-based local login, web --qr for Telegram QR login, or
begin to send a Telegram login code, verify to submit the code from stdin, and
password to submit a 2FA password from stdin when required. Commands emit JSON
on stdout.`,
}

// AuthBeginCmd starts an auth flow by sending a verification code.
var AuthBeginCmd = &cobra.Command{
	Use:   "begin",
	Short: "Send Telegram login code",
	Run:   runAuthBegin,
}

// AuthVerifyCmd verifies a login code read from stdin.
var AuthVerifyCmd = &cobra.Command{
	Use:   "verify",
	Short: "Verify Telegram login code from stdin",
	Run:   runAuthVerify,
}

// AuthPasswordCmd verifies a 2FA password read from stdin.
var AuthPasswordCmd = &cobra.Command{
	Use:   "password",
	Short: "Verify Telegram 2FA password from stdin",
	Run:   runAuthPassword,
}

// AuthStatusCmd reports local auth state.
var AuthStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show local authentication status",
	Run:   runAuthStatus,
}

// AddAuthCommand adds the nested auth command to the root command.
func AddAuthCommand(rootCmd *cobra.Command) {
	rootCmd.AddCommand(AuthCmd)

	AuthCmd.AddCommand(AuthBeginCmd, AuthVerifyCmd, AuthPasswordCmd, AuthWebCmd, AuthStatusCmd)

	addBeginFlags(AuthBeginCmd)
	addBeginFlags(AuthWebCmd)
	addStateFlags(AuthVerifyCmd)
	addStateFlags(AuthPasswordCmd)

	AuthVerifyCmd.Flags().BoolVar(&authCodeStdin, "code-stdin", false, "Read Telegram login code from stdin")
	AuthPasswordCmd.Flags().BoolVar(&authPassStdin, "password-stdin", false, "Read Telegram 2FA password from stdin")
	AuthVerifyCmd.Flags().BoolVar(&authReload, "reload-server", true, "Reload running IPC server after successful login")
	AuthPasswordCmd.Flags().BoolVar(&authReload, "reload-server", true, "Reload running IPC server after successful login")
	AuthWebCmd.Flags().BoolVar(&authReload, "reload-server", true, "Reload running IPC server after successful login")
	AuthWebCmd.Flags().BoolVar(&authWebQR, "qr", true, "Use QR code authentication flow")
	AuthWebCmd.Flags().IntVar(&authWebPort, "port", 0, "Local web auth port (0 chooses a free port)")
	AuthWebCmd.Flags().BoolVar(&authWebMock, "mock", false, "Use mock web auth data without Telegram")
	AuthWebCmd.Flags().BoolVar(&authWebMockSaved, "mock-saved-session", false, "Expose a saved session in mock web auth")
}

func addBeginFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&authAppID, "app-id", os.Getenv("TELEGRAM_APP_ID"), "Telegram API App ID")
	cmd.Flags().StringVar(&authAppHash, "app-hash", os.Getenv("TELEGRAM_APP_HASH"), "Telegram API App Hash")
	cmd.Flags().StringVar(&authStateDir, "state-dir", authflow.DefaultStateDir(), "Directory for temporary auth state")
	cmd.Flags().DurationVar(&authStateTTL, "state-ttl", authflow.DefaultStateTTL, "Temporary auth state lifetime")
}

func addStateFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&authStateID, "state-id", "", "Auth state ID returned by auth begin")
	cmd.Flags().StringVar(&authStateDir, "state-dir", authflow.DefaultStateDir(), "Directory for temporary auth state")
}

func runAuthBegin(_ *cobra.Command, _ []string) {
	runAuthBeginWithRuntime(authRuntimeFromGlobals())
}

func runAuthBeginWithRuntime(runtime authRuntimeConfig) {
	_ = godotenv.Load()

	cfg, err := runtime.authConfig(runtime.Phone)
	if err != nil {
		failJSON(err.Error())
	}
	if cfg.Phone == "" {
		failJSON("phone is required")
	}

	backend := newAuthBackend(cfg)
	ctx := context.Background()
	result, err := backend.SendCode(ctx, cfg.Phone)
	if err != nil {
		failJSON(fmt.Sprintf("failed to send code: %v", err))
	}
	sessionData, err := backend.ExportSession(ctx)
	if err != nil {
		failJSON(fmt.Sprintf("failed to export auth session: %v", err))
	}

	store := runtime.stateStore()
	state, err := store.Create(cfg.Phone, result.PhoneCodeHash, cfg.AppID, cfg.AppHash, sessionData, runtime.StateTTL)
	if err != nil {
		failJSON(err.Error())
	}

	writeJSON(map[string]any{
		"ok":        true,
		"next":      "code",
		"stateId":   state.ID,
		"phone":     maskPhone(state.Phone),
		"expiresAt": state.ExpiresAt,
		"timeout":   result.Timeout,
	})
}

func runAuthVerify(cmd *cobra.Command, _ []string) {
	runAuthVerifyWithRuntime(cmd, authRuntimeFromGlobals())
}

func runAuthVerifyWithRuntime(cmd *cobra.Command, runtime authRuntimeConfig) {
	state := loadStateOrExit(runtime)
	if !runtime.CodeStdin {
		failJSON("--code-stdin is required")
	}
	code, err := readSecretFromStdin("code", trimAllSpace)
	if err != nil {
		failJSON(err.Error())
	}
	if code == "" {
		failJSON("code is empty")
	}

	cfg := config.LoadFromArgs(
		state.AppID,
		state.AppHash,
		state.Phone,
		filepath.Join(defaultConfigDir(), ".agent-telegram"),
	)
	backend := newAuthBackend(cfg)
	ctx := context.Background()
	if err := importStateSession(ctx, backend, state); err != nil {
		failJSON(err.Error())
	}
	result, err := backend.SignIn(ctx, state.Phone, code, state.PhoneCodeHash)
	if err != nil {
		failJSON(fmt.Sprintf("sign in failed: %v", err))
	}
	if result.Requires2FA {
		state.Requires2FA = true
		state.TwoFactorHint = result.TwoFactorHint
		if err := persistBackendSession(ctx, backend, state, runtime.stateStore()); err != nil {
			failJSON(err.Error())
		}
		writeJSON(map[string]any{
			"ok":          true,
			"next":        "password",
			"stateId":     state.ID,
			"phone":       maskPhone(state.Phone),
			"requires2FA": true,
			"hint":        result.TwoFactorHint,
			"expiresAt":   state.ExpiresAt,
		})
		return
	}
	if !result.Success {
		failJSON(resultError(result.AuthError, "authentication failed"))
	}
	if err := persistBackendSession(ctx, backend, state, runtime.stateStore()); err != nil {
		failJSON(err.Error())
	}
	completeAuth(cmd, runtime, state)
}

func runAuthPassword(cmd *cobra.Command, _ []string) {
	runAuthPasswordWithRuntime(cmd, authRuntimeFromGlobals())
}

func runAuthPasswordWithRuntime(cmd *cobra.Command, runtime authRuntimeConfig) {
	state := loadStateOrExit(runtime)
	if !runtime.PassStdin {
		failJSON("--password-stdin is required")
	}
	password, err := readSecretFromStdin("password", trimLineEndings)
	if err != nil {
		failJSON(err.Error())
	}
	if password == "" {
		failJSON("password is empty")
	}

	cfg := config.LoadFromArgs(
		state.AppID,
		state.AppHash,
		state.Phone,
		filepath.Join(defaultConfigDir(), ".agent-telegram"),
	)
	backend := newAuthBackend(cfg)
	ctx := context.Background()
	if err := importStateSession(ctx, backend, state); err != nil {
		failJSON(err.Error())
	}
	result, err := backend.SignInWith2FA(ctx, state.Phone, password)
	if err != nil {
		failJSON(fmt.Sprintf("2FA sign in failed: %v", err))
	}
	if !result.Success {
		failJSON(resultError(result.AuthError, "2FA authentication failed"))
	}
	if err := persistBackendSession(ctx, backend, state, runtime.stateStore()); err != nil {
		failJSON(err.Error())
	}
	completeAuth(cmd, runtime, state)
}

func runAuthStatus(cmd *cobra.Command, _ []string) {
	runAuthStatusWithRuntime(cmd, authRuntimeFromGlobals())
}

func runAuthStatusWithRuntime(cmd *cobra.Command, runtime authRuntimeConfig) {
	configPath, configErr := config.ConfigPath()
	configInfo, cfgErr := os.Stat(configPath)
	serverAuthorized := false
	storageProvider := sessionstore.DefaultProvider()
	storageProfile := sessionstore.DefaultProfile
	storagePersistent := false
	provider, profile := authSessionSelection(cmd)
	if storage, err := sessionstore.Open(provider, profile); err == nil {
		selection := storage.Selection()
		storageProvider = selection.Provider
		storageProfile = selection.Profile
		storagePersistent = selection.Persistent
	}
	if status, err := ipc.NewClient(paths.DefaultSocketPath).Status(); err == nil {
		if authorized, ok := status["authorized"].(bool); ok {
			serverAuthorized = authorized
		}
		if provider, ok := status["session_storage"].(string); ok {
			storageProvider = provider
		}
		if profile, ok := status["session_profile"].(string); ok {
			storageProfile = profile
		}
		if persistent, ok := status["session_persistent"].(bool); ok {
			storagePersistent = persistent
		}
	}

	writeJSON(map[string]any{
		"ok":                true,
		"authenticated":     serverAuthorized,
		"sessionStorage":    storageProvider,
		"sessionProfile":    storageProfile,
		"sessionPersistent": storagePersistent,
		"configPath":        configPath,
		"configExists":      configErr == nil && cfgErr == nil && !configInfo.IsDir(),
		"phone":             maskPhone(firstNonEmpty(runtime.StatusPhone, os.Getenv("AGENT_TELEGRAM_PHONE"))),
	})
}

func completeAuth(cmd *cobra.Command, runtime authRuntimeConfig, state *authflow.State) {
	body, err := finishAuth(cmd, runtime, state)
	if err != nil {
		failJSON(err.Error())
	}
	writeJSON(body)
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

func loadStateOrExit(runtime authRuntimeConfig) *authflow.State {
	if runtime.StateID == "" {
		failJSON("--state-id is required")
	}
	state, err := runtime.stateStore().Load(runtime.StateID)
	if err != nil {
		failJSON(err.Error())
	}
	return state
}

func readSecret(r io.Reader, trim func(string) string) (string, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return "", fmt.Errorf("read stdin: %w", err)
	}
	return trim(string(data)), nil
}

func readSecretFromStdin(name string, trim func(string) string) (string, error) {
	if info, err := os.Stdin.Stat(); err == nil && info.Mode()&os.ModeCharDevice != 0 {
		return "", fmt.Errorf("%s must be piped on stdin", name)
	}
	return readSecret(os.Stdin, trim)
}

func trimAllSpace(value string) string {
	return strings.TrimSpace(value)
}

func trimLineEndings(value string) string {
	return strings.TrimRight(value, "\r\n")
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

func resultError(actual, fallback string) string {
	if actual != "" {
		return actual
	}
	return fallback
}

func importStateSession(ctx context.Context, backend authflow.Backend, state *authflow.State) error {
	sessionData, err := state.SessionData()
	if err != nil {
		return err
	}
	return backend.ImportSession(ctx, sessionData)
}

func persistBackendSession(
	ctx context.Context,
	backend authflow.Backend,
	state *authflow.State,
	store *authflow.StateStore,
) error {
	sessionData, err := backend.ExportSession(ctx)
	if err != nil {
		return fmt.Errorf("export auth session: %w", err)
	}
	state.SetSessionData(sessionData)
	return store.Save(state)
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
