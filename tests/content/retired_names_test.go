package content

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// retiredSkills are skills Hive no longer ships. A rename or a merge leaves
// their names in text that readers act on, so this test fails when distributed
// guidance or a page people read names one as a skill. The names stay legal
// elsewhere: the legacy catalogue, history documents and session fixtures keep
// them on purpose, and some live on as another thing's name, such as the
// `local-instructions.md` reference of harness-audit.
var retiredSkills = []string{
	"flow-research", "flow-report", "flow-figma-design", "hive-ui-criteria",
	"flow-status", "security-advisory", "flow-walkthrough", "starlight-docs-site",
	"local-instructions", "engram-init-workspace", "workspace-conventions", "docs-cleanup",
}

// retiredSkillFiles are the files this test reads: what an agent loads and
// what a person reads to use Hive.
var retiredSkillFiles = []string{"content", "site/src/content/docs", "README.md", "llms.txt", "AGENTS.md"}

// skillNameForms returns the patterns that name `name` as a skill: its folder
// in a skills path or locator, its entrypoint, its name in backticks, the word
// "skill" after it, or a slash, dollar or at command.
func skillNameForms(name string) *regexp.Regexp {
	n := regexp.QuoteMeta(name)
	return regexp.MustCompile(`(?:skills/` + n + `(?:/|\b)|skill:` + n + `\b|\.\./` + n + `/SKILL\.md|` + "`" + n + "`" + `|\b` + n + ` skills?\b|(?:^|[\s(])[/$@]` + n + `\b)`)
}

// legacyLeafPrefix is the folder prefix the rule leaves had before they moved
// into hive-core; written in two parts so this file does not name it.
var legacyLeafPrefix = regexp.MustCompile(`\b` + "hive-" + "rule-" + `[a-z]`)

// retiredNameProblems returns one line per mention of a retired skill, or of a
// leaf's former folder, in the given files' text.
func retiredNameProblems(files map[string]string) []string {
	var problems []string
	for path, text := range files {
		for n, line := range strings.Split(text, "\n") {
			for _, name := range retiredSkills {
				if skillNameForms(name).MatchString(line) {
					problems = append(problems, fmt.Sprintf("%s:%d: names the retired skill %s", path, n+1, name))
				}
			}
			if legacyLeafPrefix.MatchString(line) {
				problems = append(problems, fmt.Sprintf("%s:%d: names a rule leaf by its former folder; leaves are files of hive-core", path, n+1))
			}
		}
	}
	return problems
}

func readRetiredNameScope(t *testing.T) map[string]string {
	t.Helper()
	root := repoRoot(t)
	files := map[string]string{}
	for _, entry := range retiredSkillFiles {
		err := filepath.WalkDir(filepath.Join(root, entry), func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				if d.Name() == "node_modules" {
					return filepath.SkipDir
				}
				return nil
			}
			switch filepath.Ext(path) {
			case ".md", ".mdx", ".txt", ".json", ".yaml", ".yml":
			default:
				return nil
			}
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			rel, _ := filepath.Rel(root, path)
			files[filepath.ToSlash(rel)] = string(data)
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	return files
}

func TestNoRetiredSkillIsNamed(t *testing.T) {
	for _, problem := range retiredNameProblems(readRetiredNameScope(t)) {
		t.Error(problem)
	}
}

func TestRetiredSkillFormsAndLegalNames(t *testing.T) {
	for _, line := range []string{
		"read [the walkthrough](../flow-walkthrough/SKILL.md)",
		"run `flow-research` first",
		"the local-instructions skill writes it",
		"type /flow-report to start",
		"see skill:docs-cleanup/SKILL.md",
		"content/skills/workspace-conventions/SKILL.md",
		"read " + "hive-" + "rule-questions before asking",
	} {
		if len(retiredNameProblems(map[string]string{"x.md": line})) == 0 {
			t.Errorf("%q is not reported", line)
		}
	}
	for _, line := range []string{
		"[local instructions](references/local-instructions.md)",
		"[the next-step block](report-endings.md#the-next-step-block)",
		"`<meta name=\"flow-report-assets\" content=\"external\">`",
		"see ../../tests/fixtures/workspace-conventions/baseline/global.md",
		"the hive-core skill stores the leaves",
	} {
		if problems := retiredNameProblems(map[string]string{"x.md": line}); len(problems) != 0 {
			t.Errorf("%q is reported: %v", line, problems)
		}
	}
}
