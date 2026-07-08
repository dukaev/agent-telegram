package auth

import (
	"time"

	"agent-telegram/internal/config"
	"agent-telegram/internal/policy"
)

type webAuthSnapshot struct {
	completed   bool
	doneSent    bool
	qrMode      bool
	qrImage     string
	qrTokenURL  string
	qrExpires   time.Time
	policy      policy.Policy
	requires2FA bool
	hint        string
	phone       string
	appID       int
	appHash     string
}

func (s *webAuthSession) clientState(errMsg string) authClientState {
	snapshot := s.snapshot()
	data := snapshot.baseClientState(errMsg)
	if snapshot.completed {
		return snapshot.completedClientState(data)
	}
	if snapshot.qrMode {
		return snapshot.qrClientState(data)
	}
	if snapshot.requires2FA {
		return snapshot.passwordClientState(data)
	}
	return snapshot.codeClientState(data)
}

func (s *webAuthSession) snapshot() webAuthSnapshot {
	s.mu.Lock()
	defer s.mu.Unlock()

	return webAuthSnapshot{
		completed:   s.completed,
		doneSent:    s.doneSent,
		qrMode:      s.qrMode,
		qrImage:     s.qrImage,
		qrTokenURL:  s.qrTokenURL,
		qrExpires:   s.qrExpires,
		policy:      s.policy,
		requires2FA: s.state.Requires2FA,
		hint:        s.state.TwoFactorHint,
		phone:       s.state.Phone,
		appID:       s.state.AppID,
		appHash:     s.state.AppHash,
	}
}

func (s webAuthSnapshot) baseClientState(errMsg string) authClientState {
	currentPolicy := s.policy
	if currentPolicy.Version == 0 {
		currentPolicy = policy.Default()
	}
	return authClientState{
		Error:     errMsg,
		Completed: s.completed,
		API: authAPIState{
			AppID:   s.appID,
			Default: isDefaultAPISettings(s.appID, s.appHash),
			CanEdit: s.qrMode && !s.completed && !s.doneSent,
		},
		Policy: currentPolicy,
	}
}

func (s webAuthSnapshot) completedClientState(data authClientState) authClientState {
	if s.doneSent {
		data.Title = "Вход завершен"
		data.Message = "Настройки сохранены."
		data.Mode = "done"
		return data
	}
	data.Title = "Вход выполнен"
	data.Message = "Выбери, с кем агент может взаимодействовать."
	data.Mode = "setup"
	return data
}

func (s webAuthSnapshot) qrClientState(data authClientState) authClientState {
	data.Mode = "qr"
	data.Refresh = 1
	data.Message = "Отсканируй код в Telegram."
	if s.qrImage == "" {
		data.Title = "Готовлю QR-код"
		return data
	}
	data.Title = "Вход по QR-коду"
	data.QRImage = s.qrImage
	data.QRLink = s.qrTokenURL
	data.Refresh = qrRefreshDelay(s.qrExpires)
	if !s.qrExpires.IsZero() {
		data.Expires = s.qrExpires.Format(time.RFC3339)
	}
	return data
}

func (s webAuthSnapshot) passwordClientState(data authClientState) authClientState {
	data.Title = "Two-step verification"
	data.Mode = "password"
	data.Hint = "Enter your Telegram 2FA password."
	if s.hint != "" {
		data.Hint = "2FA hint: " + s.hint
	}
	return data
}

func (s webAuthSnapshot) codeClientState(data authClientState) authClientState {
	data.Title = "Telegram login"
	data.Mode = "code"
	data.Phone = maskPhone(s.phone)
	data.Hint = "Enter the code Telegram sent for " + data.Phone + "."
	return data
}

func isDefaultAPISettings(appID int, appHash string) bool {
	defaultID, err := config.ParseAppID(defaultAppID)
	if err != nil {
		return false
	}
	return appID == defaultID && appHash == defaultAppHash
}
