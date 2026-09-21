package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func handoffFixture(t *testing.T, body string) (string, string) {
	t.Helper()
	producer := t.TempDir()
	dst := t.TempDir()
	plan := "_support/sessions/2026-09-21-items/items.plan.md"
	fixtureFile(t, filepath.Join(producer, "final-files", plan), body)
	inv, err := inventory(filepath.Join(producer, "final-files"))
	if err != nil {
		t.Fatal(err)
	}
	b, _ := json.Marshal(result{Suite: "flows", Case: "plan", Host: "codex", Terminal: "completed", After: inv})
	fixtureFile(t, filepath.Join(producer, "run.json"), string(b))
	return producer, dst
}
func TestHandoffCopiesOnlyVerifiedSupportRecords(t *testing.T) {
	p, d := handoffFixture(t, "# Plan\n- [ ] Implement\n[source](../../../src/items.ts)\n")
	fixtureFile(t, filepath.Join(d, "src/items.ts"), "receiving source")
	fixtureFile(t, filepath.Join(p, "final-files/src/items.ts"), "producer source must not copy")
	report, err := importHandoff(p, d)
	if err != nil {
		t.Fatal(err)
	}
	if len(report.Files) != 1 || report.Host != "codex" || report.RunHash == "" {
		t.Fatalf("bad provenance: %+v", report)
	}
	b, _ := os.ReadFile(filepath.Join(d, "src/items.ts"))
	if string(b) != "receiving source" {
		t.Fatal("source replaced")
	}
}
func TestHandoffRejectsEscapesAndMissingReferences(t *testing.T) {
	for _, link := range []string{"../../../../outside.md", "/private/file.md", "missing.research.md", "../../../AGENTS.md"} {
		t.Run(link, func(t *testing.T) {
			p, d := handoffFixture(t, "# Plan\n[link]("+link+")\n")
			if _, err := importHandoff(p, d); err == nil {
				t.Fatal("unsafe handoff accepted")
			}
			if _, err := os.Stat(filepath.Join(d, "_support")); !os.IsNotExist(err) {
				t.Fatal("partial import before validation")
			}
		})
	}
}
func TestHandoffRejectsTamperingAndSymlinks(t *testing.T) {
	for _, mode := range []string{"tamper", "symlink"} {
		t.Run(mode, func(t *testing.T) {
			p, d := handoffFixture(t, "plan")
			path := filepath.Join(p, "final-files/_support/sessions/2026-09-21-items/items.plan.md")
			if mode == "tamper" {
				fixtureFile(t, path, "changed")
			} else {
				if err := os.Remove(path); err != nil {
					t.Fatal(err)
				}
				other := filepath.Join(t.TempDir(), "plan.md")
				fixtureFile(t, other, "plan")
				if err := os.Symlink(other, path); err != nil {
					t.Fatal(err)
				}
			}
			if _, err := importHandoff(p, d); err == nil {
				t.Fatal("invalid provenance accepted")
			}
		})
	}
}
func TestHandoffRequiresCompletedProducer(t *testing.T) {
	p, d := handoffFixture(t, "plan")
	fixtureFile(t, filepath.Join(p, "run.json"), `{"Suite":"flows","Case":"plan","Terminal":"timeout"}`)
	if _, err := importHandoff(p, d); err == nil {
		t.Fatal("timeout accepted")
	}
}
func TestFlowDiscoveryIsNotReadEvidence(t *testing.T) {
	r := result{Trace: traceReport{DiscoveryPresent: true, Skills: []string{"flow-plan"}, Events: []traceEvent{{Kind: "final", Text: "I read flow-plan"}}}}
	got := observeSkill(r, "flow-plan")
	if got.Discovery.Status != "pass" || got.Read.Status != "not_observed" {
		t.Fatalf("conflated advertisement/read: %+v", got)
	}
	ok := true
	r.Trace.Events = append(r.Trace.Events, traceEvent{Kind: "read", Path: "/home/.agents/skills/flow-plan/SKILL.md", Success: &ok, Line: 9})
	if observeSkill(r, "flow-plan").Read.Status != "pass" {
		t.Fatal("successful read not recognized")
	}
}
func TestFlowsSourceWritesFailResearch(t *testing.T) {
	r := result{Suite: "flows", Case: "research", Terminal: "completed", Changed: []string{"src/items.ts"}, After: map[string]item{"_support/sessions/x/items.research.md": {}}}
	f := fixture{ID: "research"}
	f.Expected.SkillRead = "flow-research"
	if a := assessFlows(r, f, t.TempDir()); a.Status != "fail" {
		t.Fatalf("unauthorized write passed: %+v", a)
	}
}
func TestTypeScriptFixturesAndIndependentContract(t *testing.T) {
	if _, err := exec.LookPath("node"); err != nil {
		t.Skip("Node unavailable")
	}
	raw, err := os.ReadFile("../fixtures/flows/cases.json")
	if err != nil {
		t.Fatal(err)
	}
	var corpus struct {
		Cases []fixture `json:"cases"`
	}
	if err = json.Unmarshal(raw, &corpus); err != nil {
		t.Fatal(err)
	}
	for _, f := range corpus.Cases {
		if f.ID == "smoke" {
			continue
		}
		t.Run(f.ID, func(t *testing.T) {
			dir := t.TempDir()
			for p, b := range f.Files {
				fixtureFile(t, filepath.Join(dir, p), b)
			}
			cmd := exec.Command("node", "--test", "tests/items.test.ts")
			cmd.Dir = dir
			if b, err := cmd.CombinedOutput(); err != nil {
				t.Fatalf("fixture fails before model: %s %v", b, err)
			}
			got := verifyFlowContract(dir)
			want := "fail"
			if f.ID == "research" {
				want = "pass"
			}
			if got.Status != want {
				t.Fatalf("independent contract = %+v, want %s", got, want)
			}
			if f.ID == "research" {
				p := filepath.Join(dir, "src/items.ts")
				fixtureFile(t, p, strings.Replace(f.Files["src/items.ts"], "items.slice(0, limit)", "items.slice(0, limit).reverse()", 1))
				if verifyFlowContract(dir).Status != "fail" {
					t.Fatal("order regression escaped contract")
				}
			}
		})
	}
}

