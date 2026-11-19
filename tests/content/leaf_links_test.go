package content

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// leafLink names a file that must link a leaf at the step where the leaf's
// rule applies, so the rule is loaded at that moment.
type leafLink struct {
	leaf string
	from string
}

// leafLinks lists, per leaf, the files that link it at the step where its rule applies.
var leafLinks = []leafLink{
	{"hive-rule-git-commit", "content/skills/flow-build/SKILL.md"},
	{"hive-rule-git-commit", "content/skills/git-workflow/SKILL.md"},
	{"hive-rule-ticket-state", "content/skills/flow-build/SKILL.md"},
	{"hive-rule-ticket-state", "content/skills/flow-close/SKILL.md"},
	{"hive-rule-ticket-state", "content/skills/flow-deploy/SKILL.md"},
	{"hive-rule-close-modes", "content/skills/flow-close/SKILL.md"},
	{"hive-rule-close-hygiene", "content/skills/flow-close/SKILL.md"},
	{"hive-rule-close-report", "content/skills/flow-close/SKILL.md"},
	{"hive-rule-close-footer", "content/skills/flow-close/SKILL.md"},
	{"hive-rule-close-report", "content/skills/git-workflow/SKILL.md"},
	{"hive-rule-close-report", "content/skills/unattended-delegation/SKILL.md"},
	{"hive-rule-merge-gate", "content/skills/git-workflow/SKILL.md"},
	{"hive-rule-merge-gate", "content/skills/flow-verify/SKILL.md"},
	{"hive-rule-merge-gate", "content/skills/flow-deploy/SKILL.md"},
	{"hive-rule-pr-reviews", "content/skills/git-workflow/SKILL.md"},
	{"hive-rule-promotion-path", "content/skills/git-workflow/SKILL.md"},
	{"hive-rule-promotion-path", "content/skills/flow-deploy/SKILL.md"},
	{"hive-rule-reuse-before-writing", "content/skills/flow-build/SKILL.md"},
	{"hive-principle-index", "content/skills/flow-build/SKILL.md"},
	{"hive-typescript", "content/skills/hive-principle-index/SKILL.md"},
	{"hive-rule-reuse-before-writing", "content/skills/flow-plan/SKILL.md"},
	{"hive-rule-tests", "content/skills/flow-build/SKILL.md"},
	{"hive-rule-tests", "content/skills/flow-plan/SKILL.md"},
	{"hive-rule-tests-no-runner", "content/skills/flow-plan/SKILL.md"},
	{"hive-rule-defect-fix", "content/skills/flow-build/SKILL.md"},
	{"hive-rule-defect-fix", "content/skills/flow-verify/SKILL.md"},
	{"hive-rule-delegation", "content/skills/flow-build/SKILL.md"},
	{"hive-rule-delegation", "content/skills/flow-deploy/SKILL.md"},
	{"hive-rule-reviewer-brief", "content/skills/flow-plan/SKILL.md"},
	{"hive-rule-reviewer-brief", "content/skills/flow-build/SKILL.md"},
	{"hive-rule-reviewer-brief", "content/skills/flow-verify/SKILL.md"},
	{"hive-rule-business-rules", "content/skills/flow-plan/SKILL.md"},
	{"hive-rule-business-rules", "content/skills/flow-build/SKILL.md"},
	{"hive-rule-business-rules", "content/skills/flow-verify/SKILL.md"},
	{"hive-rule-business-rules", "content/skills/flow-close/SKILL.md"},
	{"hive-rule-business-rules", "content/skills/flow-figma-design/SKILL.md"},
	{"hive-rule-business-ticket", "content/skills/flow-build/SKILL.md"},
	{"hive-rule-business-ticket", "content/skills/flow-figma-design/SKILL.md"},
	{"hive-rule-external-facts", "content/skills/flow-plan/SKILL.md"},
	{"hive-rule-external-effects", "content/skills/flow-plan/SKILL.md"},
	{"hive-rule-external-effects", "content/skills/flow-build/SKILL.md"},
	{"hive-rule-ui-plan", "content/skills/flow-plan/SKILL.md"},
	{"hive-rule-ui-verification", "content/skills/flow-build/SKILL.md"},
	{"hive-rule-ui-verification", "content/skills/flow-verify/SKILL.md"},
	{"hive-rule-before-captures", "content/skills/flow-build/SKILL.md"},
	{"hive-rule-walkthrough", "content/skills/flow-build/SKILL.md"},
	{"hive-rule-walkthrough", "content/skills/flow-verify/SKILL.md"},
	{"hive-rule-viewer-open", "content/skills/flow-build/SKILL.md"},
	{"hive-rule-viewer-open", "content/skills/flow-verify/SKILL.md"},
	{"hive-rule-viewer-close", "content/skills/flow-close/SKILL.md"},
	{"hive-rule-change-folder", "content/skills/flow-plan/SKILL.md"},
	{"hive-rule-checkpoint", "content/skills/flow-build/SKILL.md"},
	{"hive-rule-checkpoint", "content/skills/flow-verify/SKILL.md"},
	{"hive-rule-checks", "content/skills/flow-build/SKILL.md"},
	{"hive-rule-checks", "content/skills/flow-verify/SKILL.md"},
	{"hive-rule-review-findings", "content/skills/flow-verify/SKILL.md"},
	{"hive-rule-project-verification", "content/skills/flow-plan/SKILL.md"},
	{"hive-rule-project-verification", "content/skills/flow-build/SKILL.md"},
	{"hive-rule-project-verification", "content/skills/flow-verify/SKILL.md"},
	{"hive-rule-stored-data", "content/skills/flow-build/SKILL.md"},
	{"hive-rule-stored-data", "content/skills/flow-deploy/SKILL.md"},
	{"hive-rule-security-boundaries", "content/skills/flow-build/SKILL.md"},
	{"hive-rule-suppression-review", "content/skills/flow-verify/SKILL.md"},
	{"hive-rule-scope-changes", "content/skills/flow-deploy/SKILL.md"},
	{"hive-rule-project-settings", "content/skills/flow-deploy/SKILL.md"},
	{"hive-rule-test-accounts", "content/skills/flow-deploy/references/deployment-checks.md"},
	{"hive-rule-design-stage", "content/skills/flow-figma-design/SKILL.md"},
	{"hive-rule-design-stage", "content/skills/hive-mode/SKILL.md"},
	{"hive-rule-report-endings", "content/skills/hive-mode/SKILL.md"},
	{"hive-rule-no-comments", "content/skills/flow-verify/SKILL.md"},
	{"hive-rule-blast-radius", "content/skills/flow-verify/SKILL.md"},
	{"hive-rule-prose", "content/skills/flow-verify/SKILL.md"},
	{"hive-rule-prose", "content/skills/hive-rule-git-commit/SKILL.md"},
	{"hive-rule-interrogate", "content/skills/hive-rule-delivery-ask/SKILL.md"},
	{"hive-rule-verification-floor", "content/skills/flow-build/SKILL.md"},
	{"hive-rule-verification-floor", "content/skills/flow-verify/SKILL.md"},
	{"hive-rule-verification-floor", "content/skills/flow-plan/SKILL.md"},
	{"hive-rule-trackers", "content/skills/hive-rule-ticket-state/SKILL.md"},
	{"hive-rule-trackers", "content/skills/flow-close/SKILL.md"},
	{"hive-rule-unattended-effects", "content/skills/unattended-delegation/SKILL.md"},
	{"unattended-delegation", "content/skills/flow-build/references/e2e-coordination.md"},
	{"unattended-delegation", "content/skills/hive-rule-unattended-effects/SKILL.md"},
	{"hive-rule-checkpoint", "content/skills/hive-rule-close-modes/SKILL.md"},
	{"hive-rule-defect-fix", "content/agents/review/hive-review-code.md"},
	{"hive-principle-index", "content/agents/review/hive-review-code.md"},
	{"hive-principle-index", "content/agents/development/hive-build-owner.md"},
	{"hive-rule-review-evidence", "content/agents/review/hive-review-code.md"},
	{"hive-rule-suppression-review", "content/agents/review/hive-review-code.md"},
	{"hive-rule-security-boundaries", "content/agents/review/hive-review-security.md"},
	{"hive-rule-tests", "content/agents/quality/hive-write-tests.md"},
	{"hive-rule-delegation", "content/guidance/global.md"},
	{"hive-rule-viewer", "content/skills/hive-rule-walkthrough/SKILL.md"},
	{"hive-rule-ui-example", "content/skills/hive-rule-ui-plan/SKILL.md"},
	{"hive-rule-child-brief", "content/skills/hive-rule-delegation/SKILL.md"},
}

func linkProblems(root string, links []leafLink) []string {
	var problems []string
	for _, link := range links {
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
		{"hive-rule-c", "content/skills/flow-build/SKILL.md"},
		{"hive-rule-d", "content/skills/flow-build/SKILL.md"},
	})
	if len(got) != 1 || !strings.Contains(got[0], "does not link hive-rule-d") {
		t.Errorf("unexpected problems: %v", got)
	}
}
