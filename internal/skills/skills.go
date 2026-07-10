package skills

import (
	"cmp"
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
)

//go:embed bundled/**
var bundledFS embed.FS

// Skill describes an installable agent skill bundled with agent-telegram.
type Skill struct {
	Name             string `json:"name"`
	DisplayName      string `json:"displayName"`
	Description      string `json:"description"`
	ShortDescription string `json:"shortDescription"`
	DefaultPrompt    string `json:"defaultPrompt"`
	InstallCommand   string `json:"installCommand"`
}

var bundledSkills = []Skill{{
	Name:             "agent-telegram",
	DisplayName:      "Agent Telegram",
	Description:      "Use when Codex needs to operate Telegram through the agent-telegram CLI: authenticate, start or inspect the IPC server, list/read/send messages, interact with bots, press buttons, inspect chats/users/gifts, or debug runs with audit/log/trace commands.",
	ShortDescription: "Telegram CLI workflows for agents",
	DefaultPrompt:    "Use $agent-telegram to inspect chats, send messages, and debug Telegram bot flows.",
	InstallCommand:   "agent-telegram skills install agent-telegram",
}}

// List returns bundled installable skills.
func List() []Skill {
	out := slices.Clone(bundledSkills)
	slices.SortFunc(out, func(a, b Skill) int {
		return cmp.Compare(a.Name, b.Name)
	})
	return out
}

// Get returns a bundled skill by name.
func Get(name string) (Skill, bool) {
	for _, skill := range bundledSkills {
		if skill.Name == name {
			return skill, true
		}
	}
	return Skill{}, false
}

// Manifest returns JSON-safe skill metadata for agent discovery.
func Manifest() []map[string]any {
	items := List()
	out := make([]map[string]any, 0, len(items))
	for _, skill := range items {
		out = append(out, map[string]any{
			"name":             skill.Name,
			"displayName":      skill.DisplayName,
			"description":      skill.Description,
			"shortDescription": skill.ShortDescription,
			"defaultPrompt":    skill.DefaultPrompt,
			"installCommand":   skill.InstallCommand,
		})
	}
	return out
}

// DefaultInstallDir returns the canonical user skill directory.
func DefaultInstallDir() string {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return filepath.Join(".agents", "skills")
	}
	return filepath.Join(home, ".agents", "skills")
}

// Install copies one bundled skill into targetDir/name.
func Install(name, targetDir string, force bool) (string, error) {
	if _, ok := Get(name); !ok {
		return "", fmt.Errorf("unknown skill %q", name)
	}
	if targetDir == "" {
		targetDir = DefaultInstallDir()
	}
	sourceRoot := filepath.ToSlash(filepath.Join("bundled", name))
	destinationRoot := filepath.Join(targetDir, name)
	if _, err := os.Lstat(destinationRoot); err == nil && !force {
		return "", fmt.Errorf("skill %q already exists at %s; pass --force to overwrite", name, destinationRoot)
	} else if err != nil && !os.IsNotExist(err) {
		return "", fmt.Errorf("inspect skill destination: %w", err)
	}
	if force {
		if err := os.RemoveAll(destinationRoot); err != nil {
			return "", err
		}
	}
	if err := os.MkdirAll(destinationRoot, 0o750); err != nil {
		return "", err
	}
	if err := fs.WalkDir(bundledFS, sourceRoot, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		rel, err := filepath.Rel(sourceRoot, path)
		if err != nil {
			return err
		}
		if rel == "." {
			return nil
		}
		dst := filepath.Join(destinationRoot, rel)
		if entry.IsDir() {
			return os.MkdirAll(dst, 0o750)
		}
		data, err := bundledFS.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(dst, data, 0o600)
	}); err != nil {
		return "", err
	}
	return destinationRoot, nil
}
