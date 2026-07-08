// Package tgauth provides Telegram QR authentication functionality.
package tgauth

import (
	"context"
	"fmt"
	"time"

	"agent-telegram/internal/types"

	"github.com/gotd/td/telegram/auth/qrlogin"
	"github.com/gotd/td/tg"
)

// SignInWithQR authenticates through Telegram QR login.
func (s *Service) SignInWithQR(
	ctx context.Context,
	userID int,
	onToken func(tokenURL string, expiresAt time.Time) error,
) (*types.SignInResult, error) {
	s.logger.Info("Starting QR sign in", "user_id", userID)

	dispatcher := tg.NewUpdateDispatcher()
	loggedIn := qrlogin.OnLoginToken(dispatcher)

	client, err := s.CreateClientWithUpdateHandler(userID, dispatcher)
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}

	var authResult *tg.AuthAuthorization
	err = client.Run(ctx, func(ctx context.Context) error {
		var err error
		authResult, err = client.QR().Auth(ctx, loggedIn, func(_ context.Context, token qrlogin.Token) error {
			if onToken == nil {
				return nil
			}
			return onToken(token.URL(), token.Expires())
		})
		return err
	})
	if err != nil {
		return nil, err
	}

	if authResult == nil {
		return nil, fmt.Errorf("QR login did not return authorization")
	}

	s.logger.Info("QR authentication successful", "user_id", userID)

	return &types.SignInResult{Success: true}, nil
}
