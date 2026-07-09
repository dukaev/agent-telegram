package auth

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"image/png"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/spf13/cobra"

	"agent-telegram/internal/authflow"
	"agent-telegram/internal/config"
	"agent-telegram/internal/policy"
	"agent-telegram/internal/sessionstore"
	"agent-telegram/internal/types"
)

const testAllowUserPolicyBody = `{"policy":{"version":1,` +
	`"safeties":{"read":true,"write":true},` +
	`"peerTypes":{"users":true},"allowPeers":["user:1"]}}`

func TestWebAuthVerifyCompletesLogin(t *testing.T) {
	tmp := t.TempDir()
	resetAuthGlobals(t, tmp)
	state := createTestState(t)
	backend := &fakeAuthBackend{
		sessionData: []byte("web-session"),
		signResult:  &types.SignInResult{Success: true},
	}
	session := newTestWebAuthSession(state, backend)

	req := httptest.NewRequest(http.MethodPost, "/auth/verify?t=test-token", strings.NewReader("code=12345"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	session.handleVerify(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if backend.signedCode != "12345" {
		t.Fatalf("signed code = %q, want 12345", backend.signedCode)
	}
	result := <-session.done
	if result.err != nil {
		t.Fatal(result.err)
	}
	if result.body["ok"] != true || result.body["next"] != "done" {
		t.Fatalf("unexpected result body: %+v", result.body)
	}
	if _, err := authflow.NewStateStore(authStateDir).Load(state.ID); err == nil {
		t.Fatal("state should be removed after successful web login")
	}
	if result.body["sessionStorage"] != "memory" {
		t.Fatalf("session storage = %v", result.body["sessionStorage"])
	}
}

func TestWebAuthVerifyThenPasswordCompletes2FA(t *testing.T) {
	tmp := t.TempDir()
	resetAuthGlobals(t, tmp)
	state := createTestState(t)
	backend := &fakeAuthBackend{
		sessionData: []byte("web-2fa-session"),
		signResult:  &types.SignInResult{Requires2FA: true, TwoFactorHint: "pet"},
		passResult:  &types.SignInResult{Success: true},
	}
	session := newTestWebAuthSession(state, backend)

	req := httptest.NewRequest(http.MethodPost, "/auth/verify?t=test-token", strings.NewReader("code=12345"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	session.handleVerify(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	authState := decodeAuthState(t, rec)
	if authState.Mode != "password" || authState.Title != "Enter your password" {
		t.Fatalf("verify should return password state, got: %+v", authState)
	}
	select {
	case result := <-session.done:
		t.Fatalf("2FA prompt should not finish login yet: %+v", result)
	default:
	}
	loaded, err := authflow.NewStateStore(authStateDir).Load(state.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !loaded.Requires2FA || loaded.TwoFactorHint != "pet" {
		t.Fatalf("state should persist 2FA info: %+v", loaded)
	}

	req = httptest.NewRequest(http.MethodPost, "/auth/password?t=test-token", strings.NewReader("password=secret"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec = httptest.NewRecorder()
	session.handlePassword(rec, req)

	if backend.password != "secret" {
		t.Fatalf("password = %q, want secret", backend.password)
	}
	result := <-session.done
	if result.err != nil {
		t.Fatal(result.err)
	}
	if result.body["ok"] != true || result.body["next"] != "done" {
		t.Fatalf("unexpected result body: %+v", result.body)
	}
}

func TestWebAuthRejectsBadToken(t *testing.T) {
	tmp := t.TempDir()
	resetAuthGlobals(t, tmp)
	state := createTestState(t)
	session := newTestWebAuthSession(state, &fakeAuthBackend{})

	req := httptest.NewRequest(http.MethodGet, "/auth?t=bad-token", nil)
	rec := httptest.NewRecorder()
	session.handleAuth(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rec.Code)
	}
}

func TestWebAuthPolicyFormSavesPolicy(t *testing.T) {
	tmp := t.TempDir()
	resetAuthGlobals(t, tmp)
	state := createTestState(t)
	session := newTestWebAuthSession(state, &fakeAuthBackend{})

	form := url.Values{
		"token":       {"test-token"},
		"allow_read":  {"on"},
		"allow_write": {"on"},
		"allow_users": {"on"},
		"allow_bots":  {"on"},
		"allow_peers": {"ada, @grace"},
		"deny_peers":  {"@blocked"},
	}
	req := httptest.NewRequest(http.MethodPost, "/auth/policy?t=test-token", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	session.handlePolicy(rec, req)

	if rec.Code != http.StatusFound {
		t.Fatalf("status = %d, want 302", rec.Code)
	}
	saved, err := policy.LoadDefault()
	if err != nil {
		t.Fatal(err)
	}
	if !saved.Safeties.Read || !saved.Safeties.Write || saved.Safeties.Destructive || saved.PeerTypes.Groups {
		t.Fatalf("unexpected saved policy: %+v", saved)
	}
	if got := policy.JoinPeerList(saved.AllowPeers); got != "@ada\n@grace" {
		t.Fatalf("allow peers = %q", got)
	}
	if got := policy.JoinPeerList(saved.DenyPeers); got != "@blocked" {
		t.Fatalf("deny peers = %q", got)
	}
}

func TestWebAuthServerEndToEnd(t *testing.T) {
	tmp := t.TempDir()
	resetAuthGlobals(t, tmp)
	state := createTestState(t)
	backend := &fakeAuthBackend{
		sessionData: []byte("web-end-session"),
		signResult:  &types.SignInResult{Success: true},
	}
	session := newTestWebAuthSession(state, backend)

	server, link, err := startWebAuthServer(context.Background(), session, 0)
	if err != nil {
		t.Fatal(err)
	}
	defer shutdownWebAuthServer(server)

	form := url.Values{"code": {"12345"}}
	req, err := http.NewRequestWithContext(
		context.Background(),
		http.MethodPost,
		strings.Replace(link, "/auth?", "/auth/verify?", 1),
		strings.NewReader(form.Encode()),
	)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}

	result := <-session.done
	if result.err != nil {
		t.Fatal(result.err)
	}
	if result.body["ok"] != true || result.body["next"] != "done" {
		t.Fatalf("unexpected result: %+v", result.body)
	}
}

func TestWebAuthStateReturnsPolicy(t *testing.T) {
	tmp := t.TempDir()
	resetAuthGlobals(t, tmp)
	state := createTestState(t)
	session := newTestWebAuthSession(state, &fakeAuthBackend{})
	session.policy = policy.Policy{
		Safeties: policy.Safeties{Read: true},
		PeerTypes: policy.PeerTypes{
			Users: true,
			Bots:  true,
		},
		AllowPeers: []string{"@ada"},
	}
	session.policy.Normalize()

	req := httptest.NewRequest(http.MethodGet, "/auth/state?t=test-token", nil)
	rec := httptest.NewRecorder()
	session.handleState(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	authState := decodeAuthState(t, rec)
	if authState.Mode != "code" {
		t.Fatalf("mode = %q, want code", authState.Mode)
	}
	if !authState.Policy.Safeties.Read || !authState.Policy.PeerTypes.Bots {
		t.Fatalf("unexpected policy in state: %+v", authState.Policy)
	}
	if got := policy.JoinPeerList(authState.Policy.AllowPeers); got != "@ada" {
		t.Fatalf("allow peers = %q", got)
	}
}

func TestWebAuthStateContractForQR(t *testing.T) {
	tmp := t.TempDir()
	resetAuthGlobals(t, tmp)
	state := createTestState(t)
	session := newTestWebAuthSession(state, &fakeAuthBackend{})
	session.qrMode = true
	session.qrImage = "data:image/png;base64,AA=="
	session.qrTokenURL = "tg://login?token=a"
	session.qrExpires = time.Now().Add(time.Minute)

	req := httptest.NewRequest(http.MethodGet, "/auth/state?t=test-token", nil)
	rec := httptest.NewRecorder()
	session.handleState(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	raw := decodeRawObject(t, rec)
	requireJSONFields(
		t,
		raw,
		"title", "message", "mode", "completed", "qrImage", "qrLink", "expires", "refresh", "api", "policy",
	)
	authState := decodeAuthState(t, rec)
	if authState.Mode != "qr" || authState.Completed {
		t.Fatalf("unexpected QR state: %+v", authState)
	}
	if authState.API.AppID != state.AppID || !authState.API.CanEdit {
		t.Fatalf("unexpected API contract: %+v", authState.API)
	}
}

func TestWebAuthStateContractForSetup(t *testing.T) {
	tmp := t.TempDir()
	resetAuthGlobals(t, tmp)
	state := createTestState(t)
	session := newTestWebAuthSession(state, &fakeAuthBackend{})
	session.qrMode = true
	session.completed = true

	req := httptest.NewRequest(http.MethodGet, "/auth/state?t=test-token", nil)
	rec := httptest.NewRecorder()
	session.handleState(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	raw := decodeRawObject(t, rec)
	requireJSONFields(t, raw, "title", "message", "mode", "completed", "api", "policy")
	authState := decodeAuthState(t, rec)
	if authState.Mode != "setup" || !authState.Completed {
		t.Fatalf("unexpected setup state: %+v", authState)
	}
	if authState.API.CanEdit {
		t.Fatalf("API should not be editable after login: %+v", authState.API)
	}
}

func TestWebAuthPeersContract(t *testing.T) {
	tmp := t.TempDir()
	resetAuthGlobals(t, tmp)
	state := createTestState(t)
	session := newTestWebAuthSession(state, &fakeAuthBackend{})
	session.completed = true
	session.peersLoaded = true
	session.peers = []authPeer{{Peer: "user:1", Title: "Ada", Username: "ada", Type: "user", ID: 1}}

	req := httptest.NewRequest(http.MethodGet, "/auth/peers?t=test-token", nil)
	rec := httptest.NewRecorder()
	session.handlePeers(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	var peers authPeersState
	if err := json.Unmarshal(rec.Body.Bytes(), &peers); err != nil {
		t.Fatal(err)
	}
	if peers.Count != 1 || !peers.Loaded || peers.Loading || peers.Peers[0].Peer != "user:1" {
		t.Fatalf("unexpected peers contract: %+v", peers)
	}
}

func TestWebAuthQRCompletionWaitsForFilterSetup(t *testing.T) {
	tmp := t.TempDir()
	resetAuthGlobals(t, tmp)
	state := createTestState(t)
	session := newTestWebAuthSession(state, &fakeAuthBackend{sessionData: []byte("qr-session")})
	session.qrMode = true
	session.peerLoader = func(context.Context, *authflow.State, []byte) ([]authPeer, error) {
		return []authPeer{{Peer: "user:1", Title: "Ada", Type: "user"}}, nil
	}

	session.completeForSetup(context.Background(), map[string]any{"ok": true, "next": "done"})

	select {
	case result := <-session.done:
		t.Fatalf("QR completion should wait for filter setup: %+v", result)
	default:
	}

	req := httptest.NewRequest(http.MethodGet, "/auth/state?t=test-token", nil)
	rec := httptest.NewRecorder()
	session.handleState(rec, req)
	authState := decodeAuthState(t, rec)
	if !authState.Completed || authState.Mode != "setup" || authState.Title != "Set up access" {
		t.Fatalf("unexpected setup state: %+v", authState)
	}

	waitForPeers(t, session)
	req = httptest.NewRequest(http.MethodGet, "/auth/peers?t=test-token", nil)
	rec = httptest.NewRecorder()
	session.handlePeers(rec, req)
	var peers authPeersState
	if err := json.Unmarshal(rec.Body.Bytes(), &peers); err != nil {
		t.Fatal(err)
	}
	if !peers.Loaded || peers.Count != 1 || peers.Peers[0].Peer != "user:1" {
		t.Fatalf("unexpected peers payload: %+v", peers)
	}

	req = httptest.NewRequest(http.MethodPost, "/auth/finish?t=test-token", strings.NewReader(testAllowUserPolicyBody))
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	session.handleFinish(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body: %s", rec.Code, rec.Body.String())
	}

	result := <-session.done
	if result.err != nil {
		t.Fatal(result.err)
	}
	if result.body["ok"] != true {
		t.Fatalf("unexpected result body: %+v", result.body)
	}
	saved, err := policy.LoadDefault()
	if err != nil {
		t.Fatal(err)
	}
	if got := policy.JoinPeerList(saved.AllowPeers); got != "user:1" {
		t.Fatalf("allow peers = %q", got)
	}
}

func TestWebAuthQRCompletionFinishesAfterFilterSetup(t *testing.T) {
	tmp := t.TempDir()
	resetAuthGlobals(t, tmp)
	state := createTestState(t)
	session := newTestWebAuthSession(state, &fakeAuthBackend{sessionData: []byte("qr-session")})
	session.qrMode = true
	session.peerLoader = func(context.Context, *authflow.State, []byte) ([]authPeer, error) {
		return []authPeer{{Peer: "user:1", Title: "Ada", Type: "user"}}, nil
	}

	session.completeAsync(context.Background())

	select {
	case result := <-session.done:
		t.Fatalf("QR completion should not finish before filter setup: %+v", result)
	default:
	}
	if _, err := authflow.NewStateStore(authStateDir).Load(state.ID); err != nil {
		t.Fatalf("state should remain until filter setup finishes: %v", err)
	}

	waitForPeers(t, session)
	req := httptest.NewRequest(
		http.MethodPost,
		"/auth/finish?t=test-token",
		strings.NewReader(testAllowUserPolicyBody),
	)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	session.handleFinish(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body: %s", rec.Code, rec.Body.String())
	}

	result := <-session.done
	if result.err != nil {
		t.Fatal(result.err)
	}
	if result.body["ok"] != true || result.body["next"] != "done" {
		t.Fatalf("unexpected result body: %+v", result.body)
	}
	if _, err := authflow.NewStateStore(authStateDir).Load(state.ID); err == nil {
		t.Fatal("state should be removed after filter setup finishes")
	}
}

func TestWebAuthAPISettingsUpdatesState(t *testing.T) {
	tmp := t.TempDir()
	resetAuthGlobals(t, tmp)
	state := createTestState(t)
	session := newTestWebAuthSession(state, &fakeAuthBackend{})
	session.qrMode = true

	var gotAppID int
	var gotAppHash string
	newAuthBackend = func(cfg *config.Config) authflow.Backend {
		gotAppID = cfg.AppID
		gotAppHash = cfg.AppHash
		return &fakeAuthBackend{
			sessionData: []byte("custom-api-session"),
			signResult:  &types.SignInResult{},
		}
	}

	req := httptest.NewRequest(
		http.MethodPost,
		"/auth/api?t=test-token",
		strings.NewReader(`{"appId":"456","appHash":"custom-hash"}`),
	)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	session.handleAPISettings(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body: %s", rec.Code, rec.Body.String())
	}
	if gotAppID != 456 || gotAppHash != "custom-hash" {
		t.Fatalf("backend config = (%d, %q), want (456, custom-hash)", gotAppID, gotAppHash)
	}
	loaded, err := authflow.NewStateStore(authStateDir).Load(state.ID)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.AppID != 456 || loaded.AppHash != "custom-hash" {
		t.Fatalf("state API settings = (%d, %q), want (456, custom-hash)", loaded.AppID, loaded.AppHash)
	}
	authState := decodeAuthState(t, rec)
	if authState.API.AppID != 456 || authState.API.Default {
		t.Fatalf("unexpected API state: %+v", authState.API)
	}
}

func TestWebAuthOffersAndLoadsSavedSession(t *testing.T) {
	tmp := t.TempDir()
	resetAuthGlobals(t, tmp)
	const profile = "saved-web"
	t.Setenv(sessionstore.EnvProvider, "auth-test-persistent")
	t.Setenv(sessionstore.EnvProfile, profile)
	authTestPersistentStore.data[profile] = []byte("saved-keychain-session")
	t.Cleanup(func() { delete(authTestPersistentStore.data, profile) })

	state := createTestState(t)
	runtime := authRuntimeFromGlobals()
	session := newWebAuthSession(&cobra.Command{}, runtime, webAuthStart{
		backend: &fakeAuthBackend{},
		store:   authflow.NewStateStore(authStateDir),
		state:   state,
	}, "test-token")

	initial := session.clientState("")
	if initial.Session == nil || !initial.Session.Available || !initial.Session.SaveByDefault {
		t.Fatalf("saved session metadata = %+v", initial.Session)
	}
	if initial.Session.Provider != "auth-test-persistent" || initial.Session.Profile != profile {
		t.Fatalf("saved session selection = %+v", initial.Session)
	}

	var loaded []byte
	session.peerLoader = func(_ context.Context, _ *authflow.State, data []byte) ([]authPeer, error) {
		loaded = append([]byte(nil), data...)
		return []authPeer{{Peer: "user:1", Title: "Ada", Type: "user", ID: 1}}, nil
	}
	req := httptest.NewRequest(
		http.MethodPost,
		"/auth/session?t=test-token",
		strings.NewReader(`{"action":"use_saved"}`),
	)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	session.handleSavedSession(rec, req)

	authState := decodeAuthState(t, rec)
	if rec.Code != http.StatusOK || authState.Mode != "setup" || !authState.Completed {
		t.Fatalf("saved session state = %+v, status %d", authState, rec.Code)
	}
	if string(loaded) != "saved-keychain-session" {
		t.Fatalf("loaded session = %q", loaded)
	}
	stored, err := state.SessionData()
	if err != nil || string(stored) != "saved-keychain-session" {
		t.Fatalf("state session = %q, %v", stored, err)
	}
	if len(session.peers) != 1 || !session.peersLoaded {
		t.Fatalf("saved session peers = %+v, loaded %v", session.peers, session.peersLoaded)
	}
}

func TestWebAuthRejectsInvalidSavedSessionAndReturnsToSignIn(t *testing.T) {
	tmp := t.TempDir()
	resetAuthGlobals(t, tmp)
	const profile = "expired-web"
	t.Setenv(sessionstore.EnvProvider, "auth-test-persistent")
	t.Setenv(sessionstore.EnvProfile, profile)
	authTestPersistentStore.data[profile] = []byte("expired-session")
	t.Cleanup(func() { delete(authTestPersistentStore.data, profile) })

	state := createTestState(t)
	session := newWebAuthSession(&cobra.Command{}, authRuntimeFromGlobals(), webAuthStart{
		backend: &fakeAuthBackend{},
		store:   authflow.NewStateStore(authStateDir),
		state:   state,
	}, "test-token")
	session.peerLoader = func(context.Context, *authflow.State, []byte) ([]authPeer, error) {
		return nil, errors.New("telegram authorization expired")
	}

	req := httptest.NewRequest(
		http.MethodPost,
		"/auth/session?t=test-token",
		strings.NewReader(`{"action":"use_saved"}`),
	)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	session.handleSavedSession(rec, req)

	authState := decodeAuthState(t, rec)
	if rec.Code != http.StatusUnauthorized || authState.Mode != "qr" {
		t.Fatalf("invalid saved session state = %+v, status %d", authState, rec.Code)
	}
	if authState.Session == nil || authState.Session.Available || authState.Session.Error == "" {
		t.Fatalf("invalid saved session metadata = %+v", authState.Session)
	}
}

func TestWebAuthMockCanExposeSavedSession(t *testing.T) {
	tmp := t.TempDir()
	resetAuthGlobals(t, tmp)
	runtime := authRuntimeFromGlobals()
	runtime.WebMock = true
	runtime.WebMockSaved = true

	start, err := buildWebAuthStart(runtime)
	if err != nil {
		t.Fatal(err)
	}
	session := newWebAuthSession(&cobra.Command{}, runtime, start, "test-token")
	state := session.clientState("")
	if state.Session == nil || !state.Session.Available || state.Session.Provider != sessionstore.MemoryProvider {
		t.Fatalf("mock saved session = %+v", state.Session)
	}
}

func TestWebAuthMockQRFlowAdvancesToPeerSetup(t *testing.T) {
	tmp := t.TempDir()
	resetAuthGlobals(t, tmp)
	runtime := authRuntimeFromGlobals()
	runtime.WebMock = true
	runtime.WebQR = true

	start, err := buildWebAuthStart(runtime)
	if err != nil {
		t.Fatal(err)
	}
	session := newWebAuthSession(&cobra.Command{}, runtime, start, "test-token")

	req := httptest.NewRequest(
		http.MethodPost,
		"/auth/mock/advance?t=test-token",
		strings.NewReader(`{"action":"qr_scan"}`),
	)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	session.handleMockAdvance(rec, req)

	authState := decodeAuthState(t, rec)
	if rec.Code != http.StatusOK || authState.Mode != "setup" || !authState.Completed {
		t.Fatalf("mock qr state = %+v, status %d", authState, rec.Code)
	}
	waitForPeers(t, session)

	req = httptest.NewRequest(http.MethodGet, "/auth/peers?t=test-token", nil)
	rec = httptest.NewRecorder()
	session.handlePeers(rec, req)
	var peers authPeersState
	if err := json.Unmarshal(rec.Body.Bytes(), &peers); err != nil {
		t.Fatal(err)
	}
	if peers.Count != 5 || !peers.Loaded {
		t.Fatalf("mock peers = %+v", peers)
	}

	req = httptest.NewRequest(http.MethodPost, "/auth/finish?t=test-token", strings.NewReader(testAllowUserPolicyBody))
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	session.handleFinish(rec, req)
	result := <-session.done
	if result.err != nil || result.body["mock"] != true || result.body["next"] != "done" {
		t.Fatalf("mock finish = %#v, err %v", result.body, result.err)
	}
}

func TestWebAuthMockCodeFlowUsesMockCredentials(t *testing.T) {
	tmp := t.TempDir()
	resetAuthGlobals(t, tmp)
	runtime := authRuntimeFromGlobals()
	runtime.WebMock = true
	runtime.WebQR = false
	runtime.Phone = ""

	start, err := buildWebAuthStart(runtime)
	if err != nil {
		t.Fatal(err)
	}
	session := newWebAuthSession(&cobra.Command{}, runtime, start, "test-token")

	req := httptest.NewRequest(http.MethodGet, "/auth/state?t=test-token", nil)
	rec := httptest.NewRecorder()
	session.handleState(rec, req)
	authState := decodeAuthState(t, rec)
	if authState.Mode != "code" || authState.Mock == nil || authState.Mock.Code != mockCode {
		t.Fatalf("initial mock state = %+v", authState)
	}

	req = httptest.NewRequest(http.MethodPost, "/auth/verify?t=test-token", strings.NewReader(`{"code":"`+mockCode+`"}`))
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	session.handleVerify(rec, req)
	authState = decodeAuthState(t, rec)
	if authState.Mode != "password" || authState.Mock == nil || authState.Mock.Password != mockPassword {
		t.Fatalf("mock password state = %+v", authState)
	}

	req = httptest.NewRequest(
		http.MethodPost,
		"/auth/password?t=test-token",
		strings.NewReader(`{"password":"`+mockPassword+`"}`),
	)
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	session.handlePassword(rec, req)
	authState = decodeAuthState(t, rec)
	if rec.Code != http.StatusOK || authState.Mode != "setup" || !authState.Completed {
		t.Fatalf("mock setup state = %+v, status %d", authState, rec.Code)
	}
	waitForPeers(t, session)
}

func TestWebAuthMockCanSwitchBetweenQRAndPhone(t *testing.T) {
	tmp := t.TempDir()
	resetAuthGlobals(t, tmp)
	runtime := authRuntimeFromGlobals()
	runtime.WebMock = true
	runtime.WebQR = true

	start, err := buildWebAuthStart(runtime)
	if err != nil {
		t.Fatal(err)
	}
	session := newWebAuthSession(&cobra.Command{}, runtime, start, "test-token")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	session.setContext(ctx)
	session.startQRCodeFlow()
	defer session.cancelQRCodeFlow()

	postMode := func(body string) authClientState {
		t.Helper()
		req := httptest.NewRequest(http.MethodPost, "/auth/mode?t=test-token", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		session.handleMode(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("mode status = %d; body: %s", rec.Code, rec.Body.String())
		}
		return decodeAuthState(t, rec)
	}

	if state := postMode(`{"mode":"phone"}`); state.Mode != "phone" {
		t.Fatalf("phone-entry state = %+v", state)
	}
	if state := postMode(`{"mode":"code","phone":"+1 (555) 010-1010"}`); state.Mode != "code" || state.Phone == "" {
		t.Fatalf("code state = %+v", state)
	}
	if state := postMode(`{"mode":"qr"}`); state.Mode != "qr" {
		t.Fatalf("qr state = %+v", state)
	}
}

func TestNormalizeAuthPhone(t *testing.T) {
	phone, err := normalizeAuthPhone("+90 (555) 123-45-67")
	if err != nil || phone != "+905551234567" {
		t.Fatalf("normalizeAuthPhone = %q, %v", phone, err)
	}
	if _, err := normalizeAuthPhone("12"); err == nil {
		t.Fatal("short phone should be rejected")
	}
}

func TestAuthPeersFromChatsMapsDialogTypes(t *testing.T) {
	peers := authPeersFromChats([]map[string]any{
		{
			"type":       "user",
			"user_id":    int64(1),
			"first_name": "Ada",
			"username":   "ada",
		},
		{
			"type":       "user",
			"user_id":    int64(2),
			"first_name": "Build Bot",
			"bot":        true,
		},
		{
			"type":       "channel",
			"channel_id": int64(3),
			"title":      "Team",
			"megagroup":  true,
		},
		{
			"type":       "channel",
			"channel_id": int64(4),
			"title":      "News",
		},
	})

	want := []authPeer{
		{Peer: "user:1", Title: "Ada", Username: "ada", Type: "user", ID: 1},
		{Peer: "user:2", Title: "Build Bot", Type: "bot", ID: 2},
		{Peer: "channel:3", Title: "Team", Type: "group", ID: 3},
		{Peer: "channel:4", Title: "News", Type: "channel", ID: 4},
	}
	if !reflect.DeepEqual(peers, want) {
		t.Fatalf("peers = %+v, want %+v", peers, want)
	}
}

func TestQRCodeImageScalesCodeAcrossPNG(t *testing.T) {
	token := base64.URLEncoding.EncodeToString([]byte("telegram-login-token"))
	dataURL, err := qrCodeImage("tg://login?token=" + token)
	if err != nil {
		t.Fatal(err)
	}

	const prefix = "data:image/png;base64,"
	if !strings.HasPrefix(dataURL, prefix) {
		t.Fatalf("QR image prefix = %q", dataURL[:min(len(dataURL), len(prefix))])
	}
	raw, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(dataURL, prefix))
	if err != nil {
		t.Fatal(err)
	}
	img, err := png.Decode(bytes.NewReader(raw))
	if err != nil {
		t.Fatal(err)
	}

	bounds := img.Bounds()
	maxBlackX := bounds.Min.X
	maxBlackY := bounds.Min.Y
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			if r < 0x8000 && g < 0x8000 && b < 0x8000 {
				maxBlackX = max(maxBlackX, x)
				maxBlackY = max(maxBlackY, y)
			}
		}
	}

	if maxBlackX < bounds.Min.X+bounds.Dx()/2 || maxBlackY < bounds.Min.Y+bounds.Dy()/2 {
		t.Fatalf("QR code is not scaled across the PNG: bounds=%v maxBlack=(%d,%d)", bounds, maxBlackX, maxBlackY)
	}
}

func newTestWebAuthSession(state *authflow.State, backend *fakeAuthBackend) *webAuthSession {
	return &webAuthSession{
		cmd:     &cobra.Command{},
		runtime: authRuntimeFromGlobals(),
		backend: backend,
		store:   authflow.NewStateStore(authStateDir),
		state:   state,
		token:   "test-token",
		done:    make(chan webAuthResult, 1),
	}
}

func decodeAuthState(t *testing.T, rec *httptest.ResponseRecorder) authClientState {
	t.Helper()
	var state authClientState
	if err := json.Unmarshal(rec.Body.Bytes(), &state); err != nil {
		t.Fatalf("decode auth state: %v\nbody: %s", err, rec.Body.String())
	}
	return state
}

func decodeRawObject(t *testing.T, rec *httptest.ResponseRecorder) map[string]json.RawMessage {
	t.Helper()
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(rec.Body.Bytes(), &raw); err != nil {
		t.Fatalf("decode raw JSON: %v\nbody: %s", err, rec.Body.String())
	}
	return raw
}

func requireJSONFields(t *testing.T, raw map[string]json.RawMessage, fields ...string) {
	t.Helper()
	for _, field := range fields {
		if _, ok := raw[field]; !ok {
			t.Fatalf("missing JSON field %q in %v", field, raw)
		}
	}
}

func waitForPeers(t *testing.T, session *webAuthSession) {
	t.Helper()
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		session.mu.Lock()
		loaded := session.peersLoaded
		loading := session.peersLoading
		errMsg := session.peersError
		session.mu.Unlock()
		if loaded {
			return
		}
		if !loading && errMsg != "" {
			t.Fatalf("peer loader failed: %s", errMsg)
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("timed out waiting for peers")
}
