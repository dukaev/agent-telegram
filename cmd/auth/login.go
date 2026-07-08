// Package auth provides authentication commands.
package auth

import (
	"context"
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
)

// Default Telegram API credentials.
const (
	defaultAppID   = "35699202"
	defaultAppHash = "7e97f16795114cf3046d1aebf9de886d"
)

var (
	authAppID       string
	authAppHash     string
	authPhone       string
	authStateDir    string
	authStateTTL    time.Duration
	authStateID     string
	authCodeStdin   bool
	authPassStdin   bool
	authReload      bool
	authStatusPhone string
)

var newAuthBackend = func(cfg *config.Config) authflow.Backend {
	return authflow.NewTelegramBackend(cfg, silentLogger())
}

// AuthCmd groups headless authentication commands.
var AuthCmd = &cobra.Command{
	GroupID: "auth",
	Use:     "auth",
	Short:   "Headless Telegram authentication",
	Long: `Headless Telegram authentication commands for agentic workflows.

Use web for browser-based local login, or begin to send a Telegram login code,
verify to submit the code from stdin, and password to submit a 2FA password
from stdin when required. Commands emit JSON on stdout.`,
}

// LoginCmd is a backwards-compatible alias for auth begin.
var LoginCmd = &cobra.Command{
	GroupID: "auth",
	Use:     "login",
	Short:   "Start headless Telegram login",
	Long: `Start a headless Telegram login by sending a verification code.

This command replaces the old TUI login. It emits JSON with a stateId, then use:
  agent-telegram auth verify --state-id <id> --code-stdin

Set AGENT_TELEGRAM_PHONE instead of passing a phone number in argv.`,
	Run: runAuthBegin,
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
	addBeginFlags(LoginCmd)
	addStateFlags(AuthVerifyCmd)
	addStateFlags(AuthPasswordCmd)

	AuthVerifyCmd.Flags().BoolVar(&authCodeStdin, "code-stdin", false, "Read Telegram login code from stdin")
	AuthPasswordCmd.Flags().BoolVar(&authPassStdin, "password-stdin", false, "Read Telegram 2FA password from stdin")
	AuthVerifyCmd.Flags().BoolVar(&authReload, "reload-server", true, "Reload running IPC server after successful login")
	AuthPasswordCmd.Flags().BoolVar(&authReload, "reload-server", true, "Reload running IPC server after successful login")
	AuthWebCmd.Flags().BoolVar(&authReload, "reload-server", true, "Reload running IPC server after successful login")
	AuthWebCmd.Flags().BoolVar(&authWebQR, "qr", false, "Use QR code authentication flow")
	AuthWebCmd.Flags().IntVar(&authWebPort, "port", 0, "Local web auth port (0 chooses a free port)")
}

