// Package media provides Telegram media sending operations.
package media

import (
	"context"
	"fmt"
	"os"

	"github.com/gotd/td/telegram/uploader"
	"github.com/gotd/td/tg"
)

// Client provides media operations.
type Client struct {
	api    *tg.Client
	parent ParentClient
}

// ParentClient is an interface for accessing parent client methods.
type ParentClient interface {
	ResolvePeer(ctx context.Context, peer string) (tg.InputPeerClass, error)
}

// NewClient creates a new media client.
func NewClient(tc ParentClient) *Client {
	return &Client{
		parent: tc,
	}
}

// SetAPI sets the API client (called when the telegram client is initialized).
func (c *Client) SetAPI(api *tg.Client) {
	c.api = api
}

// resolvePeer resolves a peer string to InputPeerClass using the parent client's cache.
func (c *Client) resolvePeer(ctx context.Context, peer string) (tg.InputPeerClass, error) {
	if c.parent == nil {
		return nil, fmt.Errorf("parent client not set")
	}
	return c.parent.ResolvePeer(ctx, peer)
}

// uploadFile uploads a file to Telegram.
func uploadFile(ctx context.Context, api *tg.Client, filePath string) (tg.InputFileClass, error) {
	// #nosec G304 -- filePath is validated in handler
	file, err := os.Open(filePath)
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

	return uploadedFile, nil
}

// extractMessageID extracts message ID from Updates response.
func extractMessageID(result tg.UpdatesClass) int64 {
	switch r := result.(type) {
	case *tg.Updates:
		if len(r.Updates) > 0 {
			if msg, ok := r.Updates[0].(*tg.UpdateMessageID); ok {
				return int64(msg.ID)
			}
		}
	case *tg.UpdateShortSentMessage:
		return int64(r.ID)
	}
	return 0
}
