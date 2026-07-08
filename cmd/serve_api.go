package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/spf13/cobra"

	"agent-telegram/internal/config"
	"agent-telegram/internal/ipc"
	telegramipc "agent-telegram/internal/telegram/ipc"
	"agent-telegram/telegram"
)

var (
	serveAPIPort   int
	serveAPISecret string
	serveAPICORS   string
	serveAPIUnsafe bool
	serveAPILogout bool
)

var serveAPICmd = &cobra.Command{
	Use:   "serve-api",
	Short: "Start HTTP REST API server with Telegram",
	Long: `Start an HTTP server that exposes all Telegram actions as REST endpoints.

Each registered method is available at POST /rpc/{method} with a JSON body.

Examples:
  curl -X POST http://localhost:8080/rpc/send_message \
    -H "Authorization: Bearer <secret>" \
    -H "Content-Type: application/json" \
    -d '{"peer": "username", "message": "hello"}'

  curl http://localhost:8080/health
  curl http://localhost:8080/methods -H "Authorization: Bearer <secret>"`,
	GroupID: GroupIDServer,
	Run:     runServeAPI,
}

func init() {
	RootCmd.AddCommand(serveAPICmd)

	serveAPICmd.Flags().IntVar(&serveAPIPort, "port", 8080, "HTTP API server port")
	serveAPICmd.Flags().StringVar(&serveAPISecret, "secret", "",
		"Bearer token for auth (or AGENT_TELEGRAM_API_SECRET env)")
	serveAPICmd.Flags().StringVar(&serveAPICORS, "cors", "", "CORS allowed origins (comma-separated, empty disables CORS)")
	serveAPICmd.Flags().BoolVar(&serveAPIUnsafe, "unsafe-no-auth", false, "Allow HTTP API without auth")
	serveAPICmd.Flags().BoolVar(&serveAPILogout, "logout-on-stop", true, "Logout from Telegram when the server stops")
}

func runServeAPI(_ *cobra.Command, _ []string) {
	storedCfg, err := config.LoadStoredConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	setupLogger()

	secret := serveAPISecret
	if secret == "" {
		secret = os.Getenv("AGENT_TELEGRAM_API_SECRET")
	}
	if secret == "" && !serveAPIUnsafe {
		fmt.Fprintln(os.Stderr, "Error: --secret or AGENT_TELEGRAM_API_SECRET is required")
		fmt.Fprintln(os.Stderr, "Use --unsafe-no-auth only for trusted local development.")
		os.Exit(1)
	}

	tgClient := createTelegramClient(storedCfg.AppID, storedCfg.AppHash, telegramClientOptions{})
	logoutOnStop := boolFromEnv(envLogoutOnStop, serveAPILogout)
	var logoutOnce sync.Once
	logout := func() {
		logoutOnce.Do(func() {
			logoutTelegramClient(tgClient, logoutOnStop)
		})
	}
	ctx, cancel := setupContextWithShutdown(logout)
	startTelegramClient(ctx, tgClient)
	go waitForTelegramReady(ctx, tgClient)

	srv := createHTTPAPIServer(serveAPIPort, secret, serveAPICORS, tgClient, cancel, logout)
	srv.SetPolicyChecker(loadPolicyChecker(tgClient))
	fmt.Fprintf(os.Stderr, "REST API on http://localhost:%d\n", serveAPIPort)

	if err := srv.Start(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "HTTP server error: %v\n", err)
		cancel()
		os.Exit(1)
	}

	logout()
	cancel()
	fmt.Fprintln(os.Stderr, "serve-api stopped.")
}

func firstNonEmptyString(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

func createHTTPAPIServer(
	port int, secret, cors string,
	tgClient *telegram.Client, cancel context.CancelFunc,
	logout func(),
) *ipc.HTTPServer {
	srv := ipc.NewHTTPServer(port, secret, cors)
	ipc.RegisterPingPong(srv)
	telegramipc.RegisterHandlers(srv, tgClient)

	srv.Register("status", func(_ json.RawMessage) (any, *ipc.ErrorObject) {
		ctx, statusCancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer statusCancel()

		tgStatus := tgClient.GetStatus(ctx)

		return map[string]any{
			"status":          "running",
			"pid":             os.Getpid(),
			"session_storage": "memory",
			"initialized":     tgStatus.Initialized,
			"authorized":      tgStatus.Authorized,
			"telegram_state":  tgStatus.State,
			"username":        tgStatus.Username,
			"first_name":      tgStatus.FirstName,
			"user_id":         tgStatus.UserID,
		}, nil
	})

	srv.Register("shutdown", func(_ json.RawMessage) (any, *ipc.ErrorObject) {
		go func() {
			time.Sleep(100 * time.Millisecond)
			if logout != nil {
				logout()
			}
			cancel()
		}()
		return map[string]any{
			"success": true,
			"message": "Shutting down...",
		}, nil
	})

	srv.Register("reload_session", func(params json.RawMessage) (any, *ipc.ErrorObject) {
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

	return srv
}
