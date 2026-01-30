// Package auth provides authentication service for Telegram integration.
package auth

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"

	"agent-telegram/internal/config"
	"agent-telegram/internal/telegram"
	tea "github.com/charmbracelet/bubbletea"
)

// Result represents the result of an authentication operation.
type Result struct {
	Success       bool
	PhoneCodeHash string
	Requires2FA   bool
	TwoFactorHint string
	Error         string
}

// Service wraps the Telegram service for TUI integration.
type Service struct {
	telegramService *telegram.Service
	userID          int
	ctx             context.Context
	logger          *slog.Logger
	phoneNumber     string
	codeHash        string
}

// NewService creates a new auth service from environment variables.
// Phone is optional and can be set later via the UI.
func NewService(ctx context.Context) (*Service, error) {
	cfg, err := config.LoadFromEnvWithOptionalPhone()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	telegramService := telegram.NewService(cfg, slog.Default())

	return &Service{
		telegramService: telegramService,
		userID:          1, // Default user ID
		ctx:             ctx,
		logger:          slog.Default(),
	}, nil
}

// NewServiceWithConfig creates a new auth service with custom config.
func NewServiceWithConfig(ctx context.Context, appID int, appHash, phone string) (*Service, error) {
	sessionPath := filepath.Join(getConfigDir(), ".agent-telegram")
	cfg := config.LoadFromArgs(appID, appHash, phone, sessionPath)

	telegramService := telegram.NewService(cfg, slog.Default())

	return &Service{
		telegramService: telegramService,
		userID:          1, // Default user ID
		ctx:             ctx,
		logger:          slog.Default(),
		phoneNumber:     phone,
	}, nil
}

// SendCode sends a verification code to the phone number.
func (s *Service) SendCode(phone string) tea.Cmd {
	return func() tea.Msg {
		result, err := s.telegramService.SendCode(s.ctx, s.userID, phone)
		if err != nil {
			return Result{
				Error: fmt.Sprintf("Failed to send code: %v", err),
			}
		}

		s.phoneNumber = phone
		s.codeHash = result.PhoneCodeHash

		return Result{
			Success:       true,
			PhoneCodeHash: result.PhoneCodeHash,
		}
	}
}

// SignIn authenticates with the verification code.
func (s *Service) SignIn(code string) tea.Cmd {
	return func() tea.Msg {
		result, err := s.telegramService.SignIn(s.ctx, s.userID, s.phoneNumber, code, s.codeHash)
		if err != nil {
			return Result{
				Error: fmt.Sprintf("Sign in failed: %v", err),
			}
		}

		return Result{
			Success:       result.Success,
			Requires2FA:   result.Requires2FA,
			TwoFactorHint: result.TwoFactorHint,
			Error:         result.AuthError,
		}
	}
}

// SignInWith2FA authenticates with the 2FA password.
func (s *Service) SignInWith2FA(password string) tea.Cmd {
	return func() tea.Msg {
		result, err := s.telegramService.SignInWith2FA(s.ctx, s.userID, s.phoneNumber, password)
		if err != nil {
			return Result{
				Error: fmt.Sprintf("2FA authentication failed: %v", err),
			}
		}

		return Result{
			Success: result.Success,
		}
	}
}

// GetSessionPath returns the session path for the current user.
func (s *Service) GetSessionPath() string {
	return filepath.Join(getConfigDir(), ".agent-telegram", fmt.Sprintf("user_%d", s.userID), "session.json")
}

// GetPhoneNumber returns the stored phone number.
func (s *Service) GetPhoneNumber() string {
	return s.phoneNumber
}

// GetCodeHash returns the stored code hash.
func (s *Service) GetCodeHash() string {
	return s.codeHash
}

// SetPhoneNumber sets the phone number.
func (s *Service) SetPhoneNumber(phone string) {
	s.phoneNumber = phone
}

// SetCodeHash sets the code hash.
func (s *Service) SetCodeHash(hash string) {
	s.codeHash = hash
}

// getConfigDir returns the config directory path.
func getConfigDir() string {
	if home := os.Getenv("HOME"); home != "" {
		return home
	}
	return "."
}

// ParseAppID parses app ID from string.
func ParseAppID(appIDStr string) (int, error) {
	return strconv.Atoi(appIDStr)
}
