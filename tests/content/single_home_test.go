package content

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// homePhrase is a literal phrase of one rule and the one file that may state
// it. A rule stated in two places drifts: an agent follows the shorter copy
// and skips the rest. Active is false until the task that moves the rule to its
// leaf removes the other copies.
type homePhrase struct {
	phrase string
	home   string
	active bool
}

// homePhrases is the single-home registry of the change skill-leaves. Phrases
// marked "repeated today" appear in two or more files on the base branch, so
// the test fails there once the phrase is active and its copies remain.
var homePhrases = []homePhrase{
	{"Never run `git add -A`", "content/skills/hive-rule-git-commit/SKILL.md", true},
	// Repeated today: global.md «Delivery» and git-workflow.
	{"Before the validation, or before the push when the mode", "content/skills/hive-rule-foreign-commits/SKILL.md", true},
	{"is done when its changes are integrated", "content/skills/hive-rule-ticket-state/SKILL.md", true},
	// Repeated today: flow-build and flow-verify.
	{"adds, replaces or cancels a business rule, record it in", "content/skills/hive-rule-business-rules/SKILL.md", true},
	{"Name the cause before you edit", "content/skills/hive-rule-defect-fix/SKILL.md", true},
	// Repeated on the base: the entry router (then another name) and flow-figma-design.
	{"yourself that the host has a connection that can write", "content/skills/hive-rule-design-stage/SKILL.md", true},
	{"Never bypass branch protection, required reviews or required checks", "content/skills/hive-rule-merge-gate/SKILL.md", true},
	{"A human reviewer holds the merge in every delivery mode", "content/skills/hive-rule-merge-gate/SKILL.md", true},
	{"Treat an assigned or requested reviewer's comments as review findings", "content/skills/hive-rule-pr-reviews/SKILL.md", true},
	{"receive changes only through the declared promotion path", "content/skills/hive-rule-promotion-path/SKILL.md", true},
	{"Close a work item once", "content/skills/hive-rule-close-modes/SKILL.md", true},
	{"A finding is a defect the work found but did not fix", "content/skills/hive-rule-close-report/SKILL.md", true},
	{"A candidate is a repetition the work did not turn into a mechanism", "content/skills/hive-rule-close-report/SKILL.md", true},
	{"Take the first rung that fully does what the task needs", "content/skills/hive-rule-reuse-before-writing/SKILL.md", true},
	{"Name the realistic change to the code that would make it fail", "content/skills/hive-rule-tests/SKILL.md", true},
	{"pin its current behavior for those callers with a check through the real path", "content/skills/hive-rule-tests-no-runner/SKILL.md", true},
	{"A child does the work for the first of these reasons that applies", "content/skills/hive-rule-delegation/SKILL.md", true},
	{"Wait for children with the host's wait tool", "content/skills/hive-rule-child-brief/SKILL.md", true},
	{"Give a reviewer what it needs to judge the work and nothing that argues for it", "content/skills/hive-rule-reviewer-brief/SKILL.md", true},
	{"Never delete a rule's line or change the rule it states", "content/skills/hive-rule-business-record/SKILL.md", true},
	{"Leave every other character of the description as it was", "content/skills/hive-rule-business-ticket/SKILL.md", true},
	{"Measure each one with a read-only read", "content/skills/hive-rule-external-facts/SKILL.md", true},
	{"A class the user approved is the user's instruction for each effect of that class", "content/skills/hive-rule-external-effects/SKILL.md", true},
	{"Fix the data contract each changed screen consumes", "content/skills/hive-rule-ui-plan/SKILL.md", true},
	{"has an approved example; one still pending approval settles nothing", "content/skills/hive-rule-ui-example/SKILL.md", true},
	{"has no timeout for `open`, `eval`, `snapshot`, `screenshot` or `close`", "content/skills/hive-rule-ui-verification/SKILL.md", true},
	{"Save them in the work's `workspace/YYYY-MM-DD-<work>/images/before/`", "content/skills/hive-rule-before-captures/SKILL.md", true},
	{"axe-core finds on average 57% of WCAG issues", "content/skills/hive-rule-a11y-audit/SKILL.md", true},
	{"The calling flow owns its existing acceptance question", "content/skills/hive-rule-walkthrough/SKILL.md", true},
	{"Never start or restart the viewer service", "content/skills/hive-rule-viewer/SKILL.md", true},
	{"Never open the viewer on your own, and never ask about it on its own", "content/skills/hive-rule-viewer-open/SKILL.md", true},
	{"never close an already closed work again", "content/skills/hive-rule-viewer-close/SKILL.md", true},
	{"Earlier change folders in the repository are context, not a model", "content/skills/hive-rule-change-folder/SKILL.md", true},
	{"Tasks are never declared in the viewer", "content/skills/hive-rule-checkpoint/SKILL.md", true},
	{"never the full suite after each fix", "content/skills/hive-rule-checks/SKILL.md", true},
	{"Non-blocking findings go to the close's findings, never to another round", "content/skills/hive-rule-review-findings/SKILL.md", true},
	{"A reviewer's assertion alone does not confirm a finding", "content/skills/hive-rule-review-evidence/SKILL.md", true},
	{"Do not create a second active source", "content/skills/hive-rule-project-verification/SKILL.md", true},
	{"Duplicate keys, IDs, array entries, unknown fields and unknown versions are errors", "content/skills/hive-rule-project-verification-format/SKILL.md", true},
	{"take a read-only count of the affected records in each environment the change reaches", "content/skills/hive-rule-stored-data/SKILL.md", true},
	{"Keep data apart from executable syntax", "content/skills/hive-rule-security-boundaries/SKILL.md", true},
	{"A green check alone does not establish that a suppressed failure is fixed", "content/skills/hive-rule-suppression-review/SKILL.md", true},
	{"is allowed when it is reversible or made reversible first and disrupts no other component", "content/skills/hive-rule-unattended-effects/SKILL.md", true},
	{"Draw two or three examples only for a screen the user names", "content/skills/hive-rule-design-stage/SKILL.md", true},
	{"Three reports end differently", "content/skills/hive-rule-report-endings/SKILL.md", true},
	{"Delete every Kill", "content/skills/hive-rule-no-comments/SKILL.md", true},
	{"Name the one fact the change is safe because of", "content/skills/hive-rule-blast-radius/SKILL.md", true},
	{"Write so a tired engineer understands on the first read", "content/skills/hive-rule-prose/SKILL.md", true},
	{"Break this change. Prove each finding with a run against the real code", "content/skills/hive-rule-interrogate/SKILL.md", true},
}

