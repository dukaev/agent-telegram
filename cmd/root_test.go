package cmd

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"agent-telegram/internal/skills"
)

func TestRootLandingPrioritizesAuthentication(t *testing.T) {
	var output bytes.Buffer
	command := &cobra.Command{}
	command.SetOut(&output)

	if err := runRoot(command, nil); err != nil {
		t.Fatal(err)
	}
	text := output.String()
	authIndex := strings.Index(text, "agent-telegram auth")
	afterSignInIndex := strings.Index(text, "After sign-in")
	if authIndex < 0 {
		t.Fatalf("root landing does not offer web auth:\n%s", text)
	}
	if afterSignInIndex < 0 || authIndex > afterSignInIndex {
		t.Fatalf("authentication must appear before post-login commands:\n%s", text)
	}
	for _, expected := range []string{
		"Sign in with a QR code in your browser",
		"agent-telegram --help",
	} {
		if !strings.Contains(text, expected) {
			t.Fatalf("root landing is missing %q:\n%s", expected, text)
		}
	}
}

type failingReader struct{}

func (failingReader) Read([]byte) (int, error) {
	return 0, errors.New("input failed")
}

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
		getwd:      func() (string, error) { return "/repo", nil },
		resolve: func(string, string) (skills.OnboardingDecision, error) {
			return decision, nil
		},
		install: func(string, string, bool) (string, error) { return decision.Target, nil },
		dismissed: func(string) (bool, error) {
			return false, nil
		},
		recordDismissal: func(string) error { return nil },
	}
}

func TestMaybeOnboardSkillAutoInstallsProjectWithoutReadingInput(t *testing.T) {
	decision := skills.OnboardingDecision{
		Action:     skills.OnboardingInstallProject,
		InstallDir: "/repo/.agents/skills",
		Target:     "/repo/.agents/skills/agent-telegram",
	}
	command, _, stderr := onboardingTestCommand("")
	command.SetIn(failingReader{})
	services := onboardingTestServices(decision)
	called := false
	preferenceChecked := false
	services.dismissed = func(string) (bool, error) {
		preferenceChecked = true
		return true, nil
	}
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
	decision := skills.OnboardingDecision{
		Action:     skills.OnboardingPromptGlobal,
		InstallDir: "/home/user/.agents/skills",
		Target:     "/home/user/.agents/skills/agent-telegram",
	}
	for _, answer := range []string{"y\n", "yes\n", "Y\n", "YES\n"} {
		t.Run(strings.TrimSpace(answer), func(t *testing.T) {
			command, _, stderr := onboardingTestCommand(answer)
			services := onboardingTestServices(decision)
			called := false
			services.install = func(_ string, target string, force bool) (string, error) {
				called = true
				if target != decision.InstallDir || force {
					t.Fatalf("install target=%q force=%v", target, force)
				}
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
			services.recordDismissal = func(target string) error {
				recorded = target
				return nil
			}
			maybeOnboardSkill(command, services)
			if recorded != decision.Target {
				t.Fatalf("recorded = %q", recorded)
			}
		})
	}
}

func TestMaybeOnboardSkillSkipsIneligibleCalls(t *testing.T) {
	decision := skills.OnboardingDecision{Action: skills.OnboardingPromptGlobal, Target: "/global/agent-telegram"}
	tests := []struct {
		name   string
		mutate func(*cobra.Command, *onboardingServices)
	}{
		{"ci", func(_ *cobra.Command, services *onboardingServices) { services.ci = "1" }},
		{"stdin", func(_ *cobra.Command, services *onboardingServices) {
			services.isTerminal = func(any) bool { return false }
		}},
		{"stdout", func(_ *cobra.Command, services *onboardingServices) {
			calls := 0
			services.isTerminal = func(any) bool {
				calls++
				return calls == 1
			}
		}},
		{"flag", func(command *cobra.Command, _ *onboardingServices) {
			command.Flags().Bool("quiet", false, "quiet")
			if err := command.Flags().Set("quiet", "true"); err != nil {
				t.Fatal(err)
			}
		}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			command, _, stderr := onboardingTestCommand("y\n")
			services := onboardingTestServices(decision)
			called := false
			services.getwd = func() (string, error) {
				called = true
				return "/repo", nil
			}
			test.mutate(command, &services)
			maybeOnboardSkill(command, services)
			if called || stderr.Len() != 0 {
				t.Fatalf("called=%v stderr=%q", called, stderr.String())
			}
		})
	}
}

