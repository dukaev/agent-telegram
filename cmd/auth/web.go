package auth

import (
	"bytes"
	"context"
	"crypto/rand"
	"embed"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"log/slog"
	"mime"
	"net"
	"net/http"
	"net/url"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"

	"agent-telegram/internal/authflow"
	"agent-telegram/internal/config"
	"agent-telegram/internal/policy"
	"agent-telegram/internal/types"
	tgapp "agent-telegram/telegram"
	tgtypes "agent-telegram/telegram/types"

	"github.com/gotd/td/telegram/auth/qrlogin"
	"rsc.io/qr"
)

var authWebPort int
var authWebQR = true

//go:embed web_dist/index.html web_dist/assets/*
var authWebAssets embed.FS

// AuthWebCmd starts a local browser-based auth portal.
var AuthWebCmd = &cobra.Command{
	Use:   "web",
	Short: "Login through a local browser page",
	Long: `Start a local browser-based login flow.

The command starts a local login page. QR login is used by default. Pass
--qr=false to use the phone-code flow. The page is printed to stderr as a
one-time localhost URL, then the command waits for completion and emits JSON on
stdout.`,
	Run: runAuthWeb,
}

type webAuthResult struct {
	body map[string]any
	err  error
}

type webAuthSession struct {
	cmd     *cobra.Command
	runtime authRuntimeConfig
	backend authflow.Backend
	store   *authflow.StateStore
	state   *authflow.State
	token   string
	done    chan webAuthResult
	qrMode  bool

	qrImage     string
	qrTokenURL  string
	qrExpires   time.Time
	policy      policy.Policy
	resultBody  map[string]any
	sessionData []byte

	peers        []authPeer
	peersLoaded  bool
	peersLoading bool
	peersError   string
	peerLoader   func(context.Context, *authflow.State, []byte) ([]authPeer, error)

	mu        sync.Mutex
	completed bool
	doneSent  bool
	qrCancel  context.CancelFunc
	qrVersion int
}

func runAuthWeb(cmd *cobra.Command, _ []string) {
	runAuthWebWithRuntime(cmd, authRuntimeFromGlobals())
}

func runAuthWebWithRuntime(cmd *cobra.Command, runtime authRuntimeConfig) {
	_ = godotenv.Load()

	cfg, err := runtime.authConfig(runtime.Phone)
	if err != nil {
		failJSON(err.Error())
	}
	if !runtime.WebQR && cfg.Phone == "" {
		failJSON("phone is required")
	}

	backend := newAuthBackend(cfg)

	var result *types.SendCodeResult
	var sessionData []byte
	if !runtime.WebQR {
		result, err = backend.SendCode(context.Background(), cfg.Phone)
		if err != nil {
			failJSON(fmt.Sprintf("failed to send code: %v", err))
		}
		sessionData, err = backend.ExportSession(context.Background())
		if err != nil {
			failJSON(fmt.Sprintf("failed to export auth session: %v", err))
		}
	}

	store := runtime.stateStore()
	phone := ""
	codeHash := ""
	if !runtime.WebQR {
		phone = cfg.Phone
		codeHash = result.PhoneCodeHash
	}
	state, err := store.Create(phone, codeHash, cfg.AppID, cfg.AppHash, sessionData, runtime.StateTTL)
	if err != nil {
		failJSON(err.Error())
	}

	token, err := newWebToken()
	if err != nil {
		_ = store.Delete(state.ID)
		failJSON(err.Error())
	}

	session := &webAuthSession{
		cmd:         cmd,
		runtime:     runtime,
		backend:     backend,
		store:       store,
		state:       state,
		token:       token,
		qrMode:      runtime.WebQR,
		policy:      loadWebPolicy(),
		sessionData: append([]byte(nil), sessionData...),
		peerLoader:  loadAuthPeers,
		done:        make(chan webAuthResult, 1),
	}

	authCtx, authCancel := context.WithCancel(context.Background())
	defer authCancel()

	server, link, err := startWebAuthServer(authCtx, session, runtime.WebPort)
	if err != nil {
		_ = store.Delete(state.ID)
		failJSON(err.Error())
	}

	promptSuffix := "for this session"
	if state.Phone != "" {
		promptSuffix = "for " + maskPhone(state.Phone)
	}
	if _, err := fmt.Fprintf(
		cmd.ErrOrStderr(),
		"Open this link:\n%s\n\nWaiting for browser login %s...\n",
		link,
		promptSuffix,
	); err != nil {
		_ = store.Delete(state.ID)
		shutdownWebAuthServer(server)
		failJSON(fmt.Sprintf("failed to write auth link: %v", err))
	}

	if runtime.WebQR {
		session.startQRCodeFlow(authCtx)
	}

	select {
	case result := <-session.done:
		session.cancelQRCodeFlow()
		shutdownWebAuthServer(server)
		if result.err != nil {
			failJSON(result.err.Error())
		}
		writeJSON(result.body)
	case <-time.After(time.Until(state.ExpiresAt)):
		session.cancelQRCodeFlow()
		shutdownWebAuthServer(server)
		_ = store.Delete(state.ID)
		failJSON("auth web timed out")
	}
}

