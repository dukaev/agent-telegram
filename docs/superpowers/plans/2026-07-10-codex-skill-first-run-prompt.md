# Project-Aware Codex Skill Onboarding Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Automatically install the bundled `agent-telegram` skill into an existing project `.agents/skills` directory while requiring explicit consent for global `$HOME/.agents/skills` installation.

**Architecture:** Keep path discovery and preference persistence in `internal/skills`, independent from Cobra and terminal I/O. The root command performs eligibility checks, executes the resolver decision, auto-installs project scope, or asks for global consent through injected services that make every branch testable.

**Tech Stack:** Go 1.25.4, Cobra 1.10.2, standard library filesystem and terminal primitives, existing `embed.FS` skill bundle.

## Global Constraints

- Run onboarding only for the interactive root command with no arguments or flags, terminal stdin and stdout, and an empty `CI` variable.
- Search repository `.agents/skills` locations from CWD to the nearest `.git` boundary; without a repository inspect only CWD.
- Automatically install only into an already existing project `.agents/skills` directory.
- Require explicit `y` or `yes` before creating or installing into global `$HOME/.agents/skills`.
- Treat files, directories, valid symlinks, and dangling symlinks at skill targets as user-owned.
- Never pass `force=true` from onboarding and never rewrite an existing target.
- Recognize `${CODEX_HOME}/skills/agent-telegram` or `$HOME/.codex/skills/agent-telegram` legacy installations.
- Do not fall back globally after malformed project paths or project installation failures.
- Write onboarding messages to stderr and preserve non-interactive stdout byte-for-byte.
- Make all onboarding failures non-fatal to the root command.
- Add no third-party dependency.
- Use temporary HOME, CODEX_HOME, CWD, and repository trees in filesystem tests.

---

## File Structure

- Modify `internal/skills/skills.go` and `skills_test.go`: canonical user default and `Lstat` overwrite protection.
- Create `internal/skills/onboarding.go` and `onboarding_test.go`: project discovery, canonical/legacy checks, onboarding decision, and global preference.
- Modify `cmd/root.go` and `root_test.go`: eligibility, project auto-install, and global consent.
- Modify `cmd/sys/skills.go`, `internal/docs/docs.go`, `internal/docs/docs_test.go`, `README.md`, and `DEVELOPMENT.md`: canonical help and documentation.

### Task 1: Canonical explicit installer path and overwrite safety

**Files:**
- Modify: `internal/skills/skills.go`
- Modify: `internal/skills/skills_test.go`

**Interfaces:**
- Consumes: `os.UserHomeDir()` and the existing embedded bundle.
- Produces: `DefaultInstallDir() string` returning `$HOME/.agents/skills`, plus `Install(name, targetDir string, force bool)` that detects dangling symlinks with `os.Lstat`.

- [ ] **Step 1: Write failing canonical-path and dangling-symlink tests**

Append to `internal/skills/skills_test.go`:

```go
func TestDefaultInstallDirUsesCanonicalUserSkills(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("CODEX_HOME", filepath.Join(home, "legacy-codex-home"))
	want := filepath.Join(home, ".agents", "skills")
	if got := DefaultInstallDir(); got != want {
		t.Fatalf("DefaultInstallDir() = %q, want %q", got, want)
	}
}

func TestInstallDoesNotReplaceDanglingSymlinkWithoutForce(t *testing.T) {
	target := t.TempDir()
	destination := filepath.Join(target, "agent-telegram")
	missing := filepath.Join(target, "missing-target")
	if err := os.Symlink(missing, destination); err != nil {
		t.Fatal(err)
	}
	if _, err := Install("agent-telegram", target, false); err == nil {
		t.Fatal("install should reject an existing dangling symlink")
	}
	info, err := os.Lstat(destination)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode()&os.ModeSymlink == 0 {
		t.Fatalf("destination mode = %v, want symlink", info.Mode())
	}
}

func TestInstallCreatesCanonicalUserSkillParents(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	installed, err := Install("agent-telegram", "", false)
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(home, ".agents", "skills", "agent-telegram")
	if installed != want {
		t.Fatalf("installed = %q, want %q", installed, want)
	}
	if _, err := os.Stat(filepath.Join(installed, "SKILL.md")); err != nil {
		t.Fatal(err)
	}
}
```

- [ ] **Step 2: Run the focused tests and verify failure**

```bash
go test ./internal/skills -run 'TestDefaultInstallDirUsesCanonicalUserSkills|TestInstallDoesNotReplaceDanglingSymlinkWithoutForce|TestInstallCreatesCanonicalUserSkillParents' -count=1
```

Expected: default-path assertion reports `.codex/skills`; dangling-symlink test exposes the `os.Stat` gap.

- [ ] **Step 3: Implement the canonical default and Lstat check**

Replace `DefaultInstallDir`:

```go
// DefaultInstallDir returns the canonical user skill directory.
func DefaultInstallDir() string {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return filepath.Join(".agents", "skills")
	}
	return filepath.Join(home, ".agents", "skills")
}
```

Replace the initial destination check in `Install`:

```go
	if _, err := os.Lstat(destinationRoot); err == nil && !force {
		return "", fmt.Errorf("skill %q already exists at %s; pass --force to overwrite", name, destinationRoot)
	} else if err != nil && !os.IsNotExist(err) {
		return "", fmt.Errorf("inspect skill destination: %w", err)
	}
```

