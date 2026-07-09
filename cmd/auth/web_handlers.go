package auth

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
	"unicode"

	"agent-telegram/internal/policy"
	"agent-telegram/internal/sessionstore"
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
	qrMode := s.qrMode
	phone := s.state.Phone
	s.mu.Unlock()
	if !qrMode && phone != "" {
		writeAuthState(w, http.StatusBadRequest, s.clientState("API settings can only be changed before sending a code."))
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
	if qrMode {
		s.startQRCodeFlow()
	}
	writeAuthState(w, http.StatusOK, s.clientState(""))
}

func (s *webAuthSession) handleMode(w http.ResponseWriter, r *http.Request) {
	if !s.authorized(w, r) {
		return
	}
	mode, phone, err := parseAuthModeRequest(r)
	if err != nil {
		writeAuthState(w, http.StatusBadRequest, s.clientState("Could not change the sign-in method."))
		return
	}

	s.mu.Lock()
	completed := s.completed
	doneSent := s.doneSent
	s.mu.Unlock()
	if completed || doneSent {
		writeAuthState(w, http.StatusConflict, s.clientState("Sign-in is already complete."))
		return
	}

	switch mode {
	case "phone":
		s.cancelQRCodeFlow()
		if err := s.resetAuthMode(false); err != nil {
			writeAuthState(w, http.StatusInternalServerError, s.clientState("Could not prepare phone sign-in."))
			return
		}
	case "code":
		phone, err = normalizeAuthPhone(phone)
		if err != nil {
			writeAuthState(w, http.StatusBadRequest, s.clientState(err.Error()))
			return
		}
		s.cancelQRCodeFlow()
		if err := s.beginPhoneCode(r.Context(), phone); err != nil {
			writeAuthState(w, http.StatusBadRequest, s.clientState("Could not send the code. Check the number and try again."))
			return
		}
	case "qr":
		s.cancelQRCodeFlow()
		if err := s.resetAuthMode(true); err != nil {
			writeAuthState(w, http.StatusInternalServerError, s.clientState("Could not prepare the QR code."))
			return
		}
		s.startQRCodeFlow()
	default:
		writeAuthState(w, http.StatusBadRequest, s.clientState("Unknown sign-in method."))
		return
	}

	writeAuthState(w, http.StatusOK, s.clientState(""))
}

func (s *webAuthSession) handleSavedSession(w http.ResponseWriter, r *http.Request) {
	if !s.authorized(w, r) {
		return
	}
	action, err := parseAuthField(r, "action", trimAllSpace)
	if err != nil || action != "use_saved" {
		writeAuthState(w, http.StatusBadRequest, s.clientState("Unsupported saved-session action."))
		return
	}

	s.mu.Lock()
	if s.completed || s.doneSent {
		s.mu.Unlock()
		writeAuthState(w, http.StatusConflict, s.clientState("Sign-in is already complete."))
		return
	}
	provider := s.sessionProvider
	profile := s.sessionProfile
	available := s.savedSession
	loader := s.peerLoader
	state := *s.state
	s.mu.Unlock()
	if !available {
		writeAuthState(w, http.StatusNotFound, s.clientState("No saved session is available."))
		return
	}

	storage, err := sessionstore.Open(provider, profile)
	if err != nil {
		writeAuthState(w, http.StatusInternalServerError, s.clientState("Could not open the saved session."))
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()
	data, err := storage.LoadSession(ctx)
	if err == nil {
		if loader == nil {
			loader = loadAuthPeers
		}
		var peers []authPeer
		peers, err = loader(ctx, &state, data)
		if err == nil {
			s.cancelQRCodeFlow()
			s.mu.Lock()
			if !s.completed && !s.doneSent {
				s.completed = true
				s.sessionData = append([]byte(nil), data...)
				s.state.SetSessionData(data)
				s.peers = peers
				s.peersLoaded = true
				s.peersLoading = false
				s.peersError = ""
			}
			s.mu.Unlock()
			writeAuthState(w, http.StatusOK, s.clientState(""))
			return
		}
	}

	s.mu.Lock()
	s.savedSession = false
	s.sessionStoreError = "The saved session is no longer valid. Sign in again to replace it."
	s.mu.Unlock()
	writeAuthState(w, http.StatusUnauthorized, s.clientState("The saved session could not be verified with Telegram."))
}

func normalizeAuthPhone(value string) (string, error) {
	value = strings.TrimSpace(value)
	var digits strings.Builder
	for _, r := range value {
		if unicode.IsDigit(r) {
			digits.WriteRune(r)
			continue
		}
		if r != '+' && r != ' ' && r != '-' && r != '(' && r != ')' {
			return "", fmt.Errorf("enter a phone number in international format")
		}
	}
	if digits.Len() < 7 || digits.Len() > 15 {
		return "", fmt.Errorf("enter a phone number in international format")
	}
	return "+" + digits.String(), nil
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
	if !s.qrMode {
		writeAuthState(w, http.StatusBadRequest, s.clientState("Mock QR scan is only available in QR mode."))
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
