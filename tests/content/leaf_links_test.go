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
	{"hive-rule-ticket-state", "content/skills/flow-deploy/SKILL.md", false},
	{"hive-rule-close-modes", "content/skills/flow-close/SKILL.md", false},
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