// homeProblems returns one line per active phrase whose files under
// root/content are not exactly its home.
func homeProblems(root string, phrases []homePhrase) []string {
	files := map[string]string{}
	contentDir := filepath.Join(root, "content")
	_ = filepath.WalkDir(contentDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(path, ".md") {
			return nil
		}
		data, readErr := os.ReadFile(path)
		if readErr == nil {
			rel, _ := filepath.Rel(root, path)
			files[filepath.ToSlash(rel)] = strings.Join(strings.Fields(string(data)), " ")
		}
		return nil
	})
	var problems []string
	for _, p := range phrases {
		if !p.active {
			continue
		}
		var found []string
		for rel, text := range files {
			if strings.Contains(text, p.phrase) {
				found = append(found, rel)
			}
		}
		sort.Strings(found)
		if len(found) != 1 || found[0] != p.home {
			problems = append(problems, fmt.Sprintf("%q must appear only in %s; found in %v", p.phrase, p.home, found))
		}
	}
	return problems
}

func TestRulesHaveOneHome(t *testing.T) {
	for _, problem := range homeProblems("../..", homePhrases) {
		t.Error(problem)
	}
}

func TestRulesHaveOneHomeCatchesCopies(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "content", "skills", "hive-rule-x", "SKILL.md"), "Keep one\nhome for a rule.\n")
	writeFile(t, filepath.Join(root, "content", "skills", "flow-build", "SKILL.md"), "Also: keep one home for a rule.\n")
	got := homeProblems(root, []homePhrase{
		{"keep one home for a rule", "content/skills/hive-rule-x/SKILL.md", true},
		{"a rule nobody wrote", "content/skills/hive-rule-x/SKILL.md", true},
		{"keep one home for a rule", "content/skills/flow-build/SKILL.md", false},
	})
	if len(got) != 2 || !strings.Contains(got[0], "found in [content/skills/flow-build/SKILL.md]") || !strings.Contains(got[1], "found in []") {
		t.Errorf("unexpected problems: %v", got)
	}
}
