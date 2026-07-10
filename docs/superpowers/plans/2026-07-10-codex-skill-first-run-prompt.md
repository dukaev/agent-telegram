# Codex Skill First-Run Prompt Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Offer an explicit, one-time, interactive installation of the bundled `agent-telegram` Codex skill without changing non-interactive CLI behavior or overwriting user files.

**Architecture:** Add filesystem-only onboarding state helpers to `internal/skills`, then keep terminal detection, prompt rendering, and response handling in `cmd/root.go`. Inject prompt dependencies in tests so consent, decline, errors, TTY gating, flags, and CI behavior are deterministic without touching the developer's home directory.

**Tech Stack:** Go 1.25.4, Cobra 1.10.2, standard library `bufio`, `io`, `os`, and `path/filepath`, existing embedded skill installer.

## Global Constraints

- Require explicit `y` or `yes` consent before installing.
- Never pass `force=true` from onboarding and never inspect or rewrite an existing target entry.
- Prompt only for the root command with no positional arguments and no flags, with terminal stdin and stdout, and with an empty `CI` environment variable.
- Write prompt and status messages to stderr; preserve the existing root welcome text on stdout.
- Store target-specific dismissal state at `~/.agent-telegram/skill-prompt-dismissed` with file mode `0600` and parent mode `0700`.
- Treat onboarding as best-effort: all onboarding failures leave the root command successful.
- Add no third-party dependency.
- Use temporary `HOME` and `CODEX_HOME` directories in filesystem tests.

---

## File Structure

- Create `internal/skills/onboarding.go`: resolve the normalized target, detect existing entries with `os.Lstat`, compare the dismissal marker, and write the marker securely.
- Create `internal/skills/onboarding_test.go`: cover target resolution, existing files/directories/symlinks, target-specific dismissal, permissions, and filesystem failures.
- Modify `cmd/root.go`: run the best-effort prompt after the welcome text and expose a small dependency bundle for deterministic tests.
- Modify `cmd/root_test.go`: test TTY/CI/flag eligibility, affirmative installation, decline recording, EOF, warnings, and unchanged non-interactive output.
- Modify `internal/docs/docs.go` and `internal/docs/docs_test.go`: document and test generated agent guidance.
- Modify `README.md`: explain the optional first-run offer and explicit installation command.

### Task 1: Filesystem onboarding state

**Files:**
- Create: `internal/skills/onboarding.go`
- Create: `internal/skills/onboarding_test.go`

**Interfaces:**
- Consumes: `DefaultInstallDir() string`, `paths.ConfigDir() (string, error)`, and `paths.EnsureConfigDir() (string, error)`.
- Produces: `DefaultTarget(name string) (string, error)`, `ShouldOffer(target string) (bool, error)`, and `RecordDismissal(target string) error`.

- [ ] **Step 1: Write failing target and existing-entry tests**

Create `internal/skills/onboarding_test.go`:

```go
package skills

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultTargetUsesCodexHome(t *testing.T) {
	home := t.TempDir()
	codexHome := filepath.Join(home, "custom-codex")
	t.Setenv("HOME", home)
	t.Setenv("CODEX_HOME", codexHome)
	target, err := DefaultTarget("agent-telegram")
	if err != nil {
		t.Fatal(err)
	}
	want, _ := filepath.Abs(filepath.Join(codexHome, "skills", "agent-telegram"))
	if target != want {
		t.Fatalf("target = %q, want %q", target, want)
	}
}

func TestDefaultTargetRequiresHomeWithoutCodexHome(t *testing.T) {
	t.Setenv("HOME", "")
	t.Setenv("CODEX_HOME", "")
	if _, err := DefaultTarget("agent-telegram"); err == nil {
		t.Fatal("missing home should prevent an onboarding target")
	}
}

func TestShouldOfferOnlyWhenTargetIsAbsent(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("CODEX_HOME", filepath.Join(home, ".codex"))
	target, err := DefaultTarget("agent-telegram")
	if err != nil {
		t.Fatal(err)
	}
	if offer, err := ShouldOffer(target); err != nil || !offer {
		t.Fatalf("absent target: offer=%v err=%v", offer, err)
	}

	entries := []struct {
		name   string
		create func(string) error
	}{
		{"directory", func(path string) error { return os.MkdirAll(path, 0o700) }},
		{"file", func(path string) error {
			if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
				return err
			}
			return os.WriteFile(path, []byte("owned"), 0o600)
		}},
		{"dangling-symlink", func(path string) error {
			if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
				return err
			}
			return os.Symlink(filepath.Join(home, "missing"), path)
		}},
	}

	for _, entry := range entries {
		t.Run(entry.name, func(t *testing.T) {
			if err := os.RemoveAll(filepath.Dir(target)); err != nil {
				t.Fatal(err)
			}
			if err := entry.create(target); err != nil {
				t.Fatal(err)
			}
			if offer, err := ShouldOffer(target); err != nil || offer {
				t.Fatalf("existing target: offer=%v err=%v", offer, err)
			}
		})
	}
}
```

