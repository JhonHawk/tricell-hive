package main

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestSkillShellReadEvidence(t *testing.T) {
	path := "/home/.agents/skills/flow-report/SKILL.md"
	body := "---\nname: flow-report\ndescription: Visual reports\n---\n# Reports\nUse evidence.\n"
	for _, tc := range []struct {
		name, command, output string
		exit                  int
		want                  string
	}{
		{"wrapped-multiple", `/bin/zsh -lc "wc -l ` + path + ` && sed -n '1,220p' ` + path + ` && cat /other"`, body, 0, "pass"},
		{"literal", "cat " + path, body, 0, "pass"},
		{"failed", "cat " + path, body, 1, "not_observed"},
		{"wrong-output", "cat " + path, "I read flow-report", 0, "not_observed"},
		{"echo", "echo 'cat " + path + "'", body, 0, "not_observed"},
		{"variable", `cat "$SKILL/flow-report/SKILL.md"`, body, 0, "not_observed"},
		{"redirect", "cat " + path + " > /tmp/out", body, 0, "not_observed"},
		{"pipe", "cat " + path + " | head", body, 0, "not_observed"},
		{"or", "cat " + path + " || true", body, 0, "not_observed"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			event := map[string]any{"type": "item.completed", "item": map[string]any{"id": "cmd", "type": "command_execution", "command": tc.command, "aggregated_output": tc.output, "exit_code": tc.exit, "status": "completed"}}
			data, _ := json.Marshal(event)
			r := result{Trace: parseTrace("codex", strings.NewReader(string(data)))}
			if got := observeSkill(r, "flow-report").Read.Status; got != tc.want {
				t.Fatalf("got %s want %s", got, tc.want)
			}
			if observeSkill(r, "workspace-archive").Read.Status == "pass" {
				t.Fatal("unread skill passed")
			}
		})
	}
}

func TestSourceShellReadRequiresMatchingOutputAndPaths(t *testing.T) {
	yes := true
	f := fixture{ID: "project-state", Files: map[string]string{"src/dispatch.ts": "export const state = 1;\n", "tests/dispatch.test.ts": "assert.equal(state, 1);\n"}}
	for _, tc := range []struct{ name, command, output, want string }{
		{"cat-batch", "cat src/dispatch.ts tests/dispatch.test.ts", f.Files["src/dispatch.ts"] + f.Files["tests/dispatch.test.ts"], "pass"},
		{"missing-output", "cat src/dispatch.ts tests/dispatch.test.ts", f.Files["src/dispatch.ts"], "not_verified"},
		{"wrong-root", "cat /elsewhere/src/dispatch.ts /elsewhere/tests/dispatch.test.ts", f.Files["src/dispatch.ts"] + f.Files["tests/dispatch.test.ts"], "not_verified"},
		{"opaque-loop", `for f in src/dispatch.ts tests/dispatch.test.ts; do cat "$f"; done`, f.Files["src/dispatch.ts"] + f.Files["tests/dispatch.test.ts"], "not_verified"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := result{Root: "/fixture", Cwd: "/fixture", Terminal: "completed", Trace: traceReport{Events: []traceEvent{
				{Kind: "shell", ID: "cmd", Command: tc.command, Success: &yes},
				{Kind: "tool_result", ID: "cmd", Text: tc.output, Success: &yes},
			}}}
			a := assessFlows(r, f, t.TempDir())
			for _, c := range a.Criteria {
				if c.Criterion == "Observed repository source/test reads" {
					if c.Status != tc.want {
						t.Fatalf("got %+v", c)
					}
					return
				}
			}
			t.Fatal("criterion missing")
		})
	}
}

func TestPartialSkillReadDoesNotPassFailedCommand(t *testing.T) {
	path := "/home/.agents/skills/flow-research/SKILL.md"
	body := "---\nname: flow-research\ndescription: Research\n---\n# Research\nEvidence.\ncat: /missing: No such file or directory"
	for _, tc := range []struct {
		command, output string
		want            bool
	}{
		{"cat " + path + " && cat /missing", body, true},
		{"/bin/zsh -lc 'cat " + path + " && cat /missing'", body, true},
		{"cat " + path, body, false},
		{"echo 'cat " + path + " && cat /missing'", body, false},
		{"cat " + path + " && cat /missing", "cat: /missing: missing", false},
		{"cat /missing && cat " + path, body, false},
	} {
		no := false
		r := result{Trace: traceReport{Events: []traceEvent{
			{Kind: "shell", ID: "cmd", Command: tc.command, Success: &no},
			{Kind: "tool_result", ID: "cmd", Text: tc.output, Success: &no},
		}}}
		got := observeSkill(r, "flow-research")
		if (got.PartialRead != nil) != tc.want || got.Read.Status == "pass" {
			t.Fatalf("unexpected observation: %+v for %s", got, tc.command)
		}
		r.Trace.Events[1].ID = "other"
		if observeSkill(r, "flow-research").PartialRead != nil {
			t.Fatal("uncorrelated result accepted")
		}
	}
}
