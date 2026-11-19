package content

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// leafLink names a file that must link a leaf at the step where the leaf's
// rule applies, so the rule is loaded at that moment. Active is false until the
// task that creates the leaf writes the link.
type leafLink struct {
	leaf   string
	from   string
	active bool
}

// leafLinks is the "linked from" column of the change skill-leaves design.
var leafLinks = []leafLink{
	{"hive-rule-git-commit", "content/skills/flow-build/SKILL.md", true},
	{"hive-rule-git-commit", "content/skills/git-workflow/SKILL.md", true},
	{"hive-rule-ticket-state", "content/skills/flow-build/SKILL.md", true},
	{"hive-rule-ticket-state", "content/skills/flow-close/SKILL.md", true},
	{"hive-rule-ticket-state", "content/skills/flow-deploy/SKILL.md", true},
	{"hive-rule-close-modes", "content/skills/flow-close/SKILL.md", true},
	{"hive-rule-close-hygiene", "content/skills/flow-close/SKILL.md", true},
	{"hive-rule-close-report", "content/skills/flow-close/SKILL.md", true},
	{"hive-rule-close-footer", "content/skills/flow-close/SKILL.md", true},
	{"hive-rule-close-report", "content/skills/git-workflow/SKILL.md", true},
	{"hive-rule-close-report", "content/skills/unattended-delegation/SKILL.md", true},
	{"hive-rule-merge-gate", "content/skills/git-workflow/SKILL.md", true},
	{"hive-rule-merge-gate", "content/skills/flow-verify/SKILL.md", true},
	{"hive-rule-merge-gate", "content/skills/flow-deploy/SKILL.md", true},
	{"hive-rule-pr-reviews", "content/skills/git-workflow/SKILL.md", true},
	{"hive-rule-promotion-path", "content/skills/git-workflow/SKILL.md", true},
	{"hive-rule-promotion-path", "content/skills/flow-deploy/SKILL.md", true},
	// Links of T7, from the flows that own the step where each leaf applies.
	{"hive-rule-reuse-before-writing", "content/skills/flow-build/SKILL.md", true},
	{"hive-rule-reuse-before-writing", "content/skills/flow-plan/SKILL.md", true},
	{"hive-rule-tests", "content/skills/flow-build/SKILL.md", true},
	{"hive-rule-tests", "content/skills/flow-plan/SKILL.md", true},
	{"hive-rule-tests-no-runner", "content/skills/flow-plan/SKILL.md", true},
	{"hive-rule-defect-fix", "content/skills/flow-build/SKILL.md", true},
	{"hive-rule-defect-fix", "content/skills/flow-verify/SKILL.md", true},
	{"hive-rule-delegation", "content/skills/flow-build/SKILL.md", true},
	{"hive-rule-delegation", "content/skills/flow-deploy/SKILL.md", true},
	{"hive-rule-reviewer-brief", "content/skills/flow-plan/SKILL.md", true},
	{"hive-rule-reviewer-brief", "content/skills/flow-build/SKILL.md", true},
	{"hive-rule-reviewer-brief", "content/skills/flow-verify/SKILL.md", true},
	{"hive-rule-business-rules", "content/skills/flow-plan/SKILL.md", true},
	{"hive-rule-business-rules", "content/skills/flow-build/SKILL.md", true},
	{"hive-rule-business-rules", "content/skills/flow-verify/SKILL.md", true},
	{"hive-rule-business-rules", "content/skills/flow-close/SKILL.md", true},
	{"hive-rule-business-rules", "content/skills/flow-figma-design/SKILL.md", true},
	{"hive-rule-business-ticket", "content/skills/flow-build/SKILL.md", true},
	{"hive-rule-business-ticket", "content/skills/flow-figma-design/SKILL.md", true},
	{"hive-rule-external-facts", "content/skills/flow-plan/SKILL.md", true},
	{"hive-rule-external-effects", "content/skills/flow-plan/SKILL.md", true},
	{"hive-rule-external-effects", "content/skills/flow-build/SKILL.md", true},
	{"hive-rule-ui-plan", "content/skills/flow-plan/SKILL.md", true},
	{"hive-rule-ui-verification", "content/skills/flow-build/SKILL.md", true},
	{"hive-rule-ui-verification", "content/skills/flow-verify/SKILL.md", true},
	{"hive-rule-before-captures", "content/skills/flow-build/SKILL.md", true},
	{"hive-rule-walkthrough", "content/skills/flow-build/SKILL.md", true},
	{"hive-rule-walkthrough", "content/skills/flow-verify/SKILL.md", true},
	{"hive-rule-viewer-open", "content/skills/flow-build/SKILL.md", true},
	{"hive-rule-viewer-open", "content/skills/flow-verify/SKILL.md", true},
	{"hive-rule-viewer-close", "content/skills/flow-close/SKILL.md", true},
	{"hive-rule-change-folder", "content/skills/flow-plan/SKILL.md", true},
	{"hive-rule-checkpoint", "content/skills/flow-build/SKILL.md", true},
	{"hive-rule-checkpoint", "content/skills/flow-verify/SKILL.md", true},
	{"hive-rule-checks", "content/skills/flow-build/SKILL.md", true},
	{"hive-rule-checks", "content/skills/flow-verify/SKILL.md", true},
	{"hive-rule-review-findings", "content/skills/flow-verify/SKILL.md", true},
	{"hive-rule-project-verification", "content/skills/flow-plan/SKILL.md", true},
	{"hive-rule-project-verification", "content/skills/flow-build/SKILL.md", true},
	{"hive-rule-project-verification", "content/skills/flow-verify/SKILL.md", true},
	{"hive-rule-stored-data", "content/skills/flow-build/SKILL.md", true},
	{"hive-rule-stored-data", "content/skills/flow-deploy/SKILL.md", true},
	{"hive-rule-security-boundaries", "content/skills/flow-build/SKILL.md", true},
	{"hive-rule-suppression-review", "content/skills/flow-verify/SKILL.md", true},
	// Links of T8: deploy, figma and unattended reach the shared rules through leaves.
	{"hive-rule-scope-changes", "content/skills/flow-deploy/SKILL.md", true},
	{"hive-rule-project-settings", "content/skills/flow-deploy/SKILL.md", true},
	{"hive-rule-test-accounts", "content/skills/flow-deploy/references/deployment-checks.md", true},
	{"hive-rule-design-stage", "content/skills/flow-figma-design/SKILL.md", true},
	// flow-research links the design-stage leaf when T13 removes its copy of the offer rule.
	{"hive-rule-design-stage", "content/skills/flow-research/SKILL.md", false},
	{"hive-rule-unattended-effects", "content/skills/unattended-delegation/SKILL.md", true},
	{"hive-rule-defect-fix", "content/agents/review/hive-review-code.md", true},
	{"hive-rule-review-evidence", "content/agents/review/hive-review-code.md", true},
	{"hive-rule-suppression-review", "content/agents/review/hive-review-code.md", true},
	{"hive-rule-security-boundaries", "content/agents/review/hive-review-security.md", true},
	{"hive-rule-tests", "content/agents/quality/hive-write-tests.md", true},
	{"hive-rule-delegation", "content/guidance/global.md", true},
	{"hive-rule-viewer", "content/skills/hive-rule-walkthrough/SKILL.md", true},
	{"hive-rule-ui-example", "content/skills/hive-rule-ui-plan/SKILL.md", true},
	{"hive-rule-child-brief", "content/skills/hive-rule-delegation/SKILL.md", true},
}

