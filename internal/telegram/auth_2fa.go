// Package telegram provides 2FA authentication functionality.
package telegram

import (
	"context"
	"fmt"
	"strings"

	"agent-telegram/internal/types"

	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/tg"
)

// handleSignInError handles sign in errors, checking for 2FA requirement.
func (s *Service) handleSignInError(
	ctx context.Context,
	api *tg.Client,
	phoneNumber string,
	err error,
	result *signInResult,
) error {
	if !strings.Contains(err.Error(), "SESSION_PASSWORD_NEEDED") {
		result.authError = err.Error()
		s.logger.Error("Authentication failed",
			"phone_number", phoneNumber,
			"error", err)
		return fmt.Errorf("authentication failed: %w", err)
	}

	s.logger.Info("2FA password required", "phone_number", phoneNumber)

	passwordInfo, pwdErr := api.AccountGetPassword(ctx)
	if pwdErr != nil {
		return fmt.Errorf("failed to get 2FA info: %w", pwdErr)
	}

	result.requires2FA = true
	if passwordInfo.Hint != "" {
		result.twoFactorHint = passwordInfo.Hint
	}

	s.logger.Info("2FA info retrieved",
		"phone_number", phoneNumber,
		"hint", result.twoFactorHint)

	return nil
}

// SignInWith2FA authenticates with the 2FA password.
func (s *Service) SignInWith2FA(
	ctx context.Context,
	userID int,
	phoneNumber, password string,
) (*types.SignInResult, error) {
	s.logger.Info("Signing in with 2FA password",
		"user_id", userID,
		"phone_number", phoneNumber)

	client, err := s.CreateClient(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}

	err = client.Run(ctx, func(ctx context.Context) error {
		api := client.API()

		passwordInfo, err := api.AccountGetPassword(ctx)
		if err != nil {
			return fmt.Errorf("failed to get password info: %w", err)
		}

		srpPassword, err := auth.PasswordHash(
			[]byte(password),
			passwordInfo.SRPID,
			passwordInfo.SRPB,
			nil,
			passwordInfo.CurrentAlgo,
		)
		if err != nil {
			return fmt.Errorf("failed to create SRP password: %w", err)
		}

		authResult, err := api.AuthCheckPassword(ctx, srpPassword)
		if err != nil {
			return fmt.Errorf("failed to authenticate with 2FA password: %w", err)
		}

		switch r := authResult.(type) {
		case *tg.AuthAuthorization:
			s.logger.Info("2FA authentication successful", "phone_number", phoneNumber)
		case *tg.AuthAuthorizationSignUpRequired:
			return fmt.Errorf("account registration required for phone number: %s", phoneNumber)
		default:
			return fmt.Errorf("unexpected 2FA authentication result: %T", r)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	s.logger.Info("User authenticated with 2FA successfully",
		"user_id", userID,
		"phone_number", phoneNumber)

	return &types.SignInResult{
		Success: true,
	}, nil
}
