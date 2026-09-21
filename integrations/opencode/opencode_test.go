package opencode

import (
	"path/filepath"
	"strings"
	"testing"

	"tricell-hive/integrations/target"
)

func TestResolveOpenCodeUserTargets(t *testing.T) {
	home := t.TempDir()
	c := target.Config{Scope: "user", Home: home, OpenCodeHome: filepath.Join(home, ".config", "opencode")}
	targets, err := Resolve(c)
	if err != nil {
		t.Fatal(err)
	}
	if len(targets) != 2 || targets[0].Path != filepath.Join(c.OpenCodeHome, "AGENTS.md") || targets[1].Path != filepath.Join(home, ".agents", "skills", "workspace-conventions", "SKILL.md") {
		t.Fatalf("unexpected OpenCode targets: %#v", targets)
	}
}

func TestResolveOpenCodeRejectsProjectScope(t *testing.T) {
	if _, err := Resolve(target.Config{Scope: "project"}); err == nil || !strings.Contains(err.Error(), "unsupported") {
		t.Fatalf("expected unsupported project-scope error, got %v", err)
	}
}
