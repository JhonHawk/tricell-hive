package pi

import (
	"path/filepath"
	"strings"
	"testing"

	"tricell-hive/integrations/target"
)

func TestResolvePiUserTargets(t *testing.T) {
	home := t.TempDir()
	c := target.Config{Scope: "user", Home: home, PiHome: filepath.Join(home, ".pi", "agent")}
	targets, err := Resolve(c)
	if err != nil {
		t.Fatal(err)
	}
	if len(targets) != 3 || targets[0].Path != filepath.Join(c.PiHome, "AGENTS.md") || targets[1].Path != filepath.Join(home, ".agents", "skills", "workspace-conventions", "SKILL.md") {
		t.Fatalf("unexpected Pi targets: %#v", targets)
	}
	if ev := targets[2]; ev.Kind != "hook" || ev.Source != EventExtensionSource || ev.Path != filepath.Join(c.PiHome, "extensions", "hive-events.ts") || ev.Host != "pi" {
		t.Fatalf("unexpected event extension target: %#v", ev)
	}
}

func TestResolvePiRejectsProjectScope(t *testing.T) {
	if _, err := Resolve(target.Config{Scope: "project"}); err == nil || !strings.Contains(err.Error(), "unsupported") {
		t.Fatalf("expected unsupported project-scope error, got %v", err)
	}
}