- [ ] **Step 4: Format, test, and commit**

```bash
gofmt -w internal/skills/skills.go internal/skills/skills_test.go
go test ./internal/skills -count=1
git add internal/skills/skills.go internal/skills/skills_test.go
git commit -m "Use canonical Codex skill install path"
```

Expected: package tests pass and the commit succeeds.

### Task 2: Project-aware onboarding resolver and global preference

**Files:**
- Create: `internal/skills/onboarding.go`
- Create: `internal/skills/onboarding_test.go`

**Interfaces:**
- Consumes: `DefaultInstallDir() string`, `paths.ConfigDir()`, and `paths.EnsureConfigDir()`.
- Produces: `ResolveOnboarding(name, cwd string) (OnboardingDecision, error)`, `GlobalPromptDismissed(target string) (bool, error)`, and `RecordGlobalPromptDismissal(target string) error`.

- [ ] **Step 1: Write failing resolver tests**

Create `internal/skills/onboarding_test.go`:

```go
package skills

import (
	"os"
	"path/filepath"
	"testing"
)

func makeDir(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0o700); err != nil {
		t.Fatal(err)
	}
}

func TestResolveOnboardingChoosesNearestProjectSkills(t *testing.T) {
	home := t.TempDir()
	repo := filepath.Join(t.TempDir(), "repo")
	cwd := filepath.Join(repo, "services", "telegram")
	rootSkills := filepath.Join(repo, ".agents", "skills")
	nearSkills := filepath.Join(repo, "services", ".agents", "skills")
	makeDir(t, filepath.Join(repo, ".git"))
	makeDir(t, cwd)
	makeDir(t, rootSkills)
	makeDir(t, nearSkills)
	t.Setenv("HOME", home)
	t.Setenv("CODEX_HOME", "")

	decision, err := ResolveOnboarding("agent-telegram", cwd)
	if err != nil {
		t.Fatal(err)
	}
	if decision.Action != OnboardingInstallProject || decision.InstallDir != nearSkills {
		t.Fatalf("decision = %+v", decision)
	}
	if decision.Target != filepath.Join(nearSkills, "agent-telegram") {
		t.Fatalf("target = %q", decision.Target)
	}
}

func TestResolveOnboardingStopsForExistingApplicableSkill(t *testing.T) {
	tests := []struct {
		name string
		setup func(t *testing.T, home, rootSkills string)
	}{
		{"parent-project", func(t *testing.T, _, rootSkills string) {
			makeDir(t, filepath.Join(rootSkills, "agent-telegram"))
		}},
		{"canonical-user", func(t *testing.T, home, _ string) {
			makeDir(t, filepath.Join(home, ".agents", "skills", "agent-telegram"))
		}},
		{"legacy-user", func(t *testing.T, home, _ string) {
			makeDir(t, filepath.Join(home, ".codex", "skills", "agent-telegram"))
		}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			home := t.TempDir()
			repo := filepath.Join(t.TempDir(), "repo")
			cwd := filepath.Join(repo, "nested")
			rootSkills := filepath.Join(repo, ".agents", "skills")
			makeDir(t, filepath.Join(repo, ".git"))
			makeDir(t, filepath.Join(cwd, ".agents", "skills"))
			makeDir(t, rootSkills)
			t.Setenv("HOME", home)
			t.Setenv("CODEX_HOME", "")
			test.setup(t, home, rootSkills)
			decision, err := ResolveOnboarding("agent-telegram", cwd)
			if err != nil || decision.Action != OnboardingNone {
				t.Fatalf("decision=%+v err=%v", decision, err)
			}
		})
	}
}

func TestResolveOnboardingPromptsCanonicalGlobalWithoutProjectSkills(t *testing.T) {
	home := t.TempDir()
	repo := filepath.Join(t.TempDir(), "repo")
	makeDir(t, filepath.Join(repo, ".git"))
	t.Setenv("HOME", home)
	t.Setenv("CODEX_HOME", "")
	decision, err := ResolveOnboarding("agent-telegram", repo)
	if err != nil {
		t.Fatal(err)
	}
	wantDir := filepath.Join(home, ".agents", "skills")
	if decision.Action != OnboardingPromptGlobal || decision.InstallDir != wantDir {
		t.Fatalf("decision = %+v", decision)
	}
}

func TestResolveOnboardingNonRepositoryChecksOnlyCWD(t *testing.T) {
	home := t.TempDir()
	parent := t.TempDir()
	cwd := filepath.Join(parent, "work")
	makeDir(t, cwd)
	makeDir(t, filepath.Join(parent, ".agents", "skills"))
	t.Setenv("HOME", home)
	t.Setenv("CODEX_HOME", "")
	decision, err := ResolveOnboarding("agent-telegram", cwd)
	if err != nil || decision.Action != OnboardingPromptGlobal {
		t.Fatalf("decision=%+v err=%v", decision, err)
	}
}

func TestResolveOnboardingRejectsMalformedProjectSkills(t *testing.T) {
	home := t.TempDir()
	repo := filepath.Join(t.TempDir(), "repo")
	makeDir(t, filepath.Join(repo, ".git"))
	makeDir(t, filepath.Join(repo, ".agents"))
	if err := os.WriteFile(filepath.Join(repo, ".agents", "skills"), []byte("owned"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("HOME", home)
	if _, err := ResolveOnboarding("agent-telegram", repo); err == nil {
		t.Fatal("malformed project skills path should fail")
	}
}
```