// AddLoginCommand adds the legacy login alias to the root command.
func AddLoginCommand(rootCmd *cobra.Command) {
	rootCmd.AddCommand(LoginCmd)
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
	_ = godotenv.Load()

	cfg, err := authConfig(authPhone)
	if err != nil {
		failJSON(err.Error())
	}
	if cfg.Phone == "" {
		failJSON("phone is required")
	}

	backend := newAuthBackend(cfg)
	result, err := backend.SendCode(context.Background(), cfg.Phone)
	if err != nil {
		failJSON(fmt.Sprintf("failed to send code: %v", err))
	}

	store := authflow.NewStateStore(authStateDir)
	state, err := store.Create(cfg.Phone, result.PhoneCodeHash, cfg.AppID, cfg.AppHash, backend.SessionPath(), authStateTTL)
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
	state := loadStateOrExit()
	if !authCodeStdin {
		failJSON("--code-stdin is required")
	}
	code, err := readSecretFromStdin("code", trimAllSpace)
	if err != nil {
		failJSON(err.Error())
	}
	if code == "" {
		failJSON("code is empty")
	}

	cfg := config.LoadFromArgs(state.AppID, state.AppHash, state.Phone, sessionDirFromPath(state.SessionPath))
	backend := newAuthBackend(cfg)
	result, err := backend.SignIn(context.Background(), state.Phone, code, state.PhoneCodeHash)
	if err != nil {
		failJSON(fmt.Sprintf("sign in failed: %v", err))
	}
	if result.Requires2FA {
		state.Requires2FA = true
		state.TwoFactorHint = result.TwoFactorHint
		if err := authflow.NewStateStore(authStateDir).Save(state); err != nil {
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
	completeAuth(cmd, state)
}

func runAuthPassword(cmd *cobra.Command, _ []string) {
	state := loadStateOrExit()
	if !authPassStdin {
		failJSON("--password-stdin is required")
	}
	password, err := readSecretFromStdin("password", trimLineEndings)
	if err != nil {
		failJSON(err.Error())
	}
	if password == "" {
		failJSON("password is empty")
	}

	cfg := config.LoadFromArgs(state.AppID, state.AppHash, state.Phone, sessionDirFromPath(state.SessionPath))
	backend := newAuthBackend(cfg)
	result, err := backend.SignInWith2FA(context.Background(), state.Phone, password)
	if err != nil {
		failJSON(fmt.Sprintf("2FA sign in failed: %v", err))
	}
	if !result.Success {
		failJSON(resultError(result.AuthError, "2FA authentication failed"))
	}
	completeAuth(cmd, state)
}

func runAuthStatus(_ *cobra.Command, _ []string) {
	sessionPath := filepath.Join(defaultConfigDir(), ".agent-telegram", "session.json")
	configPath, configErr := config.ConfigPath()
	sessionInfo, sessionErr := os.Stat(sessionPath)
	configInfo, cfgErr := os.Stat(configPath)

	writeJSON(map[string]any{
		"ok":            true,
		"authenticated": sessionErr == nil && !sessionInfo.IsDir(),
		"sessionPath":   sessionPath,
		"sessionExists": sessionErr == nil && !sessionInfo.IsDir(),
		"configPath":    configPath,
		"configExists":  configErr == nil && cfgErr == nil && !configInfo.IsDir(),
		"phone":         maskPhone(firstNonEmpty(authStatusPhone, os.Getenv("AGENT_TELEGRAM_PHONE"))),
	})
}

func completeAuth(cmd *cobra.Command, state *authflow.State) {
	body, err := finishAuth(cmd, state)
	if err != nil {
		failJSON(err.Error())
	}
	writeJSON(body)
}

func finishAuth(cmd *cobra.Command, state *authflow.State) (map[string]any, error) {
	if err := config.SaveConfig(state.AppID, state.AppHash); err != nil {
		return nil, fmt.Errorf("failed to save config: %w", err)
	}
	serverReloaded := false
	if authReload {
		socketPath, _ := cmd.Flags().GetString("socket")
		if socketPath == "" {
			socketPath, _ = cmd.Root().PersistentFlags().GetString("socket")
		}
		serverReloaded = reloadServerSession(socketPath)
	}
	if err := authflow.NewStateStore(authStateDir).Delete(state.ID); err != nil {
		return nil, err
	}
	return map[string]any{
		"ok":             true,
		"next":           "done",
		"phone":          maskPhone(state.Phone),
		"sessionPath":    state.SessionPath,
		"serverReloaded": serverReloaded,
	}, nil
}

func loadStateOrExit() *authflow.State {
	if authStateID == "" {
		failJSON("--state-id is required")
	}
	state, err := authflow.NewStateStore(authStateDir).Load(authStateID)
	if err != nil {
		failJSON(err.Error())
	}
	return state
}

func authConfig(phone string) (*config.Config, error) {
	appIDStr := firstNonEmpty(authAppID, os.Getenv("TELEGRAM_APP_ID"), os.Getenv("AGENT_TELEGRAM_APP_ID"), defaultAppID)
	appHash := firstNonEmpty(authAppHash, os.Getenv("TELEGRAM_APP_HASH"), os.Getenv("AGENT_TELEGRAM_APP_HASH"), defaultAppHash)
	phone = firstNonEmpty(phone, os.Getenv("AGENT_TELEGRAM_PHONE"))

	appID, err := strconv.Atoi(appIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid app-id: %w", err)
	}
	return config.LoadFromArgs(appID, appHash, phone, filepath.Join(defaultConfigDir(), ".agent-telegram")), nil
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

func reloadServerSession(socketPath string) bool {
	if socketPath == "" {
		socketPath = "/tmp/agent-telegram.sock"
	}
	client := ipc.NewClient(socketPath)
	if _, err := client.Call("status", nil); err != nil {
		return false
	}
	if _, err := client.Call("reload_session", nil); err != nil {
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

func sessionDirFromPath(path string) string {
	if path == "" {
		return filepath.Join(defaultConfigDir(), ".agent-telegram")
	}
	return filepath.Dir(path)
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
