package claude

import (
	"path/filepath"
	"testing"

	"tricell-hive/integrations/target"
)

func TestUserTargetsUseSharedSkillAndRelativeSymlink(t *testing.T) {
	home := t.TempDir()
	c := target.Config{Scope: "user", Home: home, ClaudeHome: filepath.Join(home, ".claude")}
	targets, err := Resolve(c)
	if err != nil {
		t.Fatal(err)
	}
	if len(targets) != 3 {
		t.Fatalf("got %d targets, want 3", len(targets))
	}
	if got, want := targets[0].Path, filepath.Join(c.ClaudeHome, "CLAUDE.md"); got != want {
		t.Errorf("instruction path = %q, want %q", got, want)
	}
	shared := filepath.Join(home, ".agents", "skills", "workspace-conventions")
	if got, want := targets[1].Path, filepath.Join(shared, "SKILL.md"); got != want {
		t.Errorf("shared skill = %q, want %q", got, want)
	}
	link := filepath.Join(c.ClaudeHome, "skills", "workspace-conventions")
	if got, want := targets[2].Path, link; got != want {
		t.Errorf("skill link path = %q, want %q", got, want)
	}
	if targets[2].Kind != "symlink" {
		t.Errorf("link kind = %q, want symlink", targets[2].Kind)
	}
	wantTarget, err := filepath.Rel(filepath.Dir(link), shared)
	if err != nil {
		t.Fatal(err)
	}
	if targets[2].LinkTarget != wantTarget {
		t.Errorf("link target = %q, want %q", targets[2].LinkTarget, wantTarget)
	}
}

func TestProjectDestinationsRemainUnchanged(t *testing.T) {
	c := target.Config{Scope: "project", Home: "/synthetic", Root: "/repo", ClaudeHome: "/synthetic/.claude"}
	targets, err := Resolve(c)
	if err != nil {
		t.Fatal(err)
	}
	if len(targets) != 2 || targets[0].Path != filepath.Join(c.Root, "CLAUDE.md") || targets[1].Path != filepath.Join(c.Root, ".claude", "skills", "workspace-conventions", "SKILL.md") {
		t.Fatalf("project targets changed: %#v", targets)
	}
}
