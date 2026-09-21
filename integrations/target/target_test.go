package target

import (
	"encoding/json"
	"path/filepath"
	"testing"
)

func TestExpandHostHomesDefaultsAndSyntheticEnvironmentIsolation(t *testing.T) {
	home := t.TempDir()
	t.Setenv("PI_CODING_AGENT_DIR", filepath.Join(t.TempDir(), "pi-custom"))
	t.Setenv("GROK_HOME", filepath.Join(t.TempDir(), "grok-custom"))
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(t.TempDir(), "xdg-custom"))

	c, err := ExpandHostHomes(Config{Home: home}, true)
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]string{
		"pi":       filepath.Join(home, ".pi", "agent"),
		"grok":     filepath.Join(home, ".grok"),
		"opencode": filepath.Join(home, ".config", "opencode"),
	}
	got := map[string]string{"pi": c.PiHome, "grok": c.GrokHome, "opencode": c.OpenCodeHome}
	for host, path := range want {
		path, err = Canonical(path)
		if err != nil {
			t.Fatal(err)
		}
		if got[host] != path {
			t.Errorf("%s home = %q, want %q", host, got[host], path)
		}
	}
}

func TestExpandHostHomesUsesNativeOverrides(t *testing.T) {
	home := t.TempDir()
	pi := filepath.Join(t.TempDir(), "pi")
	grok := filepath.Join(t.TempDir(), "grok")
	xdg := filepath.Join(t.TempDir(), "xdg")
	t.Setenv("PI_CODING_AGENT_DIR", pi)
	t.Setenv("GROK_HOME", grok)
	t.Setenv("XDG_CONFIG_HOME", xdg)

	c, err := ExpandHostHomes(Config{Home: home}, false)
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]string{
		"pi":       pi,
		"grok":     grok,
		"opencode": filepath.Join(xdg, "opencode"),
	}
	got := map[string]string{"pi": c.PiHome, "grok": c.GrokHome, "opencode": c.OpenCodeHome}
	for host, path := range want {
		path, err = Canonical(path)
		if err != nil {
			t.Fatal(err)
		}
		if got[host] != path {
			t.Errorf("%s home = %q, want %q", host, got[host], path)
		}
	}
}

func TestExpandHostHomesRejectsRelativeXDGOverride(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "relative/config")
	if _, err := ExpandHostHomes(Config{Home: t.TempDir()}, false); err == nil {
		t.Fatal("relative XDG_CONFIG_HOME should be rejected")
	}
}

func TestExpandHostHomesCanonicalizesExplicitValues(t *testing.T) {
	home := t.TempDir()
	pi := filepath.Join(home, "..", filepath.Base(home), "pi")
	c, err := ExpandHostHomes(Config{Home: home, PiHome: pi, GrokHome: filepath.Join(home, "grok"), OpenCodeHome: filepath.Join(home, "opencode")}, true)
	if err != nil {
		t.Fatal(err)
	}
	want, err := Canonical(filepath.Join(home, "pi"))
	if err != nil {
		t.Fatal(err)
	}
	if c.PiHome != want {
		t.Fatalf("Pi home = %q, want canonical %q", c.PiHome, want)
	}
}

func TestNewHomesOmitFromLegacyConfigJSON(t *testing.T) {
	b, err := json.Marshal(Config{Scope: "user", Home: "/synthetic", CodexHome: "/synthetic/.codex", ClaudeHome: "/synthetic/.claude"})
	if err != nil {
		t.Fatal(err)
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(b, &fields); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"pi_home", "grok_home", "opencode_home"} {
		if _, ok := fields[key]; ok {
			t.Errorf("legacy config unexpectedly contains %q", key)
		}
	}

}
