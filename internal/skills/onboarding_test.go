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
		name  string
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
	data, err := os.ReadFile(marker) //nolint:gosec // fixed path below the temporary HOME
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

func TestResolveOnboardingTreatsAllTargetKindsAsExisting(t *testing.T) {
	creators := []struct {
		name   string
		create func(t *testing.T, target string)
	}{
		{"file", func(t *testing.T, target string) {
			if err := os.WriteFile(target, []byte("owned"), 0o600); err != nil {
				t.Fatal(err)
			}
		}},
		{"directory", func(t *testing.T, target string) { makeDir(t, target) }},
		{"valid-symlink", func(t *testing.T, target string) {
			real := target + "-real"
			makeDir(t, real)
			if err := os.Symlink(real, target); err != nil {
				t.Skipf("symlink unavailable: %v", err)
			}
		}},
		{"dangling-symlink", func(t *testing.T, target string) {
			if err := os.Symlink(target+"-missing", target); err != nil {
				t.Skipf("symlink unavailable: %v", err)
			}
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
			if err != nil || decision.Action != OnboardingNone {
				t.Fatalf("decision=%+v err=%v", decision, err)
			}
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
	if err := os.Symlink(realSkills, filepath.Join(repo, ".agents", "skills")); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}
	t.Setenv("HOME", home)
	t.Setenv("CODEX_HOME", "")

	decision, err := ResolveOnboarding("agent-telegram", repo)
	if err != nil || decision.Action != OnboardingInstallProject {
		t.Fatalf("decision=%+v err=%v", decision, err)
	}
}

func TestResolveOnboardingRejectsDanglingProjectSkillsSymlink(t *testing.T) {
	home := t.TempDir()
	repo := filepath.Join(t.TempDir(), "repo")
	makeDir(t, filepath.Join(repo, ".git"))
	makeDir(t, filepath.Join(repo, ".agents"))
	if err := os.Symlink(filepath.Join(repo, "missing"), filepath.Join(repo, ".agents", "skills")); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}
	t.Setenv("HOME", home)
	t.Setenv("CODEX_HOME", "")

	if _, err := ResolveOnboarding("agent-telegram", repo); err == nil {
		t.Fatal("dangling project skills symlink should be reported as malformed")
	}
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