- [ ] **Step 2: Verify the tests fail for the missing API**

```bash
go test ./internal/skills -run 'TestDefaultTarget|TestShouldOfferOnlyWhenTargetIsAbsent' -count=1
```

Expected: build failure containing `undefined: DefaultTarget` and `undefined: ShouldOffer`.

- [ ] **Step 3: Add failing dismissal tests**

Append:

```go
func TestRecordDismissalSuppressesOnlyMatchingTarget(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("CODEX_HOME", filepath.Join(home, "codex-a"))
	targetA, err := DefaultTarget("agent-telegram")
	if err != nil {
		t.Fatal(err)
	}
	if err := RecordDismissal(targetA); err != nil {
		t.Fatal(err)
	}
	marker := filepath.Join(home, ".agent-telegram", "skill-prompt-dismissed")
	data, err := os.ReadFile(marker)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != targetA+"\n" {
		t.Fatalf("marker = %q", data)
	}
	info, err := os.Stat(marker)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("marker mode = %v", info.Mode().Perm())
	}
	dirInfo, err := os.Stat(filepath.Dir(marker))
	if err != nil {
		t.Fatal(err)
	}
	if dirInfo.Mode().Perm() != 0o700 {
		t.Fatalf("marker directory mode = %v", dirInfo.Mode().Perm())
	}
	if offer, err := ShouldOffer(targetA); err != nil || offer {
		t.Fatalf("matching marker: offer=%v err=%v", offer, err)
	}

	t.Setenv("CODEX_HOME", filepath.Join(home, "codex-b"))
	targetB, err := DefaultTarget("agent-telegram")
	if err != nil {
		t.Fatal(err)
	}
	if offer, err := ShouldOffer(targetB); err != nil || !offer {
		t.Fatalf("different target: offer=%v err=%v", offer, err)
	}
}

func TestRecordDismissalRestoresMode0600(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	dir := filepath.Join(home, ".agent-telegram")
	marker := filepath.Join(dir, "skill-prompt-dismissed")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(marker, []byte("old\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := RecordDismissal(filepath.Join(home, "target")); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(marker)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("marker mode = %v", info.Mode().Perm())
	}
}
```

- [ ] **Step 4: Verify dismissal tests fail**

```bash
go test ./internal/skills -run TestRecordDismissal -count=1
```

Expected: build failure containing `undefined: RecordDismissal`.

- [ ] **Step 5: Implement the state helpers**

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

const dismissalFileName = "skill-prompt-dismissed"

func DefaultTarget(name string) (string, error) {
	if strings.TrimSpace(os.Getenv("CODEX_HOME")) == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("resolve skill target home: %w", err)
		}
		if home == "" {
			return "", fmt.Errorf("resolve skill target home: home directory is empty")
		}
	}
	target, err := filepath.Abs(filepath.Join(DefaultInstallDir(), name))
	if err != nil {
		return "", fmt.Errorf("resolve skill target: %w", err)
	}
	return filepath.Clean(target), nil
}

func ShouldOffer(target string) (bool, error) {
	if _, err := os.Lstat(target); err == nil {
		return false, nil
	} else if !os.IsNotExist(err) {
		return false, fmt.Errorf("inspect skill target: %w", err)
	}
	marker, err := dismissalPath()
	if err != nil {
		return false, err
	}
	data, err := os.ReadFile(marker) //nolint:gosec // fixed per-user state path
	if err != nil {
		if os.IsNotExist(err) {
			return true, nil
		}
		return false, fmt.Errorf("read skill prompt dismissal: %w", err)
	}
	return strings.TrimSpace(string(data)) != filepath.Clean(target), nil
}

