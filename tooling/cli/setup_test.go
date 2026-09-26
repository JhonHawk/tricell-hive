package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSetupMissingIsOptionalAndReadOnly(t *testing.T) {
	home := t.TempDir()
	// A real-home override must not contaminate a synthetic check.
	t.Setenv("CLAUDE_CONFIG_DIR", "/must-not-read")
	var out bytes.Buffer
	if err := setup([]string{"--home", home}, &out); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "No find-docs") || !strings.Contains(out.String(), "optional") || !strings.Contains(out.String(), "ctx7@latest setup --cli") {
		t.Fatal(out.String())
	}
	entries, err := os.ReadDir(home)
	if err != nil || len(entries) != 0 {
		t.Fatalf("setup wrote to home: %v %v", entries, err)
	}
}

func TestSetupDetectsSkillsWithoutChangingThem(t *testing.T) {
	home := t.TempDir()
	path := filepath.Join(home, ".agents", "skills", "find-docs", "SKILL.md")
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		t.Fatal(err)
	}
	original := []byte("Vendor skill; preserve exact bytes.\r\n")
	if err := os.WriteFile(path, original, 0640); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	if err := setup([]string{"--home", home}, &out); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "Skill file detected: "+path) || !strings.Contains(out.String(), "does not verify") {
		t.Fatal(out.String())
	}
	after, err := os.ReadFile(path)
	if err != nil || !bytes.Equal(after, original) {
		t.Fatal("vendor skill modified")
	}
	info, _ := os.Stat(path)
	if info.Mode().Perm() != 0640 {
		t.Fatal("mode changed")
	}
}

// TestSetupMentionsInstallerAndStaysReadOnly covers U9: setup must point the
// operator to the new interactive installer for optional capabilities, and
// must still write nothing.
func TestSetupMentionsInstallerAndStaysReadOnly(t *testing.T) {
	home := t.TempDir()
	var out bytes.Buffer
	if err := setup([]string{"--home", home}, &out); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "./install.sh") || !strings.Contains(out.String(), "hive install") {
		t.Fatalf("missing pointer to the interactive installer: %s", out.String())
	}
	entries, err := os.ReadDir(home)
	if err != nil || len(entries) != 0 {
		t.Fatalf("setup wrote to home: %v %v", entries, err)
	}
}

func TestContext7CandidatesHonorHomesAndDeduplicate(t *testing.T) {
	env := func(k string) string {
		if k == "CLAUDE_CONFIG_DIR" {
			return "/custom-claude"
		}
		return ""
	}
	paths := context7Candidates("/home/example", env)
	seen := map[string]bool{}
	for _, p := range paths {
		if seen[p] {
			t.Fatal("duplicate path")
		}
		seen[p] = true
	}
	if !seen["/custom-claude/skills/find-docs/SKILL.md"] || !seen["/home/example/.agents/skills/find-docs/SKILL.md"] || !seen["/home/example/.claude/skills/context7-mcp/SKILL.md"] {
		t.Fatal(paths)
	}
}
