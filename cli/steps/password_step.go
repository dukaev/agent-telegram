// Package steps provides login step implementations.
package steps

import (
	"github.com/charmbracelet/bubbletea"
	"agent-telegram/cli/components"
	"agent-telegram/internal/auth"
)

// PasswordStep represents the 2FA password input step.
type PasswordStep struct {
	passwordInput components.Input
	authService   *auth.Service
	twoFAHint     string
	errorMsg      string
}

// NewPasswordStep creates a new password input step.
func NewPasswordStep(authService *auth.Service, hint string) PasswordStep {
	passwordInput := components.NewInput(components.PasswordType)
	passwordInput.Focus()

	return PasswordStep{
		passwordInput: passwordInput,
		authService:   authService,
		twoFAHint:     hint,
	}
}

// Init initializes the step.
func (m PasswordStep) Init() tea.Cmd {
	return nil
}

// Update handles messages.
func (m PasswordStep) Update(msg tea.Msg) (PasswordStep, tea.Cmd) {
	if cmd, ok := components.HandleQuitKeys(msg); ok {
		return m, cmd
	}

	if msg, ok := msg.(tea.KeyMsg); ok {
		if msg.String() == components.KeyEnter {
			return m, m.Submit()
		}
	}

	// Clear error when user types
	if _, ok := msg.(tea.KeyMsg); ok {
		m.errorMsg = ""
	}

	var cmd tea.Cmd
	m.passwordInput.Model, cmd = m.passwordInput.Update(msg)
	return m, cmd
}

// Submit returns the password and signs in with 2FA.
func (m PasswordStep) Submit() tea.Cmd {
	password := m.passwordInput.Value()
	if m.authService != nil {
		return m.authService.SignInWith2FA(password)
	}
	return func() tea.Msg {
		return PasswordSubmitted(password)
	}
}

// HandleAuthResult handles the authentication result.
func (m PasswordStep) HandleAuthResult(result auth.Result) tea.Msg {
	if result.Error != "" {
		return AuthError{Step: "password", Error: result.Error}
	}
	if result.Success {
		return PasswordSubmitted(m.passwordInput.Value())
	}
	return AuthError{Step: "password", Error: "2FA authentication failed"}
}

// View renders the password input step.
func (m PasswordStep) View() string {
	inputLine := components.RenderLabeledInput(m.passwordInput.GetLabel(), m.passwordInput.View())
	return components.RenderInputViewWithError(inputLine, m.errorMsg, true)
}

// SetError sets an error message to display.
func (m PasswordStep) SetError(err string) PasswordStep {
	m.errorMsg = err
	return m
}

// GetError returns the current error message.
func (m PasswordStep) GetError() string {
	return m.errorMsg
}

// GetPassword returns the entered password.
func (m PasswordStep) GetPassword() string {
	return m.passwordInput.Value()
}

// PasswordSubmitted is sent when password is submitted.
type PasswordSubmitted string
