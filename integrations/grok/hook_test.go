package grok

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"tricell-hive/integrations/target"
)

type hookFileShape struct {
	Hooks struct {
		PreToolUse []struct {
			Matcher string `json:"matcher"`
			Hooks   []struct {
				Type    string `json:"type"`
				Command string `json:"command"`
				Timeout int    `json:"timeout"`
			} `json:"hooks"`
		} `json:"PreToolUse"`
	} `json:"hooks"`
}

func TestRenderReadHookShape(t *testing.T) {
	data, err := RenderReadHook("/opt/hive/bin/hive", "/home/u/.agents/skills")
	if err != nil {
		t.Fatal(err)
	}
	var f hookFileShape
	if err := json.Unmarshal(data, &f); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, data)
	}
	groups := f.Hooks.PreToolUse
	if len(groups) != 1 || groups[0].Matcher != "read_file" || len(groups[0].Hooks) != 1 {
		t.Fatalf("unexpected groups: %+v", groups)
	}
	h := groups[0].Hooks[0]
	want := `'/opt/hive/bin/hive' grok-hook read-whole --skills '/home/u/.agents/skills'`
	if h.Type != "command" || h.Command != want || h.Timeout != 10 {
		t.Fatalf("hook = %+v, want command %q", h, want)
	}
	if !strings.HasSuffix(string(data), "\n") {
		t.Fatal("rendered hook must end with a newline")
	}
}

// The command runs under sh -c, so each path must reach the program as one
// argument whatever characters it holds.
func TestRenderReadHookQuotingSurvivesTheShell(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "it's a $(touch pwned) `touch pwned2`;&|<>*? bin")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	exe := filepath.Join(dir, "hive")
	if err := os.WriteFile(exe, []byte("#!/bin/sh\nprintf '%s\\n' \"$@\"\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	skills := `/home/o'brien/$HOME/skills dir/it''s "q" \ end`
	data, err := RenderReadHook(exe, skills)
	if err != nil {
		t.Fatal(err)
	}
	var f hookFileShape
	if err := json.Unmarshal(data, &f); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("sh", "-c", f.Hooks.PreToolUse[0].Hooks[0].Command)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("sh: %v\n%s", err, out)
	}
	want := "grok-hook\nread-whole\n--skills\n" + skills + "\n"
	if string(out) != want {
		t.Fatalf("arguments = %q, want %q", out, want)
	}
	for _, name := range []string{"pwned", "pwned2"} {
		if _, err := os.Stat(filepath.Join(dir, name)); err == nil {
			t.Fatalf("shell expanded part of a path: %s was created", name)
		}
	}
}

func TestRenderReadHookRejectsUnusablePaths(t *testing.T) {
	for _, tc := range []struct{ exe, skills string }{
		{"", "/s"}, {"relative/hive", "/s"}, {"/h", ""}, {"/h", "rel"}, {"/h\x00x", "/s"}, {"/h", "/s\x00"},
	} {
		if _, err := RenderReadHook(tc.exe, tc.skills); err == nil {
			t.Errorf("RenderReadHook(%q, %q) accepted an unusable path", tc.exe, tc.skills)
		}
	}
}

func TestResolveAlwaysEmitsTheReadHookTarget(t *testing.T) {
	unsetCompatEnv(t)
	home := t.TempDir()
	c := target.Config{Scope: "user", Home: home, ClaudeHome: filepath.Join(home, ".claude"), GrokHome: filepath.Join(home, ".grok")}
	for _, sources := range [][]string{nil, {}, {"content/skills/flow-build/SKILL.md"}} {
		targets, err := Resolve(c, sources...)
		if err != nil {
			t.Fatal(err)
		}
		var hooks []target.Target
		for _, tg := range targets {
			if tg.Kind == "hook" {
				hooks = append(hooks, tg)
			}
		}
		if len(hooks) != 1 {
			t.Fatalf("sources %v: %d hook targets, want 1", sources, len(hooks))
		}
		h := hooks[0]
		if h.Path != filepath.Join(home, ".grok", "hooks", "hive-read-whole-references.json") ||
			h.Source != ReadHookSource || h.Host != "grok" || h.Scope != "user" || h.Context != home {
			t.Fatalf("hook target = %+v", h)
		}
	}
}
