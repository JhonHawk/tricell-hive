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

// Leaves are small files the model does not load on its own: a flow step or an
// index line names them with the condition to read them, and the agent reads
// the file. The rule leaves are files of the hidden hive-core skill, without
// frontmatter; the principles and the language leaf stay skills. These limits
// keep each leaf readable in one pass and keep a host's skill catalog from
// carrying them.
const (
	leafSizeLimit          = 4000
	leafExceptionSizeLimit = 7000
	// leafReportEndingsSizeLimit is the own limit of the report-endings leaf,
	// the one home of how every message ends (decision D11 A of the
	// reporting-family-leaves plan).
	leafReportEndingsSizeLimit = 9000
	leafDescriptionLimit       = 200
	playbookSizeLimit          = 7000
	// playbookModePhrase is the fixed phrase through which a playbook refers to
	// the delivery mode: it never fixes or replaces the mode itself.
	playbookModePhrase = "under the delivery mode in force"
	// The next three constants are what the reporting line of every playbook
	// links: how its message ends and the final-report row that the delivery
	// mode in force picks (AC7 of the reporting-family-leaves plan).
	playbookEndingsLink = "hive-core/references/report-endings.md"
	playbookReportLink  = "hive-core/references/close-report.md#variants"
	playbookReportModes = "the delivery mode in force picks"
	ruleLeavesHeading   = "## Rule leaves"
	principleIndexName  = "hive-principle-index"
	languageLeafName    = "hive-typescript"
	coreSkillName       = "hive-core"
	// legacyRulePrefix is the directory prefix the rule leaves had before they
	// moved into hive-core; no directory with it may remain.
	legacyRulePrefix = "hive-" + "rule-"
)

// leafSizeExceptions names each rule leaf, by its file name in
// hive-core/references, whose single rule cannot fit the common limit without
// rewording it; each stays at or under leafExceptionSizeLimit.
var leafSizeExceptions = map[string]bool{
	// One home for every question to the user: what to ask, the format and the
	// channel per host, read together before any round; the unanswered card,
	// a moment of its own, is its reference.
	"questions": true,
	// One moment, the merge: the CI wait, the current base, the reviewer hold
	// and the bypass grants that can lift it, read together before merging.
	"merge-gate": true,
	// One report in several variants (completion, open pull request,
	// deployment, hold, pause) plus the findings at the close, read together.
	"close-report": true,
}

// playbookKinds is the closed set of playbooks of hive-mode, one file each.
var playbookKinds = []string{
	"bugfix", "data-change", "deliverable", "dependency-security", "feature",
	"hotfix", "infra-ci", "investigate", "perf", "pr-review-deploy",
	"refactor", "ticket-status", "ui-change",
}

var deliveryModeNamePattern = regexp.MustCompile("`(interactive|automatic|ready|direct-base|hold)`")

var leafLinkPattern = regexp.MustCompile(`(hive-principle-[a-z0-9-]+|hive-typescript)/SKILL\.md`)

// ruleLeafLinkPattern matches a link to a rule leaf of hive-core and captures
// the leaf's file name without its extension.
var ruleLeafLinkPattern = regexp.MustCompile(`hive-core/references/([a-z0-9-]+)\.md`)