func TestMaybeOnboardSkillHonorsGlobalDismissal(t *testing.T) {
	decision := skills.OnboardingDecision{Action: skills.OnboardingPromptGlobal, Target: "/global/agent-telegram"}
	command, _, stderr := onboardingTestCommand("y\n")
	services := onboardingTestServices(decision)
	services.dismissed = func(string) (bool, error) { return true, nil }
	maybeOnboardSkill(command, services)
	if stderr.Len() != 0 {
		t.Fatalf("stderr=%q", stderr.String())
	}
}

func TestMaybeOnboardSkillEOFChangesNothing(t *testing.T) {
	decision := skills.OnboardingDecision{Action: skills.OnboardingPromptGlobal, Target: "/global/agent-telegram"}
	command, _, _ := onboardingTestCommand("y")
	services := onboardingTestServices(decision)
	installed := false
	recorded := false
	services.install = func(string, string, bool) (string, error) {
		installed = true
		return "", nil
	}
	services.recordDismissal = func(string) error {
		recorded = true
		return nil
	}
	maybeOnboardSkill(command, services)
	if installed || recorded {
		t.Fatalf("installed=%v recorded=%v", installed, recorded)
	}
}

func TestProjectInstallFailureDoesNotPromptGlobal(t *testing.T) {
	decision := skills.OnboardingDecision{Action: skills.OnboardingInstallProject, InstallDir: "/project", Target: "/project/agent-telegram"}
	command, _, stderr := onboardingTestCommand("y\n")
	services := onboardingTestServices(decision)
	services.install = func(string, string, bool) (string, error) {
		return "", errors.New("read-only")
	}
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
	services.install = func(string, string, bool) (string, error) {
		installed = true
		return "", nil
	}
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
	if stderr.Len() != 0 {
		t.Fatalf("stderr=%q", stderr.String())
	}
}

func TestMaybeOnboardSkillReadFailureChangesNothing(t *testing.T) {
	decision := skills.OnboardingDecision{Action: skills.OnboardingPromptGlobal, Target: "/global/agent-telegram"}
	command, _, stderr := onboardingTestCommand("")
	command.SetIn(failingReader{})
	services := onboardingTestServices(decision)
	installed := false
	recorded := false
	services.install = func(string, string, bool) (string, error) {
		installed = true
		return "", nil
	}
	services.recordDismissal = func(string) error {
		recorded = true
		return nil
	}
	maybeOnboardSkill(command, services)
	if installed || recorded || !strings.Contains(stderr.String(), "could not read skill prompt response") {
		t.Fatalf("installed=%v recorded=%v stderr=%q", installed, recorded, stderr.String())
	}
}

func TestMaybeOnboardSkillGlobalFailuresAreNonFatal(t *testing.T) {
	decision := skills.OnboardingDecision{Action: skills.OnboardingPromptGlobal, InstallDir: "/global", Target: "/global/agent-telegram"}
	tests := []struct {
		name   string
		answer string
		mutate func(*onboardingServices)
		want   string
	}{
		{"preference-read", "y\n", func(services *onboardingServices) {
			services.dismissed = func(string) (bool, error) { return false, errors.New("read-only") }
		}, "could not read global skill preference"},
		{"install", "y\n", func(services *onboardingServices) {
			services.install = func(string, string, bool) (string, error) { return "", errors.New("read-only") }
		}, "agent-telegram skills install agent-telegram"},
		{"dismissal-write", "n\n", func(services *onboardingServices) {
			services.recordDismissal = func(string) error { return errors.New("read-only") }
		}, "could not save global skill preference"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			command, _, stderr := onboardingTestCommand(test.answer)
			services := onboardingTestServices(decision)
			test.mutate(&services)
			maybeOnboardSkill(command, services)
			if !strings.Contains(stderr.String(), test.want) {
				t.Fatalf("stderr=%q", stderr.String())
			}
		})
	}
}

func TestNonInteractiveRootOutputRemainsExact(t *testing.T) {
	command, stdout, stderr := onboardingTestCommand("")
	services := onboardingTestServices(skills.OnboardingDecision{})
	services.isTerminal = func(any) bool { return false }
	if err := runRootWithServices(command, nil, services); err != nil {
		t.Fatal(err)
	}
	if stdout.String() != rootWelcome || stderr.Len() != 0 {
		t.Fatalf("stdout=%q stderr=%q", stdout.String(), stderr.String())
	}
}

func TestSubcommandDoesNotRunRootOnboarding(t *testing.T) {
	called := false
	services := onboardingTestServices(skills.OnboardingDecision{})
	services.getwd = func() (string, error) {
		called = true
		return "/repo", nil
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