func startWebAuthServer(ctx context.Context, session *webAuthSession, port int) (*http.Server, string, error) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /", session.handleRoot)
	mux.HandleFunc("GET /auth", session.handleAuth)
	mux.HandleFunc("GET /auth/state", session.handleState)
	mux.HandleFunc("GET /auth/peers", session.handlePeers)
	mux.HandleFunc("GET /auth/assets/", handleAuthAsset)
	mux.HandleFunc("POST /auth/api", session.handleAPISettings)
	mux.HandleFunc("POST /auth/verify", session.handleVerify)
	mux.HandleFunc("POST /auth/password", session.handlePassword)
	mux.HandleFunc("POST /auth/policy", session.handlePolicy)
	mux.HandleFunc("POST /auth/finish", session.handleFinish)

	lc := &net.ListenConfig{}
	ln, err := lc.Listen(ctx, "tcp", fmt.Sprintf("127.0.0.1:%d", port))
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
	serveAuthIndex(w, r)
}

func (s *webAuthSession) handleState(w http.ResponseWriter, r *http.Request) {
	if !s.authorized(w, r) {
		return
	}
	writeAuthState(w, http.StatusOK, s.clientState(""))
}

func (s *webAuthSession) handlePeers(w http.ResponseWriter, r *http.Request) {
	if !s.authorized(w, r) {
		return
	}
	s.mu.Lock()
	completed := s.completed
	payload := authPeersState{
		Peers:   append([]authPeer(nil), s.peers...),
		Loaded:  s.peersLoaded,
		Loading: s.peersLoading,
		Error:   s.peersError,
	}
	s.mu.Unlock()
	payload.Count = len(payload.Peers)
	if !completed {
		writeAuthPeers(w, http.StatusConflict, authPeersState{Error: "Login is not complete."})
		return
	}
	writeAuthPeers(w, http.StatusOK, payload)
}

func (s *webAuthSession) handlePolicy(w http.ResponseWriter, r *http.Request) {
	if !s.authorized(w, r) {
		return
	}

	nextPolicy, err := parsePolicyRequest(r)
	if err != nil {
		writeAuthResponse(w, r, http.StatusBadRequest, s.clientState("Invalid policy payload."))
		return
	}
	if err := policy.SaveDefault(nextPolicy); err != nil {
		writeAuthResponse(w, r, http.StatusInternalServerError, s.clientState("Failed to save policy."))
		return
	}
	s.mu.Lock()
	s.policy = nextPolicy
	s.mu.Unlock()
	if wantsJSON(r) {
		writeAuthState(w, http.StatusOK, s.clientState(""))
		return
	}
	http.Redirect(w, r, "/auth?t="+url.QueryEscape(s.token), http.StatusFound)
}

func (s *webAuthSession) handleAPISettings(w http.ResponseWriter, r *http.Request) {
	if !s.authorized(w, r) {
		return
	}
	if !s.qrMode {
		writeAuthState(w, http.StatusBadRequest, s.clientState("API settings are only available for QR login."))
		return
	}

	s.mu.Lock()
	completed := s.completed
	doneSent := s.doneSent
	s.mu.Unlock()
	if completed || doneSent {
		writeAuthState(w, http.StatusConflict, s.clientState("Login is already complete."))
		return
	}

	appID, appHash, err := parseAPISettingsRequest(r)
	if err != nil {
		writeAuthState(w, http.StatusBadRequest, s.clientState("Invalid API settings."))
		return
	}
	if appHash == "" {
		writeAuthState(w, http.StatusBadRequest, s.clientState("App hash is required."))
		return
	}

	if err := s.updateAPISettings(appID, appHash); err != nil {
		writeAuthState(w, http.StatusInternalServerError, s.clientState("Failed to update API settings."))
		return
	}
	s.startQRCodeFlow(r.Context())
	writeAuthState(w, http.StatusOK, s.clientState(""))
}

