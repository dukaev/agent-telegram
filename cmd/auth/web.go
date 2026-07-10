package auth

import (
	"context"
	"crypto/rand"
	"embed"
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"

	"agent-telegram/internal/authflow"
	"agent-telegram/internal/config"
	"agent-telegram/internal/policy"
)

var authWebPort int
var authWebQR = true

//go:embed web_dist/index.html web_dist/assets/*
var authWebAssets embed.FS

type webAuthResult struct {
	body map[string]any
	err  error
}

type webAuthSession struct {
	cmd     *cobra.Command
	runtime authRuntimeConfig
	backend authflow.Backend
	ctx     context.Context
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

	sessionProvider   string
	sessionProfile    string
	sessionPersistent bool
	savedSession      bool
	sessionStoreError string

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

	start, err := buildWebAuthStart(runtime)
	if err != nil {
		failJSON(err.Error())
	}

	token, err := newWebToken()
	if err != nil {
		_ = start.store.Delete(start.state.ID)
		failJSON(err.Error())
	}
	session := newWebAuthSession(cmd, runtime, start, token)

	authCtx, authCancel := context.WithCancel(context.Background())
	defer authCancel()
	session.setContext(authCtx)

	server, link, err := startWebAuthServer(authCtx, session, runtime.WebPort)
	if err != nil {
		_ = start.store.Delete(start.state.ID)
		failJSON(err.Error())
	}

	if err := printWebAuthStart(cmd, link, start.state); err != nil {
		_ = start.store.Delete(start.state.ID)
		shutdownWebAuthServer(server)
		failJSON(fmt.Sprintf("failed to write auth link: %v", err))
	}

	if runtime.WebQR {
		session.startQRCodeFlow()
	}
	waitForWebAuthResult(session, server, start)
}

func startWebAuthServer(ctx context.Context, session *webAuthSession, port int) (*http.Server, string, error) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /", session.handleRoot)
	mux.HandleFunc("GET /auth", session.handleAuth)
	mux.HandleFunc("GET /auth/state", session.handleState)
	mux.HandleFunc("GET /auth/peers", session.handlePeers)
	mux.HandleFunc("GET /auth/assets/", handleAuthAsset)
	mux.HandleFunc("POST /auth/api", session.handleAPISettings)
	mux.HandleFunc("POST /auth/mode", session.handleMode)
	mux.HandleFunc("POST /auth/session", session.handleSavedSession)
	mux.HandleFunc("POST /auth/verify", session.handleVerify)
	mux.HandleFunc("POST /auth/password", session.handlePassword)
	mux.HandleFunc("POST /auth/policy", session.handlePolicy)
	mux.HandleFunc("POST /auth/finish", session.handleFinish)
	mux.HandleFunc("POST /auth/mock/advance", session.handleMockAdvance)

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

func (s *webAuthSession) complete(ctx context.Context, w http.ResponseWriter) {
	if s.runtime.WebMock {
		s.completeForSetup(ctx, nil)
		writeAuthState(w, http.StatusOK, s.clientState(""))
		return
	}

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
	if s.runtime.WebMock {
		s.completeForSetup(ctx, nil)
		return
	}

	if err := s.persistCurrentSession(ctx); err != nil {
		s.finish(nil, err)
		return
	}
	s.completeForSetup(ctx, nil)
}

func (s *webAuthSession) setContext(ctx context.Context) {
	s.mu.Lock()
	s.ctx = ctx
	s.mu.Unlock()
}

func (s *webAuthSession) authContext() context.Context {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.ctx == nil {
		return context.Background()
	}
	return s.ctx
}

