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

func unsetCursorEnv(t *testing.T) {
	t.Helper()
	t.Setenv("GROK_CURSOR_AGENTS_ENABLED", "")
	if err := os.Unsetenv("GROK_CURSOR_AGENTS_ENABLED"); err != nil {
		t.Fatal(err)
	}
}

func TestCursorAgentsEnabled(t *testing.T) {
	cases := []struct {
		name    string
		config  *string
		env     string
		setEnv  bool
		synth   bool
		want    bool
		wantErr bool
	}{
		{name: "no config file", want: true},
		{name: "no cursor table", config: ptr("[compat.claude]\nagents = true\n"), want: true},
		{name: "table true", config: ptr("[compat.cursor]\nagents = true\n"), want: true},
		{name: "table false", config: ptr("[compat.cursor]\nagents = false # off\n"), want: false},
		{name: "other keys ignored", config: ptr("[compat.cursor]\nrules = false\n"), want: true},
		{name: "env false wins over toml true", config: ptr("[compat.cursor]\nagents = true\n"), env: "false", setEnv: true, want: false},
		{name: "env true wins over toml false", config: ptr("[compat.cursor]\nagents = false\n"), env: "true", setEnv: true, want: true},
		{name: "env ignored when synthetic", config: ptr("[compat.cursor]\nagents = true\n"), env: "false", setEnv: true, synth: true, want: true},
		{name: "bad env", env: "maybe", setEnv: true, wantErr: true},
		{name: "bad value", config: ptr("[compat.cursor]\nagents = maybe\n"), wantErr: true},
		{name: "duplicate table", config: ptr("[compat.cursor]\nagents = false\n[compat.cursor]\nagents = true\n"), wantErr: true},
		{name: "inline compat form", config: ptr("compat = { cursor = { agents = false } }\n"), wantErr: true},
		{name: "dotted root form", config: ptr("compat.cursor.agents = false\n"), wantErr: true},
		{name: "cursor key under compat", config: ptr("[compat]\ncursor = { agents = false }\n"), wantErr: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			unsetCursorEnv(t)
			home := t.TempDir()
			if tc.config != nil {
				if err := os.WriteFile(filepath.Join(home, "config.toml"), []byte(*tc.config), 0600); err != nil {
					t.Fatal(err)
				}
			}
			if tc.setEnv {
				t.Setenv("GROK_CURSOR_AGENTS_ENABLED", tc.env)
			}
			got, err := CursorAgentsEnabled(home, tc.synth)
			if (err != nil) != tc.wantErr {
				t.Fatalf("error = %v, wantErr %v", err, tc.wantErr)
			}
			if err == nil && got != tc.want {
				t.Fatalf("enabled = %v, want %v", got, tc.want)
			}
		})
	}
}

func ptr(s string) *string { return &s }
