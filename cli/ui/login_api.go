// Package ui provides login UI public API.
package ui

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"agent-telegram/internal/auth"
)

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
