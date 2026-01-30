// Package steps provides login step implementations.
package steps

import (
	"github.com/charmbracelet/bubbletea"
	"agent-telegram/cli/components"
)

// PasswordStep represents the 2FA password input step.
type PasswordStep struct {
	passwordInput components.MaskedInput
}

// NewPasswordStep creates a new password input step.
func NewPasswordStep() PasswordStep {
	passwordInput := components.NewMaskedInput(0, components.PasswordType)
	passwordInput.Focus()

	return PasswordStep{
		passwordInput: passwordInput,
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
		switch msg.String() {
		case "enter":
			return m, m.Submit()
		}
	}

	var cmd tea.Cmd
	m.passwordInput, cmd = m.passwordInput.Update(msg)
	return m, cmd
}

// Submit returns the password.
func (m PasswordStep) Submit() tea.Cmd {
	return func() tea.Msg {
		return PasswordSubmitted(m.passwordInput.Value())
	}
}

// View renders the password input step.
func (m PasswordStep) View() string {
	inputLine := components.RenderLabeledInput(m.passwordInput.GetLabel(), m.passwordInput.View())
	return components.RenderInputView(inputLine, true)
}

// GetPassword returns the entered password.
func (m PasswordStep) GetPassword() string {
	return m.passwordInput.Value()
}

// PasswordSubmitted is sent when password is submitted.
type PasswordSubmitted string