func TestHandoffRecursiveSupportReferences(t *testing.T) {
	p, d := handoffFixture(t, "# Plan\n[findings](items.research.md)\n")
	rel := "_support/sessions/2026-09-21-items/items.research.md"
	fixtureFile(t, filepath.Join(p, "final-files", rel), "# Findings\n[source](../../../src/items.ts)\n")
	fixtureFile(t, filepath.Join(d, "src/items.ts"), "original")
	inv, err := inventory(filepath.Join(p, "final-files"))
	if err != nil {
		t.Fatal(err)
	}
	raw, _ := json.Marshal(result{Suite: "flows", Case: "plan", Terminal: "completed", After: inv})
	fixtureFile(t, filepath.Join(p, "run.json"), string(raw))
	got, err := importHandoff(p, d)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Files) != 2 {
		t.Fatalf("missing companion: %+v", got)
	}
	if _, err = importHandoff(p, d); err == nil {
		t.Fatal("existing records overwritten")
	}
}
func TestAllManagedSkillsProtected(t *testing.T) {
	clearRoots(t)
	home := t.TempDir()
	before := protectedFor("grok", home)
	for _, name := range []string{"flow-research", "flow-plan", "flow-build"} {
		fixtureFile(t, filepath.Join(home, ".agents/skills", name, "SKILL.md"), name)
	}
	if got := changedProtected(before, protectedFor("grok", home)); len(got) != 3 {
		t.Fatalf("missing shared skills: %v", got)
	}
}

