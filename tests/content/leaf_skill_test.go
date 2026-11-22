package content

import (
	"encoding/json"
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
	ruleLeavesHeading  = "## Rule leaves"
	languageLeafName   = "hive-typescript"
	coreSkillName      = "hive-core"
	// principlesDomain is the domain of the principle leaves; its index is
	// principles/index.md and is not a leaf.
	principlesDomain    = "principles"
	principleIndexFile  = "index.md"
	domainsManifestPath = "assets/domains.json"
	// legacyRulePrefix and legacyPrincipleSkillPrefix are the directory
	// prefixes the rule leaves and the principles had before they moved into
	// hive-core; no directory with them may remain.
	legacyRulePrefix           = "hive-" + "rule-"
	legacyPrincipleSkillPrefix = "hive-" + "principle-"
	// legacyReferencesDir is the folder the rule leaves shared before the
	// domain folders replaced it; hive-core may not hold it.
	legacyReferencesDir = "references"
)

// leafSizeExceptions names each rule leaf, by its file name without the domain
// folder, whose single rule cannot fit the common limit without
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

// playbookKinds is the closed set of hive-mode playbooks that have a file; the
// other kinds of hive-mode's table carry what is specific to them in their row.
var playbookKinds = []string{
	"data-change", "dependency-security", "perf", "refactor", "ticket-status", "ui-change",
}

var deliveryModeNamePattern = regexp.MustCompile("`(interactive|automatic|ready|direct-base|hold)`")

// ruleLeafLinkPattern matches a link to a file of a hive-core domain folder and
// captures the domain and the file name without its extension.
var ruleLeafLinkPattern = regexp.MustCompile(`hive-core/([a-z0-9-]+)/([a-z0-9-]+)\.md`)

// principleIndexLinkPattern matches a link of the principle index to a file of
// the same folder and captures the file name without its extension.
var principleIndexLinkPattern = regexp.MustCompile(`\]\(([a-z0-9-]+)\.md(?:#[^)]*)?\)`)

// notRuleDomain reports whether a folder of hive-core cannot hold rule leaves
// even when a link names a file in it.
func notRuleDomain(domain string) bool {
	return domain == principlesDomain || domain == "scripts" || domain == "assets"
}

// readDomains returns the domain folders hive-core/assets/domains.json lists.
func readDomains(coreDir string) ([]string, error) {
	data, err := os.ReadFile(filepath.Join(coreDir, filepath.FromSlash(domainsManifestPath)))
	if err != nil {
		return nil, err
	}
	var manifest struct {
		Domains []string `json:"domains"`
	}
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("%s: %w", domainsManifestPath, err)
	}
	return manifest.Domains, nil
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

// countRuleLinks counts the links of text to rule leaves, keyed by
// "<domain>/<name>".
func countRuleLinks(text string) map[string]int {
	counts := map[string]int{}
	for _, match := range ruleLeafLinkPattern.FindAllStringSubmatch(text, -1) {
		if !notRuleDomain(match[1]) {
			counts[match[1]+"/"+match[2]]++
		}
	}
	return counts
}

func countPrincipleLinks(text string) map[string]int {
	counts := map[string]int{}
	for _, match := range principleIndexLinkPattern.FindAllStringSubmatch(text, -1) {
		counts[match[1]]++
	}
	return counts
}

// leafFileProblems checks what a rule leaf and a principle share: a title, no
// frontmatter, a first body line that starts with "Read" and the size limit.
func leafFileProblems(label string, data []byte, limit int) []string {
	var problems []string
	text := string(data)
	if strings.HasPrefix(text, "---") {
		problems = append(problems, fmt.Sprintf("%s: a leaf carries no frontmatter", label))
	}
	if !strings.HasPrefix(text, "# ") {
		problems = append(problems, fmt.Sprintf("%s: first line is not a title", label))
	}
	if !strings.HasPrefix(leafFirstBodyLine(text), "Read ") {
		problems = append(problems, fmt.Sprintf("%s: first body line does not start with \"Read\"", label))
	}
	if len(data) > limit {
		problems = append(problems, fmt.Sprintf("%s: leaf is %d bytes, limit %d", label, len(data), limit))
	}
	return problems
}