func isLeafName(name string) bool {
	switch {
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

func countRuleLinks(text string) map[string]int {
	counts := map[string]int{}
	for _, match := range ruleLeafLinkPattern.FindAllStringSubmatch(text, -1) {
		counts[match[1]]++
	}
	return counts
}

// ruleLeafProblems checks the hive-core skill and each rule leaf in its
// references folder: no frontmatter, a title, a first body line that starts
// with "Read", the size limit, and exactly one line of the global index.
func ruleLeafProblems(skillsDir string, ruleIndex string, hasRuleIndex bool) []string {
	var problems []string
	coreDir := filepath.Join(skillsDir, coreSkillName)
	coreData, err := os.ReadFile(filepath.Join(coreDir, "SKILL.md"))
	if err != nil {
		problems = append(problems, fmt.Sprintf("%s: no SKILL.md", coreSkillName))
	} else {
		coreText := string(coreData)
		if hidden, _, _ := frontmatterScalar(coreText, "disable-model-invocation"); hidden != "true" {
			problems = append(problems, fmt.Sprintf("%s: frontmatter lacks `disable-model-invocation: true`", coreSkillName))
		}
		if len(ruleLeafLinkPattern.FindAllString(coreText, -1)) > 0 {
			problems = append(problems, fmt.Sprintf("%s: SKILL.md lists leaves; the global index is their only list", coreSkillName))
		}
	}
	ruleCounts := countRuleLinks(ruleIndex)
	paths, _ := filepath.Glob(filepath.Join(coreDir, "references", "*.md"))
	sort.Strings(paths)
	present := map[string]bool{}
	for _, path := range paths {
		name := strings.TrimSuffix(filepath.Base(path), ".md")
		present[name] = true
		data, err := os.ReadFile(path)
		if err != nil {
			problems = append(problems, fmt.Sprintf("%s: %v", name, err))
			continue
		}
		text := string(data)
		if strings.HasPrefix(text, "---") {
			problems = append(problems, fmt.Sprintf("%s: a rule leaf carries no frontmatter", name))
		}
		if !strings.HasPrefix(text, "# ") {
			problems = append(problems, fmt.Sprintf("%s: first line is not a title", name))
		}
		if !strings.HasPrefix(leafFirstBodyLine(text), "Read ") {
			problems = append(problems, fmt.Sprintf("%s: first body line does not start with \"Read\"", name))
		}
		limit := leafSizeLimit
		if leafSizeExceptions[name] {
			limit = leafExceptionSizeLimit
		}
		if name == "report-endings" {
			limit = leafReportEndingsSizeLimit
		}
		if len(data) > limit {
			problems = append(problems, fmt.Sprintf("%s: leaf is %d bytes, limit %d", name, len(data), limit))
		}
		if !hasRuleIndex {
			problems = append(problems, fmt.Sprintf("%s: global.md has no %q section", name, ruleLeavesHeading))
		} else if n := ruleCounts[name]; n != 1 {
			problems = append(problems, fmt.Sprintf("%s: listed %d times in %q, want 1", name, n, ruleLeavesHeading))
		}
	}
	for name := range ruleCounts {
		if !present[name] {
			problems = append(problems, fmt.Sprintf("%q lists %s, which does not exist", ruleLeavesHeading, name))
		}
	}
	return problems
}

// leafFirstBodyLine returns the first non-empty line after the title.
func leafFirstBodyLine(text string) string {
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		if i == 0 || strings.TrimSpace(line) == "" {
			continue
		}
		return line
	}
	return ""
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
		if entry.IsDir() && strings.HasPrefix(entry.Name(), legacyRulePrefix) {
			problems = append(problems, fmt.Sprintf("%s: a %s directory remains; the rule leaves are files of %s", entry.Name(), legacyRulePrefix+"*", coreSkillName))
		}
		if entry.IsDir() && isLeafName(entry.Name()) {
			leaves[entry.Name()] = true
		}
	}

	globalText := ""
	if data, err := os.ReadFile(filepath.Join(root, "content", "guidance", "global.md")); err == nil {
		globalText = string(data)
	}
	ruleIndex, hasRuleIndex := sectionText(globalText, ruleLeavesHeading)
	problems = append(problems, ruleLeafProblems(skillsDir, ruleIndex, hasRuleIndex)...)
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
		if len(data) > leafSizeLimit {
			problems = append(problems, fmt.Sprintf("%s: SKILL.md is %d bytes, limit %d", name, len(data), leafSizeLimit))
		}
		if principleErr != nil {
			problems = append(problems, fmt.Sprintf("%s: %s has no SKILL.md", name, principleIndexName))
		} else if n := principleCounts[name]; n != 1 {
			problems = append(problems, fmt.Sprintf("%s: listed %d times in %s, want 1", name, n, principleIndexName))
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
			text := string(data)
			name := filepath.Base(path)
			reporting := playbookReporting(text)
			for _, want := range []string{playbookEndingsLink, playbookReportLink, playbookReportModes} {
				if !strings.Contains(reporting, want) {
					problems = append(problems, fmt.Sprintf("playbook %s lacks %q in its reporting line", name, want))
				}
			}
			if playbookHasStage(text, "flow-deploy") && !strings.Contains(reporting, "deployment row of") {
				problems = append(problems, fmt.Sprintf("playbook %s lacks the deployment row of close report in its reporting line", name))
			}
			for _, mode := range playbookModeNames(text) {
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

// playbookReporting returns the paragraph that starts with **Reporting.**, or
// an empty string when the playbook has none.
func playbookReporting(text string) string {
	for _, para := range strings.Split(text, "\n\n") {
		if strings.HasPrefix(strings.TrimSpace(para), "**Reporting.**") {
			return para
		}
	}
	return ""
}

// playbookHasStage reports whether the paragraph that starts with **Stages.**
// links the named flow.
func playbookHasStage(text, flow string) bool {
	for _, para := range strings.Split(text, "\n\n") {
		if strings.HasPrefix(strings.TrimSpace(para), "**Stages.**") {
			return strings.Contains(para, flow+"/SKILL.md")
		}
	}
	return false
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
	refs := filepath.Join(skills, coreSkillName, "references")
	link := func(name string) string { return "[x](skill:hive-core/references/" + name + ".md)" }
	writeFile(t, filepath.Join(root, "content", "guidance", "global.md"),
		"# Guide\n\n"+ruleLeavesHeading+"\n\n- Before a commit: "+link("a")+".\n- Again: "+link("a")+".\n- Plain: "+link("c")+".\n- Big: "+link("big")+".\n- Gone: "+link("gone")+".\n\n## Next\n")
	writeFile(t, filepath.Join(skills, coreSkillName, "SKILL.md"), leafDoc(coreSkillName, "")+link("a")+"\n")
	writeFile(t, filepath.Join(refs, "a.md"), "# A\n\nRead before a commit.\n")
	writeFile(t, filepath.Join(refs, "b.md"), "---\nname: b\n---\n\n# B\n\nRead never.\n")
	writeFile(t, filepath.Join(refs, "c.md"), "# C\n\nWhen a commit is made.\n")
	writeFile(t, filepath.Join(refs, "big.md"), "# Big\n\nRead always.\n"+strings.Repeat("x", leafSizeLimit))
	writeFile(t, filepath.Join(skills, legacyRulePrefix+"x", "keep"), "")
	writeFile(t, filepath.Join(skills, languageLeafName, "SKILL.md"), leafDoc(languageLeafName, "disable-model-invocation: true\n"))
	writeFile(t, filepath.Join(skills, "hive-mode", "playbooks", "bugfix.md"), strings.Repeat("y", playbookSizeLimit+1))
	writeFile(t, filepath.Join(skills, "hive-mode", "playbooks", "hotfix.md"), "Push "+playbookModePhrase+", then stop in `hold`.\n")
	writeFile(t, filepath.Join(skills, "hive-mode", "playbooks", "extra.md"), "Not one of the playbooks.\n")

	got := strings.Join(leafProblems(root), "\n")
	for _, want := range []string{
		legacyRulePrefix + "x: a " + legacyRulePrefix + "* directory remains",
		"hive-core: frontmatter lacks `disable-model-invocation: true`",
		"hive-core: SKILL.md lists leaves",
		"a: listed 2 times",
		"b: a rule leaf carries no frontmatter",
		"b: first line is not a title",
		"b: listed 0 times",
		"c: first body line does not start with \"Read\"",
		"big: leaf is",
		"lists gone, which does not exist",
		"hive-typescript: must stay model-invocable",
		"hive-typescript: frontmatter has no `paths`",
		"hive-typescript: hive-principle-index has no SKILL.md",
		"playbook bugfix.md is",
		"playbook hotfix.md lacks \"" + playbookEndingsLink + "\" in its reporting line",
		"playbook hotfix.md names the delivery mode `hold` outside the phrase",
		"hive-mode/playbooks holds extra.md",
		"hive-mode/playbooks lacks feature.md",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("missing problem %q in:\n%s", want, got)
		}
	}
}

func TestLegacyRuleDirectoryIsOnlyProblemWhenPresent(t *testing.T) {
	root := t.TempDir()
	skills := filepath.Join(root, "content", "skills")
	if err := os.MkdirAll(filepath.Join(skills, legacyRulePrefix+"x"), 0o755); err != nil {
		t.Fatal(err)
	}
	if got := strings.Join(leafProblems(root), "\n"); !strings.Contains(got, "a "+legacyRulePrefix+"* directory remains") {
		t.Errorf("an empty legacy directory must be reported, got:\n%s", got)
	}
	if err := os.Remove(filepath.Join(skills, legacyRulePrefix+"x")); err != nil {
		t.Fatal(err)
	}
	if got := strings.Join(leafProblems(root), "\n"); strings.Contains(got, "directory remains") {
		t.Errorf("no legacy directory remains, got:\n%s", got)
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