func RecordDismissal(target string) error {
	dir, err := paths.EnsureConfigDir()
	if err != nil {
		return err
	}
	if err := os.Chmod(dir, 0o700); err != nil {
		return fmt.Errorf("secure config directory: %w", err)
	}
	marker := filepath.Join(dir, dismissalFileName)
	file, err := os.CreateTemp(dir, ".skill-prompt-dismissed-*")
	if err != nil {
		return fmt.Errorf("create skill prompt dismissal: %w", err)
	}
	temporary := file.Name()
	defer os.Remove(temporary)
	if err := file.Chmod(0o600); err != nil {
		_ = file.Close()
		return fmt.Errorf("secure skill prompt dismissal: %w", err)
	}
	if _, err := fmt.Fprintln(file, filepath.Clean(target)); err != nil {
		_ = file.Close()
		return fmt.Errorf("write skill prompt dismissal: %w", err)
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("close skill prompt dismissal: %w", err)
	}
	if err := os.Remove(marker); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove previous skill prompt dismissal: %w", err)
	}
	if err := os.Rename(temporary, marker); err != nil {
		return fmt.Errorf("replace skill prompt dismissal: %w", err)
	}
	return nil
}

func dismissalPath() (string, error) {
	dir, err := paths.ConfigDir()
	if err != nil {
		return "", fmt.Errorf("resolve skill prompt dismissal: %w", err)
	}
	return filepath.Join(dir, dismissalFileName), nil
}
```

- [ ] **Step 6: Format, test, and commit**

```bash
gofmt -w internal/skills/onboarding.go internal/skills/onboarding_test.go
go test ./internal/skills -count=1
git add internal/skills/onboarding.go internal/skills/onboarding_test.go
git commit -m "Add Codex skill onboarding state"
```

Expected: tests report `ok agent-telegram/internal/skills`; commit succeeds.

### Task 2: Interactive root prompt

**Files:**
- Modify: `cmd/root.go`
- Modify: `cmd/root_test.go`

**Interfaces:**
- Consumes: the three Task 1 helpers and `skills.Install(string, string, bool) (string, error)`.
- Produces: `skillPromptServices`, `defaultSkillPromptServices()`, `maybeOfferSkill(*cobra.Command, skillPromptServices)`, and `isTerminal(any) bool`.

- [ ] **Step 1: Add failing consent and decline tests**

Extend `cmd/root_test.go` imports with `errors`, then append:

```go
func promptTestCommand(input string) (*cobra.Command, *bytes.Buffer, *bytes.Buffer) {
	command := &cobra.Command{Use: "agent-telegram"}
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	command.SetIn(strings.NewReader(input))
	command.SetOut(stdout)
	command.SetErr(stderr)
	return command, stdout, stderr
}

func promptTestServices() skillPromptServices {
	return skillPromptServices{
		isTerminal: func(any) bool { return true },
		target: func(string) (string, error) {
			return "/tmp/codex/skills/agent-telegram", nil
		},
		shouldOffer: func(string) (bool, error) { return true, nil },
		install: func(string, string, bool) (string, error) {
			return "/tmp/codex/skills/agent-telegram", nil
		},
		dismiss: func(string) error { return nil },
	}
}

func TestMaybeOfferSkillInstallsOnlyOnExplicitConsent(t *testing.T) {
	for _, answer := range []string{"y\n", "yes\n", "Y\n", "YES\n"} {
		t.Run(strings.TrimSpace(answer), func(t *testing.T) {
			command, _, stderr := promptTestCommand(answer)
			services := promptTestServices()
			installed := false
			services.install = func(name, target string, force bool) (string, error) {
				installed = true
				if name != "agent-telegram" || target != "" || force {
					t.Fatalf("install args = %q, %q, %v", name, target, force)
				}
				return "/tmp/codex/skills/agent-telegram", nil
			}
			maybeOfferSkill(command, services)
			if !installed || !strings.Contains(stderr.String(), "Skill installed at") {
				t.Fatalf("installed=%v stderr=%q", installed, stderr.String())
			}
		})
	}
}

