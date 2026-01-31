// Package telegram provides Telegram sign-in functionality.
package telegram

import (
	"context"
	"fmt"

	"agent-telegram/internal/types"

	"github.com/gotd/td/tg"
)

// signInResult holds the result of sign in operation.
type signInResult struct {
	requires2FA   bool
	twoFactorHint string
	authError     string
	success       bool
}

// SendCode sends a verification code to the specified phone number.
func (s *Service) SendCode(ctx context.Context, userID int, phoneNumber string) (*types.SendCodeResult, error) {
	s.logger.Info("Sending verification code",
		"user_id", userID,
		"phone_number", phoneNumber)

	client, err := s.CreateClient(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}

	var phoneCodeHash string
	var timeout int

	err = client.Run(ctx, func(ctx context.Context) error {
		api := client.API()

		sentCode, err := api.AuthSendCode(ctx, &tg.AuthSendCodeRequest{
			PhoneNumber: phoneNumber,
			APIID:       s.cfg.AppID,
			APIHash:     s.cfg.AppHash,
		})
		if err != nil {
			s.logger.Error("Failed to send verification code",
				"phone_number", phoneNumber,
				"error", err)
			return fmt.Errorf("failed to send verification code: %w", err)
		}

		switch code := sentCode.(type) {
		case *tg.AuthSentCode:
			phoneCodeHash = code.PhoneCodeHash
			timeout = code.Timeout
			s.logger.Info("Verification code sent successfully",
				"phone_number", phoneNumber,
				"phone_code_hash", phoneCodeHash,
				"timeout", timeout)
		default:
			return fmt.Errorf("unexpected response type: %T", sentCode)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to send verification code: %w", err)
	}

	return &types.SendCodeResult{
		PhoneCodeHash: phoneCodeHash,
		Timeout:       timeout,
	}, nil
}

// SignIn authenticates with the verification code.
func (s *Service) SignIn(
	ctx context.Context,
	userID int,
	phoneNumber, phoneCode, phoneCodeHash string,
) (*types.SignInResult, error) {
	s.logger.Info("Starting sign in with verification code",
		"user_id", userID,
		"phone_number", phoneNumber)

	client, err := s.CreateClient(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}

	var result signInResult

	err = client.Run(ctx, func(ctx context.Context) error {
		api := client.API()
		var err error

		authResult, err := api.AuthSignIn(ctx, &tg.AuthSignInRequest{
			PhoneNumber:   phoneNumber,
			PhoneCodeHash: phoneCodeHash,
			PhoneCode:     phoneCode,
		})

		if err != nil {
			return s.handleSignInError(ctx, api, phoneNumber, err, &result)
		}

		return s.processAuthResult(authResult, phoneNumber, &result)
	})

	if err != nil && result.authError == "" {
		return nil, err
	}

	return &types.SignInResult{
		Success:       result.success,
		Requires2FA:   result.requires2FA,
		TwoFactorHint: result.twoFactorHint,
		AuthError:     result.authError,
	}, nil
}

// processAuthResult processes the authentication result.
func (s *Service) processAuthResult(
	authResult tg.AuthAuthorizationClass,
	phoneNumber string,
	result *signInResult,
) error {
	switch r := authResult.(type) {
	case *tg.AuthAuthorization:
		s.logger.Info("Authentication successful", "phone_number", phoneNumber)
		result.success = true
	case *tg.AuthAuthorizationSignUpRequired:
		result.authError = "account registration required"
		return fmt.Errorf("account registration required for phone number: %s", phoneNumber)
	default:
		result.authError = fmt.Sprintf("unexpected authentication result type: %T", r)
		return fmt.Errorf("unexpected authentication result: %T", r)
	}
	return nil
}