// ruleLeafProblems checks the hive-core skill, its folders against
// assets/domains.json and each rule leaf of the listed domains: no frontmatter,
// a title, a first body line that starts with "Read", the size limit, a name
// unique across domains, and exactly one line of the global index.
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

	domains, err := readDomains(coreDir)
	if err != nil {
		problems = append(problems, fmt.Sprintf("%s/%s: %v", coreSkillName, domainsManifestPath, err))
	}
	listed := map[string]bool{}
	for _, domain := range domains {
		listed[domain] = true
	}
	if err == nil && !listed[principlesDomain] {
		problems = append(problems, fmt.Sprintf("%s/%s does not list the %s domain", coreSkillName, domainsManifestPath, principlesDomain))
	}
	if entries, err := os.ReadDir(coreDir); err == nil {
		for _, entry := range entries {
			name := entry.Name()
			switch {
			case !entry.IsDir() || name == "assets" || name == "scripts":
			case name == legacyReferencesDir:
				problems = append(problems, fmt.Sprintf("%s/%s: the folder remains; the leaves live in the domain folders", coreSkillName, name))
			case !listed[name]:
				problems = append(problems, fmt.Sprintf("%s/%s: folder is not listed in %s", coreSkillName, name, domainsManifestPath))
			}
		}
	}

	ruleCounts := countRuleLinks(ruleIndex)
	present := map[string]bool{}
	owner := map[string]string{}
	for _, domain := range domains {
		if domain == principlesDomain {
			continue
		}
		paths, _ := filepath.Glob(filepath.Join(coreDir, domain, "*.md"))
		sort.Strings(paths)
		if len(paths) == 0 {
			problems = append(problems, fmt.Sprintf("%s/%s: domain holds no leaf", coreSkillName, domain))
		}
		for _, path := range paths {
			name := strings.TrimSuffix(filepath.Base(path), ".md")
			key := domain + "/" + name
			present[key] = true
			if first, ok := owner[name]; ok {
				problems = append(problems, fmt.Sprintf("%s: the leaf name is in both %s/%s and %s/%s", name, coreSkillName, first, coreSkillName, domain))
			} else {
				owner[name] = domain
			}
			data, err := os.ReadFile(path)
			if err != nil {
				problems = append(problems, fmt.Sprintf("%s: %v", name, err))
				continue
			}
			limit := leafSizeLimit
			if leafSizeExceptions[name] {
				limit = leafExceptionSizeLimit
			}
			if name == "report-endings" {
				limit = leafReportEndingsSizeLimit
			}
			for _, problem := range leafFileProblems(name, data, limit) {
				problems = append(problems, strings.Replace(problem, "a leaf carries", "a rule leaf carries", 1))
			}
			if !hasRuleIndex {
				problems = append(problems, fmt.Sprintf("%s: global.md has no %q section", name, ruleLeavesHeading))
			} else if n := ruleCounts[key]; n != 1 {
				problems = append(problems, fmt.Sprintf("%s: listed %d times in %q, want 1", name, n, ruleLeavesHeading))
			}
		}
	}
	for key := range ruleCounts {
		if !present[key] {
			problems = append(problems, fmt.Sprintf("%q lists %s, which does not exist", ruleLeavesHeading, key))
		}
	}
	return problems
}

