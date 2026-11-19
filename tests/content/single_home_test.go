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
	{"Never run `git add -A`", "content/skills/hive-rule-git-commit/SKILL.md", false},
	// Repeated today: global.md «Delivery» and git-workflow.
	{"Before the validation, or before the push when the mode", "content/skills/hive-rule-foreign-commits/SKILL.md", false},
	{"is done when its changes are integrated", "content/skills/hive-rule-ticket-state/SKILL.md", false},
	// Repeated today: flow-build and flow-verify.
	{"adds, replaces or cancels a business rule, record it in", "content/skills/hive-rule-business-rules/SKILL.md", false},
	{"Name the cause before you edit", "content/skills/hive-rule-defect-fix/SKILL.md", false},
	// Repeated today: flow-research and flow-figma-design.
	{"yourself that the host has a connection that can write", "content/skills/hive-rule-design-stage/SKILL.md", false},
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
