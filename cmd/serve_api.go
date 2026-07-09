package cmd

import (
	"context"
	"fmt"
	"net"
	"os"
	"strconv"
	"sync"

	"github.com/spf13/cobra"

	"agent-telegram/internal/config"
	"agent-telegram/internal/ipc"
	telegramipc "agent-telegram/internal/telegram/ipc"
	"agent-telegram/telegram"
)

var (
	serveAPIPort      int
	serveAPIListen    string
	serveAPISecret    string
	serveAPICORS      string
	serveAPIFileRoots []string
	serveAPIUnsafe    bool
	serveAPILogout    bool
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
	serveAPICmd.Flags().StringVar(&serveAPIListen, "listen", "127.0.0.1", "HTTP listen address")
	serveAPICmd.Flags().StringVar(&serveAPISecret, "secret", "",
		"Bearer token for auth (or AGENT_TELEGRAM_API_SECRET env)")
	serveAPICmd.Flags().StringVar(&serveAPICORS, "cors", "", "CORS allowed origins (comma-separated, empty disables CORS)")
	serveAPICmd.Flags().StringSliceVar(
		&serveAPIFileRoots,
		"file-root",
		nil,
		"Allow HTTP file parameters only below these server directories (repeatable)",
	)
	serveAPICmd.Flags().BoolVar(&serveAPIUnsafe, "unsafe-no-auth", false, "Allow HTTP API without auth")
	serveAPICmd.Flags().BoolVar(&serveAPILogout, "logout-on-stop", false, "Logout from Telegram when the server stops")
}

func runServeAPI(cmd *cobra.Command, _ []string) {
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

	tgClient := createTelegramClient(storedCfg.AppID, storedCfg.AppHash, telegramClientOptions{
		Provider: firstConfigured(sessionFlagValue(cmd, "session-provider"), storedCfg.SessionProvider),
		Profile:  firstConfigured(sessionFlagValue(cmd, "profile"), storedCfg.SessionProfile),
	})
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

	address := net.JoinHostPort(serveAPIListen, strconv.Itoa(serveAPIPort))
	srv := createHTTPAPIServer(address, secret, serveAPICORS, serveAPIFileRoots, tgClient, cancel)
	srv.SetPolicyChecker(loadPolicyChecker(tgClient))
	fmt.Fprintf(os.Stderr, "REST API on http://%s\n", address)

	if err := srv.Start(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "HTTP server error: %v\n", err)
		cancel()
		os.Exit(1)
	}

	logout()
	cancel()
	fmt.Fprintln(os.Stderr, "serve-api stopped.")
}

func createHTTPAPIServer(
	address, secret, cors string,
	fileRoots []string,
	tgClient *telegram.Client, cancel context.CancelFunc,
) *ipc.HTTPServer {
	srv := ipc.NewHTTPServerOnAddress(address, secret, cors)
	srv.SetFileRoots(fileRoots)
	ipc.RegisterPingPong(srv)
	telegramipc.RegisterHandlers(srv, tgClient)

	registerControlHandlers(srv, tgClient, cancel)

	return srv
}
