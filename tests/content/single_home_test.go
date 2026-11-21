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
// and skips the rest.
type homePhrase struct {
	phrase string
	home   string
}

// homePhrases lists one literal phrase per rule leaf and the leaf that must be
// its only home under content/.
var homePhrases = []homePhrase{
	{"Never run `git add -A`", "content/skills/hive-core/references/git-commit.md"},
	{"five at most, the rest approved in one question", "content/skills/hive-core/references/questions.md"},
	{"Before the validation, or before the push when the mode", "content/skills/hive-core/references/foreign-commits.md"},
	{"is done when its changes are integrated", "content/skills/hive-core/references/ticket-state.md"},
	{"adds, replaces or cancels a business rule, record it in", "content/skills/hive-core/references/business-rules.md"},
	{"Name the cause before you edit", "content/skills/hive-core/references/defect-fix.md"},
	{"a state the code must never hold cannot be written down", "content/skills/hive-principle-type-system-discipline/SKILL.md"},
	{"Inside the boundary, trust the types", "content/skills/hive-principle-boundary-discipline/SKILL.md"},
	{"Put the domain rules in a data structure", "content/skills/hive-principle-model-the-domain/SKILL.md"},
	{"Derive each step from the state that exists now", "content/skills/hive-principle-make-operations-idempotent/SKILL.md"},
	{"move them to the new one in the same change", "content/skills/hive-principle-migrate-callers-then-delete-legacy-apis/SKILL.md"},
	{"Reader load has two axes", "content/skills/hive-principle-minimize-reader-load/SKILL.md"},
	{"whether they need one shared target", "content/skills/hive-principle-separate-before-serializing-shared-state/SKILL.md"},
	{"Write the decision before the run", "content/skills/hive-principle-explain-the-number/SKILL.md"},
	{"propose the cheap always-on signal", "content/skills/hive-core/references/defect-fix.md"},
	{"write the premise as one sentence", "content/skills/hive-core/references/defect-fix.md"},
	{"Keep a test that fails when the behavior its users observe breaks", "content/skills/hive-core/references/tests.md"},
	{"Remove dead code, redundant validators and stub references", "content/skills/hive-core/references/reuse-before-writing.md"},
	{"with no layer, abstraction or parameter the request does not need", "content/skills/hive-core/references/reuse-before-writing.md"},
	{"yourself that the host has a connection that can write", "content/skills/hive-core/references/design-stage.md"},
	{"Never bypass branch protection, required reviews or required checks", "content/skills/hive-core/references/merge-gate.md"},
	{"A human reviewer holds the merge in every delivery mode", "content/skills/hive-core/references/merge-gate.md"},
	{"Treat an assigned or requested reviewer's comments as review findings", "content/skills/hive-core/references/pr-reviews.md"},
	{"receive changes only through the declared promotion path", "content/skills/hive-core/references/promotion-path.md"},
	{"Close a work item once", "content/skills/hive-core/references/close-modes.md"},
	{"A finding is a defect the work found but did not fix", "content/skills/hive-core/references/close-report.md"},
	{"A candidate is a repetition the work did not turn into a mechanism", "content/skills/hive-core/references/close-report.md"},
	{"the delivery mode in force, or the flow that called the close, picks the row", "content/skills/hive-core/references/close-report.md"},
	{"Take the first rung that fully does what the task needs", "content/skills/hive-core/references/reuse-before-writing.md"},
	{"Name the realistic change to the code that would make it fail", "content/skills/hive-core/references/tests.md"},
	{"pin its current behavior for those callers with a check through the real path", "content/skills/hive-core/references/tests-no-runner.md"},
	{"A child does the work for the first of these reasons that applies", "content/skills/hive-core/references/delegation.md"},
	{"Wait for children with the host's wait tool", "content/skills/hive-core/references/child-brief.md"},
	{"Give a reviewer what it needs to judge the work and nothing that argues for it", "content/skills/hive-core/references/reviewer-brief.md"},
	{"Never delete a rule's line or change the rule it states", "content/skills/hive-core/references/business-record.md"},
	{"Leave every other character of the description as it was", "content/skills/hive-core/references/business-ticket.md"},
	{"Measure each one with a read-only read", "content/skills/hive-core/references/external-facts.md"},
	{"run each command it names once on the base", "content/skills/flow-plan/SKILL.md"},
	{"A class the user approved is the user's instruction for each effect of that class", "content/skills/hive-core/references/external-effects.md"},
	{"Fix the data contract each changed screen consumes", "content/skills/hive-core/references/ui-plan.md"},
	{"has an approved example; one still pending approval settles nothing", "content/skills/hive-core/references/ui-example.md"},
	{"has no timeout for `open`, `eval`, `snapshot`, `screenshot` or `close`", "content/skills/hive-core/references/ui-verification.md"},
	{"Save them in the work's `workspace/YYYY-MM-DD-<work>/images/before/`", "content/skills/hive-core/references/before-captures.md"},
	{"axe-core finds on average 57% of WCAG issues", "content/skills/hive-core/references/a11y-audit.md"},
	{"The calling flow owns its existing acceptance question", "content/skills/hive-core/references/walkthrough.md"},
	{"Never start or restart the viewer service", "content/skills/hive-core/references/viewer.md"},
	{"Never open the viewer on your own, and never ask about it on its own", "content/skills/hive-core/references/viewer-open.md"},
	{"never close an already closed work again", "content/skills/hive-core/references/viewer-close.md"},
	{"Earlier change folders in the repository are context, not a model", "content/skills/hive-core/references/change-folder.md"},
	{"Tasks are never declared in the viewer", "content/skills/hive-core/references/checkpoint.md"},
	{"never the full suite after each fix", "content/skills/hive-core/references/checks.md"},
	{"Non-blocking findings go to the close's findings, never to another round", "content/skills/hive-core/references/review-findings.md"},
	{"A reviewer's assertion alone does not confirm a finding", "content/skills/hive-core/references/review-evidence.md"},
	{"Do not create a second active source", "content/skills/hive-core/references/project-verification.md"},
	{"Duplicate keys, IDs, array entries, unknown fields and unknown versions are errors", "content/skills/hive-core/references/project-verification-format.md"},
	{"take a read-only count of the affected records in each environment the change reaches", "content/skills/hive-core/references/stored-data.md"},
	{"Keep data apart from executable syntax", "content/skills/hive-core/references/security-boundaries.md"},
	{"A green check alone does not establish that a suppressed failure is fixed", "content/skills/hive-core/references/suppression-review.md"},
	{"is allowed when it is reversible or made reversible first and disrupts no other component", "content/skills/hive-core/references/unattended-effects.md"},
	{"Draw two or three examples only for a screen the user names", "content/skills/hive-core/references/design-stage.md"},
	{"Each row below ends a message differently", "content/skills/hive-core/references/report-endings.md"},
	{"named it with its target, not whether or how to make one, and research finds it a small understood change", "content/skills/hive-core/references/report-endings.md"},
	{"Delete every Kill", "content/skills/hive-core/references/no-comments.md"},
	{"Name the one fact the change is safe because of", "content/skills/hive-core/references/blast-radius.md"},
	{"Write so a tired engineer understands on the first read", "content/skills/hive-core/references/prose.md"},
	{"Break this change. Prove each finding with a run against the real code", "content/skills/hive-core/references/interrogate.md"},
	{"single-line row per decision and one per external effect", "content/skills/unattended-delegation/SKILL.md"},
	{"A plan delivered in two or more units", "content/skills/hive-core/references/change-folder.md"},
	{"weaken, skip or remove a guard to reach done", "content/skills/hive-core/references/loops.md"},
	{"A `timeout` on one command inside the loop bounds one iteration, not the wait", "content/skills/hive-core/references/background-wait.md"},
	{"never from an unmeasured guess", "content/skills/hive-core/references/checks.md"},
	{"then its REST API with the user's global token", "content/skills/hive-core/references/trackers.md"},
	{"belongs to that work: fix it without asking or opening a ticket", "content/skills/hive-core/references/reasons-and-blockers.md"},
	{"though it changes a description the task did not create", "content/skills/hive-core/references/ticket-state.md"},
	{"use the main checkout's file", "content/skills/hive-core/references/local-write/worktree.md"},
	{"gets no second report", "content/skills/hive-core/references/checkpoint.md"},
	{"Recalled memory and prior summaries are claims about past state", "content/skills/hive-principle-verify-current-state/SKILL.md"},
	{"read its current documentation through Context7 when it is installed", "content/skills/hive-principle-verify-current-state/SKILL.md"},
	{"Ask only a choice that is the user's and changes what happens next", "content/skills/hive-core/references/questions.md"},
	{"Never give as the reason that an instruction you follow requires", "content/skills/hive-core/references/reasons-and-blockers.md"},
	{"Asking only in the text where the card applies is a failure", "content/skills/hive-core/references/questions.md"},
	{"Never write the round out on your own; act on what they said", "content/skills/hive-core/references/questions/unanswered.md"},
	{"is never asked with the question tool, because a card hides the report it closes", "content/skills/hive-core/references/report-endings.md"},
	{"Ask the delivery mode and the code review mechanism once per session and repository, in one round", "content/skills/hive-core/references/delivery-ask.md"},
	{"name them with their count and the work they belong to, and ask whether they go in the same publication", "content/skills/hive-core/references/foreign-commits.md"},
	{"Only an explicitly assigned `hive-build-owner` may build, test, fix and create local commits", "content/skills/flow-build/references/e2e-coordination.md"},
	{"allows creating or changing it only through this mode", "content/skills/harness-audit/references/engram-workspace.md"},
}

// homeProblems returns one line per phrase whose files under
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
	writeFile(t, filepath.Join(root, "content", "skills", "hive-core", "references", "x.md"), "Keep one\nhome for a rule.\n")
	writeFile(t, filepath.Join(root, "content", "skills", "flow-build", "SKILL.md"), "Also: keep one home for a rule.\n")
	got := homeProblems(root, []homePhrase{
		{"keep one home for a rule", "content/skills/hive-core/references/x.md"},
		{"a rule nobody wrote", "content/skills/hive-core/references/x.md"},
	})
	if len(got) != 2 || !strings.Contains(got[0], "found in [content/skills/flow-build/SKILL.md]") || !strings.Contains(got[1], "found in []") {
		t.Errorf("unexpected problems: %v", got)
	}
}