- [ ] **Step 2: Verify resolver tests fail for missing symbols**

```bash
go test ./internal/skills -run TestResolveOnboarding -count=1
```

Expected: build failure containing `undefined: ResolveOnboarding` and the onboarding constants.

- [ ] **Step 3: Add failing global-dismissal tests**

Append:

```go
func TestGlobalPromptDismissalMatchesOnlyCanonicalTarget(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	targetA := filepath.Join(home, ".agents", "skills", "agent-telegram")
	targetB := filepath.Join(home, "other", "agent-telegram")
	if err := RecordGlobalPromptDismissal(targetA); err != nil {
		t.Fatal(err)
	}
	dismissed, err := GlobalPromptDismissed(targetA)
	if err != nil || !dismissed {
		t.Fatalf("matching target: dismissed=%v err=%v", dismissed, err)
	}
	dismissed, err = GlobalPromptDismissed(targetB)
	if err != nil || dismissed {
		t.Fatalf("different target: dismissed=%v err=%v", dismissed, err)
	}
	marker := filepath.Join(home, ".agent-telegram", "skill-prompt-dismissed")
	data, err := os.ReadFile(marker)
	if err != nil || string(data) != filepath.Clean(targetA)+"\n" {
		t.Fatalf("marker=%q err=%v", data, err)
	}
	info, err := os.Stat(marker)
	if err != nil || info.Mode().Perm() != 0o600 {
		t.Fatalf("marker mode error: info=%v err=%v", info, err)
	}
	dirInfo, err := os.Stat(filepath.Dir(marker))
	if err != nil || dirInfo.Mode().Perm() != 0o700 {
		t.Fatalf("marker directory mode error: info=%v err=%v", dirInfo, err)
	}
}
```

- [ ] **Step 4: Implement resolver and preference storage**

Create `internal/skills/onboarding.go`:

```go
package skills

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"agent-telegram/internal/paths"
)

type OnboardingAction uint8

const (
	OnboardingNone OnboardingAction = iota
	OnboardingInstallProject
	OnboardingPromptGlobal
)

type OnboardingDecision struct {
	Action OnboardingAction
	InstallDir string
	Target string
}

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
		return OnboardingDecision{Action: OnboardingInstallProject, InstallDir: dir, Target: filepath.Join(dir, name)}, nil
	}
	return OnboardingDecision{Action: OnboardingPromptGlobal, InstallDir: canonicalDir, Target: filepath.Join(canonicalDir, name)}, nil
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
		info, err := os.Stat(candidate)
		switch {
		case err == nil && !info.IsDir():
			return nil, fmt.Errorf("project skill path is not a directory: %s", candidate)
		case err == nil:
			dirs = append(dirs, candidate)
		case !os.IsNotExist(err):
			return nil, fmt.Errorf("inspect project skill path %s: %w", candidate, err)
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

func RecordGlobalPromptDismissal(target string) error {
	dir, err := paths.EnsureConfigDir()
	if err != nil {
		return err
	}
	if err := os.Chmod(dir, 0o700); err != nil {
		return fmt.Errorf("secure config directory: %w", err)
	}
	marker := filepath.Join(dir, "skill-prompt-dismissed")
	temporary, err := os.CreateTemp(dir, ".skill-prompt-dismissed-*")
	if err != nil {
		return fmt.Errorf("create global skill dismissal: %w", err)
	}
	temporaryPath := temporary.Name()
	defer os.Remove(temporaryPath)
	if err := temporary.Chmod(0o600); err != nil {
		_ = temporary.Close()
		return err
	}
	if _, err := fmt.Fprintln(temporary, filepath.Clean(target)); err != nil {
		_ = temporary.Close()
		return err
	}
	if err := temporary.Close(); err != nil {
		return err
	}
	if err := os.Remove(marker); err != nil && !os.IsNotExist(err) {
		return err
	}
	if err := os.Rename(temporaryPath, marker); err != nil {
		return err
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
```

- [ ] **Step 5: Add Git-file, CODEX_HOME, and dangling-target coverage**

Append:

```go
func TestResolveOnboardingSupportsGitFileAndCodexHomeLegacy(t *testing.T) {
	home := t.TempDir()
	codexHome := filepath.Join(home, "custom-codex")
	repo := filepath.Join(t.TempDir(), "repo")
	makeDir(t, repo)
	if err := os.WriteFile(filepath.Join(repo, ".git"), []byte("gitdir: elsewhere\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	makeDir(t, filepath.Join(repo, ".agents", "skills"))
	makeDir(t, filepath.Join(codexHome, "skills", "agent-telegram"))
	t.Setenv("HOME", home)
	t.Setenv("CODEX_HOME", codexHome)
	decision, err := ResolveOnboarding("agent-telegram", repo)
	if err != nil || decision.Action != OnboardingNone {
		t.Fatalf("decision=%+v err=%v", decision, err)
	}
}

func TestResolveOnboardingTreatsDanglingTargetAsExisting(t *testing.T) {
	home := t.TempDir()
	repo := filepath.Join(t.TempDir(), "repo")
	skillsDir := filepath.Join(repo, ".agents", "skills")
	makeDir(t, filepath.Join(repo, ".git"))
	makeDir(t, skillsDir)
	if err := os.Symlink(filepath.Join(repo, "missing"), filepath.Join(skillsDir, "agent-telegram")); err != nil {
		t.Fatal(err)
	}
	t.Setenv("HOME", home)
	t.Setenv("CODEX_HOME", "")
	decision, err := ResolveOnboarding("agent-telegram", repo)
	if err != nil || decision.Action != OnboardingNone {
		t.Fatalf("decision=%+v err=%v", decision, err)
	}
}

func TestResolveOnboardingTreatsAllTargetKindsAsExisting(t *testing.T) {
	creators := []struct {
		name string
		create func(t *testing.T, target string)
	}{
		{"file", func(t *testing.T, target string) {
			if err := os.WriteFile(target, []byte("owned"), 0o600); err != nil { t.Fatal(err) }
		}},
		{"directory", func(t *testing.T, target string) { makeDir(t, target) }},
		{"valid-symlink", func(t *testing.T, target string) {
			real := target + "-real"
			makeDir(t, real)
			if err := os.Symlink(real, target); err != nil { t.Skipf("symlink unavailable: %v", err) }
		}},
		{"dangling-symlink", func(t *testing.T, target string) {
			if err := os.Symlink(target+"-missing", target); err != nil { t.Skipf("symlink unavailable: %v", err) }
		}},
	}
	for _, creator := range creators {
		t.Run(creator.name, func(t *testing.T) {
			home := t.TempDir()
			repo := filepath.Join(t.TempDir(), "repo")
			skillsDir := filepath.Join(repo, ".agents", "skills")
			makeDir(t, filepath.Join(repo, ".git"))
			makeDir(t, skillsDir)
			creator.create(t, filepath.Join(skillsDir, "agent-telegram"))
			t.Setenv("HOME", home)
			t.Setenv("CODEX_HOME", "")
			decision, err := ResolveOnboarding("agent-telegram", repo)
			if err != nil || decision.Action != OnboardingNone { t.Fatalf("decision=%+v err=%v", decision, err) }
		})
	}
}

func TestResolveOnboardingFollowsProjectSkillsDirectorySymlink(t *testing.T) {
	home := t.TempDir()
	repo := filepath.Join(t.TempDir(), "repo")
	realSkills := filepath.Join(t.TempDir(), "shared-skills")
	makeDir(t, filepath.Join(repo, ".git"))
	makeDir(t, filepath.Join(repo, ".agents"))
	makeDir(t, realSkills)
	if err := os.Symlink(realSkills, filepath.Join(repo, ".agents", "skills")); err != nil { t.Skipf("symlink unavailable: %v", err) }
	t.Setenv("HOME", home)
	t.Setenv("CODEX_HOME", "")
	decision, err := ResolveOnboarding("agent-telegram", repo)
	if err != nil || decision.Action != OnboardingInstallProject { t.Fatalf("decision=%+v err=%v", decision, err) }
}

func TestResolveOnboardingUsesCWDProjectSkills(t *testing.T) {
	home := t.TempDir()
	repo := filepath.Join(t.TempDir(), "repo")
	cwd := filepath.Join(repo, "module")
	makeDir(t, filepath.Join(repo, ".git"))
	makeDir(t, filepath.Join(cwd, ".agents", "skills"))
	t.Setenv("HOME", home)
	t.Setenv("CODEX_HOME", "")
	decision, err := ResolveOnboarding("agent-telegram", cwd)
	if err != nil || decision.InstallDir != filepath.Join(cwd, ".agents", "skills") {
		t.Fatalf("decision=%+v err=%v", decision, err)
	}
}

func TestResolveOnboardingStopsAtNearestRepositoryBoundary(t *testing.T) {
	home := t.TempDir()
	outer := filepath.Join(t.TempDir(), "outer")
	inner := filepath.Join(outer, "inner")
	makeDir(t, filepath.Join(outer, ".git"))
	makeDir(t, filepath.Join(outer, ".agents", "skills"))
	makeDir(t, filepath.Join(inner, ".git"))
	t.Setenv("HOME", home)
	t.Setenv("CODEX_HOME", "")
	decision, err := ResolveOnboarding("agent-telegram", inner)
	if err != nil || decision.Action != OnboardingPromptGlobal {
		t.Fatalf("decision=%+v err=%v", decision, err)
	}
}
```

- [ ] **Step 6: Format, test, and commit**

```bash
gofmt -w internal/skills/onboarding.go internal/skills/onboarding_test.go
go test ./internal/skills -count=1
git add internal/skills/onboarding.go internal/skills/onboarding_test.go
git commit -m "Add project-aware skill onboarding resolver"
```

Expected: all resolver and preference tests pass.

### Task 3: Root command project auto-install and global consent

**Files:**
- Modify: `cmd/root.go`
- Modify: `cmd/root_test.go`

**Interfaces:**
- Consumes: `skills.ResolveOnboarding`, `skills.GlobalPromptDismissed`, `skills.RecordGlobalPromptDismissal`, and `skills.Install`.
- Produces: `onboardingServices`, `defaultOnboardingServices()`, `runRootWithServices`, `maybeOnboardSkill`, and `isTerminal`.

- [ ] **Step 1: Add failing root-flow tests**

Extend `cmd/root_test.go` imports with `errors` and `agent-telegram/internal/skills`. Append:

