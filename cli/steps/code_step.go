// Package steps provides login step implementations.
package steps

import (
	"time"

	"github.com/charmbracelet/bubbletea"
	"agent-telegram/cli/components"
	"agent-telegram/internal/auth"
)

// CodeStep represents the verification code input step.
type CodeStep struct {
	codeInput   components.Input
	loader      components.Loader
	authService *auth.Service
	errorMsg    string
}

// NewCodeStep creates a new code input step.
func NewCodeStep(authService *auth.Service) CodeStep {
	codeInput := components.NewInput(components.CodeType)
	codeInput.Focus()

	return CodeStep{
		codeInput:   codeInput,
		loader:      components.NewVerifyLoader(),
		authService: authService,
	}
}

// Init initializes the step.
func (m CodeStep) Init() tea.Cmd {
	return nil
}

// Update handles messages.
//nolint:dupl // Similar to PhoneStep.Update - different loader message and input type
func (m CodeStep) Update(msg tea.Msg) (CodeStep, tea.Cmd) {
	if cmd, ok := components.HandleQuitKeys(msg); ok {
		return m, cmd
	}

	if msg, ok := msg.(tea.KeyMsg); ok && !m.loader.IsActive() {
		if msg.String() == components.KeyEnter {
			var cmd tea.Cmd
			m.loader, cmd = m.loader.Start("Verifying code...", 2*time.Second)
			return m, cmd
		}
	}

	if _, ok := msg.(components.TickMsg); ok {
		if m.loader.IsActive() {
			var cmd tea.Cmd
			m.loader, cmd = m.loader.Update()
			if !m.loader.IsActive() {
				return m, m.Submit()
			}
			return m, cmd
		}
		return m, nil
	}

	// Clear error when user types
	if _, ok := msg.(tea.KeyMsg); ok {
		m.errorMsg = ""
	}

	var cmd tea.Cmd
	m.codeInput.Model, cmd = m.codeInput.Update(msg)
	return m, cmd
}

// Submit returns the verification code and signs in.
func (m CodeStep) Submit() tea.Cmd {
	code := m.codeInput.Value()
	if m.authService != nil {
		return m.authService.SignIn(code)
	}
	return func() tea.Msg {
		return CodeSubmitted(code)
	}
}

// HandleAuthResult handles the authentication result.
func (m CodeStep) HandleAuthResult(result auth.Result) tea.Msg {
	if result.Error != "" {
		return AuthError{Step: "code", Error: result.Error}
	}
	if result.Requires2FA {
		return TwoFARequired{Hint: result.TwoFactorHint}
	}
	if result.Success {
		return CodeSubmitted(m.codeInput.Value())
	}
	return AuthError{Step: "code", Error: "Authentication failed"}
}

// View renders the code input step.
func (m CodeStep) View() string {
	if m.loader.IsActive() {
		return components.RenderLoaderView(
			m.codeInput.GetLabel(),
			m.codeInput.ViewWithSpinner(m.loader.Frame()),
		)
	}

	inputLine := components.RenderLabeledInput(m.codeInput.GetLabel(), m.codeInput.View())
	return components.RenderInputViewWithError(inputLine, m.errorMsg, true)
}

// SetError sets an error message to display.
func (m CodeStep) SetError(err string) CodeStep {
	m.errorMsg = err
	return m
}

// GetError returns the current error message.
func (m CodeStep) GetError() string {
	return m.errorMsg
}

// GetCode returns the entered code.
func (m CodeStep) GetCode() string {
	return m.codeInput.Value()
}

// CodeSubmitted is sent when code is submitted.
type CodeSubmitted string
