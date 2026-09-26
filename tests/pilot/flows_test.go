package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
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
		if f.ID != "research" {
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
			if got.Status != "pass" {
				t.Fatalf("independent contract = %+v, want pass", got)
			}
			p := filepath.Join(dir, "src/items.ts")
			fixtureFile(t, p, strings.Replace(f.Files["src/items.ts"], "items.slice(0, limit)", "items.slice(0, limit).reverse()", 1))
			if verifyFlowContract(dir).Status != "fail" {
				t.Fatal("order regression escaped contract")
			}
		})
	}
}

func TestNewFlowCasesHaveDistinctFixturesAndContracts(t *testing.T) {
	raw, err := os.ReadFile("../fixtures/flows/cases.json")
	if err != nil {
		t.Fatal(err)
	}
	var corpus struct {
		Cases []fixture `json:"cases"`
	}
	if err := json.Unmarshal(raw, &corpus); err != nil {
		t.Fatal(err)
	}
	byID := map[string]fixture{}
	for _, f := range corpus.Cases {
		byID[f.ID] = f
	}
	for _, id := range []string{"conventions-smoke", "deployed-smoke", "project-state", "adaptive-plan", "infra-plan", "direct-build", "git-delivery", "close-sequence"} {
		if _, ok := byID[id]; !ok {
			t.Fatalf("missing flow case %q", id)
		}
	}
	dir := t.TempDir()
	for p, b := range byID["direct-build"].Files {
		fixtureFile(t, filepath.Join(dir, p), b)
	}
	if got := verifyResourceNameContract(dir); got.Status != "fail" {
		t.Fatalf("unfixed resource-name fixture passed: %+v", got)
	}
	for _, tc := range []struct {
		name, source string
		want         string
	}{
		{"generic error accepted", "export function resourceName(project: string, component: string, environment: string): string {\n const parts = [project, component, environment].map((part) => part.trim().toLowerCase().replace(/\\s+/g, '-'));\n if (parts.some((part) => part === '')) throw new Error('empty segment');\n return parts.join('-');\n}\n", "pass"},
		{"type error accepted", "export function resourceName(project: string, component: string, environment: string): string {\n const parts = [project, component, environment].map((part) => part.trim().toLowerCase().replace(/\\s+/g, '-'));\n if (parts.some((part) => part === '')) throw new TypeError('empty segment');\n return parts.join('-');\n}\n", "pass"},
		{"missing rejection fails", "export function resourceName(project: string, component: string, environment: string): string {\n return [project, component, environment].map((part) => part.trim().toLowerCase().replace(/\\s+/g, '-')).join('-');\n}\n", "fail"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fixtureFile(t, filepath.Join(dir, "src/resource-names.ts"), tc.source)
			if got := verifyResourceNameContract(dir); got.Status != tc.want {
				t.Fatalf("contract = %+v, want %s", got, tc.want)
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
	for _, name := range []string{"flow-research", "flow-plan", "flow-build", "git-workflow"} {
		fixtureFile(t, filepath.Join(home, ".agents/skills", name, "SKILL.md"), name)
	}
	if got := changedProtected(before, protectedFor("grok", home)); len(got) != 4 {
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

func TestInfraReferenceReadAndOrderingAreSeparateFromOutcome(t *testing.T) {
	yes := true
	ref := traceEvent{Kind: "read", Path: "/home/.agents/skills/flow-plan/references/infra-naming.md", Success: &yes, Line: 2}
	write := traceEvent{Kind: "write", Path: "_support/sessions/2026-09-21-infra/media.plan.md", Success: &yes, Line: 4}
	r := result{Root: "/fixture", Cwd: "/fixture", Trace: traceReport{Events: []traceEvent{ref, write}}}
	got := observeInfraReference(r)
	if got.Read.Status != "pass" || got.BeforeFirstPlanWrite.Status != "pass" {
		t.Fatalf("unexpected observation: %+v", got)
	}
}
func TestPlanReferenceAndClaudeAliasProtected(t *testing.T) {
	clearRoots(t)
	home := t.TempDir()
	canonical := filepath.Join(home, ".agents/skills/flow-plan")
	ref := filepath.Join(canonical, "references/plan-format.md")
	infraRef := filepath.Join(canonical, "references/infra-naming.md")
	fixtureFile(t, filepath.Join(canonical, "SKILL.md"), "skill")
	fixtureFile(t, ref, "format one")
	fixtureFile(t, infraRef, "infra one")
	alias := filepath.Join(home, ".claude/skills/flow-plan")
	if err := os.MkdirAll(filepath.Dir(alias), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("../../.agents/skills/flow-plan", alias); err != nil {
		t.Fatal(err)
	}
	before := protectedFor("codex", home)
	fixtureFile(t, ref, "format two")
	fixtureFile(t, infraRef, "infra two")
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

func TestGitDeliverySetupAndInspectionAreScoped(t *testing.T) {
	root := t.TempDir()
	for p, body := range map[string]string{
		"src/resource-names.ts":        "export const resourceName = () => 'old';\n",
		"tests/resource-names.test.ts": "export {};\n",
		"notes/unrelated-staged.md":    "Keep this staged work unchanged.\n",
		"notes/unrelated-unstaged.md":  "Keep this unstaged work unchanged.\n",
	} {
		fixtureFile(t, filepath.Join(root, p), body)
	}
	if err := initGit(root); err != nil {
		t.Fatal(err)
	}
	before, err := setupGitDelivery(root)
	if err != nil {
		t.Fatal(err)
	}
	fixtureFile(t, filepath.Join(root, "src/resource-names.ts"), "export const resourceName = () => 'hive-api-qa';\n")
	for _, args := range [][]string{{"add", "src/resource-names.ts"}, {"commit", "-qm", "fix: format resource name", "--", "src/resource-names.ts"}, {"push", "fixture", "main"}} {
		if _, err := gitOutput(root, args...); err != nil {
			t.Fatal(err)
		}
	}
	after := inspectGitDelivery(root, before)
	if !after.InitialHeadAncestor || !after.RemoteMatchesHead || !after.StagedPreserved || !after.UnstagedPreserved {
		t.Fatalf("invalid delivery inspection: %+v", after)
	}
	if len(after.CommitPaths) != 1 || after.CommitPaths[0] != "src/resource-names.ts" {
		t.Fatalf("unexpected commit paths: %v", after.CommitPaths)
	}
}

func TestGitDeliveryInspectionRetainsRevertedRangePaths(t *testing.T) {
	root := t.TempDir()
	for p, body := range map[string]string{
		"src/resource-names.ts":        "export const resourceName = () => 'old';\n",
		"tests/resource-names.test.ts": "export {};\n",
		"notes/unrelated-staged.md":    "Keep this staged work unchanged.\n",
		"notes/unrelated-unstaged.md":  "Keep this unstaged work unchanged.\n",
		"README.md":                    "original\n",
	} {
		fixtureFile(t, filepath.Join(root, p), body)
	}
	if err := initGit(root); err != nil {
		t.Fatal(err)
	}
	before, err := setupGitDelivery(root)
	if err != nil {
		t.Fatal(err)
	}
	for _, body := range []string{"unauthorized but later reverted\n", "original\n"} {
		fixtureFile(t, filepath.Join(root, "README.md"), body)
		for _, args := range [][]string{{"add", "README.md"}, {"commit", "-qm", "test: temporary README change", "--", "README.md"}} {
			if _, err := gitOutput(root, args...); err != nil {
				t.Fatal(err)
			}
		}
	}
	after := inspectGitDelivery(root, before)
	if !slices.Contains(after.CommitPaths, "README.md") {
		t.Fatalf("reverted path escaped range audit: %+v", after)
	}
}

func closeSequenceFixtureRoot(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	for p, body := range map[string]string{
		"src/greeting.ts":             "export const greet = () => 'old';\n",
		"tests/greeting.test.ts":      "export {};\n",
		"notes/unrelated-staged.md":   "Keep this staged work unchanged.\n",
		"notes/unrelated-unstaged.md": "Keep this unstaged work unchanged.\n",
	} {
		fixtureFile(t, filepath.Join(root, p), body)
	}
	if err := initGit(root); err != nil {
		t.Fatal(err)
	}
	if _, err := setupGitDelivery(root); err != nil {
		t.Fatal(err)
	}
	return root
}

func TestCloseSequenceBranchAbsentDetectsLeftoverLocalAndRemoteBranches(t *testing.T) {
	root := closeSequenceFixtureRoot(t)
	r := result{Root: root}
	// Finding 2: no leftover fix/* branch is not enough on its own — a run
	// that never created or merged one at all must not pass either.
	if got := closeSequenceBranchAbsent(r); got.Status != "fail" {
		t.Fatalf("no fix/* branch and no merge observed must fail: %+v", got)
	}
	if _, err := gitOutput(root, "checkout", "-qb", "fix/greeting"); err != nil {
		t.Fatal(err)
	}
	if _, err := gitOutput(root, "push", "-qu", "fixture", "fix/greeting"); err != nil {
		t.Fatal(err)
	}
	if got := closeSequenceBranchAbsent(r); got.Status != "fail" {
		t.Fatalf("leftover local and remote branch must fail: %+v", got)
	}
	for _, args := range [][]string{{"checkout", "-q", "main"}, {"branch", "-D", "fix/greeting"}, {"push", "fixture", "--delete", "fix/greeting"}} {
		if _, err := gitOutput(root, args...); err != nil {
			t.Fatal(err)
		}
	}
	if got := closeSequenceBranchAbsent(r); got.Status != "pass" {
		t.Fatalf("deleted branch must pass again: %+v", got)
	}
}

func TestCloseSequenceCleanupLabelRequiresCaseInsensitiveLine(t *testing.T) {
	for _, tc := range []struct{ name, text, want string }{
		{"spanish label", "Reporte final.\nLimpieza: se eliminó la rama fix/greeting.\n¿Seguimos?", "pass"},
		{"english label", "Final report.\nCLEANUP: removed the temporary branch.", "pass"},
		{"missing label", "Reporte final sin mención del tema.", "fail"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := result{Trace: traceReport{Events: []traceEvent{{Kind: "text", Role: "assistant", Text: tc.text}}}}
			if got := closeSequenceCleanupLabel(r); got.Status != tc.want {
				t.Fatalf("got %+v want %s", got, tc.want)
			}
		})
	}
	if got := closeSequenceCleanupLabel(result{}); got.Status != "not_observed" {
		t.Fatalf("no assistant text must be not_observed: %+v", got)
	}
}

// TestCloseSequenceCleanupLabelChecksTextAfterLastWorkEvent is finding 8:
// checking only the single, literal last assistant text missed the shape
// where the report (with the cleanup label) and a separate close question
// are two different messages — closeQuestionAfterReport's own
// unanswered-question shortcut can leave a short, label-free continuation as
// the true last text event even though the report already stated the
// cleanup.
func TestCloseSequenceCleanupLabelChecksTextAfterLastWorkEvent(t *testing.T) {
	yes := true
	for _, tc := range []struct {
		name   string
		events []traceEvent
		want   string
	}{
		{
			name: "cleanup mentioned in the report; trailing text after an unanswered close question omits it",
			events: []traceEvent{
				{Kind: "shell", Command: "git branch -d fix/greeting", Success: &yes},
				{Kind: "tool_result", Success: &yes},
				{Kind: "text", Role: "assistant", Text: "Reporte final.\nLimpieza: se elimino la rama fix/greeting."},
				{Kind: "question", Tool: "ask_user_question", ID: "q1"},
				{Kind: "tool_result", ID: "q1"},
				{Kind: "text", Role: "assistant", Text: "Como no llego respuesta, dejo el resumen sin mas detalle."},
			},
			want: "pass",
		},
		{
			name: "neither the report nor the trailing continuation mentions cleanup",
			events: []traceEvent{
				{Kind: "shell", Command: "git branch -d fix/greeting", Success: &yes},
				{Kind: "tool_result", Success: &yes},
				{Kind: "text", Role: "assistant", Text: "Reporte final sin mencion del tema."},
				{Kind: "question", Tool: "ask_user_question", ID: "q1"},
				{Kind: "tool_result", ID: "q1"},
				{Kind: "text", Role: "assistant", Text: "Como no llego respuesta, dejo el resumen sin mas detalle."},
			},
			want: "fail",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := result{Trace: traceReport{Events: tc.events}}
			if got := closeSequenceCleanupLabel(r).Status; got != tc.want {
				t.Fatalf("got %s want %s", got, tc.want)
			}
		})
	}
}

// TestFixBranchEverExistedDetectsCreationEvenAfterDeletion is finding 2's
// existence signal in isolation: HEAD's own reflog records a fix/* checkout
// or merge and survives after the branch ref itself is deleted.
func TestFixBranchEverExistedDetectsCreationEvenAfterDeletion(t *testing.T) {
	root := closeSequenceFixtureRoot(t)
	if fixBranchEverExisted(root) {
		t.Fatal("a fresh fixture with no fix/* branch must report false")
	}
	if _, err := gitOutput(root, "checkout", "-qb", "fix/greeting"); err != nil {
		t.Fatal(err)
	}
	if !fixBranchEverExisted(root) {
		t.Fatal("a just-created fix/* branch must be detected")
	}
	for _, args := range [][]string{{"checkout", "-q", "main"}, {"branch", "-D", "fix/greeting"}} {
		if _, err := gitOutput(root, args...); err != nil {
			t.Fatal(err)
		}
	}
	if !fixBranchEverExisted(root) {
		t.Fatal("a deleted fix/* branch must still be detected via HEAD's own reflog")
	}
}

// TestCloseSequenceFlowBuildReadIsObservationalNotGating is finding 9: the
// flow-build read record must never fail the close-sequence case, whether or
// not the run actually read it, and it must report "pass" when a native read
// of flow-build/SKILL.md is present in the trace, distinct from git-workflow.
func TestCloseSequenceFlowBuildReadIsObservationalNotGating(t *testing.T) {
	root := closeSequenceFixtureRoot(t)
	f := fixture{ID: "close-sequence"}
	f.Expected.SkillRead = "git-workflow"
	base := result{Root: root, Cwd: root, Suite: "flows", Case: "close-sequence", Terminal: "completed", Trace: traceReport{Events: []traceEvent{{Kind: "text", Role: "assistant", Text: "Reporte final.\nLimpieza completada.\n¿Continuamos con otra tarea?"}}}}
	a := assessFlows(base, f, t.TempDir())
	byName := map[string]criterionAssessment{}
	for _, c := range a.Criteria {
		byName[c.Criterion] = c
	}
	got, ok := byName["close_sequence_flow_build_read"]
	if !ok {
		t.Fatal("close_sequence_flow_build_read criterion missing")
	}
	if got.Status != "not_observed" {
		t.Fatalf("expected not_observed with no flow-build read in the trace: %+v", got)
	}
	if got.Status == "fail" {
		t.Fatal("close_sequence_flow_build_read must never fail; it is observational")
	}
	yes := true
	withRead := base
	withRead.Trace.Events = append([]traceEvent{{Kind: "read", Path: "/home/.agents/skills/flow-build/SKILL.md", Success: &yes}}, base.Trace.Events...)
	a2 := assessFlows(withRead, f, t.TempDir())
	for _, c := range a2.Criteria {
		if c.Criterion == "close_sequence_flow_build_read" {
			if c.Status != "pass" {
				t.Fatalf("expected pass with an observed flow-build read: %+v", c)
			}
			return
		}
	}
	t.Fatal("close_sequence_flow_build_read criterion missing from the second assessment")
}

// closeSequenceGitDeliveryRoot builds a close-sequence fixture, applies the
// task's own fix/<slug> branch/merge sequence with the given commit path and
// message, and returns assessFlows' resulting "Scoped local Git delivery
// preserves unrelated index and worktree changes" criterion — finding 3: the
// same assertions setupGitDelivery/inspectGitDelivery already give
// git-delivery must also apply to close-sequence.
func closeSequenceGitDeliveryRoot(t *testing.T, commitPath, commitBody, commitMessage string) criterionAssessment {
	t.Helper()
	root := t.TempDir()
	for p, body := range map[string]string{
		"src/greeting.ts":             "export const greet = () => 'old';\n",
		"tests/greeting.test.ts":      "export {};\n",
		"notes/unrelated-staged.md":   "Keep this staged work unchanged.\n",
		"notes/unrelated-unstaged.md": "Keep this unstaged work unchanged.\n",
	} {
		fixtureFile(t, filepath.Join(root, p), body)
	}
	if err := initGit(root); err != nil {
		t.Fatal(err)
	}
	before, err := setupGitDelivery(root)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := gitOutput(root, "checkout", "-qb", "fix/greeting"); err != nil {
		t.Fatal(err)
	}
	fixtureFile(t, filepath.Join(root, commitPath), commitBody)
	for _, args := range [][]string{
		{"add", commitPath},
		{"commit", "-qm", commitMessage, "--", commitPath},
		{"push", "-qu", "fixture", "fix/greeting"},
		{"checkout", "-q", "main"},
		{"merge", "--ff-only", "fix/greeting"},
		{"push", "fixture", "main"},
	} {
		if _, err := gitOutput(root, args...); err != nil {
			t.Fatal(err)
		}
	}
	after := inspectGitDelivery(root, before)
	r := result{Root: root, Cwd: root, Suite: "flows", Case: "close-sequence", Terminal: "completed", GitDelivery: after, Trace: traceReport{Events: []traceEvent{{Kind: "text", Role: "assistant", Text: "Reporte final.\nLimpieza completada.\n¿Continuamos con otra tarea?"}}}}
	f := fixture{ID: "close-sequence"}
	f.Expected.SkillRead = "git-workflow"
	a := assessFlows(r, f, t.TempDir())
	for _, c := range a.Criteria {
		if c.Criterion == "Scoped local Git delivery preserves unrelated index and worktree changes" {
			return c
		}
	}
	t.Fatal("git delivery criterion missing for close-sequence")
	return criterionAssessment{}
}

func TestCloseSequenceGitDeliveryAssertionsMirrorGitDelivery(t *testing.T) {
	got := closeSequenceGitDeliveryRoot(t, "src/greeting.ts", "export const greet = (name) => `Hello, ${name}!`;\n", "fix: punctuate greeting")
	if got.Status != "pass" {
		t.Fatalf("expected pass for a clean fix/greeting merge: %+v", got)
	}
}

func TestCloseSequenceGitDeliveryFlagsUnrelatedNotesCommit(t *testing.T) {
	got := closeSequenceGitDeliveryRoot(t, "notes/unrelated-staged.md", "modified by mistake\n", "oops: touched notes")
	if got.Status != "fail" {
		t.Fatalf("expected fail for a commit touching notes/*: %+v", got)
	}
}

func TestAssessFlowsWiresCloseSequenceCriteria(t *testing.T) {
	root := closeSequenceFixtureRoot(t)
	r := result{
		Root: root, Cwd: root, Suite: "flows", Case: "close-sequence", Terminal: "completed",
		Trace: traceReport{Events: []traceEvent{{Kind: "text", Role: "assistant", Text: "Reporte final.\nLimpieza completada.\n¿Continuamos con otra tarea?"}}},
	}
	f := fixture{ID: "close-sequence"}
	f.Expected.SkillRead = "git-workflow"
	a := assessFlows(r, f, t.TempDir())
	byName := map[string]string{}
	for _, c := range a.Criteria {
		byName[c.Criterion] = c.Status
	}
	if byName["close_question_after_report"] != "pass" {
		t.Fatalf("close_question_after_report missing or wrong: %v", byName)
	}
	// Finding 2: close-sequence requires an observed merge; a raw
	// not_observed from mergedBranchDeleted must not let the case pass.
	if byName["merged_branch_deleted"] != "fail" {
		t.Fatalf("merged_branch_deleted expected fail (no merge observed, but close-sequence requires one): %v", byName)
	}
	if byName["close_sequence_branch_absent"] != "fail" {
		t.Fatalf("close_sequence_branch_absent expected fail (no fix/* branch or merge ever observed): %v", byName)
	}
	if byName["close_sequence_cleanup_label"] != "pass" {
		t.Fatalf("close_sequence_cleanup_label expected pass: %v", byName)
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

func TestDeployedSmokeForbidsFixtureWritesAndNamesNewSkills(t *testing.T) {
	want := []string{"flow-report", "engram-init-workspace", "starlight-docs-site", "unattended-delegation", "workspace-archive"}
	if !slices.Equal(flowLoadingSkills("deployed-smoke"), want) {
		t.Fatalf("skills %v", flowLoadingSkills("deployed-smoke"))
	}
	a := assessFlows(result{Case: "deployed-smoke", Terminal: "completed", Changed: []string{"README.md"}}, fixture{ID: "deployed-smoke"}, t.TempDir())
	if a.Status == "pass" {
		t.Fatal("unread skills or fixture write passed")
	}
	if len(a.Skills) != len(want) {
		t.Fatalf("observed %d skills", len(a.Skills))
	}
	found := false
	for _, c := range a.Criteria {
		if c.Criterion == "Final writes within authorized fixture scope" {
			found = true
			if c.Status != "fail" {
				t.Fatalf("write allowed: %+v", c)
			}
		}
	}
	if !found {
		t.Fatal("missing write criterion")
	}
}