```go
type failingReader struct{}

func (failingReader) Read([]byte) (int, error) { return 0, errors.New("input failed") }

func onboardingTestCommand(input string) (*cobra.Command, *bytes.Buffer, *bytes.Buffer) {
	command := &cobra.Command{Use: "agent-telegram"}
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	command.SetIn(strings.NewReader(input))
	command.SetOut(stdout)
	command.SetErr(stderr)
	return command, stdout, stderr
}

func onboardingTestServices(decision skills.OnboardingDecision) onboardingServices {
	return onboardingServices{
		isTerminal: func(any) bool { return true },
		getwd: func() (string, error) { return "/repo", nil },
		resolve: func(string, string) (skills.OnboardingDecision, error) { return decision, nil },
		install: func(string, string, bool) (string, error) { return decision.Target, nil },
		dismissed: func(string) (bool, error) { return false, nil },
		recordDismissal: func(string) error { return nil },
	}
}

func TestMaybeOnboardSkillAutoInstallsProjectWithoutReadingInput(t *testing.T) {
	decision := skills.OnboardingDecision{Action: skills.OnboardingInstallProject, InstallDir: "/repo/.agents/skills", Target: "/repo/.agents/skills/agent-telegram"}
	command, _, stderr := onboardingTestCommand("")
	command.SetIn(failingReader{})
	services := onboardingTestServices(decision)
	called := false
	preferenceChecked := false
	services.dismissed = func(string) (bool, error) { preferenceChecked = true; return true, nil }
	services.install = func(name, target string, force bool) (string, error) {
		called = true
		if name != "agent-telegram" || target != decision.InstallDir || force {
			t.Fatalf("install args = %q %q %v", name, target, force)
		}
		return decision.Target, nil
	}
	maybeOnboardSkill(command, services)
	if !called || preferenceChecked || !strings.Contains(stderr.String(), "Installed project Codex skill") {
		t.Fatalf("called=%v preferenceChecked=%v stderr=%q", called, preferenceChecked, stderr.String())
	}
}

func TestMaybeOnboardSkillRequiresGlobalConsent(t *testing.T) {
	decision := skills.OnboardingDecision{Action: skills.OnboardingPromptGlobal, InstallDir: "/home/user/.agents/skills", Target: "/home/user/.agents/skills/agent-telegram"}
	for _, answer := range []string{"y\n", "yes\n", "Y\n", "YES\n"} {
		t.Run(strings.TrimSpace(answer), func(t *testing.T) {
			command, _, stderr := onboardingTestCommand(answer)
			services := onboardingTestServices(decision)
			called := false
			services.install = func(_ string, target string, force bool) (string, error) {
				called = true
				if target != decision.InstallDir || force { t.Fatalf("install target=%q force=%v", target, force) }
				return decision.Target, nil
			}
			maybeOnboardSkill(command, services)
			if !called || !strings.Contains(stderr.String(), "globally for Codex") {
				t.Fatalf("called=%v stderr=%q", called, stderr.String())
			}
		})
	}
}

func TestMaybeOnboardSkillRecordsGlobalDecline(t *testing.T) {
	decision := skills.OnboardingDecision{Action: skills.OnboardingPromptGlobal, Target: "/global/agent-telegram"}
	for _, answer := range []string{"\n", "n\n", "no\n", "later\n"} {
		t.Run(strings.TrimSpace(answer), func(t *testing.T) {
			command, _, _ := onboardingTestCommand(answer)
			services := onboardingTestServices(decision)
			recorded := ""
			services.recordDismissal = func(target string) error { recorded = target; return nil }
			maybeOnboardSkill(command, services)
			if recorded != decision.Target { t.Fatalf("recorded = %q", recorded) }
		})
	}
}
```

- [ ] **Step 2: Verify missing root helpers**

```bash
go test ./cmd -run TestMaybeOnboardSkill -count=1
```

Expected: build failure containing `undefined: onboardingServices` and `undefined: maybeOnboardSkill`.

- [ ] **Step 3: Add eligibility, EOF, failure, and compatibility tests**

Append:

