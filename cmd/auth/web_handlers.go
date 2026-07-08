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
