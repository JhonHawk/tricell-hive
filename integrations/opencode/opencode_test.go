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

func TestContextLoaderTargetsRequireExactSource(t *testing.T) {
	home := t.TempDir()
	c := target.Config{Scope: "user", Home: home, OpenCodeHome: filepath.Join(home, "custom-opencode")}
	for _, source := range []string{ContextLoaderSource, "integrations/opencode/other.js", "integrations/opencode/context-loader.json"} {
		ts, err := Resolve(c, source)
		if err != nil {
			t.Fatal(err)
		}
		var plugins []target.Target
		for _, target := range ts {
			if target.Kind == "plugin" {
				plugins = append(plugins, target)
			}
		}
		want := 0
		if source == ContextLoaderSource {
			want = 2
		}
		if len(plugins) != want {
			t.Fatalf("source %q: %d plugin destinations, want %d", source, len(plugins), want)
		}
		for _, target := range plugins {
			if filepath.Base(target.Path) != "index.js" && filepath.Base(target.Path) != "context-loader.json" {
				t.Fatal("unexpected plugin filename")
			}
			if filepath.Dir(target.Path) != filepath.Join(c.OpenCodeHome, "hive", "context-loader") || target.Host != "opencode" || target.Scope != "user" {
				t.Fatal("plugin is not user-level inactive delivery")
			}
		}
	}
}
