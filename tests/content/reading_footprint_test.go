package content

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
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
	question := []string{global, skills + "hive-core/references/questions.md"}
	cycle := append(append([]string{}, question...),
		skills+"hive-core/references/delivery-ask.md", skills+"hive-core/references/project-settings.md", skills+"hive-core/references/ticket-state.md",
		skills+"flow-plan/SKILL.md", skills+"flow-build/SKILL.md", skills+"flow-close/SKILL.md",
		skills+"hive-mode/SKILL.md", skills+"flow-verify/SKILL.md",
		"content/agents/review/hive-review-plan.md",
		"content/agents/development/hive-build-backend.md",
		"content/agents/quality/hive-verify-task.md",
	)
	// Minimal skills (skills-rebuild): sources absent from the baseline commit.
	moved := map[string]bool{
		skills + "flow-verify/SKILL.md":                     true,
		skills + "hive-core/references/questions.md":        true,
		skills + "hive-core/references/delivery-ask.md":     true,
		skills + "hive-core/references/project-settings.md": true,
		skills + "hive-core/references/ticket-state.md":     true,
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
		skills + "hive-core/references/questions.md",
		skills + "hive-core/references/delivery-ask.md", skills + "hive-core/references/project-settings.md",
		skills + "hive-core/references/ticket-state.md",
		skills + "hive-core/references/close-modes.md", skills + "hive-core/references/close-hygiene.md", skills + "hive-core/references/close-report.md",
		skills + "hive-mode/SKILL.md", skills + "flow-plan/SKILL.md",
		skills + "flow-build/SKILL.md", skills + "flow-verify/SKILL.md", skills + "flow-close/SKILL.md",
		skills + "hive-core/references/delegation.md", skills + "hive-core/references/child-brief.md", skills + "hive-core/references/reviewer-brief.md",
		skills + "flow-plan/references/plan-handoff.md",
		skills + "hive-core/references/business-rules.md", skills + "hive-core/references/business-record.md", skills + "hive-core/references/business-ticket.md",
		skills + "project-layout/references/layout.md", skills + "flow-close/references/support-material.md",
		skills + "hive-core/references/project-verification.md", skills + "hive-core/references/project-verification-format.md", skills + "hive-core/references/suppression-review.md",
		skills + "hive-core/references/reuse-before-writing.md", skills + "hive-core/references/tests.md", skills + "hive-core/references/checkpoint.md",
		skills + "hive-core/references/checks.md", skills + "hive-core/references/change-folder.md", skills + "hive-core/references/review-findings.md",
		skills + "hive-core/references/review-evidence.md",
		skills + "hive-core/references/report-endings.md", skills + "hive-core/references/git-commit.md", skills + "hive-core/references/merge-gate.md",
		"content/agents/review/hive-review-plan.md", "content/agents/development/hive-build-backend.md",
		"content/agents/quality/hive-verify-task.md", "content/agents/review/hive-review-code.md",
	}
	e2e := append(append([]string{}, cycle...),
		skills+"flow-build/references/e2e-coordination.md", skills+"flow-plan/references/execution-record.md",
		skills+"hive-core/references/defect-fix.md", skills+"git-workflow/SKILL.md",
		skills+"flow-close/references/branch-cleanup.md", "content/agents/development/hive-build-owner.md",
		skills+"hive-core/references/pr-reviews.md",
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
				for _, path := range []string{skills + "hive-core/references/git-commit.md", skills + "hive-core/references/merge-gate.md"} {
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
				// that growth on 2026-11-21 (D12 A), measured again after the fixes of
				// its review findings.
				const reportingFamilyAllowance = 9230
				// The change reuse-sound-base grew rung 2 of reuse-before-writing and the
				// code reviewer's check of what new code reuses; the user accepted that
				// growth on 2026-11-21 (D1 A); measured again after decision D2 A, on
				// development at 01efd12c.
				const reuseSoundBaseAllowance = 1111
				// The change change-folder-versioned-check added the pointer of flow-close
				// step 3 to `tasks.py versioned`, whose output carries the rest of the
				// procedure; the user accepted that growth on 2026-11-21 (D4 A).
				const changeFolderVersionedAllowance = 125
				// The change project-skill-registry-mobile added the brief's copy of a
				// project's ## Skills lines to child-brief and reviewer-brief; the user
				// accepted that growth on 2026-11-21 (D6 A), on development at 867f5463.
				const skillRegistryAllowance = 353
				allowance := planReviewAllowance + reportingFamilyAllowance + reuseSoundBaseAllowance + changeFolderVersionedAllowance + skillRegistryAllowance
				if like >= declared.bytes+allowance {
					t.Errorf("plan-build-required union without the moments the baseline did not read is %d B, not below the baseline %d B plus the %d B allowance", like, declared.bytes, allowance)
				}
			}
		})
	}
	t.Run("plan-build-packets", func(t *testing.T) {
		t.Logf("predicate: the plan-build-required route when each flow it enters reads its stage packet whole at entry; a source a packet holds is not read again")
		packetReads := planBuildPacketReads(t, root, cycle)
		if len(packetReads.reads) > planBuildPacketReadLimit {
			t.Errorf("plan/build with packets declares %d file reads, over %d: %v", len(packetReads.reads), planBuildPacketReadLimit, packetReads.reads)
		}
		t.Logf("file reads: %d (without packets: %d)", len(packetReads.reads), len(cycle))
		t.Logf("bytes without repetitions: %d B", packetReads.unique)
		t.Logf("leaves the packets bring that the route did not read: %d B %v", packetReads.newBytes, packetReads.newLeaves)
		t.Logf("bytes with repetitions (each packet counts its members again): %d B", packetReads.repeated)
	})
	t.Log("Explicit source unions, not transitive loading or observed tokens. Excludes native metadata/rendering, repeated reads, project inputs, history and host policies. Executable resources are measured separately, not presumed read by a model. No aggregate quota; each entrypoint/reference retains its 16000-byte check.")
}

