package content

import (
	"fmt"
	"os"
	"os/exec"
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
	common := []string{global, skills + "flow-close/references/reporting.md"}
	question := append(append([]string{}, common...), skills+"hive-core/references/questions.md")
	cycle := append(append([]string{}, question...),
		skills+"flow-plan/SKILL.md", skills+"flow-build/SKILL.md", skills+"flow-close/SKILL.md",
		skills+"flow-research/SKILL.md", // planning's exploration/delegation criterion
		skills+"flow-verify/SKILL.md", skills+"flow-verify/references/review-and-gates.md",
		skills+"hive-core/references/project-settings.md",
		skills+"hive-core/references/language.md",
		skills+"hive-core/references/memory.md",
		skills+"flow-plan/references/change-records.md",
		skills+"flow-plan/references/plan-format.md",
		skills+"flow-build/references/plan-execution.md",
		skills+"flow-build/references/tdd.md",
		skills+"flow-plan/references/plan-review.md",
		skills+"flow-plan/references/delivery-decisions.md",
		skills+"flow-build/references/verification.md",
		// Split out of the three entries above (issue #204).
		skills+"flow-plan/references/review-decisions.md",
		skills+"flow-plan/references/plan-criteria-and-markers.md",
		skills+"flow-plan/references/plan-handoff.md",
		skills+"flow-build/references/foreseen-authorizations.md",
		// The build route of plan-handoff gives the plan's size (issue #259).
		skills+"hive-core/references/option-scope.md",
		// Split out of reporting; read between build tasks (issue #265).
		skills+"flow-close/references/work-item-boundary.md",
		// Split out of flow-build at its size limit (follow-up of #269).
		skills+"flow-build/references/iterate-and-integrate.md",
		skills+"flow-research/references/delegation.md",
		skills+"workspace-conventions/references/artifact-layout.md",
		"content/agents/review/hive-review-plan.md",
		"content/agents/development/hive-build-backend.md",
		"content/agents/quality/hive-verify-task.md",
		// Several readers require these resources; union excludes repeats.
		skills+"flow-close/references/reporting.md", skills+"flow-plan/references/plan-review.md",
	)
	moved := map[string]bool{
		skills + "flow-close/references/reporting.md":                  true,
		skills + "hive-core/references/questions.md":                   true,
		skills + "hive-core/references/project-settings.md":            true,
		skills + "hive-core/references/language.md":                    true,
		skills + "hive-core/references/memory.md":                      true,
		skills + "flow-build/references/plan-execution.md":             true,
		skills + "flow-build/references/tdd.md":                        true,
		skills + "flow-verify/SKILL.md":                                true,
		skills + "flow-verify/references/review-and-gates.md":          true,
		skills + "flow-research/references/delegation.md":              true,
		skills + "workspace-conventions/references/artifact-layout.md": true,
		skills + "flow-plan/references/review-decisions.md":            true,
		skills + "flow-plan/references/plan-criteria-and-markers.md":   true,
		skills + "flow-plan/references/plan-handoff.md":                true,
		skills + "flow-build/references/foreseen-authorizations.md":    true,
		skills + "hive-core/references/option-scope.md":                true,
		skills + "flow-close/references/work-item-boundary.md":         true,
		skills + "flow-build/references/iterate-and-integrate.md":      true,
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
				return exec.Command("git", "show", "d5164f71:"+path).Output()
			})
			after := measure("B working tree", scenario.paths, func(path string) ([]byte, error) {
				return os.ReadFile("../../" + path)
			})
			t.Logf("A union: %d bytes, %d characters; B/C declared source union: %d bytes, %d characters; byte change: %+d (%.2f%%)", before.bytes, before.characters, after.bytes, after.characters, after.bytes-before.bytes, 100*float64(after.bytes-before.bytes)/float64(before.bytes))
		})
	}
	t.Log("C uses B's canonical sources; mapped plan-review injection adds no unique source to this cycle. Raw source frontmatter is counted. This is not per-request context or actual tokens. Native rendering and skill/role discovery metadata, plugin metadata/wrappers, repeated text across readers/requests, history, project inputs and host policy are not measured. Pilots must observe delivery and token usage separately.")
}