func TestMaybeOfferSkillRecordsCompletedDeclines(t *testing.T) {
	for _, answer := range []string{"\n", "n\n", "no\n", "later\n"} {
		t.Run(strings.TrimSpace(answer), func(t *testing.T) {
			command, _, _ := promptTestCommand(answer)
			services := promptTestServices()
			installed := false
			dismissed := ""
			services.install = func(string, string, bool) (string, error) {
				installed = true
				return "", nil
			}
			services.dismiss = func(target string) error {
				dismissed = target
				return nil
			}
			maybeOfferSkill(command, services)
			if installed || dismissed != "/tmp/codex/skills/agent-telegram" {
				t.Fatalf("installed=%v dismissed=%q", installed, dismissed)
			}
		})
	}
}
```

- [ ] **Step 2: Verify missing prompt symbols**

```bash
go test ./cmd -run TestMaybeOfferSkill -count=1
```

Expected: build failure containing `undefined: skillPromptServices` and `undefined: maybeOfferSkill`.

- [ ] **Step 3: Add eligibility, EOF, error, and compatibility tests**

Append:

```go
func TestMaybeOfferSkillSkipsIneligibleInvocations(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*cobra.Command, *skillPromptServices)
	}{
		{"ci", func(_ *cobra.Command, services *skillPromptServices) { services.ci = "true" }},
		{"non-terminal-stdin", func(_ *cobra.Command, services *skillPromptServices) {
			services.isTerminal = func(any) bool { return false }
		}},
		{"non-terminal-stdout", func(_ *cobra.Command, services *skillPromptServices) {
			calls := 0
			services.isTerminal = func(any) bool {
				calls++
				return calls == 1
			}
		}},
		{"root-flag", func(command *cobra.Command, _ *skillPromptServices) {
			command.Flags().Bool("quiet", false, "quiet")
			if err := command.Flags().Set("quiet", "true"); err != nil {
				t.Fatal(err)
			}
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			command, _, stderr := promptTestCommand("y\n")
			services := promptTestServices()
			called := false
			services.target = func(string) (string, error) {
				called = true
				return "", nil
			}
			test.mutate(command, &services)
			maybeOfferSkill(command, services)
			if called || stderr.Len() != 0 {
				t.Fatalf("called=%v stderr=%q", called, stderr.String())
			}
		})
	}
}

func TestMaybeOfferSkillTreatsEOFAsNoDecision(t *testing.T) {
	command, _, _ := promptTestCommand("y")
	services := promptTestServices()
	installed := false
	dismissed := false
	services.install = func(string, string, bool) (string, error) {
		installed = true
		return "", nil
	}
	services.dismiss = func(string) error {
		dismissed = true
		return nil
	}
	maybeOfferSkill(command, services)
	if installed || dismissed {
		t.Fatalf("EOF changed state: installed=%v dismissed=%v", installed, dismissed)
	}
}

type failingReader struct{}

func (failingReader) Read([]byte) (int, error) {
	return 0, errors.New("input failed")
}

func TestMaybeOfferSkillReadFailureChangesNoState(t *testing.T) {
	command, _, stderr := promptTestCommand("")
	command.SetIn(failingReader{})
	services := promptTestServices()
	installed := false
	dismissed := false
	services.install = func(string, string, bool) (string, error) {
		installed = true
		return "", nil
	}
	services.dismiss = func(string) error {
		dismissed = true
		return nil
	}
	maybeOfferSkill(command, services)
	if installed || dismissed || !strings.Contains(stderr.String(), "could not read skill prompt response") {
		t.Fatalf("installed=%v dismissed=%v stderr=%q", installed, dismissed, stderr.String())
	}
}

func TestMaybeOfferSkillSilentlySkipsTargetResolutionFailure(t *testing.T) {
	command, _, stderr := promptTestCommand("y\n")
	services := promptTestServices()
	services.target = func(string) (string, error) {
		return "", errors.New("home unavailable")
	}
	maybeOfferSkill(command, services)
	if stderr.Len() != 0 {
		t.Fatalf("stderr=%q", stderr.String())
	}
}

func TestMaybeOfferSkillReportsBestEffortErrors(t *testing.T) {
	tests := []struct {
		name   string
		answer string
		mutate func(*skillPromptServices)
		want   string
	}{
		{"inspect", "y\n", func(s *skillPromptServices) {
			s.shouldOffer = func(string) (bool, error) { return false, errors.New("denied") }
		}, "Warning: could not inspect Codex skill onboarding"},
		{"install", "y\n", func(s *skillPromptServices) {
			s.install = func(string, string, bool) (string, error) { return "", errors.New("read-only") }
		}, "agent-telegram skills install agent-telegram"},
		{"dismiss", "n\n", func(s *skillPromptServices) {
			s.dismiss = func(string) error { return errors.New("read-only") }
		}, "Warning: could not save skill prompt preference"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			command, _, stderr := promptTestCommand(test.answer)
			services := promptTestServices()
			test.mutate(&services)
			maybeOfferSkill(command, services)
			if !strings.Contains(stderr.String(), test.want) {
				t.Fatalf("stderr=%q, want %q", stderr.String(), test.want)
			}
		})
	}
}

