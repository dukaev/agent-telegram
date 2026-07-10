package auth

import (
	"net/http"
	"net/url"

	"agent-telegram/internal/policy"
)

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
	if err := s.savePolicy(nextPolicy); err != nil {
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
	s.startQRCodeFlow()
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
		if err := s.savePolicy(nextPolicy); err != nil {
			writeAuthState(w, http.StatusInternalServerError, s.clientState("Failed to save policy."))
			return
		}
		s.mu.Lock()
		s.policy = nextPolicy
		s.mu.Unlock()
	}

	if body == nil {
		var err error
		body, err = s.finishBody()
		if err != nil {
			writeAuthState(w, http.StatusInternalServerError, s.clientState(err.Error()))
			return
		}
	}
	s.finish(body, nil)
	writeAuthState(w, http.StatusOK, s.clientState(""))
}

func (s *webAuthSession) handleMockAdvance(w http.ResponseWriter, r *http.Request) {
	if !s.authorized(w, r) {
		return
	}
	if !s.runtime.WebMock {
		http.NotFound(w, r)
		return
	}
	action, err := parseAuthField(r, "action", trimAllSpace)
	if err != nil {
		writeAuthState(w, http.StatusBadRequest, s.clientState("Invalid mock payload."))
		return
	}
	if action != "qr_scan" {
		writeAuthState(w, http.StatusBadRequest, s.clientState("Unsupported mock action."))
		return
	}
	s.completeForSetup(r.Context(), nil)
	writeAuthState(w, http.StatusOK, s.clientState(""))
}

func (s *webAuthSession) savePolicy(nextPolicy policy.Policy) error {
	if s.runtime.WebMock {
		return nil
	}
	return policy.SaveDefault(nextPolicy)
}

func (s *webAuthSession) finishBody() (map[string]any, error) {
	if s.runtime.WebMock {
		return s.mockFinishBody()
	}
	return finishAuth(s.cmd, s.runtime, s.state)
}
