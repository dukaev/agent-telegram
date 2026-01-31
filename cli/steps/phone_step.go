// Package steps provides login step implementations.
package steps

import (
	"time"

	"github.com/charmbracelet/bubbletea"
	"agent-telegram/cli/components"
	"agent-telegram/internal/auth"
)

// PhoneStep represents the phone input step.
type PhoneStep struct {
	phoneInput  components.Input
	loader      components.Loader
	authService *auth.Service
}

// NewPhoneStep creates a new phone input step.
func NewPhoneStep(authService *auth.Service) PhoneStep {
	phoneInput := components.NewInput(components.PhoneType)
	phoneInput.Focus()

	return PhoneStep{
		phoneInput:  phoneInput,
		loader:      components.NewLoader(),
		authService: authService,
	}
}

// Init initializes the step.
func (m PhoneStep) Init() tea.Cmd {
	return nil
}

// Update handles messages.
//nolint:dupl // Similar to CodeStep.Update - different loader message and input type
func (m PhoneStep) Update(msg tea.Msg) (PhoneStep, tea.Cmd) {
	if cmd, ok := components.HandleQuitKeys(msg); ok {
		return m, cmd
	}

	if msg, ok := msg.(tea.KeyMsg); ok && !m.loader.IsActive() {
		if msg.String() == components.KeyEnter {
			var cmd tea.Cmd
			m.loader, cmd = m.loader.Start("Sending code...", 2*time.Second)
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

	var cmd tea.Cmd
	m.phoneInput.Model, cmd = m.phoneInput.Update(msg)
	return m, cmd
}

// Submit returns the phone number and sends verification code.
func (m PhoneStep) Submit() tea.Cmd {
	phone := m.phoneInput.Value()
	if m.authService != nil {
		return m.authService.SendCode(phone)
	}
	return func() tea.Msg {
		return PhoneSubmitted(phone)
	}
}

// HandleAuthResult handles the authentication result.
func (m PhoneStep) HandleAuthResult(result auth.Result) tea.Msg {
	if result.Error != "" {
		return AuthError{Step: "phone", Error: result.Error}
	}
	if result.Success {
		return PhoneCodeSent{Phone: m.phoneInput.Value(), CodeHash: result.PhoneCodeHash}
	}
	return PhoneSubmitted(m.phoneInput.Value())
}

// View renders the phone input step.
func (m PhoneStep) View() string {
	if m.loader.IsActive() {
		return components.RenderLoaderView(
			m.phoneInput.GetLabel(),
			m.phoneInput.ViewWithSpinner(m.loader.Frame()),
			m.loader.Frame(),
		)
	}

	inputLine := components.RenderLabeledInput(m.phoneInput.GetLabel(), m.phoneInput.View())
	return components.RenderInputView(inputLine, true)
}

// GetPhone returns the entered phone number.
func (m PhoneStep) GetPhone() string {
	return m.phoneInput.Value()
}

// PhoneSubmitted is sent when phone is submitted.
type PhoneSubmitted string