func TestNonInteractiveRootOutputIsUnchanged(t *testing.T) {
	command, stdout, stderr := promptTestCommand("")
	services := promptTestServices()
	services.isTerminal = func(any) bool { return false }
	if err := runRootWithServices(command, nil, services); err != nil {
		t.Fatal(err)
	}
	if stdout.String() != rootWelcome || stderr.Len() != 0 {
		t.Fatalf("stdout=%q stderr=%q", stdout.String(), stderr.String())
	}
}

func TestSubcommandDoesNotRunRootOnboarding(t *testing.T) {
	services := promptTestServices()
	called := false
	services.target = func(string) (string, error) {
		called = true
		return "", nil
	}
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
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	if called {
		t.Fatal("subcommand ran root onboarding")
	}
}
```

- [ ] **Step 4: Implement the prompt**

Add `bufio`, `io`, `strings`, and `agent-telegram/internal/skills` imports to `cmd/root.go`, then add:

```go
const bundledSkillName = "agent-telegram"

type skillPromptServices struct {
	isTerminal  func(any) bool
	target      func(string) (string, error)
	shouldOffer func(string) (bool, error)
	install     func(string, string, bool) (string, error)
	dismiss     func(string) error
	ci          string
}

func defaultSkillPromptServices() skillPromptServices {
	return skillPromptServices{
		isTerminal: isTerminal, target: skills.DefaultTarget,
		shouldOffer: skills.ShouldOffer, install: skills.Install,
		dismiss: skills.RecordDismissal, ci: os.Getenv("CI"),
	}
}

func runRoot(cmd *cobra.Command, args []string) error {
	return runRootWithServices(cmd, args, defaultSkillPromptServices())
}

func runRootWithServices(cmd *cobra.Command, _ []string, services skillPromptServices) error {
	if _, err := fmt.Fprint(cmd.OutOrStdout(), rootWelcome); err != nil {
		return err
	}
	maybeOfferSkill(cmd, services)
	return nil
}

func maybeOfferSkill(cmd *cobra.Command, services skillPromptServices) {
	if services.ci != "" || cmd.Flags().NFlag() != 0 ||
		!services.isTerminal(cmd.InOrStdin()) || !services.isTerminal(cmd.OutOrStdout()) {
		return
	}
	target, err := services.target(bundledSkillName)
	if err != nil {
		return
	}
	offer, err := services.shouldOffer(target)
	if err != nil {
		fmt.Fprintf(cmd.ErrOrStderr(), "Warning: could not inspect Codex skill onboarding: %v\n", err)
		return
	}
	if !offer {
		return
	}
	fmt.Fprintf(cmd.ErrOrStderr(), "\nInstall the Agent Telegram skill for Codex?\nTarget: %s\nThis adds CLI usage instructions for Codex. [y/N] ", target)
	answer, err := bufio.NewReader(cmd.InOrStdin()).ReadString('\n')
	if err != nil {
		if err != io.EOF {
			fmt.Fprintf(cmd.ErrOrStderr(), "\nWarning: could not read skill prompt response: %v\n", err)
		}
		return
	}
	switch strings.ToLower(strings.TrimSpace(answer)) {
	case "y", "yes":
		installedPath, err := services.install(bundledSkillName, "", false)
		if err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "Warning: could not install Codex skill: %v\nInstall it later with: agent-telegram skills install agent-telegram\n", err)
			return
		}
		fmt.Fprintf(cmd.ErrOrStderr(), "Skill installed at %s\n", installedPath)
	default:
		if err := services.dismiss(target); err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "Warning: could not save skill prompt preference: %v\n", err)
		}
	}
}

