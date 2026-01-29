// Package cli provides interactive UI components for the agent-telegram tool.
package cli

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// MaskedInput is a text input that displays underscores for empty positions.
type MaskedInput struct {
	input       textinput.Model
	placeholder string
	maxLength   int
	focused     bool
	cursorPos   int
	focusedColor lipgloss.Color
	normalColor  lipgloss.Color
}

// NewMaskedInput creates a new masked input with the specified length.
func NewMaskedInput(length int) MaskedInput {
	ti := textinput.New()
	ti.Prompt = ""
	ti.CharLimit = length
	ti.Width = length

	return MaskedInput{
		input:        ti,
		placeholder:  strings.Repeat("_", length),
		maxLength:    length,
		focused:      false,
		cursorPos:    0,
		focusedColor: telegramBlue,
		normalColor:  normalColor,
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
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	// Track cursor position based on input length
	m.cursorPos = len(m.input.Value())
	return m, cmd
}

// Value returns the current input value.
func (m MaskedInput) Value() string {
	return m.input.Value()
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
	} else {
		currentCell = "_"
	}

	if m.focused {
		result.WriteString(lipgloss.NewStyle().
			Foreground(lipgloss.Color("white")).
			Background(m.focusedColor).
			Render(currentCell))
	} else {
		result.WriteString(lipgloss.NewStyle().Foreground(m.normalColor).Faint(true).Render(currentCell))
	}

	// Characters after cursor (faint)
	for i := m.cursorPos + 1; i < m.maxLength; i++ {
		if i < len(value) {
			result.WriteString(lipgloss.NewStyle().Foreground(m.normalColor).Faint(true).Render(string(value[i])))
		} else {
			result.WriteString(lipgloss.NewStyle().Foreground(m.normalColor).Faint(true).Render("_"))
		}
	}

	return result.String()
}
