// Package authflow provides a headless Telegram authentication workflow.
package authflow

import (
	"context"
	"log/slog"
	"path/filepath"

	"agent-telegram/internal/config"
	"agent-telegram/internal/tgauth"
	"agent-telegram/internal/types"
)

const defaultUserID = 1

// Backend is the Telegram authentication surface used by the headless CLI.
type Backend interface {
	SendCode(ctx context.Context, phone string) (*types.SendCodeResult, error)
	SignIn(ctx context.Context, phone, code, codeHash string) (*types.SignInResult, error)
	SignInWith2FA(ctx context.Context, phone, password string) (*types.SignInResult, error)
	SessionPath() string
}

// TelegramBackend adapts tgauth.Service to the headless auth workflow.
type TelegramBackend struct {
	service     *tgauth.Service
	userID      int
	sessionPath string
}

// NewTelegramBackend creates a real Telegram auth backend.
func NewTelegramBackend(cfg *config.Config, logger *slog.Logger) *TelegramBackend {
	return &TelegramBackend{
		service:     tgauth.NewService(cfg, logger),
		userID:      defaultUserID,
		sessionPath: filepath.Join(cfg.SessionPath, "session.json"),
	}
}

// SendCode sends a login code to the phone number.
func (b *TelegramBackend) SendCode(ctx context.Context, phone string) (*types.SendCodeResult, error) {
	return b.service.SendCode(ctx, b.userID, phone)
}

// SignIn completes login with a Telegram code.
func (b *TelegramBackend) SignIn(
	ctx context.Context,
	phone, code, codeHash string,
) (*types.SignInResult, error) {
	return b.service.SignIn(ctx, b.userID, phone, code, codeHash)
}

// SignInWith2FA completes login with a Telegram 2FA password.
func (b *TelegramBackend) SignInWith2FA(ctx context.Context, phone, password string) (*types.SignInResult, error) {
	return b.service.SignInWith2FA(ctx, b.userID, phone, password)
}

// SessionPath returns the session file path used by this backend.
func (b *TelegramBackend) SessionPath() string {
	return b.sessionPath
}
