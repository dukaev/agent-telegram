package skills

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestManifestIncludesAgentTelegramSkill(t *testing.T) {
	items := Manifest()
	if len(items) == 0 {
		t.Fatal("skills manifest should not be empty")
	}
	if items[0]["name"] != "agent-telegram" {
		t.Fatalf("first skill name = %v, want agent-telegram", items[0]["name"])
	}
	if !strings.Contains(items[0]["installCommand"].(string), "skills install agent-telegram") {
		t.Fatalf("install command should mention skills install: %+v", items[0])
	}
}

func TestInstallBundledSkill(t *testing.T) {
	target := t.TempDir()
	installedPath, err := Install("agent-telegram", target, false)
	if err != nil {
		t.Fatal(err)
	}
	if installedPath != filepath.Join(target, "agent-telegram") {
		t.Fatalf("installed path = %q", installedPath)
	}
	for _, rel := range []string{"SKILL.md", filepath.Join("agents", "openai.yaml")} {
		if _, err := os.Stat(filepath.Join(installedPath, rel)); err != nil {
			t.Fatalf("expected installed file %s: %v", rel, err)
		}
	}
	if _, err := Install("agent-telegram", target, false); err == nil {
		t.Fatal("install without force should fail when skill already exists")
	}
	if _, err := Install("agent-telegram", target, true); err != nil {
		t.Fatalf("install with force should overwrite: %v", err)
	}
}