func isTerminal(value any) bool {
	file, ok := value.(*os.File)
	if !ok {
		return false
	}
	info, err := file.Stat()
	return err == nil && info.Mode()&os.ModeCharDevice != 0
}
```

Keep the root welcome, `Execute`, flags, and command registration unchanged.

- [ ] **Step 5: Format, test, and commit**

```bash
gofmt -w cmd/root.go cmd/root_test.go
go test ./cmd -run 'TestRoot|TestMaybeOfferSkill|TestNonInteractiveRootOutput' -count=1
go test ./cmd ./internal/skills -count=1
git add cmd/root.go cmd/root_test.go
git commit -m "Offer Codex skill on first interactive run"
```

Expected: both packages report `ok`; tests confirm `install(..., force=false)` and unchanged non-interactive output. Cobra merges parsed root persistent flags into `cmd.Flags()` before `RunE`, so `NFlag()` covers the production global-flag path.

### Task 3: User and agent documentation

**Files:**
- Modify: `README.md`
- Modify: `internal/docs/docs.go`
- Modify: `internal/docs/docs_test.go`

**Interfaces:**
- Consumes: `docs.GenerateLLMMarkdown(rootCmd) string`.
- Produces: Quick Start copy and generated guidance explaining interactive and explicit installation.

- [ ] **Step 1: Add a failing generated-guidance test**

Append to `internal/docs/docs_test.go`:

```go
func TestGenerateLLMMarkdownMentionsSkillOnboarding(t *testing.T) {
	root := &cobra.Command{Use: "agent-telegram"}
	body := GenerateLLMMarkdown(root)
	for _, want := range []string{
		"interactive no-argument run may offer",
		"agent-telegram skills install agent-telegram",
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("generated docs missing %q:\n%s", want, body)
		}
	}
}
```

- [ ] **Step 2: Verify the new phrase is absent**

```bash
go test ./internal/docs -run TestGenerateLLMMarkdownMentionsSkillOnboarding -count=1
```

Expected: failure containing `generated docs missing "interactive no-argument run may offer"`.

- [ ] **Step 3: Update generated guidance and README**

Replace the current skill bullet in `internal/docs/docs.go` with:

```go
b.WriteString("- An interactive no-argument run may offer to install the bundled Codex skill; for automation or manual setup, use `agent-telegram skills install agent-telegram`.\n")
```

After the Quick Start code block in `README.md`, add:

````markdown
On the first interactive run with no arguments, `agent-telegram` may offer to
install its bundled Codex skill. The default answer is no, existing skill files
are never overwritten, and automated or manual setups can install it explicitly:

```bash
agent-telegram skills install agent-telegram
```
````

- [ ] **Step 4: Format, verify, and commit documentation**

```bash
gofmt -w internal/docs/docs.go internal/docs/docs_test.go
go test ./internal/docs -count=1
go run . docs check --target README.md
git add README.md internal/docs/docs.go internal/docs/docs_test.go
git commit -m "Document Codex skill onboarding"
```

Expected: docs tests pass, docs check prints JSON with `"ok": true`, and commit succeeds.

### Task 4: Full verification and required release cycle

**Files:**
- Verify all Task 1-3 changes.
- Preserve the pre-existing unstaged edit in the design spec unless the user separately authorizes changing it.

**Interfaces:**
- Consumes: the complete implementation and the repository release workflow in `CLAUDE.md`.
- Produces: a built, tested, pushed, released, and locally installed patch version.

- [ ] **Step 1: Run the complete local verification suite**

```bash
go test ./... -count=1
go build ./...
go vet ./...
go run . docs check --target README.md
```

Expected: all packages pass, build and vet exit zero, and docs check reports `"ok": true`.

- [ ] **Step 2: Inspect the final diff and state**

```bash
git diff --check
git status --short
git log -5 --oneline
```

Expected: no whitespace errors; only explicitly preserved pre-existing user changes may remain unstaged; implementation commits are at the tip of `main`.

- [ ] **Step 3: Push and create the patch release**

```bash
git push origin main
make release
```

Expected: `main` and the next `vX.Y.Z` patch tag are pushed successfully.

- [ ] **Step 4: Wait for the release workflow**

```bash
gh run list --limit 1 --json databaseId,headBranch,displayTitle,status,conclusion
RUN_ID=$(gh run list --limit 1 --json databaseId --jq '.[0].databaseId')
gh run watch "$RUN_ID" --exit-status
```

Expected: the newest release workflow completes with conclusion `success`.

- [ ] **Step 5: Install and smoke-test the npm release**

```bash
VERSION=$(git describe --tags --abbrev=0 | sed 's/^v//')
npm install -g "agent-telegram@$VERSION"
CI=1 agent-telegram
agent-telegram skills list
```

Expected: npm installs the new version; `CI=1 agent-telegram` prints the welcome without prompting; `skills list` includes `agent-telegram` and its explicit install command.

- [ ] **Step 6: Report exact release evidence**

Include the released version, verification commands, workflow URL, and preserved unrelated working-tree changes in the final handoff. Do not claim the interactive path was manually exercised unless it was run in a real TTY with temporary `HOME` and `CODEX_HOME`.
