package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"html"
	"net"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"

	"agent-telegram/internal/authflow"
)

var authWebPort int

// AuthWebCmd starts a local browser-based auth portal.
var AuthWebCmd = &cobra.Command{
	Use:   "web",
	Short: "Login through a local browser page",
	Long: `Start a local browser-based login flow.

The command sends a Telegram login code, prints a one-time localhost URL to
stderr, waits for the browser form to complete, then emits final JSON on stdout.`,
	Run: runAuthWeb,
}

type webAuthResult struct {
	body map[string]any
	err  error
}

type webAuthSession struct {
	cmd     *cobra.Command
	backend authflow.Backend
	store   *authflow.StateStore
	state   *authflow.State
	token   string
	done    chan webAuthResult

	mu        sync.Mutex
	completed bool
}

func runAuthWeb(cmd *cobra.Command, _ []string) {
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

	token, err := newWebToken()
	if err != nil {
		_ = store.Delete(state.ID)
		failJSON(err.Error())
	}

	session := &webAuthSession{
		cmd:     cmd,
		backend: backend,
		store:   store,
		state:   state,
		token:   token,
		done:    make(chan webAuthResult, 1),
	}

	server, link, err := startWebAuthServer(session, authWebPort)
	if err != nil {
		_ = store.Delete(state.ID)
		failJSON(err.Error())
	}

	fmt.Fprintf(cmd.ErrOrStderr(), "Open this link:\n%s\n\nWaiting for browser login for %s...\n", link, maskPhone(state.Phone))

	select {
	case result := <-session.done:
		shutdownWebAuthServer(server)
		if result.err != nil {
			failJSON(result.err.Error())
		}
		writeJSON(result.body)
	case <-time.After(time.Until(state.ExpiresAt)):
		shutdownWebAuthServer(server)
		_ = store.Delete(state.ID)
		failJSON("auth web timed out")
	}
}

func startWebAuthServer(session *webAuthSession, port int) (*http.Server, string, error) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /", session.handleRoot)
	mux.HandleFunc("GET /auth", session.handleAuth)
	mux.HandleFunc("POST /auth/verify", session.handleVerify)
	mux.HandleFunc("POST /auth/password", session.handlePassword)

	ln, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		return nil, "", fmt.Errorf("start local auth listener: %w", err)
	}

	server := &http.Server{
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}
	go func() {
		if err := server.Serve(ln); err != nil && err != http.ErrServerClosed {
			session.finish(nil, fmt.Errorf("local auth server error: %w", err))
		}
	}()

	link := (&url.URL{
		Scheme:   "http",
		Host:     ln.Addr().String(),
		Path:     "/auth",
		RawQuery: "t=" + url.QueryEscape(session.token),
	}).String()
	return server, link, nil
}

func shutdownWebAuthServer(server *http.Server) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_ = server.Shutdown(ctx)
}

func (s *webAuthSession) handleRoot(w http.ResponseWriter, r *http.Request) {
	if r.URL.Query().Get("t") != s.token {
		http.NotFound(w, r)
		return
	}
	http.Redirect(w, r, "/auth?t="+url.QueryEscape(s.token), http.StatusFound)
}

func (s *webAuthSession) handleAuth(w http.ResponseWriter, r *http.Request) {
	if !s.authorized(w, r) {
		return
	}
	s.mu.Lock()
	requires2FA := s.state.Requires2FA
	hint := s.state.TwoFactorHint
	completed := s.completed
	s.mu.Unlock()

	switch {
	case completed:
		renderAuthPage(w, authPageData{Title: "Login complete", Message: "You can close this page."})
	case requires2FA:
		renderAuthPage(w, passwordPageData(s.token, hint, ""))
	default:
		renderAuthPage(w, codePageData(s.token, maskPhone(s.state.Phone), ""))
	}
}

func (s *webAuthSession) handleVerify(w http.ResponseWriter, r *http.Request) {
	if !s.authorized(w, r) {
		return
	}
	if err := r.ParseForm(); err != nil {
		renderAuthPage(w, codePageData(s.token, maskPhone(s.state.Phone), "Invalid form"))
		return
	}
	code := trimAllSpace(r.FormValue("code"))
	if code == "" {
		renderAuthPage(w, codePageData(s.token, maskPhone(s.state.Phone), "Code is required"))
		return
	}

	result, err := s.backend.SignIn(r.Context(), s.state.Phone, code, s.state.PhoneCodeHash)
	if err != nil {
		renderAuthPage(w, codePageData(s.token, maskPhone(s.state.Phone), "Sign in failed"))
		return
	}
	if result.Requires2FA {
		s.mu.Lock()
		s.state.Requires2FA = true
		s.state.TwoFactorHint = result.TwoFactorHint
		s.mu.Unlock()
		if err := s.store.Save(s.state); err != nil {
			s.finish(nil, err)
			renderAuthPage(w, authPageData{Title: "Login failed", Error: "Failed to save auth state."})
			return
		}
		renderAuthPage(w, passwordPageData(s.token, result.TwoFactorHint, ""))
		return
	}
	if !result.Success {
		renderAuthPage(w, codePageData(s.token, maskPhone(s.state.Phone), resultError(result.AuthError, "Authentication failed")))
		return
	}
	s.complete(w)
}

