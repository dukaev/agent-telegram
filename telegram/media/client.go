// Package media provides Telegram media sending operations.
package media

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/uploader"
	"github.com/gotd/td/tg"
)

// Client provides media operations.
type Client struct {
	api      *tg.Client
	telegram *telegram.Client
}

// NewClient creates a new media client.
func NewClient(tc *telegram.Client) *Client {
	return &Client{
		telegram: tc,
	}
}

// SetAPI sets the API client (called when the telegram client is initialized).
func (c *Client) SetAPI(api *tg.Client) {
	c.api = api
}

// resolvePeer resolves peer from username.
func resolvePeer(ctx context.Context, api *tg.Client, peer string) (tg.InputPeerClass, error) {
	peer = strings.TrimPrefix(peer, "@")

	peerClass, err := api.ContactsResolveUsername(ctx, &tg.ContactsResolveUsernameRequest{Username: peer})
	if err != nil {
		return nil, fmt.Errorf("failed to resolve peer @%s: %w", peer, err)
	}

	switch p := peerClass.Peer.(type) {
	case *tg.PeerUser:
		return &tg.InputPeerUser{
			UserID:     p.UserID,
			AccessHash: getAccessHash(peerClass, p.UserID),
		}, nil
	case *tg.PeerChat:
		return &tg.InputPeerChat{
			ChatID: p.ChatID,
		}, nil
	case *tg.PeerChannel:
		return &tg.InputPeerChannel{
			ChannelID:  p.ChannelID,
			AccessHash: getAccessHash(peerClass, p.ChannelID),
		}, nil
	default:
		return nil, fmt.Errorf("unsupported peer type: %T", p)
	}
}

// getAccessHash extracts access hash from the resolved peer.
func getAccessHash(peerClass *tg.ContactsResolvedPeer, id int64) int64 {
	for _, chat := range peerClass.Chats {
		switch c := chat.(type) {
		case *tg.Channel:
			if c.ID == id {
				return c.AccessHash
			}
		case *tg.Chat:
			if c.ID == id {
				return 0
			}
		}
	}
	for _, user := range peerClass.Users {
		if u, ok := user.(*tg.User); ok && u.ID == id {
			return u.AccessHash
		}
	}
	return 0
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
