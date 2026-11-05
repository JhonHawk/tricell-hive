package grok

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestRenderSessionEventsHookRunsTheProgramDirectly(t *testing.T) {
	exe := `/Users/o'brien/a b/hive`
	out, err := RenderSessionEventsHook(exe)
	if err != nil {
		t.Fatal(err)
	}
	var f struct {
		Hooks map[string][]struct {
			Hooks []struct {
				Type    string   `json:"type"`
				Command string   `json:"command"`
				Args    []string `json:"args"`
				Timeout int      `json:"timeout"`
			} `json:"hooks"`
		} `json:"hooks"`
	}
	if err := json.Unmarshal(out, &f); err != nil {
		t.Fatalf("not JSON: %v\n%s", err, out)
	}
	if len(f.Hooks) != 4 {
		t.Fatalf("events = %v", f.Hooks)
	}
	for _, event := range []string{"SessionStart", "UserPromptSubmit", "Stop", "SessionEnd"} {
		groups := f.Hooks[event]
		if len(groups) != 1 || len(groups[0].Hooks) != 1 {
			t.Fatalf("%s: %+v", event, groups)
		}
		h := groups[0].Hooks[0]
		if h.Type != "command" || h.Command != exe || strings.Join(h.Args, " ") != "hook grok "+event || h.Timeout <= 0 {
			t.Fatalf("%s handler = %+v", event, h)
		}
	}
	// A present args array makes command one program: nothing is quoted, so the
	// path survives spaces and apostrophes.
	if strings.Contains(string(out), `'\''`) || strings.Contains(string(out), `\"`) {
		t.Fatalf("the path was shell-quoted:\n%s", out)
	}
}

func TestRenderSessionEventsHookRejectsAnUnusablePath(t *testing.T) {
	for _, bad := range []string{"", "hive", "relative/hive", "/a\x00b"} {
		if _, err := RenderSessionEventsHook(bad); err == nil {
			t.Fatalf("accepted %q", bad)
		}
	}
}
