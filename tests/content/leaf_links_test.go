package content

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// leafLink names a file that must link a leaf at the step where the leaf's
// rule applies, so the rule is loaded at that moment.
type leafLink struct {
	leaf string
	from string
}

// leafLinks lists, per leaf, the files that link it at the step where its rule
// applies. A rule leaf is named by its file name without its domain folder; the
// principle index by its path under hive-core; the language leaf and a flow
// skill by their skill name.
var leafLinks = []leafLink{
	{"git-commit", "content/skills/flow-build/SKILL.md"},
	{"git-commit", "content/skills/git-workflow/SKILL.md"},
	{"ticket-state", "content/skills/flow-build/SKILL.md"},
	{"ticket-state", "content/skills/flow-close/SKILL.md"},
	{"ticket-state", "content/skills/flow-deploy/SKILL.md"},
	{"close-modes", "content/skills/flow-close/SKILL.md"},
	{"close-hygiene", "content/skills/flow-close/SKILL.md"},
	{"close-cleanup", "content/skills/flow-close/SKILL.md"},
	{"proportionality", "content/skills/flow-build/SKILL.md"},
	{"close-report", "content/skills/flow-close/SKILL.md"},
	{"report-endings", "content/skills/flow-close/SKILL.md"},
	{"close-report", "content/skills/git-workflow/SKILL.md"},
	{"close-report", "content/skills/unattended-delegation/SKILL.md"},
	{"merge-gate", "content/skills/git-workflow/SKILL.md"},
	{"merge-gate", "content/skills/flow-verify/SKILL.md"},
	{"merge-gate", "content/skills/flow-deploy/SKILL.md"},
	{"pr-reviews", "content/skills/git-workflow/SKILL.md"},
	{"promotion-path", "content/skills/git-workflow/SKILL.md"},
	{"promotion-path", "content/skills/flow-deploy/SKILL.md"},
	{"reuse-before-writing", "content/skills/flow-build/SKILL.md"},
	{"principles/index", "content/skills/flow-build/SKILL.md"},
	{"hive-typescript", "content/skills/hive-core/principles/index.md"},
	{"reuse-before-writing", "content/skills/flow-plan/SKILL.md"},
	{"tests", "content/skills/flow-build/SKILL.md"},
	{"tests", "content/skills/flow-plan/SKILL.md"},
	{"tests-no-runner", "content/skills/flow-plan/SKILL.md"},
	{"defect-fix", "content/skills/flow-build/SKILL.md"},
	{"defect-fix", "content/skills/flow-verify/SKILL.md"},
	{"delegation", "content/skills/flow-build/SKILL.md"},
	{"delegation", "content/skills/flow-deploy/SKILL.md"},
	{"reviewer-brief", "content/skills/flow-plan/SKILL.md"},
	{"reviewer-brief", "content/skills/flow-build/SKILL.md"},
	{"reviewer-brief", "content/skills/flow-verify/SKILL.md"},
	{"business-rules", "content/skills/flow-plan/SKILL.md"},
	{"business-rules", "content/skills/flow-build/SKILL.md"},
	{"business-rules", "content/skills/flow-verify/SKILL.md"},
	{"business-rules", "content/skills/flow-close/SKILL.md"},
	{"business-rules", "content/skills/create-design/SKILL.md"},
	{"business-ticket", "content/skills/flow-build/SKILL.md"},
	{"business-ticket", "content/skills/create-design/SKILL.md"},
	{"external-facts", "content/skills/flow-plan/SKILL.md"},
	{"external-effects", "content/skills/flow-plan/SKILL.md"},
	{"external-effects", "content/skills/flow-build/SKILL.md"},
	{"ui-plan", "content/skills/flow-plan/SKILL.md"},
	{"ui-verification", "content/skills/flow-build/SKILL.md"},
	{"ui-verification", "content/skills/flow-verify/SKILL.md"},
	{"before-captures", "content/skills/flow-build/SKILL.md"},
	{"walkthrough", "content/skills/flow-build/SKILL.md"},
	{"walkthrough", "content/skills/flow-verify/SKILL.md"},
	{"viewer-open", "content/skills/flow-build/references/first-view.md"},
	{"viewer-open", "content/skills/flow-verify/SKILL.md"},
	{"viewer-close", "content/skills/flow-close/SKILL.md"},
	{"plan-record", "content/skills/flow-plan/SKILL.md"},
	{"checkpoint", "content/skills/flow-build/SKILL.md"},
	{"checkpoint", "content/skills/flow-verify/SKILL.md"},
	{"checks", "content/skills/flow-build/SKILL.md"},
	{"checks", "content/skills/flow-verify/SKILL.md"},
	{"review-findings", "content/skills/flow-verify/SKILL.md"},
	{"project-verification", "content/skills/flow-plan/SKILL.md"},
	{"project-verification", "content/skills/flow-build/SKILL.md"},
	{"project-verification", "content/skills/flow-verify/SKILL.md"},
	{"stored-data", "content/skills/flow-build/SKILL.md"},
	{"stored-data", "content/skills/flow-deploy/SKILL.md"},
	{"security-boundaries", "content/skills/flow-build/SKILL.md"},
	{"suppression-review", "content/skills/flow-verify/SKILL.md"},
	{"scope-changes", "content/skills/flow-deploy/SKILL.md"},
	{"project-settings", "content/skills/flow-deploy/SKILL.md"},
	{"test-accounts", "content/skills/flow-deploy/references/deployment-checks.md"},
	{"design-stage", "content/skills/create-design/SKILL.md"},
	{"design-stage", "content/skills/hive-mode/SKILL.md"},
	{"report-endings", "content/skills/hive-mode/SKILL.md"},
	{"no-comments", "content/skills/flow-verify/SKILL.md"},
	{"blast-radius", "content/skills/flow-verify/SKILL.md"},
	{"prose", "content/skills/flow-verify/SKILL.md"},
	{"prose", "content/skills/hive-core/git/git-commit.md"},
	{"interrogate", "content/skills/flow-verify/SKILL.md"},
	{"verification-floor", "content/skills/flow-build/SKILL.md"},
	{"verification-floor", "content/skills/flow-verify/SKILL.md"},
	{"verification-floor", "content/skills/flow-plan/SKILL.md"},
	{"trackers", "content/skills/hive-core/tracker/ticket-state.md"},
	{"trackers", "content/skills/flow-close/SKILL.md"},
	{"unattended-effects", "content/skills/unattended-delegation/SKILL.md"},
	{"loops", "content/skills/unattended-delegation/SKILL.md"},
	{"loops", "content/skills/hive-mode/playbooks/perf.md"},
	{"external-effects", "content/skills/hive-core/plan/loops.md"},
	{"unattended-effects", "content/skills/hive-core/plan/loops.md"},
	{"unattended-delegation", "content/skills/flow-build/references/e2e-coordination.md"},
	{"unattended-delegation", "content/skills/hive-core/delegation/unattended-effects.md"},
	{"checkpoint", "content/skills/hive-core/close/close-modes.md"},
	{"defect-fix", "content/agents/review/hive-review-code.md"},
	{"principles/index", "content/agents/review/hive-review-code.md"},
	{"principles/index", "content/agents/development/hive-build-owner.md"},
	{"principles/index", "content/agents/development/hive-ticket-owner.md"},
	{"review-evidence", "content/agents/review/hive-review-code.md"},
	{"suppression-review", "content/agents/review/hive-review-code.md"},
	{"security-boundaries", "content/agents/review/hive-review-security.md"},
	{"tests", "content/agents/quality/hive-write-tests.md"},
	{"delegation", "content/guidance/global.md"},
	{"viewer", "content/skills/hive-core/ui/walkthrough.md"},
	{"ui-example", "content/skills/hive-core/ui/ui-plan.md"},
	{"child-brief", "content/skills/hive-core/delegation/delegation.md"},
}

