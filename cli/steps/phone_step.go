// Package steps provides login step implementations.
package steps

import (
	"time"

	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"agent-telegram/cli/components"
	"agent-telegram/pkg/common"
)

// PhoneStep represents the phone input step.
type PhoneStep struct {
	phoneInput components.MaskedInput
	focusIndex int // 0 = phone input, 1 = button
	loader     components.Loader
}

// NewPhoneStep creates a new phone input step.
func NewPhoneStep() PhoneStep {
	phoneInput := components.NewMaskedInput(12, components.PhoneType)
	phoneInput.Focus()

	return PhoneStep{
		phoneInput: phoneInput,
		focusIndex: 0,
		loader:     components.NewLoader(),
	}
}

// Init initializes the step.
func (m PhoneStep) Init() tea.Cmd {
	return nil
}

// Update handles messages.
func (m PhoneStep) Update(msg tea.Msg) (PhoneStep, tea.Cmd) {
	if cmd, ok := components.HandleQuitKeys(msg); ok {
		return m, cmd
	}

	if msg, ok := msg.(tea.KeyMsg); ok && !m.loader.IsActive() {
		switch msg.String() {
		case "enter":
			if m.focusIndex == 0 {
				m.focusIndex = 1
				m.phoneInput.Blur()
				return m, nil
			}
			var cmd tea.Cmd
			m.loader, cmd = m.loader.Start("Sending code...", 2*time.Second)
			return m, cmd
		case "tab", "shift+tab":
			if m.focusIndex == 0 {
				m.focusIndex = 1
				m.phoneInput.Blur()
			} else {
				m.focusIndex = 0
				m.phoneInput.Focus()
			}
			return m, nil
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
	m.phoneInput, cmd = m.phoneInput.Update(msg)
	return m, cmd
}

// Submit returns the phone number.
func (m PhoneStep) Submit() tea.Cmd {
	return func() tea.Msg {
		return PhoneSubmitted(m.phoneInput.Value())
	}
}

// View renders the phone input step.
func (m PhoneStep) View() string {
	if m.loader.IsActive() {
		return components.RenderLoaderView(
			m.phoneInput.GetLabel(),
			m.phoneInput.ValueOnlyColored(common.TelegramBlue),
			m.loader.Frame(),
		)
	}

	inputLine := components.RenderLabeledInput(m.phoneInput.GetLabel(), m.phoneInput.View())

	button := "[send code]"
	if m.focusIndex == 1 {
		button = lipgloss.NewStyle().Foreground(common.TelegramBlue).Render("[send code]")
	}

	return components.RenderInputView(inputLine+"\n"+button, false)
}

// GetPhone returns the entered phone number.
func (m PhoneStep) GetPhone() string {
	return m.phoneInput.Value()
}

// PhoneSubmitted is sent when phone is submitted.
type PhoneSubmitted string
