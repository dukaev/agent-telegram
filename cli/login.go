// Package cli provides interactive UI components for the agent-telegram tool.
package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"agent-telegram/cli/steps"
	"agent-telegram/internal/auth"
	"agent-telegram/pkg/common"
)

// Step interface represents a login step.
type Step interface {
	Init() tea.Cmd
	View() string
}

// LoginModel manages the login flow.
type LoginModel struct {
	ctx         context.Context
	authService *auth.Service
	currentStep Step
	phone       string
	twoFAHint   string
	quitting    bool
	successMsg  string
	errorMsg    string
	sessionPath string
}

// NewLoginModel creates a new login model starting with phone step.
func NewLoginModel(ctx context.Context, authService *auth.Service) LoginModel {
	return LoginModel{
		ctx:         ctx,
		authService: authService,
		currentStep: steps.NewPhoneStep(authService),
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

	// Handle specific message types
	switch msg := msg.(type) {
	case auth.Result:
		if msg.Error != "" {
			m.errorMsg = msg.Error
			m.quitting = true
			return m, tea.Quit
		}

		// Check current step and transition accordingly
		switch m.currentStep.(type) {
		case steps.PhoneStep:
			// Phone code sent successfully, move to code step
			if msg.Success {
				m.phone = m.authService.GetPhoneNumber()
				m.currentStep = steps.NewCodeStep(m.authService)
				return m, m.currentStep.Init()
			}
		case steps.CodeStep:
			// Code verified
			if msg.Requires2FA {
				m.twoFAHint = msg.TwoFactorHint
				m.currentStep = steps.NewPasswordStep(m.authService, msg.TwoFactorHint)
				return m, m.currentStep.Init()
			}
			if msg.Success {
				// Login successful without 2FA
				m.sessionPath = m.authService.GetSessionPath()
				m.quitting = true
				m.successMsg = fmt.Sprintf("✓ Login successful as user %s", m.phone)
				return m, tea.Quit
			}
		case steps.PasswordStep:
			// 2FA authentication successful
			if msg.Success {
				m.sessionPath = m.authService.GetSessionPath()
				m.quitting = true
				m.successMsg = fmt.Sprintf("✓ Login successful as user %s", m.phone)
				return m, tea.Quit
			}
		}

	case steps.AuthError:
		m.errorMsg = msg.Error
		m.quitting = true
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
		if m.errorMsg != "" {
			return common.ErrorStyle.Render("✗ Error: " + m.errorMsg)
		}
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

// IsSuccess returns true if login was successful.
func (m LoginModel) IsSuccess() bool {
	return m.successMsg != ""
}

// GetError returns the error message if any.
func (m LoginModel) GetError() string {
	return m.errorMsg
}

// GetPhone returns the entered phone number.
func (m LoginModel) GetPhone() string {
	return m.phone
}

// GetSessionPath returns the session file path.
func (m LoginModel) GetSessionPath() string {
	return m.sessionPath
}

// SaveEnvFile saves the phone number to a .env file in the project directory.
func SaveEnvFile(projectDir, phone string) error {
	envPath := filepath.Join(projectDir, ".env")
	content := fmt.Sprintf("# Telegram API credentials\nTELEGRAM_PHONE=%s\n", phone)
	return os.WriteFile(envPath, []byte(content), 0600)
}

// RunLoginUIWithAuth runs the interactive login UI with Telegram authentication.
// Returns session path on success, error on failure.
func RunLoginUIWithAuth(ctx context.Context, authService *auth.Service) (sessionPath string, err error) {
	p := tea.NewProgram(NewLoginModel(ctx, authService))
	m, err := p.Run()
	if err != nil {
		return "", fmt.Errorf("could not start program: %w", err)
	}

	if m, ok := m.(LoginModel); ok {
		if m.GetError() != "" {
			return "", fmt.Errorf("authentication failed: %s", m.GetError())
		}
		if m.IsSuccess() {
			return m.GetSessionPath(), nil
		}
		return "", fmt.Errorf("authentication incomplete")
	}

	return "", fmt.Errorf("unexpected model type")
}
