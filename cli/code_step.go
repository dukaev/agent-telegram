// Package cli provides interactive UI components for the agent-telegram tool.
package cli

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// CodeStep represents the verification code input step.
type CodeStep struct {
	input    MaskedInput
	phone    string
	focused  bool
	quitting bool
}

// NewCodeStep creates a new code input step.
func NewCodeStep(phone string) CodeStep {
	input := NewMaskedInput(5)
	input.Focus()

	return CodeStep{
		input:   input,
		phone:   phone,
		focused: true,
	}
}

// Init initializes the step.
func (m CodeStep) Init() tea.Cmd {
	return nil
}

// Update handles messages.
func (m CodeStep) Update(msg tea.Msg) (CodeStep, tea.Cmd) {
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

// Submit returns the verification code.
func (m CodeStep) Submit() tea.Cmd {
	return func() tea.Msg {
		return CodeSubmitted(m.input.Value())
	}
}

// View renders the code input step.
func (m CodeStep) View() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("Verification Code\n\n"))
	b.WriteString(lipgloss.NewStyle().Foreground(normalColor).Render("Enter the code sent to your phone\n\n"))

	label := labelStyle.Render("Code")
	b.WriteString(label + "  " + m.input.View())

	button := "  [submit]"
	if m.focused {
		button = "  " + lipgloss.NewStyle().Foreground(telegramBlue).Render("[submit]")
	}
	b.WriteString("\n\n" + button)

	b.WriteString("\n\n" + helpStyle.Render("enter: submit • q: quit"))

	return b.String()
}

// GetPhone returns the phone number.
func (m CodeStep) GetPhone() string {
	return m.phone
}

// GetCode returns the entered code.
func (m CodeStep) GetCode() string {
	return m.input.Value()
}

// CodeSubmitted is sent when code is submitted.
type CodeSubmitted string