func (s *webAuthSession) handlePassword(w http.ResponseWriter, r *http.Request) {
	if !s.authorized(w, r) {
		return
	}
	if err := r.ParseForm(); err != nil {
		renderAuthPage(w, passwordPageData(s.token, s.state.TwoFactorHint, "Invalid form"))
		return
	}
	password := trimLineEndings(r.FormValue("password"))
	if password == "" {
		renderAuthPage(w, passwordPageData(s.token, s.state.TwoFactorHint, "Password is required"))
		return
	}

	result, err := s.backend.SignInWith2FA(r.Context(), s.state.Phone, password)
	if err != nil {
		renderAuthPage(w, passwordPageData(s.token, s.state.TwoFactorHint, "2FA sign in failed"))
		return
	}
	if !result.Success {
		renderAuthPage(w, passwordPageData(s.token, s.state.TwoFactorHint, resultError(result.AuthError, "2FA authentication failed")))
		return
	}
	s.complete(w)
}

func (s *webAuthSession) complete(w http.ResponseWriter) {
	body, err := finishAuth(s.cmd, s.state)
	if err != nil {
		s.finish(nil, err)
		renderAuthPage(w, authPageData{Title: "Login failed", Error: "Failed to save Telegram session."})
		return
	}
	s.finish(body, nil)
	renderAuthPage(w, authPageData{Title: "Login complete", Message: "You can close this page and return to the terminal."})
}

func (s *webAuthSession) finish(body map[string]any, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.completed {
		return
	}
	s.completed = true
	s.done <- webAuthResult{body: body, err: err}
}

func (s *webAuthSession) authorized(w http.ResponseWriter, r *http.Request) bool {
	if r.URL.Query().Get("t") == s.token || r.FormValue("token") == s.token {
		setAuthHeaders(w)
		return true
	}
	http.NotFound(w, r)
	return false
}

func newWebToken() (string, error) {
	var buf [32]byte
	if _, err := rand.Read(buf[:]); err != nil {
		return "", fmt.Errorf("generate web auth token: %w", err)
	}
	return hex.EncodeToString(buf[:]), nil
}

type authPageData struct {
	Title   string
	Message string
	Error   string
	Action  string
	Token   string
	Field   string
	Type    string
	Label   string
	Hint    string
	Button  string
}

func codePageData(token, phone, errMsg string) authPageData {
	return authPageData{
		Title:  "Telegram login",
		Hint:   "Enter the code Telegram sent for " + phone + ".",
		Error:  errMsg,
		Action: "/auth/verify",
		Token:  token,
		Field:  "code",
		Type:   "text",
		Label:  "Code",
		Button: "Verify code",
	}
}

func passwordPageData(token, hint, errMsg string) authPageData {
	if hint != "" {
		hint = "2FA hint: " + hint
	} else {
		hint = "Enter your Telegram 2FA password."
	}
	return authPageData{
		Title:  "Two-step verification",
		Hint:   hint,
		Error:  errMsg,
		Action: "/auth/password",
		Token:  token,
		Field:  "password",
		Type:   "password",
		Label:  "Password",
		Button: "Complete login",
	}
}

func renderAuthPage(w http.ResponseWriter, data authPageData) {
	setAuthHeaders(w)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	fmt.Fprintf(w, `<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>%s</title>
<style>
:root { color-scheme: light dark; font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif; }
body { margin: 0; min-height: 100vh; display: grid; place-items: center; background: #f5f7fb; color: #101828; }
main { width: min(420px, calc(100vw - 32px)); }
h1 { font-size: 24px; margin: 0 0 12px; }
p { line-height: 1.5; color: #475467; }
form { display: grid; gap: 12px; }
label { font-weight: 600; }
input { box-sizing: border-box; width: 100%%; height: 44px; padding: 0 12px; font-size: 18px; border: 1px solid #cfd7e6; border-radius: 8px; }
button { height: 44px; border: 0; border-radius: 8px; background: #1976d2; color: white; font-size: 16px; font-weight: 700; cursor: pointer; }
.panel { background: white; border: 1px solid #e3e8f2; border-radius: 8px; padding: 24px; box-shadow: 0 12px 30px rgba(16, 24, 40, .08); }
.error { color: #b42318; font-weight: 600; }
@media (prefers-color-scheme: dark) {
  body { background: #101828; color: #f9fafb; }
  .panel { background: #182230; border-color: #344054; }
  p { color: #d0d5dd; }
  input { background: #101828; border-color: #475467; color: #f9fafb; }
}
</style>
</head>
<body><main><section class="panel">
<h1>%s</h1>
`, html.EscapeString(data.Title), html.EscapeString(data.Title))

	if data.Error != "" {
		fmt.Fprintf(w, "<p class=\"error\">%s</p>\n", html.EscapeString(data.Error))
	}
	if data.Message != "" {
		fmt.Fprintf(w, "<p>%s</p>\n", html.EscapeString(data.Message))
	}
	if data.Action != "" {
		fmt.Fprintf(w, `<p>%s</p>
<form method="post" action="%s?t=%s" autocomplete="off">
<input type="hidden" name="token" value="%s">
<label for="%s">%s</label>
<input id="%s" name="%s" type="%s" autofocus required>
<button type="submit">%s</button>
</form>
`, html.EscapeString(data.Hint),
			html.EscapeString(data.Action),
			url.QueryEscape(data.Token),
			html.EscapeString(data.Token),
			html.EscapeString(data.Field),
			html.EscapeString(data.Label),
			html.EscapeString(data.Field),
			html.EscapeString(data.Field),
			html.EscapeString(data.Type),
			html.EscapeString(data.Button),
		)
	}
	fmt.Fprint(w, "</section></main></body></html>")
}

func setAuthHeaders(w http.ResponseWriter) {
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Referrer-Policy", "no-referrer")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Content-Security-Policy", "default-src 'none'; style-src 'unsafe-inline'; form-action 'self'; frame-ancestors 'none'; base-uri 'none'")
}
