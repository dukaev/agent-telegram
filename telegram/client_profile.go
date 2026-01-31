// Package telegram provides Telegram client profile functionality.
package telegram

import (
	"context"
	"fmt"
	"os"

	"github.com/gotd/td/tg"
	"github.com/gotd/td/telegram/uploader"
)

// UpdateProfile updates the user's profile.
func (c *Client) UpdateProfile(ctx context.Context, params UpdateProfileParams) (*UpdateProfileResult, error) {
	if c.client == nil {
		return nil, fmt.Errorf("client not initialized")
	}

	api := c.client.API()

	_, err := api.AccountUpdateProfile(ctx, &tg.AccountUpdateProfileRequest{
		FirstName: params.FirstName,
		LastName:  params.LastName,
		About:     params.Bio,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update profile: %w", err)
	}

	return &UpdateProfileResult{
		Success: true,
	}, nil
}

// UpdateAvatar updates the user's avatar/profile photo.
func (c *Client) UpdateAvatar(ctx context.Context, params UpdateAvatarParams) (*UpdateAvatarResult, error) {
	if c.client == nil {
		return nil, fmt.Errorf("client not initialized")
	}

	api := c.client.API()

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

	u := uploader.NewUploader(api)
	upload := uploader.NewUpload(fileInfo.Name(), file, fileInfo.Size())

	uploadedFile, err := u.Upload(ctx, upload)
	if err != nil {
		return nil, fmt.Errorf("failed to upload file: %w", err)
	}

	_, err = api.PhotosUploadProfilePhoto(ctx, &tg.PhotosUploadProfilePhotoRequest{
		File: uploadedFile,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update avatar: %w", err)
	}

	return &UpdateAvatarResult{
		Success: true,
	}, nil
}
