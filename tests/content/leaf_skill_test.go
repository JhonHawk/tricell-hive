package content

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
	"unicode/utf8"
)

// Leaf skills are small skills the model does not load on its own: a flow step
// or an index line names them with the condition to read them, and the agent
// reads the file. These limits keep each leaf readable in one pass and keep a
// host's skill catalog from carrying them.
const (
	leafSizeLimit          = 4000
	leafExceptionSizeLimit = 7000
	// leafReportEndingsSizeLimit is the own limit of hive-rule-report-endings,
	// the one home of how every message ends (decision D11 A of the
	// reporting-family-leaves plan).
	leafReportEndingsSizeLimit = 9000
	leafDescriptionLimit       = 200
	playbookSizeLimit          = 7000
	// playbookModePhrase is the fixed phrase through which a playbook refers to
	// the delivery mode: it never fixes or replaces the mode itself.
	playbookModePhrase = "under the delivery mode in force"
	ruleLeavesHeading  = "## Rule leaves"
	principleIndexName = "hive-principle-index"
	languageLeafName   = "hive-typescript"
)

// leafSizeExceptions names each leaf whose single rule cannot fit the common
// limit without rewording it; each stays at or under leafExceptionSizeLimit.
var leafSizeExceptions = map[string]bool{
	// One home for every question to the user: what to ask, the format and the
	// channel per host, read together before any round; the unanswered card,
	// a moment of its own, is its reference.
	"hive-rule-questions": true,
	// One moment, the merge: the CI wait, the current base, the reviewer hold
	// and the bypass grants that can lift it, read together before merging.
	"hive-rule-merge-gate": true,
	// One report in several variants (completion, open pull request,
	// deployment, hold, pause) plus the findings at the close, read together.
	"hive-rule-close-report": true,
}

// playbookKinds is the closed set of playbooks of hive-mode, one file each.
var playbookKinds = []string{
	"bugfix", "data-change", "deliverable", "dependency-security", "feature",
	"hotfix", "infra-ci", "investigate", "perf", "pr-review-deploy",
	"refactor", "ticket-status", "ui-change",
}

var deliveryModeNamePattern = regexp.MustCompile("`(interactive|automatic|ready|direct-base|hold)`")

var leafLinkPattern = regexp.MustCompile(`(hive-rule-[a-z0-9-]+|hive-principle-[a-z0-9-]+|hive-typescript)/SKILL\.md`)

func isLeafName(name string) bool {
	switch {
	case strings.HasPrefix(name, "hive-rule-"):
		return true
	case strings.HasPrefix(name, "hive-principle-"):
		return name != principleIndexName
	}
	return name == languageLeafName
}

// sectionText returns the text of the `## ` section that starts with heading,
// up to the next `## ` heading.
func sectionText(text, heading string) (string, bool) {
	start := strings.Index(text, "\n"+heading+"\n")
	if start < 0 {
		return "", false
	}
	rest := text[start+len(heading)+2:]
	if end := strings.Index(rest, "\n## "); end >= 0 {
		rest = rest[:end]
	}
	return rest, true
}

func countLinks(text string) map[string]int {
	counts := map[string]int{}
	for _, match := range leafLinkPattern.FindAllStringSubmatch(text, -1) {
		counts[match[1]]++
	}
	return counts
}

