// Package components provides reusable UI components.
package components

import (
	"strings"
	"unicode"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"agent-telegram/pkg/common"
)

// InputType defines the type of input field.
type InputType int

const (
	PhoneType InputType = iota
	CodeType
	PasswordType
)

// MaskedInput is a text input that displays underscores for empty positions.
type MaskedInput struct {
	input        textinput.Model
	maxLength    int
	minLength    int // minimum length to display with underscores
	focused      bool
	cursorPos    int
	focusedColor lipgloss.Color
	normalColor  lipgloss.Color
	inputType    InputType
}

// NewMaskedInput creates a new masked input with the specified length.
func NewMaskedInput(length int, inputType InputType) MaskedInput {
	ti := textinput.New()
	ti.Prompt = ""
	ti.CharLimit = length
	ti.Width = length

	// For password, hide input
	if inputType == PasswordType {
		ti.EchoMode = textinput.EchoPassword
		ti.EchoCharacter = '•'
	}

	minLen := 0
	if inputType == PasswordType {
		minLen = 5
	}

	return MaskedInput{
		input:        ti,
		maxLength:    length,
		minLength:    minLen,
		focused:      false,
		cursorPos:    0,
		focusedColor: common.TelegramBlue,
		normalColor:  common.NormalColor,
		inputType:    inputType,
	}
}

// Focus sets focus on the input.
func (m *MaskedInput) Focus() tea.Cmd {
	m.focused = true
	return m.input.Focus()
}

// Blur removes focus from the input.
func (m *MaskedInput) Blur() {
	m.focused = false
	m.input.Blur()
}

// SetFocusedColor sets the color for focused state.
func (m *MaskedInput) SetFocusedColor(color lipgloss.Color) {
	m.focusedColor = color
}

// SetNormalColor sets the color for normal state.
func (m *MaskedInput) SetNormalColor(color lipgloss.Color) {
	m.normalColor = color
}

// Update handles messages.
func (m MaskedInput) Update(msg tea.Msg) (MaskedInput, tea.Cmd) {
	// Filter input based on type
	if msg, ok := msg.(tea.KeyMsg); ok {
		msg = m.filterInput(msg)
		if msg.String() == "" {
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	// Track cursor position based on input length
	m.cursorPos = len(m.input.Value())
	return m, cmd
}

// filterInput validates input based on input type
func (m MaskedInput) filterInput(msg tea.KeyMsg) tea.KeyMsg {
	runes := []rune(msg.String())
	if len(runes) == 0 {
		return msg
	}
	r := runes[0]

	switch m.inputType {
	case PhoneType:
		// Allow: +, (, ) and digits
		if r == '+' || r == '(' || r == ')' || unicode.IsDigit(r) {
			return msg
		}
	case CodeType:
		// Allow: digits only
		if unicode.IsDigit(r) {
			return msg
		}
	case PasswordType:
		// Allow: any character
		return msg
	}

	return tea.KeyMsg{}
}

// Value returns the current input value.
func (m MaskedInput) Value() string {
	return m.input.Value()
}

// GetLabel returns the label with icon for the input type.
func (m MaskedInput) GetLabel() string {
	switch m.inputType {
	case PhoneType:
		return " Phone:"
	case CodeType:
		return " Code:"
	case PasswordType:
		return "󰦯 Password:"
	default:
		return "Input:"
	}
}

// View renders the input with the current cell highlighted.
func (m MaskedInput) View() string {
	value := m.input.Value()

	// Build display with underscores for remaining positions
	var result strings.Builder

	// Characters before cursor (with their color)
	for i := 0; i < m.cursorPos; i++ {
		if i < len(value) {
			if m.focused {
				result.WriteString(lipgloss.NewStyle().Foreground(m.focusedColor).Render(string(value[i])))
			} else {
				result.WriteString(lipgloss.NewStyle().Foreground(m.normalColor).Faint(true).Render(string(value[i])))
			}
		}
	}

	// Current cursor position (highlighted with background)
	var currentCell string
	if m.cursorPos < len(value) {
		currentCell = string(value[m.cursorPos])
	} else if m.cursorPos < m.maxLength {
		currentCell = "_"
	} else {
		// At max length, no cursor cell to show
		return result.String()
	}

	if m.focused {
		result.WriteString(lipgloss.NewStyle().
			Foreground(lipgloss.Color("white")).
			Background(m.focusedColor).
			Render(currentCell))
	} else {
		result.WriteString(lipgloss.NewStyle().Foreground(m.normalColor).Faint(true).Render(currentCell))
	}

	// Characters after cursor (only show entered characters, no underscores)
	for i := m.cursorPos + 1; i < len(value); i++ {
		result.WriteString(lipgloss.NewStyle().Foreground(m.normalColor).Faint(true).Render(string(value[i])))
	}

	return result.String()
}

// ValueOnly renders only the input value without cursor highlighting (for loading state).
func (m MaskedInput) ValueOnly() string {
	value := m.input.Value()
	var result strings.Builder

	for i := 0; i < len(value); i++ {
		result.WriteString(lipgloss.NewStyle().Foreground(m.normalColor).Faint(true).Render(string(value[i])))
	}

	return result.String()
}

// ValueOnlyColored renders the input value in the specified color.
func (m MaskedInput) ValueOnlyColored(color lipgloss.Color) string {
	value := m.input.Value()
	var result strings.Builder

	// Display entered characters
	for i := 0; i < len(value); i++ {
		result.WriteString(lipgloss.NewStyle().Foreground(color).Render(string(value[i])))
	}

	// Add underscores for minimum length
	for i := len(value); i < m.minLength; i++ {
		result.WriteString(lipgloss.NewStyle().Foreground(color).Render("_"))
	}

	return result.String()
}
