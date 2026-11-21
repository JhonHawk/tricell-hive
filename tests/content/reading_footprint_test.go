package content

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"unicode/utf8"
)

type readingSize struct{ bytes, characters int }

// readingUnion counts each declared source once, not each request or reader.
func readingUnion(paths []string, read func(string) ([]byte, error)) (readingSize, error) {
	var size readingSize
	seen := map[string]bool{}
	for _, path := range paths {
		if seen[path] {
			continue
		}
		data, err := read(path)
		if err != nil {
			return readingSize{}, fmt.Errorf("required source %s: %w", path, err)
		}
		seen[path] = true
		size.bytes += len(data)
		size.characters += utf8.RuneCount(data)
	}
	return size, nil
}

func TestGuidanceReadingFootprintUnion(t *testing.T) {
	reads := map[string]int{}
	got, err := readingUnion([]string{"one", "two", "one"}, func(path string) ([]byte, error) {
		reads[path]++
		return []byte("é"), nil
	})
	if err != nil || got != (readingSize{4, 2}) || reads["one"] != 1 || reads["two"] != 1 {
		t.Fatalf("unique union: got %+v, reads %v, error %v", got, reads, err)
	}
	if _, err := readingUnion([]string{"missing"}, func(string) ([]byte, error) {
		return nil, os.ErrNotExist
	}); err == nil {
		t.Fatal("missing required source must fail measurement")
	}
}

func TestGuidanceReadingFootprint(t *testing.T) {
	const global = "content/guidance/global.md"
	const skills = "content/skills/"
	// Predicates are explicit scenarios, not a transitive inventory of links.
	common := []string{global}
	// The question rules live in rule leaves the index and the flow steps link.
	question := []string{global, skills + "hive-rule-questions/SKILL.md"}
	cycle := append(append([]string{}, question...),
		skills+"hive-rule-delivery-ask/SKILL.md", skills+"hive-rule-project-settings/SKILL.md", skills+"hive-rule-ticket-state/SKILL.md",
		skills+"flow-plan/SKILL.md", skills+"flow-build/SKILL.md", skills+"flow-close/SKILL.md",
		skills+"hive-mode/SKILL.md", skills+"flow-verify/SKILL.md",
		"content/agents/review/hive-review-plan.md",
		"content/agents/development/hive-build-backend.md",
		"content/agents/quality/hive-verify-task.md",
	)
	// Minimal skills (skills-rebuild): sources absent from the baseline commit.
	moved := map[string]bool{
		skills + "flow-verify/SKILL.md":                true,
		skills + "hive-rule-questions/SKILL.md":        true,
		skills + "hive-rule-delivery-ask/SKILL.md":     true,
		skills + "hive-rule-project-settings/SKILL.md": true,
		skills + "hive-rule-ticket-state/SKILL.md":     true,
	}
	scenarios := []struct {
		name, predicate string
		paths           []string
	}{
		{"global-only", "kernel alone; not a complete chat reply", []string{global}},
		{"simple-query", "ordinary factual reply; no activity invocation, decision, delegation or artifact", common},
		{"question", "standalone user decision with native card or text; no activity invocation or artifact", question},
		{"plan-build-cycle", "retained backend runtime plan; resolve settings and hold/review choice; read-only plan reviewer, delegated backend implementation and task verifier; local tests; pause via Close; no UI, security/data migration, deployment, Git execution or incidental finding", cycle},
	}
	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			t.Logf("predicate: %s", scenario.predicate)
			var baseline []string
			for _, path := range scenario.paths {
				if !moved[path] {
					baseline = append(baseline, path)
				}
			}
			measure := func(label string, paths []string, read func(string) ([]byte, error)) readingSize {
				t.Helper()
				size, err := readingUnion(paths, func(path string) ([]byte, error) {
					data, err := read(path)
					if err == nil {
						t.Logf("%s source %s: %d bytes, %d characters", label, path, len(data), utf8.RuneCount(data))
					}
					return data, err
				})
				if err != nil {
					t.Fatal(err)
				}
				return size
			}
			before := measure("A d5164f71", baseline, func(path string) ([]byte, error) {
				return exec.Command("git", "show", "d5164f71:"+baselinePath(path)).Output()
			})
			after := measure("B working tree", scenario.paths, func(path string) ([]byte, error) {
				return os.ReadFile("../../" + path)
			})
			t.Logf("A union: %d bytes, %d characters; B/C declared source union: %d bytes, %d characters; byte change: %+d (%.2f%%)", before.bytes, before.characters, after.bytes, after.characters, after.bytes-before.bytes, 100*float64(after.bytes-before.bytes)/float64(before.bytes))
		})
	}
	t.Log("C uses B's canonical sources; mapped plan-review injection adds no unique source to this cycle. Raw source frontmatter is counted. This is not per-request context or actual tokens. Native rendering and skill/role discovery metadata, plugin metadata/wrappers, repeated text across readers/requests, history, project inputs and host policy are not measured. Pilots must observe delivery and token usage separately.")
}