// leafProblems checks every leaf skill under root/content/skills against its
// invariants and returns one line per problem.
func leafProblems(root string) []string {
	var problems []string
	skillsDir := filepath.Join(root, "content", "skills")
	entries, err := os.ReadDir(skillsDir)
	if err != nil {
		return []string{err.Error()}
	}
	leaves := map[string]bool{}
	for _, entry := range entries {
		if entry.IsDir() && isLeafName(entry.Name()) {
			leaves[entry.Name()] = true
		}
	}

	globalText := ""
	if data, err := os.ReadFile(filepath.Join(root, "content", "guidance", "global.md")); err == nil {
		globalText = string(data)
	}
	ruleIndex, hasRuleIndex := sectionText(globalText, ruleLeavesHeading)
	ruleCounts := countLinks(ruleIndex)
	principleCounts := map[string]int{}
	principleIndexPath := filepath.Join(skillsDir, principleIndexName, "SKILL.md")
	principleData, principleErr := os.ReadFile(principleIndexPath)
	if principleErr == nil {
		principleCounts = countLinks(string(principleData))
	}

	names := make([]string, 0, len(leaves))
	for name := range leaves {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		path := filepath.Join(skillsDir, name, "SKILL.md")
		data, err := os.ReadFile(path)
		if err != nil {
			problems = append(problems, fmt.Sprintf("%s: no SKILL.md", name))
			continue
		}
		text := string(data)
		description, ok, err := frontmatterScalar(text, "description")
		switch {
		case err != nil || !ok:
			problems = append(problems, fmt.Sprintf("%s: description is missing or not a single-line scalar", name))
		case utf8.RuneCountInString(description) > leafDescriptionLimit:
			problems = append(problems, fmt.Sprintf("%s: description has %d characters, limit %d", name, utf8.RuneCountInString(description), leafDescriptionLimit))
		}
		hidden, _, _ := frontmatterScalar(text, "disable-model-invocation")
		if name == languageLeafName {
			if hidden == "true" {
				problems = append(problems, fmt.Sprintf("%s: must stay model-invocable so `paths` can load it", name))
			}
			if !strings.Contains(frontmatterBlock(text), "\npaths:") {
				problems = append(problems, fmt.Sprintf("%s: frontmatter has no `paths`", name))
			}
		} else if hidden != "true" {
			problems = append(problems, fmt.Sprintf("%s: frontmatter lacks `disable-model-invocation: true`", name))
		}
		limit := leafSizeLimit
		if leafSizeExceptions[name] {
			limit = leafExceptionSizeLimit
		}
		if name == "hive-rule-report-endings" {
			limit = leafReportEndingsSizeLimit
		}
		if len(data) > limit {
			problems = append(problems, fmt.Sprintf("%s: SKILL.md is %d bytes, limit %d", name, len(data), limit))
		}
		if strings.HasPrefix(name, "hive-rule-") {
			if !hasRuleIndex {
				problems = append(problems, fmt.Sprintf("%s: global.md has no %q section", name, ruleLeavesHeading))
			} else if n := ruleCounts[name]; n != 1 {
				problems = append(problems, fmt.Sprintf("%s: listed %d times in %q, want 1", name, n, ruleLeavesHeading))
			}
		} else {
			if principleErr != nil {
				problems = append(problems, fmt.Sprintf("%s: %s has no SKILL.md", name, principleIndexName))
			} else if n := principleCounts[name]; n != 1 {
				problems = append(problems, fmt.Sprintf("%s: listed %d times in %s, want 1", name, n, principleIndexName))
			}
		}
	}
	for name := range ruleCounts {
		if !leaves[name] {
			problems = append(problems, fmt.Sprintf("%q lists %s, which does not exist", ruleLeavesHeading, name))
		}
	}
	for name := range principleCounts {
		if !leaves[name] {
			problems = append(problems, fmt.Sprintf("%s lists %s, which does not exist", principleIndexName, name))
		}
	}

	playbooks, _ := filepath.Glob(filepath.Join(skillsDir, "hive-mode", "playbooks", "*.md"))
	present := map[string]bool{}
	for _, path := range playbooks {
		present[strings.TrimSuffix(filepath.Base(path), ".md")] = true
		info, err := os.Stat(path)
		if err == nil && info.Size() > playbookSizeLimit {
			problems = append(problems, fmt.Sprintf("playbook %s is %d bytes, limit %d", filepath.Base(path), info.Size(), playbookSizeLimit))
		}
		if data, err := os.ReadFile(path); err == nil {
			for _, mode := range playbookModeNames(string(data)) {
				problems = append(problems, fmt.Sprintf("playbook %s names the delivery mode %s outside the phrase %q", filepath.Base(path), mode, playbookModePhrase))
			}
		}
	}
	wanted := map[string]bool{}
	for _, kind := range playbookKinds {
		wanted[kind] = true
		if !present[kind] {
			problems = append(problems, fmt.Sprintf("hive-mode/playbooks lacks %s.md", kind))
		}
	}
	for kind := range present {
		if !wanted[kind] {
			problems = append(problems, fmt.Sprintf("hive-mode/playbooks holds %s.md, which is not one of the %d playbooks", kind, len(playbookKinds)))
		}
	}
	sort.Strings(problems)
	return problems
}

