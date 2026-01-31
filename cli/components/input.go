// Package components provides reusable UI components.
package components

import (
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
	"agent-telegram/pkg/common"
)

// InputType defines the type of input field.
type InputType int

const (
	// PhoneType is a phone number input.
	PhoneType InputType = iota
	// CodeType is a verification code input.
	CodeType
	// PasswordType is a 2FA password input.
	PasswordType
)

// InputBorderStyle is the border style for focused input
var InputBorderStyle = lipgloss.NewStyle().
	Border(lipgloss.RoundedBorder()).
	BorderForeground(common.TelegramBlue).
	Padding(0, 1)

// Input is a styled text input component.
type Input struct {
	textinput.Model
	inputType InputType
}

// NewInput creates a new styled input.
func NewInput(inputType InputType) Input {
	ti := textinput.New()
	ti.TextStyle = common.InputStyle
	ti.PlaceholderStyle = common.InputPlaceholderStyle

	switch inputType {
	case PhoneType:
		ti.Placeholder = "+1 234 567 8900"
		ti.CharLimit = 20
		ti.Width = 30
	case CodeType:
		ti.Placeholder = "12345"
		ti.CharLimit = 5
		ti.Width = 10
	case PasswordType:
		ti.Placeholder = "Your 2FA password"
		ti.EchoMode = textinput.EchoPassword
		ti.EchoCharacter = '•'
	}

	return Input{
		Model:     ti,
		inputType: inputType,
	}
}

// View renders the input with border.
func (m Input) View() string {
	icon := lipgloss.NewStyle().Foreground(common.TelegramBlue).Render(m.GetIcon())
	content := icon + " " + m.Model.View()

	if m.Focused() {
		return InputBorderStyle.Render(content)
	}
	return content
}

// ViewWithSpinner renders the input with spinner icon instead of cursor.
func (m Input) ViewWithSpinner(spinnerFrame string) string {
	value := m.Value()
	if value == "" {
		value = m.Placeholder
	}

	spinner := lipgloss.NewStyle().Foreground(common.TelegramBlue).Render(spinnerFrame)
	content := spinner + " " + value

	return InputBorderStyle.Render(content)
}

// GetIcon returns the icon for the input type.
func (m Input) GetIcon() string {
	switch m.inputType {
	case PhoneType:
		return ""
	case CodeType:
		return ""
	case PasswordType:
		return "󰦯"
	default:
		return ">"
	}
}

// GetLabel returns the label for the input type.
func (m Input) GetLabel() string {
	switch m.inputType {
	case PhoneType:
		return "Phone:"
	case CodeType:
		return "Code:"
	case PasswordType:
		return "Password:"
	default:
		return "Input:"
	}
}
