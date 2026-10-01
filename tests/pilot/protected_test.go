package main

import (
	"os"
	"path/filepath"
	"testing"
)

func fixtureFile(t *testing.T, p, s string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(p), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(s), 0600); err != nil {
		t.Fatal(err)
	}
}
func clearRoots(t *testing.T) {
	t.Helper()
	for _, k := range []string{"CODEX_HOME", "CLAUDE_CONFIG_DIR", "GROK_HOME", "PI_CODING_AGENT_DIR", "XDG_CONFIG_HOME"} {
		t.Setenv(k, "")
	}
}
func TestOtherWorkerTrustDoesNotAlertClaude(t *testing.T) {
	clearRoots(t)
	home := t.TempDir()
	path := filepath.Join(home, ".codex/config.toml")
	fixtureFile(t, path, "model = \"original\"\n")
	claudeBefore, codexBefore := protectedFor("claude", home), protectedFor("codex", home)
	fixtureFile(t, path, "model = \"original\"\n[projects.\"/fixture\"]\ntrust_level = \"trusted\"\n")
	if got := changedProtected(claudeBefore, protectedFor("claude", home)); len(got) != 0 {
		t.Fatalf("cross-worker false alert: %v", got)
	}
	if got := changedProtected(codexBefore, protectedFor("codex", home)); len(got) != 1 || got[0] != path {
		t.Fatalf("owner must see delta: %v", got)
	}
	fixtureFile(t, filepath.Join(home, ".agents/skills/workspace-conventions/SKILL.md"), "modified")
	if got := changedProtected(claudeBefore, protectedFor("claude", home)); len(got) == 0 {
		t.Fatal("shared guidance mutation was ignored")
	}
}

// TestProtectedForHomeIgnoresRunnerEnvironment is finding 5: protectedFor
// resolves CODEX_HOME/CLAUDE_CONFIG_DIR/GROK_HOME from this process's own
// environment, which is correct for the real home but wrong for a guidance
// variant's shadow home — a runner invoked with a real CODEX_HOME exported
// must not have the shadow-home audit collapse onto that real path.
// protectedForHome must resolve every path directly under the shadow home
// regardless of what the runner's own environment declares, while
// protectedFor keeps honoring the environment for a real home, unchanged.
func TestProtectedForHomeIgnoresRunnerEnvironment(t *testing.T) {
	realCodexHome := t.TempDir()
	realGrokHome := t.TempDir()
	realClaudeHome := t.TempDir()
	t.Setenv("CODEX_HOME", realCodexHome)
	t.Setenv("GROK_HOME", realGrokHome)
	t.Setenv("CLAUDE_CONFIG_DIR", realClaudeHome)
	shadowHome := t.TempDir()
	for _, host := range []string{"codex", "grok", "claude"} {
		for p := range protectedForHome(host, shadowHome) {
			if under(p, realCodexHome) || under(p, realGrokHome) || under(p, realClaudeHome) {
				t.Fatalf("%s: shadow protected path collapsed onto the runner's real environment: %s", host, p)
			}
			if !under(p, shadowHome) {
				t.Fatalf("%s: shadow protected path escaped the shadow home: %s", host, p)
			}
		}
	}
	// The real-home variant must still honor the runner's own environment;
	// the fix must not touch this existing, correct behavior.
	found := false
	for p := range protectedFor("codex", t.TempDir()) {
		if under(p, realCodexHome) {
			found = true
		}
	}
	if !found {
		t.Fatal("protectedFor must still honor CODEX_HOME for a real home")
	}
}

func TestLinkIdentityAndFollowedPayloadAreProtected(t *testing.T) {
	home := t.TempDir()
	target := filepath.Join(home, "canonical")
	link := filepath.Join(home, "link")
	fixtureFile(t, filepath.Join(target, "SKILL.md"), "one")
	if err := os.Symlink(target, link); err != nil {
		t.Fatal(err)
	}
	before := fingerprint(link, map[string]bool{})
	fixtureFile(t, filepath.Join(target, "SKILL.md"), "two")
	after := fingerprint(link, map[string]bool{})
	if before == after {
		t.Fatal("target mutation hidden")
	}
	other := filepath.Join(home, "other")
	fixtureFile(t, filepath.Join(other, "SKILL.md"), "two")
	if err := os.Remove(link); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(other, link); err != nil {
		t.Fatal(err)
	}
	if after == fingerprint(link, map[string]bool{}) {
		t.Fatal("link retarget with equal payload hidden")
	}
}

// TestOpenCodeCLISettingsAreProtected covers issue #81: OpenCode V2 keeps its
// terminal-client settings in the global cli.json, so a pilot that changed it
// must leave a fingerprint delta for the opencode worker.
func TestOpenCodeCLISettingsAreProtected(t *testing.T) {
	clearRoots(t)
	home := t.TempDir()
	path := filepath.Join(home, ".config/opencode/cli.json")
	fixtureFile(t, path, "{}\n")
	before := protectedFor("opencode", home)
	fixtureFile(t, path, "{\"theme\":\"changed\"}\n")
	if got := changedProtected(before, protectedFor("opencode", home)); len(got) != 1 || got[0] != path {
		t.Fatalf("cli.json change not detected: %v", got)
	}
}
