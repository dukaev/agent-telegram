package auth

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/spf13/cobra"

	"agent-telegram/internal/authflow"
	"agent-telegram/internal/config"
	"agent-telegram/internal/policy"
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

	cfg, err := runtime.authConfig(runtime.Phone)
	if err != nil {
		return webAuthStart{}, err
	}
	if !runtime.WebQR && cfg.Phone == "" {
		return webAuthStart{}, fmt.Errorf("phone is required")
	}

	backend := newAuthBackend(cfg)
	phone, codeHash, sessionData, err := beginWebAuthBackend(runtime, cfg, backend)
	if err != nil {
		return webAuthStart{}, err
	}

	store := runtime.stateStore()
	state, err := store.Create(phone, codeHash, cfg.AppID, cfg.AppHash, sessionData, runtime.StateTTL)
	if err != nil {
		return webAuthStart{}, err
	}
	return webAuthStart{backend: backend, store: store, state: state, sessionData: sessionData}, nil
}

func beginWebAuthBackend(
	runtime authRuntimeConfig,
	cfg *config.Config,
	backend authflow.Backend,
) (phone string, codeHash string, sessionData []byte, err error) {
	if runtime.WebQR {
		return "", "", nil, nil
	}

	result, err := backend.SendCode(context.Background(), cfg.Phone)
	if err != nil {
		return "", "", nil, fmt.Errorf("failed to send code: %w", err)
	}
	sessionData, err = backend.ExportSession(context.Background())
	if err != nil {
		return "", "", nil, fmt.Errorf("failed to export auth session: %w", err)
	}
	return cfg.Phone, result.PhoneCodeHash, sessionData, nil
}

func newWebAuthSession(
	cmd *cobra.Command,
	runtime authRuntimeConfig,
	start webAuthStart,
	token string,
) *webAuthSession {
	return &webAuthSession{
		cmd:         cmd,
		runtime:     runtime,
		backend:     start.backend,
		ctx:         context.Background(),
		store:       start.store,
		state:       start.state,
		token:       token,
		qrMode:      runtime.WebQR,
		policy:      webAuthInitialPolicy(runtime),
		sessionData: append([]byte(nil), start.sessionData...),
		peerLoader:  webAuthPeerLoader(runtime),
		done:        make(chan webAuthResult, 1),
	}
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
		failJSON("auth web timed out")
	}
}