// baselinePath maps a current path to the path it had at a baseline commit:
// hive-mode was flow-research there, so its entry compares with the old one.
func baselinePath(path string) string {
	return strings.Replace(path, "content/skills/hive-mode/", "content/skills/flow-research/", 1)
}

// Compare the same pre-existing sources on both sides; new sources are separate.
// The legacy scenario above remains identifiable rather than being silently expanded.
func TestRequiredProcedureReadingFootprint(t *testing.T) {
	const base = "2bb96478"
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	skills := "content/skills/"
	cycle := []string{
		"content/guidance/global.md",
		skills + "hive-rule-questions/SKILL.md",
		skills + "hive-rule-delivery-ask/SKILL.md", skills + "hive-rule-project-settings/SKILL.md",
		skills + "hive-rule-ticket-state/SKILL.md",
		skills + "hive-rule-close-modes/SKILL.md", skills + "hive-rule-close-hygiene/SKILL.md", skills + "hive-rule-close-report/SKILL.md",
		skills + "hive-mode/SKILL.md", skills + "flow-plan/SKILL.md",
		skills + "flow-build/SKILL.md", skills + "flow-verify/SKILL.md", skills + "flow-close/SKILL.md",
		skills + "hive-rule-delegation/SKILL.md", skills + "hive-rule-child-brief/SKILL.md", skills + "hive-rule-reviewer-brief/SKILL.md",
		skills + "flow-plan/references/plan-handoff.md",
		skills + "hive-rule-business-rules/SKILL.md", skills + "hive-rule-business-record/SKILL.md", skills + "hive-rule-business-ticket/SKILL.md",
		skills + "project-layout/references/layout.md", skills + "flow-close/references/support-material.md",
		skills + "hive-rule-project-verification/SKILL.md", skills + "hive-rule-project-verification-format/SKILL.md", skills + "hive-rule-suppression-review/SKILL.md",
		skills + "hive-rule-reuse-before-writing/SKILL.md", skills + "hive-rule-tests/SKILL.md", skills + "hive-rule-checkpoint/SKILL.md",
		skills + "hive-rule-checks/SKILL.md", skills + "hive-rule-change-folder/SKILL.md", skills + "hive-rule-review-findings/SKILL.md",
		skills + "hive-rule-review-evidence/SKILL.md",
		skills + "hive-rule-report-endings/SKILL.md", skills + "hive-rule-git-commit/SKILL.md", skills + "hive-rule-merge-gate/SKILL.md",
		"content/agents/review/hive-review-plan.md", "content/agents/development/hive-build-backend.md",
		"content/agents/quality/hive-verify-task.md", "content/agents/review/hive-review-code.md",
	}
	e2e := append(append([]string{}, cycle...),
		skills+"flow-build/references/e2e-coordination.md", skills+"flow-plan/references/execution-record.md",
		skills+"hive-rule-defect-fix/SKILL.md", skills+"git-workflow/SKILL.md",
		skills+"flow-close/references/branch-cleanup.md", "content/agents/development/hive-build-owner.md",
		skills+"hive-rule-pr-reviews/SKILL.md",
	)
	// baselineCycle is the plan-build-required list as 2bb96478 declared it: the
	// sources the same scenario read before the rules moved into leaves.
	baselineCycle := []string{
		"content/guidance/global.md",
		skills + "flow-research/SKILL.md", skills + "flow-plan/SKILL.md",
		skills + "flow-build/SKILL.md", skills + "flow-verify/SKILL.md", skills + "flow-close/SKILL.md",
		skills + "flow-build/references/delegation.md",
		skills + "flow-plan/references/plan-handoff.md", skills + "flow-plan/references/business-rules.md",
		skills + "project-layout/references/layout.md", skills + "flow-close/references/support-material.md",
		skills + "flow-verify/references/project-verification.md", skills + "flow-verify/references/suppression-review.md",
		"content/agents/review/hive-review-plan.md", "content/agents/development/hive-build-backend.md",
		"content/agents/quality/hive-verify-task.md", "content/agents/review/hive-review-code.md",
	}
	// A source is new when its baseline path does not exist at the base commit.
	added := func(path string) bool {
		return exec.Command("git", "-C", root, "cat-file", "-e", base+":"+baselinePath(path)).Run() != nil
	}
	for _, scenario := range []struct {
		name, predicate string
		paths           []string
	}{
		{"plan-build-required", "retained backend plan, delegated implementation and independent reviews; project verification contract maintained; hold and pause; no UI, external system, data migration, defect correction or unattended authority", cycle},
		{"e2e", "same backend plan with explicit multiunit E2E grant, local owners, correction, replacement, independent verification and authorized Git integration; no UI, data migration, deployment or unattended authority", e2e},
		{"executable-resources", "execution rather than mandatory source reading: task/plan, project contract, scoped Git and cleanup helpers; never added to the prose union", []string{skills + "flow-plan/scripts/tasks.py", skills + "flow-plan/scripts/check_change.py", skills + "flow-verify/scripts/check_contract.py", skills + "git-workflow/scripts/git_state.py", skills + "flow-close/scripts/cleanup_branches.py"}},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			t.Logf("predicate: %s", scenario.predicate)
			before, comparable, fresh, prose, references, code := 0, 0, 0, 0, 0, 0
			_, err := readingUnion(scenario.paths, func(path string) ([]byte, error) {
				data, err := os.ReadFile(filepath.Join(root, path))
				if err != nil {
					return nil, err
				}
				kind := "prose"
				if strings.Contains(path, "/references/") {
					kind = "reference"
					references += len(data)
				} else if strings.HasSuffix(path, ".py") {
					kind = "code"
					code += len(data)
				} else {
					prose += len(data)
				}
				if added(path) {
					fresh += len(data)
					t.Logf("new %s %s: %d B", kind, path, len(data))
				} else {
					old, err := exec.Command("git", "-C", root, "show", base+":"+baselinePath(path)).Output()
					if err != nil {
						return nil, fmt.Errorf("baseline %s: %w", path, err)
					}
					before += len(old)
					comparable += len(data)
					t.Logf("same-path %s %s: %d -> %d B", kind, path, len(old), len(data))
				}
				return data, nil
			})
			if err != nil {
				t.Fatal(err)
			}
			t.Logf("same-path baseline %s: %d B -> %d B (%+d); new sources %d B; total %d B; prose %d B; references %d B; executable code %d B", base, before, comparable, comparable-before, fresh, comparable+fresh, prose, references, code)
			if scenario.name == "plan-build-required" {
				declared, err := readingUnion(baselineCycle, func(path string) ([]byte, error) {
					return exec.Command("git", "-C", root, "show", base+":"+path).Output()
				})
				if err != nil {
					t.Fatal(err)
				}
				// The baseline reads files that no longer exist; the candidate reads
				// their rules through leaves, which the new-source line above reports apart.
				// git-commit and merge-gate select moments the baseline list never read
				// (it held no git-workflow), so they are reported apart too and the
				// comparison with the baseline is made without them.
				extra := 0
				for _, path := range []string{skills + "hive-rule-git-commit/SKILL.md", skills + "hive-rule-merge-gate/SKILL.md"} {
					data, err := os.ReadFile(filepath.Join(root, path))
					if err != nil {
						t.Fatal(err)
					}
					extra += len(data)
				}
				like := comparable + fresh - extra
				t.Logf("plan/build union at %s (its declared list): %d B; candidate with every selected leaf: %d B (%+d B, %+.2f%%), of which new sources %d B", base, declared.bytes, comparable+fresh, comparable+fresh-declared.bytes, 100*float64(comparable+fresh-declared.bytes)/float64(declared.bytes), fresh)
				t.Logf("moments the baseline list did not read (git-commit, merge-gate): %d B; candidate without them: %d B (%+d B, %+.2f%%)", extra, like, like-declared.bytes, 100*float64(like-declared.bytes)/float64(declared.bytes))
				// The change conditional-plan-reviewers grew the union by 182 B after its
				// cuts (the domain condition and counting rules of flow-plan step 5 and the
				// plan reviewers' reread in reviewer-brief); the user accepted that growth
				// on 2026-11-21.
				const planReviewAllowance = 182
				// The change reporting-family-leaves moved the text of next-step-block and
				// close-footer, which this list did not read, into report-endings and
				// close-report, which it reads whole at each stage's end; the user accepted
				// that growth on 2026-11-21 (D12 A).
				const reportingFamilyAllowance = 8970
				allowance := planReviewAllowance + reportingFamilyAllowance
				if like >= declared.bytes+allowance {
					t.Errorf("plan-build-required union without the moments the baseline did not read is %d B, not below the baseline %d B plus the %d B allowance", like, declared.bytes, allowance)
				}
			}
		})
	}
	t.Log("Explicit source unions, not transitive loading or observed tokens. Excludes native metadata/rendering, repeated reads, project inputs, history and host policies. Executable resources are measured separately, not presumed read by a model. No aggregate quota; each entrypoint/reference retains its 16000-byte check.")
}
