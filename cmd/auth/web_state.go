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
	mock        bool
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
		mock:        s.runtime.WebMock,
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
	state := authClientState{
		Error:     errMsg,
		Completed: s.completed,
		API: authAPIState{
			AppID:   s.appID,
			Default: isDefaultAPISettings(s.appID, s.appHash),
			CanEdit: !s.completed && !s.doneSent && (s.qrMode || s.phone == ""),
		},
		Policy: currentPolicy,
	}
	if s.mock {
		state.Mock = &authMockInfo{Enabled: true, Code: mockCode, Password: mockPassword}
	}
	return state
}

func (s webAuthSnapshot) completedClientState(data authClientState) authClientState {
	if s.doneSent {
		data.Title = "Всё готово"
		data.Message = "Настройки сохранены."
		data.Mode = "done"
		return data
	}
	data.Title = "Настрой доступ"
	data.Message = "Выбери диалоги, с которыми сможет работать агент."
	data.Mode = "setup"
	return data
}

func (s webAuthSnapshot) qrClientState(data authClientState) authClientState {
	data.Mode = "qr"
	data.Refresh = 1
	data.Message = "Отсканируй код в приложении Telegram."
	if s.qrImage == "" {
		data.Title = "Войдите в Telegram"
		return data
	}
	data.Title = "Войдите в Telegram"
	data.QRImage = s.qrImage
	data.QRLink = s.qrTokenURL
	data.Refresh = qrRefreshDelay(s.qrExpires)
	if !s.qrExpires.IsZero() {
		data.Expires = s.qrExpires.Format(time.RFC3339)
	}
	return data
}

func (s webAuthSnapshot) passwordClientState(data authClientState) authClientState {
	data.Title = "Введите пароль"
	data.Mode = "password"
	data.Hint = "Введи пароль двухэтапной аутентификации Telegram."
	if s.hint != "" {
		data.Hint = "Подсказка: " + s.hint
	}
	return data
}

func (s webAuthSnapshot) codeClientState(data authClientState) authClientState {
	data.Title = "Войдите по номеру"
	data.Phone = maskPhone(s.phone)
	if s.phone == "" {
		data.Mode = "phone"
		data.Hint = "Укажи номер, к которому привязан аккаунт Telegram."
		return data
	}
	data.Mode = "code"
	data.Title = "Введите код"
	data.Hint = "Telegram отправил код для " + data.Phone + "."
	return data
}

func isDefaultAPISettings(appID int, appHash string) bool {
	defaultID, err := config.ParseAppID(defaultAppID)
	if err != nil {
		return false
	}
	return appID == defaultID && appHash == defaultAppHash
}