func linkProblems(root string, links []leafLink) []string {
	var problems []string
	for _, link := range links {
		if !link.active {
			continue
		}
		data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(link.from)))
		if err != nil {
			problems = append(problems, fmt.Sprintf("%s: %v", link.from, err))
			continue
		}
		if !strings.Contains(string(data), link.leaf+"/SKILL.md") {
			problems = append(problems, fmt.Sprintf("%s does not link %s", link.from, link.leaf))
		}
	}
	return problems
}

func TestLeavesAreLinkedWhereTheyApply(t *testing.T) {
	for _, problem := range linkProblems("../..", leafLinks) {
		t.Error(problem)
	}
}

func TestLeavesAreLinkedCatchesMissingLink(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "content", "skills", "flow-build", "SKILL.md"), "Commit, reading [c](../hive-rule-c/SKILL.md).\n")
	got := linkProblems(root, []leafLink{
		{"hive-rule-c", "content/skills/flow-build/SKILL.md", true},
		{"hive-rule-d", "content/skills/flow-build/SKILL.md", true},
		{"hive-rule-e", "content/skills/flow-build/SKILL.md", false},
	})
	if len(got) != 1 || !strings.Contains(got[0], "does not link hive-rule-d") {
		t.Errorf("unexpected problems: %v", got)
	}
}
