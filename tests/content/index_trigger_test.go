package content

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// Each line of the «Rule leaves» index names the moment to read a leaf, and the
// leaf's first body line names the same moment after "Read". A moment worded
// twice drifts: the index is what an agent sees at every session, the leaf is
// what it reads at the moment. The trigger of an index line therefore has to be
// the leaf's own first sentence.

var (
	indexLeafLink    = regexp.MustCompile(`: \[[^\]]+\]\(skill:hive-core/([a-z0-9-]+/[a-z0-9-]+)\.md\)`)
	markdownLinkText = regexp.MustCompile(`\[([^\]]*)\]\([^)]*\)`)
)

// reduceLinks replaces each markdown link with its text.
func reduceLinks(text string) string {
	return markdownLinkText.ReplaceAllString(text, "$1")
}

// leafTrigger returns the first sentence of the leaf's first body line after
// "Read ", with links reduced to their text and without the closing period.
func leafTrigger(leafText string) (string, bool) {
	line := leafFirstBodyLine(leafText)
	if !strings.HasPrefix(line, "Read ") {
		return "", false
	}
	sentence := reduceLinks(strings.TrimPrefix(line, "Read "))
	if i := strings.Index(sentence, ". "); i >= 0 {
		sentence = sentence[:i]
	}
	return strings.TrimSuffix(strings.TrimSpace(sentence), "."), true
}

// indexTriggers returns, for each leaf link of an index line, the trigger text
// that precedes it. Several leaves on one line are separated by "; ".
func indexTriggers(line string) (leaves, triggers []string) {
	rest := strings.TrimPrefix(line, "- ")
	for {
		loc := indexLeafLink.FindStringSubmatchIndex(rest)
		if loc == nil {
			return leaves, triggers
		}
		triggers = append(triggers, reduceLinks(rest[:loc[0]]))
		leaves = append(leaves, rest[loc[2]:loc[3]])
		rest = rest[loc[1]:]
		sep := strings.Index(rest, "; ")
		if sep < 0 {
			return leaves, triggers
		}
		rest = rest[sep+2:]
	}
}

// sameTrigger compares two triggers, equal when they differ at most in the case
// of the first letter.
func sameTrigger(a, b string) bool {
	if a == "" || b == "" {
		return a == b
	}
	return strings.ToLower(a[:1])+a[1:] == strings.ToLower(b[:1])+b[1:]
}

// triggerProblems returns one line per index trigger that is not the first
// sentence of its leaf, and per index line that links no leaf.
func triggerProblems(root string) []string {
	globalData, err := os.ReadFile(filepath.Join(root, "content", "guidance", "global.md"))
	if err != nil {
		return []string{err.Error()}
	}
	section, ok := sectionText(string(globalData), ruleLeavesHeading)
	if !ok {
		return []string{fmt.Sprintf("global.md has no %q section", ruleLeavesHeading)}
	}
	var problems []string
	lines := 0
	for _, line := range strings.Split(section, "\n") {
		if !strings.HasPrefix(line, "- ") {
			continue
		}
		lines++
		leaves, triggers := indexTriggers(line)
		if len(leaves) == 0 {
			problems = append(problems, fmt.Sprintf("index line links no rule leaf: %s", line))
			continue
		}
		for i, leaf := range leaves {
			data, err := os.ReadFile(filepath.Join(root, "content", "skills", coreSkillName, filepath.FromSlash(leaf)+".md"))
			if err != nil {
				problems = append(problems, fmt.Sprintf("%s: %v", leaf, err))
				continue
			}
			want, ok := leafTrigger(string(data))
			if !ok {
				problems = append(problems, fmt.Sprintf("%s: first body line does not start with \"Read\"", leaf))
				continue
			}
			if !sameTrigger(triggers[i], want) {
				problems = append(problems, fmt.Sprintf("%s: index trigger %q is not the leaf's first sentence %q", leaf, triggers[i], want))
			}
		}
	}
	if lines == 0 {
		problems = append(problems, "the index has no lines")
	}
	return problems
}

func TestIndexTriggersAreTheLeavesFirstSentence(t *testing.T) {
	for _, problem := range triggerProblems("../..") {
		t.Error(problem)
	}
}

func TestIndexTriggersCatchDrift(t *testing.T) {
	root := t.TempDir()
	global := func(lines ...string) string {
		return "# Guide\n\n" + ruleLeavesHeading + "\n\nIntro.\n\n" + strings.Join(lines, "\n") + "\n\n## Next\n"
	}
	link := func(name string) string { return ": [x](skill:hive-core/build/" + name + ".md)." }
	writeFile(t, filepath.Join(root, "content", "skills", coreSkillName, "build", "a.md"), "# A\n\nRead before a [commit](commit.md) of a task, in any flow. More.\n")
	writeFile(t, filepath.Join(root, "content", "skills", coreSkillName, "build", "b.md"), "# B\n\nRead at steps 4 and 5.\n")
	writeFile(t, filepath.Join(root, "content", "guidance", "global.md"),
		global("- Before a commit of a task, in any flow"+link("a"),
			"- When a close starts: [x](skill:hive-core/build/b.md); at steps 4 and 5"+link("b")))
	if got := triggerProblems(root); len(got) != 1 || !strings.Contains(got[0], "build/b: index trigger") {
		t.Fatalf("only b drifts, got %v", got)
	}
	writeFile(t, filepath.Join(root, "content", "guidance", "global.md"),
		global("- Before a commit of a task in any flow"+link("a"), "- Nothing linked."))
	got := strings.Join(triggerProblems(root), "\n")
	for _, want := range []string{"build/a: index trigger", "index line links no rule leaf"} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in:\n%s", want, got)
		}
	}
}
