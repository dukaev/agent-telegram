package auth

import (
	"time"

	"agent-telegram/internal/config"
	"agent-telegram/internal/policy"
)

type webAuthSnapshot struct {
	completed         bool
	doneSent          bool
	qrImage           string
	qrTokenURL        string
	qrExpires         time.Time
	policy            policy.Policy
	mock              bool
	appID             int
	appHash           string
	sessionProvider   string
	sessionProfile    string
	sessionPersistent bool
	savedSession      bool
	sessionStoreError string
}

func (s *webAuthSession) clientState(errMsg string) authClientState {
	snapshot := s.snapshot()
	data := snapshot.baseClientState(errMsg)
	if snapshot.completed {
		return snapshot.completedClientState(data)
	}
	return snapshot.qrClientState(data)
}

func (s *webAuthSession) snapshot() webAuthSnapshot {
	s.mu.Lock()
	defer s.mu.Unlock()

	return webAuthSnapshot{
		completed:         s.completed,
		doneSent:          s.doneSent,
		qrImage:           s.qrImage,
		qrTokenURL:        s.qrTokenURL,
		qrExpires:         s.qrExpires,
		policy:            s.policy,
		mock:              s.runtime.WebMock,
		appID:             s.state.AppID,
		appHash:           s.state.AppHash,
		sessionProvider:   s.sessionProvider,
		sessionProfile:    s.sessionProfile,
		sessionPersistent: s.sessionPersistent,
		savedSession:      s.savedSession,
		sessionStoreError: s.sessionStoreError,
	}
}

func (s webAuthSnapshot) baseClientState(errMsg string) authClientState {
	currentPolicy := s.policy
	if currentPolicy.Version == 0 {
		currentPolicy = policy.Default()
	}
	state := authClientState{
		Error:     errMsg,
		Completed: s.completed,
		API: authAPIState{
			AppID:   s.appID,
			Default: isDefaultAPISettings(s.appID, s.appHash),
			CanEdit: !s.completed && !s.doneSent,
		},
		Policy: currentPolicy,
	}
	if s.mock {
		state.Mock = &authMockInfo{Enabled: true}
	}
	if s.sessionProvider != "" {
		state.Session = &authSessionInfo{
			Provider:      s.sessionProvider,
			Profile:       s.sessionProfile,
			Persistent:    s.sessionPersistent,
			Available:     s.savedSession,
			SaveByDefault: s.sessionPersistent,
			Error:         s.sessionStoreError,
		}
	}
	return state
}

func (s webAuthSnapshot) completedClientState(data authClientState) authClientState {
	if s.doneSent {
		data.Title = "You're all set"
		data.Message = "Your settings have been saved."
		data.Mode = "done"
		return data
	}
	data.Title = "Set up access"
	data.Message = "Choose the chats your agent can work with."
	data.Mode = "setup"
	return data
}

func (s webAuthSnapshot) qrClientState(data authClientState) authClientState {
	data.Mode = "qr"
	data.Refresh = 1
	data.Message = "Scan the code with the Telegram app."
	if s.qrImage == "" {
		data.Title = "Sign in to Telegram"
		return data
	}
	data.Title = "Sign in to Telegram"
	data.QRImage = s.qrImage
	data.QRLink = s.qrTokenURL
	data.Refresh = qrRefreshDelay(s.qrExpires)
	if !s.qrExpires.IsZero() {
		data.Expires = s.qrExpires.Format(time.RFC3339)
	}
	return data
}

func isDefaultAPISettings(appID int, appHash string) bool {
	defaultID, err := config.ParseAppID(defaultAppID)
	if err != nil {
		return false
	}
	return appID == defaultID && appHash == defaultAppHash
}
