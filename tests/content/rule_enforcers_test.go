package content

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// The rule-enforcers table lists which of Hive's rules a mechanism enforces
// (root AGENTS.md «Measurement»). A row whose mechanism path no longer exists
// would claim an enforcement that is gone, so every path in its Mechanism
// column must resolve from the repository root.

const (
	ruleEnforcersPath = "../../_support/docs/harness-engineering/rule-enforcers.md"
	repositoryRoot    = "../.."
)

var backtickPath = regexp.MustCompile("`([^`]+)`")

func TestRuleEnforcerPathsExist(t *testing.T) {
	b, err := os.ReadFile(ruleEnforcersPath)
	if err != nil {
		t.Fatal(err)
	}
	for _, err := range checkRuleEnforcers(string(b), repositoryRoot) {
		t.Error(err)
	}
}

func TestRuleEnforcersDetectMissingPath(t *testing.T) {
	b, err := os.ReadFile(ruleEnforcersPath)
	if err != nil {
		t.Fatal(err)
	}
	doc := strings.Replace(string(b), "`tests/content/skill_size_test.go`", "`tests/content/no_such_enforcer_test.go`", 1)
	if doc == string(b) {
		t.Fatal("the mutation did not change the document; update the test to the table's current rows")
	}
	if len(checkRuleEnforcers(doc, repositoryRoot)) == 0 {
		t.Error("checkRuleEnforcers accepted a mechanism path that does not exist")
	}
}

// checkRuleEnforcers returns one error per Mechanism cell path that does not
// exist under root, and one when the table has no rows.
func checkRuleEnforcers(doc, root string) []error {
	var errs []error
	rows := 0
	for _, line := range strings.Split(doc, "\n") {
		cells := strings.Split(strings.Trim(strings.TrimSpace(line), "|"), "|")
		if !strings.HasPrefix(strings.TrimSpace(line), "|") || len(cells) != 5 {
			continue
		}
		mechanism := strings.TrimSpace(cells[2])
		if mechanism == "Mechanism" || strings.HasPrefix(mechanism, "---") {
			continue
		}
		rows++
		for _, m := range backtickPath.FindAllStringSubmatch(mechanism, -1) {
			path := m[1]
			if strings.ContainsAny(path, " ") || !strings.Contains(path, "/") {
				continue
			}
			if _, err := os.Stat(filepath.Join(root, path)); err != nil {
				errs = append(errs, fmt.Errorf("rule-enforcers.md: mechanism %s does not exist", path))
			}
		}
	}
	if rows == 0 {
		errs = append(errs, fmt.Errorf("rule-enforcers.md has no table rows"))
	}
	return errs
}