// leafLinkPatternFor returns the pattern a file must match to link the leaf: a
// rule leaf is a file of a hive-core domain folder, linked by its path or, from
// a sibling leaf, by its file name; a name with a folder, such as the principle
// index, is linked by that path; any other leaf is a skill, linked by its
// entrypoint.
func leafLinkPatternFor(leaf string) *regexp.Regexp {
	if leaf == languageLeafName || leaf == "unattended-delegation" {
		return regexp.MustCompile(regexp.QuoteMeta(leaf + "/SKILL.md"))
	}
	if strings.Contains(leaf, "/") {
		return regexp.MustCompile(`(^|[/(])` + regexp.QuoteMeta(leaf+".md"))
	}
	return regexp.MustCompile(`(^|[/(])` + regexp.QuoteMeta(leaf+".md"))
}

func linkProblems(root string, links []leafLink) []string {
	var problems []string
	for _, link := range links {
		data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(link.from)))
		if err != nil {
			problems = append(problems, fmt.Sprintf("%s: %v", link.from, err))
			continue
		}
		if !leafLinkPatternFor(link.leaf).MatchString(string(data)) {
			problems = append(problems, fmt.Sprintf("%s does not link %s", link.from, link.leaf))
		}
	}
	return problems
}

func TestFlowBuildLinksFirstView(t *testing.T) {
	data, err := os.ReadFile("../../content/skills/flow-build/SKILL.md")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "(references/first-view.md)") {
		t.Error("content/skills/flow-build/SKILL.md does not link references/first-view.md")
	}
}

func TestLeavesAreLinkedWhereTheyApply(t *testing.T) {
	for _, problem := range linkProblems("../..", leafLinks) {
		t.Error(problem)
	}
}

func TestLeavesAreLinkedCatchesMissingLink(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "content", "skills", "flow-build", "SKILL.md"), "Commit, reading [c](../hive-core/build/c.md).\n")
	got := linkProblems(root, []leafLink{
		{"c", "content/skills/flow-build/SKILL.md"},
		{"principles/index", "content/skills/flow-build/SKILL.md"},
		{"d", "content/skills/flow-build/SKILL.md"},
	})
	if len(got) != 2 || !strings.Contains(got[0], "does not link principles/index") || !strings.Contains(got[1], "does not link d") {
		t.Errorf("unexpected problems: %v", got)
	}
}
