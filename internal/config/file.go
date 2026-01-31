// Package config provides configuration management for the agent-telegram application.
package config

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/knadh/koanf/v2"
	"github.com/knadh/koanf/parsers/json"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
)

// SupportedFileExtensions returns the list of supported file extensions.
func SupportedFileExtensions() []string {
	return []string{".json", ".yaml", ".yml"}
}

// IsSupportedFile returns true if the file extension is supported.
func IsSupportedFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	for _, supported := range SupportedFileExtensions() {
		if ext == supported {
			return true
		}
	}
	return false
}

// parserForFile returns the appropriate parser for the given file extension.
func parserForFile(path string) (koanf.Parser, error) {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".json":
		return json.Parser(), nil
	case ".yaml", ".yml":
		return yaml.Parser(), nil
	default:
		return nil, fmt.Errorf("unsupported config file format: %s (supported: %v)",
			filepath.Ext(path), SupportedFileExtensions())
	}
}

// loadFromFile loads configuration from a file (JSON or YAML) into a koanf instance.
func loadFromFile(k *koanf.Koanf, path string) error {
	parser, err := parserForFile(path)
	if err != nil {
		return err
	}
	return k.Load(file.Provider(path), parser)
}
