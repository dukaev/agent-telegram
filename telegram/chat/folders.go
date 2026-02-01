// Package chat provides folder operations.
package chat

import (
	"context"
	"fmt"

	"github.com/gotd/td/tg"
	"agent-telegram/telegram/types"
)

// GetFolders retrieves all chat folders.
func (c *Client) GetFolders(ctx context.Context, _ types.GetFoldersParams) (*types.GetFoldersResult, error) {
	if err := c.CheckInitialized(); err != nil {
		return nil, err
	}

	result, err := c.API.MessagesGetDialogFilters(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get folders: %w", err)
	}

	folders := make([]types.ChatFolder, 0)

	for _, filter := range result.Filters {
		if f, ok := filter.(*tg.DialogFilter); ok {
			folder := types.ChatFolder{
				ID:                 f.ID,
				Title:              f.Title.Text,
				IncludeContacts:    f.Contacts,
				IncludeNonContacts: f.NonContacts,
				IncludeGroups:      f.Groups,
				IncludeChannels:    f.Broadcasts,
				IncludeBots:        f.Bots,
			}
			folders = append(folders, folder)
		}
	}

	return &types.GetFoldersResult{
		Folders: folders,
		Count:   len(folders),
	}, nil
}

// CreateFolder creates a new chat folder.
func (c *Client) CreateFolder(ctx context.Context, params types.CreateFolderParams) (*types.CreateFolderResult, error) {
	if err := c.CheckInitialized(); err != nil {
		return nil, err
	}

	// Get existing filters to find next ID
	existing, err := c.API.MessagesGetDialogFilters(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get existing folders: %w", err)
	}

	// Find max ID
	maxID := 1
	for _, filter := range existing.Filters {
		if f, ok := filter.(*tg.DialogFilter); ok {
			if f.ID >= maxID {
				maxID = f.ID + 1
			}
		}
	}

	// Build included peers
	includePeers := make([]tg.InputPeerClass, 0)
	for _, peer := range params.IncludedChats {
		inputPeer, err := c.ResolvePeer(ctx, peer)
		if err != nil {
			continue // Skip unresolvable peers
		}
		includePeers = append(includePeers, inputPeer)
	}

	// Build excluded peers
	excludePeers := make([]tg.InputPeerClass, 0)
	for _, peer := range params.ExcludedChats {
		inputPeer, err := c.ResolvePeer(ctx, peer)
		if err != nil {
			continue
		}
		excludePeers = append(excludePeers, inputPeer)
	}

	filter := &tg.DialogFilter{
		ID:           maxID,
		Title:        tg.TextWithEntities{Text: params.Title},
		Contacts:     params.IncludeContacts,
		NonContacts:  params.IncludeNonContacts,
		Groups:       params.IncludeGroups,
		Broadcasts:   params.IncludeChannels,
		Bots:         params.IncludeBots,
		IncludePeers: includePeers,
		ExcludePeers: excludePeers,
		PinnedPeers:  []tg.InputPeerClass{},
	}

	_, err = c.API.MessagesUpdateDialogFilter(ctx, &tg.MessagesUpdateDialogFilterRequest{
		ID:     maxID,
		Filter: filter,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create folder: %w", err)
	}

	return &types.CreateFolderResult{
		Success: true,
		ID:      maxID,
	}, nil
}

// DeleteFolder deletes a chat folder.
func (c *Client) DeleteFolder(ctx context.Context, params types.DeleteFolderParams) (*types.DeleteFolderResult, error) {
	if err := c.CheckInitialized(); err != nil {
		return nil, err
	}

	_, err := c.API.MessagesUpdateDialogFilter(ctx, &tg.MessagesUpdateDialogFilterRequest{
		ID: params.ID,
		// No filter = delete
	})
	if err != nil {
		return nil, fmt.Errorf("failed to delete folder: %w", err)
	}

	return &types.DeleteFolderResult{Success: true}, nil
}