```go
func TestMaybeOnboardSkillSkipsIneligibleCalls(t *testing.T) {
	decision := skills.OnboardingDecision{Action: skills.OnboardingPromptGlobal, Target: "/global/agent-telegram"}
	tests := []struct { name string; mutate func(*cobra.Command, *onboardingServices) }{
		{"ci", func(_ *cobra.Command, s *onboardingServices) { s.ci = "1" }},
		{"stdin", func(_ *cobra.Command, s *onboardingServices) { s.isTerminal = func(any) bool { return false } }},
		{"stdout", func(_ *cobra.Command, s *onboardingServices) {
			calls := 0
			s.isTerminal = func(any) bool { calls++; return calls == 1 }
		}},
		{"flag", func(command *cobra.Command, _ *onboardingServices) {
			command.Flags().Bool("quiet", false, "quiet")
			if err := command.Flags().Set("quiet", "true"); err != nil { t.Fatal(err) }
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			command, _, stderr := onboardingTestCommand("y\n")
			services := onboardingTestServices(decision)
			called := false
			services.getwd = func() (string, error) { called = true; return "/repo", nil }
			test.mutate(command, &services)
			maybeOnboardSkill(command, services)
			if called || stderr.Len() != 0 { t.Fatalf("called=%v stderr=%q", called, stderr.String()) }
		})
	}
}

func TestMaybeOnboardSkillHonorsGlobalDismissal(t *testing.T) {
	decision := skills.OnboardingDecision{Action: skills.OnboardingPromptGlobal, Target: "/global/agent-telegram"}
	command, _, stderr := onboardingTestCommand("y\n")
	services := onboardingTestServices(decision)
	services.dismissed = func(string) (bool, error) { return true, nil }
	maybeOnboardSkill(command, services)
	if stderr.Len() != 0 { t.Fatalf("stderr=%q", stderr.String()) }
}

func TestMaybeOnboardSkillEOFChangesNothing(t *testing.T) {
	decision := skills.OnboardingDecision{Action: skills.OnboardingPromptGlobal, Target: "/global/agent-telegram"}
	command, _, _ := onboardingTestCommand("y")
	services := onboardingTestServices(decision)
	installed := false
	recorded := false
	services.install = func(string, string, bool) (string, error) { installed = true; return "", nil }
	services.recordDismissal = func(string) error { recorded = true; return nil }
	maybeOnboardSkill(command, services)
	if installed || recorded { t.Fatalf("installed=%v recorded=%v", installed, recorded) }
}

func TestProjectInstallFailureDoesNotPromptGlobal(t *testing.T) {
	decision := skills.OnboardingDecision{Action: skills.OnboardingInstallProject, InstallDir: "/project", Target: "/project/agent-telegram"}
	command, _, stderr := onboardingTestCommand("y\n")
	services := onboardingTestServices(decision)
	services.install = func(string, string, bool) (string, error) { return "", errors.New("read-only") }
	maybeOnboardSkill(command, services)
	if !strings.Contains(stderr.String(), "could not install project Codex skill") || strings.Contains(stderr.String(), "globally") {
		t.Fatalf("stderr=%q", stderr.String())
	}
}

func TestMaybeOnboardSkillResolverFailureWarnsWithoutInstall(t *testing.T) {
	command, _, stderr := onboardingTestCommand("y\n")
	services := onboardingTestServices(skills.OnboardingDecision{})
	services.resolve = func(string, string) (skills.OnboardingDecision, error) {
		return skills.OnboardingDecision{}, errors.New("malformed project path")
	}
	installed := false
	services.install = func(string, string, bool) (string, error) { installed = true; return "", nil }
	maybeOnboardSkill(command, services)
	if installed || !strings.Contains(stderr.String(), "could not inspect Codex skill locations") {
		t.Fatalf("installed=%v stderr=%q", installed, stderr.String())
	}
}

func TestMaybeOnboardSkillGetwdFailureIsSilent(t *testing.T) {
	command, _, stderr := onboardingTestCommand("y\n")
	services := onboardingTestServices(skills.OnboardingDecision{})
	services.getwd = func() (string, error) { return "", errors.New("cwd unavailable") }
	maybeOnboardSkill(command, services)
	if stderr.Len() != 0 { t.Fatalf("stderr=%q", stderr.String()) }
}

func TestMaybeOnboardSkillReadFailureChangesNothing(t *testing.T) {
	decision := skills.OnboardingDecision{Action: skills.OnboardingPromptGlobal, Target: "/global/agent-telegram"}
	command, _, stderr := onboardingTestCommand("")
	command.SetIn(failingReader{})
	services := onboardingTestServices(decision)
	installed := false
	recorded := false
	services.install = func(string, string, bool) (string, error) { installed = true; return "", nil }
	services.recordDismissal = func(string) error { recorded = true; return nil }
	maybeOnboardSkill(command, services)
	if installed || recorded || !strings.Contains(stderr.String(), "could not read skill prompt response") {
		t.Fatalf("installed=%v recorded=%v stderr=%q", installed, recorded, stderr.String())
	}
}

func TestMaybeOnboardSkillGlobalFailuresAreNonFatal(t *testing.T) {
	decision := skills.OnboardingDecision{Action: skills.OnboardingPromptGlobal, InstallDir: "/global", Target: "/global/agent-telegram"}
	tests := []struct {
		name string
		answer string
		mutate func(*onboardingServices)
		want string
	}{
		{"preference-read", "y\n", func(s *onboardingServices) {
			s.dismissed = func(string) (bool, error) { return false, errors.New("read-only") }
		}, "could not read global skill preference"},
		{"install", "y\n", func(s *onboardingServices) {
			s.install = func(string, string, bool) (string, error) { return "", errors.New("read-only") }
		}, "agent-telegram skills install agent-telegram"},
		{"dismissal-write", "n\n", func(s *onboardingServices) {
			s.recordDismissal = func(string) error { return errors.New("read-only") }
		}, "could not save global skill preference"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			command, _, stderr := onboardingTestCommand(test.answer)
			services := onboardingTestServices(decision)
			test.mutate(&services)
			maybeOnboardSkill(command, services)
			if !strings.Contains(stderr.String(), test.want) { t.Fatalf("stderr=%q", stderr.String()) }
		})
	}
}

func TestNonInteractiveRootOutputRemainsExact(t *testing.T) {
	command, stdout, stderr := onboardingTestCommand("")
	services := onboardingTestServices(skills.OnboardingDecision{})
	services.isTerminal = func(any) bool { return false }
	if err := runRootWithServices(command, nil, services); err != nil { t.Fatal(err) }
	if stdout.String() != rootWelcome || stderr.Len() != 0 { t.Fatalf("stdout=%q stderr=%q", stdout.String(), stderr.String()) }
}

func TestSubcommandDoesNotRunRootOnboarding(t *testing.T) {
	called := false
	services := onboardingTestServices(skills.OnboardingDecision{})
	services.getwd = func() (string, error) { called = true; return "/repo", nil }
	root := &cobra.Command{
		Use: "agent-telegram",
		RunE: func(command *cobra.Command, args []string) error {
			return runRootWithServices(command, args, services)
		},
	}
	root.AddCommand(&cobra.Command{Use: "status", Run: func(*cobra.Command, []string) {}})
	root.SetArgs([]string{"status"})
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})
	if err := root.Execute(); err != nil { t.Fatal(err) }
	if called { t.Fatal("subcommand ran root onboarding") }
}
```

