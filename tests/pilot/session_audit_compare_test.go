package main

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
)

func fakeAuditor(byFile map[string][]criterionAssessment) auditRunner {
	return func(path, host string) (auditReport, error) {
		c, ok := byFile[path]
		if !ok {
			return auditReport{}, fmt.Errorf("unexpected %s", path)
		}
		return auditReport{Host: host, File: path, Criteria: c}, nil
	}
}

func TestAuditCompareShowsOnlyChangedStatusesWithoutSessionText(t *testing.T) {
	path := "/home/u/.codex/sessions/2026/11/23/rollout-a.jsonl"
	secret := "SESSION-TEXT-MUST-NOT-APPEAR"
	oldRun := fakeAuditor(map[string][]criterionAssessment{path: {
		{Criterion: "kept", Status: "pass"},
		{Criterion: "became_fail", Status: "pass"},
		{Criterion: "became_pass", Status: "fail", Evidence: []string{"rollout-a.jsonl:408 text " + secret, "rollout-a.jsonl:421 text " + secret}},
	}})
	newRun := fakeAuditor(map[string][]criterionAssessment{path: {
		{Criterion: "kept", Status: "pass"},
		{Criterion: "became_fail", Status: "fail", Evidence: []string{"rollout-a.jsonl:496 text " + secret}},
		{Criterion: "became_pass", Status: "pass"},
	}})
	var out bytes.Buffer
	if err := compareAudit(&out, []string{path}, "", oldRun, newRun); err != nil {
		t.Fatal(err)
	}
	want := "rollout-a.jsonl · became_fail · pass → fail · 496\n" +
		"rollout-a.jsonl · became_pass · fail → pass · 408,421\n"
	if out.String() != want {
		t.Fatalf("got:\n%s\nwant:\n%s", out.String(), want)
	}
	if strings.Contains(out.String(), secret) {
		t.Fatal("session text leaked")
	}
}

func TestAuditCompareDeducesHostFromPathAndHostFlagOverrides(t *testing.T) {
	for path, want := range map[string]string{
		"/h/.claude/projects/-p/abc.jsonl":        "claude",
		"/h/.codex/sessions/2026/11/23/r.jsonl":   "codex",
		"/h/.grok/sessions/proj/id/updates.jsonl": "grok",
	} {
		var got string
		run := func(p, host string) (auditReport, error) { got = host; return auditReport{}, nil }
		if err := compareAudit(&bytes.Buffer{}, []string{path}, "", run, run); err != nil {
			t.Fatal(err)
		}
		if got != want {
			t.Errorf("%s: host %q, want %q", path, got, want)
		}
	}
	run := func(p, host string) (auditReport, error) { return auditReport{Host: host}, nil }
	if err := compareAudit(&bytes.Buffer{}, []string{"/elsewhere/x.jsonl"}, "", run, run); err == nil {
		t.Fatal("unknown path without --host must fail")
	}
	var got string
	run = func(p, host string) (auditReport, error) { got = host; return auditReport{}, nil }
	if err := compareAudit(&bytes.Buffer{}, []string{"/elsewhere/x.jsonl"}, "grok", run, run); err != nil || got != "grok" {
		t.Fatalf("override: host %q err %v", got, err)
	}
}