func TestPlanReferenceReadAndOrderingAreSeparateFromOutcome(t *testing.T) {
	yes := true
	ref := traceEvent{Kind: "read", Path: "/home/.agents/skills/flow-plan/references/plan-format.md", Success: &yes, Line: 2}
	write := traceEvent{Kind: "write", Path: "_support/sessions/2026-09-21-items/items.plan.md", Success: &yes, Line: 4}
	for _, tc := range []struct {
		name        string
		events      []traceEvent
		read, order string
	}{
		{"self-report", []traceEvent{{Kind: "final", Text: "I read plan-format.md"}}, "not_observed", "not_verified"},
		{"read-only", []traceEvent{ref}, "pass", "not_verified"},
		{"before-write", []traceEvent{ref, write}, "pass", "pass"},
		{"after-write", []traceEvent{write, {Kind: "read", Path: ref.Path, Success: &yes, Line: 6}}, "pass", "fail"},
		{"shell-write", []traceEvent{ref, {Kind: "shell", Command: "cat > _support/sessions/2026-09-21-items/items.plan.md", Success: &yes, Line: 5}}, "pass", "pass"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := result{Root: "/fixture", Cwd: "/fixture", Trace: traceReport{Events: tc.events}}
			got := observePlanReference(r)
			if got.Read.Status != tc.read || got.BeforeFirstPlanWrite.Status != tc.order {
				t.Fatalf("unexpected observation: %+v", got)
			}
		})
	}
}
func TestPlanReferenceAndClaudeAliasProtected(t *testing.T) {
	clearRoots(t)
	home := t.TempDir()
	canonical := filepath.Join(home, ".agents/skills/flow-plan")
	ref := filepath.Join(canonical, "references/plan-format.md")
	fixtureFile(t, filepath.Join(canonical, "SKILL.md"), "skill")
	fixtureFile(t, ref, "format one")
	alias := filepath.Join(home, ".claude/skills/flow-plan")
	if err := os.MkdirAll(filepath.Dir(alias), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("../../.agents/skills/flow-plan", alias); err != nil {
		t.Fatal(err)
	}
	before := protectedFor("codex", home)
	fixtureFile(t, ref, "format two")
	after := protectedFor("codex", home)
	changed := changedProtected(before, after)
	seen := map[string]bool{}
	for _, p := range changed {
		seen[p] = true
	}
	if !seen[canonical] || !seen[alias] {
		t.Fatalf("reference not protected through both paths: %v", changed)
	}
}

func TestOpenCodeNativeSkillSourceRequiresSuccessfulMatchingCall(t *testing.T) {
	yes, no := true, false
	body := "<skill_content name=\"flow-plan\">\n# Skill: flow-plan\n\n# Prepare a plan\n" + strings.Repeat("Substantive source instructions. ", 8) + "\n</skill_content>"
	for _, tc := range []struct {
		name         string
		call, output traceEvent
		want         string
	}{
		{"actual", traceEvent{Kind: "skill_invocation", ID: "1", Success: &yes}, traceEvent{Kind: "tool_result", ID: "1", Text: body, Success: &yes}, "pass"},
		{"failed", traceEvent{Kind: "skill_invocation", ID: "1", Success: &no}, traceEvent{Kind: "tool_result", ID: "1", Text: body, Success: &no}, "not_observed"},
		{"unrelated", traceEvent{Kind: "skill_invocation", ID: "2", Success: &yes}, traceEvent{Kind: "tool_result", ID: "1", Text: body, Success: &yes}, "not_observed"},
		{"advertisement", traceEvent{Kind: "skill_invocation", ID: "1", Success: &yes}, traceEvent{Kind: "tool_result", ID: "1", Text: "Launching skill: flow-plan", Success: &yes}, "not_observed"},
		{"selfclaim", traceEvent{Kind: "skill_invocation", ID: "1", Success: &yes}, traceEvent{Kind: "final", ID: "1", Text: body, Success: &yes}, "not_observed"},
		{"other-skill", traceEvent{Kind: "skill_invocation", ID: "1", Success: &yes}, traceEvent{Kind: "tool_result", ID: "1", Text: strings.ReplaceAll(body, "flow-plan", "flow-build"), Success: &yes}, "not_observed"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := result{Trace: traceReport{Events: []traceEvent{tc.call, tc.output}}}
			if got := observeSkill(r, "flow-plan").Read.Status; got != tc.want {
				t.Fatalf("got %s want %s", got, tc.want)
			}
		})
	}
}
func TestClaudeNativeSyntheticSkillSource(t *testing.T) {
	body := "Base directory for this skill: /home/.claude/skills/flow-plan\n\n# Prepare a plan\n" + strings.Repeat("Substantive source instructions. ", 8)
	for _, tc := range []struct {
		name            string
		synthetic       bool
		typ, body, want string
	}{
		{"actual", true, "user", body, "pass"},
		{"assistant-selfclaim", true, "assistant", body, "not_observed"},
		{"ordinary-user", false, "user", body, "not_observed"},
		{"launch-only", true, "user", "Launching skill: flow-plan", "not_observed"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			events := []map[string]any{
				{"type": "assistant", "message": map[string]any{"content": []any{map[string]any{"type": "tool_use", "id": "1", "name": "Skill", "input": map[string]any{"skill": "flow-plan"}}}}},
				{"type": "user", "message": map[string]any{"content": []any{map[string]any{"type": "tool_result", "tool_use_id": "1", "content": "Launching skill: flow-plan"}}}},
				{"type": tc.typ, "isSynthetic": tc.synthetic, "message": map[string]any{"content": []any{map[string]any{"type": "text", "text": tc.body}}}},
			}
			var lines []string
			for _, e := range events {
				b, _ := json.Marshal(e)
				lines = append(lines, string(b))
			}
			r := result{Trace: parseTrace("claude", strings.NewReader(strings.Join(lines, "\n")))}
			if got := observeSkill(r, "flow-plan").Read.Status; got != tc.want {
				t.Fatalf("got %s want %s", got, tc.want)
			}
		})
	}
}

