// Package types provides common types for Telegram chat folders.
package types

// GetFoldersParams holds parameters for GetFolders.
type GetFoldersParams struct {
	// No required params
}

// Validate validates GetFoldersParams.
func (p GetFoldersParams) Validate() error {
	return nil
}

// ChatFolder represents a chat folder.
type ChatFolder struct {
	ID                 int      `json:"id"`
	Title              string   `json:"title"`
	IncludedChats      []string `json:"includedChats,omitempty"`
	ExcludedChats      []string `json:"excludedChats,omitempty"`
	IncludeContacts    bool     `json:"includeContacts,omitempty"`
	IncludeNonContacts bool     `json:"includeNonContacts,omitempty"`
	IncludeGroups      bool     `json:"includeGroups,omitempty"`
	IncludeChannels    bool     `json:"includeChannels,omitempty"`
	IncludeBots        bool     `json:"includeBots,omitempty"`
}

// GetFoldersResult is the result of GetFolders.
type GetFoldersResult struct {
	Folders []ChatFolder `json:"folders"`
	Count   int          `json:"count"`
}

// CreateFolderParams holds parameters for CreateFolder.
type CreateFolderParams struct {
	Title              string   `json:"title" validate:"required"`
	IncludedChats      []string `json:"includedChats,omitempty"`
	ExcludedChats      []string `json:"excludedChats,omitempty"`
	IncludeContacts    bool     `json:"includeContacts,omitempty"`
	IncludeNonContacts bool     `json:"includeNonContacts,omitempty"`
	IncludeGroups      bool     `json:"includeGroups,omitempty"`
	IncludeChannels    bool     `json:"includeChannels,omitempty"`
	IncludeBots        bool     `json:"includeBots,omitempty"`
}

// Validate validates CreateFolderParams.
func (p CreateFolderParams) Validate() error {
	return ValidateStruct(p)
}

// CreateFolderResult is the result of CreateFolder.
type CreateFolderResult struct {
	Success bool `json:"success"`
	ID      int  `json:"id"`
}

// DeleteFolderParams holds parameters for DeleteFolder.
type DeleteFolderParams struct {
	ID int `json:"id" validate:"required"`
}

// Validate validates DeleteFolderParams.
func (p DeleteFolderParams) Validate() error {
	return ValidateStruct(p)
}

// DeleteFolderResult is the result of DeleteFolder.
type DeleteFolderResult struct {
	Success bool `json:"success"`
}
