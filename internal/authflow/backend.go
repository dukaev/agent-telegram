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
	SendCode(ctx context.Context, phone string) (*types.SendCodeResult, error)
	SignIn(ctx context.Context, phone, code, codeHash string) (*types.SignInResult, error)
	SignInWith2FA(ctx context.Context, phone, password string) (*types.SignInResult, error)
	SignInWithQR(
		ctx context.Context,
		onToken func(tokenURL string, expiresAt time.Time) error,
	) (*types.SignInResult, error)
	ImportSession(ctx context.Context, data []byte) error
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

// SignInWithQR completes login through Telegram QR code flow.
func (b *TelegramBackend) SignInWithQR(
	ctx context.Context,
	onToken func(tokenURL string, expiresAt time.Time) error,
) (*types.SignInResult, error) {
	return b.service.SignInWithQR(ctx, b.userID, onToken)
}

// ImportSession restores temporary auth session bytes before the next auth step.
func (b *TelegramBackend) ImportSession(ctx context.Context, data []byte) error {
	return b.service.ImportSession(ctx, data)
}

// ExportSession returns the current temporary auth session bytes.
func (b *TelegramBackend) ExportSession(_ context.Context) ([]byte, error) {
	return b.service.ExportSession(), nil
}
