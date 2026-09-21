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