- [ ] **Step 4: Implement root onboarding orchestration**

Add `bufio`, `io`, `strings`, and `agent-telegram/internal/skills` imports to `cmd/root.go`, then add:

```go
const bundledSkillName = "agent-telegram"

type onboardingServices struct {
	isTerminal func(any) bool
	getwd func() (string, error)
	resolve func(string, string) (skills.OnboardingDecision, error)
	install func(string, string, bool) (string, error)
	dismissed func(string) (bool, error)
	recordDismissal func(string) error
	ci string
}

func defaultOnboardingServices() onboardingServices {
	return onboardingServices{
		isTerminal: isTerminal, getwd: os.Getwd, resolve: skills.ResolveOnboarding,
		install: skills.Install, dismissed: skills.GlobalPromptDismissed,
		recordDismissal: skills.RecordGlobalPromptDismissal, ci: os.Getenv("CI"),
	}
}

func runRoot(cmd *cobra.Command, args []string) error {
	return runRootWithServices(cmd, args, defaultOnboardingServices())
}

func runRootWithServices(cmd *cobra.Command, _ []string, services onboardingServices) error {
	if _, err := fmt.Fprint(cmd.OutOrStdout(), rootWelcome); err != nil { return err }
	maybeOnboardSkill(cmd, services)
	return nil
}

func maybeOnboardSkill(cmd *cobra.Command, services onboardingServices) {
	if services.ci != "" || cmd.Flags().NFlag() != 0 || !services.isTerminal(cmd.InOrStdin()) || !services.isTerminal(cmd.OutOrStdout()) { return }
	cwd, err := services.getwd()
	if err != nil { return }
	decision, err := services.resolve(bundledSkillName, cwd)
	if err != nil {
		fmt.Fprintf(cmd.ErrOrStderr(), "Warning: could not inspect Codex skill locations: %v\n", err)
		return
	}
	switch decision.Action {
	case skills.OnboardingNone:
		return
	case skills.OnboardingInstallProject:
		installed, err := services.install(bundledSkillName, decision.InstallDir, false)
		if err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "Warning: could not install project Codex skill: %v\n", err)
			return
		}
		fmt.Fprintf(cmd.ErrOrStderr(), "Installed project Codex skill at %s\n", installed)
	case skills.OnboardingPromptGlobal:
		promptGlobalSkill(cmd, services, decision)
	}
}

func promptGlobalSkill(cmd *cobra.Command, services onboardingServices, decision skills.OnboardingDecision) {
	dismissed, err := services.dismissed(decision.Target)
	if err != nil {
		fmt.Fprintf(cmd.ErrOrStderr(), "Warning: could not read global skill preference: %v\n", err)
		return
	}
	if dismissed { return }
	fmt.Fprintf(cmd.ErrOrStderr(), "\nInstall the Agent Telegram skill globally for Codex?\nTarget: %s\nThis makes the skill available across projects. [y/N] ", decision.Target)
	answer, err := bufio.NewReader(cmd.InOrStdin()).ReadString('\n')
	if err != nil {
		if err != io.EOF { fmt.Fprintf(cmd.ErrOrStderr(), "\nWarning: could not read skill prompt response: %v\n", err) }
		return
	}
	switch strings.ToLower(strings.TrimSpace(answer)) {
	case "y", "yes":
		installed, err := services.install(bundledSkillName, decision.InstallDir, false)
		if err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "Warning: could not install global Codex skill: %v\nInstall it later with: agent-telegram skills install agent-telegram\n", err)
			return
		}
		fmt.Fprintf(cmd.ErrOrStderr(), "Installed global Codex skill at %s\n", installed)
	default:
		if err := services.recordDismissal(decision.Target); err != nil { fmt.Fprintf(cmd.ErrOrStderr(), "Warning: could not save global skill preference: %v\n", err) }
	}
}

func isTerminal(value any) bool {
	file, ok := value.(*os.File)
	if !ok { return false }
	info, err := file.Stat()
	return err == nil && info.Mode()&os.ModeCharDevice != 0
}
```

- [ ] **Step 5: Format, test, and commit**

```bash
gofmt -w cmd/root.go cmd/root_test.go
go test ./cmd -run 'TestRoot|TestMaybeOnboardSkill|TestProjectInstallFailure|TestNonInteractiveRoot' -count=1
go test ./cmd ./internal/skills -count=1
git add cmd/root.go cmd/root_test.go
git commit -m "Add project-aware Codex skill onboarding"
```

Expected: project installation never reads stdin, global installation requires consent, and root output compatibility is preserved.

### Task 4: Canonical path and onboarding documentation