func (s *webAuthSession) handleFinish(w http.ResponseWriter, r *http.Request) {
	if !s.authorized(w, r) {
		return
	}
	s.mu.Lock()
	completed := s.completed
	body := s.resultBody
	s.mu.Unlock()
	if !completed {
		writeAuthState(w, http.StatusConflict, s.clientState("Login is not complete."))
		return
	}

	nextPolicy, ok, err := parseOptionalPolicyRequest(r)
	if err != nil {
		writeAuthState(w, http.StatusBadRequest, s.clientState("Invalid policy payload."))
		return
	}
	if ok {
		if err := policy.SaveDefault(nextPolicy); err != nil {
			writeAuthState(w, http.StatusInternalServerError, s.clientState("Failed to save policy."))
			return
		}
		s.mu.Lock()
		s.policy = nextPolicy
		s.mu.Unlock()
	}

	if body == nil {
		var err error
		body, err = finishAuth(s.cmd, s.runtime, s.state)
		if err != nil {
			writeAuthState(w, http.StatusInternalServerError, s.clientState(err.Error()))
			return
		}
	}
	s.finish(body, nil)
	writeAuthState(w, http.StatusOK, s.clientState(""))
}

func (s *webAuthSession) handleVerify(w http.ResponseWriter, r *http.Request) {
	if !s.authorized(w, r) {
		return
	}
	if s.qrMode {
		writeAuthState(w, http.StatusBadRequest, s.clientState("QR mode does not use a verification code."))
		return
	}

	code, err := parseAuthField(r, "code", trimAllSpace)
	if err != nil {
		writeAuthState(w, http.StatusBadRequest, s.clientState("Invalid verification payload."))
		return
	}
	if code == "" {
		writeAuthState(w, http.StatusBadRequest, s.clientState("Code is required."))
		return
	}

	if err := importStateSession(r.Context(), s.backend, s.state); err != nil {
		writeAuthState(w, http.StatusInternalServerError, s.clientState("Failed to restore auth session."))
		return
	}
	result, err := s.backend.SignIn(r.Context(), s.state.Phone, code, s.state.PhoneCodeHash)
	if err != nil {
		writeAuthState(w, http.StatusBadRequest, s.clientState("Sign in failed."))
		return
	}
	if result.Requires2FA {
		s.mu.Lock()
		s.state.Requires2FA = true
		s.state.TwoFactorHint = result.TwoFactorHint
		s.mu.Unlock()
		if err := s.persistCurrentSession(r.Context()); err != nil {
			s.finish(nil, err)
			writeAuthState(w, http.StatusInternalServerError, s.clientState("Failed to save auth state."))
			return
		}
		writeAuthState(w, http.StatusOK, s.clientState(""))
		return
	}
	if !result.Success {
		writeAuthState(w, http.StatusBadRequest, s.clientState(resultError(result.AuthError, "Authentication failed")))
		return
	}
	s.complete(r.Context(), w)
}

func (s *webAuthSession) handlePassword(w http.ResponseWriter, r *http.Request) {
	if !s.authorized(w, r) {
		return
	}
	if s.qrMode {
		writeAuthState(w, http.StatusBadRequest, s.clientState("QR mode does not use a 2FA form."))
		return
	}

	password, err := parseAuthField(r, "password", trimLineEndings)
	if err != nil {
		writeAuthState(w, http.StatusBadRequest, s.clientState("Invalid password payload."))
		return
	}
	if password == "" {
		writeAuthState(w, http.StatusBadRequest, s.clientState("Password is required."))
		return
	}

	if err := importStateSession(r.Context(), s.backend, s.state); err != nil {
		writeAuthState(w, http.StatusInternalServerError, s.clientState("Failed to restore auth session."))
		return
	}
	result, err := s.backend.SignInWith2FA(r.Context(), s.state.Phone, password)
	if err != nil {
		writeAuthState(w, http.StatusBadRequest, s.clientState("2FA sign in failed."))
		return
	}
	if !result.Success {
		writeAuthState(w, http.StatusBadRequest, s.clientState(resultError(result.AuthError, "2FA authentication failed")))
		return
	}
	s.complete(r.Context(), w)
}

