package auth

import (
	"context"
	"fmt"
	"time"

	"agent-telegram/internal/authflow"
	"agent-telegram/internal/config"
	"agent-telegram/internal/types"
)

type mockAuthBackend struct {
	sessionData []byte
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
	backend := newMockAuthBackend()

	store := runtime.stateStore()
	state, err := store.Create("", "", appID, appHash, backend.sessionData, runtime.StateTTL)
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
