// Package telegram provides Telegram client operations including authentication
// and other Telegram API interactions.
package telegram

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"agent-telegram/internal/config"

	"github.com/gotd/td/session"
	"github.com/gotd/td/telegram"
)

// Service handles Telegram operations including authentication.
type Service struct {
	cfg    *config.Config
	logger *slog.Logger
}

// NewService creates a new Telegram service.
func NewService(cfg *config.Config, logger *slog.Logger) *Service {
	if logger == nil {
		logger = slog.Default()
	}
	return &Service{
		cfg:    cfg,
		logger: logger,
	}
}

// NewServiceFromEnv creates a new Telegram service from environment variables.
func NewServiceFromEnv(logger *slog.Logger) (*Service, error) {
	cfg, err := config.LoadFromEnv()
	if err != nil {
		return nil, err
	}
	return NewService(cfg, logger), nil
}

// CreateClient creates a new Telegram client for the given user ID.
func (s *Service) CreateClient(userID int) (*telegram.Client, error) {
	// Create user session directory
	sessionDir := filepath.Join(s.cfg.SessionPath, fmt.Sprintf("user_%d", userID))
	if err := os.MkdirAll(sessionDir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create session directory: %w", err)
	}

	// Create session storage
	sessionStorage := &session.FileStorage{
		Path: filepath.Join(sessionDir, "session.json"),
	}

	client := telegram.NewClient(s.cfg.AppID, s.cfg.AppHash, telegram.Options{
		SessionStorage: sessionStorage,
	})

	return client, nil
}