// principleProblems checks the principle domain: every file of principles/
// other than index.md has no frontmatter, a title, a first body line that
// starts with "Read", the size limit, and exactly one link in principles/index.md.
func principleProblems(skillsDir string) []string {
	var problems []string
	dir := filepath.Join(skillsDir, coreSkillName, principlesDomain)
	indexData, indexErr := os.ReadFile(filepath.Join(dir, principleIndexFile))
	if indexErr != nil {
		problems = append(problems, fmt.Sprintf("%s/%s/%s: %v", coreSkillName, principlesDomain, principleIndexFile, indexErr))
	}
	counts := countPrincipleLinks(string(indexData))
	paths, _ := filepath.Glob(filepath.Join(dir, "*.md"))
	sort.Strings(paths)
	present := map[string]bool{}
	for _, path := range paths {
		name := strings.TrimSuffix(filepath.Base(path), ".md")
		if filepath.Base(path) == principleIndexFile {
			continue
		}
		present[name] = true
		data, err := os.ReadFile(path)
		if err != nil {
			problems = append(problems, fmt.Sprintf("%s: %v", name, err))
			continue
		}
		for _, problem := range leafFileProblems(name, data, leafSizeLimit) {
			problems = append(problems, strings.Replace(problem, "a leaf carries", "a principle carries", 1))
		}
		if indexErr == nil {
			if n := counts[name]; n != 1 {
				problems = append(problems, fmt.Sprintf("%s: listed %d times in %s/%s, want 1", name, n, principlesDomain, principleIndexFile))
			}
		}
	}
	if len(present) == 0 {
		problems = append(problems, fmt.Sprintf("%s/%s: domain holds no principle", coreSkillName, principlesDomain))
	}
	for name := range counts {
		if !present[name] {
			problems = append(problems, fmt.Sprintf("%s/%s lists %s, which does not exist", principlesDomain, principleIndexFile, name))
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
		if entry.IsDir() && strings.HasPrefix(entry.Name(), legacyPrincipleSkillPrefix) {
			problems = append(problems, fmt.Sprintf("%s: a %s directory remains; the principles are files of %s/%s", entry.Name(), legacyPrincipleSkillPrefix+"*", coreSkillName, principlesDomain))
		}
		if entry.IsDir() && entry.Name() == languageLeafName {
			leaves[entry.Name()] = true
		}
	}

	globalText := ""
	if data, err := os.ReadFile(filepath.Join(root, "content", "guidance", "global.md")); err == nil {
		globalText = string(data)
	}
	ruleIndex, hasRuleIndex := sectionText(globalText, ruleLeavesHeading)
	problems = append(problems, ruleLeafProblems(skillsDir, ruleIndex, hasRuleIndex)...)
	problems = append(problems, principleProblems(skillsDir)...)

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
	core := filepath.Join(skills, coreSkillName)
	link := func(domain, name string) string { return "[x](skill:hive-core/" + domain + "/" + name + ".md)" }
	writeFile(t, filepath.Join(root, "content", "guidance", "global.md"),
		"# Guide\n\n"+ruleLeavesHeading+"\n\n- Before a commit: "+link("build", "a")+".\n- Again: "+link("build", "a")+".\n- Plain: "+link("build", "c")+".\n- Big: "+link("build", "big")+".\n- Gone: "+link("build", "gone")+".\n\n## Next\n")
	writeFile(t, filepath.Join(core, "SKILL.md"), leafDoc(coreSkillName, "")+link("build", "a")+"\n")
	writeFile(t, filepath.Join(core, "assets", "domains.json"), `{"domains": ["build", "plan", "empty", "principles"]}`)
	writeFile(t, filepath.Join(core, "build", "a.md"), "# A\n\nRead before a commit.\n")
	writeFile(t, filepath.Join(core, "build", "b.md"), "---\nname: b\n---\n\n# B\n\nRead never.\n")
	writeFile(t, filepath.Join(core, "build", "c.md"), "# C\n\nWhen a commit is made.\n")
	writeFile(t, filepath.Join(core, "build", "big.md"), "# Big\n\nRead always.\n"+strings.Repeat("x", leafSizeLimit))
	writeFile(t, filepath.Join(core, "plan", "a.md"), "# A again\n\nRead before a plan.\n")
	writeFile(t, filepath.Join(core, "empty", "keep"), "")
	writeFile(t, filepath.Join(core, "stray", "z.md"), "# Z\n\nRead never.\n")
	writeFile(t, filepath.Join(core, legacyReferencesDir, "old.md"), "# Old\n\nRead never.\n")
	writeFile(t, filepath.Join(core, principlesDomain, principleIndexFile), "# Principles\n\nRead first.\n\n- [one](p1.md)\n- [again](p1.md)\n- [two](p2.md)\n- [three](p3.md)\n- [gone](p5.md)\n")
	writeFile(t, filepath.Join(core, principlesDomain, "p1.md"), "---\nname: p1\n---\n\n# P1\n\nRead when you design.\n")
	writeFile(t, filepath.Join(core, principlesDomain, "p2.md"), "# P2\n\nRead when you write.\n"+strings.Repeat("x", leafSizeLimit))
	writeFile(t, filepath.Join(core, principlesDomain, "p3.md"), "# P3\n\nWhen you review.\n")
	writeFile(t, filepath.Join(core, principlesDomain, "p4.md"), "# P4\n\nRead when you ship.\n")
	writeFile(t, filepath.Join(skills, legacyRulePrefix+"x", "keep"), "")
	writeFile(t, filepath.Join(skills, legacyPrincipleSkillPrefix+"x", "keep"), "")
	writeFile(t, filepath.Join(skills, languageLeafName, "SKILL.md"), leafDoc(languageLeafName, "disable-model-invocation: true\n"))
	writeFile(t, filepath.Join(skills, "hive-mode", "playbooks", "perf.md"), strings.Repeat("y", playbookSizeLimit+1))
	writeFile(t, filepath.Join(skills, "hive-mode", "playbooks", "refactor.md"), "Push "+playbookModePhrase+", then stop in `hold`.\n")
	writeFile(t, filepath.Join(skills, "hive-mode", "playbooks", "extra.md"), "Not one of the playbooks.\n")

	got := strings.Join(leafProblems(root), "\n")
	for _, want := range []string{
		legacyRulePrefix + "x: a " + legacyRulePrefix + "* directory remains",
		legacyPrincipleSkillPrefix + "x: a " + legacyPrincipleSkillPrefix + "* directory remains",
		"hive-core: frontmatter lacks `disable-model-invocation: true`",
		"hive-core: SKILL.md lists leaves",
		"hive-core/references: the folder remains",
		"hive-core/stray: folder is not listed in assets/domains.json",
		"hive-core/empty: domain holds no leaf",
		"a: the leaf name is in both hive-core/build and hive-core/plan",
		"a: listed 2 times",
		"b: a rule leaf carries no frontmatter",
		"b: first line is not a title",
		"b: listed 0 times",
		"c: first body line does not start with \"Read\"",
		"big: leaf is",
		"lists build/gone, which does not exist",
		"p1: a principle carries no frontmatter",
		"p1: listed 2 times in principles/index.md",
		"p2: leaf is",
		"p3: first body line does not start with \"Read\"",
		"p4: listed 0 times in principles/index.md",
		"principles/index.md lists p5, which does not exist",
		"hive-typescript: must stay model-invocable",
		"hive-typescript: frontmatter has no `paths`",
		"playbook perf.md is",
		"playbook refactor.md names the delivery mode `hold` outside the phrase",
		"hive-mode/playbooks holds extra.md",
		"hive-mode/playbooks lacks ui-change.md",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("missing problem %q in:\n%s", want, got)
		}
	}
}

func TestLeafSkillInvariantsNeedTheDomainManifest(t *testing.T) {
	root := t.TempDir()
	core := filepath.Join(root, "content", "skills", coreSkillName)
	writeFile(t, filepath.Join(core, "SKILL.md"), leafDoc(coreSkillName, "disable-model-invocation: true\n"))
	if got := strings.Join(leafProblems(root), "\n"); !strings.Contains(got, "hive-core/assets/domains.json: ") {
		t.Errorf("a missing manifest must be reported, got:\n%s", got)
	}
	writeFile(t, filepath.Join(core, "assets", "domains.json"), `{"domains": ["build"]}`)
	if got := strings.Join(leafProblems(root), "\n"); !strings.Contains(got, "does not list the principles domain") {
		t.Errorf("a manifest without principles must be reported, got:\n%s", got)
	}
	writeFile(t, filepath.Join(core, "assets", "domains.json"), `{"domains": ["build", "principles"]}`)
	writeFile(t, filepath.Join(core, principlesDomain, principleIndexFile), "# Principles\n")
	if got := strings.Join(leafProblems(root), "\n"); !strings.Contains(got, "hive-core/principles: domain holds no principle") {
		t.Errorf("a principles domain without a principle must be reported, got:\n%s", got)
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
