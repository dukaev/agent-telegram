package auth

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"agent-telegram/internal/authflow"
	"agent-telegram/internal/types"
)

func TestWebAuthVerifyCompletesLogin(t *testing.T) {
	tmp := t.TempDir()
	resetAuthGlobals(t, tmp)
	state := createTestState(t, tmp)
	backend := &fakeAuthBackend{
		sessionPath: state.SessionPath,
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
	if _, err := filepath.Abs(result.body["sessionPath"].(string)); err != nil {
		t.Fatalf("session path should be a path: %v", err)
	}
}

func TestWebAuthVerifyThenPasswordCompletes2FA(t *testing.T) {
	tmp := t.TempDir()
	resetAuthGlobals(t, tmp)
	state := createTestState(t, tmp)
	backend := &fakeAuthBackend{
		sessionPath: state.SessionPath,
		signResult:  &types.SignInResult{Requires2FA: true, TwoFactorHint: "pet"},
		passResult:  &types.SignInResult{Success: true},
	}
	session := newTestWebAuthSession(state, backend)

	req := httptest.NewRequest(http.MethodPost, "/auth/verify?t=test-token", strings.NewReader("code=12345"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	session.handleVerify(rec, req)

	if !strings.Contains(rec.Body.String(), "Two-step verification") {
		t.Fatalf("verify should render password form, got: %s", rec.Body.String())
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
	state := createTestState(t, tmp)
	session := newTestWebAuthSession(state, &fakeAuthBackend{sessionPath: state.SessionPath})

	req := httptest.NewRequest(http.MethodGet, "/auth?t=bad-token", nil)
	rec := httptest.NewRecorder()
	session.handleAuth(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rec.Code)
	}
}

func TestWebAuthServerEndToEnd(t *testing.T) {
	tmp := t.TempDir()
	resetAuthGlobals(t, tmp)
	state := createTestState(t, tmp)
	backend := &fakeAuthBackend{
		sessionPath: state.SessionPath,
		signResult:  &types.SignInResult{Success: true},
	}
	session := newTestWebAuthSession(state, backend)

	server, link, err := startWebAuthServer(session, 0)
	if err != nil {
		t.Fatal(err)
	}
	defer shutdownWebAuthServer(server)

	resp, err := http.PostForm(strings.Replace(link, "/auth?", "/auth/verify?", 1), url.Values{"code": {"12345"}})
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

func newTestWebAuthSession(state *authflow.State, backend *fakeAuthBackend) *webAuthSession {
	return &webAuthSession{
		cmd:     &cobra.Command{},
		backend: backend,
		store:   authflow.NewStateStore(authStateDir),
		state:   state,
		token:   "test-token",
		done:    make(chan webAuthResult, 1),
	}
}
