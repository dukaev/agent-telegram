// Package telegram provides Telegram client operations including authentication
// and other Telegram API interactions.
package telegram

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"agent-telegram/internal/config"
	"agent-telegram/internal/types"

	"github.com/gotd/td/session"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/tg"
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

// signInResult holds the result of sign in operation.
type signInResult struct {
	requires2FA   bool
	twoFactorHint string
	authError     string
	success       bool
}

// SignIn authenticates with the verification code.
func (s *Service) SignIn(ctx context.Context, userID int, phoneNumber, phoneCode, phoneCodeHash string) (*types.SignInResult, error) {
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

// handleSignInError handles sign in errors, checking for 2FA requirement.
func (s *Service) handleSignInError(ctx context.Context, api *tg.Client, phoneNumber string, err error, result *signInResult) error {
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

// processAuthResult processes the authentication result.
func (s *Service) processAuthResult(authResult tg.AuthAuthorizationClass, phoneNumber string, result *signInResult) error {
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

// SignInWith2FA authenticates with the 2FA password.
func (s *Service) SignInWith2FA(ctx context.Context, userID int, phoneNumber, password string) (*types.SignInResult, error) {
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
