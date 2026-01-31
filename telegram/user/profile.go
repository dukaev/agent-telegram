// Package user provides Telegram user profile operations.
package user

import (
	"context"
	"fmt"
	"os"

	"github.com/gotd/td/tg"
	"github.com/gotd/td/telegram/uploader"
	"agent-telegram/telegram/types"
)

// UpdateProfile updates the user's profile.
func (c *Client) UpdateProfile(
	ctx context.Context, params types.UpdateProfileParams,
) (*types.UpdateProfileResult, error) {
	if c.api == nil {
		return nil, fmt.Errorf("client not initialized")
	}

	_, err := c.api.AccountUpdateProfile(ctx, &tg.AccountUpdateProfileRequest{
		FirstName: params.FirstName,
		LastName:  params.LastName,
		About:     params.Bio,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update profile: %w", err)
	}

	return &types.UpdateProfileResult{
		Success: true,
	}, nil
}

// UpdateAvatar updates the user's avatar/profile photo.
func (c *Client) UpdateAvatar(ctx context.Context, params types.UpdateAvatarParams) (*types.UpdateAvatarResult, error) {
	if c.api == nil {
		return nil, fmt.Errorf("client not initialized")
	}

	// #nosec G304 -- filePath is validated in handler
	file, err := os.Open(params.File)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer func() { _ = file.Close() }()

	fileInfo, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}

	u := uploader.NewUploader(c.api)
	upload := uploader.NewUpload(fileInfo.Name(), file, fileInfo.Size())

	uploadedFile, err := u.Upload(ctx, upload)
	if err != nil {
		return nil, fmt.Errorf("failed to upload file: %w", err)
	}

	_, err = c.api.PhotosUploadProfilePhoto(ctx, &tg.PhotosUploadProfilePhotoRequest{
		File: uploadedFile,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update avatar: %w", err)
	}

	return &types.UpdateAvatarResult{
		Success: true,
	}, nil
}
