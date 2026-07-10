// Package tgauth provides Telegram authentication service for the agent-telegram application.
package tgauth

import (
	"context"
	"fmt"
	"log/slog"

	"agent-telegram/internal/config"
	tgapp "agent-telegram/telegram"

	gottg "github.com/gotd/td/telegram"
)

// Service handles Telegram authentication operations.
type Service struct {
	cfg            *config.Config
	logger         *slog.Logger
	sessionStorage *tgapp.EnvStorage
}

// NewService creates a new Telegram auth service.
func NewService(cfg *config.Config, logger *slog.Logger) *Service {
	if logger == nil {
		logger = slog.Default()
	}
	return &Service{
		cfg:            cfg,
		logger:         logger,
		sessionStorage: tgapp.NewMemoryStorage(nil),
	}
}

// CreateClient creates a new Telegram client.
func (s *Service) CreateClient(userID int) (*gottg.Client, error) {
	return s.CreateClientWithUpdateHandler(userID, nil)
}

// CreateClientWithUpdateHandler creates a new Telegram client with optional update handler.
func (s *Service) CreateClientWithUpdateHandler(_ int, updateHandler gottg.UpdateHandler) (*gottg.Client, error) {
	if s.sessionStorage == nil {
		return nil, fmt.Errorf("session storage is not initialized")
	}
	client := gottg.NewClient(s.cfg.AppID, s.cfg.AppHash, gottg.Options{
		SessionStorage: s.sessionStorage,
		UpdateHandler:  updateHandler,
	})

	return client, nil
}

// ImportSession loads raw Telegram session bytes into this auth service.
func (s *Service) ImportSession(ctx context.Context, data []byte) error {
	if s.sessionStorage == nil {
		s.sessionStorage = tgapp.NewMemoryStorage(nil)
	}
	if len(data) == 0 {
		return nil
	}
	return s.sessionStorage.StoreSession(ctx, data)
}

// ExportSession returns a copy of the current in-memory Telegram session.
func (s *Service) ExportSession() []byte {
	if s.sessionStorage == nil {
		return nil
	}
	return s.sessionStorage.ExportSession()
}
