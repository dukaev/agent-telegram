package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"agent-telegram/internal/config"
	"agent-telegram/internal/ipc"
	telegramipc "agent-telegram/internal/telegram/ipc"
	"agent-telegram/telegram"
)

var (
	serveAPIPort    int
	serveAPISecret  string
	serveAPISession string
	serveAPICORS    string
	serveAPIUnsafe  bool
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
	serveAPICmd.Flags().StringVar(&serveAPISession, "session", "", "Path to Telegram session file")
	serveAPICmd.Flags().StringVar(&serveAPICORS, "cors", "", "CORS allowed origins (comma-separated, empty disables CORS)")
	serveAPICmd.Flags().BoolVar(&serveAPIUnsafe, "unsafe-no-auth", false, "Allow HTTP API without auth")
}

func runServeAPI(_ *cobra.Command, _ []string) {
	storedCfg, err := config.LoadStoredConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	setupLogger()
	ctx, cancel := setupContext()

	secret := serveAPISecret
	if secret == "" {
		secret = os.Getenv("AGENT_TELEGRAM_API_SECRET")
	}
	if secret == "" && !serveAPIUnsafe {
		fmt.Fprintln(os.Stderr, "Error: --secret or AGENT_TELEGRAM_API_SECRET is required")
		fmt.Fprintln(os.Stderr, "Use --unsafe-no-auth only for trusted local development.")
		os.Exit(1)
	}

	sessionPath := serveAPISession
	if sessionPath == "" {
		sessionPath = getSessionPath()
	}

	tgClient := createTelegramClient(storedCfg.AppID, storedCfg.AppHash, sessionPath)
	startTelegramClient(ctx, tgClient)
	waitForTelegramReady(ctx, tgClient)

	srv := createHTTPAPIServer(serveAPIPort, secret, serveAPICORS, tgClient, cancel)
	fmt.Fprintf(os.Stderr, "REST API on http://localhost:%d\n", serveAPIPort)

	if err := srv.Start(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "HTTP server error: %v\n", err)
		cancel()
		os.Exit(1)
	}

	cancel()
	fmt.Fprintln(os.Stderr, "serve-api stopped.")
}

func createHTTPAPIServer(
	port int, secret, cors string,
	tgClient *telegram.Client, cancel context.CancelFunc,
) *ipc.HTTPServer {
	srv := ipc.NewHTTPServer(port, secret, cors)
	ipc.RegisterPingPong(srv)
	telegramipc.RegisterHandlers(srv, tgClient)

	srv.Register("status", func(_ json.RawMessage) (any, *ipc.ErrorObject) {
		ctx, statusCancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer statusCancel()

		tgStatus := tgClient.GetStatus(ctx)
		sessionPath, _ := tgClient.GetSessionPath()

		return map[string]any{
			"status":         "running",
			"pid":            os.Getpid(),
			"session_path":   sessionPath,
			"initialized":    tgStatus.Initialized,
			"authorized":     tgStatus.Authorized,
			"telegram_state": tgStatus.State,
			"username":       tgStatus.Username,
			"first_name":     tgStatus.FirstName,
			"user_id":        tgStatus.UserID,
		}, nil
	})

	srv.Register("shutdown", func(_ json.RawMessage) (any, *ipc.ErrorObject) {
		go func() {
			time.Sleep(100 * time.Millisecond)
			cancel()
		}()
		return map[string]any{
			"success": true,
			"message": "Shutting down...",
		}, nil
	})

	srv.Register("reload_session", func(_ json.RawMessage) (any, *ipc.ErrorObject) {
		tgClient.Reload()
		return map[string]any{
			"success": true,
			"message": "Session reload initiated",
		}, nil
	})

	return srv
}
