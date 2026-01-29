// Package cli provides interactive UI components for the agent-telegram tool.
package cli

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// PhoneStep represents the phone input step.
type PhoneStep struct {
	input    MaskedInput
	focused  bool
	quitting bool
}

// NewPhoneStep creates a new phone input step.
func NewPhoneStep() PhoneStep {
	input := NewMaskedInput(12)
	input.Focus()

	return PhoneStep{
		input:   input,
		focused: true,
	}
}

// Init initializes the step.
func (m PhoneStep) Init() tea.Cmd {
	return nil
}

// Update handles messages.
func (m PhoneStep) Update(msg tea.Msg) (PhoneStep, tea.Cmd) {
	if msg, ok := msg.(tea.KeyMsg); ok {
		switch msg.String() {
		case "ctrl+c", "esc", "q":
			m.quitting = true
			return m, tea.Quit
		case "enter":
			return m, m.Submit()
		}
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

// Submit returns the phone number.
func (m PhoneStep) Submit() tea.Cmd {
	return func() tea.Msg {
		return PhoneSubmitted(m.input.Value())
	}
}

// View renders the phone input step.
func (m PhoneStep) View() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("Telegram Login"))
	b.WriteString("\n\n")

	b.WriteString(labelStyle.Render("Phone:"))
	b.WriteString(" ")
	b.WriteString(m.input.View())
	b.WriteString("\n\n")

	button := "  [send code]"
	if m.focused {
		button = "  " + lipgloss.NewStyle().Foreground(telegramBlue).Render("[send code]")
	}
	b.WriteString(button)

	return b.String()
}

// GetPhone returns the entered phone number.
func (m PhoneStep) GetPhone() string {
	return m.input.Value()
}

// PhoneSubmitted is sent when phone is submitted.
type PhoneSubmitted string
