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

// homePhrases lists one literal phrase per rule and the file that must be
// its only home under content/.
var homePhrases = []homePhrase{
	{"A small understood change is one whose how leaves no decision", "content/skills/hive-core/plan/proportionality.md"},
	{"updating the pending-work records in persistent memory that name it", "content/skills/hive-core/close/close-cleanup.md"},
	{"A mechanical change is a rename, a move", "content/skills/hive-core/plan/proportionality.md"},
	{"An operational task is work that leaves no repository change to deliver", "content/skills/hive-core/plan/proportionality.md"},
	{"recommending the new session so that", "content/skills/hive-core/plan/scope-changes.md"},
	{"Never run `git add -A`", "content/skills/hive-core/git/git-commit.md"},
	{"five at most, the rest approved in one question", "content/skills/hive-core/messages/questions.md"},
	{"Before the validation, or before the push when the mode", "content/skills/hive-core/git/foreign-commits.md"},
	{"is done when its changes are integrated", "content/skills/hive-core/tracker/ticket-state.md"},
	{"adds, replaces or cancels a business rule, record it in", "content/skills/hive-core/business/business-rules.md"},
	{"Name the cause before you edit", "content/skills/hive-core/build/defect-fix.md"},
	{"a state the code must never hold cannot be written down", "content/skills/hive-core/principles/type-system-discipline.md"},
	{"Inside the boundary, trust the types", "content/skills/hive-core/principles/boundary-discipline.md"},
	{"Put the domain rules in a data structure", "content/skills/hive-core/principles/model-the-domain.md"},
	{"Derive each step from the state that exists now", "content/skills/hive-core/principles/make-operations-idempotent.md"},
	{"move them to the new one in the same change", "content/skills/hive-core/principles/migrate-callers-then-delete-legacy-apis.md"},
	{"Reader load has two axes", "content/skills/hive-core/principles/minimize-reader-load.md"},
	{"whether they need one shared target", "content/skills/hive-core/principles/separate-before-serializing-shared-state.md"},
	{"Write the decision before the run", "content/skills/hive-core/principles/explain-the-number.md"},
	{"propose the cheap always-on signal", "content/skills/hive-core/build/defect-fix.md"},
	{"write the premise as one sentence", "content/skills/hive-core/build/defect-fix.md"},
	{"Keep a test that fails when the behavior its users observe breaks", "content/skills/hive-core/build/tests.md"},
	{"Remove dead code, redundant validators and stub references", "content/skills/hive-core/plan/reuse-before-writing.md"},
	{"with no layer, abstraction or parameter the request does not need", "content/skills/hive-core/plan/reuse-before-writing.md"},
	{"It is not when (a) it is a workaround", "content/skills/hive-core/plan/reuse-before-writing.md"},
	{"yourself that the host has a connection that can write", "content/skills/hive-core/ui/design-stage.md"},
	{"Never bypass branch protection, required reviews or required checks", "content/skills/hive-core/git/merge-gate.md"},
	{"A human reviewer holds the merge in every delivery mode", "content/skills/hive-core/git/merge-gate.md"},
	{"Treat an assigned or requested reviewer's comments as review findings", "content/skills/hive-core/git/pr-reviews.md"},
	{"receive changes only through the declared promotion path", "content/skills/hive-core/git/promotion-path.md"},
	{"Close a work item once", "content/skills/hive-core/close/close-modes.md"},
	{"A finding is a defect the work found but did not fix", "content/skills/hive-core/close/close-report.md"},
	{"A candidate is a repetition the work did not turn into a mechanism", "content/skills/hive-core/close/close-report.md"},
	{"the delivery mode in force, or the flow that called the close, picks the row", "content/skills/hive-core/close/close-report.md"},
	{"Take the first rung that fully does what the task needs", "content/skills/hive-core/plan/reuse-before-writing.md"},
	{"Name the realistic change to the code that would make it fail", "content/skills/hive-core/build/tests.md"},
	{"pin its current behavior for those callers with a check through the real path", "content/skills/hive-core/build/tests-no-runner.md"},
	{"A child does the work for the first of these reasons that applies", "content/skills/hive-core/delegation/delegation.md"},
	{"Wait for children with the host's wait tool", "content/skills/hive-core/delegation/child-brief.md"},
	{"Give a reviewer what it needs to judge the work and nothing that argues for it", "content/skills/hive-core/delegation/reviewer-brief.md"},
	{"Never delete a rule's line or change the rule it states", "content/skills/hive-core/business/business-record.md"},
	{"Leave every other character of the description as it was", "content/skills/hive-core/business/business-ticket.md"},
	{"Measure each one with a read-only read", "content/skills/hive-core/plan/external-facts.md"},
	{"run each command it names once on the base", "content/skills/flow-plan/SKILL.md"},
	{"A class the user approved is the user's instruction for each effect of that class", "content/skills/hive-core/plan/external-effects.md"},
	{"Fix the data contract each changed screen consumes", "content/skills/hive-core/ui/ui-plan.md"},
	{"has an approved example; one still pending approval settles nothing", "content/skills/hive-core/ui/ui-example.md"},
	{"has no timeout for `open`, `eval`, `snapshot`, `screenshot` or `close`", "content/skills/hive-core/ui/ui-verification.md"},
	{"Save them in the work's `workspace/YYYY-MM-DD-<work>/images/before/`", "content/skills/hive-core/ui/before-captures.md"},
	{"axe-core finds on average 57% of WCAG issues", "content/skills/hive-core/ui/a11y-audit.md"},
	{"The calling flow owns its existing acceptance question", "content/skills/hive-core/ui/walkthrough.md"},
	{"Never start or restart the viewer service", "content/skills/hive-core/ui/viewer.md"},
	{"Never open the viewer on your own, and never ask about it on its own", "content/skills/hive-core/ui/viewer-open.md"},
	{"do not close an already closed work again", "content/skills/hive-core/ui/viewer-close.md"},
	{"Tasks are never declared in the viewer", "content/skills/hive-core/plan/checkpoint.md"},
	{"never the full suite after each fix", "content/skills/hive-core/build/checks.md"},
	{"Its non-blocking findings are sorted as above with no fix of their own", "content/skills/hive-core/review/review-findings.md"},
	{"Each takes the first of these decisions that matches it", "content/skills/hive-core/review/review-findings.md"},
	{"A reviewer's assertion alone does not confirm a finding", "content/skills/hive-core/review/review-evidence.md"},
	{"Do not create a second active source", "content/skills/hive-core/build/project-verification.md"},
	{"Duplicate keys, IDs, array entries, unknown fields and unknown versions are errors", "content/skills/hive-core/build/project-verification-format.md"},
	{"take a read-only count of the affected records in each environment the change reaches", "content/skills/hive-core/build/stored-data.md"},
	{"Keep data apart from executable syntax", "content/skills/hive-core/build/security-boundaries.md"},
	{"A green check alone does not establish that a suppressed failure is fixed", "content/skills/hive-core/review/suppression-review.md"},
	{"is allowed when it is reversible or made reversible first and disrupts no other component", "content/skills/hive-core/delegation/unattended-effects.md"},
	{"Draw two or three examples only for a screen the user names", "content/skills/hive-core/ui/design-stage.md"},
	{"Each row below ends a message differently", "content/skills/hive-core/messages/report-endings.md"},
	{"named it with its target, not whether or how to make one, and research finds it a small understood change", "content/skills/hive-core/messages/report-endings.md"},
	{"Delete every Kill", "content/skills/hive-core/review/no-comments.md"},
	{"restates a rule the code", "content/skills/hive-core/review/no-comments.md"},
	{"Name the one fact the change is safe because of", "content/skills/hive-core/review/blast-radius.md"},
	{"Write so a tired engineer understands on the first read", "content/skills/hive-core/build/prose.md"},
	{"Break this change. Prove each finding with a run against the real code", "content/skills/hive-core/review/interrogate.md"},
	{"single-line row per decision and one per external effect", "content/skills/unattended-delegation/SKILL.md"},
	{"Earlier plans in the repository are context, not a model", "content/skills/hive-core/plan/plan-record.md"},
	{"A plan delivered in two or more units", "content/skills/hive-core/plan/plan-record.md"},
	{"weaken, skip or remove a guard to reach done", "content/skills/hive-core/plan/loops.md"},
	{"A `timeout` on one command inside the loop bounds one iteration, not the wait", "content/skills/hive-core/build/background-wait.md"},
	{"never from an unmeasured guess", "content/skills/hive-core/build/checks.md"},
	{"then its REST API with the user's global token", "content/skills/hive-core/tracker/trackers.md"},
	{"belongs to that work: fix it without asking or opening a ticket", "content/skills/hive-core/messages/reasons-and-blockers.md"},
	{"even as it edits a description the task did not create", "content/guidance/global.md"},
	{"use the main checkout's file", "content/skills/hive-core/local/local-write/worktree.md"},
	{"gets no second report", "content/skills/hive-core/plan/checkpoint.md"},
	{"Recalled memory and prior summaries are claims about past state", "content/skills/hive-core/principles/verify-current-state.md"},
	{"read its current documentation through Context7 when it is installed", "content/skills/hive-core/principles/verify-current-state.md"},
	{"Ask only a choice that is the user's and changes what happens next", "content/skills/hive-core/messages/questions.md"},
	{"Never give as the reason that an instruction you follow requires", "content/skills/hive-core/messages/reasons-and-blockers.md"},
	{"is a failure on every host and in every mode", "content/skills/hive-core/messages/questions.md"},
	{"The round appears once per turn, in the card or in the text", "content/skills/hive-core/messages/questions.md"},
	{"Never write the round out on your own; act on what they said", "content/skills/hive-core/messages/questions/unanswered.md"},
	{"is never asked with the question tool, because a card hides the report it closes", "content/skills/hive-core/messages/report-endings.md"},
	{"Ask the delivery mode and the code review mechanism once per session and repository, in one round", "content/skills/hive-core/git/delivery-ask.md"},
	{"name them with their count and the work they belong to, and ask whether they go in the same publication", "content/skills/hive-core/git/foreign-commits.md"},
	{"Only an explicitly assigned `hive-build-owner` may build, test, fix and create local commits", "content/skills/flow-build/references/e2e-coordination.md"},
	{"Only in the swarm mode, write the plan file and checkpoint of your assigned tickets", "content/agents/development/hive-ticket-owner.md"},
	{"allows creating or changing it only through this mode", "content/skills/audit-harness/references/engram-workspace.md"},
	{"Before the work a line names, read that skill's SKILL.md whole.", "content/skills/audit-harness/references/instruction-files.md"},
	{"copy into the brief of every child, builder or reviewer, each of its lines whose trigger applies", "content/skills/hive-core/delegation/child-brief.md"},
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
