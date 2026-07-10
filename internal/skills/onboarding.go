package skills

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"agent-telegram/internal/paths"
)

// OnboardingAction describes the next safe onboarding step.
type OnboardingAction uint8

const (
	// OnboardingNone means the skill already exists at an applicable location.
	OnboardingNone OnboardingAction = iota
	// OnboardingInstallProject installs into an existing project skills directory.
	OnboardingInstallProject
	// OnboardingPromptGlobal requires consent before global installation.
	OnboardingPromptGlobal
)

// OnboardingDecision identifies the selected scope and install target.
type OnboardingDecision struct {
	Action     OnboardingAction
	InstallDir string
	Target     string
}

// ResolveOnboarding decides whether to do nothing, install in-project, or ask
// before installing globally.
func ResolveOnboarding(name, cwd string) (OnboardingDecision, error) {
	absCWD, err := filepath.Abs(cwd)
	if err != nil {
		return OnboardingDecision{}, fmt.Errorf("resolve working directory: %w", err)
	}
	projectDirs, err := projectSkillDirs(filepath.Clean(absCWD))
	if err != nil {
		return OnboardingDecision{}, err
	}
	canonicalDir, err := canonicalUserInstallDir()
	if err != nil {
		return OnboardingDecision{}, err
	}
	legacyDir, err := legacyUserInstallDir()
	if err != nil {
		return OnboardingDecision{}, err
	}

	allDirs := append(append([]string{}, projectDirs...), canonicalDir)
	if legacyDir != canonicalDir {
		allDirs = append(allDirs, legacyDir)
	}
	for _, dir := range allDirs {
		exists, err := pathEntryExists(filepath.Join(dir, name))
		if err != nil {
			return OnboardingDecision{}, err
		}
		if exists {
			return OnboardingDecision{Action: OnboardingNone}, nil
		}
	}

	if len(projectDirs) > 0 {
		dir := projectDirs[0]
		return OnboardingDecision{
			Action:     OnboardingInstallProject,
			InstallDir: dir,
			Target:     filepath.Join(dir, name),
		}, nil
	}
	return OnboardingDecision{
		Action:     OnboardingPromptGlobal,
		InstallDir: canonicalDir,
		Target:     filepath.Join(canonicalDir, name),
	}, nil
}

func projectSkillDirs(cwd string) ([]string, error) {
	root, found, err := repositoryRoot(cwd)
	if err != nil {
		return nil, err
	}
	limit := cwd
	if found {
		limit = root
	}

	var dirs []string
	for current := cwd; ; current = filepath.Dir(current) {
		candidate := filepath.Join(current, ".agents", "skills")
		_, lstatErr := os.Lstat(candidate)
		switch {
		case lstatErr == nil:
			info, statErr := os.Stat(candidate)
			if statErr != nil {
				return nil, fmt.Errorf("inspect project skill path %s: %w", candidate, statErr)
			}
			if !info.IsDir() {
				return nil, fmt.Errorf("project skill path is not a directory: %s", candidate)
			}
			dirs = append(dirs, candidate)
		case os.IsNotExist(lstatErr):
			// The project has not opted into skills at this level.
		default:
			return nil, fmt.Errorf("inspect project skill path %s: %w", candidate, lstatErr)
		}

		if current == limit {
			break
		}
		parent := filepath.Dir(current)
		if parent == current {
			break
		}
	}
	return dirs, nil
}

func repositoryRoot(cwd string) (string, bool, error) {
	for current := cwd; ; current = filepath.Dir(current) {
		_, err := os.Lstat(filepath.Join(current, ".git"))
		if err == nil {
			return current, true, nil
		}
		if !os.IsNotExist(err) {
			return "", false, fmt.Errorf("inspect repository boundary: %w", err)
		}
		parent := filepath.Dir(current)
		if parent == current {
			return "", false, nil
		}
	}
}

func canonicalUserInstallDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve user home directory: %w", err)
	}
	if home == "" {
		return "", fmt.Errorf("resolve user home directory: home is empty")
	}
	return filepath.Join(home, ".agents", "skills"), nil
}

func legacyUserInstallDir() (string, error) {
	if codexHome := strings.TrimSpace(os.Getenv("CODEX_HOME")); codexHome != "" {
		return filepath.Abs(filepath.Join(codexHome, "skills"))
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve legacy home directory: %w", err)
	}
	if home == "" {
		return "", fmt.Errorf("resolve legacy home directory: home is empty")
	}
	return filepath.Join(home, ".codex", "skills"), nil
}

func pathEntryExists(path string) (bool, error) {
	_, err := os.Lstat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, fmt.Errorf("inspect skill target %s: %w", path, err)
}

// GlobalPromptDismissed reports whether the user declined this global target.
func GlobalPromptDismissed(target string) (bool, error) {
	marker, err := globalDismissalPath()
	if err != nil {
		return false, err
	}
	data, err := os.ReadFile(marker) //nolint:gosec // fixed per-user state path
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, fmt.Errorf("read global skill dismissal: %w", err)
	}
	return strings.TrimSpace(string(data)) == filepath.Clean(target), nil
}

// RecordGlobalPromptDismissal securely records a declined global target.
func RecordGlobalPromptDismissal(target string) error {
	dir, err := paths.EnsureConfigDir()
	if err != nil {
		return err
	}
	if err := os.Chmod(dir, 0o700); err != nil { //nolint:gosec // directory needs owner execute permission
		return fmt.Errorf("secure config directory: %w", err)
	}
	marker := filepath.Join(dir, "skill-prompt-dismissed")
	temporary, err := os.CreateTemp(dir, ".skill-prompt-dismissed-*")
	if err != nil {
		return fmt.Errorf("create global skill dismissal: %w", err)
	}
	temporaryPath := temporary.Name()
	defer func() {
		_ = os.Remove(temporaryPath)
	}()
	if err := temporary.Chmod(0o600); err != nil {
		_ = temporary.Close()
		return fmt.Errorf("secure global skill dismissal: %w", err)
	}
	if _, err := fmt.Fprintln(temporary, filepath.Clean(target)); err != nil {
		_ = temporary.Close()
		return fmt.Errorf("write global skill dismissal: %w", err)
	}
	if err := temporary.Close(); err != nil {
		return fmt.Errorf("close global skill dismissal: %w", err)
	}
	if err := os.Remove(marker); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove previous global skill dismissal: %w", err)
	}
	if err := os.Rename(temporaryPath, marker); err != nil {
		return fmt.Errorf("replace global skill dismissal: %w", err)
	}
	return nil
}

func globalDismissalPath() (string, error) {
	dir, err := paths.ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "skill-prompt-dismissed"), nil
}
