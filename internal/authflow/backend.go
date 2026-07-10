// Package authflow provides a headless Telegram authentication workflow.
package authflow

import (
	"context"
	"log/slog"
	"time"

	"agent-telegram/internal/config"
	"agent-telegram/internal/tgauth"
	"agent-telegram/internal/types"
)

const defaultUserID = 1

// Backend is the Telegram authentication surface used by the headless CLI.
type Backend interface {
	SignInWithQR(
		ctx context.Context,
		onToken func(tokenURL string, expiresAt time.Time) error,
	) (*types.SignInResult, error)
	ExportSession(ctx context.Context) ([]byte, error)
}

// TelegramBackend adapts tgauth.Service to the headless auth workflow.
type TelegramBackend struct {
	service *tgauth.Service
	userID  int
}

// NewTelegramBackend creates a real Telegram auth backend.
func NewTelegramBackend(cfg *config.Config, logger *slog.Logger) *TelegramBackend {
	return &TelegramBackend{
		service: tgauth.NewService(cfg, logger),
		userID:  defaultUserID,
	}
}

// SignInWithQR completes login through Telegram QR code flow.
func (b *TelegramBackend) SignInWithQR(
	ctx context.Context,
	onToken func(tokenURL string, expiresAt time.Time) error,
) (*types.SignInResult, error) {
	return b.service.SignInWithQR(ctx, b.userID, onToken)
}

// ExportSession returns the current temporary auth session bytes.
func (b *TelegramBackend) ExportSession(_ context.Context) ([]byte, error) {
	return b.service.ExportSession(), nil
}
