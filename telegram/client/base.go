// Package client provides shared base functionality for Telegram domain clients.
package client

import (
	"context"
	"fmt"

	"github.com/gotd/td/tg"
)

// BaseClient provides shared functionality for all domain clients.
// Domain clients should embed this struct to inherit common methods.
type BaseClient struct {
	API    *tg.Client
	Parent ParentClient
}

// ParentClient is an interface for accessing parent client methods.
type ParentClient interface {
	ResolvePeer(ctx context.Context, peer string) (tg.InputPeerClass, error)
}

// SetAPI sets the API client (called when the telegram client is initialized).
func (b *BaseClient) SetAPI(api *tg.Client) {
	b.API = api
}

// ResolvePeer resolves a peer string to InputPeerClass using the parent client's cache.
func (b *BaseClient) ResolvePeer(ctx context.Context, peer string) (tg.InputPeerClass, error) {
	if b.Parent == nil {
		return nil, fmt.Errorf("parent client not set")
	}
	return b.Parent.ResolvePeer(ctx, peer)
}

// IsInitialized returns true if the API client is set.
func (b *BaseClient) IsInitialized() bool {
	return b.API != nil
}

// CheckInitialized returns an error if the API client is not set.
func (b *BaseClient) CheckInitialized() error {
	if !b.IsInitialized() {
		return fmt.Errorf("api client not set")
	}
	return nil
}