// planBuildPacketReadLimit is the most file reads the plan/build route may
// declare once its flows read stage packets (D6 A).
const planBuildPacketReadLimit = 26

// planBuildRouteFlows are the flows the plan/build route enters; each reads its
// stage packets at entry.
var planBuildRouteFlows = []string{"hive-mode", "flow-plan", "flow-build", "flow-verify", "flow-close"}

type packetRoute struct {
	reads            []string
	unique, repeated int
	newBytes         int
	newLeaves        []string
}

// planBuildPacketReads replaces each source of the route that a packet of the
// route's flows holds by that packet, read once. The packet's size is the sum
// of its members plus one heading per member and its end line, before the
// manager rewrites links. The comparison with the baseline stays on the route's
// own sources: the leaves a packet adds and the repeated members are reported
// apart and do not count against it.
func planBuildPacketReads(t *testing.T, root string, route []string) packetRoute {
	t.Helper()
	skills := "content/skills/"
	size := func(path string) int {
		data, err := os.ReadFile(filepath.Join(root, path))
		if err != nil {
			t.Fatal(err)
		}
		return len(data)
	}
	held := map[string]bool{}
	var result packetRoute
	for _, flow := range planBuildRouteFlows {
		manifest := readStageManifest(t, flow)
		keys := make([]string, 0, len(manifest))
		for key := range manifest {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			members := manifest[key]
			packet := len("<!-- hive packet end: " + key + " ·  -->\n")
			for _, member := range members {
				held[skills+member] = true
				packet += size(skills+member) + len("## "+strings.TrimSuffix(member, ".md")+"\n\n") + len(member)
			}
			result.reads = append(result.reads, skills+flow+"/"+key+".md")
			result.repeated += packet
		}
	}
	inRoute := map[string]bool{}
	for _, path := range route {
		inRoute[path] = true
		if !held[path] {
			result.reads = append(result.reads, path)
			result.repeated += size(path)
		}
	}
	for path := range held {
		result.unique += size(path)
		if !inRoute[path] {
			result.newBytes += size(path)
			result.newLeaves = append(result.newLeaves, strings.TrimSuffix(filepath.Base(path), ".md"))
		}
	}
	for _, path := range route {
		if !held[path] {
			result.unique += size(path)
		}
	}
	sort.Strings(result.newLeaves)
	return result
}