func TestPlanReferenceLiteralShellReads(t *testing.T) {
	yes, no := true, false
	path := "/home/.claude/skills/flow-plan/references/plan-format.md"
	body := "# Retained plan format\n\n## What the document must carry\nSource body"
	for _, tc := range []struct {
		name, command, body string
		success             *bool
		want                string
	}{
		{"cat", "for f in source; do echo \"$f\"; done; cat " + path + "; node --version", body, &yes, "pass"},
		{"sed", "sed -n '1,260p' " + path, body, &yes, "pass"},
		{"failed", "cat " + path, body, &no, "not_observed"},
		{"echo", "echo 'cat " + path + "'", body, &yes, "not_observed"},
		{"echo-semicolon", "echo 'dummy; cat " + path + "'", body, &yes, "not_observed"},
		{"no-source-output", "cat " + path, "I read the reference", &yes, "not_observed"},
		{"variable", "cat \"$REF/flow-plan/references/plan-format.md\"", body, &yes, "not_observed"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := result{Root: "/fixture", Cwd: "/fixture", Trace: traceReport{Events: []traceEvent{
				{Kind: "shell", ID: "1", Command: tc.command, Success: tc.success, Line: 1},
				{Kind: "tool_result", ID: "1", Text: tc.body, Success: tc.success, Line: 2},
				{Kind: "write", Path: "_support/sessions/x/items.plan.md", Success: &yes, Line: 3},
			}}}
			got := observePlanReference(r)
			if got.Read.Status != tc.want {
				t.Fatalf("got %+v want %s", got, tc.want)
			}
			if tc.want == "pass" && got.BeforeFirstPlanWrite.Status != "pass" {
				t.Fatalf("ordering: %+v", got)
			}
		})
	}
}

func TestFlowTaskPromptBindsActualFixtureWithoutInjectingPolicy(t *testing.T) {
	root := "/repo/_support/workspace/round/claude-plan/fixture-123"
	original := "Investiga y conserva un plan."
	got := flowTaskPrompt(original, root)
	for _, required := range []string{original, root, "único repositorio de trabajo", "repositorio padre", "carpetas hermanas", "archivos del evaluador", "skills instaladas", "no cambia tus permisos", "ni constituye aislamiento"} {
		if !strings.Contains(got, required) {
			t.Fatalf("missing scope clause %q", required)
		}
	}
	if strings.Contains(got, "<topic>.plan.md") || strings.Contains(got, "_support/sessions") {
		t.Fatal("placement/naming policy duplicated in prompt")
	}
	if other := flowTaskPrompt(original, "/different/fixture"); other == got || strings.Contains(other, root) {
		t.Fatal("fixture root is hardcoded or leaked")
	}
}
func TestFlowAssessmentDetectsParentWriteOutsideInventory(t *testing.T) {
	yes := true
	r := result{Root: "/parent/fixture", Cwd: "/parent/fixture", Suite: "flows", Terminal: "completed", Trace: traceReport{Events: []traceEvent{{Kind: "write", Path: "/parent/_support/sessions/limit-items.research.md", Success: &yes, Line: 8}}}}
	f := fixture{ID: "research"}
	f.Expected.SkillRead = "flow-research"
	a := assessFlows(r, f, t.TempDir())
	found := false
	for _, c := range a.Criteria {
		if c.Criterion == "No observed writes outside fixture scope" {
			found = true
			if c.Status != "fail" {
				t.Fatal("parent write passed")
			}
		}
	}
	if !found {
		t.Fatal("missing external write criterion")
	}
}
