package antigravity

import (
	"path/filepath"
	"strings"
	"testing"

	"tricell-hive/integrations/target"
)

func TestResolveAntigravityUserTargets(t *testing.T) {
	home := t.TempDir()
	c := target.Config{Scope: "user", Home: home, AntigravityHome: filepath.Join(home, ".gemini", "config")}
	targets, err := Resolve(c)
	if err != nil {
		t.Fatal(err)
	}
	if len(targets) != 2 {
		t.Fatalf("unexpected Antigravity targets: %#v", targets)
	}
	if targets[0].Kind != "skill" || targets[0].Path != filepath.Join(home, ".agents", "skills", "workspace-conventions", "SKILL.md") {
		t.Fatalf("unexpected shared skill: %#v", targets[0])
	}
	if targets[1].Kind != "symlink" || targets[1].LinkTarget != "../../../.agents/skills/workspace-conventions" {
		t.Fatalf("unexpected skill link: %#v", targets[1])
	}
}

func TestResolveAntigravityRejectsProjectScope(t *testing.T) {
	if _, err := Resolve(target.Config{Scope: "project"}); err == nil || !strings.Contains(err.Error(), "unsupported") {
		t.Fatalf("expected unsupported project-scope error, got %v", err)
	}
}