func (s *webAuthSession) complete(ctx context.Context, w http.ResponseWriter) {
	if err := s.persistCurrentSession(ctx); err != nil {
		s.finish(nil, err)
		writeAuthState(w, http.StatusInternalServerError, s.clientState("Failed to save Telegram session."))
		return
	}
	body, err := finishAuth(s.cmd, s.runtime, s.state)
	if err != nil {
		s.finish(nil, err)
		writeAuthState(w, http.StatusInternalServerError, s.clientState("Failed to save Telegram session."))
		return
	}
	s.finish(body, nil)
	writeAuthState(w, http.StatusOK, s.clientState(""))
}

func (s *webAuthSession) completeAsync(ctx context.Context) {
	if err := s.persistCurrentSession(ctx); err != nil {
		s.finish(nil, err)
		return
	}
	s.completeForSetup(ctx, nil)
}

func (s *webAuthSession) startQRCodeFlow(ctx context.Context) {
	s.mu.Lock()
	timeout := time.Until(s.state.ExpiresAt)
	if timeout <= 0 {
		timeout = time.Second
	}
	previousCancel := s.qrCancel
	s.qrVersion++
	version := s.qrVersion
	backend := s.backend
	s.qrImage = ""
	s.qrTokenURL = ""
	s.qrExpires = time.Time{}
	s.mu.Unlock()

	if previousCancel != nil {
		previousCancel()
	}

	flowCtx, cancel := context.WithTimeout(ctx, timeout)
	s.mu.Lock()
	if s.qrVersion == version {
		s.qrCancel = cancel
	}
	s.mu.Unlock()

	go func() {
		defer func() {
			s.mu.Lock()
			if s.qrVersion == version {
				s.qrCancel = nil
			}
			s.mu.Unlock()
			cancel()
		}()

		result, err := backend.SignInWithQR(flowCtx, func(tokenURL string, expiresAt time.Time) error {
			return s.updateQRCode(version, tokenURL, expiresAt)
		})
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(flowCtx.Err(), context.Canceled) {
				return
			}
			s.finish(nil, err)
			return
		}
		if result == nil || !result.Success {
			s.finish(nil, fmt.Errorf("QR authentication failed"))
			return
		}
		s.completeAsync(context.WithoutCancel(flowCtx))
	}()
}

func (s *webAuthSession) cancelQRCodeFlow() {
	s.mu.Lock()
	cancel := s.qrCancel
	s.qrCancel = nil
	s.qrVersion++
	s.mu.Unlock()
	if cancel != nil {
		cancel()
	}
}

func (s *webAuthSession) updateQRCode(version int, tokenURL string, expiresAt time.Time) error {
	qrImage, err := qrCodeImage(tokenURL)
	if err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.completed || version != s.qrVersion {
		return nil
	}
	s.qrImage = qrImage
	s.qrTokenURL = tokenURL
	s.qrExpires = expiresAt
	return nil
}

func (s *webAuthSession) updateAPISettings(appID int, appHash string) error {
	s.mu.Lock()
	phone := s.state.Phone
	s.mu.Unlock()

	cfg := config.LoadFromArgs(appID, appHash, phone, "")
	backend := newAuthBackend(cfg)

	s.mu.Lock()
	s.backend = backend
	s.state.AppID = appID
	s.state.AppHash = appHash
	s.state.SetSessionData(nil)
	s.sessionData = nil
	s.qrImage = ""
	s.qrTokenURL = ""
	s.qrExpires = time.Time{}
	state := *s.state
	s.mu.Unlock()

	return s.store.Save(&state)
}

func (s *webAuthSession) finish(body map[string]any, err error) {
	s.mu.Lock()
	if s.doneSent {
		s.mu.Unlock()
		return
	}
	s.doneSent = true
	if err == nil {
		s.completed = true
		s.resultBody = body
	}
	s.mu.Unlock()
	s.done <- webAuthResult{body: body, err: err}
}

