// Package cli provides interactive UI components for the agent-telegram tool.
package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	// Telegram brand colors
	// Primary: Telegram Blue (#0088cc / #2AABEE)
	// Neutral: Gray tones
	telegramBlue = lipgloss.Color("#0088cc")  // Telegram signature blue
	normalColor  = lipgloss.Color("#6c6c6c")  // Medium gray for secondary text

	// Telegram brand styles
	titleStyle = lipgloss.NewStyle().
			Foreground(telegramBlue).
			Bold(true)

	labelStyle = lipgloss.NewStyle().
			Foreground(normalColor)

	helpStyle = lipgloss.NewStyle().
			Foreground(normalColor)
)

type step int

const (
	stepPhone step = iota
	stepCode
)

type model struct {
	step       step
	focusIndex int
	inputs     []textinput.Model
	quitting   bool
	phone      string // saved phone number
}

func initialModel() model {
	m := model{
		step:   stepPhone,
		inputs: make([]textinput.Model, 1),
	}

	t := textinput.New()
	t.Prompt = ""
	t.Placeholder = "+1234567890"
	t.Width = 15
	t.Focus()
	m.inputs[0] = t

	return m
}

func codeModel(phone string) model {
	m := model{
		step:   stepCode,
		inputs: make([]textinput.Model, 1),
		phone:  phone,
	}

	t := textinput.New()
	t.Prompt = ""
	t.Placeholder = "12345"
	t.CharLimit = 5
	t.Width = 5
	t.Focus()

	m.inputs[0] = t

	return m
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if msg, ok := msg.(tea.KeyMsg); ok {
		switch msg.String() {
		case "ctrl+c", "esc", "q":
			m.quitting = true
			return m, tea.Quit

		case "tab", "shift+tab", "enter", "up", "down":
			return m.handleNavigation(msg.String())

		default:
			// Continue to update inputs
		}
	}

	cmd := m.updateInputs(msg)
	return m, cmd
}

func (m model) handleNavigation(key string) (tea.Model, tea.Cmd) {
	// Handle submit on enter
	if key == "enter" && m.focusIndex == len(m.inputs) {
		return m.handleSubmit()
	}

	// Navigate focus
	m.updateFocus(key)

	// Update input focus states
	cmds := make([]tea.Cmd, len(m.inputs))
	for i := range m.inputs {
		if i == m.focusIndex {
			cmds[i] = m.inputs[i].Focus()
		} else {
			m.inputs[i].Blur()
		}
	}

	return m, tea.Batch(cmds...)
}

func (m model) handleSubmit() (tea.Model, tea.Cmd) {
	if m.step == stepPhone {
		phone := m.inputs[0].Value()
		return codeModel(phone), nil
	}
	m.quitting = true
	return m, tea.Quit
}

func (m *model) updateFocus(key string) {
	if key == "up" || key == "shift+tab" {
		m.focusIndex--
	} else {
		m.focusIndex++
	}

	if m.focusIndex > len(m.inputs) {
		m.focusIndex = 0
	} else if m.focusIndex < 0 {
		m.focusIndex = len(m.inputs)
	}
}

func (m *model) updateInputs(msg tea.Msg) tea.Cmd {
	cmds := make([]tea.Cmd, len(m.inputs))
	for i := range m.inputs {
		m.inputs[i], cmds[i] = m.inputs[i].Update(msg)
	}
	return tea.Batch(cmds...)
}

func (m model) View() string {
	if m.quitting {
		return ""
	}

	if m.step == stepPhone {
		return m.phoneView()
	}
	return m.codeView()
}

func (m model) phoneView() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("Telegram Login"))
	b.WriteString("  ")

	b.WriteString(labelStyle.Render("Phone:"))
	b.WriteString(" ")

	input := m.renderInput(0)
	b.WriteString(input)

	button := m.renderButton("[send code]", telegramBlue)
	b.WriteString(button)

	return b.String()
}

func (m model) codeView() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("Verification Code\n\n"))
	b.WriteString(lipgloss.NewStyle().Foreground(normalColor).Render("Enter the code sent to your phone\n\n"))

	label := labelStyle.Render("Code")
	input := m.inputs[0].View()

	if m.focusIndex == len(m.inputs) {
		b.WriteString(label + "  " + lipgloss.NewStyle().Faint(true).Render(input))
		b.WriteString("\n\n  [submit]")
	} else {
		b.WriteString(label)
		b.WriteString("  ")
		b.WriteString(lipgloss.NewStyle().Foreground(telegramBlue).Render(input))
		b.WriteString("\n\n" + lipgloss.NewStyle().Foreground(telegramBlue).Render("  [submit]"))
	}

	b.WriteString("\n\n" + helpStyle.Render("enter: submit • q: quit"))

	return b.String()
}

func (m model) renderInput(index int) string {
	input := m.inputs[index].View()
	if m.focusIndex == index {
		return lipgloss.NewStyle().Foreground(telegramBlue).Render(input)
	}
	return lipgloss.NewStyle().Faint(true).Render(input)
}

func (m model) renderButton(text string, color lipgloss.Color) string {
	if m.focusIndex == len(m.inputs) {
		return "  " + lipgloss.NewStyle().Foreground(color).Render("[" + text + "]")
	}
	return "  [" + text + "]"
}

func (m model) GetCredentials() (phone string) {
	return m.phone
}

func (m model) GetCode() string {
	if len(m.inputs) > 0 {
		return m.inputs[0].Value()
	}
	return ""
}

// RunLoginUI runs the interactive login UI and returns the phone number and verification code.
func RunLoginUI() (phone, code string, err error) {
	p := tea.NewProgram(initialModel())
	m, err := p.Run()
	if err != nil {
		return "", "", fmt.Errorf("could not start program: %w", err)
	}

	if m, ok := m.(model); ok {
		return m.GetCredentials(), m.GetCode(), nil
	}

	return "", "", fmt.Errorf("unexpected model type")
}

// SaveEnvFile saves the phone number to a .env file in the project directory.
func SaveEnvFile(projectDir, phone string) error {
	envPath := filepath.Join(projectDir, ".env")
	content := fmt.Sprintf("# Telegram API credentials\nTELEGRAM_PHONE=%s\n", phone)

	return os.WriteFile(envPath, []byte(content), 0600)
}
