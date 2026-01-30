// Package cli provides interactive UI components for the agent-telegram tool.
package cli

import (
	"fmt"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"agent-telegram/cli/steps"
	"agent-telegram/pkg/common"
)

// Step interface represents a login step.
type Step interface {
	Init() tea.Cmd
	View() string
}

// LoginModel manages the login flow.
type LoginModel struct {
	currentStep Step
	phone       string
	code        string
	password    string
	quitting    bool
	successMsg  string
}

// NewLoginModel creates a new login model starting with phone step.
func NewLoginModel() LoginModel {
	return LoginModel{
		currentStep: steps.NewPhoneStep(),
	}
}

// Init initializes the login model.
func (m LoginModel) Init() tea.Cmd {
	if step := m.currentStep; step != nil {
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
	case steps.PhoneSubmitted:
		m.phone = string(msg)
		m.currentStep = steps.NewCodeStep()
		return m, m.currentStep.Init()

	case steps.CodeSubmitted:
		m.code = string(msg)
		m.currentStep = steps.NewPasswordStep()
		return m, m.currentStep.Init()

	case steps.PasswordSubmitted:
		m.password = string(msg)
		m.quitting = true
		m.successMsg = fmt.Sprintf("Login as user %s", m.phone)
		return m, tea.Quit

	case tea.KeyMsg:
		if msg.String() == "ctrl+c" || msg.String() == "q" {
			m.quitting = true
			return m, tea.Quit
		}
	}

	// Delegate update to current step
	var cmd tea.Cmd
	switch step := m.currentStep.(type) {
	case steps.PhoneStep:
		m.currentStep, cmd = step.Update(msg)
	case steps.CodeStep:
		m.currentStep, cmd = step.Update(msg)
	case steps.PasswordStep:
		m.currentStep, cmd = step.Update(msg)
	}

	return m, cmd
}

// View renders the current step.
func (m LoginModel) View() string {
	if m.quitting {
		if m.successMsg != "" {
			return common.TitleStyle.Render(m.successMsg)
		}
		return ""
	}

	if m.currentStep != nil {
		return m.currentStep.View()
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

// GetPassword returns the entered 2FA password.
func (m LoginModel) GetPassword() string {
	return m.password
}

// RunLoginUI runs the interactive login UI and returns the phone number, code, and password.
func RunLoginUI() (phone, code, password string, err error) {
	p := tea.NewProgram(NewLoginModel())
	m, err := p.Run()
	if err != nil {
		return "", "", "", fmt.Errorf("could not start program: %w", err)
	}

	if m, ok := m.(LoginModel); ok {
		return m.GetPhone(), m.GetCode(), m.GetPassword(), nil
	}

	return "", "", "", fmt.Errorf("unexpected model type")
}

// SaveEnvFile saves the phone number to a .env file in the project directory.
func SaveEnvFile(projectDir, phone string) error {
	envPath := filepath.Join(projectDir, ".env")
	content := fmt.Sprintf("# Telegram API credentials\nTELEGRAM_PHONE=%s\n", phone)
	return os.WriteFile(envPath, []byte(content), 0600)
}
