package main

import "testing"

func TestProjectStateSeparatesLocationAndNaming(t *testing.T) {
	for _, tc := range []struct {
		name             string
		paths            []string
		location, naming string
	}{
		{"correct", []string{"_support/sessions/2026-09-21-dis-42/dis-42.research.md"}, "pass", "pass"},
		{"claude-report", []string{"_support/sessions/2026-09-21-dis-42/dis-42.report.md"}, "pass", "fail"},
		{"codex-root", []string{"dis-42.research.md"}, "fail", "pass"},
		{"html-root", []string{"dis-42.report.html"}, "fail", "fail"},
		{"missing", nil, "fail", "fail"},
		{"two", []string{"_support/sessions/x/a.research.md", "_support/sessions/x/b.research.md"}, "fail", "fail"},
		{"no-session-folder", []string{"_support/sessions/a.research.md"}, "fail", "pass"},
		{"escape", []string{"_support/sessions/../../a.research.md"}, "fail", "pass"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := fixture{ID: "project-state", Files: map[string]string{"README.md": "fixture"}}
			r := result{Before: map[string]item{"_support/sessions/old/old.research.md": {}}, After: map[string]item{"README.md": {}, "_support/sessions/old/old.research.md": {}}}
			for _, p := range tc.paths {
				r.After[p] = item{}
			}
			got := projectStateRecordCriteria(r, f)
			if got[0].Status != tc.location || got[1].Status != tc.naming {
				t.Fatalf("got %+v", got)
			}
		})
	}
}
