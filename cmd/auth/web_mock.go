package auth

import (
	"context"
	"fmt"
	"time"

	"agent-telegram/internal/authflow"
	"agent-telegram/internal/config"
	"agent-telegram/internal/types"
)

const (
	mockPhone    = "+15550101010"
	mockCode     = "22222"
	mockPassword = "mock-password"
)

type mockAuthBackend struct {
	sessionData []byte
	imported    []byte
}

func newMockAuthBackend() *mockAuthBackend {
	return &mockAuthBackend{sessionData: []byte("mock-web-auth-session")}
}

func mockWebAuthStart(runtime authRuntimeConfig) (webAuthStart, error) {
	appID, err := config.ParseAppID(firstNonEmpty(runtime.AppID, defaultAppID))
	if err != nil {
		return webAuthStart{}, fmt.Errorf("invalid mock app-id: %w", err)
	}
	appHash := firstNonEmpty(runtime.AppHash, defaultAppHash)
	phone := firstNonEmpty(runtime.Phone, mockPhone)
	backend := newMockAuthBackend()

	var codeHash string
	if !runtime.WebQR {
		codeHash = "mock-code-hash"
	}
	store := runtime.stateStore()
	state, err := store.Create(phone, codeHash, appID, appHash, backend.sessionData, runtime.StateTTL)
	if err != nil {
		return webAuthStart{}, err
	}
	return webAuthStart{
		backend:     backend,
		store:       store,
		state:       state,
		sessionData: append([]byte(nil), backend.sessionData...),
	}, nil
}

func (b *mockAuthBackend) SendCode(_ context.Context, _ string) (*types.SendCodeResult, error) {
	return &types.SendCodeResult{PhoneCodeHash: "mock-code-hash", Timeout: 30}, nil
}

func (b *mockAuthBackend) SignIn(_ context.Context, _, code, _ string) (*types.SignInResult, error) {
	if code != mockCode {
		return &types.SignInResult{AuthError: "Use mock code " + mockCode}, nil
	}
	return &types.SignInResult{Requires2FA: true, TwoFactorHint: "mock-password"}, nil
}

func (b *mockAuthBackend) SignInWith2FA(_ context.Context, _ string, password string) (*types.SignInResult, error) {
	if password != mockPassword {
		return &types.SignInResult{AuthError: "Use mock password " + mockPassword}, nil
	}
	return &types.SignInResult{Success: true}, nil
}

func (b *mockAuthBackend) SignInWithQR(
	ctx context.Context,
	onToken func(tokenURL string, expiresAt time.Time) error,
) (*types.SignInResult, error) {
	if onToken != nil {
		if err := onToken(mockQRTokenURL(), time.Now().Add(5*time.Minute)); err != nil {
			return nil, err
		}
	}
	<-ctx.Done()
	return nil, ctx.Err()
}

func (b *mockAuthBackend) ImportSession(_ context.Context, data []byte) error {
	b.imported = append([]byte(nil), data...)
	return nil
}

func (b *mockAuthBackend) ExportSession(_ context.Context) ([]byte, error) {
	return append([]byte(nil), b.sessionData...), nil
}

func mockQRTokenURL() string {
	return "tg://login?token=bW9jay13ZWItYXV0aC10b2tlbg=="
}

func mockAuthPeers(_ context.Context, _ *authflow.State, _ []byte) ([]authPeer, error) {
	return []authPeer{
		{Peer: "user:101", Title: "Ada Lovelace", Username: "ada", Type: "user", ID: 101},
		{Peer: "user:202", Title: "Grace Hopper", Username: "grace", Type: "user", ID: 202},
		{Peer: "chat:303", Title: "Agent QA Group", Type: "group", ID: 303},
		{Peer: "channel:404", Title: "Release Notes", Username: "release_notes", Type: "channel", ID: 404},
		{Peer: "user:505", Title: "Build Bot", Username: "buildbot", Type: "bot", ID: 505},
	}, nil
}

func (s *webAuthSession) mockFinishBody() (map[string]any, error) {
	if err := s.store.Delete(s.state.ID); err != nil {
		return nil, err
	}
	return map[string]any{
		"ok":             true,
		"next":           "done",
		"phone":          maskPhone(s.state.Phone),
		"sessionStorage": "memory",
		"serverReloaded": false,
		"mock":           true,
	}, nil
}