// playbookModeNames returns each delivery mode a playbook names in backticks
// outside the fixed phrase. A playbook refers to the mode in force and never
// fixes or replaces it, so it carries no mode name of its own.
func playbookModeNames(text string) []string {
	text = strings.ReplaceAll(text, playbookModePhrase, "")
	var names []string
	for _, match := range deliveryModeNamePattern.FindAllStringSubmatch(text, -1) {
		names = append(names, match[0])
	}
	return names
}

// frontmatterBlock returns the frontmatter text, starting with a newline, or
// an empty string when the document has none.
func frontmatterBlock(text string) string {
	if !strings.HasPrefix(text, "---\n") {
		return ""
	}
	end := strings.Index(text[4:], "\n---\n")
	if end < 0 {
		return ""
	}
	return "\n" + text[4:4+end]
}

func TestLeafSkillInvariants(t *testing.T) {
	for _, problem := range leafProblems("../..") {
		t.Error(problem)
	}
}

func writeFile(t *testing.T, path, text string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}
}

func leafDoc(name, extra string) string {
	return "---\nname: " + name + "\ndescription: Read before a commit.\n" + extra + "---\n\n# Body\n"
}

func TestLeafSkillInvariantsCatchViolations(t *testing.T) {
	root := t.TempDir()
	skills := filepath.Join(root, "content", "skills")
	writeFile(t, filepath.Join(root, "content", "guidance", "global.md"),
		"# Guide\n\n"+ruleLeavesHeading+"\n\n- **Before a commit:** read [a](skill:hive-rule-a/SKILL.md).\n- **Again:** read [a](skill:hive-rule-a/SKILL.md).\n- **Gone:** read [x](skill:hive-rule-gone/SKILL.md).\n\n## Next\n")
	writeFile(t, filepath.Join(skills, "hive-rule-a", "SKILL.md"), leafDoc("hive-rule-a", "disable-model-invocation: true\n"))
	writeFile(t, filepath.Join(skills, "hive-rule-b", "SKILL.md"), leafDoc("hive-rule-b", ""))
	writeFile(t, filepath.Join(skills, "hive-rule-big", "SKILL.md"), leafDoc("hive-rule-big", "disable-model-invocation: true\n")+strings.Repeat("x", leafSizeLimit))
	writeFile(t, filepath.Join(skills, languageLeafName, "SKILL.md"), leafDoc(languageLeafName, "disable-model-invocation: true\n"))
	writeFile(t, filepath.Join(skills, "hive-mode", "playbooks", "bugfix.md"), strings.Repeat("y", playbookSizeLimit+1))
	writeFile(t, filepath.Join(skills, "hive-mode", "playbooks", "hotfix.md"), "Push "+playbookModePhrase+", then stop in `hold`.\n")
	writeFile(t, filepath.Join(skills, "hive-mode", "playbooks", "extra.md"), "Not one of the playbooks.\n")

	got := strings.Join(leafProblems(root), "\n")
	for _, want := range []string{
		"hive-rule-a: listed 2 times",
		"hive-rule-b: frontmatter lacks `disable-model-invocation: true`",
		"hive-rule-b: listed 0 times",
		"hive-rule-big: SKILL.md is",
		"lists hive-rule-gone, which does not exist",
		"hive-typescript: must stay model-invocable",
		"hive-typescript: frontmatter has no `paths`",
		"hive-typescript: hive-principle-index has no SKILL.md",
		"playbook bugfix.md is",
		"playbook hotfix.md names the delivery mode `hold` outside the phrase",
		"hive-mode/playbooks holds extra.md",
		"hive-mode/playbooks lacks feature.md",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("missing problem %q in:\n%s", want, got)
		}
	}
}

func TestPlaybookModeNames(t *testing.T) {
	if got := playbookModeNames("Stops " + playbookModePhrase + "; no `hold` here."); len(got) != 1 || got[0] != "`hold`" {
		t.Errorf("want only `hold`, got %v", got)
	}
	if got := playbookModeNames("Pushes " + playbookModePhrase + ". The mode is automatic."); len(got) != 0 {
		t.Errorf("the phrase and a bare word are allowed, got %v", got)
	}
}