func (s *webAuthSession) completeForSetup(ctx context.Context, body map[string]any) {
	s.mu.Lock()
	if s.doneSent || s.completed {
		s.mu.Unlock()
		return
	}
	s.completed = true
	s.resultBody = body
	s.peersLoading = true
	s.peersLoaded = false
	s.peersError = ""
	s.mu.Unlock()

	go s.loadPeers(ctx)
}

func (s *webAuthSession) loadPeers(ctx context.Context) {
	loader := s.peerLoader
	if loader == nil {
		loader = loadAuthPeers
	}
	s.mu.Lock()
	state := *s.state
	sessionData := append([]byte(nil), s.sessionData...)
	s.mu.Unlock()

	ctx, cancel := context.WithTimeout(ctx, 25*time.Second)
	defer cancel()
	peers, err := loader(ctx, &state, sessionData)

	s.mu.Lock()
	defer s.mu.Unlock()
	s.peersLoading = false
	if err != nil {
		s.peersError = err.Error()
		return
	}
	s.peers = peers
	s.peersLoaded = true
	s.peersError = ""
}

func (s *webAuthSession) persistCurrentSession(ctx context.Context) error {
	sessionData, err := s.backend.ExportSession(ctx)
	if err != nil {
		return fmt.Errorf("export auth session: %w", err)
	}
	s.mu.Lock()
	s.sessionData = append([]byte(nil), sessionData...)
	s.state.SetSessionData(sessionData)
	state := *s.state
	s.mu.Unlock()
	return s.store.Save(&state)
}