**Files:**
- Modify: `cmd/sys/skills.go`
- Modify: `internal/docs/docs.go`
- Modify: `internal/docs/docs_test.go`
- Modify: `README.md`
- Modify: `DEVELOPMENT.md`

**Interfaces:**
- Consumes: `skills.DefaultInstallDir()` and `docs.GenerateLLMMarkdown`.
- Produces: help and generated guidance matching `.agents/skills` behavior.

- [ ] **Step 1: Add a failing generated-guidance test**

Append to `internal/docs/docs_test.go`:

```go
func TestGenerateLLMMarkdownMentionsProjectAwareSkillOnboarding(t *testing.T) {
	root := &cobra.Command{Use: "agent-telegram"}
	body := GenerateLLMMarkdown(root)
	for _, want := range []string{"existing project `.agents/skills`", "global `$HOME/.agents/skills` installation requires consent", "agent-telegram skills install agent-telegram"} {
		if !strings.Contains(body, want) { t.Fatalf("generated docs missing %q:\n%s", want, body) }
	}
}
```

- [ ] **Step 2: Verify the docs test fails**

```bash
go test ./internal/docs -run TestGenerateLLMMarkdownMentionsProjectAwareSkillOnboarding -count=1
```

Expected: failure for the new project-aware phrase.

- [ ] **Step 3: Update help and generated guidance**

Use this `Long` text in `cmd/sys/skills.go`:

```go
Long: `List and install agent skills bundled with agent-telegram.

Skills help AI agents discover best practices for using the CLI. The explicit
install command defaults to $HOME/.agents/skills; use --target to override it.`,
```

Change the `--target` help:

```go
SkillsInstallCmd.Flags().StringVar(&skillInstallTarget, "target", "", "Skill install directory (default: $HOME/.agents/skills)")
```

Replace the generated skill guidance in `internal/docs/docs.go`:

```go
b.WriteString("- An interactive no-argument run auto-installs into an existing project `.agents/skills`; global `$HOME/.agents/skills` installation requires consent. For manual setup, use `agent-telegram skills install agent-telegram`.\n")
```

- [ ] **Step 4: Update README and DEVELOPMENT**

After README Quick Start, add:

````markdown
On an interactive run with no arguments, `agent-telegram` automatically adds
its bundled skill to the nearest existing project `.agents/skills`. It never
creates a project skill directory. If no project skill directory exists, it
asks before installing globally to `$HOME/.agents/skills`.

Manual installation remains available:

```bash
agent-telegram skills install agent-telegram
```
````

Replace the legacy skill-location paragraph in `DEVELOPMENT.md` with:

```markdown
Bundled Codex skills live under `internal/skills/bundled` and are exposed via
`agent-telegram skills list`, `agent-telegram skills install`, CLI/HTTP
manifests, and generated docs. The explicit installer defaults to the canonical
user directory `$HOME/.agents/skills`; `--target` supports custom and legacy
locations. Interactive onboarding auto-installs only into an existing project
`.agents/skills` and requires consent for global installation.
```

- [ ] **Step 5: Format, verify, and commit documentation**

```bash
gofmt -w cmd/sys/skills.go internal/docs/docs.go internal/docs/docs_test.go
go test ./internal/docs ./cmd/sys -count=1
go run . docs check --target README.md
git diff --check
git add cmd/sys/skills.go internal/docs/docs.go internal/docs/docs_test.go README.md DEVELOPMENT.md
git commit -m "Document project-aware skill onboarding"
```

Expected: docs tests and generated-doc check pass; commit succeeds.

### Task 5: Full verification and required release cycle

**Files:**
- Verify all files changed by Tasks 1-4.
- Do not modify unrelated working-tree files.

**Interfaces:**
- Consumes: completed onboarding behavior and the release workflow in `CLAUDE.md`.
- Produces: tested, pushed, released, and locally installed patch version.

- [ ] **Step 1: Run complete verification**

```bash
go test ./... -count=1
go build ./...
go vet ./...
go run . docs check --target README.md
```

Expected: all packages pass; build and vet exit zero; docs check returns `"ok": true`.

- [ ] **Step 2: Inspect final scope**

```bash
git diff --check
git status --short
git log -8 --oneline
```

Expected: no whitespace errors, no uncommitted implementation files, and focused Task 1-4 commits at the tip of `main`.

- [ ] **Step 3: Push and create the patch release**

```bash
git push origin main
make release
```

Expected: `main` and the next `vX.Y.Z` patch tag are pushed.

- [ ] **Step 4: Watch the release workflow**

```bash
gh run list --limit 1 --json databaseId,displayTitle,status,conclusion,url
RUN_ID=$(gh run list --limit 1 --json databaseId --jq '.[0].databaseId')
gh run watch "$RUN_ID" --exit-status
```

Expected: newest release workflow finishes with `success`.

- [ ] **Step 5: Install and smoke-test the release**

```bash
VERSION=$(git describe --tags --abbrev=0 | sed 's/^v//')
npm install -g "agent-telegram@$VERSION"
CI=1 agent-telegram
agent-telegram skills path
agent-telegram skills list
```

Expected: npm installs the release; CI invocation never prompts; `skills path` reports `$HOME/.agents/skills`; bundled skill remains listed.

- [ ] **Step 6: Report release evidence**

Include the version, verification commands, GitHub Actions URL, npm result, and preserved unrelated worktree state. Do not claim a real-TTY project auto-install unless it was exercised with temporary HOME and repository directories.
