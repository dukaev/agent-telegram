// Package tgauth provides Telegram authentication service for the agent-telegram application.
package tgauth

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"agent-telegram/internal/config"

	"github.com/gotd/td/session"
	gottg "github.com/gotd/td/telegram"
)

// Service handles Telegram authentication operations.
type Service struct {
	cfg    *config.Config
	logger *slog.Logger
}

// NewService creates a new Telegram auth service.
func NewService(cfg *config.Config, logger *slog.Logger) *Service {
	if logger == nil {
		logger = slog.Default()
	}
	return &Service{
		cfg:    cfg,
		logger: logger,
	}
}

// NewServiceFromEnv creates a new Telegram auth service from environment variables.
func NewServiceFromEnv(logger *slog.Logger) (*Service, error) {
	cfg, err := config.LoadFromEnv()
	if err != nil {
		return nil, err
	}
	return NewService(cfg, logger), nil
}

// CreateClient creates a new Telegram client.
func (s *Service) CreateClient(_ int) (*gottg.Client, error) {
	// Create session directory
	if err := os.MkdirAll(s.cfg.SessionPath, 0700); err != nil {
		return nil, fmt.Errorf("failed to create session directory: %w", err)
	}

	// Create session storage - use same path as serve command
	sessionStorage := &session.FileStorage{
		Path: filepath.Join(s.cfg.SessionPath, "session.json"),
	}

	client := gottg.NewClient(s.cfg.AppID, s.cfg.AppHash, gottg.Options{
		SessionStorage: sessionStorage,
	})

	return client, nil
}
