// Package sessionstore provides pluggable storage for Telegram sessions.
package sessionstore

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"
)

const (
	// EnvProvider selects a provider by name.
	EnvProvider = "AGENT_TELEGRAM_SESSION_PROVIDER"
	// EnvProfile selects the session profile within a provider.
	EnvProfile = "AGENT_TELEGRAM_PROFILE"
	// DefaultProfile is used when no profile is configured.
	DefaultProfile = "default"
)

var ErrNotFound = errors.New("session not found")

// Store persists opaque Telegram session bytes for named profiles.
type Store interface {
	Provider() string
	Persistent() bool
	Load(ctx context.Context, profile string) ([]byte, error)
	Save(ctx context.Context, profile string, data []byte) error
	Delete(ctx context.Context, profile string) error
}

// Factory creates a provider. Third-party providers can register a factory at
// process startup without coupling auth or daemon packages to the backend.
type Factory func() (Store, error)

// ProviderInfo describes a registered session provider.
type ProviderInfo struct {
	Name       string `json:"name"`
	Default    bool   `json:"default"`
	Persistent bool   `json:"persistent"`
}

// Selection identifies the configured session location.
type Selection struct {
	Provider   string `json:"provider"`
	Profile    string `json:"profile"`
	Persistent bool   `json:"persistent"`
}

var providerRegistry = struct {
	sync.RWMutex
	factories map[string]Factory
}{factories: make(map[string]Factory)}

// RegisterProvider adds a session provider factory. Names are normalized to
// lowercase and duplicate registrations panic during program initialization.
func RegisterProvider(name string, factory Factory) {
	name = normalizeProvider(name)
	if name == "" || factory == nil {
		panic("sessionstore: provider name and factory are required")
	}
	providerRegistry.Lock()
	defer providerRegistry.Unlock()
	if _, exists := providerRegistry.factories[name]; exists {
		panic("sessionstore: provider already registered: " + name)
	}
	providerRegistry.factories[name] = factory
}

// Open creates a storage adapter for provider and profile.
func Open(provider, profile string) (*Storage, error) {
	provider = normalizeProvider(provider)
	if provider == "" {
		provider = normalizeProvider(os.Getenv(EnvProvider))
		if provider == "" {
			provider = DefaultProvider()
		}
	}
	if strings.TrimSpace(profile) == "" {
		profile = os.Getenv(EnvProfile)
	}
	profile, err := normalizeProfile(profile)
	if err != nil {
		return nil, err
	}

	providerRegistry.RLock()
	factory := providerRegistry.factories[provider]
	providerRegistry.RUnlock()
	if factory == nil {
		return nil, fmt.Errorf("unknown session provider %q (available: %s)", provider, strings.Join(ProviderNames(), ", "))
	}
	store, err := factory()
	if err != nil {
		return nil, fmt.Errorf("open session provider %q: %w", provider, err)
	}
	return NewStorage(store, profile), nil
}

// OpenDefault opens the provider/profile selected by environment or platform defaults.
func OpenDefault() (*Storage, error) {
	return Open("", "")
}

// DefaultProvider returns the native provider preferred on this build.
func DefaultProvider() string {
	return platformDefaultProvider()
}

// ProviderNames returns registered providers in stable order.
func ProviderNames() []string {
	providerRegistry.RLock()
	names := make([]string, 0, len(providerRegistry.factories))
	for name := range providerRegistry.factories {
		names = append(names, name)
	}
	providerRegistry.RUnlock()
	sort.Strings(names)
	return names
}

// Providers returns provider capabilities without exposing session contents.
func Providers() []ProviderInfo {
	defaultName := DefaultProvider()
	names := ProviderNames()
	result := make([]ProviderInfo, 0, len(names))
	for _, name := range names {
		storage, err := Open(name, DefaultProfile)
		result = append(result, ProviderInfo{
			Name:       name,
			Default:    name == defaultName,
			Persistent: err == nil && storage.Persistent(),
		})
	}
	return result
}

func normalizeProvider(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func normalizeProfile(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		value = DefaultProfile
	}
	if len(value) > 64 {
		return "", fmt.Errorf("session profile is too long")
	}
	for _, r := range value {
		if r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '-' || r == '_' || r == '.' {
			continue
		}
		return "", fmt.Errorf("session profile %q contains unsupported characters", value)
	}
	return value, nil
}
