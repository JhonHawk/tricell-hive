package content

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
)

// retiredSkills are skills Hive no longer ships. A rename or a merge leaves
// their names in text that readers act on, so this test fails when distributed
// guidance or a page people read names one as a skill. The names stay legal
// elsewhere: the legacy catalogue, history documents and session fixtures keep
// them on purpose, and some live on as another thing's name, such as the
// `local-instructions.md` reference of audit-harness. Add a name here when a
// skill is retired.
var retiredSkills = []string{
	"flow-research", "flow-report", "flow-figma-design", "hive-ui-criteria",
	"flow-status", "security-advisory", "flow-walkthrough", "starlight-docs-site",
	"local-instructions", "engram-init-workspace", "workspace-conventions", "docs-cleanup",
	"flow-security", "workspace-archive", "comment-audit", "test-audit", "harness-audit",
}

// upstreamMentions name a third-party skill that shares a retired name: the
// attribution keeps the upstream's own name and URL, so its exact text is
// masked before the retired forms are matched.
var upstreamMentions = []string{
	"the `test-audit` skill of [openclaw/openclaw](https://github.com/openclaw/openclaw/tree/main/.agents/skills/test-audit)",
}

// retiredSkillFiles are the files this test reads: what an agent loads and
// what a person reads to use Hive.
var retiredSkillFiles = []string{"content", "site/src/content/docs", "README.md", "llms.txt", "AGENTS.md"}

// skillNameForms holds, per retired skill, the pattern that names it as a
// skill: its folder in a skills path or locator, its entrypoint, its name in
// backticks, the word "skill" after it (also across a line break), or a slash,
// dollar or at command after any character that is not part of a word or a
// path, such as a space, a backtick, an underscore, a quote, a bracket or a
// table bar. A name is matched whole: one that only starts with a retired
// name, such as flow-report-assets, is another name.
var skillNameForms = func() map[string]*regexp.Regexp {
	forms := map[string]*regexp.Regexp{}
	const end = `(?:$|[^A-Za-z0-9-])`
	for _, name := range retiredSkills {
		n := regexp.QuoteMeta(name)
		forms[name] = regexp.MustCompile(`(?m)(?:` +
			`skills/` + n + `(?:/|` + end + `)` +
			`|skill:` + n + end +
			`|\.\./` + n + `/SKILL\.md` +
			"|`" + n + "`" +
			`|(?:^|[^A-Za-z0-9_-])` + n + `\s+skills?` + end +
			`|(?:^|[^A-Za-z0-9./-])[/$@]` + n + end +
			`)`)
	}
	return forms
}()

// legacyLeafPrefix is the folder prefix the rule leaves had before they moved
// into hive-core; written in two parts so this file does not name it.
var legacyLeafPrefix = regexp.MustCompile(`\b` + "hive-" + "rule-" + `[a-z]`)

// legacyPrincipleSkillName matches the skill-name prefix the principles had before
// they moved into hive-core/principles; written in two parts for the same reason.
// A principle is a file there, not a skill.
var legacyPrincipleSkillName = regexp.MustCompile(`\b` + "hive-" + "principle-" + `[a-z]`)

// legacyReferencesPath is the folder the rule leaves shared before the domain
// folders replaced it; written in two parts for the same reason.
var legacyReferencesPath = regexp.MustCompile(regexp.QuoteMeta("hive-core/" + "references/"))

// retiredNameProblems returns one line per mention of a retired skill, or of a
// leaf's or a principle's former home, in the given files' text, sorted by file and line.
func retiredNameProblems(files map[string]string) []string {
	var problems []string
	lineOf := func(text string, offset int) int { return strings.Count(text[:offset], "\n") + 1 }
	for path, text := range files {
		for _, mention := range upstreamMentions {
			text = strings.ReplaceAll(text, mention, strings.Repeat(" ", len(mention)))
		}
		for _, name := range retiredSkills {
			for _, m := range skillNameForms[name].FindAllStringIndex(text, -1) {
				problems = append(problems, fmt.Sprintf("%s:%d: names the retired skill %s", path, lineOf(text, m[0]), name))
			}
		}
		for _, m := range legacyLeafPrefix.FindAllStringIndex(text, -1) {
			problems = append(problems, fmt.Sprintf("%s:%d: names a rule leaf by its former folder; leaves are files of hive-core", path, lineOf(text, m[0])))
		}
		for _, m := range legacyPrincipleSkillName.FindAllStringIndex(text, -1) {
			problems = append(problems, fmt.Sprintf("%s:%d: names a principle as a skill; principles are files of hive-core/principles", path, lineOf(text, m[0])))
		}
		for _, m := range legacyReferencesPath.FindAllStringIndex(text, -1) {
			problems = append(problems, fmt.Sprintf("%s:%d: names the former hive-core references folder; leaves live in the domain folders", path, lineOf(text, m[0])))
		}
	}
	sort.Strings(problems)
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
	for _, text := range []string{
		"read [the walkthrough](../flow-walkthrough/SKILL.md)",
		"run `flow-research` first",
		"the local-instructions skill writes it",
		"and the flow-report\nskill writes it",
		"type /flow-report to start",
		"type `/flow-status` to see the backlog",
		"run `$flow-research` in Codex",
		"**@flow-report** in Cursor",
		"__/flow-status__ in bold",
		"set the command to \"/flow-status\"",
		"say '/flow-status' aloud",
		"<code>/flow-report</code>",
		"|/flow-report|",
		"see [/flow-report](https://x) for details",
		"see skill:docs-cleanup/SKILL.md",
		"content/skills/workspace-conventions/SKILL.md",
		"run `flow-security` before the push",
		"read " + "hive-" + "rule-questions before asking",
		"read " + "hive-" + "principle-index before writing",
		"content/skills/" + "hive-" + "principle-boundary-discipline/SKILL.md",
		"the leaf is `hive-core/" + "references/questions.md`",
		"run `harness-audit` on the repository",
		"see https://github.com/openclaw/openclaw/tree/main/.agents/skills/test-audit",
	} {
		if len(retiredNameProblems(map[string]string{"x.md": text})) == 0 {
			t.Errorf("%q is not reported", text)
		}
	}
	for _, text := range []string{
		"[local instructions](references/local-instructions.md)",
		"[the next-step block](report-endings.md#the-next-step-block)",
		"`<meta name=\"flow-report-assets\" content=\"external\">`",
		"see ../../tests/fixtures/workspace-conventions/baseline/global.md",
		"the hive-core skill stores the leaves",
		"read hive-core/principles/index.md before writing",
		"`/flow-report-assets/app.css`",
		"content/skills/flow-report-assets/x",
		"the use-flow-status skill is another one",
		"https://example.com/flow-report",
		"> Adapted from the `test-audit` skill of [openclaw/openclaw](https://github.com/openclaw/openclaw/tree/main/.agents/skills/test-audit) (MIT).",
	} {
		if problems := retiredNameProblems(map[string]string{"x.md": text}); len(problems) != 0 {
			t.Errorf("%q is reported: %v", text, problems)
		}
	}
}
