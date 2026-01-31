// Package ui provides interactive UI components for the agent-telegram tool.
package ui

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"agent-telegram/cli/steps"
	"agent-telegram/internal/auth"
	"agent-telegram/pkg/common"
)

// Step interface represents a login step.
type Step interface {
	Init() tea.Cmd
	View() string
}

// messageHandlerCheck defines a message type check and its handler.
type messageHandlerCheck struct {
	check   func(tea.Msg) bool
	handle  func(LoginModel, tea.Msg) (tea.Model, tea.Cmd)
}

// LoginModel manages the login flow.
type LoginModel struct {
	ctx            context.Context
	authService    *auth.Service
	currentStep    Step
	phone          string
	twoFAHint      string
	quitting       bool
	successMsg     string
	errorMsg       string
	sessionPath    string
	messageHandlers []messageHandlerCheck
}

// NewLoginModel creates a new login model starting with phone step.
func NewLoginModel(ctx context.Context, authService *auth.Service) LoginModel {
	m := LoginModel{
		ctx:         ctx,
		authService: authService,
		currentStep: steps.NewPhoneStep(authService),
		messageHandlers: []messageHandlerCheck{
			{
				check:  func(msg tea.Msg) bool { _, ok := msg.(auth.Result); return ok },
				handle: func(m LoginModel, msg tea.Msg) (tea.Model, tea.Cmd) { return m.handleAuthResultMsg(msg.(auth.Result)) },
			},
			{
				check:  func(msg tea.Msg) bool { _, ok := msg.(steps.PhoneCodeSent); return ok },
				handle: func(m LoginModel, msg tea.Msg) (tea.Model, tea.Cmd) { return m.handlePhoneCodeSentMsg(msg.(steps.PhoneCodeSent)) },
			},
			{
				check:  func(msg tea.Msg) bool { _, ok := msg.(steps.TwoFARequired); return ok },
				handle: func(m LoginModel, msg tea.Msg) (tea.Model, tea.Cmd) { return m.handleTwoFARequiredMsg(msg.(steps.TwoFARequired)) },
			},
			{
				check:  func(msg tea.Msg) bool { _, ok := msg.(steps.PhoneSubmitted); return ok },
				handle: func(m LoginModel, msg tea.Msg) (tea.Model, tea.Cmd) { return m.handlePhoneSubmittedMsg(msg.(steps.PhoneSubmitted)) },
			},
			{
				check:  func(msg tea.Msg) bool { _, ok := msg.(steps.CodeSubmitted); return ok },
				handle: func(m LoginModel, _ tea.Msg) (tea.Model, tea.Cmd) { return m.handleCodeSubmittedMsg() },
			},
			{
				check:  func(msg tea.Msg) bool { _, ok := msg.(steps.PasswordSubmitted); return ok },
				handle: func(m LoginModel, _ tea.Msg) (tea.Model, tea.Cmd) { return m.handlePasswordSubmittedMsg() },
			},
			{
				check:  func(msg tea.Msg) bool { _, ok := msg.(steps.AuthError); return ok },
				handle: func(m LoginModel, msg tea.Msg) (tea.Model, tea.Cmd) { return m.handleAuthErrorMsg(msg.(steps.AuthError)) },
			},
			{
				check:  func(msg tea.Msg) bool { _, ok := msg.(tea.KeyMsg); return ok },
				handle: func(m LoginModel, msg tea.Msg) (tea.Model, tea.Cmd) { return m.handleKeyMsgMsg(msg.(tea.KeyMsg)) },
			},
		},
	}
	return m
}

// Init initializes the login model.
func (m LoginModel) Init() tea.Cmd {
	if step := m.currentStep; step != nil {
		return step.Init()
	}
	return nil
}

// Update handles messages and transitions between steps.
func (m LoginModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.quitting {
		return m, tea.Quit
	}

	// Find handler for this message type
	for _, handler := range m.messageHandlers {
		if handler.check(msg) {
			return handler.handle(m, msg)
		}
	}

	// Delegate update to current step
	return m.updateCurrentStep(msg)
}

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

func (m LoginModel) handleKeyMsgMsg(keyMsg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if keyMsg.String() == "ctrl+c" || keyMsg.String() == "q" {
		m.quitting = true
		return m, tea.Quit
	}
	return m, nil
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

// View renders the current step.
func (m LoginModel) View() string {
	if m.quitting {
		if m.errorMsg != "" {
			return common.ErrorStyle.Render("✗ Error: " + m.errorMsg)
		}
		if m.successMsg != "" {
			return common.TitleStyle.Render(m.successMsg)
		}
		return ""
	}

	if m.currentStep != nil {
		return m.currentStep.View()
	}

	return ""
}

// IsSuccess returns true if login was successful.
func (m LoginModel) IsSuccess() bool {
	return m.successMsg != ""
}

// GetError returns the error message if any.
func (m LoginModel) GetError() string {
	return m.errorMsg
}

// GetPhone returns the entered phone number.
func (m LoginModel) GetPhone() string {
	return m.phone
}

// GetSessionPath returns the session file path.
func (m LoginModel) GetSessionPath() string {
	return m.sessionPath
}

// SaveEnvFile saves the phone number to a .env file in the project directory.
func SaveEnvFile(projectDir, phone string) error {
	envPath := filepath.Join(projectDir, ".env")
	content := fmt.Sprintf("# Telegram API credentials\nTELEGRAM_PHONE=%s\n", phone)
	return os.WriteFile(envPath, []byte(content), 0600)
}

// RunLoginUIWithAuth runs the interactive login UI with Telegram authentication.
// Returns session path on success, error on failure.
func RunLoginUIWithAuth(ctx context.Context, authService *auth.Service) (sessionPath string, err error) {
	p := tea.NewProgram(NewLoginModel(ctx, authService))
	m, err := p.Run()
	if err != nil {
		return "", fmt.Errorf("could not start program: %w", err)
	}

	if m, ok := m.(LoginModel); ok {
		if m.GetError() != "" {
			return "", fmt.Errorf("authentication failed: %s", m.GetError())
		}
		if m.IsSuccess() {
			return m.GetSessionPath(), nil
		}
		return "", fmt.Errorf("authentication incomplete")
	}

	return "", fmt.Errorf("unexpected model type")
}
