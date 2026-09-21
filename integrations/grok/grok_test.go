package grok

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"tricell-hive/integrations/target"
)

func unsetCompatEnv(t *testing.T) {
	t.Helper()
	old, ok := os.LookupEnv("GROK_CLAUDE_AGENTS_ENABLED")
	if err := os.Unsetenv("GROK_CLAUDE_AGENTS_ENABLED"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if ok {
			_ = os.Setenv("GROK_CLAUDE_AGENTS_ENABLED", old)
		} else {
			_ = os.Unsetenv("GROK_CLAUDE_AGENTS_ENABLED")
		}
	})
}

func TestResolveGrokUserTargets(t *testing.T) {
	unsetCompatEnv(t)
	home := t.TempDir()
	c := target.Config{Scope: "user", Home: home, ClaudeHome: filepath.Join(home, ".claude-override"), GrokHome: filepath.Join(home, ".grok")}
	targets, err := Resolve(c)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{
		filepath.Join(home, ".claude", "CLAUDE.md"),
		filepath.Join(home, ".agents", "skills", "workspace-conventions", "SKILL.md"),
	}
	if len(targets) != len(want) {
		t.Fatalf("got %d targets, want %d", len(targets), len(want))
	}
	for i := range want {
		if targets[i].Path != want[i] {
			t.Errorf("target %d = %q, want %q", i, targets[i].Path, want[i])
		}
	}
}

func TestResolveFailsWhenClaudeCompatibilityIsDisabled(t *testing.T) {
	unsetCompatEnv(t)
	home := t.TempDir()
	if err := os.WriteFile(filepath.Join(home, "config.toml"), []byte("[compat.claude]\nagents = false\n"), 0600); err != nil {
		t.Fatal(err)
	}
	c := target.Config{Scope: "user", Home: home, ClaudeHome: filepath.Join(home, ".claude"), GrokHome: home}
	if _, err := Resolve(c); err == nil || !strings.Contains(err.Error(), "disabled") {
		t.Fatalf("Resolve error = %v, want disabled compatibility error", err)
	}
}

func TestEnvironmentOverridesGrokCompatibilityConfig(t *testing.T) {
	home := t.TempDir()
	if err := os.WriteFile(filepath.Join(home, "config.toml"), []byte("[compat.claude]\nagents = false\n"), 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("GROK_CLAUDE_AGENTS_ENABLED", "true")
	c := target.Config{Scope: "user", Home: home, ClaudeHome: filepath.Join(home, ".claude"), GrokHome: home}
	if _, err := Resolve(c); err != nil {
		t.Fatalf("env override should enable compatibility: %v", err)
	}
	t.Setenv("GROK_CLAUDE_AGENTS_ENABLED", "unsupported")
	if _, err := Resolve(c); err == nil || !strings.Contains(err.Error(), "unsupported") {
		t.Fatalf("unsupported env value error = %v", err)
	}
}

func TestResolveRejectsProjectScope(t *testing.T) {
	_, err := Resolve(target.Config{Scope: "project"})
	if err == nil || !strings.Contains(err.Error(), "unsupported") {
		t.Fatalf("Resolve error = %v, want unsupported project scope", err)
	}
}

func TestSyntheticResolveIgnoresRealCompatibilityEnvironment(t *testing.T) {
	home := t.TempDir()
	t.Setenv("GROK_CLAUDE_AGENTS_ENABLED", "false")
	c := target.Config{Scope: "user", Home: home, ClaudeHome: filepath.Join(home, ".claude"), GrokHome: filepath.Join(home, ".grok"), Synthetic: true}
	if _, err := Resolve(c); err != nil {
		t.Fatalf("synthetic resolution read host compatibility environment: %v", err)
	}
}

func TestResolveRejectsUnsupportedCompatibilityTOMLForms(t *testing.T) {
	unsetCompatEnv(t)
	forms := []string{
		"compat   = { claude = { agents = false } }\n",
		"[compat]\nclaude = { agents = false }\n",
		"[\"compat\".\"claude\"]\nagents = false\n",
		"[compat.claude]\n\"agents\" = false\n",
	}
	for i, form := range forms {
		t.Run(string(rune('a'+i)), func(t *testing.T) {
			home := t.TempDir()
			if err := os.WriteFile(filepath.Join(home, "config.toml"), []byte(form), 0600); err != nil {
				t.Fatal(err)
			}
			c := target.Config{Scope: "user", Home: home, ClaudeHome: filepath.Join(home, ".claude"), GrokHome: home}
			if _, err := Resolve(c); err == nil || !strings.Contains(err.Error(), "unsupported") {
				t.Fatalf("Resolve error = %v, want unsupported config error", err)
			}
		})
	}
}
