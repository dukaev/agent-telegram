// Package ui provides login message handlers.
package ui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"agent-telegram/cli/steps"
	"agent-telegram/internal/auth"
)

// Message handler functions
func (m LoginModel) handleAuthResultMsg(result auth.Result) (tea.Model, tea.Cmd) {
	if result.Error != "" {
		m.errorMsg = result.Error
		m.quitting = true
		return m, tea.Quit
	}

	switch m.currentStep.(type) {
	case steps.PhoneStep:
		return m.transitionFromPhone(result)
	case steps.CodeStep:
		return m.transitionFromCode(result)
	case steps.PasswordStep:
		return m.transitionFromPassword(result)
	}

	return m, nil
}

func (m LoginModel) handlePhoneCodeSentMsg(sent steps.PhoneCodeSent) (tea.Model, tea.Cmd) {
	m.phone = sent.Phone
	m.currentStep = steps.NewCodeStep(m.authService)
	return m, m.currentStep.Init()
}

func (m LoginModel) handleTwoFARequiredMsg(required steps.TwoFARequired) (tea.Model, tea.Cmd) {
	m.twoFAHint = required.Hint
	m.currentStep = steps.NewPasswordStep(m.authService, required.Hint)
	return m, m.currentStep.Init()
}

func (m LoginModel) handlePhoneSubmittedMsg(submitted steps.PhoneSubmitted) (tea.Model, tea.Cmd) {
	m.phone = string(submitted)
	m.currentStep = steps.NewCodeStep(m.authService)
	return m, m.currentStep.Init()
}

func (m LoginModel) handleCodeSubmittedMsg() (tea.Model, tea.Cmd) {
	m.finishSuccess()
	return m, tea.Quit
}

func (m LoginModel) handlePasswordSubmittedMsg() (tea.Model, tea.Cmd) {
	m.finishSuccess()
	return m, tea.Quit
}

func (m LoginModel) handleAuthErrorMsg(err steps.AuthError) (tea.Model, tea.Cmd) {
	m.errorMsg = err.Error
	m.quitting = true
	return m, tea.Quit
}

// transitionFromPhone handles transition from phone step.
func (m LoginModel) transitionFromPhone(msg auth.Result) (tea.Model, tea.Cmd) {
	if msg.Success {
		if m.authService != nil {
			m.phone = m.authService.GetPhoneNumber()
		}
		m.currentStep = steps.NewCodeStep(m.authService)
		return m, m.currentStep.Init()
	}
	return m, nil
}

// transitionFromCode handles transition from code step.
func (m LoginModel) transitionFromCode(msg auth.Result) (tea.Model, tea.Cmd) {
	if msg.Requires2FA {
		m.twoFAHint = msg.TwoFactorHint
		m.currentStep = steps.NewPasswordStep(m.authService, msg.TwoFactorHint)
		return m, m.currentStep.Init()
	}
	if msg.Success {
		m.finishSuccess()
		return m, tea.Quit
	}
	return m, nil
}

// transitionFromPassword handles transition from password step.
func (m LoginModel) transitionFromPassword(msg auth.Result) (tea.Model, tea.Cmd) {
	if msg.Success {
		m.finishSuccess()
		return m, tea.Quit
	}
	return m, nil
}

// updateCurrentStep delegates update to current step.
func (m LoginModel) updateCurrentStep(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch step := m.currentStep.(type) {
	case steps.PhoneStep:
		m.currentStep, cmd = step.Update(msg)
	case steps.CodeStep:
		m.currentStep, cmd = step.Update(msg)
	case steps.PasswordStep:
		m.currentStep, cmd = step.Update(msg)
	}
	return m, cmd
}

// finishSuccess completes the login flow successfully.
func (m *LoginModel) finishSuccess() {
	if m.authService != nil {
		m.sessionPath = m.authService.GetSessionPath()
	} else {
		m.sessionPath = "(mock mode - no session saved)"
	}
	m.quitting = true
	m.successMsg = fmt.Sprintf("✓ Login successful as user %s", m.phone)
}
