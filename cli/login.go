// Package cli provides interactive UI components for the agent-telegram tool.
package cli

import (
	"fmt"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	// Telegram brand colors
	// Primary: Telegram Blue (#0088cc / #2AABEE)
	// Neutral: Gray tones
	telegramBlue = lipgloss.Color("#0088cc") // Telegram signature blue
	normalColor  = lipgloss.Color("#6c6c6c") // Medium gray for secondary text

	// Telegram brand styles
	titleStyle = lipgloss.NewStyle().
			Foreground(telegramBlue).
			Bold(true)

	labelStyle = lipgloss.NewStyle().
			Foreground(normalColor)

	helpStyle = lipgloss.NewStyle().
			Foreground(normalColor)
)

// LoginModel manages the login flow between phone and code steps.
type LoginModel struct {
	currentStep interface{}
	phone       string
	code        string
	quitting    bool
}

// NewLoginModel creates a new login model starting with phone step.
func NewLoginModel() LoginModel {
	return LoginModel{
		currentStep: NewPhoneStep(),
	}
}

// Init initializes the login model.
func (m LoginModel) Init() tea.Cmd {
	if step, ok := m.currentStep.(interface{ Init() tea.Cmd }); ok {
		return step.Init()
	}
	return nil
}

// Update handles messages and transitions between steps.
func (m LoginModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.quitting {
		return m, tea.Quit
	}

	switch msg := msg.(type) {
	case PhoneSubmitted:
		m.phone = string(msg)
		m.currentStep = NewCodeStep(m.phone)
		return m, m.currentStep.(interface{ Init() tea.Cmd }).Init()

	case CodeSubmitted:
		m.code = string(msg)
		m.quitting = true
		return m, tea.Quit

	case tea.KeyMsg:
		if msg.String() == "ctrl+c" || msg.String() == "q" {
			m.quitting = true
			return m, tea.Quit
		}
	}

	// Delegate update to current step
	switch step := m.currentStep.(type) {
	case PhoneStep:
		updatedStep, cmd := step.Update(msg)
		m.currentStep = updatedStep
		return m, cmd

	case CodeStep:
		updatedStep, cmd := step.Update(msg)
		m.currentStep = updatedStep
		return m, cmd
	}

	return m, nil
}

// View renders the current step.
func (m LoginModel) View() string {
	if m.quitting {
		return ""
	}

	switch step := m.currentStep.(type) {
	case PhoneStep:
		return step.View()
	case CodeStep:
		return step.View()
	}

	return ""
}

// GetPhone returns the entered phone number.
func (m LoginModel) GetPhone() string {
	return m.phone
}

// GetCode returns the entered verification code.
func (m LoginModel) GetCode() string {
	return m.code
}

// RunLoginUI runs the interactive login UI and returns the phone number and verification code.
func RunLoginUI() (phone, code string, err error) {
	p := tea.NewProgram(NewLoginModel())
	m, err := p.Run()
	if err != nil {
		return "", "", fmt.Errorf("could not start program: %w", err)
	}

	if m, ok := m.(LoginModel); ok {
		return m.GetPhone(), m.GetCode(), nil
	}

	return "", "", fmt.Errorf("unexpected model type")
}

// SaveEnvFile saves the phone number to a .env file in the project directory.
func SaveEnvFile(projectDir, phone string) error {
	envPath := filepath.Join(projectDir, ".env")
	content := fmt.Sprintf("# Telegram API credentials\nTELEGRAM_PHONE=%s\n", phone)
	return os.WriteFile(envPath, []byte(content), 0600)
}
