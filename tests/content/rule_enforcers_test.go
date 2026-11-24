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
	for name, mutate := range map[string][2]string{
		"renamed criterion":      {"`close_before_branch_cleanup`", "`no_such_criterion`"},
		"unknown status":         {"| detects | enforced |", "| detects | maybe |"},
		"no mechanism, enforced": {"| none | — | judgment |", "| none | — | enforced |"},
	} {
		doc := strings.Replace(string(b), mutate[0], mutate[1], 1)
		if doc == string(b) {
			t.Fatalf("%s: the mutation did not change the document; update the test to the table's current rows", name)
		}
		if len(checkRuleEnforcers(doc, repositoryRoot)) == 0 {
			t.Errorf("%s: checkRuleEnforcers accepted it", name)
		}
	}
}

var (
	enforcerKinds    = map[string]bool{"prevents": true, "detects": true, "—": true}
	enforcerStatuses = map[string]bool{"enforced": true, "judgment": true}
)

// checkRuleEnforcers returns one error per Mechanism cell path that does not
// exist under root, per criterion name in parentheses that the preceding Go
// file does not define as a string, per Kind or Status outside its values,
// per row without a mechanism that is not marked judgment, and one when the
// table has no rows.
func checkRuleEnforcers(doc, root string) []error {
	var errs []error
	rows := 0
	for _, line := range strings.Split(doc, "\n") {
		cells := strings.Split(strings.Trim(strings.TrimSpace(line), "|"), "|")
		if !strings.HasPrefix(strings.TrimSpace(line), "|") || len(cells) != 5 {
			continue
		}
		mechanism := strings.TrimSpace(cells[2])
		kind, status := strings.TrimSpace(cells[3]), strings.TrimSpace(cells[4])
		if mechanism == "Mechanism" || strings.HasPrefix(mechanism, "---") {
			continue
		}
		rows++
		if !enforcerKinds[kind] || !enforcerStatuses[status] {
			errs = append(errs, fmt.Errorf("rule-enforcers.md: row %q has kind %q and status %q", cells[0], kind, status))
		}
		if mechanism == "none" && status != "judgment" {
			errs = append(errs, fmt.Errorf("rule-enforcers.md: row %q has no mechanism but is %q", cells[0], status))
		}
		lastGo := ""
		for _, m := range backtickPath.FindAllStringSubmatch(mechanism, -1) {
			token := m[1]
			if !strings.Contains(token, "/") {
				if lastGo == "" || strings.ContainsAny(token, " ") {
					continue
				}
				src, err := os.ReadFile(filepath.Join(root, lastGo))
				if err == nil && !strings.Contains(string(src), `"`+token+`"`) {
					errs = append(errs, fmt.Errorf("rule-enforcers.md: %s defines no criterion %q", lastGo, token))
				}
				continue
			}
			if strings.ContainsAny(token, " ") {
				continue
			}
			if _, err := os.Stat(filepath.Join(root, token)); err != nil {
				errs = append(errs, fmt.Errorf("rule-enforcers.md: mechanism %s does not exist", token))
				continue
			}
			if strings.HasSuffix(token, ".go") {
				lastGo = token
			}
		}
	}
	if rows == 0 {
		errs = append(errs, fmt.Errorf("rule-enforcers.md has no table rows"))
	}
	return errs
}
