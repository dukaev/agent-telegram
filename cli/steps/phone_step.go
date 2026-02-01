// Package steps provides login step implementations.
package steps

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"agent-telegram/cli/components"
	"agent-telegram/internal/auth"
)

// PhoneStep represents the phone input step.
type PhoneStep struct {
	phoneInput  components.Input
	loader      components.Loader
	authService *auth.Service
	errorMsg    string
	waiting     bool // true while waiting for API response
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

	// Handle Enter key - start loading and submit immediately
	if msg, ok := msg.(tea.KeyMsg); ok && !m.waiting {
		if msg.String() == components.KeyEnter {
			m.waiting = true
			m.errorMsg = ""
			var loaderCmd tea.Cmd
			// Start loader (will be stopped when response comes)
			m.loader, loaderCmd = m.loader.Start("Connecting to Telegram...", 10*time.Second)
			// Submit immediately and batch with loader animation
			return m, tea.Batch(loaderCmd, m.Submit())
		}
	}

	// Keep loader animating while waiting
	if _, ok := msg.(components.TickMsg); ok {
		if m.waiting && m.loader.IsActive() {
			var cmd tea.Cmd
			m.loader, cmd = m.loader.Update()
			return m, cmd
		}
		return m, nil
	}

	// Clear error when user types (only if not waiting)
	if _, ok := msg.(tea.KeyMsg); ok && !m.waiting {
		m.errorMsg = ""
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
	// Show loader while waiting for API response
	if m.waiting && m.loader.IsActive() {
		return components.RenderLoaderView(
			m.phoneInput.GetLabel(),
			m.phoneInput.ViewWithSpinner(m.loader.Frame()),
		)
	}

	// Show input with optional error
	inputLine := components.RenderLabeledInput(m.phoneInput.GetLabel(), m.phoneInput.View())
	return components.RenderInputViewWithError(inputLine, m.errorMsg, true)
}

// SetError sets an error message to display and stops waiting.
func (m PhoneStep) SetError(err string) PhoneStep {
	m.errorMsg = err
	m.waiting = false
	m.loader = components.NewLoader() // Reset loader
	return m
}

// GetError returns the current error message.
func (m PhoneStep) GetError() string {
	return m.errorMsg
}

// GetPhone returns the entered phone number.
func (m PhoneStep) GetPhone() string {
	return m.phoneInput.Value()
}

// PhoneSubmitted is sent when phone is submitted.
type PhoneSubmitted string