func (s *webAuthSession) startQRCodeFlow() {
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

	flowCtx, cancel := context.WithTimeout(s.authContext(), timeout)
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
	var sessionData []byte
	if s.runtime.WebMock {
		mockBackend := newMockAuthBackend()
		backend = mockBackend
		sessionData = mockBackend.sessionData
	}

	s.mu.Lock()
	s.backend = backend
	s.state.AppID = appID
	s.state.AppHash = appHash
	s.state.SetSessionData(sessionData)
	s.sessionData = append([]byte(nil), sessionData...)
	s.qrImage = ""
	s.qrTokenURL = ""
	s.qrExpires = time.Time{}
	state := *s.state
	s.mu.Unlock()

	return s.store.Save(&state)
}

func (s *webAuthSession) resetAuthMode(qrMode bool) error {
	s.mu.Lock()
	appID := s.state.AppID
	appHash := s.state.AppHash
	s.mu.Unlock()

	backend := newAuthBackend(config.LoadFromArgs(appID, appHash, "", ""))
	var sessionData []byte
	if s.runtime.WebMock {
		mockBackend := newMockAuthBackend()
		backend = mockBackend
		sessionData = mockBackend.sessionData
	}

	s.mu.Lock()
	s.backend = backend
	s.qrMode = qrMode
	s.state.Phone = ""
	s.state.PhoneCodeHash = ""
	s.state.Requires2FA = false
	s.state.TwoFactorHint = ""
	s.state.SetSessionData(sessionData)
	s.sessionData = append([]byte(nil), sessionData...)
	s.qrImage = ""
	s.qrTokenURL = ""
	s.qrExpires = time.Time{}
	state := *s.state
	s.mu.Unlock()

	return s.store.Save(&state)
}

func (s *webAuthSession) beginPhoneCode(ctx context.Context, phone string) error {
	s.mu.Lock()
	appID := s.state.AppID
	appHash := s.state.AppHash
	s.mu.Unlock()

	backend := newAuthBackend(config.LoadFromArgs(appID, appHash, phone, ""))
	if s.runtime.WebMock {
		backend = newMockAuthBackend()
	}
	result, err := backend.SendCode(ctx, phone)
	if err != nil {
		return fmt.Errorf("send Telegram code: %w", err)
	}
	sessionData, err := backend.ExportSession(ctx)
	if err != nil {
		return fmt.Errorf("export auth session: %w", err)
	}

	s.mu.Lock()
	s.backend = backend
	s.qrMode = false
	s.state.Phone = phone
	s.state.PhoneCodeHash = result.PhoneCodeHash
	s.state.Requires2FA = false
	s.state.TwoFactorHint = ""
	s.state.SetSessionData(sessionData)
	s.sessionData = append([]byte(nil), sessionData...)
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
	Title     string           `json:"title"`
	Message   string           `json:"message,omitempty"`
	Error     string           `json:"error,omitempty"`
	Mode      string           `json:"mode"`
	Completed bool             `json:"completed"`
	Phone     string           `json:"phone,omitempty"`
	Hint      string           `json:"hint,omitempty"`
	QRImage   string           `json:"qrImage,omitempty"`
	QRLink    string           `json:"qrLink,omitempty"`
	Expires   string           `json:"expires,omitempty"`
	Refresh   int              `json:"refresh,omitempty"`
	API       authAPIState     `json:"api"`
	Policy    policy.Policy    `json:"policy"`
	Mock      *authMockInfo    `json:"mock,omitempty"`
	Session   *authSessionInfo `json:"session,omitempty"`
}

type authSessionInfo struct {
	Provider      string `json:"provider"`
	Profile       string `json:"profile"`
	Persistent    bool   `json:"persistent"`
	Available     bool   `json:"available"`
	SaveByDefault bool   `json:"saveByDefault"`
	Error         string `json:"error,omitempty"`
}

type authAPIState struct {
	AppID   int  `json:"appId"`
	Default bool `json:"default"`
	CanEdit bool `json:"canEdit"`
}

type authMockInfo struct {
	Enabled  bool   `json:"enabled"`
	Code     string `json:"code,omitempty"`
	Password string `json:"password,omitempty"`
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