func (s *webAuthSession) authorized(w http.ResponseWriter, r *http.Request) bool {
	if r.URL.Query().Get("t") == s.token {
		setAuthHeaders(w)
		return true
	}
	if strings.Contains(r.Header.Get("Content-Type"), "application/x-www-form-urlencoded") {
		_ = r.ParseForm()
	}
	if r.Form.Get("token") == s.token {
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

type authClientState struct {
	Title     string        `json:"title"`
	Message   string        `json:"message,omitempty"`
	Error     string        `json:"error,omitempty"`
	Mode      string        `json:"mode"`
	Completed bool          `json:"completed"`
	Phone     string        `json:"phone,omitempty"`
	Hint      string        `json:"hint,omitempty"`
	QRImage   string        `json:"qrImage,omitempty"`
	QRLink    string        `json:"qrLink,omitempty"`
	Expires   string        `json:"expires,omitempty"`
	Refresh   int           `json:"refresh,omitempty"`
	API       authAPIState  `json:"api"`
	Policy    policy.Policy `json:"policy"`
}

type authAPIState struct {
	AppID   int  `json:"appId"`
	Default bool `json:"default"`
	CanEdit bool `json:"canEdit"`
}

type authPeersState struct {
	Peers   []authPeer `json:"peers"`
	Count   int        `json:"count"`
	Loaded  bool       `json:"loaded"`
	Loading bool       `json:"loading"`
	Error   string     `json:"error,omitempty"`
}

type authPeer struct {
	Peer     string `json:"peer"`
	Title    string `json:"title"`
	Username string `json:"username,omitempty"`
	Type     string `json:"type"`
	ID       int64  `json:"id,omitempty"`
}

func (s *webAuthSession) clientState(errMsg string) authClientState {
	s.mu.Lock()
	completed := s.completed
	doneSent := s.doneSent
	qrMode := s.qrMode
	qrImage := s.qrImage
	qrTokenURL := s.qrTokenURL
	qrExpires := s.qrExpires
	currentPolicy := s.policy
	requires2FA := s.state.Requires2FA
	hint := s.state.TwoFactorHint
	phone := s.state.Phone
	appID := s.state.AppID
	appHash := s.state.AppHash
	s.mu.Unlock()

	if currentPolicy.Version == 0 {
		currentPolicy = policy.Default()
	}

	data := authClientState{
		Error:     errMsg,
		Completed: completed,
		API: authAPIState{
			AppID:   appID,
			Default: isDefaultAPISettings(appID, appHash),
			CanEdit: qrMode && !completed && !doneSent,
		},
		Policy: currentPolicy,
	}
	if completed {
		if doneSent {
			data.Title = "Вход завершен"
			data.Message = "Настройки сохранены."
			data.Mode = "done"
			return data
		}
		data.Title = "Вход выполнен"
		data.Message = "Выбери, с кем агент может взаимодействовать."
		data.Mode = "setup"
		return data
	}

	if qrMode {
		data.Mode = "qr"
		data.Refresh = 1
		data.Message = "Отсканируй код в Telegram."
		if qrImage == "" {
			data.Title = "Готовлю QR-код"
			return data
		}
		data.Title = "Вход по QR-коду"
		data.QRImage = qrImage
		data.QRLink = qrTokenURL
		data.Refresh = qrRefreshDelay(qrExpires)
		if !qrExpires.IsZero() {
			data.Expires = qrExpires.Format(time.RFC3339)
		}
		return data
	}

	if requires2FA {
		data.Title = "Two-step verification"
		data.Mode = "password"
		if hint != "" {
			data.Hint = "2FA hint: " + hint
		} else {
			data.Hint = "Enter your Telegram 2FA password."
		}
		return data
	}

	data.Title = "Telegram login"
	data.Mode = "code"
	data.Phone = maskPhone(phone)
	data.Hint = "Enter the code Telegram sent for " + data.Phone + "."
	return data
}

func isDefaultAPISettings(appID int, appHash string) bool {
	defaultID, err := config.ParseAppID(defaultAppID)
	if err != nil {
		return false
	}
	return appID == defaultID && appHash == defaultAppHash
}

func serveAuthIndex(w http.ResponseWriter, r *http.Request) {
	setAuthHeaders(w)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	data, err := authWebAssets.ReadFile("web_dist/index.html")
	if err != nil {
		http.Error(w, "auth web assets are missing", http.StatusInternalServerError)
		return
	}
	http.ServeContent(w, r, "index.html", time.Time{}, bytes.NewReader(data))
}

func handleAuthAsset(w http.ResponseWriter, r *http.Request) {
	setAuthHeaders(w)

	assetPath := strings.TrimPrefix(r.URL.Path, "/auth/")
	assetPath = path.Clean(assetPath)
	if !strings.HasPrefix(assetPath, "assets/") {
		http.NotFound(w, r)
		return
	}

	data, err := authWebAssets.ReadFile("web_dist/" + assetPath)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	if contentType := mime.TypeByExtension(path.Ext(assetPath)); contentType != "" {
		w.Header().Set("Content-Type", contentType)
	}
	http.ServeContent(w, r, path.Base(assetPath), time.Time{}, bytes.NewReader(data))
}

func writeAuthResponse(w http.ResponseWriter, r *http.Request, status int, state authClientState) {
	if wantsJSON(r) {
		writeAuthState(w, status, state)
		return
	}
	http.Error(w, state.Error, status)
}

func writeAuthState(w http.ResponseWriter, status int, state authClientState) {
	setAuthHeaders(w)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(state); err != nil {
		slog.Debug("failed to write auth state", "error", err)
	}
}

func writeAuthPeers(w http.ResponseWriter, status int, state authPeersState) {
	setAuthHeaders(w)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(state); err != nil {
		slog.Debug("failed to write auth peers", "error", err)
	}
}

func parseAuthField(r *http.Request, name string, trim func(string) string) (string, error) {
	if isJSONRequest(r) {
		var body map[string]string
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			return "", err
		}
		return trim(body[name]), nil
	}
	if err := r.ParseForm(); err != nil {
		return "", err
	}
	return trim(r.FormValue(name)), nil
}

func parsePolicyRequest(r *http.Request) (policy.Policy, error) {
	if isJSONRequest(r) {
		var nextPolicy policy.Policy
		if err := json.NewDecoder(r.Body).Decode(&nextPolicy); err != nil {
			return policy.Policy{}, err
		}
		nextPolicy.Normalize()
		return nextPolicy, nil
	}
	if err := r.ParseForm(); err != nil {
		return policy.Policy{}, err
	}
	nextPolicy := policy.Policy{
		Safeties: policy.Safeties{
			Read:        checkboxOn(r.FormValue("allow_read")),
			Write:       checkboxOn(r.FormValue("allow_write")),
			Destructive: checkboxOn(r.FormValue("allow_destructive")),
			Paid:        checkboxOn(r.FormValue("allow_paid")),
		},
		PeerTypes: policy.PeerTypes{
			Users:    checkboxOn(r.FormValue("allow_users")),
			Groups:   checkboxOn(r.FormValue("allow_groups")),
			Channels: checkboxOn(r.FormValue("allow_channels")),
			Bots:     checkboxOn(r.FormValue("allow_bots")),
		},
		AllowPeers: policy.SplitPeerList(r.FormValue("allow_peers")),
		DenyPeers:  policy.SplitPeerList(r.FormValue("deny_peers")),
	}
	nextPolicy.Normalize()
	return nextPolicy, nil
}

func parseOptionalPolicyRequest(r *http.Request) (policy.Policy, bool, error) {
	if !isJSONRequest(r) {
		return policy.Policy{}, false, nil
	}

	var raw map[string]json.RawMessage
	if err := json.NewDecoder(r.Body).Decode(&raw); err != nil {
		if errors.Is(err, io.EOF) {
			return policy.Policy{}, false, nil
		}
		return policy.Policy{}, false, err
	}

	if payload, ok := raw["policy"]; ok {
		var nextPolicy policy.Policy
		if err := json.Unmarshal(payload, &nextPolicy); err != nil {
			return policy.Policy{}, false, err
		}
		nextPolicy.Normalize()
		return nextPolicy, true, nil
	}

	if _, ok := raw["safeties"]; !ok {
		return policy.Policy{}, false, nil
	}
	data, err := json.Marshal(raw)
	if err != nil {
		return policy.Policy{}, false, err
	}
	var nextPolicy policy.Policy
	if err := json.Unmarshal(data, &nextPolicy); err != nil {
		return policy.Policy{}, false, err
	}
	nextPolicy.Normalize()
	return nextPolicy, true, nil
}

func parseAPISettingsRequest(r *http.Request) (int, string, error) {
	if isJSONRequest(r) {
		var body struct {
			AppID      string `json:"appId"`
			AppHash    string `json:"appHash"`
			UseDefault bool   `json:"useDefault"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			return 0, "", err
		}
		if body.UseDefault {
			appID, err := config.ParseAppID(defaultAppID)
			if err != nil {
				return 0, "", err
			}
			return appID, defaultAppHash, nil
		}
		appID, err := config.ParseAppID(strings.TrimSpace(body.AppID))
		if err != nil {
			return 0, "", err
		}
		return appID, strings.TrimSpace(body.AppHash), nil
	}

	if err := r.ParseForm(); err != nil {
		return 0, "", err
	}
	if checkboxOn(r.FormValue("use_default")) {
		appID, err := config.ParseAppID(defaultAppID)
		if err != nil {
			return 0, "", err
		}
		return appID, defaultAppHash, nil
	}
	appID, err := config.ParseAppID(strings.TrimSpace(r.FormValue("app_id")))
	if err != nil {
		return 0, "", err
	}
	return appID, strings.TrimSpace(r.FormValue("app_hash")), nil
}

func wantsJSON(r *http.Request) bool {
	return strings.Contains(r.Header.Get("Accept"), "application/json") || isJSONRequest(r)
}

func isJSONRequest(r *http.Request) bool {
	return strings.Contains(r.Header.Get("Content-Type"), "application/json")
}

func loadAuthPeers(ctx context.Context, state *authflow.State, sessionData []byte) ([]authPeer, error) {
	if len(sessionData) == 0 {
		return nil, fmt.Errorf("auth session is empty")
	}
	client := tgapp.NewClient(state.AppID, state.AppHash).WithSessionStorage(tgapp.NewMemoryStorage(sessionData))

	clientCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	errCh := make(chan error, 1)
	go func() {
		errCh <- client.Start(clientCtx)
	}()

	select {
	case <-client.Ready():
	case err := <-errCh:
		if err == nil {
			return nil, fmt.Errorf("telegram client stopped before it became ready")
		}
		return nil, err
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	result, err := client.Chat().GetChats(ctx, &tgtypes.GetChatsParams{Limit: 100})
	cancel()
	if err != nil {
		return nil, err
	}
	select {
	case err := <-errCh:
		if err != nil && !errors.Is(err, context.Canceled) {
			return nil, err
		}
	case <-time.After(2 * time.Second):
	}

	return authPeersFromChats(result.Chats), nil
}

func authPeersFromChats(chats []map[string]any) []authPeer {
	peers := make([]authPeer, 0, len(chats))
	for _, chat := range chats {
		peer := authPeerFromChat(chat)
		if peer.Peer == "" {
			continue
		}
		peers = append(peers, peer)
	}
	return peers
}

func authPeerFromChat(chat map[string]any) authPeer {
	rawType := stringValue(chat, "type")
	username := stringValue(chat, "username")
	title := stringValue(chat, "title")
	firstName := stringValue(chat, "first_name")
	lastName := stringValue(chat, "last_name")
	if title == "" {
		title = strings.TrimSpace(firstName + " " + lastName)
	}
	if title == "" && username != "" {
		title = "@" + username
	}

	var id int64
	var kind string
	var peer string
	switch rawType {
	case "user":
		id = int64Value(chat, "user_id")
		kind = "user"
		if boolValue(chat, "bot") {
			kind = "bot"
		}
		if id != 0 {
			peer = fmt.Sprintf("user:%d", id)
		}
	case "chat":
		id = int64Value(chat, "chat_id")
		kind = "group"
		if id != 0 {
			peer = fmt.Sprintf("chat:%d", id)
		}
	case "channel":
		id = int64Value(chat, "channel_id")
		kind = "channel"
		if boolValue(chat, "megagroup") {
			kind = "group"
		}
		if id != 0 {
			peer = fmt.Sprintf("channel:%d", id)
		}
	}

	if peer == "" {
		peer = stringValue(chat, "peer")
	}
	peer = policy.NormalizePeer(peer)
	if title == "" {
		title = peer
	}
	return authPeer{
		Peer:     peer,
		Title:    title,
		Username: username,
		Type:     kind,
		ID:       id,
	}
}

func stringValue(m map[string]any, key string) string {
	if value, ok := m[key].(string); ok {
		return value
	}
	return ""
}

func boolValue(m map[string]any, key string) bool {
	if value, ok := m[key].(bool); ok {
		return value
	}
	return false
}

func int64Value(m map[string]any, key string) int64 {
	switch value := m[key].(type) {
	case int64:
		return value
	case int:
		return int64(value)
	case float64:
		return int64(value)
	case json.Number:
		out, _ := value.Int64()
		return out
	default:
		return 0
	}
}

func qrRefreshDelay(expiresAt time.Time) int {
	if expiresAt.IsZero() {
		return 1
	}
	seconds := int(time.Until(expiresAt).Seconds())
	if seconds <= 0 {
		return 1
	}
	if seconds > 3 {
		return 3
	}
	return seconds + 1
}

func qrCodeImage(tokenURL string) (string, error) {
	if _, err := qrlogin.ParseTokenURL(tokenURL); err != nil {
		return "", fmt.Errorf("parse QR token URL: %w", err)
	}
	code, err := qr.Encode(tokenURL, qr.M)
	if err != nil {
		return "", fmt.Errorf("render QR image: %w", err)
	}

	const (
		quietZone = 4
		scale     = 12
	)
	side := (code.Size + quietZone*2) * scale
	img := image.NewGray(image.Rect(0, 0, side, side))
	for i := range img.Pix {
		img.Pix[i] = 0xff
	}

	black := color.Gray{Y: 0x00}
	for y := 0; y < code.Size; y++ {
		for x := 0; x < code.Size; x++ {
			if !code.Black(x, y) {
				continue
			}
			startX := (x + quietZone) * scale
			startY := (y + quietZone) * scale
			for yy := startY; yy < startY+scale; yy++ {
				for xx := startX; xx < startX+scale; xx++ {
					img.SetGray(xx, yy, black)
				}
			}
		}
	}

	var out bytes.Buffer
	if err := png.Encode(&out, img); err != nil {
		return "", fmt.Errorf("encode QR image: %w", err)
	}
	return "data:image/png;base64," + base64.StdEncoding.EncodeToString(out.Bytes()), nil
}

func loadWebPolicy() policy.Policy {
	p, err := policy.LoadDefault()
	if err != nil {
		return policy.Default()
	}
	return p
}

func checkboxOn(value string) bool {
	return value == "on" || value == "true" || value == "1"
}

func setAuthHeaders(w http.ResponseWriter) {
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Referrer-Policy", "no-referrer")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Content-Security-Policy", strings.Join([]string{
		"default-src 'none'",
		"img-src data: 'self'",
		"script-src 'self'",
		"style-src 'self' 'unsafe-inline'",
		"connect-src 'self'",
		"form-action 'self'",
		"frame-ancestors 'none'",
		"base-uri 'none'",
	}, "; "))
}
