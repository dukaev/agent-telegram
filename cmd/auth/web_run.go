package auth

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/spf13/cobra"

	"agent-telegram/internal/authflow"
	"agent-telegram/internal/policy"
	"agent-telegram/internal/sessionstore"
)

type webAuthStart struct {
	backend     authflow.Backend
	store       *authflow.StateStore
	state       *authflow.State
	sessionData []byte
}

func buildWebAuthStart(runtime authRuntimeConfig) (webAuthStart, error) {
	if runtime.WebMock {
		return mockWebAuthStart(runtime)
	}

	cfg, err := runtime.authConfig()
	if err != nil {
		return webAuthStart{}, err
	}
	backend := newAuthBackend(cfg)

	store := runtime.stateStore()
	state, err := store.Create("", "", cfg.AppID, cfg.AppHash, nil, runtime.StateTTL)
	if err != nil {
		return webAuthStart{}, err
	}
	return webAuthStart{backend: backend, store: store, state: state}, nil
}

func newWebAuthSession(
	cmd *cobra.Command,
	runtime authRuntimeConfig,
	start webAuthStart,
	token string,
) *webAuthSession {
	session := &webAuthSession{
		cmd:         cmd,
		runtime:     runtime,
		backend:     start.backend,
		ctx:         context.Background(),
		store:       start.store,
		state:       start.state,
		token:       token,
		policy:      webAuthInitialPolicy(runtime),
		sessionData: append([]byte(nil), start.sessionData...),
		peerLoader:  webAuthPeerLoader(runtime),
		done:        make(chan webAuthResult, 1),
	}
	session.configureSessionStore()
	return session
}

func (s *webAuthSession) configureSessionStore() {
	provider, profile := authSessionSelection(s.cmd)
	if s.runtime.WebMock {
		provider = sessionstore.MemoryProvider
		profile = "mock"
	}
	storage, err := sessionstore.Open(provider, profile)
	if err != nil {
		s.sessionProvider = firstNonEmpty(provider, sessionstore.DefaultProvider())
		s.sessionProfile = firstNonEmpty(profile, sessionstore.DefaultProfile)
		s.sessionStoreError = err.Error()
		return
	}
	selection := storage.Selection()
	s.sessionProvider = selection.Provider
	s.sessionProfile = selection.Profile
	s.sessionPersistent = selection.Persistent
}

func webAuthInitialPolicy(runtime authRuntimeConfig) policy.Policy {
	if runtime.WebMock {
		return policy.Default()
	}
	return loadWebPolicy()
}

func webAuthPeerLoader(runtime authRuntimeConfig) func(context.Context, *authflow.State, []byte) ([]authPeer, error) {
	if runtime.WebMock {
		return mockAuthPeers
	}
	return loadAuthPeers
}

func printWebAuthStart(cmd *cobra.Command, link string, state *authflow.State) error {
	promptSuffix := "for this session"
	if state.Phone != "" {
		promptSuffix = "for " + maskPhone(state.Phone)
	}
	_, err := fmt.Fprintf(
		cmd.ErrOrStderr(),
		"Open this link:\n%s\n\nWaiting for browser login %s...\n",
		link,
		promptSuffix,
	)
	return err
}

func waitForWebAuthResult(session *webAuthSession, server *http.Server, start webAuthStart) {
	select {
	case result := <-session.done:
		session.cancelQRCodeFlow()
		shutdownWebAuthServer(server)
		if result.err != nil {
			failJSON(result.err.Error())
		}
		writeJSON(result.body)
	case <-time.After(time.Until(start.state.ExpiresAt)):
		session.cancelQRCodeFlow()
		shutdownWebAuthServer(server)
		_ = start.store.Delete(start.state.ID)
		failJSON("auth timed out")
	}
}
